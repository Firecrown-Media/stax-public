package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	// Test version
	if cfg.Version != 2 {
		t.Errorf("expected version 2, got %d", cfg.Version)
	}

	// Test provider default
	if cfg.Provider != "wpengine" {
		t.Errorf("expected default provider 'wpengine', got %q", cfg.Provider)
	}

	// Test ProviderConfig is initialized
	if cfg.ProviderConfig == nil {
		t.Error("expected ProviderConfig to be initialized, got nil")
	}

	// Test project defaults
	if cfg.Project.Type != "wordpress-multisite" {
		t.Errorf("expected project type 'wordpress-multisite', got %q", cfg.Project.Type)
	}
	if cfg.Project.Mode != "subdomain" {
		t.Errorf("expected project mode 'subdomain', got %q", cfg.Project.Mode)
	}

	// Test DDEV defaults
	if cfg.DDEV.PHPVersion != "8.1" {
		t.Errorf("expected PHP version '8.1', got %q", cfg.DDEV.PHPVersion)
	}
	if cfg.DDEV.MySQLVersion != "8.0" {
		t.Errorf("expected MySQL version '8.0', got %q", cfg.DDEV.MySQLVersion)
	}
	if cfg.DDEV.WebserverType != "nginx-fpm" {
		t.Errorf("expected webserver type 'nginx-fpm', got %q", cfg.DDEV.WebserverType)
	}
	if cfg.DDEV.NodeJSVersion != "20" {
		t.Errorf("expected Node.js version '20', got %q", cfg.DDEV.NodeJSVersion)
	}

	// Test provider_config defaults
	if env, _ := cfg.ProviderConfig["environment"].(string); env != "production" {
		t.Errorf("expected provider_config.environment 'production', got %q", env)
	}
	if gw, _ := cfg.ProviderConfig["ssh_gateway"].(string); gw != "ssh.wpengine.net" {
		t.Errorf("expected provider_config.ssh_gateway 'ssh.wpengine.net', got %q", gw)
	}

	// Test WordPress defaults
	if cfg.WordPress.Version != "latest" {
		t.Errorf("expected WordPress version 'latest', got %q", cfg.WordPress.Version)
	}
	if cfg.WordPress.TablePrefix != "wp_" {
		t.Errorf("expected table prefix 'wp_', got %q", cfg.WordPress.TablePrefix)
	}

	// Test logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("expected log level 'info', got %q", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected log format 'json', got %q", cfg.Logging.Format)
	}

	// Test snapshots defaults
	if !cfg.Snapshots.AutoSnapshotBeforePull {
		t.Error("expected auto snapshot before pull to be enabled")
	}
	if cfg.Snapshots.Retention.Auto != 7 {
		t.Errorf("expected auto retention 7 days, got %d", cfg.Snapshots.Retention.Auto)
	}

	// Test performance defaults
	if cfg.Performance.ParallelDownloads != 4 {
		t.Errorf("expected 4 parallel downloads, got %d", cfg.Performance.ParallelDownloads)
	}
}

func TestConfig_ToYAML(t *testing.T) {
	cfg := &Config{
		Version: 2,
		Project: ProjectConfig{
			Name: "test-project",
			Type: "wordpress-multisite",
			Mode: "subdomain",
		},
		Provider: "wpengine",
		ProviderConfig: map[string]any{
			"install":     "testinstall",
			"environment": "production",
		},
	}

	data, err := cfg.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty YAML data")
	}

	// Verify it's valid YAML by unmarshaling
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Errorf("generated YAML is invalid: %v", err)
	}

	// Verify specific fields
	if result["version"] != 2 {
		t.Errorf("expected version 2, got %v", result["version"])
	}
}

func TestConfig_FromYAML(t *testing.T) {
	yamlData := `version: 2
project:
  name: test-from-yaml
  type: wordpress-multisite
  mode: subdomain
provider: wpengine
provider_config:
  install: yamlinstall
  environment: staging
ddev:
  php_version: "8.2"
  mysql_version: "8.0"
`

	cfg := &Config{}
	err := cfg.FromYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("FromYAML() failed: %v", err)
	}

	// Verify fields
	if cfg.Version != 2 {
		t.Errorf("expected version 2, got %d", cfg.Version)
	}
	if cfg.Project.Name != "test-from-yaml" {
		t.Errorf("expected project name 'test-from-yaml', got %q", cfg.Project.Name)
	}
	install, _ := cfg.ProviderConfig["install"].(string)
	if install != "yamlinstall" {
		t.Errorf("expected provider_config.install 'yamlinstall', got %q", install)
	}
	if cfg.DDEV.PHPVersion != "8.2" {
		t.Errorf("expected PHP version '8.2', got %q", cfg.DDEV.PHPVersion)
	}
}

func TestConfig_RoundTrip(t *testing.T) {
	// Test that marshaling and unmarshaling preserves data
	original := Defaults()
	original.Project.Name = "roundtrip-test"
	original.ProviderConfig["install"] = "testinstall"

	// Marshal to YAML
	data, err := original.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() failed: %v", err)
	}

	// Unmarshal back
	result := &Config{}
	if err := result.FromYAML(data); err != nil {
		t.Fatalf("FromYAML() failed: %v", err)
	}

	// Compare key fields
	if result.Project.Name != original.Project.Name {
		t.Errorf("project name: got %q, want %q", result.Project.Name, original.Project.Name)
	}
	origInstall, _ := original.ProviderConfig["install"].(string)
	resultInstall, _ := result.ProviderConfig["install"].(string)
	if resultInstall != origInstall {
		t.Errorf("provider_config.install: got %q, want %q", resultInstall, origInstall)
	}
	if result.DDEV.PHPVersion != original.DDEV.PHPVersion {
		t.Errorf("PHP version: got %q, want %q", result.DDEV.PHPVersion, original.DDEV.PHPVersion)
	}
}

func TestProjectConfig(t *testing.T) {
	tests := []struct {
		name   string
		config ProjectConfig
		check  func(t *testing.T, cfg ProjectConfig)
	}{
		{
			name: "multisite subdomain",
			config: ProjectConfig{
				Name: "test",
				Type: "wordpress-multisite",
				Mode: "subdomain",
			},
			check: func(t *testing.T, cfg ProjectConfig) {
				if cfg.Type != "wordpress-multisite" {
					t.Errorf("expected 'wordpress-multisite', got %q", cfg.Type)
				}
				if cfg.Mode != "subdomain" {
					t.Errorf("expected 'subdomain', got %q", cfg.Mode)
				}
			},
		},
		{
			name: "single site",
			config: ProjectConfig{
				Name: "single",
				Type: "wordpress",
				Mode: "single",
			},
			check: func(t *testing.T, cfg ProjectConfig) {
				if cfg.Type != "wordpress" {
					t.Errorf("expected 'wordpress', got %q", cfg.Type)
				}
				if cfg.Mode != "single" {
					t.Errorf("expected 'single', got %q", cfg.Mode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.config)
		})
	}
}

func TestNetworkConfig(t *testing.T) {
	config := NetworkConfig{
		Domain: "test.local",
		Title:  "Test Network",
		Sites: []SiteConfig{
			{
				Name:   "Site 1",
				Slug:   "site1",
				Domain: "site1.test.local",
				Active: true,
			},
		},
	}

	if config.Domain != "test.local" {
		t.Errorf("expected domain 'test.local', got %q", config.Domain)
	}

	if len(config.Sites) != 1 {
		t.Errorf("expected 1 site, got %d", len(config.Sites))
	}

	if config.Sites[0].Name != "Site 1" {
		t.Errorf("expected site name 'Site 1', got %q", config.Sites[0].Name)
	}
}

func TestSearchReplaceConfig(t *testing.T) {
	config := SearchReplaceConfig{
		Network: []SearchReplacePair{
			{Old: "https://old.com", New: "https://new.local"},
		},
		Sites: []SiteSearchReplace{
			{
				Old: "https://site1.old.com",
				New: "https://site1.new.local",
				URL: "https://site1.new.local",
			},
		},
		SkipColumns: []string{"guid"},
		SkipTables:  []string{"wp_logs"},
	}

	if len(config.Network) != 1 {
		t.Errorf("expected 1 network pair, got %d", len(config.Network))
	}

	if config.Network[0].Old != "https://old.com" {
		t.Errorf("expected old 'https://old.com', got %q", config.Network[0].Old)
	}

	if len(config.Sites) != 1 {
		t.Errorf("expected 1 site pair, got %d", len(config.Sites))
	}

	if len(config.SkipColumns) != 1 || config.SkipColumns[0] != "guid" {
		t.Errorf("expected skip columns ['guid'], got %v", config.SkipColumns)
	}
}

func TestBuildConfig(t *testing.T) {
	config := BuildConfig{
		Scripts: BuildScriptsConfig{
			Main: "build.sh",
			PreBuild: []string{
				"composer install",
			},
		},
		Composer: BuildComposerConfig{
			Optimize: true,
			NoDev:    true,
		},
		NPM: BuildNPMConfig{
			BuildCommand: "npm run build",
			DevCommand:   "npm run dev",
		},
	}

	if config.Scripts.Main != "build.sh" {
		t.Errorf("expected main script 'build.sh', got %q", config.Scripts.Main)
	}

	if !config.Composer.Optimize {
		t.Error("expected composer optimize to be true")
	}

	if config.NPM.BuildCommand != "npm run build" {
		t.Errorf("expected npm build command 'npm run build', got %q", config.NPM.BuildCommand)
	}
}

func TestMediaConfig(t *testing.T) {
	config := MediaConfig{
		ProxyEnabled:     true,
		PrimarySource:    "wpengine",
		WPEngineFallback: true,
		Cache: CacheConfig{
			Enabled:   true,
			Directory: ".stax/media-cache",
			MaxSize:   "1GB",
			TTL:       86400,
		},
	}

	if !config.ProxyEnabled {
		t.Error("expected proxy to be enabled")
	}

	if config.PrimarySource != "wpengine" {
		t.Errorf("expected primary source 'wpengine', got %q", config.PrimarySource)
	}

	if !config.Cache.Enabled {
		t.Error("expected cache to be enabled")
	}

	if config.Cache.TTL != 86400 {
		t.Errorf("expected TTL 86400, got %d", config.Cache.TTL)
	}
}

func TestSnapshotsConfig(t *testing.T) {
	config := SnapshotsConfig{
		Directory:                "~/.stax/snapshots",
		AutoSnapshotBeforePull:   true,
		AutoSnapshotBeforeImport: true,
		Retention: RetentionConfig{
			Auto:   7,
			Manual: 30,
		},
		Compression: "gzip",
	}

	if !config.AutoSnapshotBeforePull {
		t.Error("expected auto snapshot before pull to be enabled")
	}

	if config.Retention.Auto != 7 {
		t.Errorf("expected auto retention 7, got %d", config.Retention.Auto)
	}

	if config.Compression != "gzip" {
		t.Errorf("expected compression 'gzip', got %q", config.Compression)
	}
}

func TestPerformanceConfig(t *testing.T) {
	config := PerformanceConfig{
		ParallelDownloads:       4,
		RsyncBandwidthLimit:     1000,
		DatabaseImportBatchSize: 1000,
	}

	if config.ParallelDownloads != 4 {
		t.Errorf("expected 4 parallel downloads, got %d", config.ParallelDownloads)
	}

	if config.RsyncBandwidthLimit != 1000 {
		t.Errorf("expected bandwidth limit 1000, got %d", config.RsyncBandwidthLimit)
	}

	if config.DatabaseImportBatchSize != 1000 {
		t.Errorf("expected batch size 1000, got %d", config.DatabaseImportBatchSize)
	}
}
