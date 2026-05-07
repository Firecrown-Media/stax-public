# Stax

**A CLI tool for WordPress development with WPEngine integration.**

Stax automates the repetitive parts of WordPress development — setting up local environments, syncing databases, pulling files, running builds — so you can stay focused on building sites.

---

## What is Stax?

Stax is a command-line tool that connects your local [DDEV](https://ddev.readthedocs.io/) environment to your WPEngine hosting. With one command you can:

- Spin up a fully configured WordPress environment
- Pull a database from WPEngine and have URLs automatically replaced
- Sync wp-content files from production or staging
- Create database snapshots before risky operations
- Run builds, linting, and dev servers

Stax works with single sites and WordPress multisite networks, and stores credentials securely in the macOS Keychain.

---

## Prerequisites

- **macOS 12+** (Monterey or later)
- **Docker** — [Docker Desktop](https://www.docker.com/products/docker-desktop) or Colima
- **DDEV** — [Install guide](https://ddev.readthedocs.io/en/stable/users/install/)
- **WPEngine account** with API access enabled

---

## Installation

```bash
brew install firecrown-media/stax/stax
```

Verify:

```bash
stax --version
```

### Build from source

```bash
git clone https://github.com/firecrown-media/stax-public.git
cd stax-public
go mod download
make build
make install
```

---

## Quick Start

```bash
# 1. Store your WPEngine credentials
stax setup

# 2. Initialize a project
mkdir my-site && cd my-site
stax init

# 3. Start the environment
stax start

# 4. Pull the database
stax db pull

# Your site is now running at https://my-site.ddev.site
```

---

## Configuration

Stax projects are configured with a `.stax.yml` file in your project directory.

```yaml
version: 2

project:
  name: my-site
  type: wordpress          # wordpress | wordpress-multisite
  mode: single             # single | subdomain | subdirectory

provider: wpengine
provider_config:
  install: my-install
  environment: production  # production | staging | development
  ssh_gateway: ssh.wpengine.net

ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  webserver_type: nginx-fpm
  nodejs_version: "20"

wordpress:
  version: latest
  locale: en_US
  table_prefix: wp_
```

Generate a template:

```bash
stax init --template > .stax.yml
```

---

## Commands

### Setup & Initialization

#### `stax setup`
Configure WPEngine and GitHub credentials. Stores them securely in the macOS Keychain.

```bash
stax setup

# Non-interactive
stax setup \
  --wpengine-user=user@example.com \
  --wpengine-password=mypassword \
  --ssh-key-path=~/.ssh/id_rsa
```

#### `stax init`
Initialize a new Stax project in the current directory. Runs interactively by default.

```bash
# Interactive (prompts for everything)
stax init

# Non-interactive with all flags
stax init \
  --name=my-site \
  --type=wordpress \
  --install=my-install \
  --environment=production \
  --php=8.1 \
  --mysql=8.0 \
  --start \
  --pull-db

# Import an existing DDEV project
stax init --from-ddev

# Print a config template to stdout
stax init --template

# Show example config with comments
stax init --show-example
```

**Key flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | directory name | Project name |
| `--type` | `wordpress` | `wordpress` or `wordpress-multisite` |
| `--mode` | `subdomain` | Multisite mode: `subdomain` or `subdirectory` |
| `--install` | | WPEngine install name |
| `--environment` | `production` | WPEngine environment |
| `--php` | `8.1` | PHP version |
| `--mysql` | `8.0` | MySQL version |
| `--repo` | | Git repository URL to clone |
| `--branch` | `main` | Git branch |
| `--start` | | Start DDEV after init |
| `--pull-db` | | Pull database after init |
| `--pull-files` | | Pull files after init |
| `--skip-wordpress` | | Skip WordPress core download |

---

### Environment

#### `stax start`
Start the DDEV environment.

```bash
stax start

# Start with Xdebug enabled
stax start --xdebug

# Start and run the build
stax start --build
```

#### `stax stop`
Stop the DDEV environment.

```bash
stax stop

# Stop all DDEV projects
stax stop --all
```

#### `stax restart`
Restart the DDEV environment.

```bash
stax restart
```

#### `stax status`
Show the current environment status — DDEV, WPEngine connection, configuration.

```bash
stax status
```

---

### Database

#### `stax db pull`
Pull a database from WPEngine, import it locally, and automatically replace URLs.

```bash
# Pull from the configured environment
stax db pull

# Pull from staging
stax db pull --environment=staging

# Skip automatic URL replacement
stax db pull --skip-replace

# Skip creating a snapshot before pull
stax db pull --snapshot=false

# Exclude specific tables
stax db pull --exclude-tables=wp_logs,wp_actionscheduler_logs
```

**Key flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--environment` | from config | `production`, `staging`, or `development` |
| `--snapshot` | `true` | Create a snapshot before pulling |
| `--skip-replace` | | Skip URL search-replace after import |
| `--skip-logs` | | Exclude log tables from export |
| `--skip-transients` | | Exclude transient data |
| `--skip-spam` | | Exclude spam comments |
| `--exclude-tables` | | Comma-separated tables to exclude |

#### `stax db push`
Push your local database to WPEngine.

```bash
# Dry run first
stax db push --environment=staging --dry-run

# Push to staging
stax db push --environment=staging

# Push without creating a remote backup first
stax db push --environment=staging --skip-backup
```

#### `stax snapshot`
Create and manage local database snapshots.

```bash
# Create a snapshot
stax snapshot

# Create with a description
stax snapshot --description "before-migration"

# List snapshots
stax snapshot list

# Restore a snapshot
stax snapshot restore <name>

# Delete a snapshot
stax snapshot delete <name>

# Clean up old snapshots
stax snapshot clean
```

---

### Files

#### `stax files pull`
Sync files from WPEngine to your local `wp-content` directory using rsync over SSH.

```bash
# Pull all wp-content
stax files pull

# Pull from staging
stax files pull --environment=staging

# Pull only themes
stax files pull --themes-only

# Pull only plugins
stax files pull --plugins-only

# Pull only mu-plugins
stax files pull --mu-plugins-only

# Exclude uploads directory
stax files pull --exclude-uploads

# Dry run (preview without syncing)
stax files pull --dry-run

# Delete local files not present on remote
stax files pull --delete

# Limit bandwidth to 1000 KB/s
stax files pull --bandwidth-limit=1000
```

#### `stax files push`
Push local files to WPEngine.

```bash
stax files push --environment=staging --dry-run
stax files push --environment=staging --themes-only
```

---

### Configuration

#### `stax config`
View and modify your `.stax.yml` configuration from the command line.

```bash
# Show all configuration
stax config show

# Get a specific value
stax config get ddev.php_version
stax config get provider_config.install

# Set a value
stax config set ddev.php_version 8.2
stax config set provider_config.environment staging

# Show config as JSON or YAML
stax config show --format=json
stax config show --format=yaml
```

#### `stax config migrate`
Migrate a `.stax.yml` from an older schema version to the current version (v2).

```bash
# Preview what would change
stax config migrate --dry-run

# Migrate (creates a backup automatically)
stax config migrate
```

#### `stax validate`
Validate your `.stax.yml` against the schema.

```bash
stax validate
```

---

### Diagnostics

#### `stax doctor`
Run diagnostics and check for common issues. Can auto-fix some problems.

```bash
stax doctor

# Auto-fix detected issues
stax doctor --fix

# Check a specific project directory
stax doctor --project-dir=/path/to/project
```

Checks include:
- Docker and DDEV installation
- `.stax.yml` presence and validity
- WPEngine environment configuration
- SSH key and credential setup

---

### Build & Development

#### `stax build`
Run the project build (Composer, npm, build scripts).

```bash
stax build

# Composer only
stax build composer

# npm only
stax build npm
```

#### `stax dev`
Start development mode with file watching and hot-reload.

```bash
stax dev

# Dev mode for a specific theme
stax dev theme my-theme

# Stop dev mode
stax dev stop
```

#### `stax lint`
Run PHP CodeSniffer.

```bash
# Check for issues
stax lint check

# Auto-fix issues
stax lint fix

# Lint only staged files (useful pre-commit)
stax lint --staged
```

---

### WPEngine

#### `stax list`
List all WPEngine installs on your account.

```bash
stax list
```

#### `stax wpengine`
Global WPEngine discovery and management.

```bash
# List all installs
stax wpengine list

# Get details for a specific install
stax wpengine info my-install
```

---

### GitHub Actions

#### `stax actions setup`
Generate GitHub Actions workflow files for WPEngine deployment.

```bash
# Basic setup (main → production)
stax actions setup --prod-install my-install

# With staging branch
stax actions setup \
  --production=main \
  --staging=develop \
  --prod-install=my-install \
  --stage-install=my-install-staging

# Overwrite existing files
stax actions setup --force
```

Creates:
- `.github/workflows/deploy.yml` — deployment workflow using `wpengine/github-action-wpe-site-deploy`
- `.github/CODEOWNERS` — template code owners file

---

### Other Commands

```bash
stax repo init         # Initialize Git for an existing WordPress site
stax media setup-proxy # Configure DDEV nginx media proxying
stax media status      # Show media proxy status
stax version           # Show version and feature status
stax man               # Generate man page
```

---

## Common Workflows

### Starting a new project from scratch

```bash
mkdir my-site && cd my-site
stax init --name=my-site --install=my-install --start --pull-db
# Done — site is running at https://my-site.ddev.site
```

### Onboarding to an existing team project

```bash
cd ~/Sites
stax init \
  --name=team-site \
  --repo=https://github.com/org/team-site.git \
  --install=teaminstall \
  --start \
  --pull-db \
  --pull-files
```

### Daily development

```bash
stax start                            # Start the environment
stax db pull --environment=staging    # Sync latest database
# ... do your work ...
stax stop                             # Shut down at end of day
```

### Pulling a production database safely

```bash
stax snapshot --description "before-prod-pull"
stax db pull --environment=production
# Something's wrong?
stax snapshot restore before-prod-pull
```

### Deploying to staging via CI

```bash
# Set up GitHub Actions (one-time)
stax actions setup \
  --production=main \
  --staging=develop \
  --prod-install=mysite \
  --stage-install=mysite-staging

git add .github/
git commit -m "chore: add GitHub Actions deployment"
git push  # → triggers deploy to staging
```

### Importing an existing DDEV project

```bash
cd existing-ddev-project
stax init --from-ddev
# Stax reads existing DDEV config and creates .stax.yml
```

---

## Configuration Reference

### Version

```yaml
version: 2  # Required. Only version 2 is supported.
```

### project

```yaml
project:
  name: my-site               # Project name (used for DDEV domain)
  type: wordpress             # wordpress | wordpress-multisite
  mode: single                # single | subdomain | subdirectory
  description: My site        # Optional description
```

### provider

```yaml
provider: wpengine            # Hosting provider (currently: wpengine)

provider_config:
  install: my-install         # WPEngine install name
  environment: production     # production | staging | development
  ssh_gateway: ssh.wpengine.net
  account_name: myaccount     # Optional
```

### ddev

```yaml
ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  mysql_type: mysql           # mysql | mariadb
  webserver_type: nginx-fpm
  nodejs_version: "20"
  composer_version: "2"
  xdebug_enabled: false
  mutagen_enabled: true       # Improves performance on macOS
```

### wordpress

```yaml
wordpress:
  version: latest
  locale: en_US
  table_prefix: wp_
```

### network (multisite only)

```yaml
network:
  domain: my-site.ddev.site
  title: My Network
  sites:
    - name: site-one
      slug: site-one
      title: Site One
      domain: site-one.my-site.ddev.site
      provider_domain: site-one.wpengine.com
      active: true
```

### repository

```yaml
repository:
  url: https://github.com/org/repo.git
  branch: main
  depth: 1
```

### media

```yaml
media:
  proxy_enabled: true         # Serve uploads from WPEngine on-demand
```

### snapshots

```yaml
snapshots:
  directory: ~/.stax/snapshots
  auto_snapshot_before_pull: true
  retention:
    auto: 7    # days
    manual: 30 # days
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `STAX_PROVIDER` | Override the provider (e.g. `wpengine`) |
| `STAX_PROJECT_NAME` | Override the project name |
| `STAX_WPENGINE_INSTALL` | Override the WPEngine install name |
| `STAX_WPENGINE_ENVIRONMENT` | Override the WPEngine environment |
| `STAX_DDEV_PHP_VERSION` | Override the PHP version |

---

## Credential Setup

Credentials are stored securely in the macOS Keychain via `stax setup`. You'll need:

1. **WPEngine API credentials** — from your WPEngine account under API Access
2. **SSH private key** — for database and file sync (from WPEngine SSH Gateway settings)

```bash
stax setup
```

WPEngine API credentials can also be provided via environment variables for CI:

```bash
WPENGINE_API_USER=user@example.com
WPENGINE_API_PASSWORD=yourpassword
```

---

## Schema Migration

If you have a `.stax.yml` from before v1.0.0, migrate it:

```bash
# Preview
stax config migrate --dry-run

# Apply (backs up the original automatically)
stax config migrate
```

**Before (v1 format):**
```yaml
wpengine:
  install: my-install
  environment: production
```

**After (v2 format):**
```yaml
version: 2
provider: wpengine
provider_config:
  install: my-install
  environment: production
```

---

## Troubleshooting

**Environment not starting?**
```bash
stax doctor
```

**Database pull fails with credential error?**
```bash
stax setup   # Re-enter credentials
```

**URLs not replaced after db pull?**
```bash
# Run manually
ddev wp search-replace 'https://my-install.wpengine.com' 'https://my-site.ddev.site' --all-tables
```

**Snapshot won't restore?**
```bash
stax snapshot list   # Check available snapshots
```

For more help: `stax --help` or `stax <command> --help`
