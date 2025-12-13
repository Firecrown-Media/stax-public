# Stax Project Constitution

## Project Identity

**Stax** is a WordPress development CLI tool that streamlines local development workflows with hosting provider integration. It enables WordPress developers to work efficiently with source-based development, Git workflows, and automated deployments.

## Mission Statement

Reduce the friction of WordPress development by providing a single, intuitive CLI that handles environment setup, database synchronization, file management, and deployment configuration.

## Core Principles

### 1. Developer Experience First

- Commands should be intuitive and memorable
- Error messages must be actionable with clear next steps
- Defaults should work for 90% of use cases
- Complex operations should be achievable with simple commands
- Progress feedback for long-running operations

### 2. Provider Agnostic Architecture

- All hosting-specific code behind provider interfaces
- New providers addable without core changes
- WPEngine is the reference implementation
- Future providers (WPVIP, AWS) follow same patterns
- Provider capabilities clearly documented

### 3. Security by Default

- Credentials never stored in plain text files
- macOS Keychain for all sensitive data
- No credentials in configuration files that might be committed
- Secure SSH key handling for remote operations
- Warn users about potential security issues

### 4. DDEV as Foundation

- DDEV provides container orchestration
- Stax orchestrates DDEV, not Docker directly
- Leverage DDEV's WordPress-specific optimizations
- Don't reinvent what DDEV already does well
- Ensure compatibility with DDEV updates

### 5. Git-First Workflow

- Configuration files designed for version control
- Automatic .gitignore generation for WordPress
- Support source-based development patterns
- Enable GitFlow and similar branching strategies
- GitHub Actions integration for CI/CD

### 6. Non-Destructive by Default

- Database snapshots before dangerous operations
- Confirmation prompts for destructive actions
- Clear warnings when operations affect remote systems
- Easy rollback mechanisms

## Technical Standards

### Code Quality

- **Language**: Go 1.23+ with standard formatting (gofmt)
- **Linting**: golangci-lint with project configuration
- **Testing**: >80% coverage for core packages
- **Race Detection**: Enabled in all test runs
- **Error Handling**: Context-rich errors with pkg/errors

### CLI Conventions

- **Framework**: Cobra for command structure
- **Configuration**: Viper for config management
- **Output**: pkg/ui for consistent terminal formatting
- **Input**: pkg/prompts with Safe* variants for non-interactive support
- **Progress**: Spinners and progress bars for long operations

### Documentation Standards

- **Framework**: Diataxis (tutorials, how-tos, reference, explanation)
- **Man Pages**: Generated from code comments
- **AI Context**: CLAUDE.md maintained for AI assistant context
- **Spec-Kit**: .speckit/ for specifications and plans

### Release Process

- **Commits**: Conventional commits (feat:, fix:, docs:, chore:)
- **Versioning**: Semantic versioning via release-please
- **Building**: GoReleaser for cross-platform builds
- **Distribution**: Homebrew tap for macOS installation

## Architectural Decisions

### Configuration Hierarchy

1. Command-line flags (highest priority)
2. Environment variables
3. Project configuration (.stax.yml)
4. Global configuration (~/.stax/config.yml)
5. Built-in defaults (lowest priority)

### Package Structure

```
cmd/           # CLI command definitions
pkg/
  ├── config/      # Configuration management
  ├── credentials/ # Secure credential storage
  ├── ddev/        # DDEV integration
  ├── errors/      # Enhanced error types
  ├── git/         # Git operations
  ├── prerequisites/ # Dependency checking
  ├── prompts/     # User input handling
  ├── provider/    # Provider interface
  ├── providers/   # Provider implementations
  ├── snapshot/    # Database snapshots
  ├── sync/        # File synchronization
  ├── ui/          # Terminal output
  ├── wordpress/   # WordPress operations
  └── wpengine/    # WPEngine-specific code
```

### Provider Interface Contract

All providers must implement:
- Authentication and connection testing
- Site listing and metadata retrieval
- Database export (import where supported)
- File synchronization
- Version detection (PHP, MySQL, WordPress)

Optional capabilities:
- Remote command execution
- Backup management
- Deployment triggering
- Media/CDN management

## User Personas

### Primary: WordPress Developer

- Works on custom themes and plugins
- Needs to sync database and files from production/staging
- Uses Git for version control
- Deploys via Git push or CI/CD
- Comfortable with command line

### Secondary: Agency Team Lead

- Sets up projects for team members
- Defines branching and deployment strategies
- Configures branch protection and code review
- Needs consistent environments across team

## Quality Gates

### Before Merge

- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] No race conditions (`go test -race`)
- [ ] Documentation updated
- [ ] CHANGELOG entry (via conventional commits)

### Before Release

- [ ] Integration tests pass
- [ ] Man page regenerated
- [ ] README reflects changes
- [ ] Breaking changes documented

## Evolution Guidelines

### Adding New Commands

1. Fits within existing command hierarchy
2. Follows naming conventions of similar commands
3. Includes --help with examples
4. Has both unit and integration tests
5. Documented in reference section

### Adding New Providers

1. Implements full Provider interface
2. Documents supported capabilities
3. Has provider-specific tests
4. Added to provider documentation
5. Configuration examples provided

### Breaking Changes

1. Major version bump required
2. Migration guide provided
3. Deprecation warnings in prior version
4. Clear communication in CHANGELOG
