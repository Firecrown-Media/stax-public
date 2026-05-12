# Migration Checklist Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `stax migrate checklist` — a command that generates a per-site migration checklist pre-populated from config and existing artifacts, serving as both an execution plan and a post-migration record.

**Architecture:** A new `pkg/migration/checklist.go` file holds the `ChecklistData` struct, Go template, and `WriteChecklist()` helper, following the same pattern as `report.go`. `service.go` gets `Checklist()` (detects artifacts, populates data, calls `WriteChecklist`) and an updated `Publish()` that includes the checklist alongside the report. `cmd/migrate.go` gets a new `migrateChecklistCmd` and the publish command is updated to pass the checklist path.

**Tech Stack:** Go, `text/template`, `os/exec` (git log for VIP commit SHA), Cobra flags.

---

## File Structure

| File | Change | Responsibility |
|------|--------|----------------|
| `pkg/migration/checklist.go` | CREATE | `ChecklistData` struct, markdown template, `WriteChecklist()` |
| `pkg/migration/checklist_test.go` | CREATE | Template output tests — section presence, pre-check logic |
| `pkg/migration/service.go` | MODIFY | Add `ChecklistOptions`, `Checklist()`; update `PublishOptions`, `Publish()` |
| `pkg/migration/service_test.go` | MODIFY | `TestChecklist_*` integration tests |
| `cmd/migrate.go` | MODIFY | Add `migrateChecklistCmd`; update `migratePublishCmd` |
| `docs/runbooks/migration.md` | MODIFY | Insert Step 8 (checklist), renumber Steps 9–10 |

---

## Task 1: Create `pkg/migration/checklist.go`

**Files:**
- Create: `pkg/migration/checklist.go`

- [ ] **Step 1: Write the failing test in `pkg/migration/checklist_test.go`**

```go
package migration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/firecrown-media/stax/pkg/migration"
)

func TestWriteChecklist_AllSections(t *testing.T) {
	data := migration.ChecklistData{
		Install:      "mysite",
		Destination:  "vip",
		Domain:       "mysite.com",
		GeneratedAt:  "2026-05-12 10:00:00 CDT",
		ReportPath:   ".stax/mysite-migration-report.md",
		ReportExists: false,
		SQLPath:      ".stax/mysite-export.sql",
		SQLExists:    false,
		VIPCommitSHA: "",
		PullDone:     false,
		ExportDone:   false,
		ReportDone:   false,
		PublishDone:  false,
	}
	outputPath := filepath.Join(t.TempDir(), "checklist.md")
	if err := migration.WriteChecklist(data, outputPath); err != nil {
		t.Fatalf("WriteChecklist() failed: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read checklist: %v", err)
	}
	s := string(content)
	for _, want := range []string{
		"Migration Checklist: mysite",
		"mysite.com",
		"Pre-migration steps",
		"QA Checklist",
		"DNS Cutover",
		"Post-launch Validation",
		"Sign-off",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("checklist missing section %q", want)
		}
	}
}

func TestWriteChecklist_PreCheckedItems(t *testing.T) {
	data := migration.ChecklistData{
		Install:      "mysite",
		Destination:  "vip",
		Domain:       "mysite.com",
		GeneratedAt:  "2026-05-12 10:00:00 CDT",
		ReportPath:   ".stax/mysite-migration-report.md",
		ReportExists: true,
		SQLPath:      ".stax/mysite-export.sql",
		SQLExists:    true,
		VIPCommitSHA: "abc1234",
		PullDone:     true,
		ExportDone:   true,
		ReportDone:   true,
		PublishDone:  true,
	}
	outputPath := filepath.Join(t.TempDir(), "checklist.md")
	if err := migration.WriteChecklist(data, outputPath); err != nil {
		t.Fatalf("WriteChecklist() failed: %v", err)
	}
	content, _ := os.ReadFile(outputPath)
	s := string(content)

	checks := []struct {
		substr string
		want   bool
	}{
		{"[x] `stax migrate pull`", true},
		{"[x] `stax migrate export`", true},
		{"[x] `stax migrate report`", true},
		{"[x] `stax migrate publish`", true},
		{"[ ] `stax migrate import`", true}, // import is never pre-checked
		{"abc1234", true},
	}
	for _, c := range checks {
		got := strings.Contains(s, c.substr)
		if got != c.want {
			t.Errorf("expected Contains(%q)=%v, got %v", c.substr, c.want, got)
		}
	}
}

func TestWriteChecklist_UncheckedWhenNoArtifacts(t *testing.T) {
	data := migration.ChecklistData{
		Install:     "mysite",
		Destination: "vip",
		Domain:      "mysite.com",
		GeneratedAt: "2026-05-12 10:00:00 CDT",
	}
	outputPath := filepath.Join(t.TempDir(), "checklist.md")
	if err := migration.WriteChecklist(data, outputPath); err != nil {
		t.Fatalf("WriteChecklist() failed: %v", err)
	}
	content, _ := os.ReadFile(outputPath)
	s := string(content)

	for _, step := range []string{"pull", "export", "audit", "compare", "import", "report", "publish"} {
		substr := "[ ] `stax migrate " + step + "`"
		if !strings.Contains(s, substr) {
			t.Errorf("step %q should be unchecked when no artifacts, missing %q", step, substr)
		}
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd /Users/geoff/_projects/fc/stax
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestWriteChecklist -v 2>&1 | head -20
```

Expected: FAIL — `migration.ChecklistData` undefined, `migration.WriteChecklist` undefined.

- [ ] **Step 3: Create `pkg/migration/checklist.go`**

```go
package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ChecklistData holds all data passed to the migration checklist template.
type ChecklistData struct {
	Install      string
	Destination  string
	Domain       string
	GeneratedAt  string
	ReportPath   string
	ReportExists bool
	SQLPath      string
	SQLExists    bool
	VIPCommitSHA string // empty if not yet published to VIP repo
	PullDone     bool   // true if wp-content/ dir detected
	ExportDone   bool   // true if SQL dump detected
	ReportDone   bool   // true if migration report detected
	PublishDone  bool   // true if VIP commit SHA found
}

var checklistTmpl = template.Must(template.New("checklist").Funcs(template.FuncMap{
	"check": func(done bool) string {
		if done {
			return "x"
		}
		return " "
	},
}).Parse(checklistTemplate))

// WriteChecklist writes the checklist markdown to outputPath, creating parent dirs as needed.
func WriteChecklist(data ChecklistData, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create checklist file: %w", err)
	}
	defer f.Close()
	return checklistTmpl.Execute(f, data)
}

const checklistTemplate = `# Migration Checklist: {{ .Install }}

**Site:** {{ .Install }}
**Destination:** {{ .Destination }}
**Domain:** {{ .Domain }}
**Generated:** {{ .GeneratedAt }}

---

## Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Migration report | {{ .ReportPath }} | {{ if .ReportExists }}Present{{ else }}Not found{{ end }} |
| SQL dump | {{ .SQLPath }} | {{ if .SQLExists }}Present{{ else }}Not found{{ end }} |
| VIP repo commit | {{ if .VIPCommitSHA }}{{ .VIPCommitSHA }}{{ else }}—{{ end }} | {{ if .VIPCommitSHA }}Present{{ else }}Not found{{ end }} |

---

## Pre-migration steps

- [{{ check .PullDone }}] ` + "`stax migrate pull`" + ` — files downloaded to ` + "`wp-content/`" + `
- [{{ check .ExportDone }}] ` + "`stax migrate export`" + ` — SQL dump at ` + "`{{ .SQLPath }}`" + `
- [{{ check .ReportDone }}] ` + "`stax migrate audit`" + ` — phpcs results in migration report
- [{{ check .ReportDone }}] ` + "`stax migrate compare`" + ` — file comparison in migration report
- [ ] ` + "`stax migrate import`" + ` — database imported to VIP
- [{{ check .ReportDone }}] ` + "`stax migrate report`" + ` — report at ` + "`{{ .ReportPath }}`" + `
- [{{ check .PublishDone }}] ` + "`stax migrate publish`" + ` — artifacts uploaded to S3, committed to VIP repo{{ if .VIPCommitSHA }} ({{ .VIPCommitSHA }}){{ end }}

---

## QA Checklist

Run on the VIP staging environment before DNS cutover.

- [ ] Front page loads
- [ ] Key pages load (about, contact, category pages)
- [ ] Forms work
- [ ] No broken images

---

## DNS Cutover

- [ ] Reduce TTL on ` + "`{{ .Domain }}`" + ` to 300s (do this 24–48h before cutover)
- [ ] Confirm TTL reduction has propagated
- [ ] Swap A/CNAME record for ` + "`{{ .Domain }}`" + ` to VIP
- [ ] Verify DNS resolution has updated

---

## Post-launch Validation

Run on ` + "`{{ .Domain }}`" + ` after DNS cutover.

- [ ] Front page loads on ` + "`{{ .Domain }}`" + `
- [ ] Key pages load on ` + "`{{ .Domain }}`" + `
- [ ] Forms work on ` + "`{{ .Domain }}`" + `
- [ ] No broken images on ` + "`{{ .Domain }}`" + `

---

## Sign-off

**Operator:** ___________________________
**Date:** ___________________________
`
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestWriteChecklist -v
```

Expected: PASS — all 3 `TestWriteChecklist_*` tests pass.

- [ ] **Step 5: Commit**

```bash
git add pkg/migration/checklist.go pkg/migration/checklist_test.go
git commit -m "feat: add ChecklistData type, template, and WriteChecklist helper"
```

---

## Task 2: Add `Checklist()` to `pkg/migration/service.go`

**Files:**
- Modify: `pkg/migration/service.go`
- Modify: `pkg/migration/service_test.go`

**Context:** `service.go` already imports `fmt`, `io`, `os`, `os/exec`, `path/filepath`, `strings`, `time` and the `config`, `provider`, `ui` packages. `config.ProviderConfigString` extracts string values from `cfg.ProviderConfig`. The `requireDestination(cfg)` helper returns an error if `cfg.Migration.Destination` is empty.

- [ ] **Step 1: Write failing tests in `pkg/migration/service_test.go`**

Add these three tests after the existing `TestReportOptions_RepoPath` test:

```go
func TestChecklist_MissingDestination(t *testing.T) {
	cfg := cfgWithDest("")
	err := migration.Checklist(cfg, migration.ChecklistOptions{Domain: "example.com"})
	if err == nil {
		t.Fatal("expected error for missing migration.destination")
	}
	if !strings.Contains(err.Error(), "migration.destination") {
		t.Errorf("error should mention migration.destination, got: %v", err)
	}
}

func TestChecklist_GeneratesOutput(t *testing.T) {
	registerTestProviders(t)
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "checklist.md")

	cfg := cfgWithDest("test-dst")
	err := migration.Checklist(cfg, migration.ChecklistOptions{
		Domain:     "example.com",
		OutputPath: outputPath,
		ProjectDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("Checklist() failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read checklist: %v", err)
	}
	s := string(content)
	for _, want := range []string{
		"Migration Checklist: mysite",
		"example.com",
		"Pre-migration steps",
		"QA Checklist",
		"DNS Cutover",
		"Post-launch Validation",
		"Sign-off",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("checklist missing %q", want)
		}
	}
}

func TestChecklist_PreChecksArtifacts(t *testing.T) {
	registerTestProviders(t)
	tmpDir := t.TempDir()

	// Create artifacts: report, SQL, wp-content
	staxDir := filepath.Join(tmpDir, ".stax-artifacts")
	_ = os.MkdirAll(staxDir, 0755)
	reportPath := filepath.Join(staxDir, "mysite-migration-report.md")
	sqlPath := filepath.Join(staxDir, "mysite-export.sql")
	_ = os.WriteFile(reportPath, []byte("# Report\n"), 0644)
	_ = os.WriteFile(sqlPath, []byte("-- SQL\n"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "wp-content"), 0755)

	outputPath := filepath.Join(tmpDir, "checklist.md")
	cfg := cfgWithDest("test-dst")
	err := migration.Checklist(cfg, migration.ChecklistOptions{
		Domain:      "example.com",
		OutputPath:  outputPath,
		ProjectDir:  tmpDir,
		ReportPath:  reportPath,
		SQLPath:     sqlPath,
	})
	if err != nil {
		t.Fatalf("Checklist() failed: %v", err)
	}

	content, _ := os.ReadFile(outputPath)
	s := string(content)

	if !strings.Contains(s, "[x] `stax migrate pull`") {
		t.Error("pull step should be pre-checked (wp-content dir exists)")
	}
	if !strings.Contains(s, "[x] `stax migrate export`") {
		t.Error("export step should be pre-checked (SQL file exists)")
	}
	if !strings.Contains(s, "[x] `stax migrate report`") {
		t.Error("report step should be pre-checked (report file exists)")
	}
	if !strings.Contains(s, "[ ] `stax migrate import`") {
		t.Error("import step should never be pre-checked")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run "TestChecklist_" -v 2>&1 | head -20
```

Expected: FAIL — `migration.Checklist` undefined, `migration.ChecklistOptions` undefined.

- [ ] **Step 3: Add `ChecklistOptions` and `Checklist()` to `pkg/migration/service.go`**

Add after the `Status()` function (around line 205) and before `const defaultAssetsBucket`:

```go
// ChecklistOptions configures stax migrate checklist.
type ChecklistOptions struct {
	Domain     string // required: live domain, e.g. astronomytn.com
	RepoPath   string // optional: path to VIP repo for commit SHA detection
	OutputPath string // defaults to .stax/<install>-checklist.md
	ProjectDir string // for wp-content/ detection
	ReportPath string // defaults to .stax/<install>-migration-report.md
	SQLPath    string // defaults to .stax/<install>-export.sql
}

// Checklist generates a per-site migration checklist pre-populated from config
// and existing artifacts.
func Checklist(cfg *config.Config, opts ChecklistOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}
	install := config.ProviderConfigString(cfg.ProviderConfig, "install")

	if opts.OutputPath == "" {
		opts.OutputPath = filepath.Join(".stax", install+"-checklist.md")
	}
	if opts.ReportPath == "" {
		opts.ReportPath = filepath.Join(".stax", install+"-migration-report.md")
	}
	if opts.SQLPath == "" {
		opts.SQLPath = filepath.Join(".stax", install+"-export.sql")
	}

	_, reportErr := os.Stat(opts.ReportPath)
	_, sqlErr := os.Stat(opts.SQLPath)

	pullDone := false
	if opts.ProjectDir != "" {
		if _, err := os.Stat(filepath.Join(opts.ProjectDir, "wp-content")); err == nil {
			pullDone = true
		}
	}

	var vipCommitSHA string
	if opts.RepoPath != "" {
		out, err := exec.Command("git", "-C", opts.RepoPath, "log", "--oneline", "-1", "--", "docs/migration-report.md").Output()
		if err == nil {
			parts := strings.Fields(string(out))
			if len(parts) > 0 {
				vipCommitSHA = strings.TrimSpace(parts[0])
			}
		}
	}

	data := ChecklistData{
		Install:      install,
		Destination:  cfg.Migration.Destination,
		Domain:       opts.Domain,
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05 MST"),
		ReportPath:   opts.ReportPath,
		ReportExists: reportErr == nil,
		SQLPath:      opts.SQLPath,
		SQLExists:    sqlErr == nil,
		VIPCommitSHA: vipCommitSHA,
		PullDone:     pullDone,
		ExportDone:   sqlErr == nil,
		ReportDone:   reportErr == nil,
		PublishDone:  vipCommitSHA != "",
	}

	if err := WriteChecklist(data, opts.OutputPath); err != nil {
		return err
	}
	ui.Success("Checklist written to %s", opts.OutputPath)
	return nil
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run "TestChecklist_" -v
```

Expected: PASS — all three tests pass.

- [ ] **Step 5: Run full migration package tests to check for regressions**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -v 2>&1 | tail -20
```

Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/migration/service.go pkg/migration/service_test.go
git commit -m "feat: add Checklist() service function with artifact detection"
```

---

## Task 3: Update `Publish()` to include the checklist

**Files:**
- Modify: `pkg/migration/service.go` (lines 209–282)

**Context:** `Publish()` currently uploads the report and SQL to S3, copies the report to `<repo>/docs/migration-report.md`, and commits both in a single `git commit`. The checklist should be included in that same commit when it exists.

Current `PublishOptions` (line 210):
```go
type PublishOptions struct {
    RepoPath   string
    ReportPath string
    SQLPath    string
}
```

- [ ] **Step 1: Add `ChecklistPath` to `PublishOptions` and update `Publish()`**

Replace `PublishOptions` and the `Publish()` function in `service.go`. The full updated function:

```go
// PublishOptions configures stax migrate publish.
type PublishOptions struct {
	RepoPath      string // path to local VIP repo checkout (required)
	ReportPath    string // path to migration-report.md; defaults to .stax/<install>-migration-report.md
	SQLPath       string // path to SQL export; defaults to .stax/<install>-export.sql
	ChecklistPath string // path to checklist; defaults to .stax/<install>-checklist.md
}

// Publish uploads the migration report, SQL export, and checklist to S3, then commits
// the report and checklist to the VIP repo docs/ folder and pushes.
func Publish(cfg *config.Config, opts PublishOptions) error {
	install := config.ProviderConfigString(cfg.ProviderConfig, "install")

	if opts.ReportPath == "" {
		opts.ReportPath = filepath.Join(".stax", install+"-migration-report.md")
	}
	if opts.SQLPath == "" {
		opts.SQLPath = filepath.Join(".stax", install+"-export.sql")
	}
	if opts.ChecklistPath == "" {
		opts.ChecklistPath = filepath.Join(".stax", install+"-checklist.md")
	}

	if _, err := os.Stat(opts.ReportPath); err != nil {
		return fmt.Errorf("report not found at %s: run 'stax migrate report' first", opts.ReportPath)
	}

	if _, err := os.Stat(opts.RepoPath); err != nil {
		return fmt.Errorf("VIP repo not found at %s", opts.RepoPath)
	}

	s3Prefix := fmt.Sprintf("s3://%s/vip-migration/%s", defaultAssetsBucket, install)

	ui.Info("Uploading report to S3...")
	if err := runExec("aws", "s3", "cp", opts.ReportPath, s3Prefix+"/migration-report.md"); err != nil {
		return fmt.Errorf("S3 upload of report failed: %w", err)
	}

	if _, err := os.Stat(opts.SQLPath); err == nil {
		ui.Info("Uploading SQL export to S3...")
		s3SQL := fmt.Sprintf("%s/%s-export.sql", s3Prefix, install)
		if err := runExec("aws", "s3", "cp", opts.SQLPath, s3SQL); err != nil {
			return fmt.Errorf("S3 upload of SQL failed: %w", err)
		}
	}

	checklistExists := false
	if _, err := os.Stat(opts.ChecklistPath); err == nil {
		checklistExists = true
		ui.Info("Uploading checklist to S3...")
		if err := runExec("aws", "s3", "cp", opts.ChecklistPath, s3Prefix+"/checklist.md"); err != nil {
			return fmt.Errorf("S3 upload of checklist failed: %w", err)
		}
	}

	docsDir := filepath.Join(opts.RepoPath, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs/ in VIP repo: %w", err)
	}
	if err := copyFile(opts.ReportPath, filepath.Join(docsDir, "migration-report.md")); err != nil {
		return fmt.Errorf("failed to copy report to VIP repo: %w", err)
	}
	if checklistExists {
		if err := copyFile(opts.ChecklistPath, filepath.Join(docsDir, "checklist.md")); err != nil {
			return fmt.Errorf("failed to copy checklist to VIP repo: %w", err)
		}
	}

	ui.Info("Committing to VIP repo...")
	if err := runExec("git", "-C", opts.RepoPath, "add", "docs/migration-report.md"); err != nil {
		return fmt.Errorf("git add (report) failed: %w", err)
	}
	if checklistExists {
		if err := runExec("git", "-C", opts.RepoPath, "add", "docs/checklist.md"); err != nil {
			return fmt.Errorf("git add (checklist) failed: %w", err)
		}
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
	ui.Info("  S3 report:    %s/migration-report.md", s3Prefix)
	ui.Info("  S3 SQL:       %s/%s-export.sql", s3Prefix, install)
	if checklistExists {
		ui.Info("  S3 checklist: %s/checklist.md", s3Prefix)
	}
	ui.Info("  VIP repo:     %s/docs/ (commit %s)", opts.RepoPath, sha)
	return nil
}
```

- [ ] **Step 2: Run existing publish tests to confirm they still pass**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run "TestPublish_" -v
```

Expected: `TestPublish_MissingReport` and `TestPublish_MissingRepo` both PASS (they fail before reaching S3/checklist logic).

- [ ] **Step 3: Run full migration package tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -v 2>&1 | tail -20
```

Expected: all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add pkg/migration/service.go
git commit -m "feat: include checklist in stax migrate publish"
```

---

## Task 4: Add `migrateChecklistCmd` to `cmd/migrate.go`

**Files:**
- Modify: `cmd/migrate.go`

**Context:** `cmd/migrate.go` declares package-level vars for all flag values (`migRepoPath`, `migOutputPath`, etc.) and registers commands in `init()`. All commands call `loadConfigForCommand()` then the corresponding `migration.*()` function. `getProjectDir()` returns the absolute path to the working directory (defined in `cmd/root.go`).

Current var block (lines 15–29):
```go
var (
    migDestination   string
    migLocalPath     string
    migRepoPath      string
    migExportPath    string
    migSQLPath       string
    migOutputPath    string
    migPullDryRun    bool
    migImportDryRun  bool
    migSlug          string
    migThemesOnly    bool
    migPluginsOnly   bool
    migMuPluginsOnly bool
    migSeverity      int
)
```

- [ ] **Step 1: Add `migDomain` to the var block**

In `cmd/migrate.go`, add `migDomain string` to the existing var block:

```go
var (
	migDestination   string
	migLocalPath     string
	migRepoPath      string
	migExportPath    string
	migSQLPath       string
	migOutputPath    string
	migPullDryRun    bool
	migImportDryRun  bool
	migSlug          string
	migThemesOnly    bool
	migPluginsOnly   bool
	migMuPluginsOnly bool
	migSeverity      int
	migDomain        string
)
```

- [ ] **Step 2: Add `migrateChecklistCmd` after `migratePublishCmd`**

Add this command definition after `migratePublishCmd` (around line 249):

```go
var migrateChecklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "Generate per-site migration checklist with pre-populated artifacts",
	Long: `Generate a migration checklist pre-populated from config and existing artifacts.
The checklist tracks pre-migration steps, QA, DNS cutover, and post-launch validation.
Re-run at any time to update the artifact status.`,
	Example: `  stax migrate checklist --domain=astronomytn.com
  stax migrate checklist --domain=astronomytn.com --repo=../vip-repo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migDestination != "" {
			cfg.Migration.Destination = migDestination
		}
		return migration.Checklist(cfg, migration.ChecklistOptions{
			Domain:     migDomain,
			RepoPath:   migRepoPath,
			OutputPath: migOutputPath,
			ProjectDir: getProjectDir(),
		})
	},
}
```

- [ ] **Step 3: Register the command and its flags in `init()`**

In the `init()` function, after the `migratePublishCmd` registration block, add:

```go
migrateCmd.AddCommand(migrateChecklistCmd)
migrateChecklistCmd.Flags().StringVar(&migDomain, "domain", "", "live domain (e.g. astronomytn.com)")
_ = migrateChecklistCmd.MarkFlagRequired("domain")
migrateChecklistCmd.Flags().StringVar(&migRepoPath, "repo", "", "path to local VIP repo checkout (for commit SHA detection)")
migrateChecklistCmd.Flags().StringVar(&migOutputPath, "output", "", "output path (default: .stax/<install>-checklist.md)")
```

- [ ] **Step 4: Update `migratePublishCmd`'s `RunE` to pass `ChecklistPath`**

The current publish `RunE` (around line 234) calls `migration.Publish` with three fields. Update it to pass `ChecklistPath`:

```go
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
		RepoPath:      migRepoPath,
		ReportPath:    filepath.Join(".stax", install+"-migration-report.md"),
		SQLPath:       filepath.Join(".stax", install+"-export.sql"),
		ChecklistPath: filepath.Join(".stax", install+"-checklist.md"),
	})
},
```

- [ ] **Step 5: Build to verify no compile errors**

```bash
PATH="/opt/homebrew/bin:$PATH" make build 2>&1
```

Expected: build succeeds with no errors.

- [ ] **Step 6: Verify checklist command appears in help**

```bash
./.bin/stax migrate --help 2>&1 | grep checklist
```

Expected: output contains `checklist`.

- [ ] **Step 7: Run full test suite**

```bash
PATH="/opt/homebrew/bin:$PATH" make test 2>&1 | tail -20
```

Expected: all tests PASS.

- [ ] **Step 8: Commit**

```bash
git add cmd/migrate.go
git commit -m "feat: add stax migrate checklist command"
```

---

## Task 5: Update `docs/runbooks/migration.md`

**Files:**
- Modify: `docs/runbooks/migration.md`

**Context:** The current runbook has Steps 1–9. Step 7 is "Generate report", Step 8 is "Review and annotate the report", Step 9 is "Publish". The checklist should be inserted as the new Step 8 (after the report is generated, before the operator annotates it) — it gives the operator a structured document to work through during annotation and QA. Steps 8 and 9 become Steps 9 and 10.

- [ ] **Step 1: Insert new Step 8 and renumber Steps 9–10**

After the "## Step 7: Generate report" section and before "## Check status at any time" (the second occurrence), insert:

```markdown
## Step 8: Generate checklist

```bash
stax migrate checklist --domain=<live-domain>
```

Generates `.stax/<install>-checklist.md` — a per-site migration checklist pre-populated with artifact status and site-specific details. The checklist tracks all migration steps, QA sign-off, DNS cutover, and post-launch validation.

Re-run after any step completes to update the pre-checked items. If you have a VIP repo checkout, pass `--repo` to detect the publish commit SHA:

```bash
stax migrate checklist --domain=<live-domain> --repo=../my-vip-repo
```
```

Then rename the existing `## Step 8: Review and annotate the report` to `## Step 9: Review and annotate the report`, and rename `## Step 9: Publish` to `## Step 10: Publish`.

- [ ] **Step 2: Verify the file renders correctly (section count check)**

```bash
grep "^## Step" docs/runbooks/migration.md
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
## Step 8: Generate checklist
## Step 9: Review and annotate the report
## Step 10: Publish
```

- [ ] **Step 3: Run gofmt to ensure no formatting issues were introduced**

```bash
PATH="/opt/homebrew/bin:$PATH" make fmt
git diff --name-only
```

Expected: no Go files changed (only the markdown file was touched).

- [ ] **Step 4: Commit**

```bash
git add docs/runbooks/migration.md
git commit -m "docs: add checklist step to migration runbook, renumber steps 9-10"
```
