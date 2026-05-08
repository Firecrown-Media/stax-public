package config

import (
	"gopkg.in/yaml.v3"
)

// Config represents the complete Stax configuration
type Config struct {
	Version int `yaml:"version"`

	// Project metadata
	Project ProjectConfig `yaml:"project"`

	// Provider selection (e.g. "wpengine")
	Provider string `yaml:"provider"`

	// Provider-specific configuration — decoded by the provider's Authenticate()
	ProviderConfig map[string]any `yaml:"provider_config,omitempty"`

	// Network and sites configuration
	Network NetworkConfig `yaml:"network"`

	// DDEV configuration
	DDEV DDEVConfig `yaml:"ddev"`

	// GitHub repository configuration
	Repository RepositoryConfig `yaml:"repository,omitempty"`

	// Build process configuration
	Build BuildConfig `yaml:"build,omitempty"`

	// WordPress configuration
	WordPress WordPressConfig `yaml:"wordpress,omitempty"`

	// Remote media configuration
	Media MediaConfig `yaml:"media,omitempty"`

	// Credentials (reference only)
	Credentials CredentialsConfig `yaml:"credentials,omitempty"`

	// Logging and debugging
	Logging LoggingConfig `yaml:"logging,omitempty"`

	// Snapshots
	Snapshots SnapshotsConfig `yaml:"snapshots,omitempty"`

	// Performance tuning
	Performance PerformanceConfig `yaml:"performance,omitempty"`

	// Migration configuration for stax migrate commands
	Migration MigrationConfig `yaml:"migration,omitempty"`
}

// ProjectConfig represents project metadata
type ProjectConfig struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"` // wordpress, wordpress-multisite
	Mode        string `yaml:"mode"` // subdomain, subdirectory, single
	Description string `yaml:"description,omitempty"`
}

// NetworkConfig represents multisite network configuration
type NetworkConfig struct {
	Domain     string       `yaml:"domain"`
	Title      string       `yaml:"title,omitempty"`
	AdminEmail string       `yaml:"admin_email,omitempty"`
	Sites      []SiteConfig `yaml:"sites,omitempty"`
}

// SiteConfig represents an individual site in the network
type SiteConfig struct {
	Name           string `yaml:"name"`
	Slug           string `yaml:"slug"`
	Title          string `yaml:"title"`
	Domain         string `yaml:"domain"`
	ProviderDomain string `yaml:"provider_domain"`
	Active         bool   `yaml:"active"`
}

// DDEVConfig represents DDEV configuration
type DDEVConfig struct {
	PHPVersion          string              `yaml:"php_version"`
	MySQLVersion        string              `yaml:"mysql_version"`
	MySQLType           string              `yaml:"mysql_type,omitempty"` // mysql, mariadb
	WebserverType       string              `yaml:"webserver_type"`
	RouterHTTPPort      string              `yaml:"router_http_port,omitempty"`
	RouterHTTPSPort     string              `yaml:"router_https_port,omitempty"`
	MailHogPort         string              `yaml:"mailhog_port,omitempty"`
	NFSMountEnabled     bool                `yaml:"nfs_mount_enabled"`
	MutagenEnabled      bool                `yaml:"mutagen_enabled"`
	XdebugEnabled       bool                `yaml:"xdebug_enabled"`
	UseDNSWhenPossible  bool                `yaml:"use_dns_when_possible,omitempty"`
	NodeJSVersion       string              `yaml:"nodejs_version,omitempty"`
	ComposerVersion     string              `yaml:"composer_version,omitempty"`
	AdditionalHostnames []string            `yaml:"additional_hostnames,omitempty"`
	AdditionalFQDNs     []string            `yaml:"additional_fqdns,omitempty"`
	CustomCommands      []DDEVCustomCommand `yaml:"custom_commands,omitempty"`
	Hooks               DDEVHooks           `yaml:"hooks,omitempty"`
}

// DDEVCustomCommand represents a custom DDEV command
type DDEVCustomCommand struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Enabled     bool   `yaml:"enabled"`
}

// DDEVHooks represents DDEV lifecycle hooks
type DDEVHooks struct {
	PreStart  []DDEVHook `yaml:"pre_start,omitempty"`
	PostStart []DDEVHook `yaml:"post_start,omitempty"`
	PreStop   []DDEVHook `yaml:"pre_stop,omitempty"`
	PostStop  []DDEVHook `yaml:"post_stop,omitempty"`
}

// DDEVHook represents a single hook command
type DDEVHook struct {
	Exec string `yaml:"exec"`
}

// RepositoryConfig represents GitHub repository configuration
type RepositoryConfig struct {
	URL        string       `yaml:"url"`
	Branch     string       `yaml:"branch"`
	Private    bool         `yaml:"private"`
	Depth      int          `yaml:"depth,omitempty"`
	Submodules bool         `yaml:"submodules"`
	Deploy     DeployConfig `yaml:"deploy,omitempty"`
}

// DeployConfig represents deployment configuration
type DeployConfig struct {
	Workflow string       `yaml:"workflow"`
	OnPush   OnPushConfig `yaml:"on_push,omitempty"`
}

// OnPushConfig represents on-push deployment configuration
type OnPushConfig struct {
	Branches []string `yaml:"branches,omitempty"`
}

// BuildConfig represents build process configuration
type BuildConfig struct {
	// Scripts configuration
	Scripts BuildScriptsConfig `yaml:"scripts,omitempty"`

	// Composer configuration
	Composer BuildComposerConfig `yaml:"composer,omitempty"`

	// NPM configuration
	NPM BuildNPMConfig `yaml:"npm,omitempty"`

	// PHPCS configuration
	PHPCS BuildPHPCSConfig `yaml:"phpcs,omitempty"`

	// Git hooks configuration
	Hooks BuildHooksConfig `yaml:"hooks,omitempty"`

	// Watch mode configuration
	Watch WatchConfig `yaml:"watch,omitempty"`
}

// BuildScriptsConfig represents build scripts configuration
type BuildScriptsConfig struct {
	// Main build script path
	Main string `yaml:"main,omitempty"`

	// Additional build scripts
	Additional []string `yaml:"additional,omitempty"`

	// Pre-build hooks
	PreBuild []string `yaml:"pre_build,omitempty"`

	// Post-build hooks
	PostBuild []string `yaml:"post_build,omitempty"`
}

// BuildComposerConfig represents composer build configuration
type BuildComposerConfig struct {
	// Install arguments
	InstallArgs string `yaml:"install_args,omitempty"`

	// Timeout in seconds
	Timeout int `yaml:"timeout,omitempty"`

	// Skip platform requirements check
	IgnorePlatformReqs bool `yaml:"ignore_platform_reqs,omitempty"`

	// Optimize autoloader
	Optimize bool `yaml:"optimize"`

	// Skip dev dependencies
	NoDev bool `yaml:"no_dev"`
}

// BuildNPMConfig represents NPM build configuration
type BuildNPMConfig struct {
	// Install arguments
	InstallArgs string `yaml:"install_args,omitempty"`

	// Build command
	BuildCommand string `yaml:"build_command"`

	// Dev command (npm start)
	DevCommand string `yaml:"dev_command,omitempty"`

	// Timeout in seconds
	Timeout int `yaml:"timeout,omitempty"`

	// Use legacy peer deps
	LegacyPeerDeps bool `yaml:"legacy_peer_deps"`
}

// BuildPHPCSConfig represents PHPCS configuration
type BuildPHPCSConfig struct {
	// Config file path
	Config string `yaml:"config,omitempty"`

	// Coding standard
	Standard string `yaml:"standard,omitempty"`

	// File extensions to check
	Extensions string `yaml:"extensions,omitempty"`

	// Patterns to ignore
	Ignore string `yaml:"ignore,omitempty"`

	// Show sniff codes
	ShowSniffs bool `yaml:"show_sniffs,omitempty"`
}

// BuildHooksConfig represents git hooks configuration
type BuildHooksConfig struct {
	// Enable pre-commit hook
	PreCommit bool `yaml:"pre_commit,omitempty"`

	// Enable pre-push hook
	PrePush bool `yaml:"pre_push,omitempty"`

	// Enable commit-msg hook
	CommitMsg bool `yaml:"commit_msg,omitempty"`
}

// WatchConfig represents watch mode configuration
type WatchConfig struct {
	Enabled bool     `yaml:"enabled"`
	Paths   []string `yaml:"paths,omitempty"`
	Command string   `yaml:"command"`
}

// WordPressConfig represents WordPress configuration
type WordPressConfig struct {
	Version       string                 `yaml:"version,omitempty"`
	Locale        string                 `yaml:"locale,omitempty"`
	Constants     map[string]interface{} `yaml:"constants,omitempty"`
	TablePrefix   string                 `yaml:"table_prefix,omitempty"`
	SearchReplace SearchReplaceConfig    `yaml:"search_replace,omitempty"`
}

// SearchReplaceConfig represents search-replace configuration
type SearchReplaceConfig struct {
	Network     []SearchReplacePair `yaml:"network,omitempty"`
	Sites       []SiteSearchReplace `yaml:"sites,omitempty"`
	SkipColumns []string            `yaml:"skip_columns,omitempty"`
	SkipTables  []string            `yaml:"skip_tables,omitempty"`
}

// SearchReplacePair represents a search-replace pair
type SearchReplacePair struct {
	Old string `yaml:"old"`
	New string `yaml:"new"`
}

// SiteSearchReplace represents search-replace for a specific site
type SiteSearchReplace struct {
	Old string `yaml:"old"`
	New string `yaml:"new"`
	URL string `yaml:"url"`
}

// MediaConfig represents remote media configuration
type MediaConfig struct {
	ProxyEnabled     bool           `yaml:"proxy_enabled"`
	PrimarySource    string         `yaml:"primary_source,omitempty"`
	BunnyCDN         BunnyCDNConfig `yaml:"bunnycdn,omitempty"`
	WPEngineFallback bool           `yaml:"wpengine_fallback"`
	Cache            CacheConfig    `yaml:"cache,omitempty"`
}

// BunnyCDNConfig represents BunnyCDN configuration
type BunnyCDNConfig struct {
	Hostname    string `yaml:"hostname"`
	PullZone    string `yaml:"pull_zone"`
	StorageZone string `yaml:"storage_zone"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Directory string `yaml:"directory"`
	MaxSize   string `yaml:"max_size"`
	TTL       int    `yaml:"ttl"`
}

// CredentialsConfig represents credentials references
type CredentialsConfig struct {
	WPEngine CredentialRef `yaml:"wpengine,omitempty"`
	GitHub   CredentialRef `yaml:"github,omitempty"`
	SSH      CredentialRef `yaml:"ssh,omitempty"`
}

// CredentialRef represents a keychain reference
type CredentialRef struct {
	KeychainService string `yaml:"keychain_service"`
	KeychainAccount string `yaml:"keychain_account"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"` // debug, info, warn, error
	File       string `yaml:"file"`
	MaxSize    string `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Format     string `yaml:"format"` // json, text
	Timestamp  bool   `yaml:"timestamp"`
	Caller     bool   `yaml:"caller"`
}

// SnapshotsConfig represents snapshot configuration
type SnapshotsConfig struct {
	Directory                string          `yaml:"directory"`
	AutoSnapshotBeforePull   bool            `yaml:"auto_snapshot_before_pull"`
	AutoSnapshotBeforeImport bool            `yaml:"auto_snapshot_before_import"`
	Retention                RetentionConfig `yaml:"retention,omitempty"`
	Compression              string          `yaml:"compression,omitempty"`
}

// RetentionConfig represents snapshot retention configuration
type RetentionConfig struct {
	Auto   int `yaml:"auto"`   // days
	Manual int `yaml:"manual"` // days
}

// PerformanceConfig represents performance tuning configuration
type PerformanceConfig struct {
	ParallelDownloads       int `yaml:"parallel_downloads"`
	RsyncBandwidthLimit     int `yaml:"rsync_bandwidth_limit"` // KB/s
	DatabaseImportBatchSize int `yaml:"database_import_batch_size"`
}

// MigrationConfig represents migration destination configuration.
type MigrationConfig struct {
	Destination string `yaml:"destination"` // e.g. "vip"
}

// ToYAML converts the config to YAML
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// FromYAML populates the config from YAML
func (c *Config) FromYAML(data []byte) error {
	return yaml.Unmarshal(data, c)
}

// Defaults returns a config with default values
func Defaults() *Config {
	return &Config{
		Version:  2,
		Provider: "wpengine",
		ProviderConfig: map[string]any{
			"environment": "production",
			"ssh_gateway": "ssh.wpengine.net",
		},
		Project: ProjectConfig{
			Type: "wordpress-multisite",
			Mode: "subdomain",
		},
		DDEV: DDEVConfig{
			PHPVersion:         "8.1",
			MySQLVersion:       "8.0",
			MySQLType:          "mysql",
			WebserverType:      "nginx-fpm",
			RouterHTTPPort:     "80",
			RouterHTTPSPort:    "443",
			MailHogPort:        "8025",
			NFSMountEnabled:    false,
			MutagenEnabled:     false,
			XdebugEnabled:      false,
			UseDNSWhenPossible: true,
			NodeJSVersion:      "20",
			ComposerVersion:    "2",
		},
		Repository: RepositoryConfig{
			Branch:     "main",
			Private:    true,
			Depth:      1,
			Submodules: false,
		},
		WordPress: WordPressConfig{
			Version:     "latest",
			Locale:      "en_US",
			TablePrefix: "wp_",
		},
		Media: MediaConfig{
			ProxyEnabled:     true,
			WPEngineFallback: true,
			Cache: CacheConfig{
				Enabled:   true,
				Directory: ".stax/media-cache",
				MaxSize:   "1GB",
				TTL:       86400,
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       "~/.stax/logs/stax.log",
			MaxSize:    "10MB",
			MaxBackups: 5,
			MaxAge:     30,
			Format:     "json",
			Timestamp:  true,
			Caller:     false,
		},
		Snapshots: SnapshotsConfig{
			Directory:                "~/.stax/snapshots",
			AutoSnapshotBeforePull:   true,
			AutoSnapshotBeforeImport: true,
			Retention: RetentionConfig{
				Auto:   7,
				Manual: 30,
			},
			Compression: "gzip",
		},
		Performance: PerformanceConfig{
			ParallelDownloads:       4,
			RsyncBandwidthLimit:     0,
			DatabaseImportBatchSize: 1000,
		},
	}
}

// ProviderConfigString safely extracts a string value from a provider_config map.
// Returns "" if m is nil, key is absent, or the value is not a string.
func ProviderConfigString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
