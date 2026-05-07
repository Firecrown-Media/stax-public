# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Spec-Driven Development

This project uses [spec-kit](https://github.com/github/spec-kit) for spec-driven development. Key files:

- **`.speckit/constitution.md`** - Project principles, standards, and architectural decisions
- **`.speckit/specs/`** - Feature specifications (the "what" and "why")
- **`.speckit/plans/`** - Implementation plans (the "how")
- **`.speckit/tasks/`** - Generated task lists

When making significant changes, consult the constitution and relevant specs/plans first.

## Project Overview

Stax is a WordPress development CLI tool that integrates with WPEngine hosting and DDEV local development environments. It's written in Go 1.24 using the Cobra CLI framework and Viper for configuration management.

**Key Capabilities**:
- Initialize WordPress projects with DDEV configuration
- Sync databases between WPEngine and local environments (pull/push)
- Remote media proxying (on-demand fetching from CDN/WPEngine)
- File synchronization via rsync over SSH
- Database snapshots for safe restore operations
- Multisite WordPress support (subdomain and subdirectory modes)
- Secure credential storage via macOS Keychain

## Development Commands

### Build and Test
```bash
make build              # Build binary with version info (uses ldflags)
make test               # Run unit tests with race detection
make test-unit          # Same as `make test`
make test-integration   # Run integration tests (requires RUN_INTEGRATION_TESTS=true)
make test-coverage      # Generate HTML coverage report in coverage.html

# Run a single test file
go test -v -race ./pkg/config/...

# Run a specific test function
go test -v -race ./pkg/config/... -run TestFunctionName

# Code quality
make fmt                # Format code with gofmt
make vet                # Run go vet
make lint               # Run golangci-lint (config in .golangci.yml)

# Installation
make install            # Install to /usr/local/bin with man page

# Release testing
make release-snapshot   # Test release build with GoReleaser (no publish)
```

### Running the Binary
```bash
go run main.go <command>     # During development (slower, no version info)
./stax <command>             # After `make build` (faster, has version info)
```

## Architecture Overview

### Hybrid Repository Model (3 Repositories)

1. **Development Repository** (private): `github.com/Firecrown-Media/stax`
   - Main development happens here
   - Conventional commits trigger release-please

2. **Public Mirror** (public): `github.com/Firecrown-Media/stax-public`
   - Release-please creates tags here
   - GoReleaser builds and publishes releases

3. **Homebrew Tap**: `github.com/Firecrown-Media/homebrew-stax`
   - Formula updated automatically by GoReleaser
   - Users install via: `brew install firecrown-media/stax/stax`

### Release Workflow

```
Conventional Commit → Release-Please PR → Merge → Tag in Public Mirror
→ GoReleaser Build → GitHub Releases + Homebrew Formula Update
```

**Important**: Version is set via git tags using ldflags in Makefile:
```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -X github.com/firecrown-media/stax/cmd.Version=$(VERSION)
```

### CLI Command Structure

The CLI is built with Cobra. All commands are defined in `cmd/` directory:

**Root Command** (`cmd/root.go`):
- Global flags: `--config`, `--verbose`, `--debug`, `--quiet`, `--no-color`, `--project-dir`
- `PersistentPreRunE`: Loads `.stax.yml` config for most commands (skips for setup, version, init)
- Uses `pkg/ui` for consistent terminal output

**Command Categories**:
- **Project**: `init`, `setup`, `start`, `stop`, `restart`, `status`, `shell`
- **Database**: `db pull`, `db push`, `db snapshot list|create|restore|delete`
- **Files**: `files pull` (rsync from WPEngine)
- **Media**: `media setup`, `media status` (remote media proxy)
- **Credentials**: `credentials set|get|delete` (macOS Keychain)
- **Diagnostics**: `diagnose` (system validation)

### Package Structure

#### Core Packages

**`pkg/config`**: Configuration management
- `Config` struct maps to `.stax.yml` YAML structure (schema version 2)
- Top-level `provider` string + `provider_config map[string]any` for provider-specific settings (replaces old `wpengine` block)
- Nested configs: `ProjectConfig`, `NetworkConfig`, `DDEVConfig`, etc.
- `manager.go`: Load, validate, save configuration files
- `validate.go`: Schema version and required-field validation

**`pkg/ddev`**: DDEV integration
- `Manager` type wraps DDEV CLI operations
- `IsRunning()`: Check container status
- `Exec()`: Run commands in DDEV containers
- `nginx.go`: Generate media proxy nginx configs
- `GenerateMediaProxyConfig()`: Creates `.ddev/nginx_full/media-proxy.conf`

**`pkg/wpengine`**: WPEngine SSH operations
- `SSHClient` for database exports and rsync file transfers
- Handles SSH gateway connections
- `rsync.go`: File synchronization helpers

**`pkg/database`**: Database service layer
- `service.go`: Pull/push/export logic extracted from cmd/ — all DB business logic lives here

**`pkg/files`**: File sync service layer
- `service.go`: `Pull()` and `Push()` — rsync sync logic extracted from cmd/

**`pkg/init`**: Project initialization service
- `service.go`: Init workflow logic extracted from cmd/

**`pkg/actions`**: Build/deploy actions service
- `service.go`: Build and deploy action logic extracted from cmd/

**`pkg/wordpress`**: WordPress/WP-CLI operations
- `CLI` type wraps WP-CLI commands (via `ddev wp`)
- `SearchReplace()`: URL replacement with verification
- `SearchReplaceWithOptions()`: Advanced search-replace with `--all-tables`, `--network`
- `VerifySiteURL()`: Critical validation after URL changes
- Multisite support: network-wide and per-site operations

**`pkg/credentials`**: Secure credential storage
- Uses macOS Keychain via `github.com/keybase/go-keychain`
- Service name: `com.firecrownmedia.stax`
- Stores WPEngine SSH passwords and API tokens
- `helpers.go`: High-level Get/Set/Delete operations

**`pkg/prompts`**: Interactive user input
- `IsInteractive()`: TTY detection (important for CI/CD)
- `Safe*` functions: Return defaults in non-interactive mode
- Validators: `WPEngineInstallPrompt()`, `DomainPrompt()`, etc.
- **Critical**: Always use `Safe*` variants to avoid hanging in non-interactive contexts

**`pkg/ui`**: Terminal output formatting
- `Success()`, `Error()`, `Warning()`, `Info()`: Colored output
- `Section()`: Visual separators
- Respects `--no-color` and `--quiet` flags

**`pkg/errors`**: Enhanced error types
- `DDEVError`, `WPEngineError`, `ConfigError`: Context-rich errors
- `enhanced.go`: Error codes and catalog for user-friendly messages

**`pkg/snapshot`**: Database snapshot management
- Create timestamped backups before dangerous operations
- List, restore, and delete snapshots
- Stored in `.ddev/db_snapshots/`

#### Supporting Packages

- **`pkg/build`**: Build automation and deployment
- **`pkg/diagnostics`**: System requirement validation
- **`pkg/git`**: Git operations wrapper
- **`pkg/provider`**: Multi-provider abstraction (WPEngine, others)
- **`pkg/security`**: Security scanning and validation
- **`pkg/system`**: System utilities and detection
- **`pkg/testutil`**: Test helpers and fixtures

## Key Development Patterns

### Adding a New Command

1. Create `cmd/<command>.go`:
```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/firecrown-media/stax/pkg/config"
    "github.com/firecrown-media/stax/pkg/ui"
)

var myCmd = &cobra.Command{
    Use:   "my-command",
    Short: "Brief description",
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)
    myCmd.Flags().StringVar(&myFlag, "my-flag", "", "Flag description")
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Get config (loaded by PersistentPreRunE in root.go)
    cfg, err := config.Load(getProjectDir())
    if err != nil {
        return err
    }

    ui.Section("My Command")
    // ... implementation
    ui.Success("Done!")
    return nil
}
```

2. Update `cmd/root.go` if the command should skip config loading

### Configuration Loading

**Automatic**: Most commands automatically load `.stax.yml` via `rootCmd.PersistentPreRunE`

**Skip Config Loading**: Commands in `skipConfigCommands` slice (in `cmd/root.go`) don't require `.stax.yml`:
- `setup`, `version`, `completion`, `man`, `list`, `doctor`, `init`, `start`, `stop`, `restart`, `status`, `wpengine`, `config`

**Project Directory**: Resolved via `getProjectDir()`:
- Uses `--project-dir` flag if provided
- Otherwise uses current working directory
- Always converts to absolute path

### DDEV Integration Pattern

```go
import "github.com/firecrown-media/stax/pkg/ddev"

// Check if DDEV is installed
if !ddev.IsInstalled() {
    return errors.New("DDEV not installed")
}

// Create manager
manager := ddev.NewManager(projectDir)

// Check if running
running, err := manager.IsRunning()
if err != nil {
    return err
}

if !running {
    return errors.New("DDEV not running - run 'stax start' first")
}

// Execute command in container
output, err := manager.Exec("wp", "option", "get", "siteurl")
```

### WP-CLI Operations Pattern

```go
import "github.com/firecrown-media/stax/pkg/wordpress"

// Create CLI wrapper (uses DDEV by default)
cli := wordpress.NewCLI(projectDir)

// Search-replace with verification
opts := wordpress.SearchReplaceOptions{
    Network:     cfg.Project.Type == "wordpress-multisite",
    SkipColumns: []string{"guid"},
    DryRun:      false,
}

if err := cli.SearchReplaceWithOptions(oldURL, newURL, opts); err != nil {
    return err
}

// CRITICAL: Always verify URL changes
matches, actualURL, err := cli.VerifySiteURL(expectedURL)
if !matches {
    return fmt.Errorf("URL verification failed: expected %s, got %s", expectedURL, actualURL)
}
```

### Interactive vs Non-Interactive Mode

**Always use Safe* variants** to prevent hanging in CI/CD:

```go
import "github.com/firecrown-media/stax/pkg/prompts"

// BAD - will hang in non-interactive mode
input, err := prompts.PromptInput("Enter value", "default")

// GOOD - returns default in non-interactive mode
input, err := prompts.SafePromptInput("Enter value", "default", true)

// Check if interactive
if prompts.IsInteractive() {
    // Show interactive picker
} else {
    // Use defaults or flags
}
```

### Error Handling

```go
import (
    "github.com/firecrown-media/stax/pkg/errors"
    "github.com/firecrown-media/stax/pkg/ui"
)

// Create context-rich errors
if !running {
    return &errors.DDEVError{
        Message: "DDEV containers not running",
        Err:     fmt.Errorf("run 'stax start' to start containers"),
    }
}

// Display errors consistently
if err != nil {
    ui.Error(fmt.Sprintf("Operation failed: %v", err))
    return err
}
```

## Critical Implementation Details

### Database Pull/Push Flow

**Pull** (`cmd/db.go:runDBPull`):
1. Validate DDEV is running
2. Get WPEngine credentials from Keychain
3. Create database snapshot (backup before overwrite)
4. SSH to WPEngine, export database to temp file
5. Download SQL dump via SSH
6. Import into DDEV via `ddev import-db`
7. Run search-replace to update URLs
8. **Verify** URLs were actually changed (critical!)
9. Flush WordPress cache

**Push** (similar flow, opposite direction):
1. Confirm with user (destructive operation!)
2. Create remote snapshot if possible
3. Export from DDEV
4. Upload to WPEngine via SSH
5. Import on WPEngine
6. Run search-replace on remote
7. Verify URLs

### Media Proxy Configuration

**Purpose**: Fetch media files on-demand from CDN/WPEngine instead of syncing entire uploads directory

**Implementation** (`pkg/ddev/nginx.go`):
- Generates `.ddev/nginx_full/media-proxy.conf`
- nginx config with `try_files` → `@proxy_media` → `@wpengine_fallback`
- Optional caching with `proxy_cache_path`
- Must be wrapped in `server { }` block (validated by `ValidateNginxMediaProxyConfig`)

**Setup**:
```bash
stax media setup  # Prompts for CDN URL and WPEngine URL
```

**Generated Config Pattern**:
```nginx
proxy_cache_path /var/cache/nginx/media levels=1:2 keys_zone=media_cache:10m max_size=10g;

server {
    location ~ ^/wp-content/uploads/(.*)$ {
        try_files $uri @proxy_media;
    }

    location @proxy_media {
        proxy_pass https://cdn.example.com;
        proxy_cache media_cache;
        # ... headers and error handling
    }

    location @wpengine_fallback {
        proxy_pass https://example.wpengine.com;
    }
}
```

### URL Search-Replace for Multisite

**Challenge**: Multisite installations require network-wide + per-site replacements

**Solution** (`cmd/db_helpers.go:runSearchReplace`):
1. Detect multisite via `cfg.Project.Type == "wordpress-multisite"`
2. Run network-wide replacement with `--network` flag
3. If subdomain mode, iterate through `cfg.Network.Sites[]`
4. Run per-site replacements with `--url=<site-domain>` flag
5. Each operation uses `--all-tables --skip-columns=guid --skip-themes --skip-plugins`

**Flags Explained**:
- `--all-tables`: Ensure all tables are searched (not just core WP tables)
- `--skip-columns=guid`: Don't modify WordPress GUIDs (best practice)
- `--skip-themes --skip-plugins`: Prevent DNS lookups during search-replace
- `--network`: Apply to all sites in multisite

### Credential Management

**Storage**: macOS Keychain service `com.firecrownmedia.stax`

**Account Pattern**: `{install}.{credential-type}`
- Example: `mysite-prod.ssh-password`
- Example: `mysite-prod.api-token`

**Usage**:
```go
import "github.com/firecrown-media/stax/pkg/credentials"

// Store
err := credentials.Set(install, "ssh-password", password)

// Retrieve
password, err := credentials.Get(install, "ssh-password")

// Delete
err := credentials.Delete(install, "ssh-password")
```

## Testing

### Unit Tests

**Location**: `*_test.go` files alongside source files

**Run**: `make test` or `go test -race ./...`

**Pattern**:
```go
func TestMyFunction(t *testing.T) {
    // Use t.TempDir() for temporary directories
    tmpDir := t.TempDir()

    // Test implementation
    result, err := MyFunction(tmpDir)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

**Test Utilities**: `pkg/testutil` provides helpers for config creation, DDEV mocking

### Integration Tests

**Location**: `*_integration_test.go` files

**Run**: `RUN_INTEGRATION_TESTS=true make test-integration`

**Requirements**: Actual DDEV installation, Docker running

**Guard**:
```go
func TestIntegrationDBPull(t *testing.T) {
    if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
        t.Skip("Skipping integration test")
    }
    // ... test implementation
}
```

### Coverage

**Generate**: `make test-coverage`

**View**: Opens `coverage.html` in browser

**Goal**: Maintain >80% coverage for critical packages (config, wordpress, ddev)

## Common Issues and Solutions

### "DDEV not running" False Positives

**Problem**: `stax status` reports "DDEV not running" even when containers are up

**Cause**: Path resolution issues when using `--project-dir` or running from outside project

**Fix** (implemented in v2.12.3):
- `getProjectDir()` always returns absolute paths via `filepath.Abs()`
- Validate `.stax.yml` exists before checking DDEV status
- Enhanced error messages show which directory was checked

### URL Replacement Silent Failures

**Problem**: `wp search-replace` completes successfully but URLs aren't updated

**Cause**: Missing `--all-tables` flag or incorrect table prefix handling

**Fix** (implemented):
- Always use `SearchReplaceWithOptions()` with `--all-tables`
- Mandatory verification via `VerifySiteURL()` after every replacement
- Return error if verification fails, even if search-replace reported success

### Non-Interactive Hangs

**Problem**: Commands hang indefinitely in CI/CD or when piped

**Cause**: Using `prompts.PromptInput()` instead of `prompts.SafePromptInput()`

**Fix**:
- TTY detection via `prompts.IsInteractive()`
- All command implementations use `Safe*` variants
- Non-interactive mode returns defaults or errors if required values missing

### Nginx Configuration Errors

**Problem**: `ddev restart` fails with nginx syntax errors after media proxy setup

**Cause**: Generated nginx config had bare `location` blocks (not wrapped in `server {}`)

**Fix** (implemented):
- `GenerateMediaProxyConfig()` wraps all directives in `server {}` block
- `ValidateNginxMediaProxyConfig()` validates structure before writing
- Cache path (`proxy_cache_path`) at HTTP level, locations inside server block

## Important Files for Common Tasks

**Adding a feature**:
- `cmd/<feature>.go`: Command implementation
- `pkg/<feature>/`: Business logic package
- `README.md`: Update feature list and usage examples
- Tests: `pkg/<feature>/*_test.go`

**Fixing a bug**:
- `pkg/errors/enhanced.go`: Add/update error codes
- Relevant `pkg/` or `cmd/` file
- Add regression test

**Updating configuration schema**:
- `pkg/config/config.go`: Update `Config` struct
- `pkg/config/validate.go`: Update validation if new required fields added
- `pkg/config/manager.go`: Handle migration if needed
- Update `.stax.yml` examples in docs

**Adding a new command with business logic**:
- Create `pkg/<feature>/service.go` for the business logic
- Create `cmd/<feature>.go` as a thin wrapper that calls the service
- Commands in `cmd/` should only handle flag parsing, UI output, and calling service functions

**Release**:
- Commit using conventional commits: `feat:`, `fix:`, `docs:`, `chore:`
- Release-please creates PR with CHANGELOG
- Merge PR triggers automated release via GoReleaser

## Dependencies to Know

**CLI Framework**:
- `github.com/spf13/cobra`: Commands, flags, help
- `github.com/spf13/viper`: Configuration management

**Security**:
- `github.com/keybase/go-keychain`: macOS Keychain access
- `golang.org/x/crypto`: SSH operations (keep updated!)

**Utilities**:
- `gopkg.in/yaml.v3`: YAML parsing for `.stax.yml`
- `github.com/stretchr/testify`: Test assertions (if added)

## Documentation Files

**Spec-Kit** (spec-driven development):
- `.speckit/constitution.md`: Project principles and standards
- `.speckit/specs/`: Feature specifications
- `.speckit/plans/`: Implementation plans
- `.speckit/tasks/`: Generated task lists

**User Documentation**:
- `README.md`: User-facing documentation
- `CHANGELOG.md`: Auto-generated by release-please
- `docs/`: Full documentation (Diataxis framework)

**Development**:
- `.goreleaser.yml`: Release configuration
- `Makefile`: Build commands and targets

---

**Last Updated**: 2026-05-07 (Stax v2.21.x)
