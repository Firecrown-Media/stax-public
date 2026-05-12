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
stax migrate compare --repo=../my-vip-repo
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
stax migrate report --repo=../my-vip-repo
```

Generates a comprehensive VIP-style migration document at `.stax/<install>-migration-report.md`. The report covers plugin/theme compatibility, WPEngine MU plugins removed, database analysis, media stats, file comparison gaps, and known issues. Review it before filling in Operator Notes.

## Step 8: Generate checklist

```bash
stax migrate checklist --domain=<live-domain>
```

Generates `.stax/<install>-checklist.md` — a per-site migration checklist pre-populated with artifact status and site-specific details. The checklist tracks all migration steps, QA sign-off, DNS cutover, and post-launch validation.

Re-run after any step completes to update the pre-checked items. Pass `--repo` to include the VIP publish commit SHA in the output:

```bash
stax migrate checklist --domain=<live-domain> --repo=../my-vip-repo
```

## Check status at any time

```bash
stax migrate status
```

Shows which migration steps have run and their outcomes.

## Step 9: Review and annotate the report

Open `.stax/<install>-migration-report.md` and fill in the **Operator Notes** section:

- Which incompatible plugins were addressed and how (updated, deactivated, or accepted as-is)
- Any site-specific issues encountered during the migration steps
- Confirmation of the WPEngine MU plugin removal list

Don't skip this step. The report goes to the VIP repo and is the permanent migration record.

## Step 10: Publish

```bash
stax migrate publish --repo=../my-vip-repo
```

Uploads the report, SQL export, and checklist (when present) to S3, copies the report and checklist to `<vip-repo>/docs/`, commits, and pushes.

## Check status at any time

```bash
stax migrate status
```

Shows which migration steps have run and their outcomes.

## WPEngine-specific considerations

These apply to every site and are auto-flagged in the report.

**WPEngine MU plugins** — always removed on VIP:

| Plugin | Reason |
|--------|--------|
| `wpe-cache-plugin` | Caching handled by WPVIP infrastructure |
| `wpe-wp-sign-on-plugin` | WPEngine-specific — VIP incompatible |
| `wpe-update-source-selector` | WPEngine-specific — VIP incompatible |
| `force-strong-passwords` | Functionality managed by WPVIP |
| `slt-force-strong-passwords` | WPEngine-specific — VIP incompatible |
| `wpengine-security-auditor` | WPEngine-specific — VIP incompatible |

**Table prefix** — WPEngine often uses `wp_2_` instead of `wp_`. The report detects this. When it's present, search-replace must cover both the prefix conversion (`wp_2_` → `wp_`) and URL changes.

**PHP files in uploads** — VIP does not allow PHP files as media. The report lists any PHP files found in `wp-content/uploads/`. Remove them before or after migration.

**Media (uploads)** — `stax migrate pull` deliberately excludes `wp-content/uploads/`. VIP manages media separately via `vip import media`, which pulls from a URL rather than a file transfer. The standard approach:

1. Keep the WPEngine site live and pointing at the existing domain during migration.
2. After DNS cutover, run `vip import media` to pull uploads directly from the old site's URL.
3. Alternatively, provide VIP with an S3 bucket or SFTP drop of the uploads directory.

`stax` does not orchestrate this — VIP's own tooling handles it. Document the chosen approach in the Operator Notes section of the migration report.

**Third-party domain whitelisting** — services like Google reCAPTCHA, ad networks, and SSO providers need domain whitelisting after DNS cutover. This is not a migration blocker but must be documented in Operator Notes and actioned post-launch.

## What not to do

- Don't skip `stax migrate audit` even if you're confident the plugins are clean. VIP enforces the standard at deployment, and it's faster to catch issues locally.
- Don't import the database without validating it first. `stax migrate import` runs validation automatically — don't bypass it by importing directly with WP-CLI.
- Don't run `stax migrate publish` without filling in the Operator Notes section first. The report is a permanent record.
