package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
)

// CreateTestConfig creates a test configuration with default values
func CreateTestConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := config.Defaults()
	cfg.Project.Name = "test-project"
	cfg.ProviderConfig["install"] = "testinstall"
	cfg.Network.Domain = "test.local"

	return cfg
}

// CreateTestConfigFile creates a test configuration file
func CreateTestConfigFile(t *testing.T, dir string, cfg *config.Config) string {
	t.Helper()

	configPath := filepath.Join(dir, ".stax.yml")
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	return configPath
}

// CreateMultisiteConfig creates a multisite test configuration
func CreateMultisiteConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := CreateTestConfig(t)
	cfg.Project.Type = "wordpress-multisite"
	cfg.Project.Mode = "subdomain"
	cfg.Network.Sites = []config.SiteConfig{
		{
			Name:           "Site 1",
			Slug:           "site1",
			Title:          "Test Site 1",
			Domain:         "site1.test.local",
			ProviderDomain: "site1.wpengine.com",
			Active:         true,
		},
		{
			Name:           "Site 2",
			Slug:           "site2",
			Title:          "Test Site 2",
			Domain:         "site2.test.local",
			ProviderDomain: "site2.wpengine.com",
			Active:         true,
		},
	}

	return cfg
}

// MockWPEngineAPIResponse creates a mock WPEngine API response
func MockWPEngineAPIResponse(t *testing.T, data interface{}) []byte {
	t.Helper()

	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal mock response: %v", err)
	}

	return jsonData
}

// CreateMockDatabaseDump creates a mock database dump file
func CreateMockDatabaseDump(t *testing.T, path string) {
	t.Helper()

	sqlContent := `-- MySQL dump
--
-- Host: localhost    Database: testdb
-- ------------------------------------------------------

DROP TABLE IF EXISTS wp_posts;
CREATE TABLE wp_posts (
  ID bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  post_title text NOT NULL,
  post_content longtext NOT NULL,
  PRIMARY KEY (ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO wp_posts (ID, post_title, post_content) VALUES
(1, 'Hello World', 'Welcome to WordPress. This is your first post.');

DROP TABLE IF EXISTS wp_options;
CREATE TABLE wp_options (
  option_id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  option_name varchar(191) NOT NULL,
  option_value longtext NOT NULL,
  PRIMARY KEY (option_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO wp_options (option_id, option_name, option_value) VALUES
(1, 'siteurl', 'https://example.wpengine.com'),
(2, 'home', 'https://example.wpengine.com');

DROP TABLE IF EXISTS wp_blogs;
CREATE TABLE wp_blogs (
  blog_id bigint(20) NOT NULL AUTO_INCREMENT,
  site_id bigint(20) NOT NULL DEFAULT '0',
  domain varchar(200) NOT NULL DEFAULT '',
  path varchar(100) NOT NULL DEFAULT '',
  PRIMARY KEY (blog_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO wp_blogs (blog_id, site_id, domain, path) VALUES
(1, 1, 'example.wpengine.com', '/'),
(2, 1, 'site1.wpengine.com', '/');
`

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(sqlContent), 0644); err != nil {
		t.Fatalf("failed to write mock database dump: %v", err)
	}
}

// CreateMockSSHResponse creates a mock SSH command response
func CreateMockSSHResponse(command string, output string) map[string]string {
	return map[string]string{
		"command": command,
		"output":  output,
	}
}

// AssertConfigEqual asserts that two configs are equal
func AssertConfigEqual(t *testing.T, got, want *config.Config) {
	t.Helper()

	if got.Project.Name != want.Project.Name {
		t.Errorf("Project.Name: got %q, want %q", got.Project.Name, want.Project.Name)
	}

	gotInstall, _ := got.ProviderConfig["install"].(string)
	wantInstall, _ := want.ProviderConfig["install"].(string)
	if gotInstall != wantInstall {
		t.Errorf("ProviderConfig.install: got %q, want %q", gotInstall, wantInstall)
	}

	if got.Network.Domain != want.Network.Domain {
		t.Errorf("Network.Domain: got %q, want %q", got.Network.Domain, want.Network.Domain)
	}
}

// CreateMockComposerJSON creates a mock composer.json file
func CreateMockComposerJSON(t *testing.T, dir string) {
	t.Helper()

	composerJSON := `{
  "name": "test/project",
  "type": "project",
  "require": {
    "php": ">=8.1",
    "wordpress/wordpress": "^6.0"
  },
  "require-dev": {
    "phpunit/phpunit": "^9.0"
  },
  "scripts": {
    "test": "phpunit"
  }
}`

	path := filepath.Join(dir, "composer.json")
	if err := os.WriteFile(path, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}
}

// CreateMockPackageJSON creates a mock package.json file
func CreateMockPackageJSON(t *testing.T, dir string) {
	t.Helper()

	packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "scripts": {
    "build": "webpack --mode=production",
    "dev": "webpack --mode=development --watch",
    "lint": "eslint src"
  },
  "dependencies": {
    "react": "^18.0.0"
  },
  "devDependencies": {
    "webpack": "^5.0.0",
    "eslint": "^8.0.0"
  }
}`

	path := filepath.Join(dir, "package.json")
	if err := os.WriteFile(path, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}
}

// CreateMockDDEVConfig creates a mock DDEV config.yaml file
func CreateMockDDEVConfig(t *testing.T, dir string, projectName string) {
	t.Helper()

	ddevConfig := fmt.Sprintf(`name: %s
type: php
docroot: ""
php_version: "8.1"
webserver_type: nginx-fpm
router_http_port: "80"
router_https_port: "443"
mysql_version: "8.0"
`, projectName)

	ddevDir := filepath.Join(dir, ".ddev")
	if err := os.MkdirAll(ddevDir, 0755); err != nil {
		t.Fatalf("failed to create .ddev directory: %v", err)
	}

	path := filepath.Join(ddevDir, "config.yaml")
	if err := os.WriteFile(path, []byte(ddevConfig), 0644); err != nil {
		t.Fatalf("failed to write DDEV config: %v", err)
	}
}

// CreateMockWordPressInstall creates a mock WordPress installation
func CreateMockWordPressInstall(t *testing.T, dir string) {
	t.Helper()

	// Create wp-config.php
	wpConfig := `<?php
define('DB_NAME', 'testdb');
define('DB_USER', 'db');
define('DB_PASSWORD', 'db');
define('DB_HOST', 'db');
define('DB_CHARSET', 'utf8mb4');
define('DB_COLLATE', '');

$table_prefix = 'wp_';

define('WP_DEBUG', true);

if (!defined('ABSPATH')) {
    define('ABSPATH', __DIR__ . '/');
}

require_once ABSPATH . 'wp-settings.php';
`
	wpConfigPath := filepath.Join(dir, "wp-config.php")
	if err := os.WriteFile(wpConfigPath, []byte(wpConfig), 0644); err != nil {
		t.Fatalf("failed to write wp-config.php: %v", err)
	}

	// Create wp-content directories
	dirs := []string{
		"wp-content/themes",
		"wp-content/plugins",
		"wp-content/mu-plugins",
		"wp-content/uploads",
	}

	for _, d := range dirs {
		path := filepath.Join(dir, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", path, err)
		}
	}
}

// CreateGitRepo creates a mock git repository
func CreateGitRepo(t *testing.T, dir string) {
	t.Helper()

	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create a minimal git config
	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
`
	configPath := filepath.Join(gitDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}
}
