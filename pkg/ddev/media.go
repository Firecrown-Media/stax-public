package ddev

import (
	"fmt"
	"os"
	"path/filepath"
)

// IsMediaProxyConfigured checks if media proxy nginx config exists
func IsMediaProxyConfigured(projectPath string) bool {
	configPath := filepath.Join(projectPath, ".ddev", "nginx_full", "media-proxy.conf")
	_, err := os.Stat(configPath)
	return err == nil
}

// RemoveMediaProxyConfig removes the media proxy configuration
func RemoveMediaProxyConfig(projectPath string) error {
	nginxDir := filepath.Join(projectPath, ".ddev", "nginx_full")
	configPath := filepath.Join(nginxDir, "media-proxy.conf")
	scriptPath := filepath.Join(nginxDir, "setup-media-cache.sh")

	// Remove media proxy config
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove media proxy config: %w", err)
	}

	// Remove cache setup script (if exists)
	if err := os.Remove(scriptPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache setup script: %w", err)
	}

	// Also try to remove old cache-config.conf if it exists (backward compatibility)
	oldCachePath := filepath.Join(nginxDir, "cache-config.conf")
	os.Remove(oldCachePath) // Ignore errors for backward compat cleanup

	return nil
}

// GetMediaProxyConfig reads the current media proxy configuration
func GetMediaProxyConfig(projectPath string) (string, error) {
	return ReadNginxConfig(projectPath, "media-proxy.conf")
}

// TestMediaProxy performs basic validation of media proxy setup
func TestMediaProxy(projectPath string) error {
	// Check if config exists
	if !IsMediaProxyConfigured(projectPath) {
		return fmt.Errorf("media proxy is not configured")
	}

	// Check if DDEV is running
	mgr := NewManager(projectPath)
	running, err := mgr.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check DDEV status: %w", err)
	}
	if !running {
		return fmt.Errorf("DDEV is not running")
	}

	// Validate nginx configuration
	if err := ValidateNginxConfig(projectPath); err != nil {
		return fmt.Errorf("nginx configuration is invalid: %w", err)
	}

	return nil
}

// EnableMediaProxy enables media proxy by creating/updating configuration
func EnableMediaProxy(projectPath string, options MediaProxyOptions) error {
	options.Enabled = true
	return GenerateMediaProxyConfig(projectPath, options)
}

// DisableMediaProxy disables media proxy by removing configuration
func DisableMediaProxy(projectPath string) error {
	return RemoveMediaProxyConfig(projectPath)
}

// GetMediaProxyStatus returns detailed status of media proxy configuration
func GetMediaProxyStatus(projectPath string) (*MediaProxyStatus, error) {
	status := &MediaProxyStatus{
		Configured: IsMediaProxyConfigured(projectPath),
		ConfigPath: filepath.Join(projectPath, ".ddev", "nginx_full", "media-proxy.conf"),
	}

	// Check if DDEV is running
	mgr := NewManager(projectPath)
	running, err := mgr.IsRunning()
	if err == nil {
		status.Running = running
	}

	// If configured, try to read config
	if status.Configured {
		config, err := GetMediaProxyConfig(projectPath)
		if err == nil {
			status.ConfigContent = config
		}

		// Validate if running
		if status.Running {
			if err := ValidateNginxConfig(projectPath); err != nil {
				status.Valid = false
				status.ValidationError = err.Error()
			} else {
				status.Valid = true
			}
		}
	}

	return status, nil
}

// UpdateMediaProxyOptions updates media proxy configuration with new options
func UpdateMediaProxyOptions(projectPath string, updates map[string]interface{}) error {
	// Get current options (if they exist)
	options := GetDefaultMediaProxyOptions()

	// Apply updates
	for key, value := range updates {
		switch key {
		case "enabled":
			if v, ok := value.(bool); ok {
				options.Enabled = v
			}
		case "cdn_url":
			if v, ok := value.(string); ok {
				options.CDNURL = v
			}
		case "wpengine_url":
			if v, ok := value.(string); ok {
				options.WPEngineURL = v
			}
		case "cache_enabled":
			if v, ok := value.(bool); ok {
				options.CacheEnabled = v
			}
		case "cache_ttl":
			if v, ok := value.(string); ok {
				options.CacheTTL = v
			}
		case "cache_max_size":
			if v, ok := value.(string); ok {
				options.CacheMaxSize = v
			}
		default:
			return fmt.Errorf("unknown option: %s", key)
		}
	}

	return GenerateMediaProxyConfig(projectPath, options)
}

// MediaProxyStatus represents the status of media proxy configuration
type MediaProxyStatus struct {
	Configured      bool
	Running         bool
	Valid           bool
	ConfigPath      string
	ConfigContent   string
	ValidationError string
}
