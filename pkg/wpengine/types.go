package wpengine

import (
	"time"
)

// Install represents a WPEngine installation
type Install struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	PrimaryDomain string `json:"primary_domain"`
	PHPVersion    string `json:"php_version"`
	Environment   string `json:"environment"`
}

// InstallDetails represents detailed information about a WPEngine installation
type InstallDetails struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	PrimaryDomain    string `json:"primary_domain"`
	PHPVersion       string `json:"php_version"`
	MySQLVersion     string `json:"mysql_version"`
	WordPressVersion string `json:"wordpress_version"`
	Environment      string `json:"environment"`
	DiskUsage        struct {
		Used  int64 `json:"used"`
		Total int64 `json:"total"`
	} `json:"disk_usage"`
	Domains []string `json:"domains"`
}

// Backup represents a WPEngine backup
type Backup struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
	Status    string    `json:"status"`
}

// SSHConfig represents SSH connection configuration
type SSHConfig struct {
	Host       string
	Port       int
	User       string
	PrivateKey string
	Install    string
}

// DatabaseOptions represents database export options
type DatabaseOptions struct {
	ExcludeTables  []string
	SkipLogs       bool
	SkipTransients bool
	SkipSpam       bool
	Compress       bool
	ExtraFlags     []string // additional flags passed to wp db export
}

// SyncOptions represents file synchronization options
type SyncOptions struct {
	Source              string
	Destination         string
	Include             []string
	Exclude             []string
	Delete              bool
	DryRun              bool
	BandwidthLimit      int // KB/s
	Progress            bool
	PreservePermissions bool   // Preserve file permissions during sync
	ProjectDir          string // Project directory for loading .staxignore
}

// ExportOptions represents database export options
type ExportOptions struct {
	SkipLogs       bool
	SkipTransients bool
	SkipSpam       bool
	ExcludeTables  []string
	Compress       bool
}

// ListInstallsResponse represents the API response for listing installs
type ListInstallsResponse struct {
	Results []Install `json:"results"`
	Count   int       `json:"count"`
	Next    string    `json:"next,omitempty"`
	Prev    string    `json:"prev,omitempty"`
}

// ListBackupsResponse represents the API response for listing backups
type ListBackupsResponse struct {
	Results []Backup `json:"results"`
	Count   int      `json:"count"`
}

// CreateBackupRequest represents the request to create a backup
type CreateBackupRequest struct {
	Description string `json:"description"`
}

// CreateBackupResponse represents the response from creating a backup
type CreateBackupResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// InstallInfo is an alias for backward compatibility
type InstallInfo = InstallDetails
