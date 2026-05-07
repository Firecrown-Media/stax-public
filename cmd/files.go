package cmd

import (
	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/files"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/spf13/cobra"
)

// filesCmd represents the files command group
var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "✓ File synchronization operations",
	Long:  `Manage file synchronization between WPEngine and local environment including themes, plugins, and uploads.`,
}

var (
	filesEnvironment         string
	filesThemesOnly          bool
	filesPluginsOnly         bool
	filesMuPluginsOnly       bool
	filesUploadsOnly         bool
	filesExcludeUploads      bool
	filesDryRun              bool
	filesDelete              bool
	filesBandwidthLimit      int
	filesInclude             string
	filesExclude             string
	filesVerify              bool
	filesPreservePermissions bool
)

// filesPullCmd represents the files:pull command
var filesPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull files from WPEngine",
	Long: `Pull files from WPEngine to your local environment.

This command will:
  - Connect to WPEngine via SSH
  - Sync wp-content directory (or specific subdirectories)
  - Transfer files using rsync over SSH
  - Verify file integrity after transfer

By default, this syncs the entire wp-content directory. Use flags to
limit the sync to specific directories like themes or plugins.`,
	Example: `  # Basic pull (all wp-content)
  stax files pull

  # Pull only themes
  stax files pull --themes-only

  # Pull only plugins
  stax files pull --plugins-only

  # Pull only mu-plugins
  stax files pull --mu-plugins-only

  # Pull without uploads directory
  stax files pull --exclude-uploads

  # Pull from staging environment
  stax files pull --environment=staging

  # Dry run to see what would be synced
  stax files pull --dry-run

  # Delete local files not on remote
  stax files pull --delete

  # Limit bandwidth to 1000 KB/s
  stax files pull --bandwidth-limit=1000

  # Preserve file permissions
  stax files pull --preserve-permissions

  # Custom includes and excludes
  stax files pull --include="*.php,*.js" --exclude="*.log,cache/"`,
	RunE: runFilesPull,
}

// filesPushCmd represents the files:push command
var filesPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push files to WPEngine",
	Long: `Push files from your local environment to WPEngine.

This command will:
  - Connect to WPEngine via SSH
  - Sync local wp-content directory to remote (or specific subdirectories)
  - Transfer files using rsync over SSH
  - Verify file integrity after transfer

By default, this syncs the entire wp-content directory. Use flags to
limit the sync to specific directories like themes or plugins.

WARNING: This operation modifies files on the remote server. Always use
--dry-run first to preview changes, especially when using --delete mode.`,
	Example: `  # Dry run to see what would be pushed
  stax files push --dry-run

  # Basic push (all wp-content)
  stax files push

  # Push only themes
  stax files push --themes-only

  # Push only plugins
  stax files push --plugins-only

  # Push only mu-plugins
  stax files push --mu-plugins-only

  # Push only uploads
  stax files push --uploads-only

  # Push to staging environment
  stax files push --environment=staging

  # Delete remote files not in local (DANGEROUS!)
  stax files push --delete

  # Limit bandwidth to 1000 KB/s
  stax files push --bandwidth-limit=1000

  # Preserve file permissions
  stax files push --preserve-permissions

  # Custom includes and excludes
  stax files push --include="*.php,*.js" --exclude="*.log,cache/"`,
	RunE: runFilesPush,
}

func init() {
	rootCmd.AddCommand(filesCmd)
	filesCmd.AddCommand(filesPullCmd)
	filesCmd.AddCommand(filesPushCmd)

	// Flags for pull
	filesPullCmd.Flags().StringVar(&filesEnvironment, "environment", "", "WPEngine environment (default: from config)")
	filesPullCmd.Flags().BoolVar(&filesThemesOnly, "themes-only", false, "sync only themes directory")
	filesPullCmd.Flags().BoolVar(&filesPluginsOnly, "plugins-only", false, "sync only plugins directory")
	filesPullCmd.Flags().BoolVar(&filesMuPluginsOnly, "mu-plugins-only", false, "sync only mu-plugins directory")
	filesPullCmd.Flags().BoolVar(&filesExcludeUploads, "exclude-uploads", false, "exclude uploads directory")
	filesPullCmd.Flags().BoolVar(&filesDryRun, "dry-run", false, "show what would be transferred without syncing")
	filesPullCmd.Flags().BoolVar(&filesDelete, "delete", false, "delete local files not present on remote")
	filesPullCmd.Flags().IntVar(&filesBandwidthLimit, "bandwidth-limit", 0, "bandwidth limit in KB/s (0 = unlimited)")
	filesPullCmd.Flags().StringVar(&filesInclude, "include", "", "comma-separated patterns to include")
	filesPullCmd.Flags().StringVar(&filesExclude, "exclude", "", "comma-separated patterns to exclude")
	filesPullCmd.Flags().BoolVar(&filesVerify, "verify", false, "verify file checksums after sync (slower for large sites)")
	filesPullCmd.Flags().BoolVar(&filesPreservePermissions, "preserve-permissions", false, "preserve file permissions during sync")

	// Flags for push
	filesPushCmd.Flags().StringVar(&filesEnvironment, "environment", "", "WPEngine environment (default: from config)")
	filesPushCmd.Flags().BoolVar(&filesThemesOnly, "themes-only", false, "sync only themes directory")
	filesPushCmd.Flags().BoolVar(&filesPluginsOnly, "plugins-only", false, "sync only plugins directory")
	filesPushCmd.Flags().BoolVar(&filesMuPluginsOnly, "mu-plugins-only", false, "sync only mu-plugins directory")
	filesPushCmd.Flags().BoolVar(&filesUploadsOnly, "uploads-only", false, "sync only uploads directory")
	filesPushCmd.Flags().BoolVar(&filesDryRun, "dry-run", false, "show what would be transferred without syncing")
	filesPushCmd.Flags().BoolVar(&filesDelete, "delete", false, "delete remote files not present locally (DANGEROUS!)")
	filesPushCmd.Flags().IntVar(&filesBandwidthLimit, "bandwidth-limit", 0, "bandwidth limit in KB/s (0 = unlimited)")
	filesPushCmd.Flags().StringVar(&filesInclude, "include", "", "comma-separated patterns to include")
	filesPushCmd.Flags().StringVar(&filesExclude, "exclude", "", "comma-separated patterns to exclude")
	filesPushCmd.Flags().BoolVar(&filesVerify, "verify", false, "verify file checksums after sync (slower for large sites)")
	filesPushCmd.Flags().BoolVar(&filesPreservePermissions, "preserve-permissions", false, "preserve file permissions during sync")
}

func runFilesPull(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}
	p, err := provider.NewAuthenticatedProvider(cfg)
	if err != nil {
		return err
	}
	return files.Pull(p, cfg, files.SyncFlags{
		Environment:         filesEnvironment,
		ThemesOnly:          filesThemesOnly,
		PluginsOnly:         filesPluginsOnly,
		MuPluginsOnly:       filesMuPluginsOnly,
		ExcludeUploads:      filesExcludeUploads,
		DryRun:              filesDryRun,
		Delete:              filesDelete,
		BandwidthLimit:      filesBandwidthLimit,
		Include:             filesInclude,
		Exclude:             filesExclude,
		Verify:              filesVerify,
		PreservePermissions: filesPreservePermissions,
		ProjectDir:          getProjectDir(),
	})
}

func runFilesPush(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}
	p, err := provider.NewAuthenticatedProvider(cfg)
	if err != nil {
		return err
	}
	return files.Push(p, cfg, files.SyncFlags{
		Environment:         filesEnvironment,
		ThemesOnly:          filesThemesOnly,
		PluginsOnly:         filesPluginsOnly,
		MuPluginsOnly:       filesMuPluginsOnly,
		UploadsOnly:         filesUploadsOnly,
		DryRun:              filesDryRun,
		Delete:              filesDelete,
		BandwidthLimit:      filesBandwidthLimit,
		Include:             filesInclude,
		Exclude:             filesExclude,
		Verify:              filesVerify,
		PreservePermissions: filesPreservePermissions,
		ProjectDir:          getProjectDir(),
	})
}

// loadConfigForCommand loads configuration for a command
func loadConfigForCommand() (*config.Config, error) {
	cfg, err := config.Load(cfgFile, projectDir)
	if err != nil {
		return nil, errors.NewConfigNotFoundError(cfgFile, err)
	}
	return cfg, nil
}
