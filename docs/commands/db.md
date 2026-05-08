Reference for all `stax db` subcommands and flags.

## stax db pull

Pull the remote database, import it locally, and run URL search-replace.

```bash
stax db pull [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--environment` | string | from config | WPEngine environment (production, staging) |
| `--snapshot` | bool | true | Create a local snapshot before importing |
| `--skip-replace` | bool | false | Skip URL search-replace after import |
| `--exclude-tables` | string | — | Comma-separated table names to exclude |
| `--skip-logs` | bool | true | Exclude log tables |
| `--skip-transients` | bool | true | Exclude transient tables |
| `--skip-spam` | bool | true | Exclude spam/trash tables |
| `--sanitize` | bool | false | Sanitize user data after import |

**Examples:**

```bash
# Pull from production (default)
stax db pull

# Pull from staging
stax db pull --environment=staging

# Pull without creating a snapshot first
stax db pull --snapshot=false

# Pull and skip URL replacement
stax db pull --skip-replace

# Exclude specific tables
stax db pull --exclude-tables=wp_users,wp_usermeta
```

---

## stax db push

Push the local database to a remote WPEngine environment.

```bash
stax db push [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--environment` | string | — | Target environment (required: staging or production) |
| `--dry-run` | bool | false | Preview the operation without making changes |
| `--skip-backup` | bool | false | Skip creating a remote backup before import |
| `--skip-replace` | bool | false | Skip URL replacement after import |

**Examples:**

```bash
# Dry run to staging
stax db push --environment=staging --dry-run

# Push to staging
stax db push --environment=staging

# Push to production (prompts for confirmation)
stax db push --environment=production

# Push without creating a remote backup
stax db push --environment=staging --skip-backup
```

---

## stax db snapshot

Create and manage timestamped local database snapshots.

### stax db snapshot (create)

```bash
stax db snapshot [--description string]
```

Creates a snapshot in `.ddev/db_snapshots/`. Snapshots are retained for 30 days by default.

```bash
stax db snapshot
stax db snapshot --description "before-plugin-update"
```

### stax db snapshot list

```bash
stax db snapshot list
```

Lists all local snapshots with their names and creation timestamps.

### stax db snapshot restore

```bash
stax db snapshot restore <name>
```

Restores the named snapshot, replacing the current local database.

### stax db snapshot delete

```bash
stax db snapshot delete <name>
```

Permanently deletes the named snapshot.

### stax db snapshot clean

```bash
stax db snapshot clean
```

Removes snapshots older than the configured retention period.
