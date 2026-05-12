package migration

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// PluginCompatResult is one row in the plugin or theme compatibility table.
type PluginCompatResult struct {
	Name          string
	Status        string // "Active" (pulled plugins are assumed active)
	Version       string // from plugin/theme header, or "unknown"
	VIPCompatible bool   // true if no phpcs errors
	Issues        string // joined unique phpcs error messages
	Notes         string // pre-populated for known premium plugins
}

// WPEMUPlugin is a WPEngine-specific MU plugin detected in mu-plugins/.
type WPEMUPlugin struct {
	Name   string
	Reason string
}

// MediaStats summarises wp-content/uploads/.
type MediaStats struct {
	TotalSizeHuman string
	FileCount      int
	ExcludedFiles  []string // PHP files not allowed on VIP
}

// SQLAnalysis holds findings from scanning the SQL export.
type SQLAnalysis struct {
	DetectedPrefix  string   // e.g. "wp_2_"
	CollationIssues []string // non-utf8mb4 collations found
}

// known WPEngine MU plugins that must always be removed.
var wpeMUPluginReasons = map[string]string{
	"wpe-cache-plugin":           "Caching handled by WPVIP infrastructure",
	"wpe-wp-sign-on-plugin":      "WPEngine-specific — VIP incompatible",
	"wpe-update-source-selector": "WPEngine-specific — VIP incompatible",
	"force-strong-passwords":     "Functionality managed by WPVIP",
	"slt-force-strong-passwords": "WPEngine-specific — VIP incompatible",
	"wpengine-security-auditor":  "WPEngine-specific — VIP incompatible",
}

// known premium plugin notes.
var premiumPluginNotes = map[string]string{
	"advanced-custom-fields-pro": "Premium plugin — manual update required",
	"gravityforms":               "Premium plugin — manual update required",
	"relevanssi-premium":         "Premium plugin — manual update required",
}

// DetectWPEMUPlugins scans wpContentPath/mu-plugins/ for known WPEngine MU plugins.
func DetectWPEMUPlugins(wpContentPath string) []WPEMUPlugin {
	muPluginsDir := filepath.Join(wpContentPath, "mu-plugins")
	entries, err := os.ReadDir(muPluginsDir)
	if err != nil {
		return nil
	}
	var found []WPEMUPlugin
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if reason, ok := wpeMUPluginReasons[e.Name()]; ok {
			found = append(found, WPEMUPlugin{Name: e.Name(), Reason: reason})
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Name < found[j].Name })
	return found
}

// BuildPluginResults groups audit violations by plugin directory and produces
// one PluginCompatResult per plugin found in wpContentPath/plugins/.
func BuildPluginResults(audit *AuditReport, wpContentPath string) []PluginCompatResult {
	return buildCompatResults(audit, wpContentPath, "plugins", ParsePluginHeader)
}

// BuildThemeResults groups audit violations by theme directory and produces
// one PluginCompatResult per theme found in wpContentPath/themes/.
func BuildThemeResults(audit *AuditReport, wpContentPath string) []PluginCompatResult {
	return buildCompatResults(audit, wpContentPath, "themes", ParseThemeHeader)
}

func buildCompatResults(audit *AuditReport, wpContentPath, subdir string, headerFn func(string) (string, string)) []PluginCompatResult {
	targetDir := filepath.Join(wpContentPath, subdir)
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil
	}

	// Map directory name → unique error messages
	dirErrors := map[string]map[string]bool{}
	for _, f := range audit.Files {
		rel, err := filepath.Rel(targetDir, f.FilePath)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		parts := strings.SplitN(rel, string(os.PathSeparator), 2)
		if len(parts) == 0 {
			continue
		}
		dirName := parts[0]
		for _, msg := range f.Messages {
			if msg.Type == "ERROR" {
				if dirErrors[dirName] == nil {
					dirErrors[dirName] = map[string]bool{}
				}
				dirErrors[dirName][msg.Message] = true
			}
		}
	}

	var results []PluginCompatResult
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name, version := headerFn(filepath.Join(targetDir, e.Name()))
		errMsgs := dirErrors[e.Name()]
		var issues []string
		for msg := range errMsgs {
			issues = append(issues, msg)
		}
		sort.Strings(issues)

		results = append(results, PluginCompatResult{
			Name:          name,
			Status:        "Active",
			Version:       version,
			VIPCompatible: len(errMsgs) == 0,
			Issues:        strings.Join(issues, "; "),
			Notes:         premiumPluginNotes[e.Name()],
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	return results
}

// ParsePluginHeader reads a WordPress plugin header from pluginDir.
// It tries <dirName>.php first, then all PHP files at the directory root.
// Falls back to humanizing the directory name.
func ParsePluginHeader(pluginDir string) (name, version string) {
	dirName := filepath.Base(pluginDir)
	entries, _ := os.ReadDir(pluginDir)
	candidates := []string{filepath.Join(pluginDir, dirName+".php")}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".php") && e.Name() != dirName+".php" {
			candidates = append(candidates, filepath.Join(pluginDir, e.Name()))
		}
	}
	for _, path := range candidates {
		n, v := readWPFileHeader(path, "Plugin Name", "Version")
		if n != "" {
			return n, v
		}
	}
	return humanizeDirName(dirName), "unknown"
}

// ParseThemeHeader reads a WordPress theme header from style.css in themeDir.
func ParseThemeHeader(themeDir string) (name, version string) {
	n, v := readWPFileHeader(filepath.Join(themeDir, "style.css"), "Theme Name", "Version")
	if n != "" {
		return n, v
	}
	return humanizeDirName(filepath.Base(themeDir)), "unknown"
}

// readWPFileHeader scans the first 30 lines of path for WordPress-style
// "Key: Value" header comments.
func readWPFileHeader(path, nameKey, versionKey string) (name, version string) {
	f, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 30 && scanner.Scan(); i++ {
		line := strings.TrimSpace(scanner.Text())
		// Strip leading comment characters: * , //
		line = strings.TrimLeft(line, "/* ")
		if after, ok := strings.CutPrefix(line, nameKey+":"); ok {
			name = strings.TrimSpace(after)
		}
		if after, ok := strings.CutPrefix(line, versionKey+":"); ok {
			version = strings.TrimSpace(after)
		}
		if name != "" && version != "" {
			return
		}
	}
	return
}

// humanizeDirName converts "advanced-custom-fields-pro" → "Advanced Custom Fields Pro".
func humanizeDirName(s string) string {
	words := strings.Split(strings.ReplaceAll(s, "_", "-"), "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// humanizeBytes converts a byte count to a human-readable string.
func humanizeBytes(b int64) string {
	if b == 0 {
		return "unknown"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// AnalyzeMedia walks wpContentPath/uploads/ counting files and identifying
// PHP files excluded by VIP.
func AnalyzeMedia(wpContentPath string) MediaStats {
	uploadsPath := filepath.Join(wpContentPath, "uploads")
	var stats MediaStats
	var totalBytes int64
	_ = filepath.WalkDir(uploadsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		totalBytes += info.Size()
		stats.FileCount++
		if strings.ToLower(filepath.Ext(path)) == ".php" {
			stats.ExcludedFiles = append(stats.ExcludedFiles, filepath.Base(path))
		}
		return nil
	})
	stats.TotalSizeHuman = humanizeBytes(totalBytes)
	return stats
}

// AnalyzeSQLExport scans sqlPath for table prefix and collation issues using
// grep for efficiency on large files. Returns empty SQLAnalysis if sqlPath is "".
func AnalyzeSQLExport(sqlPath string) SQLAnalysis {
	if sqlPath == "" {
		return SQLAnalysis{}
	}
	var analysis SQLAnalysis

	// Detect table prefix from first CREATE TABLE statement
	if out, err := exec.Command("grep", "-m", "1", "^CREATE TABLE", sqlPath).Output(); err == nil {
		analysis.DetectedPrefix = extractTablePrefix(strings.TrimSpace(string(out)))
	}

	// Detect non-utf8mb4 collations (scan up to 20 occurrences)
	if out, err := exec.Command("grep", "-m", "20", "COLLATE", sqlPath).Output(); err == nil {
		seen := map[string]bool{}
		for _, line := range strings.Split(string(out), "\n") {
			if !strings.Contains(line, "utf8mb4") {
				if col := extractCollationName(line); col != "" && !seen[col] {
					seen[col] = true
					analysis.CollationIssues = append(analysis.CollationIssues, col)
				}
			}
		}
	}
	return analysis
}

// extractTablePrefix parses "CREATE TABLE `wp_2_posts`" → "wp_2_".
func extractTablePrefix(createLine string) string {
	start := strings.Index(createLine, "`")
	if start < 0 {
		return ""
	}
	end := strings.Index(createLine[start+1:], "`")
	if end < 0 {
		return ""
	}
	tableName := createLine[start+1 : start+1+end]
	// Find last underscore segment — prefix is everything up to the last word
	idx := strings.LastIndex(tableName, "_")
	if idx < 0 {
		return ""
	}
	return tableName[:idx+1]
}

// extractCollationName parses a COLLATE clause to return the collation name.
func extractCollationName(line string) string {
	idx := strings.Index(strings.ToUpper(line), "COLLATE")
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(line[idx+len("COLLATE"):])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], "`,'")
}

// enrichedReportData holds all data passed to the enriched report template.
type enrichedReportData struct {
	Install        string
	Destination    string
	GeneratedAt    string
	PluginResults  []PluginCompatResult
	ThemeResults   []PluginCompatResult
	WPEMUPlugins   []WPEMUPlugin
	SQLAnalysis    SQLAnalysis
	SQLPath        string
	SQLSizeHuman   string
	MediaStats     MediaStats
	MissingFromVIP []string
	MissingFromWPE []string
	TotalErrors    int
	TotalWarnings  int
}

var enrichedReportTmpl = template.Must(template.New("enriched-report").Funcs(template.FuncMap{
	"join": strings.Join,
	"sort": func(s []string) []string {
		cp := make([]string, len(s))
		copy(cp, s)
		sort.Strings(cp)
		return cp
	},
}).Parse(enrichedReportTemplate))

const enrichedReportTemplate = `# Migration Report: {{ .Install }}

**Site:** {{ .Install }}
**Generated:** {{ .GeneratedAt }}
**Destination:** {{ .Destination }}

---

## 1. Plugin Compatibility

| Name | Status | Version | VIP Compatible | Compatibility Issues | Notes |
|------|--------|---------|----------------|----------------------|-------|
{{ range .PluginResults -}}
| {{ .Name }} | {{ .Status }} | {{ .Version }} | {{ if .VIPCompatible }}Yes{{ else }}No{{ end }} | {{ .Issues }} | {{ .Notes }} |
{{ end }}
---

## 2. WPEngine Plugins Removed

{{ if .WPEMUPlugins -}}
| Plugin | Reason |
|--------|--------|
{{ range .WPEMUPlugins -}}
| {{ .Name }} | {{ .Reason }} |
{{ end }}
{{- else -}}
No WPEngine-specific MU plugins detected.
{{ end }}
---

## 3. Theme Compatibility

| Name | Status | Version | VIP Compatible | Compatibility Issues |
|------|--------|---------|----------------|----------------------|
{{ range .ThemeResults -}}
| {{ .Name }} | {{ .Status }} | {{ .Version }} | {{ if .VIPCompatible }}Yes{{ else }}No{{ end }} | {{ .Issues }} |
{{ end }}
---

## 4. Database Analysis

**Table prefix detected:** {{ if .SQLAnalysis.DetectedPrefix }}` + "`{{ .SQLAnalysis.DetectedPrefix }}`" + `{{ else }}` + "`wp_`" + ` (standard){{ end }}
{{ if .SQLAnalysis.CollationIssues -}}
**Collation issues:** {{ join .SQLAnalysis.CollationIssues ", " }}
{{- else -}}
**Collation:** No incompatible collations detected.
{{ end }}
**Export file:** {{ if .SQLPath }}{{ .SQLPath }} ({{ .SQLSizeHuman }}){{ else }}Not provided.{{ end }}

---

## 5. Media Migration

**Total size:** {{ .MediaStats.TotalSizeHuman }}
**File count:** {{ .MediaStats.FileCount }}
{{ if .MediaStats.ExcludedFiles -}}
**Excluded (PHP files not allowed on VIP):** {{ len .MediaStats.ExcludedFiles }} file(s) — {{ join .MediaStats.ExcludedFiles ", " }}
{{ end }}
---

## 6. File Comparison

### Present in WPEngine, missing from VIP repo

{{ if .MissingFromVIP -}}
{{ range (sort .MissingFromVIP) -}}
- {{ . }}
{{ end }}
{{- else -}}
None.
{{ end }}

### Present in VIP repo, missing from WPEngine

{{ if .MissingFromWPE -}}
{{ range (sort .MissingFromWPE) -}}
- {{ . }}
{{ end }}
{{- else -}}
None.
{{ end }}

---

## 7. Known Issues / Operator Notes

### Standard VIP post-launch considerations

- **Third-party domain whitelisting**: Services like Google reCAPTCHA, Firecrown dashboard, and ad networks require domain whitelisting after DNS cutover. Not a migration blocker.
- **Plugin constants via WPVIP envvars**: Plugins requiring PHP constants (e.g., API keys) must have them set via WPVIP environment variables in ` + "`vip-config.php`" + `.
{{ if .SQLAnalysis.DetectedPrefix -}}
- **Table prefix conversion**: Prefix ` + "`{{ .SQLAnalysis.DetectedPrefix }}`" + ` detected — search-replace must cover prefix conversion and URL changes.
{{ end }}
### Operator Notes

<!-- Add site-specific observations here before running stax migrate publish -->

---

## 8. Summary

**phpcs audit:** {{ .TotalErrors }} errors, {{ .TotalWarnings }} warnings
**Missing from VIP repo:** {{ len .MissingFromVIP }} items
**Missing from WPEngine:** {{ len .MissingFromWPE }} items
**WPEngine MU plugins to remove:** {{ len .WPEMUPlugins }}
**Export size:** {{ .SQLSizeHuman }}
`
