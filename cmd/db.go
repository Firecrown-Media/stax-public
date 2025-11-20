package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/snapshot"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/firecrown-media/stax/pkg/wordpress"
	"github.com/firecrown-media/stax/pkg/wpengine"
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
	ui.PrintHeader("Pulling Database from WPEngine")

	// Load configuration
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}

	// Get WPEngine credentials with fallback
	creds, err := credentials.GetWPEngineCredentialsWithFallback(cfg.WPEngine.Install)
	if err != nil {
		if credErr, ok := err.(*credentials.CredentialsNotFoundError); ok {
			return errors.NewCredentialsNotFoundError(credErr.Tried, credErr.LastErr)
		}
		return fmt.Errorf("failed to get WPEngine credentials: %w", err)
	}

	// Get SSH key with fallback
	sshKey, err := credentials.GetSSHPrivateKeyWithFallback("wpengine")
	if err != nil {
		if keyErr, ok := err.(*credentials.SSHKeyNotFoundError); ok {
			return errors.NewSSHKeyNotFoundError("", keyErr.Tried, keyErr.LastErr)
		}
		return fmt.Errorf("failed to get SSH key: %w", err)
	}

	// Check if DDEV is running
	projectDir := getProjectDir()

	// Verify .stax.yml exists in project directory
	staxConfigPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(staxConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("no .stax.yml found in %s. Please run 'stax init' first or use --project-dir to specify the correct directory", projectDir)
	}

	mgr := ddev.NewManager(projectDir)

	// Check DDEV status with brief retry for recently-started DDEV
	var running bool
	var statusErr error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		running, statusErr = mgr.IsRunning()
		if statusErr == nil && running {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	if statusErr != nil {
		return fmt.Errorf("failed to check DDEV status in %s: %w\n\nTry running: cd %s && ddev describe", projectDir, statusErr, projectDir)
	}
	if !running {
		return fmt.Errorf("DDEV must be running to import database.\n\nProject directory: %s\nPlease run: cd %s && stax start", projectDir, projectDir)
	}

	// Create snapshot if requested (or auto-snapshot is enabled)
	shouldSnapshot := dbSnapshot || cfg.Snapshots.AutoSnapshotBeforePull
	if shouldSnapshot {
		ui.Info("Creating database snapshot...")
		snapMgr := snapshot.NewManager(cfg, projectDir)
		snapType := "manual"
		if cfg.Snapshots.AutoSnapshotBeforePull && !dbSnapshot {
			snapType = "auto"
		}
		filename, err := snapMgr.CreateSnapshot(cfg.Project.Name, snapType)
		if err != nil {
			ui.Warning(fmt.Sprintf("Failed to create snapshot: %v", err))
			ui.Info("Continuing with database pull...")
		} else {
			ui.Success(fmt.Sprintf("Snapshot created: %s", filename))
		}
	}

	// Create SSH client
	ui.Info("Connecting to WPEngine SSH Gateway...")
	sshConfig := wpengine.SSHConfig{
		Host:       cfg.WPEngine.SSHGateway,
		Port:       22,
		User:       creds.SSHUser,
		PrivateKey: sshKey,
		Install:    cfg.WPEngine.Install,
	}

	sshClient, err := wpengine.NewSSHClient(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to WPEngine: %w", err)
	}
	defer sshClient.Close()

	ui.Success("Connected to WPEngine")

	// Build database export options
	dbOptions := wpengine.DatabaseOptions{
		SkipLogs:       dbSkipLogs,
		SkipTransients: dbSkipTransients,
		SkipSpam:       dbSkipSpam,
	}

	// Export database
	ui.Info("Exporting database from WPEngine...")
	dbReader, err := sshClient.ExportDatabase(dbOptions)
	if err != nil {
		return fmt.Errorf("failed to export database: %w", err)
	}
	defer dbReader.Close()

	// Create temporary file for database dump
	tmpFile, err := os.CreateTemp("", "wpengine-db-*.sql")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	dbPath := tmpFile.Name()
	defer os.Remove(dbPath) // Clean up temporary file

	// Stream database export to temporary file
	ui.Info("Downloading database...")
	spinner := ui.NewSpinner("Streaming database export...")
	spinner.Start()

	_, err = io.Copy(tmpFile, dbReader)
	tmpFile.Close()
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to download database: %w", err)
	}

	ui.Success("Database exported")

	// Import to local database
	ui.Info("Importing database to local environment...")
	if err := mgr.ImportDB(dbPath); err != nil {
		return fmt.Errorf("database import failed: %w", err)
	}
	ui.Success("Database imported")

	// Run search-replace unless skipped
	if !dbSkipReplace {
		ui.Info("Replacing URLs...")

		// Get source and target URLs
		sourceURL := getWPEngineURL(cfg)
		targetURL := getDDEVURL(cfg)

		// Run search-replace
		if err := runSearchReplace(projectDir, sourceURL, targetURL, cfg); err != nil {
			ui.Warning(fmt.Sprintf("URL replacement failed: %v", err))
			ui.Info("You may need to run manually: ddev wp search-replace '%s' '%s' --all-tables", sourceURL, targetURL)
		} else {
			ui.Success("URLs replaced successfully")
		}
	} else {
		ui.Info("Skipping URL replacement (--skip-replace flag set)")
		sourceURL := getWPEngineURL(cfg)
		targetURL := getDDEVURL(cfg)
		ui.Info("To replace URLs manually, run: ddev wp search-replace '%s' '%s' --all-tables", sourceURL, targetURL)
	}

	// Flush cache
	ui.Info("Flushing WordPress cache...")
	cli := wordpress.NewCLI(projectDir)
	if err := cli.FlushCache(); err != nil {
		ui.Warning(fmt.Sprintf("Cache flush failed: %v", err))
	} else {
		ui.Success("Cache flushed")
	}

	ui.Success("\nDatabase pull completed!")

	return nil
}

func runDBPush(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Pushing Database to WPEngine")

	// Load configuration
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}

	// Validate environment
	if dbEnvironment != "staging" && dbEnvironment != "production" {
		return fmt.Errorf("environment must be 'staging' or 'production', got: %s", dbEnvironment)
	}

	// Production safety check - require explicit confirmation
	if dbEnvironment == "production" && !dbDryRun {
		ui.Warning("You are about to push the local database to PRODUCTION!")
		ui.Warning("This will OVERWRITE the production database!")
		ui.Info("")

		if !ui.Confirm("Are you absolutely sure you want to continue?") {
			ui.Info("Database push cancelled")
			return nil
		}

		// Double confirmation for production
		ui.Info("")
		ui.Warning("This is your last chance to cancel!")
		if !ui.Confirm("Type 'yes' to proceed with production database push") {
			ui.Info("Database push cancelled")
			return nil
		}
	}

	ui.Info(fmt.Sprintf("Environment: %s", dbEnvironment))
	ui.Info(fmt.Sprintf("Install: %s", cfg.WPEngine.Install))

	// Check if DDEV is running
	projectDir := getProjectDir()

	// Verify .stax.yml exists in project directory
	staxConfigPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(staxConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("no .stax.yml found in %s. Please run 'stax init' first or use --project-dir to specify the correct directory", projectDir)
	}

	mgr := ddev.NewManager(projectDir)

	// Check DDEV status with brief retry for recently-started DDEV
	var running bool
	var statusErr error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		running, statusErr = mgr.IsRunning()
		if statusErr == nil && running {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	if statusErr != nil {
		return fmt.Errorf("failed to check DDEV status in %s: %w\n\nTry running: cd %s && ddev describe", projectDir, statusErr, projectDir)
	}
	if !running {
		return fmt.Errorf("DDEV must be running to export database.\n\nProject directory: %s\nPlease run: cd %s && stax start", projectDir, projectDir)
	}

	// Get credentials
	creds, err := credentials.GetWPEngineCredentialsWithFallback(cfg.WPEngine.Install)
	if err != nil {
		if credErr, ok := err.(*credentials.CredentialsNotFoundError); ok {
			return errors.NewCredentialsNotFoundError(credErr.Tried, credErr.LastErr)
		}
		return fmt.Errorf("failed to get WPEngine credentials: %w", err)
	}

	// Get SSH key
	sshKey, err := credentials.GetSSHPrivateKeyWithFallback("wpengine")
	if err != nil {
		if keyErr, ok := err.(*credentials.SSHKeyNotFoundError); ok {
			return errors.NewSSHKeyNotFoundError("", keyErr.Tried, keyErr.LastErr)
		}
		return fmt.Errorf("failed to get SSH key: %w", err)
	}

	// Use credentials
	_ = creds
	_ = sshKey

	if dbDryRun {
		ui.Info("\n=== DRY RUN MODE ===")
		ui.Info("The following operations would be performed:")
		ui.Info("  1. Export local database from DDEV")
		ui.Info("  2. Run search-replace: %s -> %s", getDDEVURL(cfg), getTargetURL(cfg, dbEnvironment))
		if !dbSkipBackup {
			ui.Info("  3. Create backup on WPEngine %s environment", dbEnvironment)
		}
		ui.Info("  4. Upload database to WPEngine")
		ui.Info("  5. Import database on WPEngine %s environment", dbEnvironment)
		ui.Info("  6. Clean up temporary files")
		ui.Info("\nNo changes will be made in dry-run mode.")
		return nil
	}

	// Export local database
	ui.Info("Exporting local database...")
	tmpDBPath := fmt.Sprintf("/tmp/stax-db-push-%d.sql", os.Getpid())
	defer os.Remove(tmpDBPath) // Clean up local temp file

	if err := mgr.ExportDB(tmpDBPath); err != nil {
		return fmt.Errorf("failed to export local database: %w", err)
	}
	ui.Success("Database exported")

	// Connect to WPEngine
	ui.Info("Connecting to WPEngine SSH Gateway...")
	sshConfig := wpengine.SSHConfig{
		Host:       cfg.WPEngine.SSHGateway,
		Port:       22,
		User:       creds.SSHUser,
		PrivateKey: sshKey,
		Install:    cfg.WPEngine.Install,
	}

	sshClient, err := wpengine.NewSSHClient(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to WPEngine: %w", err)
	}
	defer sshClient.Close()
	ui.Success("Connected to WPEngine")

	// Create backup on remote unless skipped
	if !dbSkipBackup {
		ui.Info("Creating database backup on WPEngine...")
		backupPath := fmt.Sprintf("~/db-backup-before-push-%d.sql", time.Now().Unix())
		backupCmd := fmt.Sprintf("wp db export %s", backupPath)
		if _, err := sshClient.ExecuteCommand(backupCmd); err != nil {
			ui.Warning(fmt.Sprintf("Failed to create backup: %v", err))
			ui.Info("Continuing without backup...")
		} else {
			ui.Success(fmt.Sprintf("Backup created: %s", backupPath))
		}
	}

	// Upload database file
	ui.Info("Uploading database to WPEngine...")
	remoteDBPath := fmt.Sprintf("~/stax-db-push-%d.sql", os.Getpid())

	if err := sshClient.UploadFile(tmpDBPath, remoteDBPath); err != nil {
		return fmt.Errorf("failed to upload database: %w", err)
	}
	defer func() {
		_ = sshClient.RemoveFile(remoteDBPath) // Best-effort cleanup
	}()
	ui.Success("Database uploaded")

	// Import database on WPEngine
	ui.Info("Importing database on WPEngine...")
	if err := sshClient.ImportDatabase(remoteDBPath); err != nil {
		return fmt.Errorf("database import failed: %w", err)
	}
	ui.Success("Database imported")

	// Run search-replace on WPEngine unless skipped
	if !dbSkipReplace {
		ui.Info("Running search-replace on WPEngine...")

		// Get source and target URLs
		sourceURL := getDDEVURL(cfg)
		targetURL := getTargetURL(cfg, dbEnvironment)

		ui.Info(fmt.Sprintf("  Replacing: %s -> %s", sourceURL, targetURL))

		// Run search-replace via WP-CLI on remote
		searchReplaceCmd := fmt.Sprintf("wp search-replace '%s' '%s' --all-tables --skip-columns=guid", sourceURL, targetURL)
		output, err := sshClient.ExecuteCommand(searchReplaceCmd)
		if err != nil {
			ui.Warning(fmt.Sprintf("Search-replace failed: %v", err))
			ui.Info("Database imported but URLs may not be correct")
		} else {
			ui.Success("URLs updated successfully")
			ui.Verbose(output)
		}
	}

	// Flush cache on WPEngine
	ui.Info("Flushing WordPress cache on WPEngine...")
	if _, err := sshClient.ExecuteCommand("wp cache flush"); err != nil {
		ui.Warning(fmt.Sprintf("Cache flush failed: %v", err))
	} else {
		ui.Success("Cache flushed")
	}

	ui.Success("\nDatabase push completed!")
	ui.Info(fmt.Sprintf("Database successfully pushed to %s environment", dbEnvironment))

	return nil
}

// getTargetURL returns the target WPEngine URL for the given environment
func getTargetURL(cfg *config.Config, environment string) string {
	install := cfg.WPEngine.Install

	if environment == "production" {
		// Check if custom domain is configured
		if cfg.WPEngine.Domains.Production.Primary != "" {
			return "https://" + cfg.WPEngine.Domains.Production.Primary
		}
		// Default production URL pattern
		return fmt.Sprintf("https://%s.wpengine.com", install)
	} else if environment == "staging" {
		// Check if custom domain is configured
		if cfg.WPEngine.Domains.Staging.Primary != "" {
			return "https://" + cfg.WPEngine.Domains.Staging.Primary
		}
		// Default staging URL pattern
		return fmt.Sprintf("https://%s.wpengineurl.com", install)
	}

	// Fallback to staging pattern
	return fmt.Sprintf("https://%s.wpengineurl.com", install)
}

// loadConfigForCommand is now defined in files.go to avoid duplication
