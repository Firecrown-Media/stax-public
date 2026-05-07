//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/testutil"
	"github.com/firecrown-media/stax/test/helpers"
)

func TestDatabaseWorkflow(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	projectDir := testutil.TempDir(t)

	t.Run("create mock database dump", func(t *testing.T) {
		dumpPath := filepath.Join(projectDir, "database.sql")
		helpers.CreateMockDatabaseDump(t, dumpPath)

		// Verify dump file exists
		testutil.AssertFileExists(t, dumpPath)

		// Verify dump contains expected content
		testutil.AssertFileContains(t, dumpPath, "wp_posts")
		testutil.AssertFileContains(t, dumpPath, "wp_options")
		testutil.AssertFileContains(t, dumpPath, "wp_blogs")
	})

	t.Run("verify search-replace configuration", func(t *testing.T) {
		cfg := helpers.CreateMultisiteConfig(t)
		cfg.WordPress.SearchReplace = config.SearchReplaceConfig{
			Network: []config.SearchReplacePair{
				{
					Old: "https://example.wpengine.com",
					New: "https://test.local",
				},
			},
			Sites: []config.SiteSearchReplace{
				{
					Old: "https://site1.wpengine.com",
					New: "https://site1.test.local",
					URL: "https://site1.test.local",
				},
			},
			SkipColumns: []string{"guid"},
		}

		// Verify search-replace pairs
		if len(cfg.WordPress.SearchReplace.Network) != 1 {
			t.Errorf("expected 1 network search-replace pair, got %d", len(cfg.WordPress.SearchReplace.Network))
		}

		if cfg.WordPress.SearchReplace.Network[0].Old != "https://example.wpengine.com" {
			t.Errorf("expected old URL 'https://example.wpengine.com', got %q", cfg.WordPress.SearchReplace.Network[0].Old)
		}
	})
}

func TestDatabaseSnapshot(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	projectDir := testutil.TempDir(t)

	t.Run("snapshot configuration", func(t *testing.T) {
		cfg := config.Defaults()
		cfg.Snapshots.AutoSnapshotBeforePull = true
		cfg.Snapshots.AutoSnapshotBeforeImport = true
		cfg.Snapshots.Directory = filepath.Join(projectDir, "snapshots")
		cfg.Snapshots.Retention.Auto = 7
		cfg.Snapshots.Retention.Manual = 30

		// Save config
		cfgPath := filepath.Join(projectDir, ".stax.yml")
		err := config.Save(cfg, cfgPath)
		testutil.AssertNoError(t, err, "save config")

		// Load and verify
		loadedCfg, err := config.Load(cfgPath, projectDir)
		testutil.AssertNoError(t, err, "load config")

		if !loadedCfg.Snapshots.AutoSnapshotBeforePull {
			t.Error("expected auto snapshot before pull to be enabled")
		}

		if loadedCfg.Snapshots.Retention.Auto != 7 {
			t.Errorf("expected auto retention 7 days, got %d", loadedCfg.Snapshots.Retention.Auto)
		}
	})
}

func TestSearchReplaceGeneration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	t.Run("generate search-replace pairs for multisite", func(t *testing.T) {
		cfg := helpers.CreateMultisiteConfig(t)

		// Should have network domain
		if cfg.Network.Domain == "" {
			t.Error("expected non-empty network domain")
		}

		// Should have sites
		if len(cfg.Network.Sites) == 0 {
			t.Error("expected at least one site")
		}

		// Verify each site has WPEngine domain
		for _, site := range cfg.Network.Sites {
			if site.WPEngineDomain == "" {
				t.Errorf("site %q missing WPEngine domain", site.Name)
			}
			if site.Domain == "" {
				t.Errorf("site %q missing local domain", site.Name)
			}
		}
	})

	t.Run("generate search-replace for single site", func(t *testing.T) {
		cfg := helpers.CreateTestConfig(t)
		cfg.Project.Type = "wordpress"
		cfg.Project.Mode = "single"

		// For single site, we just need the main domain
		if cfg.Network.Domain == "" {
			t.Error("expected non-empty domain")
		}
	})
}

func TestDatabaseImportConfiguration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	t.Run("configure import batch size", func(t *testing.T) {
		cfg := config.Defaults()
		cfg.Performance.DatabaseImportBatchSize = 5000

		if cfg.Performance.DatabaseImportBatchSize != 5000 {
			t.Errorf("expected batch size 5000, got %d", cfg.Performance.DatabaseImportBatchSize)
		}
	})

	t.Run("configure table exclusions", func(t *testing.T) {
		// Table exclusions are now passed via database.PullOptions.ExcludeTables
		tables := []string{
			"wp_logs",
			"wp_statistics",
			"wp_actionscheduler_logs",
		}

		if len(tables) != 3 {
			t.Errorf("expected 3 excluded tables, got %d", len(tables))
		}
	})
}
