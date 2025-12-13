# Stax v3.0 Refactor Implementation Plan

## Overview

This plan outlines the complete refactoring of Stax using spec-driven development with spec-kit. The refactor maintains all existing functionality while improving architecture, adding GitHub Actions deployment support, and reorganizing documentation using the Diataxis framework.

## Current State Analysis

### Existing Capabilities (Must Preserve)
1. **Project Initialization**: `stax init` - Interactive WordPress project setup
2. **Environment Management**: `start`, `stop`, `restart`, `status`, `shell`
3. **Database Operations**: `db pull`, `db push`, `db snapshot list|create|restore|delete`
4. **File Synchronization**: `files pull` via rsync over SSH
5. **Media Proxy**: `media setup`, `media status` - nginx proxy to WPEngine/CDN
6. **Credentials Management**: macOS Keychain integration
7. **Multi-provider Architecture**: WPEngine implemented, WPVIP/AWS scaffolded
8. **Multisite Support**: Subdomain and subdirectory modes
9. **Build Automation**: Composer, npm, custom scripts
10. **Diagnostics**: `diagnose`, `doctor`, `validate`

### Architecture Strengths to Keep
- Provider interface abstraction (`pkg/provider/interface.go`)
- Cobra CLI framework with Viper configuration
- DDEV integration for local development
- Secure credential storage via macOS Keychain
- Configuration migration support

### Areas for Improvement
- Documentation scattered across multiple formats
- No GitHub Actions deployment setup
- No automated .gitignore generation
- No workflow for converting existing WPEngine sites to Git repos
- Thoughts framework to be replaced with spec-kit

## Desired End State

After this refactor is complete:

1. **Spec-Kit Integration**: `.speckit/` directory with constitution, specs, and plans
2. **Diataxis Documentation**: Four-quadrant documentation structure
3. **Git Repository Setup**: `stax repo init` command to bootstrap Git repos for existing sites
4. **GitHub Actions**: `stax actions setup` to generate deployment workflows
5. **Smart File Sync**: Configurable sync with proper .gitignore generation
6. **Dependency Detection**: Clear guidance for installing DDEV, Docker, etc.
7. **Improved Provider Architecture**: Cleaner abstractions for future providers

### Verification Criteria
- All existing commands work identically (backward compatible)
- New `stax repo init` successfully creates GitHub-ready WordPress repos
- New `stax actions setup` generates working GitHub Actions workflows
- Documentation passes Diataxis structure review
- `make test` passes with >80% coverage on core packages
- `make lint` passes with no errors

## What We're NOT Doing

- Rewriting working provider implementations (WPEngine)
- Changing the release process (Homebrew, release-please, GoReleaser)
- Adding new hosting providers (WPVIP, AWS) - just improving abstractions
- Changing credential storage mechanism
- Supporting non-macOS platforms (future work)

---

## Implementation Phases

### Phase 1: Spec-Kit Foundation & Cleanup

**Objective**: Set up spec-kit, remove thoughts framework, establish project constitution.

#### 1.1 Remove Thoughts Framework

**Files to Remove**:
- `thoughts/` directory
- Any references to thoughts in documentation

**Commands**:
```bash
rm -rf thoughts/
git add -A && git commit -m "chore: remove thoughts framework in favor of spec-kit"
```

#### 1.2 Initialize Spec-Kit

**Create Directory Structure**:
```
.speckit/
├── constitution.md      # Project principles and standards
├── specs/               # Feature specifications
├── plans/               # Implementation plans (this file)
└── tasks/               # Generated task lists
```

**Spec-Kit Constitution** (`.speckit/constitution.md`):
```markdown
# Stax Project Constitution

## Project Identity
Stax is a WordPress development CLI tool that streamlines local development workflows with hosting provider integration.

## Core Principles

### 1. Developer Experience First
- Commands should be intuitive and memorable
- Error messages must be actionable
- Defaults should work for 90% of use cases

### 2. Provider Agnostic Architecture
- All hosting-specific code behind provider interfaces
- New providers addable without core changes
- WPEngine is reference implementation

### 3. Security by Default
- Credentials never stored in plain text
- macOS Keychain for sensitive data
- No credentials in configuration files

### 4. DDEV as Foundation
- DDEV provides container orchestration
- Stax orchestrates DDEV, not Docker directly
- Leverage DDEV's WordPress expertise

### 5. Git-First Workflow
- Configuration files designed for version control
- .gitignore generation for WordPress best practices
- Support source-based development workflows

## Technical Standards

### Code Quality
- Go 1.23+ with standard formatting (gofmt)
- golangci-lint for static analysis
- >80% test coverage for core packages
- Race detection enabled in tests

### CLI Conventions
- Cobra for command structure
- Viper for configuration
- pkg/ui for consistent output formatting
- pkg/prompts for interactive input (with Safe* variants)

### Documentation
- Diataxis framework (tutorials, how-tos, reference, explanation)
- Man pages generated from code
- CLAUDE.md for AI assistant context

### Release Process
- Conventional commits (feat:, fix:, docs:, chore:)
- Release-please for changelog and versioning
- GoReleaser for builds
- Homebrew tap for distribution
```

#### 1.3 Update CLAUDE.md

Update the existing CLAUDE.md to reference spec-kit and the new architecture direction.

**Success Criteria - Phase 1**:

#### Automated Verification:
- [ ] `thoughts/` directory removed: `[ ! -d thoughts/ ]`
- [ ] `.speckit/` directory exists with constitution
- [ ] `make lint` passes
- [ ] `make test` passes

#### Manual Verification:
- [ ] Constitution accurately reflects project values
- [ ] CLAUDE.md updated with spec-kit references

---

### Phase 2: Documentation Restructure (Diataxis)

**Objective**: Reorganize all documentation following the Diataxis framework.

#### 2.1 New Documentation Structure

```
docs/
├── README.md                    # Documentation hub/index
├── tutorials/                   # Learning-oriented
│   ├── README.md
│   ├── first-project.md         # Your first Stax project
│   ├── database-workflow.md     # Learning database sync
│   └── multisite-setup.md       # Setting up multisite
├── how-to/                      # Task-oriented
│   ├── README.md
│   ├── install-stax.md          # How to install Stax
│   ├── sync-database.md         # How to sync database
│   ├── setup-media-proxy.md     # How to configure media proxy
│   ├── push-database.md         # How to push database changes
│   ├── manage-snapshots.md      # How to use database snapshots
│   ├── setup-github-actions.md  # How to set up CI/CD (NEW)
│   ├── convert-site-to-git.md   # How to add Git to existing site (NEW)
│   └── troubleshoot-common.md   # How to fix common issues
├── reference/                   # Information-oriented
│   ├── README.md
│   ├── commands.md              # Complete command reference
│   ├── configuration.md         # .stax.yml specification
│   ├── providers.md             # Provider capabilities matrix
│   ├── gitignore-template.md    # WordPress .gitignore reference (NEW)
│   └── github-actions.md        # Generated workflow reference (NEW)
├── explanation/                 # Understanding-oriented
│   ├── README.md
│   ├── architecture.md          # System architecture
│   ├── why-ddev.md              # Why we chose DDEV
│   ├── provider-system.md       # Multi-provider design
│   ├── security-model.md        # Credential storage explained
│   └── release-process.md       # How releases work
└── internal/                    # Developer documentation
    ├── README.md
    ├── contributing.md
    ├── testing.md
    └── releasing.md
```

#### 2.2 Content Migration Map

| Current File | New Location | Type |
|-------------|--------------|------|
| GETTING_STARTED.md | tutorials/first-project.md | Tutorial |
| QUICK_START.md | how-to/install-stax.md | How-to |
| INSTALLATION.md | how-to/install-stax.md | How-to |
| COMMAND_REFERENCE.md | reference/commands.md | Reference |
| CONFIG_SPEC.md | reference/configuration.md | Reference |
| ARCHITECTURE.md | explanation/architecture.md | Explanation |
| MULTISITE.md | tutorials/multisite-setup.md | Tutorial |
| MEDIA_PROXY.md | how-to/setup-media-proxy.md | How-to |
| TROUBLESHOOTING.md | how-to/troubleshoot-common.md | How-to |
| FAQ.md | (distribute to appropriate sections) | Various |
| WPENGINE.md | reference/providers.md | Reference |
| guides/*.md | how-to/*.md | How-to |

#### 2.3 README.md Update

The root README.md should be simplified to:
- Brief description
- Quick install command
- Link to docs/tutorials/first-project.md
- Link to docs/ for full documentation

**Success Criteria - Phase 2**:

#### Automated Verification:
- [ ] All documentation files exist in new structure
- [ ] No broken internal links: `make docs-check` (new target)
- [ ] Old documentation files removed or redirected

#### Manual Verification:
- [ ] Each document clearly fits one Diataxis quadrant
- [ ] Tutorials provide hands-on learning experiences
- [ ] How-tos are concise and goal-focused
- [ ] Reference is comprehensive and neutral
- [ ] Explanation provides context and understanding

---

### Phase 3: Dependency Detection System

**Objective**: Implement clear dependency checking with helpful installation guidance.

#### 3.1 New Package: `pkg/prerequisites`

**File**: `pkg/prerequisites/prerequisites.go`

```go
package prerequisites

// Dependency represents a required system dependency
type Dependency struct {
    Name        string
    Command     string   // Command to check existence
    MinVersion  string   // Minimum required version
    InstallURL  string   // Documentation URL
    InstallCmd  string   // Suggested install command (informational)
    Required    bool     // Is this strictly required?
}

// DefaultDependencies returns the list of Stax prerequisites
func DefaultDependencies() []Dependency {
    return []Dependency{
        {
            Name:       "Docker",
            Command:    "docker",
            MinVersion: "20.0.0",
            InstallURL: "https://docs.docker.com/desktop/install/mac-install/",
            InstallCmd: "Download from https://docker.com/products/docker-desktop",
            Required:   true,
        },
        {
            Name:       "DDEV",
            Command:    "ddev",
            MinVersion: "1.22.0",
            InstallURL: "https://ddev.readthedocs.io/en/stable/users/install/",
            InstallCmd: "brew install ddev/ddev/ddev",
            Required:   true,
        },
        {
            Name:       "Git",
            Command:    "git",
            MinVersion: "2.0.0",
            InstallURL: "https://git-scm.com/download/mac",
            InstallCmd: "brew install git",
            Required:   true,
        },
        {
            Name:       "GitHub CLI",
            Command:    "gh",
            MinVersion: "2.0.0",
            InstallURL: "https://cli.github.com/",
            InstallCmd: "brew install gh",
            Required:   false, // Only for GitHub Actions setup
        },
    }
}

// Check verifies a dependency is installed and meets version requirements
func (d *Dependency) Check() *CheckResult

// CheckAll verifies all dependencies
func CheckAll(deps []Dependency) []CheckResult
```

#### 3.2 New Command: `stax doctor`

Enhance existing `doctor` command to include prerequisite checking:

```go
// cmd/doctor.go additions
func runDoctor(cmd *cobra.Command, args []string) error {
    ui.Section("Checking Prerequisites")

    results := prerequisites.CheckAll(prerequisites.DefaultDependencies())

    for _, result := range results {
        if result.OK {
            ui.Success("%s %s ✓", result.Dependency.Name, result.Version)
        } else if result.Dependency.Required {
            ui.Error("%s not found or version too old", result.Dependency.Name)
            ui.Info("  Install: %s", result.Dependency.InstallCmd)
            ui.Info("  Docs: %s", result.Dependency.InstallURL)
        } else {
            ui.Warning("%s not found (optional)", result.Dependency.Name)
            ui.Info("  Install: %s", result.Dependency.InstallCmd)
        }
    }

    // ... rest of doctor checks
}
```

#### 3.3 Pre-flight Checks in `stax init`

Add prerequisite verification at the start of `stax init`:

```go
// cmd/init.go additions
func runInit(cmd *cobra.Command, args []string) error {
    // Check required prerequisites first
    results := prerequisites.CheckAll(prerequisites.RequiredOnly())
    missing := prerequisites.FilterFailed(results)

    if len(missing) > 0 {
        ui.Error("Missing required dependencies:")
        for _, m := range missing {
            ui.Info("  • %s: %s", m.Dependency.Name, m.Dependency.InstallCmd)
        }
        return fmt.Errorf("please install missing dependencies and try again")
    }

    // ... rest of init
}
```

**Success Criteria - Phase 3**:

#### Automated Verification:
- [ ] `pkg/prerequisites` package exists with tests
- [ ] `stax doctor` shows dependency status
- [ ] `stax init` fails gracefully with missing deps
- [ ] `make test` passes

#### Manual Verification:
- [ ] Error messages are clear and actionable
- [ ] Install commands are accurate for macOS
- [ ] Version detection works correctly

---

### Phase 4: Git Repository Initialization

**Objective**: Add `stax repo init` command to bootstrap Git repos for existing WPEngine sites.

#### 4.1 New Command: `stax repo init`

**File**: `cmd/repo.go`

```go
// stax repo init - Initialize a Git repository for an existing WPEngine site
//
// This command:
// 1. Creates a new Git repository (or uses existing)
// 2. Syncs WordPress files from WPEngine (themes, plugins, mu-plugins)
// 3. Generates appropriate .gitignore
// 4. Creates initial commit
// 5. Optionally creates GitHub repository and pushes

var repoInitCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize a Git repository for a WordPress site",
    Long: `Initialize a Git repository for an existing WordPress site.

This command syncs custom code from WPEngine and sets up a proper
Git repository with WordPress-appropriate .gitignore.

Example:
  stax repo init --install mysite-prod --github myorg/mysite`,
    RunE: runRepoInit,
}

func init() {
    repoCmd.AddCommand(repoInitCmd)

    repoInitCmd.Flags().String("install", "", "WPEngine install name")
    repoInitCmd.Flags().String("github", "", "GitHub repository (org/repo)")
    repoInitCmd.Flags().Bool("private", true, "Create private GitHub repository")
    repoInitCmd.Flags().StringSlice("sync-dirs",
        []string{"wp-content/themes", "wp-content/plugins", "wp-content/mu-plugins"},
        "Directories to sync from WPEngine")
}
```

#### 4.2 WordPress .gitignore Template

**File**: `pkg/git/templates/wordpress.gitignore`

```gitignore
# WordPress Core (don't commit - install via composer or download)
/wp-admin/
/wp-includes/
/wp-*.php
/index.php
/license.txt
/readme.html
/xmlrpc.php

# WPEngine specific
mysql.sql
.smushit-status
.wpengine-conf/

# wp-content exceptions
/wp-content/uploads/
/wp-content/cache/
/wp-content/upgrade/
/wp-content/backup*/
/wp-content/backups/
/wp-content/blogs.dir/
/wp-content/debug.log
/wp-content/advanced-cache.php
/wp-content/object-cache.php
/wp-content/wp-cache-config.php

# Build artifacts in themes/plugins
/wp-content/themes/*/node_modules/
/wp-content/plugins/*/node_modules/
/wp-content/themes/*/.sass-cache/
/wp-content/plugins/*/.sass-cache/

# Dependency directories
/vendor/
/node_modules/

# Environment and local config
.env
.env.*
wp-config-local.php
*.local.php

# IDE and editor files
.idea/
.vscode/
*.swp
*.swo
.DS_Store
Thumbs.db

# Stax local files
.stax.local.yml

# DDEV (local development)
.ddev/
!.ddev/config.yaml
!.ddev/commands/

# Logs
*.log
logs/
```

#### 4.3 Enhanced File Sync

**File**: `pkg/sync/sync.go`

```go
package sync

// SyncConfig defines what to sync from remote
type SyncConfig struct {
    // Directories to include (relative to WordPress root)
    IncludeDirs []string

    // Patterns to exclude (applied to all directories)
    ExcludePatterns []string

    // Whether to delete local files not on remote
    DeleteOrphans bool

    // Bandwidth limit in KB/s (0 = unlimited)
    BandwidthLimit int
}

// DefaultSyncConfig returns WordPress best-practice sync settings
func DefaultSyncConfig() SyncConfig {
    return SyncConfig{
        IncludeDirs: []string{
            "wp-content/themes",
            "wp-content/plugins",
            "wp-content/mu-plugins",
        },
        ExcludePatterns: []string{
            "node_modules/",
            ".git/",
            ".sass-cache/",
            "*.log",
            ".DS_Store",
            "mysql.sql",
        },
        DeleteOrphans:  false,
        BandwidthLimit: 0,
    }
}
```

#### 4.4 Repo Init Flow

```
stax repo init --install mysite-prod --github myorg/mysite

1. Validate WPEngine credentials and install exists
2. Create directory structure if needed
3. Initialize Git repository (git init)
4. Generate .gitignore from template
5. Sync files from WPEngine:
   - wp-content/themes/
   - wp-content/plugins/
   - wp-content/mu-plugins/
   (excluding node_modules, mysql.sql, etc.)
6. Create initial commit
7. If --github specified:
   a. Check gh CLI is installed
   b. Create GitHub repository (gh repo create)
   c. Add remote and push
8. Generate .stax.yml configuration
9. Display next steps
```

**Success Criteria - Phase 4**:

#### Automated Verification:
- [ ] `stax repo init --help` shows correct usage
- [ ] `.gitignore` template generates correctly
- [ ] File sync excludes mysql.sql and node_modules
- [ ] `make test` passes for new packages

#### Manual Verification:
- [ ] Full workflow tested with real WPEngine site
- [ ] Generated .gitignore is comprehensive
- [ ] GitHub repository creation works (with gh CLI)
- [ ] Initial commit contains only intended files

---

### Phase 5: GitHub Actions Deployment Setup

**Objective**: Add `stax actions setup` to generate GitHub Actions workflows for WPEngine deployment.

#### 5.1 New Command: `stax actions setup`

**File**: `cmd/actions.go`

```go
var actionsSetupCmd = &cobra.Command{
    Use:   "setup",
    Short: "Set up GitHub Actions for WPEngine deployment",
    Long: `Generate GitHub Actions workflow files for deploying to WPEngine.

This command creates:
- .github/workflows/deploy.yml - Deployment workflow
- Recommends branch protection settings
- Provides instructions for adding secrets

Example:
  stax actions setup --production main --staging develop`,
    RunE: runActionsSetup,
}

func init() {
    actionsCmd.AddCommand(actionsSetupCmd)

    actionsSetupCmd.Flags().String("production", "main", "Branch for production deployment")
    actionsSetupCmd.Flags().String("staging", "develop", "Branch for staging deployment")
    actionsSetupCmd.Flags().String("prod-install", "", "WPEngine production install name")
    actionsSetupCmd.Flags().String("stage-install", "", "WPEngine staging install name")
}
```

#### 5.2 Workflow Template

**File**: `pkg/actions/templates/deploy.yml`

```yaml
name: Deploy to WPEngine

on:
  push:
    branches:
      - {{.ProductionBranch}}
      {{- if .StagingBranch}}
      - {{.StagingBranch}}
      {{- end}}

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Build assets
        run: npm run build

      - name: Deploy to WPEngine
        uses: wpengine/github-action-wpe-site-deploy@v3
        with:
          WPE_SSHG_KEY_PRIVATE: ${{"{{"}} secrets.WPE_SSHG_KEY_PRIVATE {{"}}"}}
          WPE_ENV: ${{"{{"}} github.ref == 'refs/heads/{{.ProductionBranch}}' && '{{.ProductionInstall}}' || '{{.StagingInstall}}' {{"}}"}}
          SRC_PATH: "wp-content/"
          REMOTE_PATH: "wp-content/"
          PHP_LINT: true
          FLAGS: -azvr --inplace --delete --exclude=".*" --exclude="node_modules/" --exclude="*.sql"
```

#### 5.3 Branch Protection Recommendations

**File**: `pkg/actions/templates/branch-protection.md`

```markdown
# Recommended Branch Protection Settings

## Production Branch ({{.ProductionBranch}})

### Settings
- ✅ Require pull request before merging
  - ✅ Require approvals: 1
  - ✅ Dismiss stale pull request approvals when new commits are pushed
- ✅ Require status checks to pass before merging
  - Required checks: `deploy`
- ✅ Require conversation resolution before merging
- ✅ Do not allow bypassing the above settings

### CODEOWNERS
Create `.github/CODEOWNERS`:
```
# Default owners for everything
* @your-org/maintainers

# Theme owners
/wp-content/themes/ @your-org/frontend-team

# Plugin owners
/wp-content/plugins/ @your-org/backend-team
```

## Staging Branch ({{.StagingBranch}})

### Settings
- ✅ Require pull request before merging
  - Require approvals: 0 (optional for staging)
- ✅ Require status checks to pass before merging

## Required GitHub Secrets

Add these secrets in Settings → Secrets and variables → Actions:

| Secret Name | Description | How to Get |
|------------|-------------|-----------|
| `WPE_SSHG_KEY_PRIVATE` | WPEngine SSH private key | WPEngine Portal → SSH Gateway |

## Workflow Permissions

Ensure Actions have appropriate permissions:
Settings → Actions → General → Workflow permissions
- ✅ Read and write permissions
```

#### 5.4 Actions Setup Flow

```
stax actions setup --production main --staging develop

1. Verify .git directory exists
2. Check for existing workflow files
3. Read .stax.yml for WPEngine install names
4. Generate .github/workflows/deploy.yml
5. Generate .github/CODEOWNERS template
6. Display:
   - Branch protection recommendations
   - Required secrets and how to obtain them
   - Next steps for GitHub configuration
```

**Success Criteria - Phase 5**:

#### Automated Verification:
- [ ] `stax actions setup --help` shows correct usage
- [ ] Workflow template generates valid YAML
- [ ] Generated workflow passes `actionlint` validation
- [ ] `make test` passes

#### Manual Verification:
- [ ] Generated workflow deploys successfully to WPEngine
- [ ] Branch protection instructions are accurate
- [ ] CODEOWNERS template is appropriate
- [ ] Secrets documentation is complete

---

### Phase 6: Provider Architecture Refinement

**Objective**: Clean up provider interfaces for better extensibility.

#### 6.1 Provider Interface Audit

Review and document current provider interface:

```go
// pkg/provider/interface.go - document each method's purpose
type Provider interface {
    // Core identification
    Name() string
    Description() string
    Capabilities() ProviderCapabilities

    // Authentication (required by all providers)
    Authenticate(credentials map[string]string) error
    TestConnection() error
    ValidateCredentials(credentials map[string]string) error

    // Site operations (required by all providers)
    ListSites() ([]Site, error)
    GetSite(identifier string) (*Site, error)
    GetSiteMetadata(site *Site) (*SiteMetadata, error)

    // Database operations
    ExportDatabase(site *Site, options DatabaseExportOptions) (io.ReadCloser, error)
    ImportDatabase(site *Site, data io.Reader, options DatabaseImportOptions) error
    GetDatabaseCredentials(site *Site) (*DatabaseCredentials, error)

    // File operations
    SyncFiles(site *Site, destination string, options SyncOptions) error
    DownloadFile(site *Site, remotePath string) (io.ReadCloser, error)
    UploadFile(site *Site, localPath, remotePath string) error

    // Environment info
    GetPHPVersion(site *Site) (string, error)
    GetMySQLVersion(site *Site) (string, error)
    GetWordPressVersion(site *Site) (string, error)
}
```

#### 6.2 Add Provider Registration Helpers

```go
// pkg/provider/registry.go
var providers = make(map[string]func() Provider)

// Register adds a provider factory to the registry
func Register(name string, factory func() Provider) {
    providers[name] = factory
}

// Get returns a provider instance by name
func Get(name string) (Provider, error) {
    factory, ok := providers[name]
    if !ok {
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
    return factory(), nil
}

// List returns all registered provider names
func List() []string {
    names := make([]string, 0, len(providers))
    for name := range providers {
        names = append(names, name)
    }
    sort.Strings(names)
    return names
}
```

#### 6.3 Document Provider Implementation Guide

Create `docs/internal/implementing-providers.md` with:
- Required interface methods
- Optional capability interfaces
- Testing requirements
- Example skeleton implementation

**Success Criteria - Phase 6**:

#### Automated Verification:
- [ ] Provider interface fully documented
- [ ] `make test` passes
- [ ] No breaking changes to WPEngine provider

#### Manual Verification:
- [ ] Documentation clearly explains how to add new providers
- [ ] Provider architecture supports future WPVIP/AWS additions

---

### Phase 7: Integration Testing & Polish

**Objective**: Ensure all new features work together and existing functionality is preserved.

#### 7.1 End-to-End Test Scenarios

Create integration tests for:

1. **Fresh Project Setup**
   ```bash
   stax init --install mysite --start
   # Verify: DDEV running, database imported, site accessible
   ```

2. **Repository Initialization**
   ```bash
   stax repo init --install mysite --github testorg/test-site
   # Verify: .git exists, .gitignore correct, files synced
   ```

3. **Actions Setup**
   ```bash
   stax actions setup --production main --staging develop
   # Verify: workflow file exists and is valid
   ```

4. **Database Workflow**
   ```bash
   stax db pull
   stax db snapshot create test-snap
   stax db restore test-snap
   # Verify: database operations work correctly
   ```

5. **Media Proxy**
   ```bash
   stax media setup
   # Verify: nginx config generated, media proxying works
   ```

#### 7.2 Backward Compatibility Tests

Verify all existing commands work:
- `stax init` (existing behavior)
- `stax start/stop/restart/status`
- `stax db pull/push`
- `stax files pull`
- `stax media setup/status`
- `stax credentials set/get/delete`
- `stax doctor/diagnose/validate`

#### 7.3 Documentation Review

- [ ] All commands documented
- [ ] All configuration options documented
- [ ] Examples tested and working
- [ ] Man page up to date

**Success Criteria - Phase 7**:

#### Automated Verification:
- [ ] `make test` passes
- [ ] `make test-integration` passes (with RUN_INTEGRATION_TESTS=true)
- [ ] `make lint` passes
- [ ] Documentation link checker passes

#### Manual Verification:
- [ ] Complete workflow tested end-to-end
- [ ] Existing user workflows unaffected
- [ ] New features intuitive to use

---

## Spec-Kit Commands for Implementation

### Phase 1: Foundation
```bash
# After manually creating .speckit/constitution.md
/speckit.analyze  # Verify constitution is complete
```

### Phase 2-7: Feature Development
For each phase, use this workflow:

```bash
# 1. Create specification for the phase
/speckit.specify
# Describe: "Phase X - [Description]"

# 2. Generate implementation plan
/speckit.plan

# 3. Generate tasks
/speckit.tasks

# 4. Implement
/speckit.implement

# 5. Validate
/speckit.analyze
```

---

## Branching Strategy

All work on single feature branch:

```
main
  └── feature/stax-v3-refactor
        ├── Phase 1 commits
        ├── Phase 2 commits
        ├── ...
        └── Phase 7 commits
```

### Commit Convention

```
feat(phase-1): set up spec-kit and remove thoughts framework
feat(phase-2): restructure documentation to Diataxis format
feat(phase-3): add dependency detection system
feat(phase-4): add stax repo init command
feat(phase-5): add GitHub Actions setup command
refactor(phase-6): improve provider architecture documentation
test(phase-7): add integration tests for new features
```

### Merge Strategy

After all phases complete and tested:
1. Create PR from `feature/stax-v3-refactor` to `main`
2. Review full changeset
3. Squash merge or regular merge based on preference
4. Release-please will create v3.0.0 release PR

---

## Timeline Estimate

This plan does not include time estimates. Work should proceed phase by phase, with each phase completed and verified before moving to the next.

---

## References

- Spec-Kit Documentation: https://github.com/github/spec-kit
- Diataxis Framework: https://diataxis.fr/
- WPEngine GitHub Action: https://github.com/wpengine/github-action-wpe-site-deploy
- DDEV Documentation: https://ddev.readthedocs.io/
- Cobra CLI Framework: https://cobra.dev/
