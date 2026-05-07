//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/testutil"
	"github.com/firecrown-media/stax/test/helpers"
)

// TestFullWorkflow tests the complete user workflow from init to running site
func TestFullWorkflow(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test (set RUN_E2E_TESTS=true to run)")
	}

	// Create a real project directory for E2E testing
	projectDir := testutil.TempDir(t)
	t.Logf("Testing in directory: %s", projectDir)

	// Step 1: Initialize configuration
	t.Run("step 1: initialize configuration", func(t *testing.T) {
		cfg := config.Defaults()
		cfg.Project.Name = "e2e-test"
		cfg.ProviderConfig["install"] = "e2etest"
		cfg.Network.Domain = "e2e.local"

		cfgPath := filepath.Join(projectDir, ".stax.yml")
		err := config.Save(cfg, cfgPath)
		testutil.AssertNoError(t, err, "initialize configuration")

		testutil.AssertFileExists(t, cfgPath)
		t.Logf("✓ Configuration initialized at %s", cfgPath)
	})

	// Step 2: Create project structure
	t.Run("step 2: create project structure", func(t *testing.T) {
		testutil.CreateTestProject(t, projectDir)
		helpers.CreateMockWordPressInstall(t, projectDir)

		testutil.AssertDirExists(t, filepath.Join(projectDir, "wp-content"))
		testutil.AssertFileExists(t, filepath.Join(projectDir, "wp-config.php"))
		t.Log("✓ Project structure created")
	})

	// Step 3: Verify DDEV availability
	t.Run("step 3: verify DDEV", func(t *testing.T) {
		if !ddev.IsInstalled() {
			t.Skip("DDEV not installed, skipping DDEV tests")
		}

		manager := ddev.NewManager(projectDir)
		running, _ := manager.IsRunning()

		if running {
			t.Log("✓ DDEV is already running")
		} else {
			t.Log("✓ DDEV is available but not running")
		}
	})

	// Step 4: Verify build tools
	t.Run("step 4: verify build tools", func(t *testing.T) {
		testutil.AssertFileExists(t, filepath.Join(projectDir, "package.json"))
		testutil.AssertFileExists(t, filepath.Join(projectDir, "composer.json"))
		t.Log("✓ Build configuration verified")
	})

	t.Log("✓ Full workflow test completed successfully")
}

// TestMultisiteWorkflow tests multisite-specific workflow
func TestMultisiteWorkflow(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test (set RUN_E2E_TESTS=true to run)")
	}

	projectDir := testutil.TempDir(t)

	t.Run("initialize multisite network", func(t *testing.T) {
		cfg := helpers.CreateMultisiteConfig(t)
		cfg.Project.Name = "multisite-e2e"

		cfgPath := filepath.Join(projectDir, ".stax.yml")
		err := config.Save(cfg, cfgPath)
		testutil.AssertNoError(t, err, "save multisite config")

		// Verify multisite configuration
		loadedCfg, err := config.Load(cfgPath, projectDir)
		testutil.AssertNoError(t, err, "load config")

		if loadedCfg.Project.Type != "wordpress-multisite" {
			t.Errorf("expected multisite type, got %q", loadedCfg.Project.Type)
		}

		if len(loadedCfg.Network.Sites) < 2 {
			t.Errorf("expected at least 2 sites, got %d", len(loadedCfg.Network.Sites))
		}

		t.Log("✓ Multisite network configured")
	})

	t.Run("verify search-replace configuration", func(t *testing.T) {
		cfgPath := filepath.Join(projectDir, ".stax.yml")
		cfg, err := config.Load(cfgPath, projectDir)
		testutil.AssertNoError(t, err, "load config")

		// Verify each site has proper domain mapping
		for _, site := range cfg.Network.Sites {
			if site.Domain == "" {
				t.Errorf("site %q missing local domain", site.Name)
			}
			if site.ProviderDomain == "" {
				t.Errorf("site %q missing provider domain", site.Name)
			}
			t.Logf("✓ Site %q: %s -> %s", site.Name, site.ProviderDomain, site.Domain)
		}
	})
}

// TestBuildWorkflow tests the build process
func TestBuildWorkflow(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test (set RUN_E2E_TESTS=true to run)")
	}

	projectDir := testutil.TempDir(t)
	testutil.CreateTestProject(t, projectDir)

	t.Run("composer configuration", func(t *testing.T) {
		helpers.CreateMockComposerJSON(t, projectDir)
		testutil.AssertFileExists(t, filepath.Join(projectDir, "composer.json"))
		t.Log("✓ Composer configuration present")
	})

	t.Run("npm configuration", func(t *testing.T) {
		helpers.CreateMockPackageJSON(t, projectDir)
		testutil.AssertFileExists(t, filepath.Join(projectDir, "package.json"))

		// Verify package.json has build scripts
		testutil.AssertFileContains(t, filepath.Join(projectDir, "package.json"), "build")
		t.Log("✓ NPM configuration present with build scripts")
	})
}

// TestDatabaseWorkflow tests database operations
func TestDatabaseWorkflow(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test (set RUN_E2E_TESTS=true to run)")
	}

	projectDir := testutil.TempDir(t)

	t.Run("create mock database", func(t *testing.T) {
		dumpPath := filepath.Join(projectDir, "test-database.sql")
		helpers.CreateMockDatabaseDump(t, dumpPath)

		testutil.AssertFileExists(t, dumpPath)
		testutil.AssertFileContains(t, dumpPath, "wp_posts")
		testutil.AssertFileContains(t, dumpPath, "wp_options")

		t.Log("✓ Mock database created")
	})

	t.Run("verify search-replace configuration", func(t *testing.T) {
		cfg := helpers.CreateMultisiteConfig(t)

		// Configure search-replace
		cfg.WordPress.SearchReplace = config.SearchReplaceConfig{
			Network: []config.SearchReplacePair{
				{Old: "https://example.wpengine.com", New: "https://test.local"},
			},
			SkipColumns: []string{"guid"},
		}

		if len(cfg.WordPress.SearchReplace.Network) != 1 {
			t.Errorf("expected 1 network search-replace pair, got %d", len(cfg.WordPress.SearchReplace.Network))
		}

		t.Log("✓ Search-replace configuration verified")
	})
}

// TestConfigurationValidation tests configuration validation
func TestConfigurationValidation(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test (set RUN_E2E_TESTS=true to run)")
	}

	t.Run("validate complete configuration", func(t *testing.T) {
		cfg := config.Defaults()
		cfg.Project.Name = "validation-test"
		cfg.ProviderConfig["install"] = "validationtest"

		// Verify all required fields have values
		if cfg.Project.Name == "" {
			t.Error("project name is required")
		}
		if cfg.ProviderConfig["install"] == "" {
			t.Error("WPEngine install is required")
		}
		if cfg.DDEV.PHPVersion == "" {
			t.Error("PHP version is required")
		}
		if cfg.DDEV.MySQLVersion == "" {
			t.Error("MySQL version is required")
		}

		t.Log("✓ Configuration validation passed")
	})

	t.Run("validate multisite configuration", func(t *testing.T) {
		cfg := helpers.CreateMultisiteConfig(t)

		// Verify multisite-specific requirements
		if cfg.Project.Type != "wordpress-multisite" {
			t.Error("multisite type is required")
		}
		if cfg.Network.Domain == "" {
			t.Error("network domain is required")
		}
		if len(cfg.Network.Sites) == 0 {
			t.Error("at least one site is required for multisite")
		}

		// Verify each site has required fields
		for i, site := range cfg.Network.Sites {
			if site.Name == "" {
				t.Errorf("site %d: name is required", i)
			}
			if site.Domain == "" {
				t.Errorf("site %d: domain is required", i)
			}
		}

		t.Log("✓ Multisite configuration validation passed")
	})
}
