# Stax Documentation

Welcome to the Stax CLI documentation. This guide helps you navigate all available documentation organized by your needs.

## Quick Navigation

- **New to Stax?** Start with [Quick Start Guide](QUICK_START.md)
- **Installing Stax?** See [Installation Guide](INSTALLATION.md)
- **Need help?** Check [Troubleshooting](TROUBLESHOOTING.md) or [FAQ](FAQ.md)
- **Looking for commands?** See [Command Reference](COMMAND_REFERENCE.md)

---

## Documentation by Role

### I'm a WordPress Developer (Want to Use Stax)

```
START HERE
    │
    ▼
┌──────────────────┐
│  README.md       │ ← What is Stax? Why use it?
└────────┬─────────┘
         ▼
┌──────────────────┐
│ INSTALLATION.md  │ ← Install Stax and prerequisites
└────────┬─────────┘
         ▼
┌──────────────────┐
│  QUICK_START.md  │ ← Get running in 5 minutes
└────────┬─────────┘
         ▼
┌──────────────────┐
│ GETTING_STARTED  │ ← Complete onboarding
└────────┬─────────┘
         ▼
┌──────────────────┐
│  USER_GUIDE.md   │ ← Daily workflows and features
└────────┬─────────┘
         ▼
┌──────────────────────────────────────────┐
│  Specialized Guides (as needed):         │
│  • MULTISITE.md                          │
│  • MEDIA_PROXY.md                        │
│  • guides/media-proxy-setup.md           │
│  • guides/troubleshooting-file-sync.md   │
│  • WPENGINE.md                           │
│  • EXAMPLES.md                           │
└──────────────────┬───────────────────────┘
                   ▼
        ┌──────────────────┐
        │  Having Issues?  │
        └─────────┬────────┘
                  ▼
     ┌────────────┴────────────┐
     ▼                         ▼
┌─────────────┐      ┌─────────────────┐
│ TROUBLE     │      │ COMMAND         │
│ SHOOTING.md │      │ REFERENCE.md    │
│             │      │                 │
│ FAQ.md      │      │                 │
└─────────────┘      └─────────────────┘
```

**Estimated Time:** 30 minutes to first working project

---

### I'm a Developer (Want to Contribute to Stax)

```
START HERE
    │
    ▼
┌──────────────────────┐
│  ../README.md        │ ← Project overview
└──────────┬───────────┘
           ▼
┌──────────────────────┐
│  CONTRIBUTING.md     │ ← How to contribute
└──────────┬───────────┘
           ▼
┌──────────────────────────────────────────┐
│  Technical Documentation:                │
│  • technical/ARCHITECTURE.md             │
│  • technical/COMMANDS.md                 │
│  • technical/CONFIG_SPEC.md              │
│  • technical/WPENGINE_INTEGRATION.md     │
│  • technical/DDEV_MULTISITE_IMPL...md    │
└──────────┬───────────────────────────────┘
           ▼
┌──────────────────────┐
│  TESTING.md          │ ← Testing guide
└──────────────────────┘
```

**Estimated Time:** 2-4 hours to understand architecture

---

### I'm a Release Manager

```
START HERE
    │
    ▼
┌──────────────────────┐
│  release/README.md   │ ← Overview of release process
└──────────┬───────────┘
           ▼
┌──────────────────────────────────────────┐
│  Release Process:                        │
│  • release/QUICK_REFERENCE.md            │
│  • release/AUTOMATED_RELEASE_PROCESS.md  │
│  • release/FEATURE_BRANCH_WORKFLOW.md    │
└──────────────────────────────────────────┘
```

**Estimated Time:** 15-30 minutes per release

---

## Documentation by Task

### Task: Install and Set Up Stax

1. [../README.md](../README.md) - Overview and prerequisites
2. [INSTALLATION.md](INSTALLATION.md) - Detailed installation
3. [QUICK_START.md](QUICK_START.md) - First project setup

**Estimated Time:** 15-30 minutes

---

### Task: Set Up a New WordPress Project

1. [QUICK_START.md](QUICK_START.md) - Basic setup
2. [GETTING_STARTED.md](GETTING_STARTED.md) - Complete walkthrough
3. [COMMAND_REFERENCE.md](COMMAND_REFERENCE.md) - Command help

**Estimated Time:** 5-10 minutes (with `stax init --start`)

---

### Task: Configure Media Proxy

1. [MEDIA_PROXY.md](MEDIA_PROXY.md) - Understanding media proxy
2. [guides/media-proxy-setup.md](guides/media-proxy-setup.md) - Step-by-step setup
3. [guides/troubleshooting-file-sync.md](guides/troubleshooting-file-sync.md) - Fix sync issues

**Estimated Time:** 5-10 minutes

---

### Task: Work with WordPress Multisite

1. [MULTISITE.md](MULTISITE.md) - Multisite guide
2. [USER_GUIDE.md](USER_GUIDE.md) - Daily multisite workflows
3. [technical/DDEV_MULTISITE_IMPLEMENTATION.md](technical/DDEV_MULTISITE_IMPLEMENTATION.md) - Technical details

**Estimated Time:** 20-30 minutes to understand

---

### Task: Sync Database from WPEngine

1. [USER_GUIDE.md](USER_GUIDE.md) - Database operations
2. [WPENGINE.md](WPENGINE.md) - WPEngine integration
3. [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Fix sync issues

**Commands:**
```bash
stax db pull                         # Pull from production
stax db pull --environment=staging   # Pull from staging
```

---

### Task: Troubleshoot Issues

1. Run `stax doctor` - Automated diagnostics
2. [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common problems and solutions
3. [FAQ.md](FAQ.md) - Frequently asked questions
4. [guides/troubleshooting-file-sync.md](guides/troubleshooting-file-sync.md) - File sync specific

---

### Task: Contribute to Stax Development

1. [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
2. [technical/ARCHITECTURE.md](technical/ARCHITECTURE.md) - System design
3. [TESTING.md](TESTING.md) - Running tests

---

## Complete Documentation Index

### Getting Started

| Document | Purpose | Audience |
|----------|---------|----------|
| [../README.md](../README.md) | Project overview, features | Everyone |
| [INSTALLATION.md](INSTALLATION.md) | Detailed installation guide | New Users |
| [QUICK_START.md](QUICK_START.md) | 5-minute quickstart | New Users |
| [GETTING_STARTED.md](GETTING_STARTED.md) | Complete onboarding | New Users |
| [USER_GUIDE.md](USER_GUIDE.md) | Comprehensive usage guide | Active Users |

### Feature Guides

| Document | Purpose | Audience |
|----------|---------|----------|
| [MULTISITE.md](MULTISITE.md) | WordPress multisite guide | Users |
| [MEDIA_PROXY.md](MEDIA_PROXY.md) | Media proxy configuration | Users |
| [WPENGINE.md](WPENGINE.md) | WPEngine integration | Users |
| [EXAMPLES.md](EXAMPLES.md) | Real-world workflows | Users |

### Step-by-Step Guides

| Document | Purpose | Audience |
|----------|---------|----------|
| [guides/README.md](guides/README.md) | Guides overview | Users |
| [guides/media-proxy-setup.md](guides/media-proxy-setup.md) | Media proxy setup | Users |
| [guides/troubleshooting-file-sync.md](guides/troubleshooting-file-sync.md) | File sync troubleshooting | Users |

### Reference Documentation

| Document | Purpose | Audience |
|----------|---------|----------|
| [COMMAND_REFERENCE.md](COMMAND_REFERENCE.md) | All commands and options | Users |
| [FAQ.md](FAQ.md) | Frequently asked questions | Users |
| [TROUBLESHOOTING.md](TROUBLESHOOTING.md) | Common issues and solutions | Users |
| [SECURITY.md](SECURITY.md) | Security best practices | Users |

### Technical Documentation

| Document | Purpose | Audience |
|----------|---------|----------|
| [technical/README.md](technical/README.md) | Technical docs overview | Contributors |
| [technical/ARCHITECTURE.md](technical/ARCHITECTURE.md) | System architecture | Contributors |
| [technical/COMMANDS.md](technical/COMMANDS.md) | CLI command structure | Contributors |
| [technical/CONFIG_SPEC.md](technical/CONFIG_SPEC.md) | Configuration schema | Contributors |
| [technical/WPENGINE_INTEGRATION.md](technical/WPENGINE_INTEGRATION.md) | WPEngine provider | Contributors |
| [technical/DDEV_MULTISITE_IMPLEMENTATION.md](technical/DDEV_MULTISITE_IMPLEMENTATION.md) | Multisite implementation | Contributors |

### Contributing & Development

| Document | Purpose | Audience |
|----------|---------|----------|
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution guidelines | Contributors |
| [TESTING.md](TESTING.md) | Testing guide | Contributors |

### Release Documentation

| Document | Purpose | Audience |
|----------|---------|----------|
| [release/README.md](release/README.md) | Release overview | Release Managers |
| [release/QUICK_REFERENCE.md](release/QUICK_REFERENCE.md) | Quick commands | Release Managers |
| [release/AUTOMATED_RELEASE_PROCESS.md](release/AUTOMATED_RELEASE_PROCESS.md) | Automation details | Release Managers |
| [release/FEATURE_BRANCH_WORKFLOW.md](release/FEATURE_BRANCH_WORKFLOW.md) | Branch workflow | Developers |

### Archived Documentation

Historical documents preserved for reference (issues now resolved):

| Document | Purpose |
|----------|---------|
| [archived/README.md](archived/README.md) | Archive overview |
| [archived/BUGFIX_SUMMARY.md](archived/BUGFIX_SUMMARY.md) | URL replacement fix |
| [archived/INTERACTIVE_MODE_FIX.md](archived/INTERACTIVE_MODE_FIX.md) | TTY detection fix |
| [archived/URL_REPLACEMENT_FIX.md](archived/URL_REPLACEMENT_FIX.md) | URL replacement details |

---

## Common Documentation Paths

### Path 1: New User Journey
README → INSTALLATION → QUICK_START → GETTING_STARTED → USER_GUIDE → Specialized Guides

**Time:** 1-2 hours to read, lifetime to master

### Path 2: Quick Setup
QUICK_START → COMMAND_REFERENCE (as needed)

**Time:** 5-10 minutes

### Path 3: Troubleshooting
`stax doctor` → TROUBLESHOOTING → FAQ → GitHub Issues

**Time:** 5-30 minutes

### Path 4: Contributing
README → CONTRIBUTING → technical/ARCHITECTURE → TESTING

**Time:** 2-4 hours to understand

### Path 5: Release
release/QUICK_REFERENCE → release/AUTOMATED_RELEASE_PROCESS

**Time:** 15-30 minutes per release

---

## Need Help?

- Check the [FAQ](FAQ.md) for common questions
- Review [Troubleshooting](TROUBLESHOOTING.md) for common issues
- Run `stax doctor` to diagnose problems
- Run `stax <command> --help` for command-specific help

---

**Version:** 2.16.1
**Last Updated:** 2025-12-12
