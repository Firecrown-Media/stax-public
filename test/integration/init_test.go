//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/testutil"
)

func TestInitWorkflow(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	// Create temporary project directory
	projectDir := testutil.TempDir(t)

	t.Run("initialize configuration", func(t *testing.T) {
		// Create default config
		cfg := config.Defaults()
		cfg.Project.Name = "integration-test"
		cfg.ProviderConfig["install"] = "testinstall"
		cfg.Network.Domain = "integration-test.local"

		// Save config
		cfgPath := filepath.Join(projectDir, ".stax.yml")
		err := config.Save(cfg, cfgPath)
		testutil.AssertNoError(t, err, "save config")

		// Verify config file exists
		testutil.AssertFileExists(t, cfgPath)

		// Load and verify config
		loadedCfg, err := config.Load(cfgPath, projectDir)
		testutil.AssertNoError(t, err, "load config")

		if loadedCfg.Project.Name != "integration-test" {
			t.Errorf("expected project name 'integration-test', got %q", loadedCfg.Project.Name)
		}
	})

	t.Run("setup project structure", func(t *testing.T) {
		// Create WordPress directory structure
		testutil.CreateTestProject(t, projectDir)

		// Verify directories exist
		testutil.AssertDirExists(t, filepath.Join(projectDir, "wp-content"))
		testutil.AssertDirExists(t, filepath.Join(projectDir, "wp-content/themes"))
		testutil.AssertDirExists(t, filepath.Join(projectDir, "wp-content/plugins"))

		// Verify package.json exists
		testutil.AssertFileExists(t, filepath.Join(projectDir, "package.json"))
	})

	t.Run("verify DDEV availability", func(t *testing.T) {
		if !ddev.IsInstalled() {
			t.Skip("DDEV not installed, skipping DDEV tests")
		}

		version, err := ddev.GetVersion()
		testutil.AssertNoError(t, err, "get DDEV version")

		if version == "" {
			t.Error("expected non-empty DDEV version")
		}

		t.Logf("DDEV version: %s", version)
	})
}

func TestConfigurationMerging(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	projectDir := testutil.TempDir(t)

	t.Run("project config overrides defaults", func(t *testing.T) {
		// Create project config with custom values
		cfg := config.Defaults()
		cfg.Project.Name = "custom-project"
		cfg.DDEV.PHPVersion = "8.2"
		cfg.ProviderConfig["environment"] = "staging"

		// Save and load
		cfgPath := filepath.Join(projectDir, ".stax.yml")
		err := config.Save(cfg, cfgPath)
		testutil.AssertNoError(t, err, "save config")

		loadedCfg, err := config.Load(cfgPath, projectDir)
		testutil.AssertNoError(t, err, "load config")

		// Verify custom values
		testutil.AssertEqual(t, loadedCfg.Project.Name, "custom-project")
		testutil.AssertEqual(t, loadedCfg.DDEV.PHPVersion, "8.2")
		env, _ := loadedCfg.ProviderConfig["environment"].(string)
		testutil.AssertEqual(t, env, "staging")
	})

	t.Run("environment variables override config", func(t *testing.T) {
		// Set environment variables
		testutil.SetEnv(t, "STAX_PROJECT_NAME", "env-project")
		testutil.SetEnv(t, "STAX_WPENGINE_INSTALL", "env-install")

		// Load config
		cfg := config.Defaults()
		cfgPath := filepath.Join(projectDir, ".stax.yml")
		err := config.Save(cfg, cfgPath)
		testutil.AssertNoError(t, err, "save config")

		loadedCfg, err := config.Load(cfgPath, projectDir)
		testutil.AssertNoError(t, err, "load config")

		// Verify environment variables took precedence
		testutil.AssertEqual(t, loadedCfg.Project.Name, "env-project")
		install, _ := loadedCfg.ProviderConfig["install"].(string)
		testutil.AssertEqual(t, install, "env-install")
	})
}

func TestMultisiteConfiguration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	projectDir := testutil.TempDir(t)

	t.Run("setup multisite network", func(t *testing.T) {
		// Create multisite config
		cfg := config.Defaults()
		cfg.Project.Name = "multisite-test"
		cfg.Project.Type = "wordpress-multisite"
		cfg.Project.Mode = "subdomain"
		cfg.Network.Domain = "multisite.local"
		cfg.Network.Sites = []config.SiteConfig{
			{
				Name:           "Site 1",
				Slug:           "site1",
				Domain:         "site1.multisite.local",
				ProviderDomain: "site1.wpengine.com",
				Active:         true,
			},
			{
				Name:           "Site 2",
				Slug:           "site2",
				Domain:         "site2.multisite.local",
				ProviderDomain: "site2.wpengine.com",
				Active:         true,
			},
		}

		// Save config
		cfgPath := filepath.Join(projectDir, ".stax.yml")
		err := config.Save(cfg, cfgPath)
		testutil.AssertNoError(t, err, "save config")

		// Load and verify
		loadedCfg, err := config.Load(cfgPath, projectDir)
		testutil.AssertNoError(t, err, "load config")

		testutil.AssertEqual(t, loadedCfg.Project.Type, "wordpress-multisite")
		testutil.AssertEqual(t, len(loadedCfg.Network.Sites), 2)
		testutil.AssertEqual(t, loadedCfg.Network.Sites[0].Name, "Site 1")
	})
}

func TestBuildConfiguration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	projectDir := testutil.TempDir(t)
	testutil.CreateTestProject(t, projectDir)

	t.Run("composer configuration", func(t *testing.T) {
		cfg := config.Defaults()
		cfg.Build.Composer.Optimize = true
		cfg.Build.Composer.NoDev = true
		cfg.Build.Composer.InstallArgs = "--no-dev --optimize-autoloader"

		cfgPath := filepath.Join(projectDir, ".stax.yml")
		err := config.Save(cfg, cfgPath)
		testutil.AssertNoError(t, err, "save config")

		loadedCfg, err := config.Load(cfgPath, projectDir)
		testutil.AssertNoError(t, err, "load config")

		if !loadedCfg.Build.Composer.Optimize {
			t.Error("expected composer optimize to be true")
		}
	})

	t.Run("npm configuration", func(t *testing.T) {
		cfg := config.Defaults()
		cfg.Build.NPM.BuildCommand = "npm run production"
		cfg.Build.NPM.DevCommand = "npm run dev"
		cfg.Build.NPM.LegacyPeerDeps = true

		cfgPath := filepath.Join(projectDir, ".stax.yml")
		err := config.Save(cfg, cfgPath)
		testutil.AssertNoError(t, err, "save config")

		loadedCfg, err := config.Load(cfgPath, projectDir)
		testutil.AssertNoError(t, err, "load config")

		testutil.AssertEqual(t, loadedCfg.Build.NPM.BuildCommand, "npm run production")
		if !loadedCfg.Build.NPM.LegacyPeerDeps {
			t.Error("expected legacy peer deps to be true")
		}
	})
}
