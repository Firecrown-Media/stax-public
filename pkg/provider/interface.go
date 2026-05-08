package provider

import (
	"io"
)

// Provider is the core interface that all hosting providers must implement
type Provider interface {
	// ===== Metadata =====

	// Name returns the provider's unique identifier (e.g., "wpengine", "aws", "wordpress-vip")
	Name() string

	// Description returns a human-readable description of the provider
	Description() string

	// Capabilities returns the provider's capability set
	Capabilities() ProviderCapabilities

	// ===== Authentication & Setup =====

	// Authenticate authenticates with the provider using provided credentials
	// credentials map contains provider-specific authentication parameters
	// Returns error if authentication fails
	Authenticate(credentials map[string]string) error

	// TestConnection tests the connection to the provider
	// Returns error if connection test fails
	TestConnection() error

	// ValidateCredentials validates credentials without establishing a connection
	// Returns error if credentials are invalid or incomplete
	ValidateCredentials(credentials map[string]string) error

	// ===== Site Management =====

	// ListSites lists all sites/installations available on this provider
	ListSites() ([]Site, error)

	// GetSite retrieves information about a specific site
	// identifier can be site ID, name, or domain (provider-dependent)
	GetSite(identifier string) (*Site, error)

	// GetSiteMetadata retrieves detailed metadata about a site
	GetSiteMetadata(site *Site) (*SiteMetadata, error)

	// ===== Database Operations =====

	// ExportDatabase exports the database from the remote site
	// Returns an io.ReadCloser that streams the database dump
	// Caller is responsible for closing the ReadCloser
	ExportDatabase(site *Site, options DatabaseExportOptions) (io.ReadCloser, error)

	// ImportDatabase imports a database to the remote site
	// data is an io.Reader containing the SQL dump
	ImportDatabase(site *Site, data io.Reader, options DatabaseImportOptions) error

	// GetDatabaseCredentials retrieves database connection credentials
	// Useful for direct database access or debugging
	GetDatabaseCredentials(site *Site) (*DatabaseCredentials, error)

	// ===== File Operations =====

	// SyncFiles synchronizes files from remote site to local destination
	// Typically syncs wp-content or specific directories
	SyncFiles(site *Site, destination string, options SyncOptions) error

	// DownloadFile downloads a single file from the remote site
	DownloadFile(site *Site, remotePath string) (io.ReadCloser, error)

	// UploadFile uploads a single file to the remote site
	UploadFile(site *Site, localPath, remotePath string) error

	// ===== Environment Information =====

	// GetPHPVersion returns the PHP version for the site
	GetPHPVersion(site *Site) (string, error)

	// GetMySQLVersion returns the MySQL/MariaDB version for the site
	GetMySQLVersion(site *Site) (string, error)

	// GetWordPressVersion returns the WordPress version for the site
	GetWordPressVersion(site *Site) (string, error)
}

// ===== Common Types =====

// Site represents a WordPress site on a hosting provider
type Site struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	PrimaryDomain string            `json:"primary_domain"`
	Environment   string            `json:"environment"` // e.g., "production", "staging"
	Status        string            `json:"status"`      // e.g., "active", "suspended"
	Provider      string            `json:"provider"`    // Provider name
	Metadata      map[string]string `json:"metadata"`    // Provider-specific metadata
}

// SiteMetadata contains detailed information about a site
type SiteMetadata struct {
	Site             *Site     `json:"site"`
	PHPVersion       string    `json:"php_version"`
	MySQLVersion     string    `json:"mysql_version"`
	WordPressVersion string    `json:"wordpress_version"`
	DiskUsage        DiskUsage `json:"disk_usage"`
	Domains          []string  `json:"domains"`
	Features         []string  `json:"features"` // e.g., "ssl", "cdn", "backups"
	CreatedAt        string    `json:"created_at"`
	UpdatedAt        string    `json:"updated_at"`
}

// DiskUsage represents disk space usage
type DiskUsage struct {
	Used  int64 `json:"used"`  // Bytes used
	Total int64 `json:"total"` // Total bytes available
}

// Environment represents a site environment
type Environment struct {
	Name         string `json:"name"` // e.g., "production", "staging"
	URL          string `json:"url"`
	Status       string `json:"status"` // e.g., "active", "inactive"
	IsDefault    bool   `json:"is_default"`
	LastDeployAt string `json:"last_deploy_at"`
}

// ===== Database Types =====

// DatabaseExportOptions configures database export behavior
type DatabaseExportOptions struct {
	ExcludeTables  []string `json:"exclude_tables"`  // Tables to exclude
	SkipLogs       bool     `json:"skip_logs"`       // Skip log tables
	SkipTransients bool     `json:"skip_transients"` // Skip transient data
	SkipSpam       bool     `json:"skip_spam"`       // Skip spam comments
	Compress       bool     `json:"compress"`        // Compress output (gzip)
	IncludePrefix  bool     `json:"include_prefix"`  // Include table prefix detection
	ExtraFlags     []string `json:"extra_flags"`     // passed to wp db export
}

// DatabaseImportOptions configures database import behavior
type DatabaseImportOptions struct {
	DropExisting  bool     `json:"drop_existing"`  // Drop existing tables
	SearchReplace []string `json:"search_replace"` // Search/replace patterns
	SkipErrors    bool     `json:"skip_errors"`    // Continue on SQL errors
}

// DatabaseCredentials contains database connection information
type DatabaseCredentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSL      bool   `json:"ssl"`
}

// ===== File Sync Types =====

// SyncOptions configures file synchronization
type SyncOptions struct {
	Source         string   `json:"source"`          // Source path (optional, provider determines default)
	Destination    string   `json:"destination"`     // Local destination path
	Include        []string `json:"include"`         // Patterns to include
	Exclude        []string `json:"exclude"`         // Patterns to exclude
	Delete         bool     `json:"delete"`          // Delete files not on remote
	DryRun         bool     `json:"dry_run"`         // Perform dry run
	BandwidthLimit int      `json:"bandwidth_limit"` // KB/s limit
	Progress       bool     `json:"progress"`        // Show progress
}

// ===== Provider Capabilities =====

// ProviderCapabilities describes what features a provider supports
type ProviderCapabilities struct {
	// Core capabilities (should always be true if implemented correctly)
	Authentication bool `json:"authentication"`
	SiteManagement bool `json:"site_management"`
	DatabaseExport bool `json:"database_export"`
	DatabaseImport bool `json:"database_import"`
	FileSync       bool `json:"file_sync"`

	// Optional capabilities
	Deployment      bool `json:"deployment"`       // Git-based deployments
	Environments    bool `json:"environments"`     // Multi-environment support
	Backups         bool `json:"backups"`          // Automated backups
	RemoteExecution bool `json:"remote_execution"` // SSH/WP-CLI
	MediaManagement bool `json:"media_management"` // CDN/media proxy
	SSHAccess       bool `json:"ssh_access"`       // Direct SSH access
	APIAccess       bool `json:"api_access"`       // REST API access

	// Advanced capabilities
	Scaling    bool `json:"scaling"`    // Auto-scaling support
	Monitoring bool `json:"monitoring"` // Performance monitoring
	Logging    bool `json:"logging"`    // Centralized logging
}

// ===== Optional Capability Interfaces =====

// Deployer interface for providers that support deployments
type Deployer interface {
	Provider

	// Deploy deploys code to the site
	Deploy(site *Site, options DeployOptions) (*Deployment, error)

	// GetDeploymentStatus checks the status of a deployment
	GetDeploymentStatus(site *Site, deploymentID string) (*DeploymentStatus, error)

	// ListDeployments lists recent deployments
	ListDeployments(site *Site) ([]Deployment, error)
}

// DeployOptions configures deployment behavior
type DeployOptions struct {
	Branch      string            `json:"branch"`      // Git branch to deploy
	Commit      string            `json:"commit"`      // Specific commit (optional)
	Message     string            `json:"message"`     // Deployment message
	Environment string            `json:"environment"` // Target environment
	Metadata    map[string]string `json:"metadata"`    // Provider-specific options
}

// Deployment represents a deployment
type Deployment struct {
	ID         string `json:"id"`
	Status     string `json:"status"` // "pending", "in_progress", "completed", "failed"
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	Message    string `json:"message"`
	DeployedAt string `json:"deployed_at"`
	DeployedBy string `json:"deployed_by"`
}

// DeploymentStatus represents deployment status details
type DeploymentStatus struct {
	Deployment *Deployment `json:"deployment"`
	Progress   int         `json:"progress"` // 0-100
	Phase      string      `json:"phase"`    // Current phase
	Logs       []string    `json:"logs"`     // Recent log lines
	Error      string      `json:"error"`    // Error message if failed
}

// EnvironmentManager interface for providers with multi-environment support
type EnvironmentManager interface {
	Provider

	// ListEnvironments lists available environments for a site
	ListEnvironments(site *Site) ([]Environment, error)

	// GetEnvironment retrieves information about a specific environment
	GetEnvironment(site *Site, environmentName string) (*Environment, error)

	// SwitchEnvironment switches to a different environment
	SwitchEnvironment(site *Site, environmentName string) error

	// CreateEnvironment creates a new environment (if supported)
	CreateEnvironment(site *Site, environmentName string, options EnvironmentOptions) error

	// DeleteEnvironment deletes an environment (if supported)
	DeleteEnvironment(site *Site, environmentName string) error
}

// EnvironmentOptions configures environment creation
type EnvironmentOptions struct {
	CloneFrom  string            `json:"clone_from"` // Clone from existing environment
	PHPVersion string            `json:"php_version"`
	Domain     string            `json:"domain"`
	Metadata   map[string]string `json:"metadata"`
}

// BackupManager interface for providers that support backups
type BackupManager interface {
	Provider

	// ListBackups lists available backups for a site
	ListBackups(site *Site) ([]Backup, error)

	// CreateBackup creates a manual backup
	CreateBackup(site *Site, description string) (*Backup, error)

	// RestoreBackup restores a site from a backup
	RestoreBackup(site *Site, backupID string, options RestoreOptions) error

	// DeleteBackup deletes a backup
	DeleteBackup(site *Site, backupID string) error

	// DownloadBackup downloads a backup archive
	DownloadBackup(site *Site, backupID string) (io.ReadCloser, error)
}

// Backup represents a backup
type Backup struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // "manual", "automatic", "scheduled"
	Description string `json:"description"`
	Size        int64  `json:"size"`
	CreatedAt   string `json:"created_at"`
	Status      string `json:"status"` // "pending", "completed", "failed"
	ExpiresAt   string `json:"expires_at"`
}

// RestoreOptions configures backup restoration
type RestoreOptions struct {
	DatabaseOnly bool   `json:"database_only"` // Restore only database
	FilesOnly    bool   `json:"files_only"`    // Restore only files
	Environment  string `json:"environment"`   // Target environment
}

// RemoteExecutor interface for providers that support remote execution
type RemoteExecutor interface {
	Provider

	// ExecuteCommand executes a shell command on the remote server
	ExecuteCommand(site *Site, command string) (string, error)

	// ExecuteWPCLI executes a WP-CLI command
	ExecuteWPCLI(site *Site, args []string) (string, error)

	// StreamCommand executes a command and streams output
	StreamCommand(site *Site, command string, stdout, stderr io.Writer) error
}

// MediaManager interface for providers with media/CDN support
type MediaManager interface {
	Provider

	// GetMediaURL returns the CDN or media proxy URL for a site
	GetMediaURL(site *Site) (string, error)

	// SupportsRemoteMedia indicates if provider supports remote media serving
	SupportsRemoteMedia() bool

	// ConfigureMedia configures media settings
	ConfigureMedia(site *Site, options MediaOptions) error

	// PurgeMediaCache purges the media cache (if CDN)
	PurgeMediaCache(site *Site, paths []string) error
}

// MediaOptions configures media/CDN settings
type MediaOptions struct {
	CDNEnabled bool     `json:"cdn_enabled"`
	CDNDomain  string   `json:"cdn_domain"`
	CacheTTL   int      `json:"cache_ttl"` // Seconds
	Excludes   []string `json:"excludes"`  // Paths to exclude from CDN
}

// Migrator interface for providers that support importing from other providers
type Migrator interface {
	Provider

	// ImportFromProvider imports a site from another provider
	ImportFromProvider(sourceProvider Provider, sourceSite *Site, options MigrationOptions) error

	// ExportToProvider exports a site to another provider
	ExportToProvider(targetProvider Provider, site *Site, options MigrationOptions) error
}

// MigrationOptions configures cross-provider migration
type MigrationOptions struct {
	IncludeDatabase bool     `json:"include_database"`
	IncludeFiles    bool     `json:"include_files"`
	IncludeMedia    bool     `json:"include_media"`
	ExcludePlugins  []string `json:"exclude_plugins"`
	ExcludeThemes   []string `json:"exclude_themes"`
	DryRun          bool     `json:"dry_run"`
}
