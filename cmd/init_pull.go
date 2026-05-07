package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/ui"
)

// pullDatabase pulls the database from WPEngine during initialization
func pullDatabase(projectDir string, cfg *config.Config) error {
	ui.Section("Pulling Database")
	ui.Info("This may take several minutes...")

	// Verify WPEngine configuration exists
	if wpeInstall(cfg) == "" {
		return fmt.Errorf("WPEngine install not configured")
	}

	// Save current directory and change to project directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			ui.Debug("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}

	// Set the environment for database pull
	dbEnvironment = wpeEnv(cfg)
	if dbEnvironment == "" {
		dbEnvironment = "production"
	}

	// Call the existing database pull function
	if err := runDBPull(nil, nil); err != nil {
		return fmt.Errorf("database pull failed: %w\n\nYou can try manually: stax db pull --environment=%s", err, dbEnvironment)
	}

	ui.Success("Database pulled successfully")
	return nil
}

// pullFiles pulls WordPress files from WPEngine.
// If media proxy is enabled in the config, it excludes the uploads directory
// and only syncs themes, plugins, and mu-plugins. Otherwise, it syncs all files
// including uploads.
//
// Media proxy configuration is found in cfg.Media.ProxyEnabled. When enabled,
// media files are served from the remote WPEngine server on-demand rather than
// being stored locally, significantly reducing sync time and disk usage.
func pullFiles(projectDir string, cfg *config.Config) error {
	ui.Section("Pulling Files")

	// Verify WPEngine configuration exists
	if wpeInstall(cfg) == "" {
		return fmt.Errorf("WPEngine install not configured")
	}

	// Save current directory and change to project directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			ui.Debug("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}

	// Set the environment for file pull
	filesEnvironment = wpeEnv(cfg)
	if filesEnvironment == "" {
		filesEnvironment = "production"
	}

	// Only exclude uploads if media proxy is enabled in config
	if cfg.Media.ProxyEnabled {
		filesExcludeUploads = true
		ui.Info("Media proxy enabled - excluding uploads directory")
		ui.Info("Syncing: themes, plugins, mu-plugins")
		ui.Info("This may take several minutes...")
	} else {
		filesExcludeUploads = false
		ui.Info("Media proxy disabled - pulling all files including uploads")
		ui.Info("Syncing: themes, plugins, mu-plugins, uploads")
		ui.Info("This may take longer due to media files...")
	}

	// Call the existing file pull function
	if err := runFilesPull(nil, nil); err != nil {
		// Provide helpful error message with next steps
		if cfg.Media.ProxyEnabled {
			return fmt.Errorf("file pull failed: %w\n\nYou can try manually: stax files pull --environment=%s --exclude-uploads", err, filesEnvironment)
		}
		return fmt.Errorf("file pull failed: %w\n\nYou can try manually: stax files pull --environment=%s", err, filesEnvironment)
	}

	ui.Success("Files pulled successfully")

	// Validate that critical directories exist and have content
	if err := validatePulledFiles(projectDir, cfg); err != nil {
		ui.Warning("File validation warnings detected:")
		ui.Warning(err.Error())
		ui.Info("\nNext steps:")
		ui.Info("  1. Check your WPEngine install has themes and plugins")
		ui.Info("  2. Verify SSH access: stax files pull --environment=%s", filesEnvironment)
		ui.Info("  3. Check .stax.yml configuration")
	}

	return nil
}

// validatePulledFiles checks that critical WordPress directories exist and have content.
// This helps catch sync issues early and provides clear feedback to users.
func validatePulledFiles(projectDir string, cfg *config.Config) error {
	// Read DDEV config to get docroot
	ddevConfig, err := ddev.ReadConfig(projectDir)
	docroot := "."
	if err == nil && ddevConfig.DocRoot != "" {
		docroot = ddevConfig.DocRoot
	}

	wpContentDir := filepath.Join(projectDir, docroot, "wp-content")
	var warnings []string

	// Check themes directory
	themesDir := filepath.Join(wpContentDir, "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		warnings = append(warnings, "themes directory not found")
	} else {
		// Check if themes directory has content
		entries, err := os.ReadDir(themesDir)
		if err == nil && len(entries) == 0 {
			warnings = append(warnings, "themes directory is empty")
		}
	}

	// Check plugins directory
	pluginsDir := filepath.Join(wpContentDir, "plugins")
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		warnings = append(warnings, "plugins directory not found")
	}

	// Check mu-plugins directory (optional, so only warn if expected)
	muPluginsDir := filepath.Join(wpContentDir, "mu-plugins")
	if _, err := os.Stat(muPluginsDir); os.IsNotExist(err) {
		ui.Info("Note: mu-plugins directory not found (this is optional)")
	}

	// Check uploads directory only if media proxy is disabled
	if !cfg.Media.ProxyEnabled {
		uploadsDir := filepath.Join(wpContentDir, "uploads")
		if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
			warnings = append(warnings, "uploads directory not found (media proxy disabled)")
		}
	}

	if len(warnings) > 0 {
		return fmt.Errorf("%s", strings.Join(warnings, "; "))
	}

	ui.Success("File validation passed")
	ui.Info("  - themes: present")
	ui.Info("  - plugins: present")

	if cfg.Media.ProxyEnabled {
		ui.Info("  - uploads: excluded (media proxy enabled)")
	} else {
		ui.Info("  - uploads: synced")
	}

	return nil
}
