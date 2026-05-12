package migration

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/ui"
)

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

	if err := enrichedReportTmpl.Execute(f, data); err != nil {
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

	docsDir := filepath.Join(opts.RepoPath, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs/ in VIP repo: %w", err)
	}
	if err := copyFile(opts.ReportPath, filepath.Join(docsDir, "migration-report.md")); err != nil {
		return fmt.Errorf("failed to copy report to VIP repo: %w", err)
	}

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
