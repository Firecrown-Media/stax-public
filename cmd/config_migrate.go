package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

func runConfigMigrateImplementation(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Configuration Migration")

	// Get config path
	configPath := config.GetProjectConfigPath(getProjectDir())
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("no configuration found - run 'stax config template > .stax.yml' to create one")
	}

	// Handle list-backups flag
	if migrateListBackups {
		return listConfigBackups(configPath)
	}

	ui.Info("Detecting configuration version from: %s", configPath)
	fmt.Println()

	// Detect current version
	version, err := config.DetectConfigVersion(configPath)
	if err != nil {
		return fmt.Errorf("failed to detect config version: %w", err)
	}

	ui.Debug("Detected version: %d, schema: %s", version.Version, version.Schema)
	ui.Success("Found configuration (version: %d, schema: %s)", version.Version, version.Schema)

	// Get migration plan
	plan, err := config.GetMigrationPlan(configPath)
	if err != nil {
		// Check if already up to date
		if version.Version >= config.CurrentVersion {
			ui.Success("Configuration is already at version %d (latest)", config.CurrentVersion)
			return nil
		}
		return fmt.Errorf("failed to create migration plan: %w", err)
	}

	// Show migration information
	ui.Warning("Migration available: v%d -> v%d", plan.FromVersion, plan.ToVersion)
	fmt.Println()

	// Show changes
	if len(plan.Changes) > 0 {
		ui.Info("Changes:")
		for _, change := range plan.Changes {
			switch change.Type {
			case "rename":
				ui.Info("  - Rename: %s -> %s", change.OldPath, change.NewPath)
			case "add":
				ui.Info("  - Add: %s (default: %v)", change.NewPath, change.NewValue)
			case "remove":
				ui.Info("  - Remove: deprecated field '%s'", change.OldPath)
			case "update":
				ui.Info("  - Update: %s (%v -> %v)", change.NewPath, change.OldValue, change.NewValue)
			case "restructure":
				ui.Info("  - Restructure: %s", change.Description)
			default:
				ui.Info("  - %s", change.Description)
			}
		}
		fmt.Println()
	} else {
		ui.Info("Changes:")
		ui.Info("  - Update version field to latest")
		ui.Info("  - Ensure all required fields have defaults")
		fmt.Println()
	}

	// Dry run mode
	if migrateDryRun {
		ui.Warning("Dry-run mode: no files will be modified")
		ui.Info("Run without --dry-run to perform the migration")
		return nil
	}

	// Confirm migration
	ui.Info("Creating backup: %s.backup.*", configPath)

	// Execute migration
	_, err = config.MigrateConfig(configPath, false)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	ui.Info("Migrating configuration...")
	ui.Info("Validating migrated config...")
	fmt.Println()

	// Success message
	ui.Success("Migration complete!")
	ui.Info("  .stax.yml updated to version %d", config.CurrentVersion)

	// List the backup that was created
	backups, err := config.ListBackups(configPath)
	if err == nil && len(backups) > 0 {
		latestBackup := backups[len(backups)-1]
		ui.Info("  Backup saved to %s", latestBackup)
	}

	return nil
}

func listConfigBackups(configPath string) error {
	ui.PrintHeader("Configuration Backups")

	backups, err := config.ListBackups(configPath)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		ui.Info("No backups found")
		return nil
	}

	ui.Info("Found %d backup(s):", len(backups))
	fmt.Println()

	for i, backup := range backups {
		// Get file info for timestamp
		info, err := os.Stat(backup)
		var timestamp string
		if err == nil {
			timestamp = info.ModTime().Format(time.RFC3339)
		} else {
			timestamp = "unknown"
		}
		ui.Info("%d. %s (created: %s)", i+1, filepath.Base(backup), timestamp)
	}

	return nil
}
