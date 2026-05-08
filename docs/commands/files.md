Reference for all `stax files` subcommands and flags.

## stax files pull

Pull wp-content files from WPEngine to your local environment via rsync over SSH.

```bash
stax files pull [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--environment` | string | from config | WPEngine environment |
| `--themes-only` | bool | false | Sync only `wp-content/themes/` |
| `--plugins-only` | bool | false | Sync only `wp-content/plugins/` |
| `--mu-plugins-only` | bool | false | Sync only `wp-content/mu-plugins/` |
| `--exclude-uploads` | bool | false | Exclude `wp-content/uploads/` |
| `--dry-run` | bool | false | Show what would be transferred without syncing |
| `--delete` | bool | false | Delete local files not present on remote |
| `--bandwidth-limit` | int | 0 | Bandwidth cap in KB/s (0 = unlimited) |
| `--include` | string | — | Comma-separated rsync include patterns |
| `--exclude` | string | — | Comma-separated rsync exclude patterns |
| `--preserve-permissions` | bool | false | Preserve file permissions during sync |
| `--verify` | bool | false | Verify file checksums after sync |

**Examples:**

```bash
# Pull all wp-content
stax files pull

# Pull themes only
stax files pull --themes-only

# Pull plugins only
stax files pull --plugins-only

# Dry run to see what would change
stax files pull --dry-run

# Pull without uploads (large sites)
stax files pull --exclude-uploads

# Pull from staging
stax files pull --environment=staging

# Limit bandwidth to 500 KB/s
stax files pull --bandwidth-limit=500
```

---

## stax files push

Push local wp-content files to a WPEngine environment via rsync over SSH.

```bash
stax files push [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--environment` | string | from config | WPEngine environment |
| `--themes-only` | bool | false | Sync only `wp-content/themes/` |
| `--plugins-only` | bool | false | Sync only `wp-content/plugins/` |
| `--mu-plugins-only` | bool | false | Sync only `wp-content/mu-plugins/` |
| `--uploads-only` | bool | false | Sync only `wp-content/uploads/` |
| `--dry-run` | bool | false | Show what would be transferred without syncing |
| `--delete` | bool | false | Delete remote files not present locally |
| `--bandwidth-limit` | int | 0 | Bandwidth cap in KB/s (0 = unlimited) |
| `--include` | string | — | Comma-separated rsync include patterns |
| `--exclude` | string | — | Comma-separated rsync exclude patterns |
| `--preserve-permissions` | bool | false | Preserve file permissions during sync |
| `--verify` | bool | false | Verify file checksums after sync |

**Examples:**

```bash
# Always dry run first
stax files push --dry-run

# Push themes to staging
stax files push --themes-only --environment=staging

# Push plugins to staging
stax files push --plugins-only --environment=staging

# Push to production (prompts for confirmation)
stax files push --environment=production

# Exclude log files and cache
stax files push --exclude="*.log,cache/"
```
