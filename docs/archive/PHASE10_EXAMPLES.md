# Phase 10: Advanced Configuration Management - Examples

This document demonstrates the usage of the new configuration management commands implemented in Phase 10.

## Commands Overview

- `stax config show` - Display current configuration in formatted output
- `stax config get <path>` - Query specific configuration values
- `stax config set <path> <value>` - Update configuration values
- `stax config list` - Alias for `show` command

## Config Show Command

### Basic Usage

Display the current configuration in a human-readable format:

```bash
stax config show
```

Output:
```
Project Configuration
  Name:        stax-cli
  Type:        wordpress-multisite
  Mode:        subdomain
  Description: Stax WordPress CLI Tool

WPEngine Configuration
  Install:     staxtest
  Environment: production
  Account:     testaccount
  SSH Gateway: ssh.wpengine.net

Network Configuration
  Domain:      stax.ddev.site
  Title:       Stax Development Network

DDEV Configuration
  PHP Version:    8.1
  MySQL Version:  8.0
  MySQL Type:     mysql
  Webserver:      nginx-fpm
  Node.js:        20
  Xdebug:         false

WordPress Configuration
  Version:     latest
  Locale:      en_US
  Table Prefix: wp_

...
```

### JSON Format

Display configuration as JSON for machine-readable output:

```bash
stax config show --format json
```

Output:
```json
{
  "Version": 1,
  "Project": {
    "Name": "stax-cli",
    "Type": "wordpress-multisite",
    "Mode": "subdomain",
    "Description": "Stax WordPress CLI Tool"
  },
  "WPEngine": {
    "Install": "staxtest",
    "Environment": "production",
    "AccountName": "testaccount",
    "SSHGateway": "ssh.wpengine.net"
  },
  ...
}
```

### YAML Format

Display configuration as YAML:

```bash
stax config show --format yaml
```

Output:
```yaml
version: 1
project:
    name: stax-cli
    type: wordpress-multisite
    mode: subdomain
    description: Stax WordPress CLI Tool
wpengine:
    install: staxtest
    environment: production
    account_name: testaccount
    ssh_gateway: ssh.wpengine.net
...
```

## Config Get Command

Query specific configuration values using dot notation.

### Project Configuration

```bash
# Get project name
$ stax config get project.name
stax-cli

# Get project type
$ stax config get project.type
wordpress-multisite

# Get project mode
$ stax config get project.mode
subdomain
```

### WPEngine Configuration

```bash
# Get WPEngine install name
$ stax config get wpengine.install
staxtest

# Get WPEngine environment
$ stax config get wpengine.environment
production

# Get SSH gateway
$ stax config get wpengine.ssh_gateway
ssh.wpengine.net
```

### DDEV Configuration

```bash
# Get PHP version
$ stax config get ddev.php_version
8.1

# Get MySQL version
$ stax config get ddev.mysql_version
8.0

# Get Xdebug status
$ stax config get ddev.xdebug_enabled
false
```

### Network Configuration

```bash
# Get network domain
$ stax config get network.domain
stax.ddev.site

# Get network title
$ stax config get network.title
Stax Development Network
```

### WordPress Configuration

```bash
# Get WordPress version
$ stax config get wordpress.version
latest

# Get locale
$ stax config get wordpress.locale
en_US

# Get table prefix
$ stax config get wordpress.table_prefix
wp_
```

### Error Handling

Invalid paths return a clear error message:

```bash
$ stax config get invalid.path
✗ Config key not found: invalid.path
Error: failed to get config value: field not found: invalid
```

## Config Set Command

Update configuration values with automatic backup creation.

### Setting String Values

```bash
# Set project name
$ stax config set project.name my-new-project
✓ Updated project.name to my-new-project
  Configuration saved to .stax.yml
  Backup saved to .stax.yml.backup.20251115-174225

# Set WPEngine environment
$ stax config set wpengine.environment staging
✓ Updated wpengine.environment to staging
  Configuration saved to .stax.yml
  Backup saved to .stax.yml.backup.20251115-174226
```

### Setting Version Numbers

```bash
# Update PHP version
$ stax config set ddev.php_version 8.2
✓ Updated ddev.php_version to 8.2
  Configuration saved to .stax.yml
  Backup saved to .stax.yml.backup.20251115-174227

# Update MySQL version
$ stax config set ddev.mysql_version 8.0
✓ Updated ddev.mysql_version to 8.0
  Configuration saved to .stax.yml
  Backup saved to .stax.yml.backup.20251115-174228
```

### Setting Boolean Values

```bash
# Enable Xdebug
$ stax config set ddev.xdebug_enabled true
✓ Updated ddev.xdebug_enabled to true
  Configuration saved to .stax.yml
  Backup saved to .stax.yml.backup.20251115-174229

# Disable Xdebug
$ stax config set ddev.xdebug_enabled false
✓ Updated ddev.xdebug_enabled to false
  Configuration saved to .stax.yml
  Backup saved to .stax.yml.backup.20251115-174230
```

### Error Handling

Invalid paths are validated before modification:

```bash
$ stax config set invalid.path value
✗ Invalid config path: invalid.path
  Use 'stax config show' to see available configuration options
Error: failed to get config value: field not found: invalid
```

## Config List Command

The `list` command is an alias for `show`:

```bash
# List configuration (pretty format)
$ stax config list

# List configuration as JSON
$ stax config list --json
```

## Global Configuration

All commands support a `--global` flag to work with global configuration:

```bash
# Show global configuration
$ stax config show --global

# Get global setting
$ stax config get ddev.php_version --global

# Set global setting
$ stax config set ddev.php_version 8.2 --global
```

## Supported Configuration Paths

### Project Paths
- `project.name` - Project name
- `project.type` - Project type (wordpress, wordpress-multisite)
- `project.mode` - Multisite mode (subdomain, subdirectory, single)
- `project.description` - Project description

### WPEngine Paths
- `wpengine.install` - WPEngine install name
- `wpengine.environment` - Environment (production, staging, development)
- `wpengine.account_name` - WPEngine account name
- `wpengine.ssh_gateway` - SSH gateway hostname

### Network Paths
- `network.domain` - Network domain
- `network.title` - Network title
- `network.admin_email` - Admin email

### DDEV Paths
- `ddev.php_version` - PHP version
- `ddev.mysql_version` - MySQL version
- `ddev.mysql_type` - MySQL type (mysql, mariadb)
- `ddev.webserver_type` - Webserver type
- `ddev.xdebug_enabled` - Xdebug enabled (true/false)
- `ddev.nodejs_version` - Node.js version
- `ddev.composer_version` - Composer version

### WordPress Paths
- `wordpress.version` - WordPress version
- `wordpress.locale` - WordPress locale
- `wordpress.table_prefix` - Database table prefix

### Repository Paths
- `repository.url` - Repository URL
- `repository.branch` - Default branch
- `repository.private` - Private repository (true/false)

### Media Paths
- `media.proxy_enabled` - Enable media proxy (true/false)
- `media.wpengine_fallback` - Enable WPEngine fallback (true/false)

### Logging Paths
- `logging.level` - Log level (debug, info, warn, error)
- `logging.file` - Log file path
- `logging.format` - Log format (json, text)

### Snapshots Paths
- `snapshots.directory` - Snapshots directory
- `snapshots.auto_snapshot_before_pull` - Auto snapshot (true/false)
- `snapshots.compression` - Compression type

### Performance Paths
- `performance.parallel_downloads` - Parallel downloads count
- `performance.rsync_bandwidth_limit` - Bandwidth limit (KB/s)
- `performance.database_import_batch_size` - Batch size

## Backup and Safety

### Automatic Backups

Every `set` operation creates a timestamped backup:

```bash
$ ls -la .stax.yml*
-rw-r--r--  1 user  staff  1234 Nov 15 17:42 .stax.yml
-rw-r--r--  1 user  staff  1234 Nov 15 17:40 .stax.yml.backup.20251115-174225
-rw-r--r--  1 user  staff  1234 Nov 15 17:41 .stax.yml.backup.20251115-174226
```

### Validation

Paths are validated before modification to prevent invalid updates:

```bash
$ stax config set ddev.xdebug_enabled invalid-bool
Error: failed to set config value: invalid boolean value: invalid-bool (use true/false)
```

## Integration with Other Commands

The configuration system integrates with other Stax commands:

```bash
# Initialize a new project with specific PHP version
$ stax init myproject
$ stax config set ddev.php_version 8.2

# Pull database from WPEngine
$ stax db pull  # Uses wpengine.install and wpengine.environment

# Start DDEV with configured settings
$ stax start    # Uses ddev.* settings
```

## Tips and Best Practices

1. **Use JSON for scripting**: When automating, use `--format json` for reliable parsing
2. **Check before setting**: Use `get` to verify current values before updating
3. **Keep backups**: Backup files are created automatically but can be cleaned up periodically
4. **Validate changes**: Use `stax config show` to verify changes after `set` operations
5. **Global vs Project**: Use global config for machine-wide defaults, project config for site-specific settings
6. **Path discovery**: Use `stax config show` to discover available configuration paths

## Common Workflows

### Switching Between Environments

```bash
# Switch to staging
$ stax config set wpengine.environment staging

# Pull staging database
$ stax db pull

# Switch back to production
$ stax config set wpengine.environment production
```

### Updating PHP Version

```bash
# Check current version
$ stax config get ddev.php_version

# Update version
$ stax config set ddev.php_version 8.3

# Restart DDEV to apply changes
$ stax restart
```

### Enabling Debug Mode

```bash
# Enable Xdebug
$ stax config set ddev.xdebug_enabled true

# Update logging level
$ stax config set logging.level debug

# Restart to apply changes
$ stax restart
```
