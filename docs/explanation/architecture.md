# Stax Architecture

## Executive Summary

Stax is a powerful CLI tool designed to streamline WordPress multisite development workflows. Built with Go and leveraging DDEV for container orchestration, Stax automates the complete setup of WordPress multisite environments with hosting provider integration, remote media proxying, and team-friendly configuration management.

## Technology Stack Decision

### Core Technologies

- **Language**: Go 1.22+
- **CLI Framework**: Cobra
- **Container Platform**: DDEV
- **Version Control**: Git
- **Package Distribution**: Homebrew (primary), GitHub Releases
- **Configuration Format**: YAML
- **Credential Storage**: macOS Keychain (via go-keychain)

### Container Platform: DDEV (Recommended)

After comprehensive evaluation of DDEV, Podman, and Docker Compose, **DDEV is the recommended platform** for the following reasons:

#### Why DDEV Over Alternatives

**1. Abstracted Complexity**
- DDEV provides pre-configured, CMS-specific environments that reduce setup time from hours to minutes
- Single `config.yaml` file consolidates all settings vs. complex Docker Compose configurations
- Built-in WordPress multisite support with wildcard subdomain SSL handling
- Team collaboration is simplified through consistent, committable configuration

**2. Mac Optimization (Apple Silicon + Intel)**
- Native support for both Apple Silicon (arm64) and Intel (amd64) architectures
- Built on Docker Desktop for Mac, which uses Apple's Hypervisor.framework for optimal performance
- Includes automatic SSL certificate generation and management via mkcert
- No manual /etc/hosts editing required (uses ddev-router for DNS resolution)

**3. PHP/MySQL Version Management**
- Simple PHP version switching via `php_version` config parameter
- Automatic container selection for correct PHP runtime
- Multiple database engines supported (MySQL 5.7, 8.0, MariaDB 10.x)
- Per-project version configuration without global system changes

**4. Built-in Development Tools**
- Xdebug pre-installed (toggle on/off for performance)
- MailHog email capture enabled by default
- WP-CLI available in container
- Composer and npm/node pre-installed
- Database management tools (phpMyAdmin addon)

**5. WordPress Multisite Support**
- Native support for both subdomain and subdirectory multisite modes
- Wildcard SSL certificates for subdomain installations (*.site.ddev.site)
- Additional hostnames easily configured for brand-specific domains
- Built-in router handles all subdomain DNS resolution

**6. Junior Developer Friendliness**
- Intuitive commands: `ddev start`, `ddev ssh`, `ddev composer`
- Automatic service health checks and error reporting
- Clear, actionable error messages
- Extensive documentation and community support
- No deep Docker knowledge required

**7. Performance**
- Mutagen file sync for improved Mac performance (optional)
- NFS mounting available for faster I/O operations
- Container resource limiting and optimization
- Faster than manual Docker Compose setups due to optimized images

**Why Not Podman**
- Immature Mac support (Podman 5 improved but still evolving)
- Requires Podman Desktop VM on Mac (similar overhead to Docker Desktop)
- Rootless networking performance penalty (2-4 Gbps vs 8-10 Gbps)
- Less ecosystem maturity for WordPress-specific tooling
- Team would need to learn new tooling vs. Docker-compatible DDEV

**Why Not Docker Compose**
- Requires manual configuration of all services (nginx/Apache, PHP, MySQL, SSL, MailHog)
- Team must maintain custom configurations vs. DDEV's maintained images
- No built-in multisite subdomain DNS resolution
- Manual SSL certificate generation and renewal
- Higher maintenance burden and setup complexity
- Longer onboarding time for junior developers

## System Architecture

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Stax CLI (Go)                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Command    │  │   Config     │  │   Credential         │  │
│  │   Router     │  │   Manager    │  │   Manager            │  │
│  │   (Cobra)    │  │   (YAML)     │  │   (Keychain)         │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘  │
│         │                 │                      │              │
│  ┌──────┴─────────────────┴──────────────────────┴───────────┐  │
│  │              Core Orchestration Layer                     │  │
│  └──────┬─────────────┬──────────────┬──────────────┬────────┘  │
│         │             │              │              │           │
├─────────┼─────────────┼──────────────┼──────────────┼───────────┤
│  ┌──────▼──────┐ ┌────▼──────┐ ┌────▼──────┐ ┌─────▼────────┐  │
│  │   DDEV      │ │ WPEngine  │ │  GitHub   │ │  WordPress   │  │
│  │   Manager   │ │  Client   │ │  Client   │ │   Manager    │  │
│  └──────┬──────┘ └────┬──────┘ └────┬──────┘ └──────┬───────┘  │
│         │             │              │               │          │
└─────────┼─────────────┼──────────────┼───────────────┼──────────┘
          │             │              │               │
┌─────────▼─────────────▼──────────────▼───────────────▼──────────┐
│                     External Systems                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │     DDEV     │  │   WPEngine   │  │       GitHub         │   │
│  │  Containers  │  │  SSH Gateway │  │    Repositories      │   │
│  │              │  │      API     │  │   (firecrown-*)      │   │
│  │  - Web/PHP   │  │              │  │                      │   │
│  │  - Database  │  │  - Database  │  │  - Actions/Hooks     │   │
│  │  - Router    │  │  - Files     │  │                      │   │
│  │  - MailHog   │  │  - Rsync     │  │                      │   │
│  └──────────────┘  └──────────────┘  └──────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

### Component Overview

#### 1. Command Router (Cobra)
**Purpose**: CLI command parsing, validation, and routing

**Responsibilities**:
- Parse command-line arguments and flags
- Validate required parameters
- Route to appropriate handler functions
- Display help text and usage information
- Handle command aliases and shortcuts

**Key Dependencies**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

#### 2. Config Manager
**Purpose**: Configuration file management and validation

**Responsibilities**:
- Read/write `.stax.yml` configuration files
- Merge global and project-specific configurations
- Validate configuration schemas
- Provide default values for missing configs
- Support environment variable overrides

**Key Files**:
- `~/.stax/config.yml` - Global configuration
- `<project>/.stax.yml` - Project-specific configuration
- `<project>/.ddev/config.yaml` - DDEV configuration (generated)

**Configuration Schema**:
```yaml
version: 1
project:
  name: firecrown-multisite
  type: wordpress-multisite
  mode: subdomain  # or subdirectory

wpengine:
  environment: production  # or staging
  install: fsmultisite

network:
  domain: firecrown.local
  sites:
    - name: flyingmag
      domain: flyingmag.firecrown.local
      wpengine_domain: flyingmag.com
    - name: planeandpilot
      domain: planeandpilot.firecrown.local
      wpengine_domain: planeandpilotmag.com

ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  webserver_type: nginx
  router_http_port: "80"
  router_https_port: "443"

repository:
  url: https://github.com/Firecrown-Media/firecrown-multisite.git
  branch: main

build:
  pre_install:
    - composer install
    - npm install
  post_install:
    - scripts/build.sh
```

#### 3. Credential Manager
**Purpose**: Secure storage and retrieval of API keys and credentials

**Responsibilities**:
- Store WPEngine API credentials in macOS Keychain
- Store GitHub tokens for private repository access
- Store SSH keys for WPEngine SSH Gateway
- Provide secure credential retrieval interface
- Support credential rotation and updates

**Key Dependencies**:
- `github.com/keybase/go-keychain` - macOS Keychain access

**Keychain Items**:
- `stax.wpengine.api_user` - WPEngine API username
- `stax.wpengine.api_password` - WPEngine API password
- `stax.github.token` - GitHub personal access token
- `stax.wpengine.ssh_key` - SSH private key for WPEngine

#### 4. DDEV Manager
**Purpose**: Interface with DDEV for container orchestration

**Responsibilities**:
- Generate DDEV configuration files
- Execute DDEV commands (start, stop, restart, delete)
- Monitor container health and status
- Configure additional services (MailHog, phpMyAdmin)
- Manage custom DDEV commands and hooks
- Handle DDEV version compatibility

**DDEV Configuration Generation**:
```yaml
# .ddev/config.yaml (generated by stax)
name: firecrown-multisite
type: wordpress
docroot: ""
php_version: "8.1"
webserver_type: nginx-fpm
router_http_port: "80"
router_https_port: "443"
xdebug_enabled: false
additional_hostnames:
  - "*.firecrown"
  - flyingmag.firecrown
  - planeandpilot.firecrown
  - finescale.firecrown
  - avweb.firecrown
additional_fqdns:
  - "*.firecrown.local"
  - flyingmag.firecrown.local
  - planeandpilot.firecrown.local
  - finescale.firecrown.local
  - avweb.firecrown.local
database:
  type: mysql
  version: "8.0"
hooks:
  post-start:
    - exec: wp search-replace fsmultisite.wpenginepowered.com firecrown.local --network
    - exec: wp search-replace flyingmag.com flyingmag.firecrown.local --url=flyingmag.com
    - exec: wp search-replace planeandpilotmag.com planeandpilot.firecrown.local --url=planeandpilotmag.com
```

**Custom DDEV Commands**:
- `.ddev/commands/web/stax-sync` - Sync database from WPEngine
- `.ddev/commands/web/stax-build` - Run build process
- `.ddev/commands/web/stax-wp` - Wrapper for WP-CLI commands

#### 5. WPEngine Client
**Purpose**: Integration with WPEngine services

**Responsibilities**:
- Authenticate with WPEngine API
- Connect to WPEngine SSH Gateway
- Download database backups (full or partial)
- Sync files via rsync
- Query environment information (PHP version, MySQL version)
- Handle WPEngine-specific domain configurations

**API Endpoints** (see WPENGINE_INTEGRATION.md for details):
- `/installs` - List available installations
- `/installs/{install_id}` - Get installation details
- `/installs/{install_id}/backups` - List database backups
- SSH Gateway - File and database access via SSH

**Key Dependencies**:
- `golang.org/x/crypto/ssh` - SSH client
- `net/http` - HTTP client for API calls

#### 6. GitHub Client
**Purpose**: Integration with GitHub repositories and workflows

**Responsibilities**:
- Clone private firecrown-* repositories
- Authenticate with GitHub API using tokens
- Trigger GitHub Actions workflows
- Monitor deployment status
- Manage repository webhooks for auto-sync

**Key Dependencies**:
- `github.com/google/go-github/v57/github` - GitHub API client
- `golang.org/x/oauth2` - OAuth2 authentication

#### 7. WordPress Manager
**Purpose**: WordPress-specific operations and configurations

**Responsibilities**:
- Execute WP-CLI commands in DDEV container
- Perform multisite search-replace operations
- Configure WordPress constants (wp-config.php)
- Manage plugin activation/deactivation
- Handle theme builds and asset compilation
- Detect table prefix from database

**WP-CLI Operations**:
- `wp core multisite-install` - Initial multisite setup
- `wp site create` - Create new subsites
- `wp search-replace` - Domain replacements
- `wp db import` - Database import
- `wp db query` - Direct SQL queries
- `wp plugin activate/deactivate` - Plugin management

## Data Flow Diagrams

### Initial Project Setup Flow

```
┌─────────────┐
│ stax init   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────┐
│ Create .stax.yml config     │
│ (interactive prompts)       │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ Validate WPEngine access    │
│ (API + SSH Gateway)         │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ Clone GitHub repository     │
│ (firecrown-multisite)       │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ Query WPEngine for versions │
│ (PHP, MySQL, WordPress)     │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ Generate DDEV config        │
│ (.ddev/config.yaml)         │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ DDEV start                  │
│ (pull images, start)        │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ Run build process           │
│ (composer, npm, build.sh)   │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ Download WPEngine database  │
│ (via SSH Gateway)           │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ Import database             │
│ (wp db import)              │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ Run search-replace          │
│ (network + all sites)       │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│ Display success + URLs      │
└─────────────────────────────┘
```

### Database Sync Flow

```
┌──────────────────┐
│ stax db:pull     │
└────────┬─────────┘
         │
         ▼
┌──────────────────────────────┐
│ Connect to WPEngine SSH      │
│ Gateway                      │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Detect table prefix          │
│ (query wp_options)           │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Export database              │
│ (partial or full)            │
│ Options:                     │
│  - Skip logs/transients      │
│  - Skip spam/trash           │
│  - Include specific tables   │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Transfer via SSH             │
│ (stream to local tmp)        │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Import to DDEV database      │
│ (via ddev mysql)             │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Run search-replace           │
│ (WPEngine → local domains)   │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Flush WordPress cache        │
│ (wp cache flush)             │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Display summary              │
│ (rows imported, sites)       │
└──────────────────────────────┘
```

### Remote Media Proxying Flow

```
┌──────────────────────────────┐
│ WordPress request for        │
│ /wp-content/uploads/...      │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Nginx checks local file      │
│ (try_files directive)        │
└────────┬─────────────────────┘
         │
         │ (not found)
         ▼
┌──────────────────────────────┐
│ Check BunnyCDN first         │
│ (via nginx proxy_pass)       │
│ https://cdn.example.com/...  │
└────────┬─────────────────────┘
         │
         │ (404)
         ▼
┌──────────────────────────────┐
│ Fallback to WPEngine         │
│ (via nginx proxy_pass)       │
│ https://wpengine.com/...     │
└────────┬─────────────────────┘
         │
         │ (200 OK)
         ▼
┌──────────────────────────────┐
│ Optional: Cache locally      │
│ (proxy_cache directive)      │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Serve to browser             │
└──────────────────────────────┘
```

## File Structure and Organization

```
stax/
├── cmd/                          # CLI command definitions
│   ├── root.go                   # Root command and global flags
│   ├── init.go                   # stax init
│   ├── start.go                  # stax start
│   ├── stop.go                   # stax stop
│   ├── restart.go                # stax restart
│   ├── delete.go                 # stax delete
│   ├── status.go                 # stax status
│   ├── ssh.go                    # stax ssh
│   ├── logs.go                   # stax logs
│   ├── database.go               # Database command group
│   │   ├── pull.go               # stax db:pull
│   │   ├── export.go             # stax db:export
│   │   ├── snapshot.go           # stax db:snapshot
│   │   └── restore.go            # stax db:restore
│   ├── wordpress.go              # WordPress command group
│   │   ├── cli.go                # stax wp <command>
│   │   ├── search_replace.go     # stax wp:search-replace
│   │   └── plugin.go             # stax wp:plugin
│   ├── wpengine.go               # WPEngine command group
│   │   ├── sync.go               # stax wpe:sync
│   │   ├── info.go               # stax wpe:info
│   │   └── backups.go            # stax wpe:backups
│   ├── config.go                 # Configuration commands
│   │   ├── set.go                # stax config:set
│   │   ├── get.go                # stax config:get
│   │   └── list.go               # stax config:list
│   └── setup.go                  # stax setup (credential setup)
│
├── pkg/                          # Shared packages
│   ├── config/                   # Configuration management
│   │   ├── config.go             # Config struct and methods
│   │   ├── loader.go             # Load/merge configurations
│   │   ├── validator.go          # Config validation
│   │   └── defaults.go           # Default values
│   │
│   ├── credentials/              # Credential management
│   │   ├── keychain.go           # macOS Keychain interface
│   │   ├── manager.go            # Credential CRUD operations
│   │   └── types.go              # Credential types
│   │
│   ├── ddev/                     # DDEV management
│   │   ├── manager.go            # DDEV operations
│   │   ├── config.go             # Config generation
│   │   ├── commands.go           # Custom DDEV commands
│   │   ├── hooks.go              # DDEV hooks management
│   │   └── nginx.go              # Nginx config for media proxy
│   │
│   ├── wpengine/                 # WPEngine integration
│   │   ├── client.go             # API client
│   │   ├── ssh.go                # SSH Gateway client
│   │   ├── database.go           # Database operations
│   │   ├── files.go              # File sync (rsync)
│   │   └── types.go              # API response types
│   │
│   ├── github/                   # GitHub integration
│   │   ├── client.go             # GitHub API client
│   │   ├── clone.go              # Repository cloning
│   │   └── workflows.go          # Actions/workflows
│   │
│   ├── wordpress/                # WordPress operations
│   │   ├── wpcli.go              # WP-CLI wrapper
│   │   ├── multisite.go          # Multisite operations
│   │   ├── search_replace.go     # Search-replace logic
│   │   ├── build.go              # Build process
│   │   └── detect.go             # Detect WP config
│   │
│   ├── ui/                       # User interface utilities
│   │   ├── spinner.go            # Loading spinners
│   │   ├── progress.go           # Progress bars
│   │   ├── prompts.go            # Interactive prompts
│   │   └── colors.go             # Terminal colors
│   │
│   └── errors/                   # Error handling
│       ├── errors.go             # Custom error types
│       └── handlers.go           # Error handlers
│
├── templates/                    # Template files
│   ├── ddev/                     # DDEV templates
│   │   ├── config.yaml.tmpl      # DDEV config template
│   │   ├── nginx-site.conf.tmpl  # Nginx config template
│   │   └── commands/             # Custom command templates
│   │       ├── stax-sync         # Database sync command
│   │       └── stax-build        # Build command
│   │
│   └── stax/                     # Stax templates
│       └── .stax.yml.tmpl        # Project config template
│
├── docs/                         # Documentation
│   ├── ARCHITECTURE.md           # This file
│   ├── COMMANDS.md               # Command reference
│   ├── CONFIG_SPEC.md            # Configuration spec
│   ├── WPENGINE_INTEGRATION.md   # WPEngine integration
│   ├── DEVELOPMENT.md            # Development guide
│   └── TROUBLESHOOTING.md        # Troubleshooting guide
│
├── .github/                      # GitHub workflows
│   └── workflows/
│       ├── test.yml              # Run tests on PR
│       ├── release.yml           # Build and release
│       └── update-homebrew-tap.yml # Update Homebrew tap
│
├── scripts/                      # Build and utility scripts
│   ├── build.sh                  # Build binary
│   ├── install.sh                # Local installation
│   └── test.sh                   # Run tests
│
├── main.go                       # Application entry point
├── go.mod                        # Go module definition
├── go.sum                        # Go dependencies
├── Makefile                      # Build automation
└── README.md                     # User-facing documentation
```

## Development Workflow Integration

### Local Development Lifecycle

1. **Developer runs**: `stax init`
   - Interactive prompts for project configuration
   - Credentials validated against WPEngine
   - Repository cloned from GitHub
   - DDEV environment created and started
   - Build process executed
   - Database pulled and imported
   - Site ready at `https://firecrown.local`

2. **Daily Development**:
   ```bash
   stax start              # Start environment
   stax ssh                # SSH into container
   stax wp plugin list     # WP-CLI commands
   stax db:pull            # Refresh database
   stax logs -f            # Tail logs
   stax stop               # Stop environment
   ```

3. **Database Refresh**:
   ```bash
   stax db:snapshot        # Save current state
   stax db:pull            # Pull latest from WPEngine
   stax db:restore         # Restore snapshot if needed
   ```

### GitHub Workflow Integration

Stax integrates with existing GitHub workflows for WPEngine deployments:

```yaml
# .github/workflows/deploy-wpengine.yml
name: Deploy to WPEngine
on:
  push:
    branches: [main, staging]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to WPEngine
        uses: wpengine/github-action-wpe-site-deploy@v3
        with:
          WPE_SSHG_KEY_PRIVATE: ${{ secrets.WPE_SSHG_KEY_PRIVATE }}
          WPE_ENV: ${{ github.ref == 'refs/heads/main' && 'production' || 'staging' }}
```

Stax can trigger and monitor these workflows:
```bash
stax wpe:deploy --environment=staging
stax wpe:deploy --environment=production --watch
```

## Build Process Preservation

Stax preserves Firecrown's existing build toolchain:

### Build Steps (Automated by Stax)

1. **Composer Install**
   ```bash
   composer install --no-dev --optimize-autoloader
   ```

2. **NPM Install**
   ```bash
   npm install
   ```

3. **Build Script Execution**
   ```bash
   scripts/build.sh
   ```
   - Builds MU-plugins
   - Compiles themes (Webpack/Gulp)
   - Generates optimized assets

4. **Code Quality Checks** (Pre-commit hooks preserved)
   - PHPCS (WordPress coding standards)
   - Husky pre-commit hooks
   - ESLint for JavaScript

### DDEV Integration

Build process runs automatically via DDEV hooks:

```yaml
# .ddev/config.yaml
hooks:
  post-start:
    - exec: composer install
    - exec: npm install
    - exec: bash scripts/build.sh
  pre-commit:
    - exec: composer run phpcs
```

## Security Considerations

### Credential Storage

- **macOS Keychain**: All sensitive credentials stored in Keychain
- **No Plain Text**: Never store credentials in config files
- **Access Control**: Keychain items require user authentication
- **Rotation Support**: Easy credential updates via `stax setup`

### SSH Key Management

- **WPEngine SSH Keys**: Stored in Keychain, loaded for SSH connections
- **GitHub Tokens**: Stored in Keychain, used for API auth
- **Auto-loading**: SSH agent integration for seamless auth

### Database Security

- **No Production Writes**: Stax is read-only against WPEngine
- **Local Snapshots**: Database snapshots stored locally only
- **Sanitization**: Option to sanitize user data on import
  ```bash
  stax db:pull --sanitize
  ```

## Performance Optimization

### Mac-Specific Optimizations

1. **Mutagen File Sync** (Optional)
   - Faster file synchronization than Docker volumes
   - Reduces I/O overhead on Mac
   - Configurable via DDEV settings

2. **NFS Mounting** (Alternative)
   - Better performance than VirtioFS
   - Trade-off: Slightly more complex setup
   - DDEV handles configuration automatically

3. **Resource Limits**
   - Configure Docker Desktop memory/CPU limits
   - Optimize for Apple Silicon efficiency
   - DDEV manages container resources

### Database Performance

1. **Partial Database Imports**
   - Skip unnecessary tables (logs, transients, spam)
   - Significantly faster import times
   - Configurable via flags

2. **Incremental Sync**
   - Track database changes
   - Only pull modified data
   - Future enhancement

### Build Performance

1. **Cached Dependencies**
   - Composer cache persisted in DDEV
   - NPM cache preserved
   - Faster subsequent builds

2. **Parallel Processing**
   - NPM/Composer run in parallel where possible
   - Asset compilation optimized

## Error Handling and Recovery

### Graceful Failures

- **Connection Errors**: Clear messages for WPEngine/GitHub connectivity issues
- **DDEV Errors**: Parse DDEV error output and provide actionable suggestions
- **Build Failures**: Capture and display build script errors with context

### Recovery Mechanisms

1. **Database Snapshots**: Automatic snapshots before risky operations
2. **Rollback Support**: `stax db:restore` to revert changes
3. **Clean Restart**: `stax delete && stax init` for nuclear option
4. **Healthchecks**: `stax status` shows detailed health info

### Logging

- **Structured Logging**: JSON logs for debugging
- **Log Levels**: DEBUG, INFO, WARN, ERROR
- **Log Location**: `~/.stax/logs/stax.log`
- **DDEV Logs**: `stax logs` proxies to DDEV container logs

## Extensibility and Future Enhancements

### Plugin System (Future)

Allow custom commands and hooks:
```go
// ~/.stax/plugins/custom-sync.go
package main

func Execute(args []string) error {
    // Custom sync logic
}
```

### Multi-Environment Support (Future)

Support for staging/production configurations:
```yaml
environments:
  production:
    wpengine_install: fsmultisite-prod
  staging:
    wpengine_install: fsmultisite-stage
```

### Team Sharing (Future)

Shared team configurations via GitHub:
```bash
stax init --from=https://github.com/Firecrown-Media/stax-configs/firecrown-multisite.yml
```

### CI/CD Integration (Future)

Run Stax in CI environments:
```bash
stax test --ci
stax deploy --environment=staging
```

## Version Compatibility Matrix

| Stax Version | DDEV Version | Go Version | macOS Version     |
|--------------|--------------|------------|-------------------|
| 1.x          | 1.22+        | 1.22+      | 12+ (Monterey+)   |
| 2.x (future) | 1.23+        | 1.23+      | 13+ (Ventura+)    |

### Dependency Versions

- **DDEV**: Minimum 1.22.0 (for latest multisite features)
- **Docker Desktop**: 4.25+ (for Apple Silicon optimization)
- **Go**: 1.22+ (for generic types and performance)
- **WP-CLI**: 2.10+ (bundled in DDEV)

## Migration from LocalWP

### Migration Path

For teams currently using LocalWP:

1. **Export LocalWP Site**:
   ```bash
   # In LocalWP:
   # - Export database to SQL file
   # - Note wp-content location
   ```

2. **Import to Stax**:
   ```bash
   stax init
   stax db:import ~/path/to/localwp-export.sql
   # wp-content already in firecrown-multisite repo
   ```

3. **Verify Configuration**:
   ```bash
   stax status
   stax wp site list
   ```

### Advantages Over LocalWP

| Feature                    | LocalWP          | Stax             |
|----------------------------|------------------|------------------|
| CLI automation             | Limited          | Full CLI         |
| Version control config     | Manual           | Automatic        |
| WPEngine integration       | Manual           | Built-in         |
| Multisite subdomain setup  | Manual /etc/hosts| Automatic        |
| Database search-replace    | Manual           | Automatic        |
| Team consistency           | Variable         | Identical        |
| Remote media               | Download all     | Proxy on-demand  |
| Build automation           | Manual           | Automatic        |
| Apple Silicon support      | Yes              | Yes (via DDEV)   |

## Summary

Stax provides a comprehensive, automated replacement for LocalWP designed for professional WordPress multisite development workflows. By leveraging DDEV's mature container platform, Go's performance and cross-platform capabilities, and deep hosting provider integration, Stax reduces setup time from hours to minutes while ensuring consistency across development teams.

The architecture prioritizes:
- **Developer Experience**: Simple, intuitive commands
- **Team Consistency**: Shared, version-controlled configuration
- **Mac Optimization**: Native support for Apple Silicon and Intel
- **Security**: Keychain-based credential management
- **Flexibility**: Support for multiple multisite modes and PHP/MySQL versions
- **Integration**: Seamless WPEngine and GitHub workflows

Next steps: Review COMMANDS.md, CONFIG_SPEC.md, and WPENGINE_INTEGRATION.md for detailed implementation specifications.
