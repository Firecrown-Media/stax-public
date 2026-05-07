package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/snapshot"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/firecrown-media/stax/pkg/wordpress"
	"github.com/firecrown-media/stax/pkg/wpengine"
)

// PullOptions controls the behaviour of a database pull.
type PullOptions struct {
	// ProjectDir is the local project directory (where .stax.yml lives).
	ProjectDir string

	// Environment overrides the environment in provider_config when non-empty.
	Environment string

	// Snapshot controls whether to snapshot before pull.
	Snapshot bool

	// SkipReplace skips URL search-replace when true.
	SkipReplace bool

	// Sanitize enables data sanitisation.
	Sanitize bool

	// ExcludeTables is a comma-separated list of tables to exclude.
	ExcludeTables string

	// SkipLogs excludes log tables.
	SkipLogs bool

	// SkipTransients excludes transient tables.
	SkipTransients bool

	// SkipSpam excludes spam/trash.
	SkipSpam bool

	// AutoSnapshotBeforePull mirrors cfg.Snapshots.AutoSnapshotBeforePull.
	AutoSnapshotBeforePull bool
}

// PushOptions controls the behaviour of a database push.
type PushOptions struct {
	// ProjectDir is the local project directory.
	ProjectDir string

	// Environment is the target environment (staging|production).
	Environment string

	// DryRun shows what would happen without making changes.
	DryRun bool

	// SkipBackup skips creating a remote backup before import.
	SkipBackup bool

	// SkipReplace skips URL search-replace when true.
	SkipReplace bool
}

// Pull pulls a database from WPEngine, imports it locally, and runs search-replace.
func Pull(cfg *config.Config, opts PullOptions) error {
	install := providerConfigString(cfg.ProviderConfig, "install")

	// Get WPEngine credentials with fallback.
	creds, err := credentials.GetWPEngineCredentialsWithFallback(install)
	if err != nil {
		if credErr, ok := err.(*credentials.CredentialsNotFoundError); ok {
			return errors.NewCredentialsNotFoundError(credErr.Tried, credErr.LastErr)
		}
		return fmt.Errorf("failed to get WPEngine credentials: %w", err)
	}

	// Get SSH key with fallback.
	sshKey, err := credentials.GetSSHPrivateKeyWithFallback("wpengine")
	if err != nil {
		if keyErr, ok := err.(*credentials.SSHKeyNotFoundError); ok {
			return errors.NewSSHKeyNotFoundError("", keyErr.Tried, keyErr.LastErr)
		}
		return fmt.Errorf("failed to get SSH key: %w", err)
	}

	projectDir := opts.ProjectDir

	// Verify .stax.yml exists in the project directory.
	staxConfigPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(staxConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("no .stax.yml found in %s. Please run 'stax init' first or use --project-dir to specify the correct directory", projectDir)
	}

	mgr := ddev.NewManager(projectDir)

	// Check DDEV status.
	if err := mgr.RequireRunning(); err != nil {
		return err
	}

	// Create snapshot if requested.
	shouldSnapshot := opts.Snapshot || opts.AutoSnapshotBeforePull
	if shouldSnapshot {
		ui.Info("Creating database snapshot...")
		snapMgr := snapshot.NewManager(cfg, projectDir)
		snapType := "manual"
		if opts.AutoSnapshotBeforePull && !opts.Snapshot {
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

	// Create SSH client.
	ui.Info("Connecting to WPEngine SSH Gateway...")
	sshConfig := wpengine.SSHConfig{
		Host:       providerConfigString(cfg.ProviderConfig, "ssh_gateway"),
		Port:       22,
		User:       creds.SSHUser,
		PrivateKey: sshKey,
		Install:    install,
	}

	sshClient, err := wpengine.NewSSHClient(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to WPEngine: %w", err)
	}
	defer sshClient.Close()

	ui.Success("Connected to WPEngine")

	// Build database export options.
	dbOptions := wpengine.DatabaseOptions{
		SkipLogs:       opts.SkipLogs,
		SkipTransients: opts.SkipTransients,
		SkipSpam:       opts.SkipSpam,
	}

	// Export database.
	ui.Info("Exporting database from WPEngine...")
	dbReader, err := sshClient.ExportDatabase(dbOptions)
	if err != nil {
		return fmt.Errorf("failed to export database: %w", err)
	}
	defer dbReader.Close()

	// Stream to a temp file.
	tmpFile, err := os.CreateTemp("", "wpengine-db-*.sql")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	dbPath := tmpFile.Name()
	defer os.Remove(dbPath)

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

	// Import to local database.
	ui.Info("Importing database to local environment...")
	if err := mgr.ImportDB(dbPath); err != nil {
		return fmt.Errorf("database import failed: %w", err)
	}
	ui.Success("Database imported")

	// Wait for database to be ready.
	ui.Info("Waiting for database to be ready...")
	if err := mgr.WaitForDatabaseReady(60 * time.Second); err != nil {
		ui.Warning(fmt.Sprintf("Database health check: %v", err))
		ui.Info("Proceeding with search-replace anyway - it may fail if database is not ready")
	}

	// Run search-replace unless skipped.
	environment := providerConfigString(cfg.ProviderConfig, "environment")
	if opts.Environment != "" {
		environment = opts.Environment
	}

	if !opts.SkipReplace {
		ui.Info("Replacing URLs...")

		targetURL := GetDDEVURL(cfg.Project.Name)
		cli := wordpress.NewCLI(projectDir)
		sourceURL := ""

		actualURL, err := cli.GetOption("siteurl", 0)
		if err == nil && actualURL != "" && actualURL != targetURL {
			ui.Info(fmt.Sprintf("Detected source URL: %s", actualURL))
			sourceURL = actualURL
		} else {
			sourceURL = GetWPEngineURL(install, environment, "")
			ui.Debug(fmt.Sprintf("Using pattern-based source URL: %s", sourceURL))
		}

		if err := RunSearchReplace(projectDir, sourceURL, targetURL, cfg); err != nil {
			ui.Warning(fmt.Sprintf("URL replacement failed: %v", err))
			ui.Info("You may need to run manually: ddev wp search-replace '%s' '%s' --all-tables", sourceURL, targetURL)
		} else {
			ui.Success("URLs replaced successfully")
		}
	} else {
		ui.Info("Skipping URL replacement (--skip-replace flag set)")
		sourceURL := GetWPEngineURL(install, environment, "")
		targetURL := GetDDEVURL(cfg.Project.Name)
		ui.Info("To replace URLs manually, run: ddev wp search-replace '%s' '%s' --all-tables", sourceURL, targetURL)
	}

	// Flush cache.
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

// Push exports the local database, optionally runs search-replace, and imports it on WPEngine.
func Push(cfg *config.Config, opts PushOptions) error {
	install := providerConfigString(cfg.ProviderConfig, "install")

	projectDir := opts.ProjectDir

	// Verify .stax.yml exists.
	staxConfigPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(staxConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("no .stax.yml found in %s. Please run 'stax init' first or use --project-dir to specify the correct directory", projectDir)
	}

	mgr := ddev.NewManager(projectDir)

	// Check DDEV status.
	if err := mgr.RequireRunning(); err != nil {
		return err
	}

	// Get credentials.
	creds, err := credentials.GetWPEngineCredentialsWithFallback(install)
	if err != nil {
		if credErr, ok := err.(*credentials.CredentialsNotFoundError); ok {
			return errors.NewCredentialsNotFoundError(credErr.Tried, credErr.LastErr)
		}
		return fmt.Errorf("failed to get WPEngine credentials: %w", err)
	}

	// Get SSH key.
	sshKey, err := credentials.GetSSHPrivateKeyWithFallback("wpengine")
	if err != nil {
		if keyErr, ok := err.(*credentials.SSHKeyNotFoundError); ok {
			return errors.NewSSHKeyNotFoundError("", keyErr.Tried, keyErr.LastErr)
		}
		return fmt.Errorf("failed to get SSH key: %w", err)
	}

	ui.Info(fmt.Sprintf("Environment: %s", opts.Environment))
	ui.Info(fmt.Sprintf("Install: %s", install))

	if opts.DryRun {
		ui.Info("\n=== DRY RUN MODE ===")
		ui.Info("The following operations would be performed:")
		ui.Info("  1. Export local database from DDEV")
		ui.Info("  2. Run search-replace: %s -> %s", GetDDEVURL(cfg.Project.Name), getTargetURL(cfg, opts.Environment))
		if !opts.SkipBackup {
			ui.Info("  3. Create backup on WPEngine %s environment", opts.Environment)
		}
		ui.Info("  4. Upload database to WPEngine")
		ui.Info("  5. Import database on WPEngine %s environment", opts.Environment)
		ui.Info("  6. Clean up temporary files")
		ui.Info("\nNo changes will be made in dry-run mode.")
		return nil
	}

	// Export local database.
	ui.Info("Exporting local database...")
	tmpDBPath := fmt.Sprintf("/tmp/stax-db-push-%d.sql", os.Getpid())
	defer os.Remove(tmpDBPath)

	if err := mgr.ExportDB(tmpDBPath); err != nil {
		return fmt.Errorf("failed to export local database: %w", err)
	}
	ui.Success("Database exported")

	// Connect to WPEngine.
	ui.Info("Connecting to WPEngine SSH Gateway...")
	sshCfg := wpengine.SSHConfig{
		Host:       providerConfigString(cfg.ProviderConfig, "ssh_gateway"),
		Port:       22,
		User:       creds.SSHUser,
		PrivateKey: sshKey,
		Install:    install,
	}

	sshClient, err := wpengine.NewSSHClient(sshCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to WPEngine: %w", err)
	}
	defer sshClient.Close()
	ui.Success("Connected to WPEngine")

	// Create remote backup unless skipped.
	if !opts.SkipBackup {
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

	// Upload database file.
	ui.Info("Uploading database to WPEngine...")
	remoteDBPath := fmt.Sprintf("~/stax-db-push-%d.sql", os.Getpid())

	if err := sshClient.UploadFile(tmpDBPath, remoteDBPath); err != nil {
		return fmt.Errorf("failed to upload database: %w", err)
	}
	defer func() {
		_ = sshClient.RemoveFile(remoteDBPath)
	}()
	ui.Success("Database uploaded")

	// Import database on WPEngine.
	ui.Info("Importing database on WPEngine...")
	if err := sshClient.ImportDatabase(remoteDBPath); err != nil {
		return fmt.Errorf("database import failed: %w", err)
	}
	ui.Success("Database imported")

	// Run search-replace on WPEngine unless skipped.
	if !opts.SkipReplace {
		ui.Info("Running search-replace on WPEngine...")

		sourceURL := GetDDEVURL(cfg.Project.Name)
		targetURL := getTargetURL(cfg, opts.Environment)

		ui.Info(fmt.Sprintf("  Replacing: %s -> %s", sourceURL, targetURL))

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

	// Flush cache on WPEngine.
	ui.Info("Flushing WordPress cache on WPEngine...")
	if _, err := sshClient.ExecuteCommand("wp cache flush"); err != nil {
		ui.Warning(fmt.Sprintf("Cache flush failed: %v", err))
	} else {
		ui.Success("Cache flushed")
	}

	ui.Success("\nDatabase push completed!")
	ui.Info(fmt.Sprintf("Database successfully pushed to %s environment", opts.Environment))

	return nil
}

// GetDDEVURL returns the local DDEV URL for the given project name.
func GetDDEVURL(projectName string) string {
	return fmt.Sprintf("https://%s.ddev.site", projectName)
}

// GetWPEngineURL returns the WPEngine URL for an install/environment combination.
// If customDomain is non-empty it is used directly.
func GetWPEngineURL(install, environment, customDomain string) string {
	if customDomain != "" {
		return "https://" + customDomain
	}

	switch environment {
	case "production":
		return fmt.Sprintf("https://%s.wpengine.com", install)
	case "staging":
		if strings.HasSuffix(install, "stg") ||
			strings.HasSuffix(install, "stage") ||
			strings.HasSuffix(install, "staging") ||
			strings.Contains(install, "stag") {
			return fmt.Sprintf("https://%s.wpengine.com", install)
		}
		return fmt.Sprintf("https://%s.wpengine.com", install)
	case "development":
		return fmt.Sprintf("https://%s.wpengine.com", install)
	default:
		return fmt.Sprintf("https://%s.wpengine.com", install)
	}
}

// RunSearchReplace executes wp search-replace via DDEV and verifies the result.
func RunSearchReplace(projectDir, from, to string, cfg *config.Config) error {
	cli := wordpress.NewCLI(projectDir)

	ui.Info(fmt.Sprintf("Replacing URLs: %s -> %s", from, to))

	isMultisite := cfg.Project.Type == "wordpress-multisite"

	if isMultisite {
		ui.Info("Detected multisite installation - running network-wide search-replace")

		opts := wordpress.SearchReplaceOptions{
			Network:     true,
			SkipColumns: []string{"guid"},
			DryRun:      false,
		}

		if err := cli.SearchReplaceWithOptions(from, to, opts); err != nil {
			return fmt.Errorf("multisite search-replace failed: %w", err)
		}

		if cfg.Project.Mode == "subdomain" && len(cfg.Network.Sites) > 0 {
			ui.Info("Running additional search-replace for subdomain sites")

			for _, site := range cfg.Network.Sites {
				if !site.Active {
					continue
				}

				if site.ProviderDomain != "" && site.Domain != "" {
					siteFrom := "https://" + site.ProviderDomain
					siteTo := "https://" + site.Domain

					ui.Info(fmt.Sprintf("Site %s: %s -> %s", site.Name, siteFrom, siteTo))

					siteOpts := wordpress.SearchReplaceOptions{
						Network:     false,
						URL:         site.Domain,
						SkipColumns: []string{"guid"},
						DryRun:      false,
					}

					if err := cli.SearchReplaceWithOptions(siteFrom, siteTo, siteOpts); err != nil {
						ui.Warning(fmt.Sprintf("Search-replace failed for site %s: %v", site.Name, err))
					}
				}
			}
		}
	} else {
		ui.Info("Detected single-site installation - running standard search-replace")

		opts := wordpress.SearchReplaceOptions{
			Network:     false,
			SkipColumns: []string{"guid"},
			DryRun:      false,
		}

		if err := cli.SearchReplaceWithOptions(from, to, opts); err != nil {
			return fmt.Errorf("search-replace failed: %w", err)
		}
	}

	// Verify the siteurl was updated.
	ui.Info("Verifying URL replacement...")
	matches, actualURL, err := cli.VerifySiteURL(to)
	if err != nil {
		return fmt.Errorf("failed to verify siteurl: %w", err)
	}

	if !matches {
		return fmt.Errorf("URL replacement verification failed: expected siteurl to be '%s' but found '%s'. The search-replace may have completed without errors but did not update the URLs correctly", to, actualURL)
	}

	ui.Success(fmt.Sprintf("Verified: siteurl is correctly set to %s", to))

	return nil
}

// getTargetURL returns the WPEngine URL for the given push environment.
func getTargetURL(cfg *config.Config, environment string) string {
	install := providerConfigString(cfg.ProviderConfig, "install")
	return GetWPEngineURL(install, environment, "")
}

// providerConfigString safely extracts a string value from provider_config.
func providerConfigString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
