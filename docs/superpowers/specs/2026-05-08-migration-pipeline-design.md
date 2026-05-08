# Migration Pipeline Design

**Date:** 2026-05-08
**Status:** Approved — pending implementation

## Overview

Stax needs a migration command group (`stax migrate`) that orchestrates moving WordPress sites from one hosting platform to another. The immediate use case is migrating 25 WPEngine sites to WordPress VIP. The design must support future migrations in either direction without locking commands to specific providers.

## Scope

This spec covers four sub-systems. Sub-systems 1 and 3 are in scope for the first implementation cycle. Sub-systems 2 and 4 follow as separate cycles.

| Sub-system | What | Cycle |
|------------|------|-------|
| 1. stax migrate commands | New command group + provider interfaces | This cycle |
| 2. AWS migration infrastructure | EC2, Terraform, S3 for reports | Next cycle |
| 3. Documentation overhaul | README, runbooks, man pages | This cycle |
| 4. Migration runbook | Per-site checklist and QA process | Next cycle |

---

## Sub-system 1: stax migrate commands

### Design principles

`stax migrate` is provider-agnostic. Source and destination are resolved from `.stax.yml`, not hardcoded. Adding a new source or destination means implementing the interfaces, not changing commands.

This follows the same convention used elsewhere in stax: context from config, flags for overrides.

### Configuration

Source is the project's existing `provider:` field — no new config required. Destination is declared once per project:

```yaml
version: 2
provider: wpengine
provider_config:
  install: astronomytn
  environment: production
migration:
  destination: vip
```

Override flag `--destination=vip` is available for one-off use.

### Commands

```
stax migrate pull       Pull files from source provider to local wp-content/
stax migrate export     Export database from source with destination-correct flags
stax migrate audit      Run phpcs compatibility audit against destination ruleset
stax migrate compare    Diff local files against the destination repo
stax migrate import     Push DB/media into destination (wraps VIP CLI)
stax migrate report     Generate combined audit + compare report as markdown
stax migrate status     Show migration state for this project
```

All commands read `migration.destination` from `.stax.yml`. Running any `stax migrate` command in a project without `migration.destination` set produces a clear error directing the user to add it.

### Provider interfaces

New package `pkg/migration/` with two interfaces:

```go
// Source is the platform being migrated away from.
type Source interface {
    PullFiles(opts PullOptions) error
    ExportDatabase(opts ExportOptions) error
}

// Destination is the platform being migrated to.
type Destination interface {
    Audit(localPath string, opts AuditOptions) (*AuditReport, error)
    ValidateDatabase(path string) error
    ImportDatabase(path string, opts ImportOptions) error
    ImportMedia(opts ImportOptions) error
    CompareFiles(localPath string) (*CompareResult, error)
}
```

A `Registry` in `pkg/migration/registry.go` resolves provider names to implementations, matching the pattern in `pkg/provider/`.

### Implementations (this cycle)

**`pkg/migration/providers/wpengine/source.go`** — `WPEngineSource`
- `PullFiles`: delegates to the existing `pkg/files` pull service, excluding uploads by default
- `ExportDatabase`: delegates to `pkg/database` export with VIP-compatible mysqldump flags (`--add-drop-table --hex-blob --no-create-db --quote-names --default-character-set=utf8mb4`)

**`pkg/migration/providers/vip/destination.go`** — `VIPDestination`
- `Audit`: runs `phpcs` with the `WordPress-VIP-Go` ruleset against `plugins/`, `themes/`, `client-mu-plugins/`
- `ValidateDatabase`: runs `vip import validate-sql` against the exported SQL file
- `ImportDatabase`: runs `vip import sql`
- `ImportMedia`: runs `vip import media`
- `CompareFiles`: diffs the downloaded WPEngine wp-content against the local VIP repo structure (plugins, themes, client-mu-plugins) — assumes `stax migrate compare` is run from inside the VIP repo checkout

### Report format

`stax migrate report` produces a markdown file at `.stax/migration-report.md` containing:
- phpcs audit results per plugin/theme with severity counts
- File comparison summary (present in WPEngine, missing from VIP repo; present in VIP repo, missing from WPEngine)
- Database export metadata (table count, size)
- Timestamp and install name

This matches the format of the audit and post-QA reports that VIP produced for the astronomy migration.

### Command location

`cmd/migrate.go` — thin wrapper following existing command patterns. All business logic in `pkg/migration/service.go`.

### Error handling

- Missing `migration.destination` in config: clear error with config snippet showing how to add it
- phpcs not installed: error with install instructions (`composer global require automattic/vip-coding-standards`)
- VIP CLI not installed: error with install instructions (`npm install -g @automattic/vip`)
- Source SSH connection failure: surfaces the underlying SSH error unchanged

---

## Sub-system 3: Documentation overhaul

### Problem

The existing `docs/` directory has 15,000+ lines across 20+ files and 7 subdirectories. Content is duplicated across `USER_GUIDE.md`, `GETTING_STARTED.md`, `QUICK_START.md`, `reference/commands.md`, `technical/COMMANDS.md`, and `COMMAND_REFERENCE.md`. Bug fix summaries and TODO tracking are mixed in with user-facing docs.

### New structure

```
README.md                        Root README: what stax is, install, quick-start, links
docs/
  runbooks/
    getting-started.md           Install, credentials, first project end-to-end
    database.md                  db pull/push, snapshots, restore
    files.md                     files pull/push, exclude patterns, .staxignore
    media-proxy.md               nginx proxy setup and verification
    migration.md                 WPEngine → VIP migration workflow
    multisite.md                 Multisite-specific operations
    troubleshooting.md           Common problems with exact error messages and fixes
  commands/
    overview.md                  All command groups with one-line descriptions
    db.md                        All db subcommands, flags, examples
    files.md                     All files subcommands, flags, examples
    migrate.md                   All migrate subcommands, flags, examples
    wp.md                        wp passthrough with examples
    config.md                    config get/set/validate
  CONTRIBUTING.md                Kept, trimmed to essentials
man/                             Generated by scripts/generate-man.sh
```

### What gets deleted

All of the following are removed:

- `docs/archived/` — bug fix summaries do not belong in user-facing docs
- `docs/explanation/`, `docs/guides/`, `docs/how-to/`, `docs/reference/`, `docs/technical/`, `docs/tutorials/` — consolidated into runbooks and commands
- `docs/USER_GUIDE.md`, `docs/GETTING_STARTED.md`, `docs/QUICK_START.md`, `docs/FAQ.md`, `docs/EXAMPLES.md`, `docs/INSTALLATION.md`, `docs/TROUBLESHOOTING.md`, `docs/WPENGINE.md`, `docs/MEDIA_PROXY.md`, `docs/MULTISITE.md`, `docs/COMMAND_REFERENCE.md`, `docs/SECURITY.md`, `docs/TESTING.md`, `docs/README.md`

### Documentation style (trains-infrastructure pattern)

Each file opens with one sentence describing what it covers. Runbooks are task-oriented — they answer "how do I do X" with copy-paste commands. Command references list every flag with its type, default, and a one-line description. The root README links out rather than duplicating.

Explicit "What not to do" sections are used for operations with non-obvious failure modes.

The `humanizer` skill is applied to all written content to remove AI writing patterns before committing.

### Man pages

`scripts/generate-man.sh` already exists. Man pages are generated from Cobra and installed to the system man path via `make man-install`. They cover every command, subcommand, flag, and argument — the authoritative flag reference. The markdown command docs in `docs/commands/` are the GitHub-readable equivalent.

---

## Out of scope for this cycle

- AWS EC2 infrastructure and Terraform (Sub-system 2)
- Per-site migration runbook and QA checklist (Sub-system 4)
- Support for migration sources other than WPEngine
- Support for migration destinations other than VIP
- Automated parallel execution across 25 sites (requires Sub-system 2)
