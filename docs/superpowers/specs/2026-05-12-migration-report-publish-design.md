# Migration Report & Publish Design

**Goal:** Enrich `stax migrate report` to generate a comprehensive VIP-style migration document automatically from prior step artifacts, and add `stax migrate publish` to distribute it to S3 and the site's VIP GitHub repo.

**Architecture:** Enrich report generation in `pkg/migration/providers/vip/`; add `publish` command to `cmd/migrate.go` and a `Publish()` method to the migration service. Rename `--vip-repo` flag to `--repo` across all migrate subcommands.

**Tech Stack:** Go, AWS S3 (via aws-cli on migration instance), Git

**Reference:** VIP astronomy migration outputs — plugin compatibility spreadsheet and post-QA analysis doc — define the expected report format.

---

## Context

This is Sub-system 4 of the Firecrown WPEngine → WordPress VIP migration pipeline. Sub-systems 1–3 are complete. This sub-system produces the per-site migration record that:

1. Documents what was found, what was changed, and what needs follow-up
2. Serves as the artifact Firecrown and VIP reference post-migration
3. Lives in the site's VIP GitHub repo `docs/` folder (synced to VIP production) and in S3

Migrations run one site at a time on the `vip-migration-instance` EC2 workstation.

---

## What We're Building

### 1. Enriched `stax migrate report`

Reads `.stax/` artifacts written by prior steps and generates `<install>-migration-report.md` with the following sections.

#### Section 1 — Plugin Compatibility

Table auto-populated from phpcs audit results:

| Name | Status | Current Version | VIP Compatible | Compatibility Issues | Notes |
|------|--------|----------------|----------------|----------------------|-------|

- VIP Compatible = Yes if no phpcs errors for that plugin, No otherwise
- Flags known premium plugins (ACF Pro, Gravity Forms, Relevanssi Premium) as "manual update required" in Notes
- Inactive plugins are included

#### Section 2 — WPEngine Plugins Removed

Auto-detects the 6 known WPEngine MU plugins in the pulled `wp-content/mu-plugins/` and lists them:

| Plugin | Reason |
|--------|--------|
| wpe-cache-plugin | Caching handled by WPVIP infrastructure |
| wpe-wp-sign-on-plugin | WPEngine-specific — VIP incompatible |
| wpe-update-source-selector | WPEngine-specific — VIP incompatible |
| force-strong-passwords | Functionality managed by WPVIP |
| slt-force-strong-passwords | WPEngine-specific — VIP incompatible |
| wpengine-security-auditor | WPEngine-specific — VIP incompatible |

#### Section 3 — Theme Compatibility

Same table format as plugins, from audit results.

#### Section 4 — Database Analysis

- **Table prefix** — detected prefix (e.g., `wp_2_`) and the target prefix (`wp_`); flags if conversion required
- **Collation** — any incompatible collations detected in the SQL export
- **Unused tables** — empty tables present in the export (e.g., leftover WooCommerce tables)

#### Section 5 — Media Migration

- Total upload size and file count
- Count and list of excluded files (PHP files not allowed as media on VIP)

#### Section 6 — Known Issues / Operator Notes

Pre-populated with standard VIP post-launch caveats:

- Third-party services (reCAPTCHA, ad networks, SSO providers, Firecrown dashboard) need domain whitelisting after DNS cutover — not a migration blocker
- Any constants required by plugins (e.g., `KSERVE_API_KEY`) must be set via WPVIP environment variables in `vip-config.php`

Blank **Operator Notes** field for site-specific observations. Operator fills this in before running `publish`.

#### Section 7 — Summary

Auto-generated checklist of completed migration steps with outcomes and overall status.

---

### 2. `stax migrate publish`

```bash
stax migrate publish --repo=../my-vip-repo
```

Steps in order:

1. Verifies `.stax/<install>-migration-report.md` exists — errors if `stax migrate report` hasn't been run
2. Uploads to S3:
   - `.stax/<install>-migration-report.md` → `s3://firecrown-assets-378073025324/vip-migration/<install>/migration-report.md`
   - `.stax/<install>-export.sql` → `s3://firecrown-assets-378073025324/vip-migration/<install>/<install>-export.sql`
3. Copies report to `<repo>/docs/migration-report.md`
4. Git commits and pushes: `docs: add migration report for <install>`
5. Prints confirmation: S3 URLs + git commit SHA

Fails clearly if:
- Report file missing (run `stax migrate report` first)
- VIP repo has uncommitted changes
- Git push fails (SSH key not configured — see README)
- S3 upload fails (IAM role issue)

Uses IAM instance profile for S3 (no credentials file needed). Uses the SSH key set up during first-launch instance setup for the git push.

---

### 3. Flag rename: `--vip-repo` → `--repo`

The `--vip-repo` flag on `stax migrate compare`, `stax migrate report`, and the new `stax migrate publish` is renamed to `--repo`. Applies consistently across all three commands.

---

### 4. `docs/runbooks/migration.md` updates

Add to the existing 7-step workflow:

**Step 8: Review and annotate the report**

Open `.stax/<install>-migration-report.md` and fill in the Operator Notes section:
- Plugin decisions (which incompatible plugins were addressed and how)
- Site-specific issues encountered
- Confirmation of WPEngine MU plugin removal

**Step 9: Publish**

```bash
stax migrate publish --repo=../my-vip-repo
```

**New "WPEngine-specific considerations" section:**

- The 6 WPEngine MU plugins are always removed — auto-flagged in the report
- Table prefix: WPEngine often uses `wp_2_` instead of `wp_`; search-replace must cover both prefix conversion and URL changes
- PHP files in uploads are excluded from media migration — VIP doesn't allow PHP as a media type
- Third-party service domain whitelisting is required after DNS cutover but is not a migration blocker

---

## File Map

**Modify:**
- `pkg/migration/providers/vip/destination.go` — enrich `GenerateReport()` with full VIP-style sections
- `pkg/migration/providers/vip/destination_test.go` — tests for enriched report content
- `pkg/migration/service.go` — add `Publish(opts PublishOptions) error` method
- `pkg/migration/service_test.go` — tests for Publish
- `cmd/migrate.go` — add `migratePublishCmd`; rename `--vip-repo` to `--repo` on compare, report, publish
- `docs/runbooks/migration.md` — add Steps 8–9 and WPEngine-specific considerations section

---

## Operator Workflow (complete 9-step)

```bash
stax migrate pull
stax migrate export
stax migrate audit
stax migrate compare --repo=../my-vip-repo
# resolve gaps — commit missing files to VIP repo
stax migrate import --sql=.stax/<install>-export.sql
stax migrate report --repo=../my-vip-repo
# open .stax/<install>-migration-report.md, fill in Operator Notes
stax migrate publish --repo=../my-vip-repo
```
