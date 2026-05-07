package cmd

import (
	"fmt"

	"github.com/firecrown-media/stax/pkg/config"
)

// generateTemplate outputs a template .stax.yml configuration to stdout
func generateTemplate() error {
	cfg := config.Defaults()
	cfg.Project.Name = "example-project"
	cfg.ProviderConfig["install"] = "example-install"
	cfg.Network.Domain = "example.ddev.site"

	data, err := cfg.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// showExample outputs an example configuration with comments to stdout
func showExample() error {
	example := `# Stax Configuration File
# Version: 1

version: 1

# Project metadata
project:
  name: myproject
  type: wordpress-multisite  # wordpress, wordpress-multisite
  mode: subdomain            # subdomain, subdirectory, single
  description: My WordPress project

# WPEngine integration
wpengine:
  install: myinstall
  environment: production    # production, staging, development
  ssh_gateway: ssh.wpengine.net
  backup:
    auto_snapshot: true
    skip_logs: true
    skip_transients: true
    skip_spam: true

# Network configuration (for multisite)
network:
  domain: myproject.ddev.site
  title: My Network
  sites: []

# DDEV configuration
ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  mysql_type: mysql
  webserver_type: nginx-fpm
  router_http_port: "80"
  router_https_port: "443"
  nfs_mount_enabled: false
  mutagen_enabled: true      # Enable on macOS for better performance
  xdebug_enabled: false
  nodejs_version: "20"
  composer_version: "2"

# Repository configuration
repository:
  url: https://github.com/org/repo.git
  branch: main
  private: true
  depth: 1
  submodules: false

# WordPress configuration
wordpress:
  version: latest
  locale: en_US
  table_prefix: wp_

# Media proxy configuration
media:
  proxy_enabled: true
  wpengine_fallback: true
  cache:
    enabled: true
    directory: .stax/media-cache
    max_size: 1GB
    ttl: 86400

# Logging configuration
logging:
  level: info
  file: ~/.stax/logs/stax.log
  format: json

# Snapshot configuration
snapshots:
  directory: ~/.stax/snapshots
  auto_snapshot_before_pull: true
  auto_snapshot_before_import: true
  retention:
    auto: 7    # days
    manual: 30 # days

# Performance configuration
performance:
  parallel_downloads: 4
  rsync_bandwidth_limit: 0
  database_import_batch_size: 1000
`

	fmt.Println(example)
	return nil
}
