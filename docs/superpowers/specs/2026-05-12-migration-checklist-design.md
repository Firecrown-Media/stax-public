# Migration Checklist Design

**Date:** 2026-05-12
**Status:** Approved

## Overview

Sub-system 4 of the WPEngine → VIP migration pipeline. Adds `stax migrate checklist` — a command that generates a per-site migration checklist pre-populated from `.stax.yml` and existing artifacts. The checklist serves as both a pre-migration execution plan and a post-migration record.

The astronomy migration (done by VIP) is the reference. `stax migrate report` already produces the analysis document; the checklist covers the procedural workflow: pre-migration steps, QA, DNS cutover, post-launch validation, and sign-off.

---

## Command

```
stax migrate checklist --domain=<live-domain> [--repo=<path>] [--output=<path>]
```

**Flags:**

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--domain` | yes | — | Live domain, e.g. `astronomytn.com` |
| `--repo` | no | — | Path to local VIP repo checkout (enables commit SHA detection) |
| `--output` | no | `.stax/<install>-checklist.md` | Output path |

**Config inputs** (from `.stax.yml`):
- `provider_config.install` — install name (e.g. `astronomytn`)
- `migration.destination` — destination name (e.g. `vip`)

---

## Artifact Detection

When the command runs, it checks for existing artifacts and pre-populates the checklist accordingly:

| Artifact | Detection | Pre-checks step? |
|----------|-----------|-----------------|
| Report | `.stax/<install>-migration-report.md` exists | Yes |
| SQL dump | `.stax/<install>-export.sql` exists | Yes |
| VIP repo commit | `git -C <repo> log --oneline -1 -- docs/migration-report.md` (only if `--repo` provided) | Yes |

Pre-migration steps with detected artifacts are marked `[x]` in the output. Steps without artifacts are left `[ ]`.

---

## Output Document Structure

### 1. Header

Install name, migration destination, live domain, generated date.

### 2. Artifacts

Paths or links to:
- Migration report (`.stax/<install>-migration-report.md`)
- SQL dump (`.stax/<install>-export.sql`)
- VIP repo commit SHA (from git log, or blank placeholder)

### 3. Pre-migration steps

Checkboxes for each `stax migrate` command in order. Steps with detected artifacts are pre-checked:

- [ ] `stax migrate pull` — files downloaded to `wp-content/`
- [ ] `stax migrate export` — SQL dump at `.stax/<install>-export.sql`
- [ ] `stax migrate audit` — phpcs clean (see report)
- [ ] `stax migrate compare` — gaps resolved in VIP repo
- [ ] `stax migrate import` — database imported to VIP
- [ ] `stax migrate report` — report at `.stax/<install>-migration-report.md`
- [ ] `stax migrate publish` — report and SQL uploaded to S3, committed to VIP repo

### 4. QA checklist

Functional checks against the VIP staging environment, all open:

- [ ] Front page loads
- [ ] Key pages load (about, contact, category pages)
- [ ] Forms work
- [ ] No broken images

### 5. DNS cutover

Steps pre-filled with the live domain:

- [ ] Reduce TTL on `<domain>` to 300s (do this 24–48h before cutover)
- [ ] Confirm TTL reduction has propagated
- [ ] Swap A/CNAME record for `<domain>` to VIP
- [ ] Verify DNS resolution has updated

### 6. Post-launch validation

Same four functional checks as QA, run against the live domain after cutover:

- [ ] Front page loads on `<domain>`
- [ ] Key pages load on `<domain>`
- [ ] Forms work on `<domain>`
- [ ] No broken images on `<domain>`

### 7. Sign-off

Blank operator name and date lines.

---

## Publish Integration

`stax migrate publish` is updated to include the checklist:

1. Upload checklist to S3: `s3://firecrown-assets-378073025324/vip-migration/<install>/checklist.md`
2. Copy checklist to `<vip-repo>/docs/checklist.md`
3. Commit both `docs/migration-report.md` and `docs/checklist.md` in a single git commit

---

## Files

| File | Change |
|------|--------|
| `pkg/migration/checklist.go` | NEW — `ChecklistData` type, markdown template, `GenerateChecklist()` |
| `pkg/migration/checklist_test.go` | NEW — tests for artifact detection, pre-check logic, template output |
| `pkg/migration/service.go` | MODIFY — add `ChecklistOptions`, `Checklist()` function; update `Publish()` |
| `pkg/migration/service_test.go` | MODIFY — add `TestChecklist_*` and updated publish tests |
| `cmd/migrate.go` | MODIFY — add `migrateChecklistCmd`, register flags |
| `docs/runbooks/migration.md` | MODIFY — add checklist step after publish, before WPEngine-specific notes |

---

## Error handling

- Missing `--domain` flag: cobra marks it required, error is automatic
- Missing `migration.destination` in config: existing `requireDestination()` check
- Output directory creation failure: `os.MkdirAll` with wrapped error

No error if artifacts are missing — the checklist generates with open checkboxes and blank artifact fields.
