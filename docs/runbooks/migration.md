Step-by-step workflow for migrating a WPEngine site to WordPress VIP using stax migrate.

## Prerequisites

| Requirement | Install |
|-------------|---------|
| phpcs with WordPress-VIP-Go ruleset | `composer global require automattic/vip-coding-standards` |
| VIP CLI | `npm install -g @automattic/vip` |
| Local VIP repo checkout | Clone your VIP Go repo to a directory next to your project |

Verify phpcs is available:

```bash
phpcs -i | grep WordPress-VIP-Go
```

## Configure the project

Add the migration destination to `.stax.yml`:

```yaml
version: 2
provider: wpengine
provider_config:
  install: my-install
  environment: production
migration:
  destination: vip
```

## Step 1: Pull files

```bash
stax migrate pull
```

Downloads plugins, themes, and mu-plugins from WPEngine. Uploads are excluded — VIP manages media separately. Files land in the local `wp-content/` directory.

## Step 2: Export the database

```bash
stax migrate export
```

Exports the WPEngine database with VIP-compatible flags (`--hex-blob`, `--quote-names`, `--default-character-set=utf8mb4`). The file is written to `.stax/<install>-export.sql`.

## Step 3: Run the phpcs audit

```bash
stax migrate audit
```

Scans the pulled plugins and themes against the WordPress-VIP-Go coding standard. The output lists files with violations grouped by severity.

Read the output before proceeding. Errors (not just warnings) must be resolved before VIP will accept the code. Focus on:

- Direct database queries (`$wpdb->query` without prepared statements)
- Use of deprecated functions
- Filesystem writes outside allowed paths

## Step 4: Compare files

```bash
stax migrate compare --vip-repo=../my-vip-repo
```

Compares the local `wp-content/` directory against the VIP repo.

- **MissingFromVIP** — files that exist on WPEngine but are not in the VIP repo. These need to be committed to the VIP repo before migration.
- **MissingFromWPE** — files that exist in the VIP repo but not on WPEngine. This is expected for VIP-specific code.

## Step 5: Resolve gaps

Commit any missing plugins or themes to the VIP repo. Don't skip this step — files not in the VIP repo won't be deployed after migration.

## Step 6: Import

```bash
stax migrate import --sql=.stax/my-install-export.sql
```

Validates the SQL file first, then imports it into the VIP environment. Common errors at this stage:

- **Character set mismatch** — the export uses `utf8mb4`; ensure the VIP database is configured for the same
- **Missing tables** — usually caused by plugins that are active on WPEngine but not yet committed to the VIP repo

## Step 7: Generate report

```bash
stax migrate report --vip-repo=../my-vip-repo
```

Writes a migration summary to `.stax/migration-report.md`. The report covers audit results, file comparison gaps, and import status. Review it with the client before go-live.

## Check status at any time

```bash
stax migrate status
```

Shows which migration steps have run and their outcomes.

## What not to do

- Don't skip `stax migrate audit` even if you're confident the plugins are clean. VIP enforces the standard at deployment, and it's faster to catch issues locally.
- Don't import the database without validating it first. `stax migrate import` runs validation automatically — don't bypass it by importing directly with WP-CLI.
