package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/ui"
)

var reportTmpl = template.Must(template.New("report").Funcs(template.FuncMap{
	"join": strings.Join,
	"sort": func(s []string) []string {
		cp := make([]string, len(s))
		copy(cp, s)
		sort.Strings(cp)
		return cp
	},
}).Parse(reportTemplate))

func requireDestination(cfg *config.Config) error {
	if cfg.Migration.Destination == "" {
		return fmt.Errorf("migration.destination is not set in .stax.yml\n\nAdd it:\n\n  migration:\n    destination: vip\n")
	}
	return nil
}

// Pull downloads wp-content from the source provider (uploads excluded).
func Pull(p provider.Provider, cfg *config.Config, opts PullOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}
	src, err := NewSource(cfg.Provider, p, cfg)
	if err != nil {
		return err
	}
	ui.Info("Pulling files from %s (uploads excluded)...", cfg.Provider)
	return src.PullFiles(opts)
}

// Export dumps the source database with VIP-compatible flags to a local SQL file.
// If opts.OutputPath is empty, it defaults to .stax/<install>-export.sql.
func Export(p provider.Provider, cfg *config.Config, opts ExportOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}
	if opts.OutputPath == "" {
		install := config.ProviderConfigString(cfg.ProviderConfig, "install")
		opts.OutputPath = filepath.Join(".stax", install+"-export.sql")
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	src, err := NewSource(cfg.Provider, p, cfg)
	if err != nil {
		return err
	}
	ui.Info("Exporting database with VIP-compatible flags...")
	if err := src.ExportDatabase(opts); err != nil {
		return err
	}
	ui.Success("Database exported to %s", opts.OutputPath)
	return nil
}

// Audit runs a phpcs compatibility audit against the WordPress-VIP-Go ruleset.
// localPath should be the downloaded wp-content directory.
func Audit(cfg *config.Config, localPath string, opts AuditOptions) (*AuditReport, error) {
	if err := requireDestination(cfg); err != nil {
		return nil, err
	}
	dest, err := NewDestination(cfg.Migration.Destination, "")
	if err != nil {
		return nil, err
	}
	ui.Info("Running phpcs audit (WordPress-VIP-Go ruleset)...")
	return dest.Audit(localPath, opts)
}

// Compare diffs the downloaded WPEngine wp-content against the local VIP repo.
// vipRepoPath must be the root of the VIP repo checkout.
func Compare(cfg *config.Config, localPath, vipRepoPath string) (*CompareResult, error) {
	if err := requireDestination(cfg); err != nil {
		return nil, err
	}
	dest, err := NewDestination(cfg.Migration.Destination, vipRepoPath)
	if err != nil {
		return nil, err
	}
	return dest.CompareFiles(localPath)
}

// Import validates then imports a SQL dump into the VIP destination.
func Import(cfg *config.Config, sqlPath string, opts ImportOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}
	dest, err := NewDestination(cfg.Migration.Destination, "")
	if err != nil {
		return err
	}
	ui.Info("Validating database...")
	if err := dest.ValidateDatabase(sqlPath); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	ui.Info("Importing database...")
	return dest.ImportDatabase(sqlPath, opts)
}

// ReportOptions configures stax migrate report.
type ReportOptions struct {
	LocalPath  string // path to downloaded wp-content
	RepoPath   string // path to VIP repo checkout
	SQLPath    string // path to exported SQL file (optional)
	OutputPath string // output markdown path; defaults to .stax/migration-report.md
}

// Report runs audit and compare, then writes a combined markdown report.
func Report(p provider.Provider, cfg *config.Config, opts ReportOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}

	install := config.ProviderConfigString(cfg.ProviderConfig, "install")

	if opts.OutputPath == "" {
		opts.OutputPath = ".stax/migration-report.md"
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

	type reportData struct {
		Install       string
		Destination   string
		GeneratedAt   string
		AuditReport   *AuditReport
		CompareResult *CompareResult
		SQLPath       string
		SQLSizeHuman  string
	}

	data := reportData{
		Install:       install,
		Destination:   cfg.Migration.Destination,
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05 MST"),
		AuditReport:   audit,
		CompareResult: compare,
		SQLPath:       opts.SQLPath,
		SQLSizeHuman:  humanizeBytes(sqlSize),
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

// Status prints the current migration configuration and the presence of key artifacts.
func Status(cfg *config.Config) error {
	install := config.ProviderConfigString(cfg.ProviderConfig, "install")

	ui.Section("Migration Status")
	ui.Info("Install:     %s", install)
	ui.Info("Source:      %s", cfg.Provider)
	ui.Info("Destination: %s", cfg.Migration.Destination)

	checks := []struct {
		label string
		path  string
	}{
		{"Files pulled", "wp-content"},
		{"Database exported", filepath.Join(".stax", install+"-export.sql")},
		{"Report generated", ".stax/migration-report.md"},
	}

	ui.Section("Artifacts")
	for _, c := range checks {
		if _, err := os.Stat(c.path); err == nil {
			ui.Success("  [x] %s (%s)", c.label, c.path)
		} else {
			ui.Info("  [ ] %s", c.label)
		}
	}
	return nil
}

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

const reportTemplate = `# Migration Report: {{ .Install }}

**Generated:** {{ .GeneratedAt }}
**Source:** wpengine/{{ .Install }}
**Destination:** {{ .Destination }}

---

## phpcs Audit (WordPress-VIP-Go)

**Total errors:** {{ .AuditReport.TotalErrors }}
**Total warnings:** {{ .AuditReport.TotalWarnings }}

{{ range .AuditReport.Files -}}
### {{ .FilePath }}

Errors: {{ .Errors }} | Warnings: {{ .Warnings }}

{{ range .Messages -}}
- Line {{ .Line }}, Col {{ .Column }} [{{ .Type }}] {{ .Message }} ({{ .Source }})
{{ end }}
{{ end }}

---

## File Comparison

### Present in WPEngine, missing from VIP repo

{{ if .CompareResult.MissingFromVIP -}}
{{ range (sort .CompareResult.MissingFromVIP) -}}
- {{ . }}
{{ end }}
{{- else }}
None.
{{ end }}

### Present in VIP repo, missing from WPEngine

{{ if .CompareResult.MissingFromWPE -}}
{{ range (sort .CompareResult.MissingFromWPE) -}}
- {{ . }}
{{ end }}
{{- else }}
None.
{{ end }}

---

## Database Export

{{ if .SQLPath -}}
File: {{ .SQLPath }}
Size: {{ .SQLSizeHuman }}
{{- else }}
No SQL export found. Run ` + "`stax migrate export`" + ` first.
{{ end }}
`
