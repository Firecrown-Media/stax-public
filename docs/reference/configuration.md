# Stax Configuration Specification

## Overview

Stax uses YAML-based configuration files to manage project settings, with support for global defaults and per-project overrides. Configuration is stored in two locations:

1. **Global Configuration**: `~/.stax/config.yml` - User-wide defaults
2. **Project Configuration**: `<project>/.stax.yml` - Project-specific settings

Project configuration takes precedence over global configuration. Environment variables can override both.

## Configuration Priority (Highest to Lowest)

1. Command-line flags
2. Environment variables
3. Project configuration (`.stax.yml`)
4. Global configuration (`~/.stax/config.yml`)
5. Built-in defaults

## Configuration Schema

### Complete Schema Example

```yaml
# .stax.yml - Project Configuration
version: 1

# Project metadata
project:
  name: firecrown-multisite
  type: wordpress-multisite
  mode: subdomain  # subdomain | subdirectory
  description: "Firecrown Media multisite development environment"

# WPEngine integration
wpengine:
  install: fsmultisite
  environment: production  # production | staging | development
  account_name: firecrown-media
  ssh_gateway: ssh.wpengine.net

  # Backup preferences
  backup:
    auto_snapshot: true
    skip_logs: true
    skip_transients: true
    skip_spam: true
    exclude_tables:
      - wp_actionscheduler_logs
      - wp_wc_admin_notes

  # Domain mapping
  domains:
    production:
      primary: fsmultisite.wpenginepowered.com
      sites:
        - flyingmag.com
        - planeandpilotmag.com
        - finescale.com
        - avweb.com
    staging:
      primary: fsmultisite-staging.wpengine.com
      sites:
        - staging-flyingmag.com
        - staging-planeandpilot.com
        - staging-finescale.com
        - staging-avweb.com

# Network and sites configuration
network:
  domain: firecrown.local
  title: "Firecrown Media Network"
  admin_email: admin@firecrown.local

  # Individual sites
  sites:
    - name: flyingmag
      slug: flyingmag
      title: "Flying Magazine"
      domain: flyingmag.firecrown.local
      wpengine_domain: flyingmag.com
      active: true

    - name: planeandpilot
      slug: planeandpilot
      title: "Plane & Pilot"
      domain: planeandpilot.firecrown.local
      wpengine_domain: planeandpilotmag.com
      active: true

    - name: finescale
      slug: finescale
      title: "FineScale Modeler"
      domain: finescale.firecrown.local
      wpengine_domain: finescale.com
      active: true

    - name: avweb
      slug: avweb
      title: "AVWeb"
      domain: avweb.firecrown.local
      wpengine_domain: avweb.com
      active: true

# DDEV configuration
ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  mysql_type: mysql  # mysql | mariadb
  webserver_type: nginx-fpm  # nginx-fpm | apache-fpm

  # Ports
  router_http_port: "80"
  router_https_port: "443"
  mailhog_port: "8025"

  # Performance
  nfs_mount_enabled: false
  mutagen_enabled: false

  # Development tools
  xdebug_enabled: false
  nodejs_version: "20"
  composer_version: "2"

  # Additional hostnames (beyond auto-generated)
  additional_hostnames: []

  # Additional fully-qualified domain names
  additional_fqdns: []

  # Custom DDEV commands
  custom_commands:
    - name: stax-sync
      description: "Sync database from WPEngine"
      enabled: true
    - name: stax-build
      description: "Run build process"
      enabled: true

  # Hooks
  hooks:
    pre_start: []
    post_start:
      - exec: composer install
      - exec: npm install
      - exec: bash scripts/build.sh
    pre_stop: []
    post_stop: []

# GitHub repository configuration
repository:
  url: https://github.com/Firecrown-Media/firecrown-multisite.git
  branch: main
  private: true

  # Clone options
  depth: 1  # Shallow clone depth (0 = full clone)
  submodules: false

  # Deployment
  deploy:
    workflow: deploy-wpengine.yml
    on_push:
      branches:
        - main
        - staging

# Build process configuration
build:
  # Pre-installation commands
  pre_install: []

  # Installation commands
  install:
    - composer install --no-dev --optimize-autoloader
    - npm install --production

  # Post-installation commands
  post_install:
    - bash scripts/build.sh

  # Watch mode (for development)
  watch:
    enabled: false
    paths:
      - wp-content/themes/**/*.scss
      - wp-content/themes/**/*.js
    command: npm run watch

# WordPress configuration
wordpress:
  version: latest  # latest | 6.4.2 | specific version
  locale: en_US

  # wp-config.php constants
  constants:
    WP_DEBUG: true
    WP_DEBUG_LOG: true
    WP_DEBUG_DISPLAY: false
    SCRIPT_DEBUG: false
    SAVEQUERIES: false
    WP_ENVIRONMENT_TYPE: local

    # Multisite-specific
    WP_ALLOW_MULTISITE: true
    MULTISITE: true
    SUBDOMAIN_INSTALL: true
    DOMAIN_CURRENT_SITE: firecrown.local
    PATH_CURRENT_SITE: /
    SITE_ID_CURRENT_SITE: 1
    BLOG_ID_CURRENT_SITE: 1

    # Custom constants
    DISABLE_WP_CRON: false
    WP_POST_REVISIONS: 5
    AUTOSAVE_INTERVAL: 160
    WP_MEMORY_LIMIT: 256M
    WP_MAX_MEMORY_LIMIT: 512M

  # Database table prefix (detected automatically if not specified)
  table_prefix: wp_

  # Search-replace configuration
  search_replace:
    network:
      - old: fsmultisite.wpenginepowered.com
        new: firecrown.local
    sites:
      - old: flyingmag.com
        new: flyingmag.firecrown.local
        url: flyingmag.com
      - old: planeandpilotmag.com
        new: planeandpilot.firecrown.local
        url: planeandpilotmag.com
      - old: finescale.com
        new: finescale.firecrown.local
        url: finescale.com
      - old: avweb.com
        new: avweb.firecrown.local
        url: avweb.com
    skip_columns:
      - guid
    skip_tables:
      - wp_users

# Remote media configuration
media:
  proxy_enabled: true

  # Primary source
  primary_source: bunnycdn

  # BunnyCDN configuration
  bunnycdn:
    hostname: cdn.firecrown.com
    pull_zone: firecrown-media
    storage_zone: firecrown-storage

  # WPEngine fallback
  wpengine_fallback: true

  # Local caching
  cache:
    enabled: true
    directory: .stax/media-cache
    max_size: 1GB
    ttl: 86400  # 24 hours in seconds

# Credentials (reference only - stored in Keychain)
credentials:
  wpengine:
    keychain_service: com.firecrown.stax.wpengine
    keychain_account: fsmultisite

  github:
    keychain_service: com.firecrown.stax.github
    keychain_account: firecrown-media

  ssh:
    keychain_service: com.firecrown.stax.ssh
    keychain_account: wpengine

# Logging and debugging
logging:
  level: info  # debug | info | warn | error
  file: ~/.stax/logs/stax.log
  max_size: 10MB
  max_backups: 5
  max_age: 30  # days

  # Format
  format: json  # json | text
  timestamp: true
  caller: false

# Snapshots
snapshots:
  directory: ~/.stax/snapshots
  auto_snapshot_before_pull: true
  auto_snapshot_before_import: true
  retention:
    auto: 7  # days
    manual: 30  # days
  compression: gzip

# Performance tuning
performance:
  parallel_downloads: 4
  rsync_bandwidth_limit: 0  # KB/s, 0 = unlimited
  database_import_batch_size: 1000

# Notifications
notifications:
  enabled: false
  on_success: false
  on_error: true
  services:
    slack:
      enabled: false
      webhook_url: ""
    email:
      enabled: false
      recipient: ""

# Custom scripts
scripts:
  before_init: []
  after_init: []
  before_start: []
  after_start: []
  before_stop: []
  after_stop: []
  before_db_pull: []
  after_db_pull: []
```

## Global Configuration

Location: `~/.stax/config.yml`

Global configuration provides defaults for all projects and stores user-wide preferences.

### Global Configuration Example

```yaml
# ~/.stax/config.yml
version: 1

# User preferences
user:
  name: Geoff Hickman
  email: geoff@firecrown.com

# Default DDEV settings
ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  webserver_type: nginx-fpm
  xdebug_enabled: false
  nodejs_version: "20"

# WPEngine defaults
wpengine:
  account_name: firecrown-media
  ssh_gateway: ssh.wpengine.net
  backup:
    auto_snapshot: true
    skip_logs: true
    skip_transients: true
    skip_spam: true

# GitHub defaults
github:
  organization: Firecrown-Media

# Logging
logging:
  level: info
  file: ~/.stax/logs/stax.log

# Snapshots
snapshots:
  directory: ~/.stax/snapshots
  retention:
    auto: 7
    manual: 30

# Media proxy defaults
media:
  proxy_enabled: true
  cache:
    enabled: true
    max_size: 1GB
```

## Project Configuration

Location: `<project>/.stax.yml`

Project configuration is specific to each project and should be committed to version control.

### Minimal Project Configuration

```yaml
version: 1

project:
  name: firecrown-multisite
  type: wordpress-multisite
  mode: subdomain

wpengine:
  install: fsmultisite
  environment: production

network:
  domain: firecrown.local
  sites:
    - name: flyingmag
      domain: flyingmag.firecrown.local
      wpengine_domain: flyingmag.com

repository:
  url: https://github.com/Firecrown-Media/firecrown-multisite.git
  branch: main
```

## Environment Variables

Environment variables can override any configuration value using the following format:

```
STAX_<SECTION>_<KEY>=value
```

### Environment Variable Mapping

| Environment Variable | Configuration Path | Example |
|---------------------|-------------------|---------|
| `STAX_PROJECT_NAME` | `project.name` | `firecrown-multisite` |
| `STAX_WPENGINE_INSTALL` | `wpengine.install` | `fsmultisite` |
| `STAX_WPENGINE_ENVIRONMENT` | `wpengine.environment` | `production` |
| `STAX_DDEV_PHP_VERSION` | `ddev.php_version` | `8.1` |
| `STAX_DDEV_MYSQL_VERSION` | `ddev.mysql_version` | `8.0` |
| `STAX_LOGGING_LEVEL` | `logging.level` | `debug` |
| `STAX_CONFIG` | (special) | Path to config file |
| `STAX_DEBUG` | (special) | `true` / `false` |
| `STAX_NO_COLOR` | (special) | `true` / `false` |

### Nested Values

For nested configuration values, use underscores to separate levels:

```bash
STAX_WPENGINE_BACKUP_AUTO_SNAPSHOT=true
STAX_DDEV_HOOKS_POST_START_0="composer install"
STAX_MEDIA_CACHE_MAX_SIZE=2GB
```

### Example Usage

```bash
# Override WPEngine environment
STAX_WPENGINE_ENVIRONMENT=staging stax db:pull

# Override PHP version
STAX_DDEV_PHP_VERSION=8.2 stax start

# Enable debug logging
STAX_DEBUG=true stax init

# Use custom config file
STAX_CONFIG=.stax.staging.yml stax start
```

## Credential Storage (macOS Keychain)

Stax stores sensitive credentials in the macOS Keychain, never in configuration files. Credentials are referenced in configuration but stored securely.

### Keychain Structure

**Service Naming Convention**: `com.firecrown.stax.<service>`

| Keychain Service | Keychain Account | Stored Value |
|-----------------|------------------|--------------|
| `com.firecrown.stax.wpengine` | `<install_name>` | API credentials (JSON) |
| `com.firecrown.stax.github` | `<organization>` | Personal access token |
| `com.firecrown.stax.ssh` | `wpengine` | SSH private key |

### WPEngine Credentials

Stored as JSON in Keychain:

```json
{
  "api_user": "myuser@example.com",
  "api_password": "mypassword",
  "ssh_user": "fsmultisite",
  "ssh_gateway": "ssh.wpengine.net"
}
```

**Keychain Item**:
- Service: `com.firecrown.stax.wpengine`
- Account: `fsmultisite` (install name)
- Password: JSON string above

### GitHub Credentials

**Keychain Item**:
- Service: `com.firecrown.stax.github`
- Account: `Firecrown-Media` (organization)
- Password: `ghp_xxxxxxxxxxxxxxxxxxxx` (personal access token)

### SSH Private Key

**Keychain Item**:
- Service: `com.firecrown.stax.ssh`
- Account: `wpengine`
- Password: SSH private key contents (PEM format)

### Credential Management Commands

```bash
# Setup credentials (interactive)
stax setup

# View stored credentials (masked)
stax config:get credentials --global

# Update specific credential
stax setup --wpengine-user=newuser@example.com

# Remove credentials
stax config:unset credentials.wpengine --global
```

### Credential Access in Code

```go
// pkg/credentials/keychain.go
func GetWPEngineCredentials(install string) (*WPEngineCredentials, error) {
    service := "com.firecrown.stax.wpengine"
    account := install

    password, err := keychain.GetPassword(service, account)
    if err != nil {
        return nil, err
    }

    var creds WPEngineCredentials
    if err := json.Unmarshal([]byte(password), &creds); err != nil {
        return nil, err
    }

    return &creds, nil
}
```

## Configuration Validation

Stax validates configuration on every command execution. Validation errors are reported clearly with suggestions.

### Validation Rules

#### Required Fields

```yaml
# Minimum required for stax init
project:
  name: required
  type: required
  mode: required

wpengine:
  install: required
  environment: required

network:
  domain: required
```

#### Field Validation

| Field | Validation Rules |
|-------|-----------------|
| `project.name` | Alphanumeric, hyphens, underscores only |
| `project.type` | Must be `wordpress`, `wordpress-multisite` |
| `project.mode` | Must be `subdomain`, `subdirectory` |
| `wpengine.environment` | Must be `production`, `staging`, `development` |
| `ddev.php_version` | Must be valid PHP version (7.4, 8.0, 8.1, 8.2, 8.3) |
| `ddev.mysql_version` | Must be valid MySQL/MariaDB version |
| `network.domain` | Valid domain format |
| `network.sites[].domain` | Valid domain format, unique |

#### Cross-field Validation

```yaml
# Subdomain mode requires wildcard-compatible domains
project:
  mode: subdomain

network:
  domain: firecrown.local  # Must not have subdomain
  sites:
    - domain: site1.firecrown.local  # Must be subdomain of network.domain
```

### Validation Command

```bash
stax config:validate
```

**Output:**
```
✓ Configuration is valid

Warnings:
  - PHP version 8.1 is older than WPEngine production (8.2)
    Recommendation: Update ddev.php_version to "8.2"

  - Xdebug is enabled (may impact performance)
    Recommendation: Disable Xdebug in production configs

Suggestions:
  - Consider enabling media proxy caching
  - Auto-snapshot is enabled (recommended)
```

### Validation Errors

```
✗ Configuration is invalid

Errors:
  1. project.name is required
  2. wpengine.install is required
  3. network.sites[0].domain is not a subdomain of firecrown.local
  4. ddev.php_version "7.3" is not supported (minimum: 7.4)

Fix these errors before running stax commands.
```

## Configuration Merging

When multiple configuration sources exist, Stax merges them with the following precedence:

1. Command-line flags (highest priority)
2. Environment variables
3. Project configuration (`.stax.yml`)
4. Global configuration (`~/.stax/config.yml`)
5. Built-in defaults (lowest priority)

### Merge Strategy

- **Scalars** (strings, numbers, booleans): Higher priority overwrites
- **Arrays**: Higher priority appends to lower priority
- **Objects**: Recursive merge

### Example Merge

**Global Config** (`~/.stax/config.yml`):
```yaml
ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  xdebug_enabled: false

wpengine:
  backup:
    skip_logs: true
    skip_transients: true
```

**Project Config** (`.stax.yml`):
```yaml
ddev:
  php_version: "8.2"  # Overrides global
  nodejs_version: "20"  # Added to global

wpengine:
  install: fsmultisite  # Added
  backup:
    skip_logs: false  # Overrides global.wpengine.backup.skip_logs
```

**Merged Result**:
```yaml
ddev:
  php_version: "8.2"          # From project
  mysql_version: "8.0"        # From global
  xdebug_enabled: false       # From global
  nodejs_version: "20"        # From project

wpengine:
  install: fsmultisite        # From project
  backup:
    skip_logs: false          # From project (overrides global)
    skip_transients: true     # From global
```

## Configuration Templates

Stax includes templates for common project types.

### WordPress Multisite (Subdomain)

```yaml
version: 1

project:
  name: {{ .ProjectName }}
  type: wordpress-multisite
  mode: subdomain

wpengine:
  install: {{ .WPEngineInstall }}
  environment: production

network:
  domain: {{ .NetworkDomain }}
  sites: []

ddev:
  php_version: "8.1"
  mysql_version: "8.0"

repository:
  url: {{ .RepoURL }}
  branch: main
```

### WordPress Multisite (Subdirectory)

```yaml
version: 1

project:
  name: {{ .ProjectName }}
  type: wordpress-multisite
  mode: subdirectory

wpengine:
  install: {{ .WPEngineInstall }}
  environment: production

network:
  domain: {{ .NetworkDomain }}
  sites: []

ddev:
  php_version: "8.1"
  mysql_version: "8.0"

wordpress:
  constants:
    SUBDOMAIN_INSTALL: false

repository:
  url: {{ .RepoURL }}
  branch: main
```

### Single WordPress Site

```yaml
version: 1

project:
  name: {{ .ProjectName }}
  type: wordpress
  mode: single

wpengine:
  install: {{ .WPEngineInstall }}
  environment: production

ddev:
  php_version: "8.1"
  mysql_version: "8.0"

repository:
  url: {{ .RepoURL }}
  branch: main
```

## Configuration Scenarios

### Scenario 1: Multi-Environment Setup

Support for production, staging, and development configurations:

**Production Config** (`.stax.yml`):
```yaml
version: 1

project:
  name: firecrown-multisite
  type: wordpress-multisite
  mode: subdomain

wpengine:
  install: fsmultisite
  environment: production

network:
  domain: firecrown.local
```

**Staging Config** (`.stax.staging.yml`):
```yaml
version: 1

project:
  name: firecrown-multisite
  type: wordpress-multisite
  mode: subdomain

wpengine:
  install: fsmultisite-staging
  environment: staging

network:
  domain: firecrown-staging.local
```

**Usage**:
```bash
# Production
stax start

# Staging
stax start --config=.stax.staging.yml
# or
STAX_CONFIG=.stax.staging.yml stax start
```

### Scenario 2: Team Sharing

Configuration designed for team sharing via Git:

**`.stax.yml` (committed to Git)**:
```yaml
version: 1

project:
  name: firecrown-multisite
  type: wordpress-multisite
  mode: subdomain

wpengine:
  install: fsmultisite
  environment: production

network:
  domain: firecrown.local
  sites:
    - name: flyingmag
      domain: flyingmag.firecrown.local
      wpengine_domain: flyingmag.com

ddev:
  php_version: "8.1"
  mysql_version: "8.0"

repository:
  url: https://github.com/Firecrown-Media/firecrown-multisite.git
  branch: main

# Note: Credentials NOT stored here - in Keychain
```

**`.stax.local.yml` (NOT committed - in .gitignore)**:
```yaml
version: 1

# Developer-specific overrides
ddev:
  xdebug_enabled: true
  mutagen_enabled: true

logging:
  level: debug
```

**Usage**:
```bash
# Use both configs
stax start --config=.stax.yml --config=.stax.local.yml
```

### Scenario 3: Different PHP Versions per Site

When different sites require different PHP versions:

**Base Config** (`.stax.yml`):
```yaml
version: 1

project:
  name: firecrown-multisite
  type: wordpress-multisite
  mode: subdomain

ddev:
  php_version: "8.1"  # Default for most sites
```

**PHP 8.2 Override** (`.stax.php82.yml`):
```yaml
version: 1

ddev:
  php_version: "8.2"
```

**Usage**:
```bash
# Test with PHP 8.2
stax start --config=.stax.php82.yml
```

### Scenario 4: CI/CD Environment

Configuration for CI/CD pipelines:

**`.stax.ci.yml`**:
```yaml
version: 1

project:
  name: firecrown-multisite-ci
  type: wordpress-multisite
  mode: subdomain

wpengine:
  install: fsmultisite-staging
  environment: staging

ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  xdebug_enabled: false

logging:
  level: info
  format: json

snapshots:
  auto_snapshot_before_pull: false  # Save time in CI

performance:
  parallel_downloads: 8  # CI has more resources
```

**Usage in GitHub Actions**:
```yaml
# .github/workflows/test.yml
- name: Run Stax tests
  run: |
    stax start --config=.stax.ci.yml
    stax wp plugin list
    stax wp core verify-checksums
```

## Configuration Best Practices

### 1. Commit Project Config to Git

✅ **Do commit**:
- `.stax.yml` - Base project configuration
- `.stax.staging.yml` - Staging environment config
- `.stax.ci.yml` - CI/CD configuration

❌ **Don't commit**:
- `.stax.local.yml` - Developer-specific overrides
- Any files with credentials or secrets

**`.gitignore`**:
```
.stax.local.yml
.stax.*.local.yml
```

### 2. Use Environment-Specific Configs

Separate configurations for each environment:

```
.stax.yml              # Production (default)
.stax.staging.yml      # Staging
.stax.development.yml  # Development
.stax.local.yml        # Local overrides (not committed)
```

### 3. Leverage Global Defaults

Set common preferences in global config:

```yaml
# ~/.stax/config.yml
ddev:
  php_version: "8.1"
  xdebug_enabled: false

wpengine:
  backup:
    skip_logs: true
    skip_transients: true

logging:
  level: info
```

### 4. Document Custom Configuration

Add comments to configuration files:

```yaml
# .stax.yml
version: 1

# Project: Firecrown Media Multisite
# Maintained by: Engineering Team
# Last updated: 2025-11-08

project:
  name: firecrown-multisite
  type: wordpress-multisite
  mode: subdomain
  description: |
    Main multisite installation for Firecrown Media brands.
    Includes Flying Magazine, Plane & Pilot, FineScale, and AVWeb.

# WPEngine production environment
wpengine:
  install: fsmultisite
  environment: production

  # Custom backup settings to reduce download time
  backup:
    skip_logs: true
    skip_transients: true
    skip_spam: true
    exclude_tables:
      - wp_actionscheduler_logs  # Large table, not needed locally
```

### 5. Validate Configuration Regularly

Run validation after configuration changes:

```bash
stax config:validate
```

### 6. Use Environment Variables for Secrets

Never store credentials in config files:

```yaml
# ❌ BAD
wpengine:
  api_user: myuser@example.com
  api_password: mypassword123

# ✅ GOOD
wpengine:
  install: fsmultisite
  # Credentials stored in Keychain via `stax setup`
```

### 7. Version Configuration Files

Track configuration changes in Git:

```bash
git add .stax.yml
git commit -m "Update PHP version to 8.2"
```

## Configuration Migration

When updating Stax versions, configuration may need migration.

### Migration Path

**v1.0 → v1.1**:
```yaml
# Old format (v1.0)
wpengine:
  site: fsmultisite
  env: prod

# New format (v1.1)
wpengine:
  install: fsmultisite
  environment: production
```

**Migration Command**:
```bash
stax config:migrate --from=1.0 --to=1.1
```

**Migration Output**:
```
🔄 Migrating configuration from v1.0 to v1.1

Changes:
  - wpengine.site → wpengine.install
  - wpengine.env → wpengine.environment (normalized values)

✓ Configuration migrated successfully
✓ Backup saved to: .stax.yml.backup.20251108
```

## Troubleshooting Configuration

### Common Issues

**Issue 1: Configuration not found**

```
Error: Configuration file not found: .stax.yml
```

**Solution**:
```bash
# Initialize new config
stax init

# Or specify config file
stax start --config=/path/to/.stax.yml
```

**Issue 2: Invalid YAML syntax**

```
Error: Failed to parse configuration: yaml: line 10: mapping values are not allowed in this context
```

**Solution**:
- Check YAML syntax (indentation, colons, quotes)
- Validate with `stax config:validate`
- Use YAML linter

**Issue 3: Conflicting configuration values**

```
Warning: PHP version mismatch
  - Global config: 8.1
  - Project config: 8.2
  - Active: 8.2 (project takes precedence)
```

**Solution**:
- Review configuration priority
- Use `stax config:list` to see merged config
- Override with environment variables if needed

**Issue 4: Credentials not found**

```
Error: WPEngine credentials not found in Keychain
```

**Solution**:
```bash
# Setup credentials
stax setup

# Or set via environment variables
WPENGINE_API_USER=user@example.com \
WPENGINE_API_PASSWORD=password \
stax db:pull
```

### Debugging Configuration

**View merged configuration**:
```bash
stax config:list --verbose
```

**Output**:
```
Configuration Sources:
  - Global: ~/.stax/config.yml ✓
  - Project: .stax.yml ✓
  - Environment: 2 variables set

Merged Configuration:
  project:
    name: firecrown-multisite (source: project)
    type: wordpress-multisite (source: project)

  ddev:
    php_version: "8.2" (source: project, overrides global: "8.1")
    mysql_version: "8.0" (source: global)
```

**Validate configuration**:
```bash
stax config:validate --verbose
```

**Check specific value**:
```bash
stax config:get ddev.php_version --source
```

**Output**:
```
8.2 (source: .stax.yml, line 15)
```

## Summary

Stax configuration provides:

- **Hierarchical configuration**: Global defaults, project-specific, environment variables
- **Secure credential storage**: macOS Keychain integration
- **Team-friendly**: Version-controlled, sharable configuration
- **Flexible**: Environment-specific configs, overrides, templates
- **Validated**: Automatic validation with helpful error messages
- **Documented**: Self-documenting with comments and descriptions

Next steps: Review WPENGINE_INTEGRATION.md for WPEngine-specific configuration details.
