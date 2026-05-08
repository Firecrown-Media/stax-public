package files

import (
	"fmt"
	"os"
	"strings"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/firecrown-media/stax/pkg/wpengine"
)

// SyncFlags holds the CLI flag values for a file sync operation.
type SyncFlags struct {
	// Environment overrides the environment from config when non-empty.
	Environment string

	// ThemesOnly restricts sync to wp-content/themes/.
	ThemesOnly bool

	// PluginsOnly restricts sync to wp-content/plugins/.
	PluginsOnly bool

	// MuPluginsOnly restricts sync to wp-content/mu-plugins/.
	MuPluginsOnly bool

	// UploadsOnly restricts sync to wp-content/uploads/ (push only).
	UploadsOnly bool

	// ExcludeUploads excludes the uploads directory from sync (pull only).
	ExcludeUploads bool

	// DryRun shows what would be transferred without making changes.
	DryRun bool

	// Delete removes destination files that are absent from the source.
	Delete bool

	// BandwidthLimit caps rsync bandwidth in KB/s (0 = unlimited).
	BandwidthLimit int

	// Include is a comma-separated list of patterns to include.
	Include string

	// Exclude is a comma-separated list of patterns to exclude.
	Exclude string

	// Verify enables checksum verification after sync.
	Verify bool

	// PreservePermissions preserves file permissions during sync.
	PreservePermissions bool

	// ProjectDir is the local project root; if empty the current working
	// directory is used.
	ProjectDir string
}

// Pull pulls files from the provider to the local environment.
func Pull(p provider.Provider, cfg *config.Config, flags SyncFlags) error {
	install := providerConfigString(cfg.ProviderConfig, "install")
	environment := flags.Environment
	if environment == "" {
		environment = providerConfigString(cfg.ProviderConfig, "environment")
	}

	ui.Info("Environment: %s", environment)
	ui.Info("Install: %s", install)

	remotePath, localPath := resolvePaths(install, flags, false)

	if flags.DryRun {
		ui.Info("DRY RUN - No files will be transferred")
	}

	site := &provider.Site{Name: install, Provider: cfg.Provider}
	syncOpts := buildProviderSyncOptions(cfg, flags, remotePath)

	ui.Info("Starting file synchronization...")
	if err := p.SyncFiles(site, localPath, syncOpts); err != nil {
		return fmt.Errorf("file sync failed: %w", err)
	}

	if !flags.DryRun {
		ui.Success("Files synchronized successfully")
	} else {
		ui.Info("Dry run completed")
	}

	ui.Success("\nFile pull completed!")
	return nil
}

// Push pushes files from the local environment to WPEngine.
func Push(p provider.Provider, cfg *config.Config, flags SyncFlags) error {
	install := providerConfigString(cfg.ProviderConfig, "install")
	environment := flags.Environment
	if environment == "" {
		environment = providerConfigString(cfg.ProviderConfig, "environment")
	}

	// Safety check: confirm production pushes.
	if environment == "production" && !flags.DryRun {
		ui.Warning("You are about to push files to PRODUCTION!")
		ui.Warning("This will modify files on your live site.")
		if !ui.Confirm("Are you sure you want to continue?") {
			ui.Info("Operation cancelled")
			return nil
		}
	}

	// Safety check: confirm delete mode.
	if flags.Delete && !flags.DryRun {
		ui.Warning("Delete mode is ENABLED!")
		ui.Warning("This will remove remote files that don't exist locally.")
		if !ui.Confirm("Are you sure you want to delete remote files?") {
			ui.Info("Operation cancelled")
			return nil
		}
	}

	// Warn about uploads.
	if flags.UploadsOnly || (!flags.ThemesOnly && !flags.PluginsOnly && !flags.MuPluginsOnly) {
		ui.Warning("Pushing uploads directory may transfer large files and take significant time")
		ui.Info("Consider using --themes-only, --plugins-only, or --mu-plugins-only for faster syncs")
	}

	ui.Info("Environment: %s", environment)
	ui.Info("Install: %s", install)

	sshClient, err := newSSHClient(cfg, install)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	ui.Success("Connected to WPEngine")

	syncOptions := BuildSyncOptions(cfg, flags)

	remotePath, localPath := resolvePaths(install, flags, true)

	// Validate local path exists before pushing.
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("local path does not exist: %s", localPath)
	}

	if flags.DryRun {
		ui.Info("DRY RUN - No files will be transferred")
		ui.Info("Review the output below to see what would be pushed")
	}

	ui.Info("Starting file push...")
	if err := sshClient.PushDirectory(localPath, remotePath, syncOptions); err != nil {
		return fmt.Errorf("file push failed: %w", err)
	}

	if !flags.DryRun {
		ui.Success("Files pushed successfully")

		ui.Info("Verifying file integrity...")
		if err := sshClient.VerifyFileIntegrity(remotePath, localPath); err != nil {
			ui.Warning("File integrity check failed: %v", err)
			ui.Info("Files were transferred but counts may differ (this is usually OK)")
		} else {
			ui.Success("File integrity verified")
		}

		if flags.Verify {
			ui.Section("Checksum Verification")
			ui.Info("Generating checksums (this may take a while for large sites)...")

			spinner := ui.NewSpinner("Verifying checksums...")
			spinner.Start()

			result, err := sshClient.VerifyFileChecksums(remotePath, localPath)
			spinner.Stop()

			if err != nil {
				ui.Warning("Checksum verification failed: %v", err)
			} else {
				printChecksumResults(result)
			}
		}
	} else {
		ui.Info("Dry run completed")
	}

	ui.Success("\nFile push completed!")
	return nil
}

// buildProviderSyncOptions constructs a provider.SyncOptions from config and CLI flags.
func buildProviderSyncOptions(cfg *config.Config, flags SyncFlags, remotePath string) provider.SyncOptions {
	opts := provider.SyncOptions{
		Source:         remotePath,
		Delete:         flags.Delete,
		DryRun:         flags.DryRun,
		BandwidthLimit: flags.BandwidthLimit,
		Progress:       true,
		Include:        []string{},
		Exclude:        wpengine.GetExcludePatterns(),
	}

	if flags.Include != "" {
		opts.Include = strings.Split(flags.Include, ",")
	}
	if flags.Exclude != "" {
		opts.Exclude = append(opts.Exclude, strings.Split(flags.Exclude, ",")...)
	}
	if flags.ExcludeUploads {
		opts.Exclude = append(opts.Exclude, "uploads/")
	}
	if flags.BandwidthLimit == 0 && cfg.Performance.RsyncBandwidthLimit > 0 {
		opts.BandwidthLimit = cfg.Performance.RsyncBandwidthLimit
	}

	return opts
}

// BuildSyncOptions constructs a wpengine.SyncOptions from config and CLI flags.
// Flag values take precedence over config values.
func BuildSyncOptions(cfg *config.Config, flags SyncFlags) wpengine.SyncOptions {
	projectDir := flags.ProjectDir

	opts := wpengine.SyncOptions{
		DryRun:              flags.DryRun,
		Delete:              flags.Delete,
		BandwidthLimit:      flags.BandwidthLimit,
		Progress:            true,
		PreservePermissions: flags.PreservePermissions,
		Include:             []string{},
		Exclude:             wpengine.GetExcludePatterns(),
		ProjectDir:          projectDir,
	}

	// Add custom includes.
	if flags.Include != "" {
		opts.Include = strings.Split(flags.Include, ",")
	}

	// Add custom excludes.
	if flags.Exclude != "" {
		customExcludes := strings.Split(flags.Exclude, ",")
		opts.Exclude = append(opts.Exclude, customExcludes...)
	}

	// Exclude uploads if requested.
	if flags.ExcludeUploads {
		opts.Exclude = append(opts.Exclude, "uploads/")
	}

	// Apply bandwidth limit from config if not specified via flag.
	if flags.BandwidthLimit == 0 && cfg.Performance.RsyncBandwidthLimit > 0 {
		opts.BandwidthLimit = cfg.Performance.RsyncBandwidthLimit
	}

	return opts
}

// newSSHClient creates and returns a connected wpengine SSH client.
func newSSHClient(cfg *config.Config, install string) (*wpengine.SSHClient, error) {
	creds, err := credentials.GetWPEngineCredentialsWithFallback(install)
	if err != nil {
		if credErr, ok := err.(*credentials.CredentialsNotFoundError); ok {
			return nil, errors.NewCredentialsNotFoundError(credErr.Tried, credErr.LastErr)
		}
		return nil, fmt.Errorf("failed to get WPEngine credentials: %w", err)
	}

	sshKey, err := credentials.GetSSHPrivateKeyWithFallback("wpengine")
	if err != nil {
		if keyErr, ok := err.(*credentials.SSHKeyNotFoundError); ok {
			return nil, errors.NewSSHKeyNotFoundError("", keyErr.Tried, keyErr.LastErr)
		}
		return nil, fmt.Errorf("failed to get SSH key: %w", err)
	}

	ui.Info("Connecting to WPEngine SSH Gateway...")
	sshConfig := wpengine.SSHConfig{
		Host:       providerConfigString(cfg.ProviderConfig, "ssh_gateway"),
		Port:       22,
		User:       creds.SSHUser,
		PrivateKey: sshKey,
		Install:    install,
	}

	if sshConfig.Host == "" {
		sshConfig.Host = "ssh.wpengine.net"
	}

	client, err := wpengine.NewSSHClient(sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WPEngine: %w", err)
	}

	return client, nil
}

// resolvePaths returns the remote and local paths for the sync operation.
// isPush controls whether UploadsOnly is considered (push only).
func resolvePaths(install string, flags SyncFlags, isPush bool) (remotePath, localPath string) {
	projectDir := flags.ProjectDir

	switch {
	case flags.ThemesOnly:
		ui.Info("Syncing themes only...")
		remotePath = fmt.Sprintf("sites/%s/wp-content/themes/", install)
		localPath = projectDir + "/wp-content/themes/"
	case flags.PluginsOnly:
		ui.Info("Syncing plugins only...")
		remotePath = fmt.Sprintf("sites/%s/wp-content/plugins/", install)
		localPath = projectDir + "/wp-content/plugins/"
	case flags.MuPluginsOnly:
		ui.Info("Syncing mu-plugins only...")
		remotePath = fmt.Sprintf("sites/%s/wp-content/mu-plugins/", install)
		localPath = projectDir + "/wp-content/mu-plugins/"
	case isPush && flags.UploadsOnly:
		ui.Info("Syncing uploads only...")
		remotePath = fmt.Sprintf("sites/%s/wp-content/uploads/", install)
		localPath = projectDir + "/wp-content/uploads/"
	default:
		ui.Info("Syncing wp-content directory...")
		remotePath = fmt.Sprintf("sites/%s/wp-content/", install)
		localPath = projectDir + "/wp-content/"
	}

	return remotePath, localPath
}

// printChecksumResults prints the results of a checksum verification.
func printChecksumResults(result *wpengine.ChecksumResult) {
	ui.Info("Total files checked: %d", result.TotalFiles)
	ui.Success("Matched files: %d", result.MatchedFiles)

	if result.MismatchedFiles > 0 {
		ui.Warning("Mismatched checksums: %d", result.MismatchedFiles)
		ui.Info("Files with different checksums:")
		for i, mismatch := range result.Mismatches {
			if i >= 10 {
				ui.Info("  ... and %d more", len(result.Mismatches)-10)
				break
			}
			ui.Info("  - %s", mismatch.RelativePath)
			ui.Verbose("    Remote: %s", mismatch.RemoteChecksum)
			ui.Verbose("    Local:  %s", mismatch.LocalChecksum)
		}
	}

	if result.MissingLocal > 0 {
		ui.Warning("Missing locally: %d", result.MissingLocal)
		ui.Info("Files that exist remotely but not locally:")
		for i, file := range result.MissingLocally {
			if i >= 10 {
				ui.Info("  ... and %d more", len(result.MissingLocally)-10)
				break
			}
			ui.Info("  - %s", file)
		}
	}

	if result.MissingRemote > 0 {
		ui.Warning("Missing remotely: %d", result.MissingRemote)
		ui.Info("Files that exist locally but not remotely:")
		for i, file := range result.MissingRemotely {
			if i >= 10 {
				ui.Info("  ... and %d more", len(result.MissingRemotely)-10)
				break
			}
			ui.Info("  - %s", file)
		}
	}

	if result.MismatchedFiles == 0 && result.MissingLocal == 0 && result.MissingRemote == 0 {
		ui.Success("All files verified successfully - checksums match!")
	} else {
		ui.Warning("Some files have checksum differences - review the details above")
	}
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
