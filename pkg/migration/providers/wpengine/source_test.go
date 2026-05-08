package wpengine_test

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/migration"
	wpe "github.com/firecrown-media/stax/pkg/migration/providers/wpengine"
	"github.com/firecrown-media/stax/pkg/provider"
)

// stubProvider satisfies provider.Provider for tests — only ExportDatabase matters here.
type stubProvider struct {
	exportErr    error
	exportOutput string
}

func (s *stubProvider) Name() string                                          { return "wpengine" }
func (s *stubProvider) Description() string                                   { return "" }
func (s *stubProvider) Capabilities() provider.ProviderCapabilities           { return provider.ProviderCapabilities{} }
func (s *stubProvider) Authenticate(_ map[string]string) error                { return nil }
func (s *stubProvider) TestConnection() error                                 { return nil }
func (s *stubProvider) ValidateCredentials(_ map[string]string) error         { return nil }
func (s *stubProvider) ListSites() ([]provider.Site, error)                   { return nil, nil }
func (s *stubProvider) GetSite(_ string) (*provider.Site, error)              { return nil, nil }
func (s *stubProvider) GetSiteMetadata(_ *provider.Site) (*provider.SiteMetadata, error) {
	return nil, nil
}
func (s *stubProvider) ExportDatabase(_ *provider.Site, _ provider.DatabaseExportOptions) (io.ReadCloser, error) {
	if s.exportErr != nil {
		return nil, s.exportErr
	}
	return io.NopCloser(strings.NewReader(s.exportOutput)), nil
}
func (s *stubProvider) ImportDatabase(_ *provider.Site, _ io.Reader, _ provider.DatabaseImportOptions) error {
	return nil
}
func (s *stubProvider) GetDatabaseCredentials(_ *provider.Site) (*provider.DatabaseCredentials, error) {
	return nil, nil
}
func (s *stubProvider) SyncFiles(_ *provider.Site, _ string, _ provider.SyncOptions) error {
	return nil
}
func (s *stubProvider) DownloadFile(_ *provider.Site, _ string) (io.ReadCloser, error) {
	return nil, nil
}
func (s *stubProvider) UploadFile(_ *provider.Site, _, _ string) error   { return nil }
func (s *stubProvider) GetPHPVersion(_ *provider.Site) (string, error)   { return "", nil }
func (s *stubProvider) GetMySQLVersion(_ *provider.Site) (string, error) { return "", nil }
func (s *stubProvider) GetWordPressVersion(_ *provider.Site) (string, error) {
	return "", nil
}

func minimalCfg() *config.Config {
	return &config.Config{
		Provider: "wpengine",
		ProviderConfig: map[string]any{
			"install":     "mysite",
			"environment": "production",
		},
	}
}

func TestWPEngineSource_ExportDatabase_WritesFile(t *testing.T) {
	p := &stubProvider{exportOutput: "-- SQL DUMP\nCREATE TABLE test;"}
	src := wpe.NewWPEngineSource(p, minimalCfg())

	outPath := t.TempDir() + "/export.sql"
	err := src.ExportDatabase(migration.ExportOptions{OutputPath: outPath})
	if err != nil {
		t.Fatalf("ExportDatabase failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(data), "CREATE TABLE test") {
		t.Errorf("output file missing expected content, got: %s", string(data))
	}
}

func TestWPEngineSource_ExportDatabase_PassesVIPFlags(t *testing.T) {
	var capturedOpts provider.DatabaseExportOptions
	p := &captureProvider{}
	p.captureFunc = func(opts provider.DatabaseExportOptions) {
		capturedOpts = opts
	}
	p.exportOutput = "-- SQL"

	src := wpe.NewWPEngineSource(p, minimalCfg())
	_ = src.ExportDatabase(migration.ExportOptions{OutputPath: t.TempDir() + "/out.sql"})

	for _, flag := range []string{"--hex-blob", "--quote-names", "--default-character-set=utf8mb4"} {
		found := false
		for _, f := range capturedOpts.ExtraFlags {
			if f == flag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected ExtraFlag %q to be passed, got: %v", flag, capturedOpts.ExtraFlags)
		}
	}
}

func TestWPEngineSource_ExportDatabase_ProviderError(t *testing.T) {
	p := &stubProvider{exportErr: errors.New("SSH connection failed")}
	src := wpe.NewWPEngineSource(p, minimalCfg())

	err := src.ExportDatabase(migration.ExportOptions{OutputPath: t.TempDir() + "/out.sql"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "database export failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

// captureProvider extends stubProvider to capture ExportDatabase options.
type captureProvider struct {
	stubProvider
	captureFunc func(provider.DatabaseExportOptions)
}

func (c *captureProvider) ExportDatabase(site *provider.Site, opts provider.DatabaseExportOptions) (io.ReadCloser, error) {
	if c.captureFunc != nil {
		c.captureFunc(opts)
	}
	return io.NopCloser(strings.NewReader(c.exportOutput)), nil
}
