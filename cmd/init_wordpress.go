package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/ui"
)

// hasWordPressCore checks if WordPress core files are present in the docroot
func hasWordPressCore(projectDir string) bool {
	// Read DDEV config to get docroot
	ddevConfig, err := ddev.ReadConfig(projectDir)
	docroot := "." // default
	if err == nil && ddevConfig.DocRoot != "" {
		docroot = ddevConfig.DocRoot
	}

	// Check for wp-includes/version.php as indicator of WordPress core
	versionPath := filepath.Join(projectDir, docroot, "wp-includes", "version.php")
	if _, err := os.Stat(versionPath); err == nil {
		return true
	}

	// Also check for wp-load.php in docroot directory
	loadPath := filepath.Join(projectDir, docroot, "wp-load.php")
	if _, err := os.Stat(loadPath); err == nil {
		return true
	}

	return false
}

// downloadWordPressCore downloads WordPress core files via DDEV to the correct docroot
func downloadWordPressCore(projectDir string, version string) error {
	// Check if DDEV is running
	mgr := ddev.NewManager(projectDir)
	running, err := mgr.IsRunning()
	if err != nil || !running {
		return fmt.Errorf("DDEV must be running to download WordPress core")
	}

	// Read DDEV config to get docroot location
	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read DDEV config: %w", err)
	}

	// Get docroot (default to "." if not specified)
	docroot := ddevConfig.DocRoot
	if docroot == "" {
		docroot = "."
	}

	// Ensure docroot directory exists
	docrootPath := filepath.Join(projectDir, docroot)
	if err := os.MkdirAll(docrootPath, 0755); err != nil {
		return fmt.Errorf("failed to create docroot directory: %w", err)
	}

	// Build WP-CLI command with --path to specify docroot
	args := []string{"wp", "core", "download", fmt.Sprintf("--path=%s", docroot)}
	if version != "" && version != "latest" {
		args = append(args, fmt.Sprintf("--version=%s", version))
		ui.Info("Downloading WordPress %s to %s/...", version, docroot)
	} else {
		ui.Info("Downloading latest WordPress to %s/...", docroot)
	}

	// Execute download
	spinner := ui.NewSpinner("Downloading WordPress core...")
	spinner.Start()

	if err := mgr.Exec(args, nil); err != nil {
		spinner.Error("Failed to download WordPress core")
		return fmt.Errorf("WordPress download failed: %w\n\nNote: WordPress should be installed to %s/ directory", err, docroot)
	}

	spinner.Success(fmt.Sprintf("WordPress core downloaded successfully to %s/", docroot))
	return nil
}

// hasWordPressConfig checks if wp-config.php exists in the docroot
func hasWordPressConfig(projectDir string) bool {
	// Read DDEV config to get docroot
	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		// If we can't read config, check default location
		configPath := filepath.Join(projectDir, "wp-config.php")
		_, err := os.Stat(configPath)
		return err == nil
	}

	docroot := ddevConfig.DocRoot
	if docroot == "" {
		docroot = "."
	}

	configPath := filepath.Join(projectDir, docroot, "wp-config.php")
	_, err = os.Stat(configPath)
	return err == nil
}

// generateWordPressConfig creates wp-config.php via DDEV and configures multisite if needed
func generateWordPressConfig(projectDir string, cfg *config.Config) error {
	ui.Section("Generating wp-config.php")

	// Check if DDEV is running
	mgr := ddev.NewManager(projectDir)
	running, err := mgr.IsRunning()
	if err != nil || !running {
		return fmt.Errorf("DDEV must be running to generate wp-config.php")
	}

	// Read DDEV config to get docroot location
	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read DDEV config: %w", err)
	}

	// Get docroot (default to "." if not specified)
	docroot := ddevConfig.DocRoot
	if docroot == "" {
		docroot = "."
	}

	// DDEV database defaults
	dbName := "db"
	dbUser := "db"
	dbPass := "db"
	dbHost := "db"

	// Build WP-CLI command to create wp-config.php in the correct docroot
	args := []string{
		"wp", "config", "create",
		fmt.Sprintf("--path=%s", docroot),
		fmt.Sprintf("--dbname=%s", dbName),
		fmt.Sprintf("--dbuser=%s", dbUser),
		fmt.Sprintf("--dbpass=%s", dbPass),
		fmt.Sprintf("--dbhost=%s", dbHost),
		"--skip-check", // Skip DB connection check (DB may not have tables yet)
	}

	// Execute config creation
	spinner := ui.NewSpinner("Creating wp-config.php...")
	spinner.Start()

	if err := mgr.Exec(args, nil); err != nil {
		spinner.Error("Failed to generate wp-config.php")
		return err
	}

	spinner.Success("wp-config.php created successfully")

	// For multisite, add additional configuration
	if isMultisite(cfg.Project.Type) {
		if err := configureMultisite(projectDir, cfg, mgr); err != nil {
			ui.Warning(fmt.Sprintf("Multisite configuration warning: %v", err))
			ui.Info("You may need to manually configure multisite in wp-config.php")
		}
	}

	ui.Success("wp-config.php generated successfully")
	return nil
}

// configureMultisite adds multisite constants to wp-config.php
func configureMultisite(projectDir string, cfg *config.Config, mgr *ddev.Manager) error {
	ui.Info("Adding multisite configuration...")

	// Determine subdomain vs subdirectory
	subdomainInstall := "true"
	if cfg.Project.Mode == "subdirectory" {
		subdomainInstall = "false"
	}

	// Get the domain from config
	domain := cfg.Network.Domain
	if domain == "" {
		domain = fmt.Sprintf("%s.ddev.site", cfg.Project.Name)
	}

	// Add multisite constants via wp-cli
	multisiteConstants := []struct {
		name  string
		value string
		raw   bool
	}{
		{"MULTISITE", "true", true},
		{"SUBDOMAIN_INSTALL", subdomainInstall, true},
		{"DOMAIN_CURRENT_SITE", fmt.Sprintf("'%s'", domain), false},
		{"PATH_CURRENT_SITE", "'/'", false},
		{"SITE_ID_CURRENT_SITE", "1", true},
		{"BLOG_ID_CURRENT_SITE", "1", true},
	}

	for _, constant := range multisiteConstants {
		args := []string{"wp", "config", "set", constant.name, constant.value}
		if constant.raw {
			args = append(args, "--raw")
		}

		if err := mgr.Exec(args, &ddev.ExecOptions{Service: "web"}); err != nil {
			return fmt.Errorf("failed to set %s: %w", constant.name, err)
		}
	}

	ui.Success("Multisite configuration added")
	return nil
}
