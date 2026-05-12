# Migration Report & Publish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enrich `stax migrate report` to generate a comprehensive VIP-style migration document and add `stax migrate publish` to upload it to S3 and commit it to the site's VIP repo.

**Architecture:** Extract report helpers to a new `pkg/migration/report.go`; enrich the template in `service.go`; add `Publish()` to `service.go`; add `migratePublishCmd` to `cmd/migrate.go`; rename `--vip-repo` to `--repo` throughout.

**Tech Stack:** Go, `text/template`, `bufio`, `os/exec` (aws cli, git), `filepath.WalkDir`

---

## Codebase orientation

**`pkg/migration/service.go`** (292 lines) — migration orchestration functions (`Pull`, `Export`, `Audit`, `Compare`, `Import`, `Report`, `Status`). Also contains `reportTmpl`, `reportTemplate` const, and `humanizeBytes`. The existing `Report()` function re-runs audit and compare then writes a simple template.

**`pkg/migration/interfaces.go`** — `AuditReport`, `CompareResult`, `AuditOptions`, etc. Do not modify.

**`pkg/migration/providers/vip/destination.go`** — `VIPDestination` with `Audit()`, `CompareFiles()`, etc. Do not modify.

**`cmd/migrate.go`** (292 lines) — uses `migVIPRepoPath` variable and `--vip-repo` flag on compare and report commands (lines 17, 247–248, 256, 259).

**`pkg/migration/service_test.go`** — uses `registerTestProviders(t)` helper with `mockDst2` and `cfgWithDest()`. Follow this pattern for new tests.

---

## File Map

**Create:**
- `pkg/migration/report.go` — report data types, plugin/theme helpers, SQL/media analysis, full report template
- `pkg/migration/report_test.go` — tests for all helpers in report.go

**Modify:**
- `pkg/migration/service.go` — rename `ReportOptions.VIPRepoPath` → `RepoPath`; remove `reportTmpl`/`reportTemplate`/`humanizeBytes` (moved to report.go); update `Report()` to collect enriched data; add `Publish()` and `PublishOptions`
- `pkg/migration/service_test.go` — add tests for `Publish()`
- `cmd/migrate.go` — rename `migVIPRepoPath` → `migRepoPath` and `--vip-repo` → `--repo`; add `migratePublishCmd`
- `docs/runbooks/migration.md` — add Steps 8–9 and WPEngine-specific considerations

---

## Task 1: Rename `--vip-repo` to `--repo`

**Files:**
- Modify: `cmd/migrate.go:17,133–134,147,181–182,201,247–248,256,259`
- Modify: `pkg/migration/service.go:119,144`

- [ ] **Step 1: Write a failing test verifying the flag name**

```go
// pkg/migration/service_test.go — add at the bottom
func TestReportOptions_RepoPath(t *testing.T) {
	opts := migration.ReportOptions{RepoPath: "/some/repo"}
	if opts.RepoPath != "/some/repo" {
		t.Errorf("expected RepoPath, got empty")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /path/to/stax
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestReportOptions_RepoPath -v
```

Expected: FAIL — `opts.RepoPath` field doesn't exist yet.

- [ ] **Step 3: Rename `VIPRepoPath` → `RepoPath` in service.go**

In `pkg/migration/service.go` line 119, change:
```go
// before
type ReportOptions struct {
	LocalPath   string // path to downloaded wp-content
	VIPRepoPath string // path to VIP repo checkout
	SQLPath     string // path to exported SQL file (optional)
	OutputPath  string // output markdown path; defaults to .stax/migration-report.md
}
```
to:
```go
// after
type ReportOptions struct {
	LocalPath  string // path to downloaded wp-content
	RepoPath   string // path to VIP repo checkout
	SQLPath    string // path to exported SQL file (optional)
	OutputPath string // output markdown path; defaults to .stax/migration-report.md
}
```

Also update line 144:
```go
// before
compare, err := Compare(cfg, opts.LocalPath, opts.VIPRepoPath)
// after
compare, err := Compare(cfg, opts.LocalPath, opts.RepoPath)
```

- [ ] **Step 4: Rename in cmd/migrate.go**

Line 17:
```go
// before
migVIPRepoPath   string
// after
migRepoPath      string
```

Lines 133–134 (compare example):
```go
// before
Example: `  stax migrate compare --vip-repo=../vip-repo
  stax migrate compare --path=../wpe/wp-content --vip-repo=../vip-repo`,
// after
Example: `  stax migrate compare --repo=../vip-repo
  stax migrate compare --path=../wpe/wp-content --repo=../vip-repo`,
```

Line 147:
```go
// before
result, err := migration.Compare(cfg, localPath, migVIPRepoPath)
// after
result, err := migration.Compare(cfg, localPath, migRepoPath)
```

Lines 181–182 (report example):
```go
// before
Example: `  stax migrate report --vip-repo=../vip-repo
  stax migrate report --path=../wpe/wp-content --vip-repo=../vip-repo --sql=.stax/mysite-export.sql`,
// after
Example: `  stax migrate report --repo=../vip-repo
  stax migrate report --path=../wpe/wp-content --repo=../vip-repo --sql=.stax/mysite-export.sql`,
```

Line 201:
```go
// before
VIPRepoPath: migVIPRepoPath,
// after
RepoPath: migRepoPath,
```

Lines 247–248:
```go
// before
migrateCompareCmd.Flags().StringVar(&migVIPRepoPath, "vip-repo", "", "path to local VIP repo checkout")
_ = migrateCompareCmd.MarkFlagRequired("vip-repo")
// after
migrateCompareCmd.Flags().StringVar(&migRepoPath, "repo", "", "path to local VIP repo checkout")
_ = migrateCompareCmd.MarkFlagRequired("repo")
```

Lines 256–259:
```go
// before
migrateReportCmd.Flags().StringVar(&migVIPRepoPath, "vip-repo", "", "path to local VIP repo checkout")
migrateReportCmd.Flags().StringVar(&migSQLPath, "sql", "", "path to SQL dump file (optional)")
migrateReportCmd.Flags().StringVar(&migOutputPath, "output", "", "output path for report (default: .stax/migration-report.md)")
_ = migrateReportCmd.MarkFlagRequired("vip-repo")
// after
migrateReportCmd.Flags().StringVar(&migRepoPath, "repo", "", "path to local VIP repo checkout")
migrateReportCmd.Flags().StringVar(&migSQLPath, "sql", "", "path to SQL dump file (optional)")
migrateReportCmd.Flags().StringVar(&migOutputPath, "output", "", "output path for report (default: .stax/migration-report.md)")
_ = migrateReportCmd.MarkFlagRequired("repo")
```

- [ ] **Step 5: Run tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... ./cmd/... -v 2>&1 | tail -20
```

Expected: all pass including `TestReportOptions_RepoPath`.

- [ ] **Step 6: Commit**

```bash
git add pkg/migration/service.go cmd/migrate.go pkg/migration/service_test.go
git commit -m "refactor(migrate): rename --vip-repo flag to --repo"
```

---

## Task 2: Create `pkg/migration/report.go` with data types and plugin helpers

**Files:**
- Create: `pkg/migration/report.go`
- Create: `pkg/migration/report_test.go`

- [ ] **Step 1: Write failing tests for plugin/theme helpers**

Create `pkg/migration/report_test.go`:

```go
package migration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/firecrown-media/stax/pkg/migration"
)

func TestParsePluginHeader(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "my-plugin")
	_ = os.MkdirAll(pluginDir, 0755)
	content := "<?php\n/**\n * Plugin Name: My Plugin\n * Version: 1.2.3\n */\n"
	_ = os.WriteFile(filepath.Join(pluginDir, "my-plugin.php"), []byte(content), 0644)

	name, version := migration.ParsePluginHeader(pluginDir)
	if name != "My Plugin" {
		t.Errorf("expected 'My Plugin', got %q", name)
	}
	if version != "1.2.3" {
		t.Errorf("expected '1.2.3', got %q", version)
	}
}

func TestParsePluginHeader_FallbackToDirName(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "advanced-custom-fields")
	_ = os.MkdirAll(pluginDir, 0755)

	name, version := migration.ParsePluginHeader(pluginDir)
	if name != "Advanced Custom Fields" {
		t.Errorf("expected 'Advanced Custom Fields', got %q", name)
	}
	if version != "unknown" {
		t.Errorf("expected 'unknown', got %q", version)
	}
}

func TestParseThemeHeader(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "my-theme")
	_ = os.MkdirAll(themeDir, 0755)
	content := "/*\nTheme Name: My Theme\nVersion: 2.0.0\n*/\n"
	_ = os.WriteFile(filepath.Join(themeDir, "style.css"), []byte(content), 0644)

	name, version := migration.ParseThemeHeader(themeDir)
	if name != "My Theme" {
		t.Errorf("expected 'My Theme', got %q", name)
	}
	if version != "2.0.0" {
		t.Errorf("expected '2.0.0', got %q", version)
	}
}

func TestDetectWPEMUPlugins(t *testing.T) {
	dir := t.TempDir()
	muPluginsDir := filepath.Join(dir, "mu-plugins")
	_ = os.MkdirAll(filepath.Join(muPluginsDir, "wpe-cache-plugin"), 0755)
	_ = os.MkdirAll(filepath.Join(muPluginsDir, "force-strong-passwords"), 0755)
	_ = os.MkdirAll(filepath.Join(muPluginsDir, "my-custom-mu-plugin"), 0755)

	detected := migration.DetectWPEMUPlugins(dir)
	if len(detected) != 2 {
		t.Fatalf("expected 2 WPE plugins, got %d: %v", len(detected), detected)
	}
	names := map[string]bool{detected[0].Name: true, detected[1].Name: true}
	if !names["wpe-cache-plugin"] || !names["force-strong-passwords"] {
		t.Errorf("unexpected detected plugins: %v", detected)
	}
}

func TestBuildPluginResults(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	_ = os.MkdirAll(filepath.Join(pluginsDir, "clean-plugin"), 0755)
	_ = os.WriteFile(filepath.Join(pluginsDir, "clean-plugin", "clean-plugin.php"),
		[]byte("<?php\n/**\n * Plugin Name: Clean Plugin\n * Version: 1.0\n */\n"), 0644)
	_ = os.MkdirAll(filepath.Join(pluginsDir, "bad-plugin"), 0755)

	audit := &migration.AuditReport{
		Files: []migration.FileAuditResult{
			{
				FilePath: filepath.Join(pluginsDir, "bad-plugin", "bad.php"),
				Errors:   1,
				Messages: []migration.AuditMessage{
					{Type: "ERROR", Message: "Direct database query", Severity: 5},
				},
			},
		},
		TotalErrors: 1,
	}

	results := migration.BuildPluginResults(audit, dir)
	if len(results) != 2 {
		t.Fatalf("expected 2 plugin results, got %d", len(results))
	}
	byName := map[string]migration.PluginCompatResult{}
	for _, r := range results {
		byName[r.Name] = r
	}
	if !byName["Clean Plugin"].VIPCompatible {
		t.Error("clean-plugin should be VIP compatible")
	}
	if byName["Bad Plugin"].VIPCompatible {
		t.Error("bad-plugin should not be VIP compatible")
	}
	if byName["Bad Plugin"].Issues == "" {
		t.Error("bad-plugin should have issues listed")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run "TestParsePluginHeader|TestParseThemeHeader|TestDetectWPEMUPlugins|TestBuildPluginResults" -v 2>&1 | tail -15
```

Expected: FAIL — functions not defined.

- [ ] **Step 3: Create `pkg/migration/report.go` with types and helpers**

```go
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
```

- [ ] **Step 4: Run tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run "TestParsePluginHeader|TestParseThemeHeader|TestDetectWPEMUPlugins|TestBuildPluginResults" -v
```

Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add pkg/migration/report.go pkg/migration/report_test.go
git commit -m "feat(migration): add report data types and plugin/theme helpers"
```

---

## Task 3: Add SQL analysis, media stats, and report template to `report.go`

**Files:**
- Modify: `pkg/migration/report.go`
- Modify: `pkg/migration/report_test.go`

- [ ] **Step 1: Write failing tests for SQL and media helpers**

Add to `pkg/migration/report_test.go`:

```go
func TestAnalyzeMedia(t *testing.T) {
	dir := t.TempDir()
	uploads := filepath.Join(dir, "uploads")
	_ = os.MkdirAll(uploads, 0755)
	_ = os.WriteFile(filepath.Join(uploads, "photo.jpg"), make([]byte, 2048), 0644)
	_ = os.WriteFile(filepath.Join(uploads, "shell.php"), make([]byte, 512), 0644)

	stats := migration.AnalyzeMedia(dir)
	if stats.FileCount != 2 {
		t.Errorf("expected 2 files, got %d", stats.FileCount)
	}
	if len(stats.ExcludedFiles) != 1 || stats.ExcludedFiles[0] != "shell.php" {
		t.Errorf("expected shell.php in excluded, got %v", stats.ExcludedFiles)
	}
	if stats.TotalSizeHuman == "" || stats.TotalSizeHuman == "unknown" {
		t.Errorf("expected non-empty size, got %q", stats.TotalSizeHuman)
	}
}

func TestAnalyzeSQLExport_Empty(t *testing.T) {
	analysis := migration.AnalyzeSQLExport("")
	if analysis.DetectedPrefix != "" {
		t.Errorf("expected empty prefix for empty path, got %q", analysis.DetectedPrefix)
	}
}

func TestAnalyzeSQLExport_Prefix(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.sql")
	_, _ = f.WriteString("CREATE TABLE `wp_2_posts` (\n  `ID` bigint(20)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;\n")
	f.Close()

	analysis := migration.AnalyzeSQLExport(f.Name())
	if analysis.DetectedPrefix != "wp_2_" {
		t.Errorf("expected 'wp_2_', got %q", analysis.DetectedPrefix)
	}
}

func TestAnalyzeSQLExport_CollationIssue(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.sql")
	_, _ = f.WriteString("CREATE TABLE `wp_posts` (\n  `post_title` text COLLATE latin1_swedish_ci\n);\n")
	f.Close()

	analysis := migration.AnalyzeSQLExport(f.Name())
	if len(analysis.CollationIssues) == 0 {
		t.Error("expected collation issues, got none")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run "TestAnalyzeMedia|TestAnalyzeSQLExport" -v 2>&1 | tail -10
```

Expected: FAIL — functions not defined.

- [ ] **Step 3: Add SQL/media helpers and report template to `report.go`**

Append to `pkg/migration/report.go`:

```go
import (
	// add these to the existing import block
	"fmt"
	"io/fs"
	"os/exec"
	"sort"
	"strings"
	"text/template"
	"time"
)
```

Replace the import block at the top of `report.go` with:

```go
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
	"time"
)
```

Append these functions and the template to `report.go`:

```go
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

// enrichedReportData holds all data passed to the report template.
type enrichedReportData struct {
	Install       string
	Destination   string
	GeneratedAt   string
	PluginResults []PluginCompatResult
	ThemeResults  []PluginCompatResult
	WPEMUPlugins  []WPEMUPlugin
	SQLAnalysis   SQLAnalysis
	SQLPath       string
	SQLSizeHuman  string
	MediaStats    MediaStats
	MissingFromVIP []string
	MissingFromWPE []string
	TotalErrors   int
	TotalWarnings int
}

var reportTmpl = template.Must(template.New("report").Funcs(template.FuncMap{
	"join": strings.Join,
	"len":  func(s []string) int { return len(s) },
	"sort": func(s []string) []string {
		cp := make([]string, len(s))
		copy(cp, s)
		sort.Strings(cp)
		return cp
	},
	"now": func() string { return time.Now().Format("2006-01-02") },
}).Parse(reportTemplate))

const reportTemplate = `# Migration Report: {{ .Install }}

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
```

- [ ] **Step 4: Run tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run "TestAnalyzeMedia|TestAnalyzeSQLExport" -v
```

Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add pkg/migration/report.go pkg/migration/report_test.go
git commit -m "feat(migration): add SQL/media analysis helpers and enriched report template"
```

---

## Task 4: Update `Report()` in `service.go` to use enriched template

**Files:**
- Modify: `pkg/migration/service.go`
- Modify: `pkg/migration/service_test.go`

- [ ] **Step 1: Write failing test for enriched report content**

Add to `pkg/migration/service_test.go`:

```go
func TestReport_EnrichedSections(t *testing.T) {
	registerTestProviders(t)

	// Set up a minimal wp-content directory
	tmpDir := t.TempDir()
	wpContent := filepath.Join(tmpDir, "wp-content")
	_ = os.MkdirAll(filepath.Join(wpContent, "plugins", "my-plugin"), 0755)
	_ = os.WriteFile(
		filepath.Join(wpContent, "plugins", "my-plugin", "my-plugin.php"),
		[]byte("<?php\n/**\n * Plugin Name: My Plugin\n * Version: 1.0\n */\n"),
		0644,
	)
	_ = os.MkdirAll(filepath.Join(wpContent, "themes", "my-theme"), 0755)
	_ = os.WriteFile(
		filepath.Join(wpContent, "themes", "my-theme", "style.css"),
		[]byte("/*\nTheme Name: My Theme\nVersion: 1.0\n*/\n"),
		0644,
	)
	_ = os.MkdirAll(filepath.Join(wpContent, "mu-plugins", "wpe-cache-plugin"), 0755)

	vipRepo := t.TempDir()
	outputPath := filepath.Join(tmpDir, "migration-report.md")

	cfg := cfgWithDest("test-dst")
	opts := migration.ReportOptions{
		LocalPath:  wpContent,
		RepoPath:   vipRepo,
		OutputPath: outputPath,
	}

	err := migration.Report(nil, cfg, opts)
	if err != nil {
		t.Fatalf("Report() failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	s := string(content)

	for _, want := range []string{
		"Plugin Compatibility",
		"My Plugin",
		"Theme Compatibility",
		"My Theme",
		"WPEngine Plugins Removed",
		"wpe-cache-plugin",
		"Operator Notes",
		"Summary",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("report missing %q", want)
		}
	}
}
```

Also add the needed import at the top of the test file:
```go
import (
	"os"
	"path/filepath"
	// existing imports...
)
```

- [ ] **Step 2: Run test to verify it fails**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestReport_EnrichedSections -v 2>&1 | tail -15
```

Expected: FAIL (report doesn't contain Plugin Compatibility etc. yet).

- [ ] **Step 3: Update `Report()` in `service.go` to use enriched data**

Remove the `reportTmpl` variable declaration and `reportTemplate` constant from `service.go` (they now live in `report.go`).

Replace the `Report()` function body (lines 125–188) in `pkg/migration/service.go`:

```go
// Report runs audit and compare, collects plugin/DB/media analysis, then writes
// a comprehensive VIP-style migration report.
func Report(p provider.Provider, cfg *config.Config, opts ReportOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}

	install := config.ProviderConfigString(cfg.ProviderConfig, "install")

	if opts.OutputPath == "" {
		opts.OutputPath = filepath.Join(".stax", install+"-migration-report.md")
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	audit, err := Audit(cfg, opts.LocalPath, AuditOptions{})
	if err != nil {
		return fmt.Errorf("audit failed: %w", err)
	}

	compare, err := Compare(cfg, opts.LocalPath, opts.RepoPath)
	if err != nil {
		return fmt.Errorf("compare failed: %w", err)
	}

	var sqlSize int64
	if opts.SQLPath != "" {
		if info, err := os.Stat(opts.SQLPath); err == nil {
			sqlSize = info.Size()
		}
	}

	data := enrichedReportData{
		Install:        install,
		Destination:    cfg.Migration.Destination,
		GeneratedAt:    time.Now().Format("2006-01-02 15:04:05 MST"),
		PluginResults:  BuildPluginResults(audit, opts.LocalPath),
		ThemeResults:   BuildThemeResults(audit, opts.LocalPath),
		WPEMUPlugins:   DetectWPEMUPlugins(opts.LocalPath),
		SQLAnalysis:    AnalyzeSQLExport(opts.SQLPath),
		SQLPath:        opts.SQLPath,
		SQLSizeHuman:   humanizeBytes(sqlSize),
		MediaStats:     AnalyzeMedia(opts.LocalPath),
		MissingFromVIP: compare.MissingFromVIP,
		MissingFromWPE: compare.MissingFromWPE,
		TotalErrors:    audit.TotalErrors,
		TotalWarnings:  audit.TotalWarnings,
	}

	f, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer f.Close()

	if err := reportTmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	ui.Success("Report written to %s", opts.OutputPath)
	return nil
}
```

Also remove the old `humanizeBytes` function from `service.go` (it's now in `report.go`) and remove the old `reportTmpl` variable and `reportTemplate` const.

Remove the now-unneeded imports from `service.go` (`sort`, `strings`, `text/template` — they moved to `report.go`). Keep `time` since `Report()` still uses `time.Now()`.

- [ ] **Step 4: Run tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestReport_EnrichedSections -v
```

Expected: PASS.

- [ ] **Step 5: Run full test suite**

```bash
PATH="/opt/homebrew/bin:$PATH" make test
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add pkg/migration/service.go pkg/migration/service_test.go
git commit -m "feat(migration): enrich report output with VIP-style plugin/DB/media sections"
```

---

## Task 5: Add `Publish()` to `service.go`

**Files:**
- Modify: `pkg/migration/service.go`
- Modify: `pkg/migration/service_test.go`

- [ ] **Step 1: Write failing tests for Publish()**

Add to `pkg/migration/service_test.go`:

```go
func TestPublish_MissingReport(t *testing.T) {
	cfg := cfgWithDest("test-dst")
	err := migration.Publish(cfg, migration.PublishOptions{
		RepoPath:   t.TempDir(),
		ReportPath: "/nonexistent/migration-report.md",
	})
	if err == nil {
		t.Fatal("expected error for missing report")
	}
	if !strings.Contains(err.Error(), "report not found") {
		t.Errorf("expected 'report not found' error, got: %v", err)
	}
}

func TestPublish_MissingRepo(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "migration-report.md")
	_ = os.WriteFile(reportPath, []byte("# Report\n"), 0644)

	cfg := cfgWithDest("test-dst")
	err := migration.Publish(cfg, migration.PublishOptions{
		RepoPath:   "/nonexistent/repo",
		ReportPath: reportPath,
	})
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestPublish -v 2>&1 | tail -10
```

Expected: FAIL — `Publish` not defined.

- [ ] **Step 3: Add `PublishOptions` and `Publish()` to `service.go`**

Add after the `Status()` function in `pkg/migration/service.go`:

```go
const defaultAssetsBucket = "firecrown-assets-378073025324"

// PublishOptions configures stax migrate publish.
type PublishOptions struct {
	RepoPath   string // path to local VIP repo checkout (required)
	ReportPath string // path to migration-report.md; defaults to .stax/<install>-migration-report.md
	SQLPath    string // path to SQL export; defaults to .stax/<install>-export.sql
}

// Publish uploads the migration report and SQL export to S3, then commits the
// report to the VIP repo docs/ folder and pushes.
func Publish(cfg *config.Config, opts PublishOptions) error {
	install := config.ProviderConfigString(cfg.ProviderConfig, "install")

	if opts.ReportPath == "" {
		opts.ReportPath = filepath.Join(".stax", install+"-migration-report.md")
	}
	if opts.SQLPath == "" {
		opts.SQLPath = filepath.Join(".stax", install+"-export.sql")
	}

	// 1. Verify report exists
	if _, err := os.Stat(opts.ReportPath); err != nil {
		return fmt.Errorf("report not found at %s: run 'stax migrate report' first", opts.ReportPath)
	}

	// 2. Verify repo exists
	if _, err := os.Stat(opts.RepoPath); err != nil {
		return fmt.Errorf("VIP repo not found at %s", opts.RepoPath)
	}

	s3Prefix := fmt.Sprintf("s3://%s/vip-migration/%s", defaultAssetsBucket, install)

	// 3. Upload report to S3
	ui.Info("Uploading report to S3...")
	if err := runExec("aws", "s3", "cp", opts.ReportPath, s3Prefix+"/migration-report.md"); err != nil {
		return fmt.Errorf("S3 upload of report failed: %w", err)
	}

	// 4. Upload SQL export to S3 (if present)
	if _, err := os.Stat(opts.SQLPath); err == nil {
		ui.Info("Uploading SQL export to S3...")
		s3SQL := fmt.Sprintf("%s/%s-export.sql", s3Prefix, install)
		if err := runExec("aws", "s3", "cp", opts.SQLPath, s3SQL); err != nil {
			return fmt.Errorf("S3 upload of SQL failed: %w", err)
		}
	}

	// 5. Copy report to VIP repo docs/
	docsDir := filepath.Join(opts.RepoPath, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs/ in VIP repo: %w", err)
	}
	if err := copyFile(opts.ReportPath, filepath.Join(docsDir, "migration-report.md")); err != nil {
		return fmt.Errorf("failed to copy report to VIP repo: %w", err)
	}

	// 6. Git commit and push
	ui.Info("Committing report to VIP repo...")
	if err := runExec("git", "-C", opts.RepoPath, "add", "docs/migration-report.md"); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	commitMsg := fmt.Sprintf("docs: add migration report for %s", install)
	if err := runExec("git", "-C", opts.RepoPath, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	shaOut, err := exec.Command("git", "-C", opts.RepoPath, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("git rev-parse failed: %w", err)
	}
	sha := strings.TrimSpace(string(shaOut))

	if err := runExec("git", "-C", opts.RepoPath, "push"); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	ui.Success("Published:")
	ui.Info("  S3 report: %s/migration-report.md", s3Prefix)
	ui.Info("  S3 SQL:    %s/%s-export.sql", s3Prefix, install)
	ui.Info("  VIP repo:  %s/docs/migration-report.md (commit %s)", opts.RepoPath, sha)
	return nil
}

func runExec(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, string(out))
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
```

Add `"io"` and `"os/exec"` and `"strings"` to the import block in `service.go` if not already present.

- [ ] **Step 4: Run tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestPublish -v
```

Expected: both `TestPublish_MissingReport` and `TestPublish_MissingRepo` pass.

- [ ] **Step 5: Run full test suite**

```bash
PATH="/opt/homebrew/bin:$PATH" make test
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add pkg/migration/service.go pkg/migration/service_test.go
git commit -m "feat(migration): add Publish() to upload report to S3 and commit to VIP repo"
```

---

## Task 6: Add `stax migrate publish` command

**Files:**
- Modify: `cmd/migrate.go`

- [ ] **Step 1: Write failing test for publish command registration**

Add to `cmd/migrate_test.go` (create if it doesn't exist):

```go
package cmd_test

import (
	"testing"

	"github.com/firecrown-media/stax/cmd"
)

func TestMigratePublishCmd_Registered(t *testing.T) {
	root := cmd.NewRootCmd()
	migrate, _, err := root.Find([]string{"migrate"})
	if err != nil || migrate == nil {
		t.Fatal("migrate command not found")
	}
	publish, _, err := migrate.Find([]string{"publish"})
	if err != nil || publish == nil {
		t.Fatal("migrate publish command not found")
	}
	repoFlag := publish.Flags().Lookup("repo")
	if repoFlag == nil {
		t.Error("publish command missing --repo flag")
	}
}
```

Check if `cmd.NewRootCmd()` exists:

```bash
grep -n "NewRootCmd\|func NewRoot" /Users/geoff/_projects/fc/stax/cmd/root.go | head -5
```

If `NewRootCmd` doesn't exist, write the test differently using `Execute` or skip the command registration test and rely on build success + integration test.

- [ ] **Step 2: Add `migratePublishCmd` to `cmd/migrate.go`**

Add the `migratePublishCmd` variable alongside the other command variables:

```go
var migratePublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Upload migration report to S3 and commit to VIP repo",
	Long: `Upload the migration report and SQL export to S3, then commit the
report to the VIP repo docs/ folder and push.

Run 'stax migrate report' first, then review and annotate the report
before running publish.`,
	Example: `  stax migrate publish --repo=../my-vip-repo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migDestination != "" {
			cfg.Migration.Destination = migDestination
		}
		install := config.ProviderConfigString(cfg.ProviderConfig, "install")
		return migration.Publish(cfg, migration.PublishOptions{
			RepoPath:   migRepoPath,
			ReportPath: filepath.Join(".stax", install+"-migration-report.md"),
			SQLPath:    filepath.Join(".stax", install+"-export.sql"),
		})
	},
}
```

In the `init()` function, after `migrateCmd.AddCommand(migrateStatusCmd)`:

```go
migrateCmd.AddCommand(migratePublishCmd)
```

And after the `migrateReportCmd` flags block, add:

```go
migratePublishCmd.Flags().StringVar(&migRepoPath, "repo", "", "path to local VIP repo checkout")
_ = migratePublishCmd.MarkFlagRequired("repo")
```

- [ ] **Step 3: Build to verify compilation**

```bash
PATH="/opt/homebrew/bin:$PATH" make build 2>&1
```

Expected: build succeeds with no errors.

- [ ] **Step 4: Verify command appears in help**

```bash
./stax migrate --help 2>&1 | grep publish
```

Expected: `publish     Upload migration report to S3 and commit to VIP repo`

- [ ] **Step 5: Run full test suite**

```bash
PATH="/opt/homebrew/bin:$PATH" make test
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add cmd/migrate.go
git commit -m "feat(migrate): add stax migrate publish command"
```

---

## Task 7: Update `docs/runbooks/migration.md`

**Files:**
- Modify: `docs/runbooks/migration.md`

- [ ] **Step 1: Add Steps 8–9 and WPEngine considerations to the runbook**

At the end of `docs/runbooks/migration.md`, after the "What not to do" section, add:

```markdown
## Step 8: Review and annotate the report

Open `.stax/<install>-migration-report.md` and fill in the **Operator Notes** section:

- Which incompatible plugins were addressed and how (updated, deactivated, or accepted as-is)
- Any site-specific issues encountered during the migration steps
- Confirmation of the WPEngine MU plugin removal list

Don't skip this step. The report goes to the VIP repo and is the permanent migration record.

## Step 9: Publish

```bash
stax migrate publish --repo=../my-vip-repo
```

Uploads the report and SQL export to S3, copies the report to `<vip-repo>/docs/migration-report.md`, commits, and pushes.

## WPEngine-specific considerations

These apply to every site and are auto-flagged in the report.

**WPEngine MU plugins** — always removed on VIP:

| Plugin | Reason |
|--------|--------|
| `wpe-cache-plugin` | Caching handled by WPVIP infrastructure |
| `wpe-wp-sign-on-plugin` | WPEngine-specific — VIP incompatible |
| `wpe-update-source-selector` | WPEngine-specific — VIP incompatible |
| `force-strong-passwords` | Functionality managed by WPVIP |
| `slt-force-strong-passwords` | WPEngine-specific — VIP incompatible |
| `wpengine-security-auditor` | WPEngine-specific — VIP incompatible |

**Table prefix** — WPEngine often uses `wp_2_` instead of `wp_`. The report detects this. When it's present, search-replace must cover both the prefix conversion (`wp_2_` → `wp_`) and URL changes.

**PHP files in uploads** — VIP does not allow PHP files as media. The report lists any PHP files found in `wp-content/uploads/`. Remove them before or after migration.

**Third-party domain whitelisting** — services like Google reCAPTCHA, ad networks, and SSO providers need domain whitelisting after DNS cutover. This is not a migration blocker but must be documented in Operator Notes and actioned post-launch.
```

- [ ] **Step 2: Verify the full runbook renders correctly**

```bash
cat docs/runbooks/migration.md | grep -E "^## Step [0-9]"
```

Expected output:
```
## Step 1: Pull files
## Step 2: Export the database
## Step 3: Run the phpcs audit
## Step 4: Compare files
## Step 5: Resolve gaps
## Step 6: Import
## Step 7: Generate report
## Step 8: Review and annotate the report
## Step 9: Publish
```

- [ ] **Step 3: Commit**

```bash
git add docs/runbooks/migration.md
git commit -m "docs(runbooks): add publish step and WPEngine-specific considerations to migration runbook"
```

---

## Self-Review

**Spec coverage:**
- Plugin compat table ✓ (Task 2 — `BuildPluginResults`)
- WPEngine MU plugin removal ✓ (Task 2 — `DetectWPEMUPlugins`)
- Theme compat ✓ (Task 2 — `BuildThemeResults`)
- DB analysis (prefix, collation) ✓ (Task 3 — `AnalyzeSQLExport`)
- Media stats ✓ (Task 3 — `AnalyzeMedia`)
- Known issues / Operator Notes ✓ (Task 3 — report template section 7)
- Summary ✓ (Task 3 — report template section 8)
- `stax migrate publish` ✓ (Tasks 5–6)
- `--repo` rename ✓ (Task 1)
- Runbook Steps 8–9 ✓ (Task 7)

**Type consistency:** `enrichedReportData` defined in Task 3 and used in Task 4's updated `Report()`. `PublishOptions` and `Publish()` defined in Task 5 and called in Task 6.

**No placeholders:** All steps have complete code.
