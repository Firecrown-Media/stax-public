Step-by-step workflow for migrating a WPEngine site to WordPress VIP using stax migrate.

## Prerequisites

| Requirement | Install |
|-------------|---------|
| phpcs with WordPress-VIP-Go ruleset | `composer global require --dev 'dealerdirect/phpcodesniffer-composer-installer:^1.0' 'wp-coding-standards/wpcs:^3.0' 'automattic/vipwpcs:^3.0'`<br/>then `composer global config --no-plugins allow-plugins.dealerdirect/phpcodesniffer-composer-installer true && composer global install` |
| VIP CLI | `npm install -g @automattic/vip` |
| Local VIP repo checkout | Clone your VIP Go repo to a directory next to your project |

Verify phpcs is available:

```bash
phpcs -i | grep WordPress-VIP-Go
```

## About the VIP destination repo

Each VIP application has its own GitHub repository under the `wpcomvip` organization (for example `wpcomvip/boatingmag-com`). VIP creates the repo when they provision the app, and operators get access through their VIP account. The migration loop assumes you have a local clone of that repo on the same machine where you are running `stax`.

### Branch model

VIP maps branches one-to-one to environments. Pushing to a branch deploys to the matching environment.

| Branch | Environment | Used during migration |
|--------|-------------|----------------------|
| `production` | live VIP production environment | Final destination after QA |
| `preprod` | optional preview environment | Often skipped during initial migration |
| `develop` | development environment | Usually skipped; VIP `develop` envs are not always provisioned |

Most migrations push directly to `production` after operator review and VIP CI passes. Do not push tags or releases manually — the deploy pipeline is the merge.

### Required directory structure

The VIP repo expects a specific layout. `stax migrate compare` and `stax migrate publish` assume these paths.

| Path | Purpose | WPEngine equivalent |
|------|---------|--------------------|
| `plugins/` | Regular WordPress plugins. One subdirectory per plugin. | `wp-content/plugins/` |
| `themes/` | WordPress themes. Active parent and child themes both live here. | `wp-content/themes/` |
| `client-mu-plugins/` | Must-use plugins owned by the site (loaded automatically every request). Use for custom drop-ins and integration shims. | `wp-content/mu-plugins/` minus the WPEngine-specific entries |
| `vip-config/vip-config.php` | Environment-aware PHP configuration loaded by VIP before WordPress boots. Use for env-conditional constants and feature flags. | No equivalent — replaces `wp-config.php` overrides that WPEngine sites put in `mu-plugins` or in dashboard-only fields |
| `images/` (optional) | Site favicons and logos served from the VIP edge | — |
| `languages/` (optional) | Custom translation files | `wp-content/languages/` |
| `private/` (optional) | Non-web-accessible files (server-side only) | — |

Do not create or commit a top-level `mu-plugins/` directory in the VIP repo. VIP manages its own platform mu-plugins (`a8c-files`, `vip-cache-manager`, `query-monitor`, `vip-support`, and so on) and serves them automatically. Custom must-use plugins go in `client-mu-plugins/`.

### Mapping WPEngine layout to VIP layout

`stax migrate pull` lands plugins, themes, and mu-plugins at `<project>/wp-content/`. The mapping you need to apply when committing to the VIP repo:

| Source (WPEngine) | Destination (VIP repo) | Notes |
|-------------------|------------------------|-------|
| `wp-content/plugins/*` | `plugins/*` | Drop `fastly`, `unfiltered-mu`, `wpe-site-migration` — covered in the WPEngine-specific section below |
| `wp-content/themes/*` | `themes/*` | Drop unused themes such as `twentytwentyfive` and inactive starter themes |
| `wp-content/mu-plugins/*` (custom only) | `client-mu-plugins/*` | Drop all WPEngine-provided mu-plugins — covered in the WPEngine-specific section below |
| `wp-content/uploads/*` | (not in repo) | Media goes through `vip import media`, never the repo |

### Composer and build tooling

VIP repos may include a top-level `composer.json` and use Composer to manage third-party plugins. If the repo uses Composer, the `vendor/` directory is committed (VIP's deploy pipeline does not run `composer install` server-side by default — this is changing, but check per-app). If you add a plugin via Composer, commit both the `composer.json` change and the resulting `vendor/` and plugin directory changes in the same commit.

If the repo does not use Composer, drop plugin folders into `plugins/` directly, same as you would for a non-Composer WordPress install.

### Branch protection and CI

VIP applies branch protection on `production`. Pull requests typically require:

- Passing PHP_CodeSniffer with the `WordPress-VIP-Go` ruleset (the same audit `stax migrate audit` runs locally)
- Passing PHPCompatibility checks for the PHP version VIP serves
- VIP-side reviewer approval for larger changes

Run `stax migrate audit` and resolve errors before pushing. Warnings are advisory but worth a pass.

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

Downloads plugins, themes, and mu-plugins from WPEngine. Uploads are excluded; media has its own step (see Step 6b below). Files land in the local `wp-content/` directory.

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

Two things to know about the verification commands:

- VIP CLI prompts for production confirmation (`y/N`) on every `wp` call. The prompt is drawn in raw terminal mode and ignores piped stdin. To drive it from a non-interactive shell use `ssh -tt` and pipe `y`:

  ```bash
  ssh -tt user@host "echo y | vip @<app>.<env> -- wp option get siteurl"
  ```

- If a plugin throws a fatal during WordPress bootstrap, wp-cli cannot run. Use `--skip-plugins=<plugin>` to bypass the broken one and deactivate it:

  ```bash
  vip @<app>.<env> -- wp --skip-plugins=<plugin> plugin deactivate <plugin>
  ```

## Step 6a: Verify the site renders

The most common post-import symptom is a "fresh WordPress install" appearance at `<app>.go-vip.net`. The page renders unstyled because something is throwing a fatal during bootstrap and VIP is serving the generic critical-error page.

Diagnose by running any wp-cli command. The fatal will print in the output:

```bash
vip @<app>.<env> -- wp option get siteurl
```

The almost-always culprit is a plugin that's active in the database but whose code (or a code dependency) isn't deployed to the VIP repo. The DB references it, WP tries to load it, autoload fails. Find the plugin in the trace and deactivate it with `--skip-plugins=` as shown above.

After the fatal is cleared, the site will return HTTP 301 from `<app>.go-vip.net` to the production domain because `siteurl`/`home` still point there. To preview at the VIP URL, search-replace the domain:

```bash
vip @<app>.<env> -- wp search-replace \
  '<prod-domain>' '<app>.go-vip.net' \
  --all-tables --skip-columns=guid --skip-themes --skip-plugins
```

Then flush both cache layers (wp-cli can't reach VIP's edge):

```bash
vip @<app>.<env> -- wp cache flush
vip @<app>.<env> cache purge-url https://<app>.go-vip.net/
```

## Step 6b: Import media

`stax migrate` does not pull or import media. Use the project sync to land uploads on the migration box, then hand the archive to `vip import media`.

```bash
# 1. Pull uploads from WPEngine. The --include/--exclude flag ordering in
# `stax files pull` excludes before includes, so a pure include pattern
# transfers nothing. Use excludes-only to drop everything except uploads:
stax files pull \
  --exclude='themes/***,plugins/***,mu-plugins/***,languages/***,upgrade/***'

# 2. Build the archive. Drop plugin work directories that aren't real media —
# VIP Files rejects PHP, CSV, XML, and similar non-media types:
tar czf <install>-uploads.tar.gz \
  --exclude='uploads/wpallimport' \
  --exclude='uploads/xml' \
  --exclude='uploads/wpseo-redirects' \
  -C wp-content uploads

# 3. Import to VIP. --force skips the production confirmation prompt;
# --saveErrorLog captures the per-file error report locally:
vip @<app>.<env> import media <install>-uploads.tar.gz \
  --force \
  --saveErrorLog=<install>-import-errors.json \
  --exportFileErrorsToJson
```

Things to expect:

- The error log will contain one entry per intermediate (resized) image (`150x150.jpg`, `300x225.jpg`, etc.). These are not errors — VIP regenerates intermediates from the original. The `Skipping intermediate image because original file is being imported` message is informational. Add `--importIntermediateImages` if your theme requires the exact pre-existing sizes, but the default behavior is what you want.
- Files that VIP's allowlist rejects (PHP, CSS, JS, ZIP, CSV, XML, HTML, LESS, JSON, .htaccess) show as `404 - not found` in the error log. They are skipped, not import-aborting.
- The import is a long upload followed by a server-side job. The CLI shows a progress meter at the upload phase, then `Status: RUNNING` while VIP processes, then `Status: COMPLETED ✓`. If you `^C` the local CLI mid-upload, you cancel only the local progress watcher — once the upload finishes, the server-side job runs to completion regardless.

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

**WPEngine MU plugins** (in `wp-content/mu-plugins/`) — always removed on VIP:

| Plugin | Reason |
|--------|--------|
| `wpe-cache-plugin` | Caching handled by WPVIP infrastructure |
| `wpe-wp-sign-on-plugin` | WPEngine-specific, VIP incompatible |
| `wpe-update-source-selector` | WPEngine-specific, VIP incompatible |
| `force-strong-passwords` | Functionality managed by WPVIP |
| `slt-force-strong-passwords` | WPEngine-specific, VIP incompatible |
| `wpengine-security-auditor` | WPEngine-specific, VIP incompatible |
| `wpengine-common` | WPEngine bootstrap shim, VIP incompatible |
| `wordpress-plugin` | WPEngine compatibility shim |
| `elementor-safe-mode.php` | WPEngine-injected; VIP runs its own safe-mode controls |

**WPEngine plugins** (in `wp-content/plugins/`) — also remove from the VIP repo if present:

| Plugin | Reason |
|--------|--------|
| `fastly` | VIP runs its own edge layer; the WPEngine Fastly plugin will conflict |
| `unfiltered-mu` | WPEngine mu-plugin loader, not needed on VIP |
| `wpe-site-migration` | WPEngine migration tool, no value on VIP |

If a plugin in this list is still in the database's `active_plugins` option after import, deactivate it explicitly. Otherwise WordPress will look for the plugin file on every request and log a warning each time. The fastest path is wp-cli:

```bash
vip @<app>.<env> -- wp plugin deactivate fastly unfiltered-mu wpe-site-migration
```

**Table prefix** — WPEngine often uses `wp_2_` instead of `wp_`. The report detects this. When it's present, search-replace must cover both the prefix conversion (`wp_2_` → `wp_`) and URL changes.

**PHP files in uploads** — VIP does not allow PHP files as media. The report lists any PHP files found in `wp-content/uploads/`. Remove them before or after migration.

**Media (uploads)** — `stax migrate pull` deliberately excludes `wp-content/uploads/`. There is no `stax migrate media` command (an open gap; see Step 6b above for the manual flow using `stax files pull` plus `vip import media`). Document the approach used in the Operator Notes section of the migration report.

**Third-party domain whitelisting** — services like Google reCAPTCHA, ad networks, and SSO providers need domain whitelisting after DNS cutover. This is not a migration blocker but must be documented in Operator Notes and actioned post-launch.

## What not to do

- Don't skip `stax migrate audit` even if you're confident the plugins are clean. VIP enforces the standard at deployment, and it's faster to catch issues locally.
- Don't run `vip import sql` via SSH pipe or any non-TTY context without setting `VIP_TOKEN`. The VIP CLI exits 0 silently on auth failure; the import looks fine but nothing happens.
- Don't run `stax migrate publish` without filling in the Operator Notes section first. The report is a permanent record.
- Don't trust `wp plugin list` alone after import. The output reflects what the DB calls "active" but does not prove the plugin code is deployed to the VIP repo. Always diff `wp-content/plugins/` from WPEngine against `<vip-repo>/plugins/` and resolve gaps before assuming the migration is clean. `stax migrate compare` flags this.
- Don't pipe answers to `vip wp` commands through plain SSH. The VIP CLI uses raw terminal mode for its production confirmation prompt and ignores piped stdin. Use `ssh -tt` and pipe `y`, or rely on `--force` where the subcommand supports it.
- Don't include plugin work directories in `vip import media`. `wp-content/uploads/wpallimport/`, `uploads/xml/`, and `uploads/wpseo-redirects/` contain CSV, XML, and dotfile data that VIP Files rejects. Exclude them at the tar step so the error log only contains real issues. The same goes for any `.php` silence-is-golden index files left over from plugins; they all fail validation.
- Don't accept `Skipping intermediate image because original file is being imported` as an error. It's the normal VIP behavior with `--importIntermediateImages` off, and the right behavior in nearly every case.
