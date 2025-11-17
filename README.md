# Stax

**A powerful CLI tool for WordPress development with seamless WPEngine integration.**

Stax streamlines your WordPress development workflow - from environment setup to database syncing - so you can focus on building great sites instead of wrestling with configuration.

## Repository Structure

This is the development repository for Stax. All development, issues, and planning happen here.

**Hybrid Architecture (3 Repositories):**
1. **Development**: [Firecrown-Media/stax](https://github.com/Firecrown-Media/stax) (this repo)
   - Development code, issues, PRs, and artifacts
   - Can be private while maintaining public distribution
2. **Public Distribution**: [Firecrown-Media/stax-public](https://github.com/Firecrown-Media/stax-public)
   - Clean source code and releases for public consumption
   - Automatically synced on every release
   - No sensitive files or development artifacts
3. **Homebrew Tap**: [Firecrown-Media/homebrew-stax](https://github.com/Firecrown-Media/homebrew-stax)
   - Homebrew formula pointing to stax-public
   - Automatically updated by GoReleaser

**Automated Release Flow:**
```
Commit → Release-please PR → Merge → GoReleaser builds
→ Release in stax-public → Sync workflow → Homebrew formula updated
→ Users: brew install firecrown-media/stax/stax
```

> **Latest Release: v2.12.5** (2025-11-16)
>
> **Recent Fix (v2.12.5):**
> - ✓ Media proxy file sync now respects configuration (PR #88)
> - ✓ Fixed upload directory exclusion logic in `stax init`
> - ✓ Added validation for critical WordPress directories
>
> **Recent Major Features:**
> - ✓ One-command WordPress setup (v2.5.0+)
> - ✓ Automatic database import and URL replacement (v2.4.0+)
> - ✓ DNS resolution - no sudo prompts for hosts file (v2.11.0+)
> - ✓ Hybrid public mirror for distribution (v2.12.0+)
> - ✓ Advanced file operations with checksums (v2.6.0+)
> - ✓ Database and file push capabilities (v2.7.0+)
> - ✓ Enhanced diagnostics with auto-fix (v2.8.0+)
> - ✓ Database snapshot management (v2.10.0+)

---

## What is Stax?

Stax is a command-line tool that makes WordPress development simple and consistent. Whether you're working on a single site or multisite network, Stax integrates with DDEV for local development and WPEngine for database synchronization.

**What Stax provides:**
- Complete environment management (start, stop, restart, status)
- Automatic database sync from WPEngine
- Support for both single site and multisite
- System diagnostics and health checks
- Configuration validation
- Secure credential storage (macOS Keychain)
- Build automation
- WP-CLI integration

## Key Features

- **One-Command Setup** - Go from empty directory to running WordPress in one command
- **Automatic WordPress Download** - WordPress core automatically downloaded during init
- **Automatic Configuration** - wp-config.php generated with correct database settings
- **Single Site & Multisite** - Full support for standard WordPress and multisite networks
- **Multisite Ready** - Multisite networks configured automatically with proper constants
- **Automatic Database Sync** - Pull databases from WPEngine with automatic URL replacement
- **Remote Media Proxying** - Serve production media from WPEngine or CDN without downloading files. nginx automatically fetches images on-demand, saving 10GB-200GB of disk space and hours of sync time. Optional caching for fast performance.
- **Safe Database Snapshots** - Create restore points before risky operations
- **Team-Friendly** - Share configuration files via Git, everyone gets identical environments
- **Build Automation** - Automatically runs Composer, npm, and build scripts
- **Secure Credentials** - Stores API keys and passwords safely in macOS Keychain

## Quick Start

Get started with Stax in under 5 minutes:

```bash
# 1. Install Stax
brew install firecrown-media/stax/stax

# 2. Initialize your project (one command!)
mkdir my-wordpress-site && cd my-wordpress-site
stax init --start

# 3. Your WordPress site is now running!
# - WordPress core downloaded automatically ✓
# - Database configured automatically ✓
# - Site accessible at https://my-wordpress-site.ddev.site ✓
# - No sudo prompts needed! ✓

stax status
```

See [Getting Started Guide](docs/GETTING_STARTED.md) for detailed walkthrough.

## Who Should Use Stax?

**Stax is perfect for:**
- WordPress developers (single site or multisite)
- Teams using WPEngine hosting
- Developers who want consistent local environments
- Anyone tired of manual database imports and search-replace
- Teams transitioning from LocalWP to a more automated workflow

**You should use Stax if you:**
- Work with WordPress (single sites or multisite networks)
- Need to frequently sync databases from production/staging
- Want identical development environments across your team
- Prefer command-line tools over GUI applications
- Work on Mac (macOS 12 Monterey or later)

## Installation

### Via Homebrew (Recommended)
```bash
brew install firecrown-media/stax/stax
```

### Verify Installation
```bash
stax --version
# Should show: stax version 2.12.5
```

### Next Steps
1. Run `stax setup` to configure credentials
2. Run `stax init` in your project directory
3. See [Getting Started Guide](docs/GETTING_STARTED.md) for detailed walkthrough

### Build from Source (Advanced)
For development or contributing:
```bash
# Clone the development repository
git clone https://github.com/firecrown-media/stax.git
cd stax

# Install dependencies and build
go mod download
make build

# Install locally
make install
```

Note: Public distribution happens through [stax-public](https://github.com/firecrown-media/stax-public) automatically.

## Prerequisites

Before using Stax, you'll need:

- **macOS 12+** (Monterey or later) - macOS only for now
- **Docker Desktop** - For running containers ([Download](https://www.docker.com/products/docker-desktop))
- **DDEV** - Container platform for WordPress ([Install Guide](https://ddev.readthedocs.io/en/stable/users/install/))
- **WPEngine Account** - With appropriate access (if using WPEngine features)
  - Requires Owner role to enable API access initially
  - Requires appropriate user role (Owner, Full User, or Partial User) for install access
  - See [WPEngine Setup Guide](docs/WPENGINE.md#getting-started) for detailed instructions
  - Official guide: [WPEngine User Roles](https://wpengine.com/support/users/)

Don't worry - the [Getting Started Guide](docs/GETTING_STARTED.md) walks you through setting up each prerequisite.

## Common Commands

### Project Setup
```bash
stax init                     # Initialize a new Stax project (interactive)
stax setup                    # Configure WPEngine and GitHub credentials
stax list                     # List available WPEngine installations
```

### Environment Management (✓ Fully Implemented in v2.0.0)
```bash
stax start                    # Start your development environment
stax start --xdebug           # Start with Xdebug enabled
stax start --build            # Start and run build process
stax stop                     # Stop your environment
stax stop --all               # Stop all DDEV projects
stax restart                  # Restart your environment
stax status                   # Show environment status
stax doctor                   # Diagnose and fix issues
stax validate                 # Validate project configuration
```

### Database Operations
```bash
stax db pull                  # Pull database from WPEngine
stax db pull --skip-backup    # Pull without local backup
stax db snapshot              # Create local database snapshot
stax db restore               # Restore from snapshot
```

### Build & Development
```bash
stax build                    # Run the build process
stax dev                      # Start development mode with file watching
stax lint                     # Run PHP CodeSniffer
```

### WordPress Operations
```bash
stax wp -- plugin list        # Run WP-CLI commands
stax wp -- search-replace     # Run search-replace operations

# View logs
stax logs -f

# Stop when done for the day
stax stop

# Restart tomorrow
stax start
```

**Detailed walkthrough:** [docs/QUICK_START.md](./docs/QUICK_START.md)

## Common Commands

```bash
# Environment Management (✓ Fully Implemented in v2.0.0)
stax start                    # Start your development environment
stax stop                     # Stop your environment
stax restart                  # Restart your environment
stax status                   # Show environment status
stax doctor                   # Diagnose and fix issues

# Database Operations
stax db pull                  # Pull database from WPEngine
stax db snapshot              # Create a database backup
stax db restore <name>        # Restore a snapshot
stax db list                  # List all snapshots

# Configuration
stax config list              # Show your configuration
stax config set <key> <value> # Update a setting

# WordPress Management
stax wp <command>             # Run any WP-CLI command
stax wp plugin list           # List plugins
stax wp site list             # List all subsites
stax wp cache flush           # Flush WordPress cache

# Build and Development
stax build                    # Run build process
stax dev                      # Start dev mode (with watch)
stax lint                     # Run code linters
```

**Full command reference:** [docs/COMMAND_REFERENCE.md](./docs/COMMAND_REFERENCE.md)

## Documentation

### Quick Reference
- **Man Page**: `man stax` - Complete command reference
- **Quick Help**: `stax --help` - Interactive help
- **Online Docs**: See `docs/` directory

### Getting Started
- [Installation Guide](./docs/INSTALLATION.md) - Detailed installation instructions
- [Quick Start](./docs/QUICK_START.md) - Get up and running in 5 minutes
- [User Guide](./docs/USER_GUIDE.md) - Comprehensive usage guide

### Workflows
- [Working with Multisite](./docs/MULTISITE.md) - Multisite-specific features
- [WPEngine Integration](./docs/WPENGINE.md) - WPEngine features and setup
- [Media Proxy Guide](./docs/MEDIA_PROXY.md) - Serve remote media without downloads
- [Real-World Examples](./docs/EXAMPLES.md) - Common scenarios and workflows

### Step-by-Step Guides
- [Media Proxy Setup](./docs/guides/media-proxy-setup.md) - Configure media proxy step-by-step
- [File Sync Troubleshooting](./docs/guides/troubleshooting-file-sync.md) - Fix file synchronization issues
- [All Guides](./docs/guides/README.md) - Browse all step-by-step guides

### Reference
- [Command Reference](./docs/COMMAND_REFERENCE.md) - All commands and options
- [Configuration](./docs/CONFIG_SPEC.md) - Configuration file reference
- [Man Page Guide](./docs/MAN_PAGE.md) - Using the man page
- [FAQ](./docs/FAQ.md) - Frequently asked questions
- [Troubleshooting](./docs/TROUBLESHOOTING.md) - Common issues and solutions

### Advanced
- [Architecture](./docs/ARCHITECTURE.md) - System architecture and design
- [Build Process](./docs/BUILD_PROCESS.md) - How builds work
- [Provider System](./docs/MULTI_PROVIDER.md) - Multi-provider support

## How It Works

Stax uses industry-standard, battle-tested tools:

1. **DDEV** - Manages Docker containers for WordPress, MySQL, and other services
2. **WP-CLI** - Handles WordPress operations and database imports
3. **WPEngine API** - Pulls databases and syncs files from your hosting
4. **macOS Keychain** - Securely stores your credentials
5. **Git** - Clones your repositories and manages code

When you run `stax init`, here's what happens:

```
1. Validates your WPEngine credentials
2. Clones your GitHub repository
3. Detects PHP/MySQL versions from WPEngine
4. Generates DDEV configuration
5. Starts Docker containers
6. Downloads WordPress core automatically
7. Generates wp-config.php with database credentials
8. Installs Composer and npm dependencies
9. Runs your build scripts
10. Pulls the database from WPEngine
11. Imports and runs search-replace automatically
12. Displays your local URLs

Total time: 2-5 minutes
```

## Why Stax?

### vs. LocalWP

| Feature | LocalWP | Stax |
|---------|---------|------|
| **Setup Speed** | 10-30 minutes | 2-5 minutes |
| **Automation** | Mostly manual | Fully automated |
| **CLI Support** | Limited | Full CLI |
| **Multisite Subdomains** | Manual hosts file | Automatic |
| **Database Sync** | Manual export/import | One command |
| **Search-Replace** | Manual | Automatic |
| **Team Consistency** | Variable | Identical |
| **Version Control** | Config not shareable | Config in Git |
| **WPEngine Integration** | None | Built-in |
| **Build Automation** | Manual | Automatic |

### vs. Manual Docker Setup

| Feature | Manual Docker | Stax |
|---------|---------------|------|
| **Configuration** | Complex YAML files | Interactive prompts |
| **SSL Certificates** | Manual setup | Automatic |
| **Database Import** | Manual commands | One command |
| **Multisite DNS** | Manual /etc/hosts | Automatic |
| **Learning Curve** | Steep | Gentle |
| **Maintenance** | High | Low |

## Real-World Examples

### Daily Development Workflow

```bash
# Monday morning
cd ~/Sites/firecrown-multisite
stax start                    # Start your environment (10 seconds)

# Pull latest database from staging
stax db pull --environment=staging

# Work on a feature
git checkout -b feature/new-header
# ... make changes ...

# Test your changes
stax wp cache flush
open https://mysite.local

# End of day
stax stop
```

### Testing with Production Data

```bash
# Create a snapshot before testing
stax db snapshot before-testing

# Pull production database
stax db pull --environment=production

# Test your migration script
stax wp db query --file=migration.sql

# Something broke? No problem!
stax db restore before-testing
```

### Onboarding a New Team Member

```bash
# New developer's machine
git clone https://github.com/mycompany/my-project.git
cd my-project

# One-time setup
stax setup

# Project is ready!
stax init

# That's it! They have the exact same environment
```

**More examples:** [docs/EXAMPLES.md](./docs/EXAMPLES.md)

## Single Site Support

Stax works great with standard single-site WordPress installations:

```yaml
# .stax.yml
project:
  name: my-site
  type: wordpress  # Single site

wordpress:
  domain: mysite.local
```

All the same features work for single sites:
- Automatic database sync from WPEngine
- Remote media proxying
- Database snapshots and restore
- Build automation
- Team-friendly configuration

**Perfect for:** Most WordPress projects, client sites, blogs, marketing sites, or any standard WordPress installation.

## Multisite Support (Optional)

Need multisite? Stax has first-class support for WordPress multisite networks in both modes:

### Subdomain Multisite

```yaml
# .stax.yml
project:
  mode: subdomain

network:
  domain: mynetwork.local
  sites:
    - name: site1
      domain: site1.mynetwork.local
    - name: site2
      domain: site2.mynetwork.local
```

Stax automatically:
- Configures wildcard SSL certificates
- Sets up DNS resolution for all subdomains
- Runs search-replace for each subsite
- Generates proper WordPress multisite config

### Subdirectory Multisite

```yaml
# .stax.yml
project:
  mode: subdirectory

network:
  domain: mynetwork.local
  sites:
    - name: site1
      path: /site1
    - name: site2
      path: /site2
```

Works seamlessly with both modes.

**Learn more:** [docs/MULTISITE.md](./docs/MULTISITE.md)

## Troubleshooting

### Common Issues

**Port conflicts**
```bash
# Check if ports are in use
stax doctor

# Fix: Stop conflicting services
sudo apachectl stop
```

**Database connection errors**
```bash
# Restart DDEV services
stax restart
```

**SSL certificate issues**
```bash
# DDEV uses mkcert - this is automatic
# If you see SSL errors, check:
ddev --version  # Make sure DDEV is up to date
```

**Can't access subsites**
```bash
# Check your configuration
stax config list

# Restart to regenerate DDEV config
stax restart
```

**SSH fails with Warp terminal**
```bash
# Known issue: Warp terminal is incompatible with WPEngine SSH
# Solution: Use alternative terminal (Terminal.app, iTerm2, etc.)
# See: docs/TROUBLESHOOTING.md for details
```

**Full troubleshooting guide:** [docs/TROUBLESHOOTING.md](./docs/TROUBLESHOOTING.md)

## Getting Help

If you run into issues:

1. **Check the docs** - Most questions are answered in our documentation
2. **Run diagnostics** - `stax doctor` will identify common problems
3. **Check logs** - `stax logs -f` shows detailed error messages
4. **Search issues** - Check GitHub issues for similar problems
5. **Ask for help** - Contact the development team or create an issue

## Contributing

Stax is an internal Firecrown Media tool, but we welcome contributions from team members!

**Development setup:**
```bash
# Clone the development repository
git clone https://github.com/firecrown-media/stax.git
cd stax

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Install locally
make install
```

**Release workflow:**
- Use conventional commits (feat:, fix:, docs:, etc.)
- Commits to main trigger release-please
- Merge release PR to automatically release
- GoReleaser builds and publishes to stax-public
- Homebrew formula updates automatically

## FAQ

**Q: Does Stax work on Windows or Linux?**
A: Currently Stax is macOS-only. Windows and Linux support may come in the future.

**Q: Can I use Stax with other hosting providers?**
A: Stax has built-in WPEngine support, but you can use it as a local development tool with any hosting provider. Database sync would be manual.

**Q: How is this different from wp-env or other tools?**
A: Stax is specifically designed for WordPress development (both single sites and multisite) with hosting integration. It's more opinionated and automated than general-purpose tools, with first-class support for WPEngine.

**Q: Do I need to download all my media files?**
A: No! Stax uses remote media proxying - it fetches media from your CDN/production server on-demand.

**Q: Can multiple team members use different configurations?**
A: The `.stax.yml` file is meant to be shared via Git for consistency. Personal preferences go in `~/.stax/config.yml`.

**More questions:** [docs/FAQ.md](./docs/FAQ.md)

## Requirements

- **macOS**: 12.0 (Monterey) or later
- **Processor**: Intel or Apple Silicon
- **RAM**: 8GB minimum, 16GB recommended
- **Disk**: 10GB free space minimum
- **Docker Desktop**: 4.25 or later
- **DDEV**: 1.22 or later

## Version History

- **v2.12.5** (2025-11-16) - Latest stable release
  - Media proxy file synchronization fix
  - Fixed upload directory exclusion logic in `stax init`
  - Added validation for critical WordPress directories
- **v2.12.4** (2025-11-16)
  - Hybrid public mirror infrastructure
  - DNS resolution (no sudo prompts)
  - Database snapshots and push/pull
  - File operations with checksums
  - Enhanced diagnostics with auto-fix
  - Complete one-command WordPress setup
  - Automatic database import and URL replacement

## License

Proprietary - Firecrown Media

## Support

For support with Stax:
- Documentation: [docs/](./docs/)
- Issues: [GitHub Issues](https://github.com/firecrown-media/stax/issues)
- Internal: Contact the development team

---

**Made with care by Firecrown Media**

[Documentation](./docs/) • [Quick Start](./docs/QUICK_START.md) • [FAQ](./docs/FAQ.md) • [Troubleshooting](./docs/TROUBLESHOOTING.md)

