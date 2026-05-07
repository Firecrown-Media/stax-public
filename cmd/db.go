package cmd

import (
	"github.com/firecrown-media/stax/pkg/database"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

// dbCmd represents the db command group
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "✓ Database operations",
	Long:  `Manage database operations including pull, push, import, export, snapshots, and queries.`,
}

var (
	dbEnvironment    string
	dbSnapshot       bool
	dbSanitize       bool
	dbSkipReplace    bool
	dbExcludeTables  string
	dbSkipLogs       bool
	dbSkipTransients bool
	dbSkipSpam       bool
	dbDryRun         bool
	dbSkipBackup     bool
)

// dbPullCmd represents the db:pull command
var dbPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull database from WPEngine",
	Long: `Pull database from WPEngine, import it locally, and run search-replace.

This command will:
  - Create a snapshot of the current database (unless --snapshot=false)
  - Connect to WPEngine SSH Gateway
  - Export the database from WPEngine
  - Transfer the database to local environment
  - Import into local DDEV database
  - Run search-replace operations (unless --skip-replace)
  - Flush WordPress cache`,
	Example: `  # Basic pull
  stax db pull

  # Pull from staging
  stax db pull --environment=staging

  # Pull without creating snapshot
  stax db pull --snapshot=false

  # Pull without automatic URL replacement (advanced users)
  stax db pull --skip-replace

  # Pull with sanitized data
  stax db pull --sanitize`,
	RunE: runDBPull,
}

// dbPushCmd represents the db:push command
var dbPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push database to WPEngine",
	Long: `Push local database to WPEngine environment.

This command will:
  - Export the local database from DDEV
  - Run search-replace to update URLs for the target environment
  - Upload the database to WPEngine
  - Import the database on WPEngine
  - Clean up temporary files

WARNING: This will overwrite the database on the target environment!`,
	Example: `  # Push to staging
  stax db push --environment=staging

  # Push to production (requires confirmation)
  stax db push --environment=production

  # Dry run to see what would happen
  stax db push --environment=staging --dry-run

  # Push without creating remote backup
  stax db push --environment=staging --skip-backup

  # Push without URL replacement
  stax db push --environment=staging --skip-replace`,
	RunE: runDBPush,
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbPullCmd)
	dbCmd.AddCommand(dbPushCmd)

	// Flags for pull
	dbPullCmd.Flags().StringVar(&dbEnvironment, "environment", "", "WPEngine environment (default: from config)")
	dbPullCmd.Flags().BoolVar(&dbSnapshot, "snapshot", true, "create snapshot before import")
	dbPullCmd.Flags().BoolVar(&dbSanitize, "sanitize", false, "sanitize user data")
	dbPullCmd.Flags().BoolVar(&dbSkipReplace, "skip-replace", false, "skip automatic URL search-replace")
	dbPullCmd.Flags().StringVar(&dbExcludeTables, "exclude-tables", "", "comma-separated tables to exclude")
	dbPullCmd.Flags().BoolVar(&dbSkipLogs, "skip-logs", true, "skip log tables")
	dbPullCmd.Flags().BoolVar(&dbSkipTransients, "skip-transients", true, "skip transient tables")
	dbPullCmd.Flags().BoolVar(&dbSkipSpam, "skip-spam", true, "skip spam/trash")

	// Flags for push
	dbPushCmd.Flags().StringVar(&dbEnvironment, "environment", "", "WPEngine environment (required: staging or production)")
	_ = dbPushCmd.MarkFlagRequired("environment") // Flag existence guaranteed in init
	dbPushCmd.Flags().BoolVar(&dbDryRun, "dry-run", false, "show what would happen without pushing")
	dbPushCmd.Flags().BoolVar(&dbSkipBackup, "skip-backup", false, "skip creating remote backup before import")
	dbPushCmd.Flags().BoolVar(&dbSkipReplace, "skip-replace", false, "skip automatic URL search-replace")
}

func runDBPull(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}
	return database.Pull(cfg, database.PullOptions{
		Environment:    dbEnvironment,
		Snapshot:       dbSnapshot,
		Sanitize:       dbSanitize,
		SkipReplace:    dbSkipReplace,
		ExcludeTables:  dbExcludeTables,
		SkipLogs:       dbSkipLogs,
		SkipTransients: dbSkipTransients,
		SkipSpam:       dbSkipSpam,
		ProjectDir:     getProjectDir(),
	})
}

func runDBPush(cmd *cobra.Command, args []string) error {
	if dbEnvironment == "production" && !dbDryRun {
		ui.Warning("You are about to push the local database to PRODUCTION!")
		if !ui.Confirm("Are you absolutely sure you want to continue?") {
			return nil
		}
		if !ui.Confirm("This is your last chance. Proceed?") {
			return nil
		}
	}
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}
	return database.Push(cfg, database.PushOptions{
		Environment: dbEnvironment,
		DryRun:      dbDryRun,
		SkipBackup:  dbSkipBackup,
		SkipReplace: dbSkipReplace,
		ProjectDir:  getProjectDir(),
	})
}

// loadConfigForCommand is now defined in files.go to avoid duplication
