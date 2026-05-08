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
