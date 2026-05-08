Pull and push databases between WPEngine and your local DDEV environment.

## Pull a database

```bash
stax db pull
```

This command exports the database from WPEngine via SSH, imports it into your local DDEV environment, then runs search-replace to swap remote URLs with your local `.ddev.site` domain.

## Pull from staging

```bash
stax db pull --environment=staging
```

Overrides the `environment` value in `.stax.yml` for this run. Useful for comparing production and staging data without changing your config.

## Skip URL replacement

```bash
stax db pull --skip-replace
```

Imports the database without running search-replace. Use this when you have already performed URL replacement locally and do not want it overwritten, or when troubleshooting replacement issues in isolation.

## Exclude tables

```bash
stax db pull --exclude-tables=wp_users,wp_usermeta
```

Skips the specified tables during import. Useful for avoiding overwriting local user accounts or role configurations with remote values.

## Snapshots

Create and manage timestamped snapshots of your local database stored in `.ddev/db_snapshots/`.

```bash
stax db snapshot create
```

Creates a named snapshot of the current local database. You will be prompted for a name if one is not provided.

```bash
stax db snapshot list
```

Lists all available snapshots with their names and creation timestamps.

```bash
stax db snapshot restore <name>
```

Restores the named snapshot, replacing the current local database. This is destructive — the existing local database is overwritten.

```bash
stax db snapshot delete <name>
```

Permanently deletes the named snapshot from `.ddev/db_snapshots/`.

## Auto-snapshot before pull

To automatically create a snapshot before every `stax db pull`, add the following to `.stax.yml`:

```yaml
snapshots:
  auto_snapshot_before_pull: true
```

This gives you a restore point before each pull without requiring a manual snapshot step.

## Push a database

Always review what will change before pushing:

```bash
stax db push --dry-run
```

The `--dry-run` flag prints a summary of the operation without making any changes. Review it carefully before proceeding.

To push to staging:

```bash
stax db push --environment=staging
```

A push to production requires interactive confirmation at the prompt. Staging pushes do not.

## What not to do

- Never push to production without reviewing `--dry-run` output first. There is no undo — the remote database is replaced.
- Never push a database that contains local test data, placeholder content, or development credentials. Sanitize the database before any push to a shared environment.
