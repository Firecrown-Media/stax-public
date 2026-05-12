package migration_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/migration"
	"github.com/firecrown-media/stax/pkg/provider"
)

type mockSrc2 struct{ pullErr, exportErr error }

func (m *mockSrc2) PullFiles(_ migration.PullOptions) error        { return m.pullErr }
func (m *mockSrc2) ExportDatabase(_ migration.ExportOptions) error { return m.exportErr }

type mockDst2 struct{ auditErr, validateErr, importErr error }

func (m *mockDst2) Audit(_ string, _ migration.AuditOptions) (*migration.AuditReport, error) {
	if m.auditErr != nil {
		return nil, m.auditErr
	}
	return &migration.AuditReport{}, nil
}
func (m *mockDst2) ValidateDatabase(_ string) error                          { return m.validateErr }
func (m *mockDst2) ImportDatabase(_ string, _ migration.ImportOptions) error { return m.importErr }
func (m *mockDst2) ImportMedia(_ migration.ImportOptions) error              { return nil }
func (m *mockDst2) CompareFiles(_ string) (*migration.CompareResult, error) {
	return &migration.CompareResult{}, nil
}

func cfgWithDest(dest string) *config.Config {
	return &config.Config{
		Provider: "test-svc",
		ProviderConfig: map[string]any{
			"install": "mysite",
		},
		Migration: config.MigrationConfig{Destination: dest},
	}
}

func registerTestProviders(t *testing.T) {
	t.Helper()
	t.Cleanup(migration.ResetRegistryForTesting)
	migration.RegisterSource("test-svc", func(_ provider.Provider, _ *config.Config) migration.Source {
		return &mockSrc2{}
	})
	migration.RegisterDestination("test-dst", func(_ string) migration.Destination {
		return &mockDst2{}
	})
}

func TestRequireDestination_MissingError(t *testing.T) {
	registerTestProviders(t)
	cfg := cfgWithDest("")
	err := migration.Pull(nil, cfg, migration.PullOptions{})
	if err == nil {
		t.Fatal("expected error for missing migration.destination")
	}
	if !strings.Contains(err.Error(), "migration.destination") {
		t.Errorf("error should mention migration.destination, got: %v", err)
	}
}

func TestAudit_MissingDestination(t *testing.T) {
	registerTestProviders(t)
	cfg := cfgWithDest("")
	_, err := migration.Audit(cfg, "/some/path", migration.AuditOptions{})
	if err == nil {
		t.Fatal("expected error for missing migration.destination")
	}
}

func TestExport_MissingDestination(t *testing.T) {
	registerTestProviders(t)
	cfg := cfgWithDest("")
	err := migration.Export(nil, cfg, migration.ExportOptions{})
	if err == nil {
		t.Fatal("expected error for missing migration.destination")
	}
}

func TestImport_Success(t *testing.T) {
	registerTestProviders(t)
	cfg := cfgWithDest("test-dst")
	if err := migration.Import(cfg, "/some/file.sql", migration.ImportOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImport_ValidateError(t *testing.T) {
	t.Cleanup(migration.ResetRegistryForTesting)
	migration.RegisterDestination("fail-validate", func(_ string) migration.Destination {
		return &mockDst2{validateErr: errors.New("SQL invalid")}
	})

	cfg := cfgWithDest("fail-validate")
	err := migration.Import(cfg, "/some/file.sql", migration.ImportOptions{})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReport_EnrichedSections(t *testing.T) {
	registerTestProviders(t)

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

func TestReportOptions_RepoPath(t *testing.T) {
	opts := migration.ReportOptions{RepoPath: "/some/repo"}
	if opts.RepoPath != "/some/repo" {
		t.Errorf("expected RepoPath, got empty")
	}
}
