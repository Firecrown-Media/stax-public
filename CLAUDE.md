# CLAUDE.md

## Superpowers & Workflow

| When | Skill |
|------|-------|
| Before any new feature | `superpowers:brainstorming` |
| Writing a plan | `create_plan` |
| Implementing | `superpowers:test-driven-development` → `implement_plan` |
| Verifying plan coverage | `validate_plan` |
| Debugging / test failures | `superpowers:systematic-debugging`, `debug` |
| Before claiming done / opening PR | `superpowers:verification-before-completion` |
| Deciding how to merge | `superpowers:finishing-a-development-branch` |
| Code review | `superpowers:requesting-code-review` |
| PR description | `describe_pr` |
| Committing | `commit` (no AI attribution) or `ci_commit` |
| Writing docs / README / help text | `humanizer` — removes AI writing patterns |

**Release workflow**: see `.claude/skills/release-workflow.md` — feature branch → conventional commit → PR → release-please → GoReleaser.

## Project Overview

Stax is a WordPress development CLI (Go + Cobra + Viper) integrating WPEngine hosting with DDEV local environments.

**Capabilities**: project init, DB pull/push, remote media proxy, rsync file sync, DB snapshots, multisite support, macOS Keychain credentials.

## Build & Test

```bash
make build              # Build binary (ldflags version)
make test               # Unit tests with race detection
make fmt                # Format with gofmt — CI fails without this, run before every commit
make vet && make lint   # go vet + golangci-lint
make test-integration   # Requires RUN_INTEGRATION_TESTS=true
make test-coverage      # HTML coverage report
make release-snapshot   # Test GoReleaser build (no publish)
```

**Go is installed via Homebrew** at `/opt/homebrew/bin/go`. If `make build` fails with "go: command not found" (Makefile uses `/bin/sh` which doesn't source `.zshrc`), prefix with:

```bash
PATH="/opt/homebrew/bin:$PATH" make build
```

## Architecture

### 3 Repositories

The split exists because GoReleaser requires a public repo to publish releases, but development stays private.

1. **Private dev**: `github.com/Firecrown-Media/stax` — all development; conventional commits trigger release-please
2. **Public mirror**: `github.com/Firecrown-Media/stax-public` — code mirror; GoReleaser builds and publishes releases from here
3. **Homebrew tap**: `github.com/Firecrown-Media/homebrew-stax` — auto-updated by GoReleaser; `brew install firecrown-media/stax/stax`

**Never push tags or releases directly** — automation handles `stax-public` and `homebrew-stax`. Only commit to `stax`.

### Commands

`PersistentPreRunE` in `cmd/root.go` auto-loads `.stax.yml` for most commands. Commands that skip it: `setup`, `version`, `completion`, `man`, `list`, `doctor`, `init`, `start`, `stop`, `restart`, `status`, `wpengine`, `config`.

- **Project**: `init`, `setup`, `start`, `stop`, `restart`, `status`, `shell`
- **Database**: `db pull`, `db push`, `db snapshot list|create|restore|delete`
- **Files**: `files pull` | **Media**: `media setup`, `media status`
- **Repo**: `repo init` | **Actions**: `actions setup`
- **Credentials**: `credentials set|get|delete` | **Diagnostics**: `doctor [--fix]`

### Packages

| Package | Purpose |
|---------|---------|
| `pkg/config` | `.stax.yml` struct, load/validate/save, version migration |
| `pkg/database` | DB pull/push/export service |
| `pkg/files` | rsync pull/push service |
| `pkg/init` | Project init workflow service |
| `pkg/actions` | GitHub Actions workflow template generation |
| `pkg/ddev` | DDEV Manager: `IsRunning()`, `Exec()`, nginx proxy config |
| `pkg/wpengine` | WPEngine API + SSH client, rsync (internal — use via provider interface) |
| `pkg/wordpress` | WP-CLI wrapper, search-replace, multisite operations |
| `pkg/provider` | Provider interface + registry; only WPEngine implemented |
| `pkg/credentials` | macOS Keychain (`com.firecrownmedia.stax`) |
| `pkg/prompts` | TTY detection, `Safe*` input variants |
| `pkg/ui` | `Success/Error/Warning/Info/Section`, `PromptPassword` |
| `pkg/errors` | `DDEVError`, `WPEngineError`, `ConfigError` |
| `pkg/snapshot` | Timestamped DB backups in `.ddev/db_snapshots/` |
| `pkg/diagnostics` | System validation with auto-fix |
| `pkg/git` | Git ops, `.gitignore` generation |
| `pkg/prerequisites` | Dependency checking (Docker, DDEV, Git, gh CLI) |
| `pkg/testutil` | Test helpers and fixtures |
| `pkg/migration` | Migration Source/Destination interfaces + WPEngine→VIP implementations |

## Development Patterns

### Config Schema v2

`.stax.yml` uses `provider:` + `provider_config:` — the old `wpengine:` block is gone. The loader rejects v1 files.

```yaml
version: 2
provider: wpengine
provider_config:
  install: my-install
  environment: production  # production | staging
  ssh_gateway: ""          # optional override
migration:
  destination: vip         # optional; enables stax migrate commands
```

Access values via `providerConfigString(cfg.ProviderConfig, "install")` — defined in each service package. Never use `cfg.WPEngine.*` (removed in v1.0.0).

### Command Structure

`cmd/` files are thin wrappers — flag parsing and calling `pkg/` services only. Business logic lives in `pkg/<feature>/service.go`. Commands pass an authenticated provider resolved via `getAuthenticatedProvider(cfg)` in `cmd/root.go`.

### Interactive Mode

Always use `Safe*` prompt variants — bare prompts hang in CI/CD:

```go
prompts.SafePromptInput("Enter value", "default", true)  // returns default when not a TTY
```

### URL Search-Replace

Always call `VerifySiteURL()` after `SearchReplaceWithOptions()` — search-replace can report success while making zero changes. For multisite: `--network --all-tables --skip-columns=guid --skip-themes --skip-plugins`, then per-site with `--url=<domain>` for subdomain mode.

### Security

- Credentials in macOS Keychain only — never in config files or logs
- Use `exec.Command(cmd, args...)`, never shell string expansion
- Sanitize all file paths; HTTPS for all API calls

### WPEngine SSH Remote Paths

WPEngine SSH home is `/home/wpe-user`. Site files live at `~/sites/{install}/`. Remote rsync paths must be **relative** (no leading `/`) so they resolve from the SSH home:

```
sites/{install}/wp-content/           ✓ correct
/sites/{install}/wp-content/          ✗ wrong — absolute path doesn't exist
```

This is enforced in `pkg/files/service.go` (`resolvePaths`) and `pkg/wpengine/files.go` (`SyncWPContent`). The `SSHSource()` helper on `SSHClient` prepends `install@host:` to build the full rsync SSH URL.

### Migration Command Design

`stax migrate` is **provider-agnostic** — source and destination are resolved from config, not hardcoded.

- **Source** = current project's `provider:` in `.stax.yml` (already there — no new config)
- **Destination** = `migration.destination: vip` in `.stax.yml`
- **Override** = `--destination=vip` flag for one-off use

Interfaces live in `pkg/migration/`:

```go
type Source interface {
    PullFiles(opts PullOptions) error
    ExportDatabase(opts ExportOptions) error
}

type Destination interface {
    Audit(localPath string, opts AuditOptions) (*AuditReport, error)
    ValidateDatabase(path string) error
    ImportDatabase(path string, opts ImportOptions) error
    ImportMedia(opts ImportOptions) error
    CompareFiles(localPath string) (*CompareResult, error)
}
```

Implementations: `pkg/migration/providers/wpengine/` (source), `pkg/migration/providers/vip/` (destination).

Commands: `stax migrate pull`, `export`, `audit`, `compare`, `import`, `report`, `status`.

### Documentation Strategy

All docs follow the **trains-infrastructure pattern**: lean README, task-oriented runbooks, generated man pages. No sprawling ALLCAPS guides.

**Structure:**
```
README.md                    # what stax is, install, 5-command quick-start, links
docs/
  runbooks/                  # task-oriented (getting-started, database, files, media-proxy,
                             #   migration, multisite, troubleshooting)
  commands/                  # reference per command group (db, files, migrate, wp, config)
  CONTRIBUTING.md            # trimmed
man/                         # generated by scripts/generate-man.sh
```

**Rules:**
- Use the `humanizer` skill for ALL documentation — README, runbooks, command refs, man page descriptions
- Each runbook is self-contained with copy-paste commands and explicit "what not to do" sections
- Man pages cover every command, subcommand, and flag (generated from Cobra)
- Do not recreate the old ALLCAPS file structure (`USER_GUIDE.md`, `FAQ.md`, `EXAMPLES.md`, etc.)
- Do not put bug fix summaries, TODO tracking, or internal implementation notes in `docs/`

## Testing

- Unit tests alongside source (`*_test.go`), use `t.TempDir()` for isolation
- Integration tests guard: `os.Getenv("RUN_INTEGRATION_TESTS") != "true"`
- Isolate credential/SSH tests by setting `HOME` to `t.TempDir()`
- Target >80% coverage for `pkg/config`, `pkg/wordpress`, `pkg/ddev`

## Common Issues

| Problem | Cause | Fix |
|---------|-------|-----|
| "DDEV not running" false positive | Path resolution | `getProjectDir()` uses `filepath.Abs()` |
| Silent URL replacement | Missing `--all-tables` | Use `SearchReplaceWithOptions()` + `VerifySiteURL()` |
| Commands hang in CI/CD | Bare `PromptInput()` | Use `Safe*` variants |
| nginx error after media setup | Missing `server {}` wrapper | `GenerateMediaProxyConfig()` wraps all directives |
| Environment mismatch in `.stax.yml` | Manual edit / setup error | `stax doctor --fix` corrects via WPEngine API |
| Zero replacements on db pull | URL pattern mismatch | Default to `.wpengine.com`; `getPossibleWPEngineURLs()` provides fallbacks |
| rsync "No such file or directory" | Absolute remote path | Remote paths must be relative: `sites/{install}/wp-content/` not `/sites/...` |
| rsync treats remote path as local | Missing SSH host prefix | `SSHSource()` on `SSHClient` builds `install@host:path` — use it in provider `SyncFiles` |
| `make build` "go: command not found" | Homebrew Go not in `/bin/sh` PATH | Use `PATH="/opt/homebrew/bin:$PATH" make build` |
