package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/firecrown-media/stax/pkg/snapshot"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	snapshotDescription string
	snapshotName        string
)

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Create a manual database snapshot",
	Long: `Create a manual database snapshot.

Manual snapshots are retained for 30 days by default (configurable via config).
Use this before making risky database changes.`,
	Example: `  # Create a snapshot
  stax db snapshot

  # Create a snapshot with description
  stax db snapshot --description "before-major-update"`,
	RunE: runSnapshotCreate,
}

// snapshotListCmd represents the snapshot list command
var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List database snapshots",
	Long: `List all database snapshots for the current project.

Shows snapshot name, type (auto/manual), size, and creation date.`,
	Example: `  # List all snapshots
  stax db snapshot list`,
	RunE: runSnapshotList,
}

// snapshotRestoreCmd represents the snapshot restore command
var snapshotRestoreCmd = &cobra.Command{
	Use:   "restore <name>",
	Short: "Restore a database snapshot",
	Long: `Restore a database from a snapshot.

WARNING: This will replace your current database!`,
	Example: `  # Restore from a snapshot
  stax db snapshot restore mysite-20250115-143022-auto.sql.gz

  # Restore using just the filename
  stax db snapshot restore mysite-20250115-143022-auto.sql.gz`,
	Args: cobra.ExactArgs(1),
	RunE: runSnapshotRestore,
}

// snapshotCleanCmd represents the snapshot clean command
var snapshotCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean old snapshots",
	Long: `Delete old snapshots based on retention policy.

By default:
  - Auto snapshots older than 7 days are deleted
  - Manual snapshots older than 30 days are deleted

Retention periods can be configured in .stax.yml under snapshots.retention.`,
	Example: `  # Clean old snapshots
  stax db snapshot clean`,
	RunE: runSnapshotClean,
}

// snapshotDeleteCmd represents the snapshot delete command
var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a specific snapshot",
	Long: `Delete a specific snapshot by name.

This permanently deletes the snapshot file and removes it from metadata.`,
	Example: `  # Delete a snapshot
  stax db snapshot delete mysite-20250115-143022-auto.sql.gz`,
	Args: cobra.ExactArgs(1),
	RunE: runSnapshotDelete,
}

func init() {
	// Add snapshot command to db command
	dbCmd.AddCommand(snapshotCmd)

	// Add subcommands
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotRestoreCmd)
	snapshotCmd.AddCommand(snapshotCleanCmd)
	snapshotCmd.AddCommand(snapshotDeleteCmd)

	// Flags for snapshot create
	snapshotCmd.Flags().StringVar(&snapshotDescription, "description", "", "snapshot description")
}

func runSnapshotCreate(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Creating Database Snapshot")

	// Load configuration
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}

	// Get project directory
	projectDir := getProjectDir()

	// Create snapshot manager
	snapMgr := snapshot.NewManager(cfg, projectDir)

	// Create snapshot
	ui.Info("Exporting database...")
	filename, err := snapMgr.CreateSnapshot(cfg.Project.Name, "manual")
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Get full path
	snapshotDir := expandPath(cfg.Snapshots.Directory)
	fullPath := filepath.Join(snapshotDir, filename)

	ui.Success("Snapshot created successfully!")
	ui.Info("  File: %s", filename)
	ui.Info("  Path: %s", fullPath)
	ui.Info("  Type: manual")
	ui.Info("  Retention: %d days", cfg.Snapshots.Retention.Manual)

	return nil
}

func runSnapshotList(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Database Snapshots")

	// Load configuration
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}

	// Get project directory
	projectDir := getProjectDir()

	// Create snapshot manager
	snapMgr := snapshot.NewManager(cfg, projectDir)

	// List snapshots
	snapshots, err := snapMgr.ListSnapshots(cfg.Project.Name)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		ui.Info("No snapshots found for this project.")
		ui.Info("\nTo create a snapshot, run: stax db snapshot")
		return nil
	}

	// Display snapshots
	ui.Info("Found %d snapshot(s):\n", len(snapshots)

	for _, snap := range snapshots {
		// Format size
		size := formatSize(snap.Size)

		// Format timestamp
		age := time.Since(snap.Timestamp)
		ageStr := formatDuration(age)

		// Display snapshot info
		ui.Info("  %s", snap.File)
		ui.Info("    Type: %s", snap.Type)
		ui.Info("    Size: %s", size)
		ui.Info("    Created: %s (%s ago)", snap.Timestamp.Format("2006-01-02 15:04:05"), ageStr)
		if snap.Description != "" {
			ui.Info("    Description: %s", snap.Description)
		}
		ui.Info("")
	}

	// Show retention policy
	ui.Info("Retention Policy:")
	ui.Info("  Auto snapshots: %d days", cfg.Snapshots.Retention.Auto)
	ui.Info("  Manual snapshots: %d days", cfg.Snapshots.Retention.Manual)

	return nil
}

func runSnapshotRestore(cmd *cobra.Command, args []string) error {
	snapshotName := args[0]

	ui.PrintHeader("Restoring Database Snapshot")

	// Load configuration
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}

	// Get project directory
	projectDir := getProjectDir()

	// Warning
	ui.Warning("This will replace your current database!")
	ui.Info("Snapshot: %s", snapshotName)
	ui.Info("")

	if !ui.Confirm("Are you sure you want to continue?") {
		ui.Info("Snapshot restore cancelled")
		return nil
	}

	// Create snapshot manager
	snapMgr := snapshot.NewManager(cfg, projectDir)

	// Restore snapshot
	ui.Info("Restoring database from snapshot...")

	// Handle both full path and filename
	snapshotPath := snapshotName
	if !filepath.IsAbs(snapshotName) {
		snapshotDir := expandPath(cfg.Snapshots.Directory)
		snapshotPath = filepath.Join(snapshotDir, snapshotName)
	}

	if err := snapMgr.RestoreSnapshot(snapshotPath); err != nil {
		return fmt.Errorf("failed to restore snapshot: %w", err)
	}

	ui.Success("Database restored successfully!")
	ui.Info("Restored from: %s", snapshotName)

	return nil
}

func runSnapshotClean(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Cleaning Old Snapshots")

	// Load configuration
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}

	// Get project directory
	projectDir := getProjectDir()

	// Create snapshot manager
	snapMgr := snapshot.NewManager(cfg, projectDir)

	// Show retention policy
	ui.Info("Retention Policy:")
	ui.Info("  Auto snapshots: %d days", cfg.Snapshots.Retention.Auto)
	ui.Info("  Manual snapshots: %d days", cfg.Snapshots.Retention.Manual)
	ui.Info("")

	// Clean snapshots
	ui.Info("Cleaning old snapshots...")
	if err := snapMgr.CleanSnapshots(cfg); err != nil {
		return fmt.Errorf("failed to clean snapshots: %w", err)
	}

	ui.Success("Snapshot cleanup completed!")

	return nil
}

func runSnapshotDelete(cmd *cobra.Command, args []string) error {
	snapshotName := args[0]

	ui.PrintHeader("Deleting Snapshot")

	// Load configuration
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}

	// Get project directory
	projectDir := getProjectDir()

	// Warning
	ui.Warning("This will permanently delete the snapshot!")
	ui.Info("Snapshot: %s", snapshotName)
	ui.Info("")

	if !ui.Confirm("Are you sure you want to continue?") {
		ui.Info("Snapshot deletion cancelled")
		return nil
	}

	// Create snapshot manager
	snapMgr := snapshot.NewManager(cfg, projectDir)

	// Delete snapshot
	ui.Info("Deleting snapshot...")
	if err := snapMgr.DeleteSnapshot(snapshotName); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	ui.Success("Snapshot deleted successfully!")
	ui.Info("Deleted: %s", snapshotName)

	return nil
}

// formatSize formats bytes as human-readable size
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats a duration as human-readable string
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%d days", days)
	}
	hours := int(d.Hours())
	if hours > 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	minutes := int(d.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	return "less than a minute"
}
