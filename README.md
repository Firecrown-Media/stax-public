# Stax - WordPress Development CLI

**A powerful CLI tool for WordPress development with seamless WPEngine integration.**

Stax streamlines your WordPress development workflow - from environment setup to database syncing - so you can focus on building great sites instead of wrestling with configuration.

## Public Distribution Repository

This is the public distribution repository for Stax. Releases and binaries are published here for Homebrew installation and general distribution.

For development, issues, and contributions, please visit the main development repository (currently private).

## Installation

### Via Homebrew (Recommended)

```bash
brew install firecrown-media/tap/stax
```

### Verify Installation

```bash
stax --version
```

## Quick Start

```bash
# 1. Configure credentials
stax setup

# 2. Initialize your project
mkdir my-wordpress-site && cd my-wordpress-site
stax init --start

# 3. Your WordPress site is now running!
stax status
```

## Key Features

- **One-Command Setup** - Go from empty directory to running WordPress in one command
- **Automatic WordPress Download** - WordPress core automatically downloaded during init
- **Automatic Configuration** - wp-config.php generated with correct database settings
- **Single Site & Multisite** - Full support for standard WordPress and multisite networks
- **Automatic Database Sync** - Pull databases from WPEngine with automatic URL replacement
- **Remote Media Proxying** - Serve production media without downloading files
- **Safe Database Snapshots** - Create restore points before risky operations
- **Team-Friendly** - Share configuration files via Git
- **Build Automation** - Automatically runs Composer, npm, and build scripts
- **Secure Credentials** - Stores API keys and passwords safely in macOS Keychain

## Documentation

For complete documentation, please visit the [main repository](https://github.com/firecrown-media/stax).

Key documentation:
- [Getting Started Guide](docs/GETTING_STARTED.md)
- [Quick Start](docs/QUICK_START.md)
- [Command Reference](docs/COMMAND_REFERENCE.md)
- [WPEngine Integration](docs/WPENGINE.md)
- [Multisite Support](docs/MULTISITE.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
- [FAQ](docs/FAQ.md)

## Common Commands

```bash
# Environment Management
stax start                    # Start your development environment
stax stop                     # Stop your environment
stax restart                  # Restart your environment
stax status                   # Show environment status
stax doctor                   # Diagnose and fix issues

# Database Operations
stax db pull                  # Pull database from WPEngine
stax db snapshot              # Create a database backup
stax db restore <name>        # Restore a snapshot

# WordPress Management
stax wp <command>             # Run any WP-CLI command
stax wp plugin list           # List plugins
stax wp cache flush           # Flush WordPress cache
```

## Prerequisites

- **macOS 12+** (Monterey or later)
- **Docker Desktop** - For running containers
- **DDEV** - Container platform for WordPress
- **WPEngine Account** - With appropriate access (if using WPEngine features)

## System Requirements

- **macOS**: 12.0 (Monterey) or later
- **Processor**: Intel or Apple Silicon
- **RAM**: 8GB minimum, 16GB recommended
- **Disk**: 10GB free space minimum
- **Docker Desktop**: 4.25 or later
- **DDEV**: 1.22 or later

## Support

For issues and support:
- Documentation: See links above
- Issues: [GitHub Issues](https://github.com/firecrown-media/stax-public/issues)

## License

Proprietary - Firecrown Media

---

**Made with care by Firecrown Media**

[Documentation](https://github.com/firecrown-media/stax) • [Installation](https://github.com/firecrown-media/stax-public/releases)
