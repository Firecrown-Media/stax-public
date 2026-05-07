package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

// configCmd represents the config command group
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `Manage Stax configuration files and values.`,
}

var (
	configGlobal bool
	configJSON   bool
	configFormat string
)

// configGetCmd represents the config:get command
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Example: `  # Get project config
  stax config get wpengine.environment

  # Get global config
  stax config get wpengine.api_user --global`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// configSetCmd represents the config:set command
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Example: `  # Set project config
  stax config set wpengine.environment staging

  # Set global config
  stax config set ddev.php_version 8.2 --global`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configShowCmd represents the config:show command
var configShowCmd = &cobra.Command{
	Use:     "show",
	Aliases: []string{"display"},
	Short:   "Show current configuration",
	Example: `  # Show project config
  stax config show

  # Show as JSON
  stax config show --format json

  # Show as YAML
  stax config show --format yaml

  # Show global config
  stax config show --global`,
	RunE: runConfigShow,
}

// configListCmd represents the config:list command
var configListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all configuration values",
	Example: `  # List project config
  stax config list

  # List global config
  stax config list --global

  # List as JSON
  stax config list --json`,
	RunE: runConfigList,
}

// configValidateCmd represents the config:validate command
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Example: `  # Validate config
  stax config validate`,
	RunE: runConfigValidate,
}

// configTemplateCmd represents the config:template command
var configTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Generate a configuration template",
	Example: `  # Generate template to stdout
  stax config template

  # Generate template to file
  stax config template > .stax.yml`,
	RunE: runConfigTemplate,
}

// configMigrateCmd represents the config:migrate command
var configMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate configuration to the latest schema version",
	Long: `Detect and migrate outdated configuration formats to the latest schema.

This command will:
  1. Detect the current configuration version
  2. Show what changes will be made
  3. Create a backup of the current configuration
  4. Migrate to the latest format
  5. Validate the migrated configuration`,
	Example: `  # Show what would be migrated (dry-run)
  stax config migrate --dry-run

  # Migrate configuration
  stax config migrate

  # List available backups
  stax config migrate --list-backups`,
	RunE: runConfigMigrate,
}

var (
	migrateDryRun      bool
	migrateListBackups bool
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configTemplateCmd)
	configCmd.AddCommand(configMigrateCmd)

	// Flags for show
	configShowCmd.Flags().BoolVar(&configGlobal, "global", false, "show global config")
	configShowCmd.Flags().StringVar(&configFormat, "format", "pretty", "output format (pretty, json, yaml)")

	// Flags for get
	configGetCmd.Flags().BoolVar(&configGlobal, "global", false, "get from global config")

	// Flags for set
	configSetCmd.Flags().BoolVar(&configGlobal, "global", false, "set in global config")

	// Flags for list
	configListCmd.Flags().BoolVar(&configGlobal, "global", false, "list global config")
	configListCmd.Flags().BoolVar(&configJSON, "json", false, "output as JSON")

	// Flags for migrate
	configMigrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "show what would change without modifying files")
	configMigrateCmd.Flags().BoolVar(&migrateListBackups, "list-backups", false, "list available configuration backups")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Load configuration
	var cfg *config.Config
	var err error

	if configGlobal {
		globalPath, err := config.GetGlobalConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get global config path: %w", err)
		}
		cfg, err = config.Load("", "")
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}
		ui.Debug("Loaded global config from: %s", globalPath)
	} else {
		projectDir := getProjectDir()
		cfgPath := config.GetProjectConfigPath(projectDir)
		cfg, err = config.Load(cfgPath, projectDir)
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}
		ui.Debug("Loaded project config from: %s", cfgPath)
	}

	// Format and display
	output, err := config.FormatConfig(cfg, configFormat)
	if err != nil {
		return fmt.Errorf("failed to format config: %w", err)
	}

	fmt.Println(output)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	// Load configuration
	var cfg *config.Config
	var err error

	if configGlobal {
		globalPath, err := config.GetGlobalConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get global config path: %w", err)
		}
		cfg, err = config.Load("", "")
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}
		ui.Debug("Loaded global config from: %s", globalPath)
	} else {
		projectDir := getProjectDir()
		cfgPath := config.GetProjectConfigPath(projectDir)
		cfg, err = config.Load(cfgPath, projectDir)
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}
		ui.Debug("Loaded project config from: %s", cfgPath)
	}

	// Get value by path
	value, err := config.GetValueByPath(cfg, key)
	if err != nil {
		ui.Error("Config key not found: %s", key)
		return fmt.Errorf("failed to get config value: %w", err)
	}

	// Display value
	fmt.Println(config.FormatValue(value))
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Determine config file path
	var cfgPath string
	if configGlobal {
		var err error
		cfgPath, err = config.GetGlobalConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get global config path: %w", err)
		}
	} else {
		projectDir := getProjectDir()
		cfgPath = config.GetProjectConfigPath(projectDir)
	}

	// Load configuration
	cfg, err := config.Load(cfgPath, "")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate path exists
	if err := config.ValidatePath(cfg, key); err != nil {
		ui.Error("Invalid config path: %s", key)
		ui.Info("Use 'stax config show' to see available configuration options")
		return err
	}

	// Create backup before modifying
	backupPath := cfgPath + ".backup." + time.Now().Format("20060102-150405")
	if err := copyFile(cfgPath, backupPath); err != nil {
		ui.Warning("Failed to create backup: %v", err)
	} else {
		ui.Debug("Created backup: %s", backupPath)
	}

	// Set the value
	if err := config.SetValueByPath(cfg, key, value); err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	// Save configuration
	if err := config.Save(cfg, cfgPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Display confirmation
	ui.Success("Updated %s to %s", key, value)
	ui.Info("  Configuration saved to %s", cfgPath)
	if _, err := os.Stat(backupPath); err == nil {
		ui.Info("  Backup saved to %s", backupPath)
	}

	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	// Load configuration
	var cfg *config.Config
	var err error

	if configGlobal {
		globalPath, err := config.GetGlobalConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get global config path: %w", err)
		}
		cfg, err = config.Load("", "")
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}
		ui.Debug("Loaded global config from: %s", globalPath)
	} else {
		projectDir := getProjectDir()
		cfgPath := config.GetProjectConfigPath(projectDir)
		cfg, err = config.Load(cfgPath, projectDir)
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}
		ui.Debug("Loaded project config from: %s", cfgPath)
	}

	// Format based on flags
	format := "pretty"
	if configJSON {
		format = "json"
	}

	// Format and display
	output, err := config.FormatConfig(cfg, format)
	if err != nil {
		return fmt.Errorf("failed to format config: %w", err)
	}

	fmt.Println(output)
	return nil
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Validating Configuration")

	// Load config from current directory
	cfg := getConfig()
	if cfg == nil {
		return fmt.Errorf("no configuration found - run 'stax config template > .stax.yml' to create one")
	}

	ui.Info("Validating .stax.yml configuration...")
	fmt.Println()

	// Run validation
	result := config.Validate(cfg)

	// Display errors
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			ui.Error("%s: %s", err.Field, err.Message)
			if err.Fix != "" {
				ui.Info("  → %s", err.Fix)
			}
			fmt.Println()
		}
	}

	// Display warnings
	if len(result.Warnings) > 0 {
		for _, warn := range result.Warnings {
			ui.Warning("%s: %s", warn.Field, warn.Message)
			if warn.Fix != "" {
				ui.Info("  → %s", warn.Fix)
			}
			fmt.Println()
		}
	}

	// Display info messages (if verbose)
	if verbose && len(result.Infos) > 0 {
		for _, info := range result.Infos {
			ui.Info("ℹ %s: %s", info.Field, info.Message)
			fmt.Println()
		}
	}

	// Summary
	fmt.Println()
	if !result.Valid {
		errorCount := len(result.Errors)
		warningCount := len(result.Warnings)

		summary := fmt.Sprintf("Validation failed with %d error(s)", errorCount)
		if warningCount > 0 {
			summary += fmt.Sprintf(", %d warning(s)", warningCount)
		}
		ui.Error("%s", summary)
		return fmt.Errorf("configuration validation failed")
	}

	if len(result.Warnings) > 0 {
		ui.Success("Configuration is valid (with %d warning(s))", len(result.Warnings))
	} else {
		ui.Success("Configuration is valid!")
	}

	return nil
}

func runConfigTemplate(cmd *cobra.Command, args []string) error {
	// Generate a template .stax.yml configuration
	template := `# Stax Configuration
# Version: 1

version: 1

# Project metadata
project:
  name: my-wordpress-project
  type: wordpress-multisite  # wordpress or wordpress-multisite
  mode: subdomain            # subdomain, subdirectory, or single
  description: My WordPress Multisite Project

# WPEngine integration
wpengine:
  install: myinstall           # WPEngine install name
  environment: production      # production, staging, or development
  account_name: myaccount      # WPEngine account name (optional)
  ssh_gateway: ssh.wpengine.net

  # Backup preferences
  backup:
    auto_snapshot: true        # Create snapshot before DB pull
    skip_logs: true           # Skip log tables
    skip_transients: true     # Skip transient data
    skip_spam: true           # Skip spam comments
    exclude_tables: []        # Additional tables to exclude

  # Domain mapping (optional)
  domains:
    production:
      primary: example.com
      sites:
        - site1.example.com
        - site2.example.com
    staging:
      primary: staging.example.com

# Network configuration (for multisite)
network:
  domain: example.local      # Local development domain
  title: My Network         # Network title
  admin_email: admin@example.local

  # Individual sites (optional - can be auto-detected)
  sites:
    - name: Main Site
      slug: main
      title: Main Site
      domain: example.local
      path: /
    - name: Site One
      slug: site1
      title: Site One
      domain: site1.example.local
      path: /

# DDEV configuration
ddev:
  name: ""                  # Auto-generated from project name
  type: wordpress
  php_version: "8.1"
  mysql_version: "8.0"
  webserver_type: nginx-fpm
  router_http_port: "80"
  router_https_port: "443"
  xdebug_enabled: false
  additional_hostnames: []
  additional_fqdns: []

# GitHub repository (optional)
repository:
  url: https://github.com/username/repo.git
  branch: main
  private: true

# Build process (optional)
build:
  enabled: true
  composer:
    enabled: true
    install_args: "--no-dev --optimize-autoloader"
  npm:
    enabled: true
    install_command: npm ci
    build_command: npm run build
    dev_command: npm run dev
    watch_command: npm run watch
  quality:
    phpcs: true
    phpcbf: true

# WordPress configuration (optional)
wordpress:
  version: latest
  locale: en_US
  timezone: America/New_York
  debug: true
  debug_log: true

# Remote media configuration
media:
  enabled: true
  strategy: bunnycdn         # bunnycdn, wpengine, or disabled
  cache_ttl: 2592000         # 30 days
  cache_max_size: 10737418240  # 10GB

# Logging and debugging (optional)
logging:
  level: info               # debug, info, warning, error
  file: .stax/logs/stax.log
  rotate: true

# Snapshots (optional)
snapshots:
  directory: .stax/snapshots
  retention_days: 30
  auto_cleanup: true

# Performance tuning (optional)
performance:
  db_import_chunk_size: 1048576  # 1MB
  parallel_downloads: 4
  connection_timeout: 30
`

	fmt.Println(template)
	return nil
}

func runConfigMigrate(cmd *cobra.Command, args []string) error {
	return runConfigMigrateImplementation(cmd, args)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}
