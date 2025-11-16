# Stax Documentation

Welcome to the Stax CLI documentation. This guide helps you navigate all available documentation organized by your needs.

## Quick Navigation

- **New to Stax?** Start with [Quick Start Guide](QUICK_START.md)
- **Installing Stax?** See [Installation Guide](INSTALLATION.md)
- **Need help?** Check [Troubleshooting](TROUBLESHOOTING.md) or [FAQ](FAQ.md)
- **Looking for commands?** See [Command Reference](COMMAND_REFERENCE.md)
- **Contributing?** Check [Implementation Roadmap](IMPLEMENTATION_ROADMAP.md) for planned work

## Documentation by User Journey

### Getting Started (5-15 minutes)

Essential documentation to get up and running quickly:

1. [Installation Guide](INSTALLATION.md) - Install Stax via Homebrew or from source
2. [Quick Start Guide](QUICK_START.md) - 5-minute guide to your first project
3. [User Guide](USER_GUIDE.md) - Complete usage guide for daily workflows

### Core Concepts

Understanding how Stax works:

- [WordPress Multisite Guide](MULTISITE.md) - Working with multisite installations
- [WPEngine Integration](WPENGINE.md) - Using Stax with WPEngine hosting
- [Examples & Workflows](EXAMPLES.md) - Real-world usage scenarios

### Guides

Step-by-step guides for common workflows:

- [Guides Hub](guides/README.md) - All available guides
- [Media Proxy Setup](guides/media-proxy-setup.md) - Configure media proxy for efficient development
- [File Sync Troubleshooting](guides/troubleshooting-file-sync.md) - Diagnose and fix file sync issues

### Reference Documentation

Quick references when you need them:

- [Command Reference](COMMAND_REFERENCE.md) - All commands with flags and examples
- [Configuration Specification](technical/CONFIG_SPEC.md) - Complete config file reference
- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common issues and solutions
- [FAQ](FAQ.md) - Frequently asked questions

## Documentation by Role

### For End Users

Documentation for developers using Stax:

- [Installation Guide](INSTALLATION.md)
- [Quick Start Guide](QUICK_START.md)
- [User Guide](USER_GUIDE.md)
- [Multisite Guide](MULTISITE.md)
- [Command Reference](COMMAND_REFERENCE.md)
- [WPEngine Guide](WPENGINE.md)
- [Examples & Workflows](EXAMPLES.md)
- [Troubleshooting](TROUBLESHOOTING.md)
- [FAQ](FAQ.md)

**Step-by-Step Guides:**
- [Guides Hub](guides/README.md)
- [Media Proxy Setup](guides/media-proxy-setup.md)
- [File Sync Troubleshooting](guides/troubleshooting-file-sync.md)

### For Developers & Contributors

Technical documentation for those working on Stax itself:

#### Architecture & Design
- [System Architecture](technical/ARCHITECTURE.md) - High-level system design
- [Command Structure](technical/COMMANDS.md) - CLI command specifications
- [Configuration Specification](technical/CONFIG_SPEC.md) - Config file schema
- [Provider Interface](PROVIDER_INTERFACE.md) - Multi-provider abstraction
- [DDEV Multisite Implementation](technical/DDEV_MULTISITE_IMPLEMENTATION.md) - Container setup details

#### Development Guides
- [Provider Development](PROVIDER_DEVELOPMENT.md) - Adding new hosting providers
- [Multi-Provider Architecture](MULTI_PROVIDER.md) - Provider system overview
- [Build Process](BUILD_PROCESS.md) - Build system integration
- [WPEngine Integration](technical/WPENGINE_INTEGRATION.md) - WPEngine provider details
- [Testing Guide](TESTING.md) - Writing and running tests

#### Project Status & History
- [Implementation Roadmap](IMPLEMENTATION_ROADMAP.md) - Current progress and planned phases
- [Project Summary](development/PROJECT_SUMMARY.md) - Complete project overview
- [Implementation Summary](development/IMPLEMENTATION_SUMMARY.md) - What was built
- [Completion Summary](development/COMPLETION_SUMMARY.md) - Final delivery status
- [Test Suite Summary](development/TEST_SUITE_SUMMARY.md) - Testing implementation

### For Security Teams

Security-focused documentation:

- [Security Overview](SECURITY.md) - Security best practices and guidelines
- [Security Audit](SECURITY_AUDIT.md) - Complete security assessment
- [Security Review Summary](SECURITY_REVIEW_SUMMARY.md) - Executive summary
- [Security Checklist](SECURITY_CHECKLIST.md) - Pre-release security verification
- [Security Quick Reference](SECURITY_QUICK_REFERENCE.md) - Quick security guide
- [Security Scan Results](SECURITY_SCAN_RESULTS.md) - Automated scan findings

### For Release Managers

Release and deployment documentation:

- [Release Process](RELEASE_PROCESS.md) - How to create releases
- [Release Commands](release/RELEASE_COMMANDS.md) - Quick command reference
- [Release Ready Guide](release/RELEASE_READY.md) - Pre-release checklist
- [Release Quick Reference](RELEASE_QUICK_REFERENCE.md) - One-command operations
- [Homebrew Installation](HOMEBREW_INSTALLATION.md) - Installing from Homebrew
- [Homebrew Tap Setup](HOMEBREW_TAP_SETUP.md) - Setting up the Homebrew tap
- [CI/CD Pipeline](CICD_PIPELINE.md) - Automated build and release
- [Deployment Summary](DEPLOYMENT_SUMMARY.md) - Deployment infrastructure
- [Man Page](MAN_PAGE.md) - Unix manual page

## Documentation Categories

### Technical Documentation (`docs/technical/`)

In-depth technical documentation for system internals:

- [ARCHITECTURE.md](technical/ARCHITECTURE.md) - System architecture and design decisions
- [COMMANDS.md](technical/COMMANDS.md) - CLI command specifications and structure
- [CONFIG_SPEC.md](technical/CONFIG_SPEC.md) - Configuration file schema and validation
- [WPENGINE_INTEGRATION.md](technical/WPENGINE_INTEGRATION.md) - WPEngine provider implementation
- [DDEV_MULTISITE_IMPLEMENTATION.md](technical/DDEV_MULTISITE_IMPLEMENTATION.md) - Container and multisite setup

### Development Documentation (`docs/development/`)

Project status, history, and internal development docs:

- [IMPLEMENTATION_ROADMAP.md](IMPLEMENTATION_ROADMAP.md) - **Active roadmap** with planned phases and timeline
- [PROJECT_SUMMARY.md](development/PROJECT_SUMMARY.md) - Complete project overview and statistics
- [IMPLEMENTATION_SUMMARY.md](development/IMPLEMENTATION_SUMMARY.md) - Implementation details
- [COMPLETION_SUMMARY.md](development/COMPLETION_SUMMARY.md) - Final delivery status
- [COMPLETE_PROJECT_FINAL.md](development/COMPLETE_PROJECT_FINAL.md) - Project completion report
- [FINAL_PROJECT_STATUS.md](development/FINAL_PROJECT_STATUS.md) - Final status snapshot
- [TEST_SUITE_SUMMARY.md](development/TEST_SUITE_SUMMARY.md) - Test implementation summary

### Release Documentation (`docs/release/`)

Release process, packaging, and deployment:

- [RELEASE_COMMANDS.md](release/RELEASE_COMMANDS.md) - Command reference for releases
- [RELEASE_READY.md](release/RELEASE_READY.md) - Pre-release checklist and verification
- [MAN_PAGE_DELIVERABLES.md](release/MAN_PAGE_DELIVERABLES.md) - Manual page deliverables
- [MAN_PAGE_IMPLEMENTATION.md](release/MAN_PAGE_IMPLEMENTATION.md) - Manual page implementation
- [DEPLOYMENT_SETUP_COMPLETE.md](release/DEPLOYMENT_SETUP_COMPLETE.md) - Deployment setup status

### Internal Documentation (`docs/.internal/`)

AI assistant context and internal tools (not user-facing):

- [claude.md](.internal/claude.md) - AI assistant context file (excluded from user docs)

## Common Tasks

### Setting Up a New Project
1. Install Stax: [Installation Guide](INSTALLATION.md)
2. Configure credentials: `stax setup`
3. Initialize project: [Quick Start](QUICK_START.md)

### Daily Development Workflow
1. Start environment: `stax start`
2. Pull database: `stax db:pull`
3. Work on code
4. Stop environment: `stax stop`

See [User Guide](USER_GUIDE.md) for detailed workflows.

### Troubleshooting Issues
1. Check [Troubleshooting Guide](TROUBLESHOOTING.md)
2. Review [FAQ](FAQ.md)
3. Run diagnostics: `stax doctor`

### Contributing to Stax
1. Read [Architecture](technical/ARCHITECTURE.md)
2. Review [Testing Guide](TESTING.md)
3. Check [Security Guidelines](SECURITY.md)

## Version Information

- **Current Version:** 1.0.0
- **Last Updated:** 2025-11-09
- **Documentation Status:** Complete

## Need Help?

- Check the [FAQ](FAQ.md) for common questions
- Review [Troubleshooting](TROUBLESHOOTING.md) for common issues
- Run `stax doctor` to diagnose problems
- Run `stax <command> --help` for command-specific help

## Contributing to Documentation

Documentation is maintained alongside code. When contributing:

1. Keep user-facing docs simple and example-driven
2. Update technical docs when changing architecture
3. Add troubleshooting entries for common issues
4. Include examples in all command documentation

## License

This documentation is part of the Stax CLI project. See the LICENSE file in the project root.
