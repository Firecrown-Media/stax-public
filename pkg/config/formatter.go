package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// FormatConfig formats the configuration for display
func FormatConfig(cfg *Config, format string) (string, error) {
	// Mask sensitive values
	masked := MaskSensitiveValues(cfg)

	switch format {
	case "json":
		data, err := json.MarshalIndent(masked, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return string(data), nil

	case "yaml", "yml":
		data, err := yaml.Marshal(masked)
		if err != nil {
			return "", fmt.Errorf("failed to marshal YAML: %w", err)
		}
		return string(data), nil

	case "pretty", "":
		return FormatPretty(masked), nil

	default:
		return "", fmt.Errorf("unsupported format: %s (use json, yaml, or pretty)", format)
	}
}

// MaskSensitiveValues creates a copy of the config with sensitive values masked
func MaskSensitiveValues(cfg *Config) *Config {
	// Create a deep copy first
	masked := *cfg

	// Mask credential references (these are just references, not actual secrets)
	// but we still mask them for clarity
	if masked.Credentials.WPEngine.KeychainAccount != "" {
		masked.Credentials.WPEngine.KeychainAccount = maskString(masked.Credentials.WPEngine.KeychainAccount)
	}
	if masked.Credentials.GitHub.KeychainAccount != "" {
		masked.Credentials.GitHub.KeychainAccount = maskString(masked.Credentials.GitHub.KeychainAccount)
	}
	if masked.Credentials.SSH.KeychainAccount != "" {
		masked.Credentials.SSH.KeychainAccount = maskString(masked.Credentials.SSH.KeychainAccount)
	}

	return &masked
}

// maskString masks a string, showing only first and last 4 characters
func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

// FormatPretty formats the config in a human-readable pretty format
func FormatPretty(cfg *Config) string {
	var sb strings.Builder

	// Project Configuration
	sb.WriteString("Project Configuration\n")
	if cfg.Project.Name != "" {
		sb.WriteString(fmt.Sprintf("  Name:        %s\n", cfg.Project.Name))
	}
	if cfg.Project.Type != "" {
		sb.WriteString(fmt.Sprintf("  Type:        %s\n", cfg.Project.Type))
	}
	if cfg.Project.Mode != "" {
		sb.WriteString(fmt.Sprintf("  Mode:        %s\n", cfg.Project.Mode))
	}
	if cfg.Project.Description != "" {
		sb.WriteString(fmt.Sprintf("  Description: %s\n", cfg.Project.Description))
	}

	// Provider Configuration
	sb.WriteString("\nProvider Configuration\n")
	if cfg.Provider != "" {
		sb.WriteString(fmt.Sprintf("  Provider:    %s\n", cfg.Provider))
	}
	for k, v := range cfg.ProviderConfig {
		sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
	}

	// Network Configuration
	if cfg.Network.Domain != "" || cfg.Network.Title != "" {
		sb.WriteString("\nNetwork Configuration\n")
		if cfg.Network.Domain != "" {
			sb.WriteString(fmt.Sprintf("  Domain:      %s\n", cfg.Network.Domain))
		}
		if cfg.Network.Title != "" {
			sb.WriteString(fmt.Sprintf("  Title:       %s\n", cfg.Network.Title))
		}
		if cfg.Network.AdminEmail != "" {
			sb.WriteString(fmt.Sprintf("  Admin Email: %s\n", cfg.Network.AdminEmail))
		}
		if len(cfg.Network.Sites) > 0 {
			sb.WriteString(fmt.Sprintf("  Sites:       %d configured\n", len(cfg.Network.Sites)))
		}
	}

	// DDEV Configuration
	sb.WriteString("\nDDEV Configuration\n")
	if cfg.DDEV.PHPVersion != "" {
		sb.WriteString(fmt.Sprintf("  PHP Version:    %s\n", cfg.DDEV.PHPVersion))
	}
	if cfg.DDEV.MySQLVersion != "" {
		sb.WriteString(fmt.Sprintf("  MySQL Version:  %s\n", cfg.DDEV.MySQLVersion))
	}
	if cfg.DDEV.MySQLType != "" {
		sb.WriteString(fmt.Sprintf("  MySQL Type:     %s\n", cfg.DDEV.MySQLType))
	}
	if cfg.DDEV.WebserverType != "" {
		sb.WriteString(fmt.Sprintf("  Webserver:      %s\n", cfg.DDEV.WebserverType))
	}
	if cfg.DDEV.NodeJSVersion != "" {
		sb.WriteString(fmt.Sprintf("  Node.js:        %s\n", cfg.DDEV.NodeJSVersion))
	}
	sb.WriteString(fmt.Sprintf("  Xdebug:         %t\n", cfg.DDEV.XdebugEnabled))

	// WordPress Configuration
	if cfg.WordPress.Version != "" || cfg.WordPress.Locale != "" {
		sb.WriteString("\nWordPress Configuration\n")
		if cfg.WordPress.Version != "" {
			sb.WriteString(fmt.Sprintf("  Version:     %s\n", cfg.WordPress.Version))
		}
		if cfg.WordPress.Locale != "" {
			sb.WriteString(fmt.Sprintf("  Locale:      %s\n", cfg.WordPress.Locale))
		}
		if cfg.WordPress.TablePrefix != "" {
			sb.WriteString(fmt.Sprintf("  Table Prefix: %s\n", cfg.WordPress.TablePrefix))
		}
	}

	// Repository Configuration
	if cfg.Repository.URL != "" {
		sb.WriteString("\nRepository Configuration\n")
		sb.WriteString(fmt.Sprintf("  URL:         %s\n", cfg.Repository.URL))
		if cfg.Repository.Branch != "" {
			sb.WriteString(fmt.Sprintf("  Branch:      %s\n", cfg.Repository.Branch))
		}
		sb.WriteString(fmt.Sprintf("  Private:     %t\n", cfg.Repository.Private))
	}

	// Media Configuration
	if cfg.Media.ProxyEnabled {
		sb.WriteString("\nMedia Configuration\n")
		sb.WriteString(fmt.Sprintf("  Proxy:       %t\n", cfg.Media.ProxyEnabled))
		if cfg.Media.PrimarySource != "" {
			sb.WriteString(fmt.Sprintf("  Source:      %s\n", cfg.Media.PrimarySource))
		}
		if cfg.Media.BunnyCDN.Hostname != "" {
			sb.WriteString(fmt.Sprintf("  BunnyCDN:    %s\n", cfg.Media.BunnyCDN.Hostname))
		}
	}

	// Build Configuration
	if cfg.Build.Composer.Optimize || cfg.Build.NPM.BuildCommand != "" {
		sb.WriteString("\nBuild Configuration\n")
		if cfg.Build.Composer.Optimize {
			sb.WriteString("  Composer:    enabled (optimized)\n")
		}
		if cfg.Build.NPM.BuildCommand != "" {
			sb.WriteString(fmt.Sprintf("  NPM Build:   %s\n", cfg.Build.NPM.BuildCommand))
		}
	}

	// Logging Configuration
	if cfg.Logging.Level != "" {
		sb.WriteString("\nLogging Configuration\n")
		sb.WriteString(fmt.Sprintf("  Level:       %s\n", cfg.Logging.Level))
		if cfg.Logging.File != "" {
			sb.WriteString(fmt.Sprintf("  File:        %s\n", cfg.Logging.File))
		}
		sb.WriteString(fmt.Sprintf("  Format:      %s\n", cfg.Logging.Format))
	}

	// Snapshots Configuration
	if cfg.Snapshots.Directory != "" {
		sb.WriteString("\nSnapshots Configuration\n")
		sb.WriteString(fmt.Sprintf("  Directory:   %s\n", cfg.Snapshots.Directory))
		sb.WriteString(fmt.Sprintf("  Auto Backup: %t\n", cfg.Snapshots.AutoSnapshotBeforePull))
		if cfg.Snapshots.Compression != "" {
			sb.WriteString(fmt.Sprintf("  Compression: %s\n", cfg.Snapshots.Compression))
		}
	}

	// Performance Configuration
	if cfg.Performance.ParallelDownloads > 0 {
		sb.WriteString("\nPerformance Configuration\n")
		sb.WriteString(fmt.Sprintf("  Parallel Downloads: %d\n", cfg.Performance.ParallelDownloads))
		if cfg.Performance.RsyncBandwidthLimit > 0 {
			sb.WriteString(fmt.Sprintf("  Bandwidth Limit:    %d KB/s\n", cfg.Performance.RsyncBandwidthLimit))
		}
	}

	return sb.String()
}

// FormatValue formats a single value for display
func FormatValue(value interface{}) string {
	if value == nil {
		return "<nil>"
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Slice, reflect.Array:
		// Format slices as JSON
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(data)
	case reflect.Struct, reflect.Map:
		// Format complex types as YAML
		data, err := yaml.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(data)
	default:
		return fmt.Sprintf("%v", value)
	}
}
