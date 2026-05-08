package wpengine

import (
	"fmt"
	"io"
	"strings"

	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/wpengine"
	"gopkg.in/yaml.v3"
)

// WPEngineProviderConfig holds typed fields from .stax.yml provider_config
type WPEngineProviderConfig struct {
	Install     string `yaml:"install"`
	Environment string `yaml:"environment"`
	AccountName string `yaml:"account_name"`
	SSHGateway  string `yaml:"ssh_gateway"`
}

// decodeProviderConfig converts the raw map[string]any from .stax.yml into typed config.
// Uses yaml round-trip for reliable type conversion.
func decodeProviderConfig(raw map[string]any) (*WPEngineProviderConfig, error) {
	data, err := yaml.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to encode provider_config: %w", err)
	}
	cfg := &WPEngineProviderConfig{
		Environment: "production",
		SSHGateway:  "ssh.wpengine.net",
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to decode provider_config: %w", err)
	}
	if cfg.Install == "" {
		return nil, fmt.Errorf("provider_config.install is required")
	}
	return cfg, nil
}

// WPEngineProvider implements the Provider interface for WPEngine
type WPEngineProvider struct {
	apiClient   *wpengine.Client
	sshClient   *wpengine.SSHClient
	install     string
	apiUser     string
	apiPassword string
	sshKey      string
	sshGateway  string
}

// Ensure WPEngineProvider implements all required interfaces
var (
	_ provider.Provider       = (*WPEngineProvider)(nil)
	_ provider.BackupManager  = (*WPEngineProvider)(nil)
	_ provider.RemoteExecutor = (*WPEngineProvider)(nil)
)

func init() {
	// Register WPEngine provider
	if err := provider.RegisterProvider("wpengine", &WPEngineProvider{}); err != nil {
		panic(fmt.Sprintf("failed to register WPEngine provider: %v", err))
	}
}

// Name returns the provider's unique identifier
func (p *WPEngineProvider) Name() string {
	return "wpengine"
}

// Description returns a human-readable description
func (p *WPEngineProvider) Description() string {
	return "WPEngine WordPress Hosting Platform"
}

// Capabilities returns the provider's capabilities
func (p *WPEngineProvider) Capabilities() provider.ProviderCapabilities {
	return GetWPEngineCapabilities()
}

// ===== Authentication & Setup =====

// ValidateCredentials validates credentials without establishing a connection
func (p *WPEngineProvider) ValidateCredentials(credentials map[string]string) error {
	required := []string{"api_user", "api_password", "install"}
	for _, key := range required {
		if credentials[key] == "" {
			return fmt.Errorf("missing required credential: %s", key)
		}
	}

	// SSH key is optional for API-only operations
	if credentials["ssh_key"] != "" {
		// Basic validation - check if it looks like a private key
		if !strings.Contains(credentials["ssh_key"], "PRIVATE KEY") {
			return fmt.Errorf("ssh_key does not appear to be a valid private key")
		}
	}

	return nil
}

// Authenticate authenticates with WPEngine
func (p *WPEngineProvider) Authenticate(credentials map[string]string) error {
	if err := p.ValidateCredentials(credentials); err != nil {
		return err
	}

	p.apiUser = credentials["api_user"]
	p.apiPassword = credentials["api_password"]
	p.install = credentials["install"]
	p.sshKey = credentials["ssh_key"]
	p.sshGateway = credentials["ssh_gateway"]

	// Create API client
	p.apiClient = wpengine.NewClient(p.apiUser, p.apiPassword, p.install)

	// Create SSH client if key provided
	if p.sshKey != "" {
		sshConfig := wpengine.SSHConfig{
			Host:       p.sshGateway,
			Install:    p.install,
			PrivateKey: p.sshKey,
		}

		sshClient, err := wpengine.NewSSHClient(sshConfig)
		if err != nil {
			return fmt.Errorf("failed to create SSH client: %w", err)
		}

		p.sshClient = sshClient
	}

	return nil
}

// AuthenticateFromConfig authenticates using .stax.yml provider_config plus keychain secrets.
// credentials map must contain: "api_user", "api_password", "ssh_key" (optional).
func (p *WPEngineProvider) AuthenticateFromConfig(providerConfig map[string]any, credentials map[string]string) error {
	cfg, err := decodeProviderConfig(providerConfig)
	if err != nil {
		return err
	}
	p.install = cfg.Install
	p.sshGateway = cfg.SSHGateway
	if p.sshGateway == "" {
		p.sshGateway = "ssh.wpengine.net"
	}
	p.apiUser = credentials["api_user"]
	p.apiPassword = credentials["api_password"]
	p.sshKey = credentials["ssh_key"]

	if p.apiUser == "" || p.apiPassword == "" {
		return fmt.Errorf("api_user and api_password are required")
	}

	p.apiClient = wpengine.NewClient(p.apiUser, p.apiPassword, p.install)

	if p.sshKey != "" {
		sshCfg := wpengine.SSHConfig{
			Host:       p.sshGateway,
			Install:    p.install,
			PrivateKey: p.sshKey,
		}
		sshClient, err := wpengine.NewSSHClient(sshCfg)
		if err != nil {
			return fmt.Errorf("failed to create SSH client: %w", err)
		}
		p.sshClient = sshClient
	}

	return nil
}

// TestConnection tests the connection to WPEngine
func (p *WPEngineProvider) TestConnection() error {
	if p.apiClient == nil {
		return fmt.Errorf("not authenticated")
	}

	if err := p.apiClient.TestConnection(); err != nil {
		return fmt.Errorf("API connection test failed: %w", err)
	}

	// Test SSH if available
	if p.sshClient != nil {
		if err := p.sshClient.TestConnection(); err != nil {
			return fmt.Errorf("SSH connection test failed: %w", err)
		}
	}

	return nil
}

// ===== Site Management =====

// ListSites lists all WPEngine installations
func (p *WPEngineProvider) ListSites() ([]provider.Site, error) {
	if p.apiClient == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	installs, err := p.apiClient.ListInstalls()
	if err != nil {
		return nil, err
	}

	sites := make([]provider.Site, len(installs))
	for i, install := range installs {
		sites[i] = provider.Site{
			ID:            install.ID,
			Name:          install.Name,
			PrimaryDomain: install.PrimaryDomain,
			Environment:   install.Environment,
			Status:        "active", // WPEngine doesn't expose status in list
			Provider:      "wpengine",
			Metadata: map[string]string{
				"php_version": install.PHPVersion,
			},
		}
	}

	return sites, nil
}

// GetSite retrieves information about a specific site
func (p *WPEngineProvider) GetSite(identifier string) (*provider.Site, error) {
	if p.apiClient == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	// Try to find by name first
	details, err := p.apiClient.GetInstallByName(identifier)
	if err != nil {
		// Try by ID
		details, err = p.apiClient.GetInstall(identifier)
		if err != nil {
			return nil, fmt.Errorf("site not found: %s", identifier)
		}
	}

	site := &provider.Site{
		ID:            details.ID,
		Name:          details.Name,
		PrimaryDomain: details.PrimaryDomain,
		Environment:   details.Environment,
		Status:        "active",
		Provider:      "wpengine",
		Metadata: map[string]string{
			"php_version":       details.PHPVersion,
			"mysql_version":     details.MySQLVersion,
			"wordpress_version": details.WordPressVersion,
		},
	}

	return site, nil
}

// GetSiteMetadata retrieves detailed metadata about a site
func (p *WPEngineProvider) GetSiteMetadata(site *provider.Site) (*provider.SiteMetadata, error) {
	if p.apiClient == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	details, err := p.apiClient.GetInstall(site.ID)
	if err != nil {
		return nil, err
	}

	metadata := &provider.SiteMetadata{
		Site:             site,
		PHPVersion:       details.PHPVersion,
		MySQLVersion:     details.MySQLVersion,
		WordPressVersion: details.WordPressVersion,
		DiskUsage: provider.DiskUsage{
			Used:  details.DiskUsage.Used,
			Total: details.DiskUsage.Total,
		},
		Domains:  details.Domains,
		Features: []string{"ssl", "cdn", "backups", "staging"},
	}

	return metadata, nil
}

// ===== Database Operations =====

// ExportDatabase exports the database from WPEngine
func (p *WPEngineProvider) ExportDatabase(site *provider.Site, options provider.DatabaseExportOptions) (io.ReadCloser, error) {
	if p.sshClient == nil {
		return nil, fmt.Errorf("SSH client not configured (SSH key required)")
	}

	// Convert provider options to WPEngine options
	wpOptions := wpengine.DatabaseOptions{
		ExcludeTables:  options.ExcludeTables,
		SkipLogs:       options.SkipLogs,
		SkipTransients: options.SkipTransients,
		SkipSpam:       options.SkipSpam,
		Compress:       options.Compress,
		ExtraFlags:     options.ExtraFlags,
	}

	return p.sshClient.ExportDatabase(wpOptions)
}

// ImportDatabase imports a database to WPEngine
func (p *WPEngineProvider) ImportDatabase(site *provider.Site, data io.Reader, options provider.DatabaseImportOptions) error {
	// WPEngine doesn't support direct database import via SSH for security
	// This would need to be done through WPEngine's portal or support
	return fmt.Errorf("database import not supported by WPEngine provider (use WPEngine portal)")
}

// GetDatabaseCredentials retrieves database credentials
func (p *WPEngineProvider) GetDatabaseCredentials(site *provider.Site) (*provider.DatabaseCredentials, error) {
	if p.sshClient == nil {
		return nil, fmt.Errorf("SSH client not configured")
	}

	// WPEngine doesn't expose database credentials directly
	// Users should use WP-CLI or wp-config.php
	return nil, fmt.Errorf("database credentials not exposed by WPEngine (use wp-config.php)")
}

// ===== File Operations =====

// SyncFiles synchronizes files from WPEngine
func (p *WPEngineProvider) SyncFiles(site *provider.Site, destination string, options provider.SyncOptions) error {
	if p.sshClient == nil {
		return fmt.Errorf("SSH client not configured")
	}

	// Convert provider options to WPEngine options.
	// If options.Source is a relative path (no SSH host prefix), build the full
	// SSH source URL so rsync knows to connect via SSH rather than treating it
	// as a local path.
	source := options.Source
	if source != "" {
		source = p.sshClient.SSHSource(source)
	}

	wpOptions := wpengine.SyncOptions{
		Source:         source,
		Destination:    destination,
		Include:        options.Include,
		Exclude:        options.Exclude,
		Delete:         options.Delete,
		DryRun:         options.DryRun,
		BandwidthLimit: options.BandwidthLimit,
		Progress:       options.Progress,
	}

	// Use WPEngine-specific sync (wp-content by default)
	return p.sshClient.SyncWPContent(destination, wpOptions)
}

// DownloadFile downloads a single file from WPEngine
func (p *WPEngineProvider) DownloadFile(site *provider.Site, remotePath string) (io.ReadCloser, error) {
	if p.sshClient == nil {
		return nil, fmt.Errorf("SSH client not configured")
	}

	// Create a pipe for streaming
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		// Execute cat command via SSH
		output, err := p.sshClient.ExecuteCommand(fmt.Sprintf("cat %s", remotePath))
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		_, err = pw.Write([]byte(output))
		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr, nil
}

// UploadFile uploads a single file to WPEngine
func (p *WPEngineProvider) UploadFile(site *provider.Site, localPath, remotePath string) error {
	// WPEngine doesn't support file uploads via SSH (read-only filesystem)
	return fmt.Errorf("file upload not supported by WPEngine (use Git deployments)")
}

// ===== Environment Information =====

// GetPHPVersion returns the PHP version
func (p *WPEngineProvider) GetPHPVersion(site *provider.Site) (string, error) {
	if site.Metadata["php_version"] != "" {
		return site.Metadata["php_version"], nil
	}

	// Get from API
	metadata, err := p.GetSiteMetadata(site)
	if err != nil {
		return "", err
	}

	return metadata.PHPVersion, nil
}

// GetMySQLVersion returns the MySQL version
func (p *WPEngineProvider) GetMySQLVersion(site *provider.Site) (string, error) {
	if site.Metadata["mysql_version"] != "" {
		return site.Metadata["mysql_version"], nil
	}

	metadata, err := p.GetSiteMetadata(site)
	if err != nil {
		return "", err
	}

	return metadata.MySQLVersion, nil
}

// GetWordPressVersion returns the WordPress version
func (p *WPEngineProvider) GetWordPressVersion(site *provider.Site) (string, error) {
	if site.Metadata["wordpress_version"] != "" {
		return site.Metadata["wordpress_version"], nil
	}

	metadata, err := p.GetSiteMetadata(site)
	if err != nil {
		return "", err
	}

	return metadata.WordPressVersion, nil
}

// ===== BackupManager Interface =====

// ListBackups lists available backups
func (p *WPEngineProvider) ListBackups(site *provider.Site) ([]provider.Backup, error) {
	if p.apiClient == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	wpBackups, err := p.apiClient.ListBackups(site.ID)
	if err != nil {
		return nil, err
	}

	backups := make([]provider.Backup, len(wpBackups))
	for i, backup := range wpBackups {
		backups[i] = provider.Backup{
			ID:          backup.ID,
			Type:        backup.Type,
			Description: "",
			Size:        backup.Size,
			CreatedAt:   backup.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Status:      backup.Status,
		}
	}

	return backups, nil
}

// CreateBackup creates a manual backup
func (p *WPEngineProvider) CreateBackup(site *provider.Site, description string) (*provider.Backup, error) {
	if p.apiClient == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	backupID, err := p.apiClient.CreateBackup(site.ID, description)
	if err != nil {
		return nil, err
	}

	return &provider.Backup{
		ID:          backupID,
		Type:        "manual",
		Description: description,
		Status:      "pending",
	}, nil
}

// RestoreBackup restores from a backup
func (p *WPEngineProvider) RestoreBackup(site *provider.Site, backupID string, options provider.RestoreOptions) error {
	// WPEngine requires backup restoration through the portal
	return fmt.Errorf("backup restoration must be done through WPEngine portal")
}

// DeleteBackup deletes a backup
func (p *WPEngineProvider) DeleteBackup(site *provider.Site, backupID string) error {
	// WPEngine doesn't allow manual backup deletion
	return fmt.Errorf("backup deletion not supported by WPEngine")
}

// DownloadBackup downloads a backup archive
func (p *WPEngineProvider) DownloadBackup(site *provider.Site, backupID string) (io.ReadCloser, error) {
	// WPEngine doesn't provide backup download API
	return nil, fmt.Errorf("backup download not supported by WPEngine API")
}

// ===== RemoteExecutor Interface =====

// ExecuteCommand executes a shell command
func (p *WPEngineProvider) ExecuteCommand(site *provider.Site, command string) (string, error) {
	if p.sshClient == nil {
		return "", fmt.Errorf("SSH client not configured")
	}

	return p.sshClient.ExecuteCommand(command)
}

// ExecuteWPCLI executes a WP-CLI command
func (p *WPEngineProvider) ExecuteWPCLI(site *provider.Site, args []string) (string, error) {
	if p.sshClient == nil {
		return "", fmt.Errorf("SSH client not configured")
	}

	return p.sshClient.GetWPCLI(args)
}

// StreamCommand executes a command and streams output
func (p *WPEngineProvider) StreamCommand(site *provider.Site, command string, stdout, stderr io.Writer) error {
	if p.sshClient == nil {
		return fmt.Errorf("SSH client not configured")
	}

	return p.sshClient.ExecuteCommandWithOutput(command, stdout, stderr)
}

// Close closes any open connections
func (p *WPEngineProvider) Close() error {
	if p.sshClient != nil {
		return p.sshClient.Close()
	}
	return nil
}
