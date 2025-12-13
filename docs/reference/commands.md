# Stax Command Reference

Complete reference for all Stax commands.

---

## Table of Contents

- [Command Structure](#command-structure)
- [Global Flags](#global-flags)
- [Core Commands](#core-commands)
- [Database Commands](#database-commands)
- [WordPress Commands](#wordpress-commands)
- [Configuration Commands](#configuration-commands)
- [Build Commands](#build-commands)
- [Provider Commands](#provider-commands)
- [Diagnostic Commands](#diagnostic-commands)

---

## Command Structure

All Stax commands follow this pattern:

```
stax [command] [subcommand] [arguments] [flags]
```

**Examples**:
```bash
stax start                          # Simple command
stax db pull                        # Command with subcommand
stax db pull --environment=staging  # With flags
stax wp plugin list                 # With arguments
```

---

## Global Flags

Available on all commands:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--help` | `-h` | bool | Show help for command |
| `--version` | | bool | Show Stax version |
| `--verbose` | `-v` | bool | Verbose output |
| `--debug` | `-d` | bool | Debug logging |
| `--quiet` | `-q` | bool | Suppress output |
| `--config` | `-c` | string | Config file path |
| `--no-color` | | bool | Disable colors |

**Examples**:
```bash
stax --version
stax start --verbose
stax db pull --debug
stax --help
```

---

## Core Commands

### stax list

List available WPEngine installs.

**Usage**:
```bash
stax list [flags]
```

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | table | Output format (table, json, yaml) |
| `--provider` | `-p` | string | wpengine | Provider to list from |
| `--filter` | `-f` | string | | Filter by install name (regex) |
| `--environment` | `-e` | string | | Filter by environment |

**Examples**:
```bash
# List all installs
stax list

# List with JSON output
stax list --output=json

# Filter by name (regex)
stax list --filter="client.*"
stax list --filter="fs-.*-prod"

# Filter by environment
stax list --environment=production
stax list --environment=staging

# Combined filters
stax list --filter="fs.*" --environment=staging

# YAML output for scripting
stax list --output=yaml
```

**Output**:
```
INSTALL NAME       ENVIRONMENT   PRIMARY DOMAIN           PHP   STATUS
myinstall          production    mysite.wpengine.com      8.1   active
myinstall-staging  staging       myinstall-staging.wpe    8.1   active
client-prod        production    client.com               8.2   active

Total: 3 installs
```

**What it does**:
- Lists all WPEngine installations accessible to your account
- Works globally without requiring a stax.yml file
- Useful for discovering install names before running `stax init`
- Supports filtering by name (regex) and environment
- Multiple output formats for different use cases

**When to use**:
- Before initializing a new project (to find install names)
- When onboarding new team members
- To audit all available WPEngine installs
- When you forget an install name
- For scripting and automation (JSON/YAML output)

**Notes**:
- Requires WPEngine credentials (run `stax setup` first)
- Uses your API credentials from keychain/env vars
- Does not require SSH key (API only)
- No stax.yml file needed
- Fast operation (typically <5 seconds)

---

### stax init

Initialize a new Stax project.

**Usage**:
```bash
stax init [flags]
```

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--name` | string | (dir name) | Project name |
| `--type` | string | wordpress-multisite | Project type |
| `--mode` | string | subdomain | Multisite mode |
| `--interactive` | bool | true | Interactive mode |

**Examples**:
```bash
# Interactive (recommended)
stax init

# Non-interactive
stax init --name=my-project --mode=subdomain

# Skip database import
stax init --skip-db

# Skip build
stax init --skip-build
```

**What it does**:
1. Creates `.stax.yml` configuration
2. Validates WPEngine credentials
3. Clones GitHub repository
4. Generates DDEV configuration
5. Starts containers
6. Installs dependencies
7. Pulls database
8. Runs search-replace
9. Displays access URLs

---

### stax start

Start the development environment.

**Usage**:
```bash
stax start [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--build` | bool | Run build after start |

**Examples**:
```bash
stax start
stax start --build
```

**Output**:
```
🚀 Starting my-project

✓ Starting DDEV containers
  Web: https://my-project.local
  Database: MySQL 8.0 (ready)

Environment started successfully!
```

---

### stax stop

Stop the development environment.

**Usage**:
```bash
stax stop
```

**Examples**:
```bash
stax stop
```

---

### stax restart

Restart the development environment.

**Usage**:
```bash
stax restart [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--build` | bool | Run build after restart |

**Examples**:
```bash
stax restart
stax restart --build
```

---

### stax status

Show environment status.

**Usage**:
```bash
stax status [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--json` | bool | Output as JSON |

**Examples**:
```bash
stax status
stax status --json
```

**Output**:
```
📊 Status: my-project

Environment: Running ✓

Containers:
  ✓ ddev-my-project-web    (healthy)
  ✓ ddev-my-project-db     (healthy)
  ✓ ddev-router            (healthy)

URLs:
  Network: https://my-project.local
  Site 1:  https://site1.my-project.local

Configuration:
  PHP:     8.1
  MySQL:   8.0

Database:
  Size:    245 MB
  Tables:  127
```

---

### stax ssh

SSH into the web container.

**Usage**:
```bash
stax ssh [command]
```

**Examples**:
```bash
# Interactive session
stax ssh

# Run single command
stax ssh "wp plugin list"
stax ssh "composer install"
```

---

### stax setup

Configure credentials.

**Usage**:
```bash
stax setup
```

**Interactive prompts for**:
- WPEngine API username
- WPEngine API password
- GitHub token (optional)
- SSH key path

**Examples**:
```bash
stax setup
```

---

## Database Commands

### stax db pull

Pull database from WPEngine.

**Usage**:
```bash
stax db pull [flags]
```

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--environment` | string | (from config) | WPEngine environment |
| `--snapshot` | bool | true | Create snapshot before import |
| `--sanitize` | bool | false | Sanitize user data |
| `--skip-logs` | bool | true | Skip log tables |
| `--skip-transients` | bool | true | Skip transient tables |
| `--skip-spam` | bool | true | Skip spam/trash |
| `--exclude-tables` | string | | Tables to exclude (comma-separated) |

**Examples**:
```bash
# Basic pull
stax db pull

# From staging
stax db pull --environment=staging

# Without snapshot
stax db pull --snapshot=false

# Sanitize data
stax db pull --sanitize

# Skip specific tables
stax db pull --exclude-tables=wp_actionscheduler_logs,wp_wc_admin_notes
```

---

### stax db snapshot

Create database snapshot.

**Usage**:
```bash
stax db snapshot [name] [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--description` | string | Snapshot description |

**Examples**:
```bash
# Auto-named
stax db snapshot

# Named
stax db snapshot before-migration

# With description
stax db snapshot pre-deploy --description="Before deployment"
```

---

### stax db restore

Restore database snapshot.

**Usage**:
```bash
stax db restore <name> [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--force` | bool | Skip confirmation |

**Examples**:
```bash
stax db restore before-migration
stax db restore latest
stax db restore before-migration --force
```

---

### stax db list

List all snapshots.

**Usage**:
```bash
stax db list [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--json` | bool | Output as JSON |

**Examples**:
```bash
stax db list
stax db list --json
```

---

### stax db export

Export database to SQL file.

**Usage**:
```bash
stax db export [file] [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--gzip` | bool | Compress with gzip |
| `--skip-logs` | bool | Skip log tables |
| `--exclude-tables` | string | Tables to exclude |

**Examples**:
```bash
# Export to default location
stax db export

# Export to specific file
stax db export ~/backups/my-backup.sql

# Compressed
stax db export --gzip
```

---

### stax db import

Import SQL file.

**Usage**:
```bash
stax db import <file> [flags]
```

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--snapshot` | bool | true | Create snapshot before import |
| `--replace` | bool | false | Run search-replace after import |

**Examples**:
```bash
stax db import backup.sql
stax db import backup.sql --snapshot=false
stax db import backup.sql --replace
```

---

## WordPress Commands

### stax wp

Execute WP-CLI commands.

**Usage**:
```bash
stax wp <command> [args...] [flags]
```

**Examples**:
```bash
stax wp plugin list
stax wp plugin activate wordpress-seo
stax wp user list
stax wp site list
stax wp cache flush
stax wp search-replace old.com new.com --dry-run
stax wp db query "SELECT * FROM wp_options LIMIT 5"
```

**All WP-CLI commands work**:
- `stax wp core ...`
- `stax wp plugin ...`
- `stax wp theme ...`
- `stax wp user ...`
- `stax wp site ...`
- `stax wp db ...`
- `stax wp cache ...`
- And many more!

**Multisite-specific**:
```bash
# Network-wide
stax wp plugin activate plugin-name --network
stax wp cache flush --network

# Site-specific
stax wp plugin list --url=site1.example.local
stax wp cache flush --url=site1.example.local
```

---

## Configuration Commands

### stax config list

List all configuration.

**Usage**:
```bash
stax config list [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--global` | bool | List global config |
| `--json` | bool | Output as JSON |

**Examples**:
```bash
stax config list
stax config list --global
stax config list --json
```

---

### stax config get

Get configuration value.

**Usage**:
```bash
stax config get <key> [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--global` | bool | Get from global config |

**Examples**:
```bash
stax config get wpengine.environment
stax config get ddev.php_version
stax config get project.mode
```

---

### stax config set

Set configuration value.

**Usage**:
```bash
stax config set <key> <value> [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--global` | bool | Set in global config |

**Examples**:
```bash
stax config set wpengine.environment staging
stax config set ddev.php_version 8.2
stax config set project.mode subdirectory

# Global setting
stax config set defaults.wpengine.environment staging --global
```

---

### stax config validate

Validate configuration.

**Usage**:
```bash
stax config validate
```

**Examples**:
```bash
stax config validate
```

---

## Build Commands

### stax build

Run build process.

**Usage**:
```bash
stax build [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--production` | bool | Production build |

**Examples**:
```bash
stax build
stax build --production
```

**What it runs**:
1. `composer install`
2. `npm install`
3. Build scripts from `.stax.yml`

---

### stax dev

Start development mode (watch).

**Usage**:
```bash
stax dev
```

**Examples**:
```bash
stax dev
```

Runs your build tools in watch mode. Changes rebuild automatically.

---

### stax lint

Run code linters.

**Usage**:
```bash
stax lint [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--fix` | bool | Auto-fix when possible |

**Examples**:
```bash
stax lint
stax lint --fix
```

**Runs**:
- PHP_CodeSniffer
- ESLint
- Stylelint
- PHPStan (if configured)

---

## Provider Commands

### stax provider list

List available providers.

**Usage**:
```bash
stax provider list
```

**Examples**:
```bash
stax provider list
```

---

### stax provider info

Show provider information.

**Usage**:
```bash
stax provider info <provider>
```

**Examples**:
```bash
stax provider info wpengine
```

---

## Diagnostic Commands

### stax doctor

Run diagnostics.

**Usage**:
```bash
stax doctor [flags]
```

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--fix` | bool | Auto-fix issues |

**Examples**:
```bash
stax doctor
stax doctor --fix
```

**Checks**:
- Stax installation
- DDEV installation
- Docker status
- Credentials
- Port availability
- SSL certificates
- Configuration validity

---

### stax logs

View container logs.

**Usage**:
```bash
stax logs [flags]
```

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--follow` / `-f` | bool | false | Follow logs |
| `--tail` | int | 100 | Number of lines |
| `--service` | string | all | Service to show |
| `--timestamp` | bool | false | Show timestamps |

**Examples**:
```bash
# Last 100 lines
stax logs

# Follow logs
stax logs -f

# Last 500 lines
stax logs --tail=500

# Web container only
stax logs --service=web

# Database logs
stax logs --service=db

# With timestamps
stax logs -f --timestamp
```

---

## Quick Reference

### Most Common Commands

```bash
# Project lifecycle
stax list           # List WPEngine installs
stax init           # Initialize project
stax start          # Start environment
stax stop           # Stop environment
stax restart        # Restart environment
stax status         # Show status

# Database
stax db pull        # Pull from WPEngine
stax db snapshot    # Create snapshot
stax db restore     # Restore snapshot
stax db list        # List snapshots

# WordPress
stax wp ...         # Any WP-CLI command
stax ssh            # SSH into container

# Development
stax build          # Run build
stax dev            # Watch mode
stax lint           # Run linters
stax logs -f        # View logs

# Configuration
stax config list    # Show config
stax setup          # Configure credentials
stax doctor         # Run diagnostics
```

---

## Environment Variables

Override config via environment variables:

| Variable | Description |
|----------|-------------|
| `STAX_CONFIG` | Config file path |
| `STAX_DEBUG` | Enable debug mode |
| `STAX_NO_COLOR` | Disable colored output |
| `STAX_WPENGINE_ENV` | WPEngine environment |

**Examples**:
```bash
STAX_DEBUG=true stax init
STAX_WPENGINE_ENV=staging stax db pull
STAX_CONFIG=/path/to/config.yml stax start
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Missing dependencies |
| 4 | Authentication failed |
| 5 | Network error |
| 10 | User cancelled |

---

## More Information

- **User Guide**: [USER_GUIDE.md](./USER_GUIDE.md)
- **Examples**: [EXAMPLES.md](./EXAMPLES.md)
- **Troubleshooting**: [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
- **FAQ**: [FAQ.md](./FAQ.md)
