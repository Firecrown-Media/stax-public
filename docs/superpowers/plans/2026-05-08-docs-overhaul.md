# Documentation Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the existing 20+ file, 15,000+ line docs directory with a lean, task-oriented structure following the trains-infrastructure documentation pattern.

**Architecture:** Two directories: `docs/runbooks/` for task-oriented how-to guides, and `docs/commands/` for per-command-group reference. Root `README.md` is the entry point — it links out rather than duplicating. Man pages are generated from Cobra via `scripts/generate-man.sh`. All content passes through the `humanizer` skill before commit.

**Tech Stack:** Markdown, existing Cobra man-page generation (`make man`)

---

## Before You Start

The `humanizer` skill removes AI writing patterns from text. **Every document written in this plan must be humanized before committing.** The process for each file is:

1. Write the draft
2. Invoke the `humanizer` skill on the content
3. Replace the file with humanized output
4. Commit

---

## File Map

**Delete entirely:**
- `docs/archived/` (directory)
- `docs/explanation/` (directory)
- `docs/guides/` (directory)
- `docs/how-to/` (directory)
- `docs/reference/` (directory)
- `docs/technical/` (directory)
- `docs/tutorials/` (directory)
- `docs/USER_GUIDE.md`
- `docs/GETTING_STARTED.md`
- `docs/QUICK_START.md`
- `docs/FAQ.md`
- `docs/EXAMPLES.md`
- `docs/INSTALLATION.md`
- `docs/TROUBLESHOOTING.md`
- `docs/WPENGINE.md`
- `docs/MEDIA_PROXY.md`
- `docs/MULTISITE.md`
- `docs/COMMAND_REFERENCE.md`
- `docs/SECURITY.md`
- `docs/TESTING.md`
- `docs/README.md`
- `docs/stax.1.template` (replaced by Cobra-generated man pages)
- `docs/release/` (directory — release process is documented in CLAUDE.md)

**Keep:**
- `docs/CONTRIBUTING.md` (trim to essentials in Task 3)
- `docs/superpowers/` (leave untouched)

**Create:**
- `docs/runbooks/getting-started.md`
- `docs/runbooks/database.md`
- `docs/runbooks/files.md`
- `docs/runbooks/media-proxy.md`
- `docs/runbooks/migration.md`
- `docs/runbooks/multisite.md`
- `docs/runbooks/troubleshooting.md`
- `docs/commands/overview.md`
- `docs/commands/db.md`
- `docs/commands/files.md`
- `docs/commands/migrate.md`
- `docs/commands/wp.md`
- `docs/commands/config.md`

**Rewrite:**
- `README.md` (root — lean navigation file)

---

## Task 1: Delete old docs structure

- [ ] **Step 1: Remove old directories**

```bash
rm -rf docs/archived docs/explanation docs/guides docs/how-to docs/reference docs/technical docs/tutorials docs/release
```

- [ ] **Step 2: Remove old flat files**

```bash
rm -f docs/USER_GUIDE.md docs/GETTING_STARTED.md docs/QUICK_START.md docs/FAQ.md \
      docs/EXAMPLES.md docs/INSTALLATION.md docs/TROUBLESHOOTING.md docs/WPENGINE.md \
      docs/MEDIA_PROXY.md docs/MULTISITE.md docs/COMMAND_REFERENCE.md docs/SECURITY.md \
      docs/TESTING.md docs/README.md docs/stax.1.template
```

- [ ] **Step 3: Verify remaining files**

```bash
find docs/ -type f | sort
```

Expected output — only these files remain:
```
docs/CONTRIBUTING.md
docs/superpowers/specs/2026-05-08-migration-pipeline-design.md
docs/superpowers/plans/2026-05-08-migrate-commands.md
docs/superpowers/plans/2026-05-08-docs-overhaul.md
```

- [ ] **Step 4: Commit**

```bash
git add -A docs/
git commit -m "docs: remove legacy docs structure (replaced by runbooks + commands)"
```

---

## Task 2: Write README.md

The root README is the entry point. It answers: what is this, how do I install it, what can I do with it, where do I go for more. It does not duplicate content from runbooks or command docs.

- [ ] **Step 1: Write the draft**

Replace `README.md` at the repo root with:

```markdown
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
```

- [ ] **Step 2: Apply humanizer**

Invoke the `humanizer` skill on the README content. Replace file with humanized output.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: rewrite README as lean navigation file"
```

---

## Task 3: Trim CONTRIBUTING.md

The existing `docs/CONTRIBUTING.md` likely contains implementation notes and bloat. Trim it to the essentials a contributor needs.

- [ ] **Step 1: Write the trimmed version**

Replace `docs/CONTRIBUTING.md` with content covering only:

1. **One-sentence intro** — what the project is and where the real repo is
2. **Dev setup** — clone, `go mod download`, `make build`
3. **Running tests** — `make test`, `make test-integration`, `make lint`
4. **Branch and commit conventions** — feature branch off main, conventional commits (`feat:`, `fix:`, `chore:`, `docs:`)
5. **Submitting a PR** — what to include, what CI checks run
6. **Release process** — one sentence pointing to release-please automation (no manual tags)

Keep it under 80 lines total.

- [ ] **Step 2: Apply humanizer**

Invoke the `humanizer` skill on the CONTRIBUTING.md content.

- [ ] **Step 3: Commit**

```bash
git add docs/CONTRIBUTING.md
git commit -m "docs: trim CONTRIBUTING.md to essential contributor workflow"
```

---

## Task 4: Write runbooks/getting-started.md

Covers: install, credentials, first project end-to-end. Target audience: a developer joining a team that uses stax.

- [ ] **Step 1: Write the draft**

Create `docs/runbooks/getting-started.md` with:

**Opening sentence:** "Everything you need to set up stax and pull your first WPEngine site to a local DDEV environment."

**Sections:**
1. **Prerequisites** — table: Docker, DDEV (link), macOS 12+, WPEngine account with SSH key
2. **Install stax** — `brew install firecrown-media/stax/stax`, verify with `stax --version`
3. **Store credentials** — `stax setup` (what it asks for: WPEngine username, API token, SSH key path), where credentials live (macOS Keychain)
4. **Initialize a project** — `cd my-ddev-project`, `stax init` (what it creates: `.stax.yml`), show the resulting YAML
5. **Pull the database** — `stax db pull`, what happens (rsync via SSH, import into DDEV, URL replace)
6. **Pull files** — `stax files pull --exclude-uploads`, why exclude uploads (size)
7. **Start the site** — `stax start`, open in browser
8. **What not to do** — don't run `stax db push` to production without a dry run; don't share `.stax.yml` credentials (there aren't any — credentials are Keychain only)

- [ ] **Step 2: Apply humanizer**

Invoke the `humanizer` skill on the content.

- [ ] **Step 3: Commit**

```bash
git add docs/runbooks/getting-started.md
git commit -m "docs: add getting-started runbook"
```

---

## Task 5: Write runbooks/database.md

Covers: db pull, db push, snapshots, restore. Target: developers who need to sync databases day-to-day.

- [ ] **Step 1: Write the draft**

Create `docs/runbooks/database.md` with:

**Opening sentence:** "Pull and push databases between WPEngine and your local DDEV environment."

**Sections:**
1. **Pull a database** — `stax db pull` (what it does: export via SSH, import into DDEV, search-replace URLs)
2. **Pull from staging** — `stax db pull --environment=staging`
3. **Skip URL replacement** — `stax db pull --skip-replace` (when to use: already have the right URLs locally)
4. **Exclude tables** — `stax db pull --exclude-tables=wp_users,wp_usermeta`
5. **Snapshots** — `stax db snapshot create`, `stax db snapshot list`, `stax db snapshot restore <name>`, `stax db snapshot delete <name>`
6. **Auto-snapshot before pull** — how to enable in `.stax.yml`: `snapshots.auto_snapshot_before_pull: true`
7. **Push a database** — `stax db push --dry-run` first, then `stax db push --environment=staging`; warn: production push requires confirmation prompt
8. **What not to do** — never `stax db push` to production without first reviewing `--dry-run` output; never push a database with local test data

- [ ] **Step 2: Apply humanizer**

Invoke the `humanizer` skill.

- [ ] **Step 3: Commit**

```bash
git add docs/runbooks/database.md
git commit -m "docs: add database runbook"
```

---

## Task 6: Write runbooks/files.md

Covers: files pull, files push, exclude patterns, .staxignore.

- [ ] **Step 1: Write the draft**

Create `docs/runbooks/files.md` with:

**Opening sentence:** "Sync wp-content files between WPEngine and your local environment using rsync over SSH."

**Sections:**
1. **Pull all wp-content** — `stax files pull` (excludes uploads by default)
2. **Pull specific directories** — `--themes-only`, `--plugins-only`, `--mu-plugins-only`
3. **Dry run** — `stax files pull --dry-run` (always use this before a destructive pull)
4. **Exclude uploads** — `stax files pull --exclude-uploads` (explicit flag; uploads are large and usually served from WPEngine CDN)
5. **Delete local files not on remote** — `stax files pull --delete` (use carefully — removes anything not in source)
6. **Bandwidth throttle** — `stax files pull --bandwidth-limit=500` (KB/s, useful on slow connections)
7. **Push files** — `stax files push --dry-run`, then `stax files push --environment=staging`
8. **Exclude patterns** — `--exclude="*.log,cache/"` or via `.staxignore` (same syntax as `.gitignore`)
9. **What not to do** — never `stax files push` to production without a dry run first; do not push the `uploads/` directory without a deliberate decision (it can be very large)

- [ ] **Step 2: Apply humanizer**

Invoke the `humanizer` skill.

- [ ] **Step 3: Commit**

```bash
git add docs/runbooks/files.md
git commit -m "docs: add files runbook"
```

---

## Task 7: Write runbooks/media-proxy.md

Covers: nginx remote media proxy setup and verification.

- [ ] **Step 1: Write the draft**

Create `docs/runbooks/media-proxy.md` with:

**Opening sentence:** "Proxy media requests from your local DDEV environment to the WPEngine or BunnyCDN origin so you don't need to download the uploads directory."

**Sections:**
1. **What the proxy does** — DDEV nginx serves `/wp-content/uploads/` from the remote URL; local requests go through to production automatically
2. **Setup** — `stax media setup` (what it writes: `.ddev/nginx_full/media-proxy.conf`)
3. **Verify** — `stax media status`
4. **Required config in .stax.yml** — show the `media.proxy_url` field example
5. **Test it** — open a media URL in the browser after `stax start`; should load without pulling uploads locally
6. **What not to do** — don't set `proxy_url` to a staging environment for a production site (CORS and mixed-content issues)

- [ ] **Step 2: Apply humanizer**

Invoke the `humanizer` skill.

- [ ] **Step 3: Commit**

```bash
git add docs/runbooks/media-proxy.md
git commit -m "docs: add media-proxy runbook"
```

---

## Task 8: Write runbooks/migration.md

Covers: WPEngine → WordPress VIP migration end-to-end workflow using `stax migrate`.

- [ ] **Step 1: Write the draft**

Create `docs/runbooks/migration.md` with:

**Opening sentence:** "Step-by-step workflow for migrating a WPEngine site to WordPress VIP using stax migrate."

**Sections:**
1. **Prerequisites** — phpcs with WordPress-VIP-Go ruleset, VIP CLI, local VIP repo checkout; show install commands
2. **Configure the project** — add `migration.destination: vip` to `.stax.yml`; show the full relevant YAML block
3. **Step 1: Pull files** — `stax migrate pull`; what gets downloaded (plugins, themes, mu-plugins — no uploads); where files land
4. **Step 2: Export the database** — `stax migrate export`; flags added (`--hex-blob`, `--quote-names`, `--default-character-set=utf8mb4`); where the file lands (`.stax/<install>-export.sql`)
5. **Step 3: Run the phpcs audit** — `stax migrate audit`; what it scans; how to read the output; what to fix before proceeding
6. **Step 4: Compare files** — `stax migrate compare --vip-repo=../my-vip-repo`; what MissingFromVIP means (needs to be added to VIP repo); what MissingFromWPE means (VIP-only code that WPEngine doesn't have)
7. **Step 5: Resolve gaps** — commit missing plugins/themes to the VIP repo; don't skip this step
8. **Step 6: Import** — `stax migrate import --sql=.stax/<install>-export.sql`; validation runs first; what errors to expect (character set mismatches, missing tables)
9. **Step 7: Generate report** — `stax migrate report --vip-repo=../my-vip-repo`; what the report contains; where it's written
10. **Check status at any time** — `stax migrate status`
11. **What not to do** — don't skip `stax migrate audit` even if you think the plugins are clean; don't import the database without running `validate-sql` first (stax import does this automatically)

- [ ] **Step 2: Apply humanizer**

Invoke the `humanizer` skill.

- [ ] **Step 3: Commit**

```bash
git add docs/runbooks/migration.md
git commit -m "docs: add migration runbook"
```

---

## Task 9: Write runbooks/multisite.md

Covers: multisite-specific operations.

- [ ] **Step 1: Write the draft**

Create `docs/runbooks/multisite.md` with:

**Opening sentence:** "Working with WordPress multisite networks using stax."

**Sections:**
1. **Supported modes** — subdomain and subdirectory; set in `.stax.yml` under `project.mode`
2. **Init a multisite project** — `stax init` with `project.type: wordpress-multisite`; what the config looks like
3. **Pull the database for a multisite** — `stax db pull` works the same; URL replacement runs for all sites in the network
4. **Network domain config** — the `network.sites[]` array in `.stax.yml`; mapping production domains to local domains
5. **Adding a new site to the network** — `stax wp` to run WP-CLI network commands
6. **What not to do** — don't manually edit `wp_blogs` to add sites — use WP-CLI via `stax wp`; multisite with subdomain mode requires wildcard DNS in DDEV

- [ ] **Step 2: Apply humanizer**

Invoke the `humanizer` skill.

- [ ] **Step 3: Commit**

```bash
git add docs/runbooks/multisite.md
git commit -m "docs: add multisite runbook"
```

---

## Task 10: Write runbooks/troubleshooting.md

Covers: exact error messages with fixes. This is the most important runbook for new users.

- [ ] **Step 1: Write the draft**

Create `docs/runbooks/troubleshooting.md` with:

**Opening sentence:** "Common stax errors with exact error messages and how to fix them."

Cover each of the following (use the exact error message as the section heading):

1. `"DDEV not running"` — check `ddev status`; run `stax start`
2. `"go: command not found"` (make build) — Go via Homebrew not in `/bin/sh` PATH; use `PATH="/opt/homebrew/bin:$PATH" make build`
3. `"failed to get WPEngine credentials"` — run `stax setup`; check Keychain for `com.firecrownmedia.stax`
4. `"rsync: [sender] change_dir failed: No such file or directory"` — remote path issue; run `stax doctor --fix`
5. `"permission denied (publickey)"` (SSH) — SSH key not loaded or wrong path in Keychain; verify with `ssh -T install@ssh.wpengine.net`
6. `"phpcs not found"` (stax migrate audit) — install with `composer global require automattic/vip-coding-standards`
7. `"VIP CLI not found"` (stax migrate import) — install with `npm install -g @automattic/vip`
8. `"migration.destination is not set"` — add `migration.destination: vip` to `.stax.yml`
9. Zero URL replacements after db pull — check that `provider_config.install` matches the WPEngine install name; run `stax doctor`
10. `"nginx error"` after `stax media setup` — config missing `server {}` block; run `stax media setup` again; check `.ddev/nginx_full/media-proxy.conf`

- [ ] **Step 2: Apply humanizer**

Invoke the `humanizer` skill.

- [ ] **Step 3: Commit**

```bash
git add docs/runbooks/troubleshooting.md
git commit -m "docs: add troubleshooting runbook"
```

---

## Task 11: Write commands/overview.md

One-line description of every command group with a link to the detailed reference.

- [ ] **Step 1: Write the draft**

Create `docs/commands/overview.md` with:

**Opening sentence:** "All stax command groups with one-line descriptions."

A table:

| Command | Description |
|---------|-------------|
| `stax setup` | Store WPEngine credentials in the macOS Keychain |
| `stax init` | Initialize a new stax project and create `.stax.yml` |
| `stax start` | Start the DDEV environment |
| `stax stop` | Stop the DDEV environment |
| `stax restart` | Restart the DDEV environment |
| `stax status` | Show DDEV and project status |
| `stax shell` | Open a shell in the DDEV web container |
| `stax db` | Database pull, push, and snapshot operations |
| `stax files` | Sync wp-content files via rsync |
| `stax media` | Configure the nginx remote media proxy |
| `stax migrate` | Orchestrate WPEngine → VIP migration |
| `stax wp` | Run WP-CLI commands in the DDEV container |
| `stax config` | Read and validate `.stax.yml` |
| `stax doctor` | Check prerequisites and auto-fix common issues |
| `stax repo` | Initialize a GitHub repository |
| `stax actions` | Set up GitHub Actions workflows |
| `stax credentials` | Manage credentials in the macOS Keychain |
| `stax version` | Print the stax version |
| `stax man` | Generate man pages |
| `stax completion` | Generate shell completion scripts |

Follow with links to the detailed docs for each command group.

- [ ] **Step 2: Apply humanizer**

- [ ] **Step 3: Commit**

```bash
git add docs/commands/overview.md
git commit -m "docs: add commands overview"
```

---

## Task 12: Write commands/db.md

All `stax db` subcommands with every flag, type, default, and a one-line description.

- [ ] **Step 1: Write the draft**

Create `docs/commands/db.md` with:

**Opening sentence:** "Reference for all `stax db` subcommands and flags."

**Subcommands:**

**`stax db pull`**
Pulls the remote database, imports it locally, and replaces URLs.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--environment` | string | from config | WPEngine environment (production, staging) |
| `--snapshot` | bool | false | Take a local snapshot before importing |
| `--skip-replace` | bool | false | Skip URL search-replace after import |
| `--exclude-tables` | string | — | Comma-separated table names to exclude |
| `--skip-logs` | bool | false | Exclude log tables |
| `--skip-transients` | bool | false | Exclude transient tables |
| `--skip-spam` | bool | false | Exclude spam/trash |
| `--sanitize` | bool | false | Sanitize user data after import |

**`stax db push`**
Pushes the local database to the remote environment.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--environment` | string | from config | Target environment |
| `--dry-run` | bool | false | Preview changes without pushing |
| `--skip-backup` | bool | false | Skip remote backup before import |
| `--skip-replace` | bool | false | Skip URL replacement |

**`stax db snapshot create`** — Create a timestamped local snapshot.
**`stax db snapshot list`** — List all local snapshots.
**`stax db snapshot restore <name>`** — Restore a snapshot by name.
**`stax db snapshot delete <name>`** — Delete a snapshot.

Include 2-3 examples per subcommand.

To get exact flag names and defaults, run:
```bash
./bin/stax db pull --help
./bin/stax db push --help
./bin/stax db snapshot --help
```
Use the actual `--help` output as the authoritative source — this plan provides the structure, not a verbatim copy.

- [ ] **Step 2: Apply humanizer**

- [ ] **Step 3: Commit**

```bash
git add docs/commands/db.md
git commit -m "docs: add db command reference"
```

---

## Task 13: Write commands/files.md

All `stax files` subcommands with every flag.

- [ ] **Step 1: Write the draft**

Create `docs/commands/files.md` with:

**Opening sentence:** "Reference for all `stax files` subcommands and flags."

Use `./bin/stax files pull --help` and `./bin/stax files push --help` as the authoritative flag source.

**Structure:** Two sections (`stax files pull`, `stax files push`), each with a flag table (same columns as db.md) and 2–3 examples.

- [ ] **Step 2: Apply humanizer**

- [ ] **Step 3: Commit**

```bash
git add docs/commands/files.md
git commit -m "docs: add files command reference"
```

---

## Task 14: Write commands/migrate.md

All `stax migrate` subcommands. This depends on the migrate commands being built (Plan 1 — Task 7).

- [ ] **Step 1: Verify migrate commands exist**

```bash
./bin/stax migrate --help
```

If migrate commands are not built yet, complete Plan 1 Task 7 first.

- [ ] **Step 2: Write the draft**

Create `docs/commands/migrate.md` with:

**Opening sentence:** "Reference for all `stax migrate` subcommands and flags."

**Sections for each subcommand:**
- `stax migrate pull` — flag table, examples
- `stax migrate export` — flag table, examples
- `stax migrate audit` — flag table, examples
- `stax migrate compare` — flag table, examples
- `stax migrate import` — flag table, examples
- `stax migrate report` — flag table, examples
- `stax migrate status` — description only (no flags except `--destination`)

Use `./bin/stax migrate <sub> --help` as the authoritative flag source.

- [ ] **Step 3: Apply humanizer**

- [ ] **Step 4: Commit**

```bash
git add docs/commands/migrate.md
git commit -m "docs: add migrate command reference"
```

---

## Task 15: Write commands/wp.md and commands/config.md

- [ ] **Step 1: Write commands/wp.md**

Create `docs/commands/wp.md` with:

**Opening sentence:** "Run WP-CLI commands inside the DDEV container."

Cover:
- Usage: `stax wp <wp-cli-args>`
- How it maps to `ddev exec wp <args>`
- 5–6 examples (cache flush, list plugins, export, create user, search-replace)
- Note: runs inside the container — paths are container paths

Apply humanizer, commit with `git commit -m "docs: add wp command reference"`.

- [ ] **Step 2: Write commands/config.md**

Create `docs/commands/config.md` with:

**Opening sentence:** "Read and validate the `.stax.yml` configuration."

Cover:
- `stax config get <key>` — print a config value by dot-path
- `stax config validate` — check `.stax.yml` is well-formed and all required fields are present
- Full `.stax.yml` annotated example (every field with type, default, description)
- `version: 2` requirement (v1 not supported)

Apply humanizer, commit with `git commit -m "docs: add config command reference"`.

---

## Task 16: Final check

- [ ] **Step 1: Verify structure**

```bash
find docs/ -name "*.md" | sort
```

Expected:
```
docs/CONTRIBUTING.md
docs/commands/config.md
docs/commands/db.md
docs/commands/files.md
docs/commands/migrate.md
docs/commands/overview.md
docs/commands/wp.md
docs/runbooks/database.md
docs/runbooks/files.md
docs/runbooks/getting-started.md
docs/runbooks/media-proxy.md
docs/runbooks/migration.md
docs/runbooks/multisite.md
docs/runbooks/troubleshooting.md
docs/superpowers/plans/2026-05-08-docs-overhaul.md
docs/superpowers/plans/2026-05-08-migrate-commands.md
docs/superpowers/specs/2026-05-08-migration-pipeline-design.md
```

- [ ] **Step 2: Verify README links**

Check every link in README.md points to an existing file:

```bash
grep -o '\[.*\](docs/[^)]*' README.md | grep -o 'docs/[^)]*' | while read f; do
  [ -f "$f" ] && echo "OK: $f" || echo "MISSING: $f"
done
```

Fix any broken links.

- [ ] **Step 3: Build and verify man page generation**

```bash
PATH="/opt/homebrew/bin:$PATH" make build
./bin/stax man --help
```

If a `make man` target exists, run it and verify it completes without errors:

```bash
PATH="/opt/homebrew/bin:$PATH" make man 2>/dev/null && echo "man pages OK" || echo "no man target"
```

- [ ] **Step 4: Final commit**

```bash
git add README.md docs/
git commit -m "docs: complete documentation overhaul (runbooks + command reference)"
```
