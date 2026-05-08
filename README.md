# Stax

A CLI tool for WordPress development with WPEngine integration.

Stax automates local environment setup, database sync, file sync, and — for teams moving off WPEngine — the full migration workflow to WordPress VIP.

---

## Install

```bash
brew install firecrown-media/stax/stax
```

Verify:

```bash
stax --version
```

---

## Quick start

```bash
# Store your WPEngine credentials
stax setup

# Initialize a project (run from your DDEV project root)
stax init

# Pull the production database
stax db pull

# Pull files (wp-content, excluding uploads)
stax files pull --exclude-uploads

# Start DDEV
stax start
```

---

## What stax does

| Command group | What it does |
|---------------|--------------|
| `stax db` | Pull and push databases, manage snapshots |
| `stax files` | Sync wp-content via rsync over SSH |
| `stax media` | Set up nginx remote media proxy |
| `stax migrate` | Orchestrate WPEngine → VIP migration |
| `stax wp` | Run WP-CLI commands in the DDEV container |
| `stax config` | Read and validate `.stax.yml` |
| `stax doctor` | Check prerequisites, auto-fix common issues |

---

## Documentation

**Runbooks** (task-oriented):

- [Getting started](docs/runbooks/getting-started.md)
- [Database operations](docs/runbooks/database.md)
- [File sync](docs/runbooks/files.md)
- [Remote media proxy](docs/runbooks/media-proxy.md)
- [WPEngine → VIP migration](docs/runbooks/migration.md)
- [Multisite](docs/runbooks/multisite.md)
- [Troubleshooting](docs/runbooks/troubleshooting.md)

**Command reference:**

- [All command groups](docs/commands/overview.md)
- [stax db](docs/commands/db.md)
- [stax files](docs/commands/files.md)
- [stax migrate](docs/commands/migrate.md)
- [stax wp](docs/commands/wp.md)
- [stax config](docs/commands/config.md)

**Man pages** (after install):

```bash
man stax
man stax-db
man stax-files
man stax-migrate
```

---

## Requirements

- macOS 12+
- Docker (Docker Desktop or Colima)
- [DDEV](https://ddev.readthedocs.io/en/stable/users/install/)
- WPEngine account with SSH access enabled

---

## Build from source

```bash
git clone https://github.com/firecrown-media/stax-public.git
cd stax-public
go mod download
PATH="/opt/homebrew/bin:$PATH" make build
```

---

## Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md).
