package wpengine

import (
	"fmt"
	"io"
	"os"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/files"
	"github.com/firecrown-media/stax/pkg/migration"
	"github.com/firecrown-media/stax/pkg/provider"
)

func init() {
	migration.RegisterSource("wpengine", func(p provider.Provider, cfg *config.Config) migration.Source {
		return NewWPEngineSource(p, cfg)
	})
}

// WPEngineSource implements migration.Source for WPEngine installs.
type WPEngineSource struct {
	provider provider.Provider
	cfg      *config.Config
}

// NewWPEngineSource constructs a WPEngineSource.
func NewWPEngineSource(p provider.Provider, cfg *config.Config) *WPEngineSource {
	return &WPEngineSource{provider: p, cfg: cfg}
}

// PullFiles downloads wp-content from WPEngine, always excluding uploads.
func (s *WPEngineSource) PullFiles(opts migration.PullOptions) error {
	return files.Pull(s.provider, s.cfg, files.SyncFlags{
		ExcludeUploads: true,
		ThemesOnly:     opts.ThemesOnly,
		PluginsOnly:    opts.PluginsOnly,
		MuPluginsOnly:  opts.MuPluginsOnly,
		DryRun:         opts.DryRun,
		ProjectDir:     opts.ProjectDir,
	})
}

// ExportDatabase exports the WPEngine database with VIP-compatible flags
// and writes the SQL dump to opts.OutputPath.
func (s *WPEngineSource) ExportDatabase(opts migration.ExportOptions) error {
	install := config.ProviderConfigString(s.cfg.ProviderConfig, "install")
	site := &provider.Site{Name: install, Provider: s.cfg.Provider}

	reader, err := s.provider.ExportDatabase(site, provider.DatabaseExportOptions{
		ExtraFlags: []string{
			"--hex-blob",
			"--quote-names",
			"--default-character-set=utf8mb4",
		},
	})
	if err != nil {
		return fmt.Errorf("database export failed: %w", err)
	}
	defer reader.Close()

	f, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", opts.OutputPath, err)
	}

	if _, err := io.Copy(f, reader); err != nil {
		f.Close()
		return fmt.Errorf("failed to write database export: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to finalize output file: %w", err)
	}
	return nil
}
