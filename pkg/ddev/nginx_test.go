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
		"location ~ ^/wp-content/uploads/(.*)$",
		"location @proxy_media",
		"location @wpengine_fallback",
		"proxy_pass " + options.CDNURL,
		"proxy_pass " + options.WPEngineURL,
		"proxy_cache media_cache",
		"proxy_set_header X-Custom value",
		"add_header X-Proxy-Source",
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

func TestGenerateMediaProxyConfig_SeparateCacheFile(t *testing.T) {
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

	// Verify cache config file was created
	cachePath := filepath.Join(tmpDir, ".ddev", "nginx_full", "cache-config.conf")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache-config.conf was not created when CacheEnabled=true")
	}

	// Read cache config
	cacheContent, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("Failed to read cache config: %v", err)
	}

	// Verify it contains proxy_cache_path
	if !strings.Contains(string(cacheContent), "proxy_cache_path") {
		t.Error("cache-config.conf should contain proxy_cache_path directive")
	}
}

func TestGenerateMediaProxyConfig_NoCacheFileWhenDisabled(t *testing.T) {
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

	// Verify cache config file was NOT created
	cachePath := filepath.Join(tmpDir, ".ddev", "nginx_full", "cache-config.conf")
	if _, err := os.Stat(cachePath); err == nil {
		t.Error("cache-config.conf should not be created when CacheEnabled=false")
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

	// Verify files exist
	configPath := filepath.Join(tmpDir, ".ddev", "nginx_full", "media-proxy.conf")
	cachePath := filepath.Join(tmpDir, ".ddev", "nginx_full", "cache-config.conf")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("media-proxy.conf was not created")
	}
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache-config.conf was not created")
	}

	// Now remove them
	err = RemoveMediaProxyConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to remove config: %v", err)
	}

	// Verify files are gone
	if _, err := os.Stat(configPath); err == nil {
		t.Error("media-proxy.conf should be removed")
	}
	if _, err := os.Stat(cachePath); err == nil {
		t.Error("cache-config.conf should be removed")
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
