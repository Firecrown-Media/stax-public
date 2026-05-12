package migration

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
