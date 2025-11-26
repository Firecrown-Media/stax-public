package ddev

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateMediaProxyConfig_BasicGeneration(t *testing.T) {
	options := MediaProxyOptions{
		Enabled:      true,
		CDNURL:       "https://example.b-cdn.net",
		WPEngineURL:  "https://example.wpengine.com",
		WPEngineHost: "example.wpengine.com",
		CacheEnabled: true,
		CacheTTL:     "30d",
		CacheMaxSize: "10g",
	}

	tmpDir := t.TempDir()
	err := GenerateMediaProxyConfig(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, ".ddev", "nginx_full", "media-proxy.conf")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("media-proxy.conf was not created")
	}

	// Read generated config
	config, err := GetMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Verify it's not empty
	if len(config) == 0 {
		t.Fatal("Generated config is empty")
	}
}

func TestGenerateMediaProxyConfig_ContainsRequiredDirectives(t *testing.T) {
	options := MediaProxyOptions{
		Enabled:      true,
		CDNName:      "BunnyCDN",
		CDNURL:       "https://example.b-cdn.net",
		WPEngineURL:  "https://example.wpengine.com",
		WPEngineHost: "example.wpengine.com",
		CacheEnabled: true,
		CacheTTL:     "7d",
		CacheMaxSize: "5g",
		ProxyHeaders: map[string]string{
			"X-Custom": "value",
		},
	}

	tmpDir := t.TempDir()
	err := GenerateMediaProxyConfig(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	config, err := GetMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Verify critical directives are present
	required := []string{
		"server {",                              // Server block wrapper
		"location ~ ^/wp-content/uploads/(.*)$", // Upload location
		"location @proxy_media",                 // Proxy location
		"location @wpengine_fallback",           // Fallback location
		"proxy_pass " + options.CDNURL,          // CDN URL
		"proxy_pass " + options.WPEngineURL,     // WPEngine URL
		"proxy_cache media_cache",               // Cache directive
		"proxy_cache_path",                      // HTTP-level cache config
		"proxy_set_header X-Custom value",       // Custom header
		"add_header X-Proxy-Source",             // Proxy source header
	}

	for _, req := range required {
		if !strings.Contains(config, req) {
			t.Errorf("Config missing required directive: %s", req)
		}
	}
}

func TestGenerateMediaProxyConfig_WithoutCache(t *testing.T) {
	options := MediaProxyOptions{
		Enabled:      true,
		CDNURL:       "https://example.b-cdn.net",
		WPEngineURL:  "https://example.wpengine.com",
		CacheEnabled: false, // Cache disabled
	}

	tmpDir := t.TempDir()
	err := GenerateMediaProxyConfig(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	config, err := GetMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Verify cache directives are NOT present when disabled
	// Note: proxy_cache_path in comments is OK, but active proxy_cache directives should not be there
	cacheDirectives := []string{
		"proxy_cache media_cache;",
		"proxy_cache_valid",
		"proxy_cache_key",
	}

	for _, directive := range cacheDirectives {
		if strings.Contains(config, directive) {
			t.Errorf("Config should not contain '%s' when CacheEnabled=false", directive)
		}
	}
}

func TestGenerateMediaProxyConfig_WithCache(t *testing.T) {
	options := MediaProxyOptions{
		Enabled:      true,
		CDNURL:       "https://example.b-cdn.net",
		WPEngineURL:  "https://example.wpengine.com",
		CacheEnabled: true,
		CacheTTL:     "30d",
		CacheMaxSize: "10g",
	}

	tmpDir := t.TempDir()
	err := GenerateMediaProxyConfig(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	// Verify cache setup script was created
	scriptPath := filepath.Join(tmpDir, ".ddev", "nginx_full", "setup-media-cache.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Fatal("setup-media-cache.sh was not created when CacheEnabled=true")
	}

	// Read the main config
	config, err := GetMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Verify it contains proxy_cache_path (now included in main config)
	if !strings.Contains(config, "proxy_cache_path") {
		t.Error("Config should contain proxy_cache_path directive when caching enabled")
	}

	// Verify cache size is correct
	if !strings.Contains(config, "max_size=10g") {
		t.Error("Config should contain correct max_size")
	}
}

func TestGenerateMediaProxyConfig_NoCacheWhenDisabled(t *testing.T) {
	options := MediaProxyOptions{
		Enabled:      true,
		CDNURL:       "https://example.b-cdn.net",
		WPEngineURL:  "https://example.wpengine.com",
		CacheEnabled: false,
	}

	tmpDir := t.TempDir()
	err := GenerateMediaProxyConfig(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	// Verify cache setup script was NOT created
	scriptPath := filepath.Join(tmpDir, ".ddev", "nginx_full", "setup-media-cache.sh")
	if _, err := os.Stat(scriptPath); err == nil {
		t.Error("setup-media-cache.sh should not be created when CacheEnabled=false")
	}

	// Read the main config
	config, err := GetMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Verify proxy_cache_path is NOT in config when caching disabled
	if strings.Contains(config, "proxy_cache_path") {
		t.Error("Config should not contain proxy_cache_path when CacheEnabled=false")
	}
}

func TestGenerateMediaProxyConfig_Defaults(t *testing.T) {
	options := MediaProxyOptions{
		Enabled:      true,
		CDNURL:       "https://example.b-cdn.net",
		WPEngineURL:  "https://example.wpengine.com",
		CacheEnabled: true,
		// No TTL, max size, etc. - test defaults are applied
	}

	tmpDir := t.TempDir()
	err := GenerateMediaProxyConfig(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	config, err := GetMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Defaults should be applied
	if !strings.Contains(config, "30d") { // Default TTL
		t.Error("Default cache TTL (30d) not applied")
	}

	// Default proxy headers should be present
	if !strings.Contains(config, "X-Real-IP") {
		t.Error("Default X-Real-IP header not applied")
	}
	if !strings.Contains(config, "X-Forwarded-For") {
		t.Error("Default X-Forwarded-For header not applied")
	}
}

func TestRemoveMediaProxyConfig(t *testing.T) {
	// First create a config
	options := MediaProxyOptions{
		Enabled:      true,
		CDNURL:       "https://example.b-cdn.net",
		WPEngineURL:  "https://example.wpengine.com",
		CacheEnabled: true,
	}

	tmpDir := t.TempDir()
	err := GenerateMediaProxyConfig(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".ddev", "nginx_full", "media-proxy.conf")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("media-proxy.conf was not created")
	}

	// Now remove it
	err = RemoveMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to remove config: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(configPath); err == nil {
		t.Error("media-proxy.conf should be removed")
	}
}

func TestRemoveMediaProxyConfig_NoErrorWhenNotExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to remove config that doesn't exist - should not error
	err := RemoveMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("RemoveMediaProxyConfig should not error when files don't exist: %v", err)
	}
}

func TestIsMediaProxyConfigured(t *testing.T) {
	tmpDir := t.TempDir()

	// Initially not configured
	if IsMediaProxyConfigured(tmpDir) {
		t.Error("Should not be configured initially")
	}

	// Generate config
	options := MediaProxyOptions{
		Enabled:     true,
		CDNURL:      "https://example.b-cdn.net",
		WPEngineURL: "https://example.wpengine.com",
	}

	err := GenerateMediaProxyConfig(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to generate config: %v", err)
	}

	// Now should be configured
	if !IsMediaProxyConfigured(tmpDir) {
		t.Error("Should be configured after GenerateMediaProxyConfig")
	}

	// Remove config
	err = RemoveMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to remove config: %v", err)
	}

	// Should not be configured anymore
	if IsMediaProxyConfigured(tmpDir) {
		t.Error("Should not be configured after RemoveMediaProxyConfig")
	}
}

func TestGetDefaultMediaProxyOptions(t *testing.T) {
	options := GetDefaultMediaProxyOptions()

	// Verify defaults are set
	if !options.Enabled {
		t.Error("Default should be enabled")
	}
	if options.CDNName != "BunnyCDN" {
		t.Errorf("Default CDN name should be 'BunnyCDN', got '%s'", options.CDNName)
	}
	if options.CacheTTL != "30d" {
		t.Errorf("Default cache TTL should be '30d', got '%s'", options.CacheTTL)
	}
	if options.CacheMaxSize != "10g" {
		t.Errorf("Default cache max size should be '10g', got '%s'", options.CacheMaxSize)
	}
	if !options.CacheEnabled {
		t.Error("Default should have cache enabled")
	}
	if options.ProxyHeaders == nil {
		t.Error("Default proxy headers should not be nil")
	}
}

func TestGenerateMediaProxyConfig_DisabledDoesNotGenerate(t *testing.T) {
	options := MediaProxyOptions{
		Enabled:     false, // Disabled
		CDNURL:      "https://example.b-cdn.net",
		WPEngineURL: "https://example.wpengine.com",
	}

	tmpDir := t.TempDir()
	err := GenerateMediaProxyConfig(tmpDir, options)
	if err != nil {
		t.Fatalf("Should not error when disabled: %v", err)
	}

	// Verify no config file was created
	configPath := filepath.Join(tmpDir, ".ddev", "nginx_full", "media-proxy.conf")
	if _, err := os.Stat(configPath); err == nil {
		t.Error("media-proxy.conf should not be created when Enabled=false")
	}
}

func TestValidateNginxMediaProxyConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid server block config",
			config: `server {
				location ~ ^/wp-content/uploads/(.*)$ {
					try_files $uri @proxy;
				}
			}`,
			shouldError: false,
		},
		{
			name: "invalid - bare location block",
			config: `location ~ ^/wp-content/uploads/(.*)$ {
				try_files $uri @proxy;
			}`,
			shouldError: true,
			errorMsg:    "must be wrapped in 'server {' block",
		},
		{
			name: "invalid - location outside server block",
			config: `# Comment
			proxy_cache_path /var/cache;
			location / {
				try_files $uri;
			}`,
			shouldError: true,
			errorMsg:    "must contain a 'server {' block",
		},
		{
			name: "valid - with cache path and server block",
			config: `proxy_cache_path /var/cache;
			server {
				location / {
					try_files $uri;
				}
			}`,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNginxMediaProxyConfig(tt.config)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}
