# Technical Documentation

This directory contains in-depth technical documentation about Stax's architecture, implementation details, and system design.

## Overview

Technical documentation is intended for:
- Developers working on Stax itself
- Contributors adding features or fixing bugs
- Technical teams evaluating Stax's architecture
- DevOps engineers integrating Stax into workflows

## Documents in This Section

### System Architecture

**[ARCHITECTURE.md](ARCHITECTURE.md)** - System Architecture & Design
- Technology stack decisions (Go, DDEV, Cobra)
- High-level architecture diagrams
- Component overview and responsibilities
- Data flow diagrams
- File structure organization
- Development workflow integration
- Performance optimization strategies
- Security considerations
- Extensibility and future enhancements

### CLI Command Structure

**[COMMANDS.md](COMMANDS.md)** - Command Specifications
- Complete command tree structure
- Command structure and patterns
- Flag definitions and usage
- Subcommand organization
- Command aliases
- Exit codes
- Environment variables
- Shell completion

### Configuration System

**[CONFIG_SPEC.md](CONFIG_SPEC.md)** - Configuration Specification
- Configuration file schema (YAML)
- All configuration options documented
- Default values
- Validation rules
- Environment variable overrides
- Configuration precedence
- Multi-source config loading
- Example configurations

### Provider Integration

**[WPENGINE_INTEGRATION.md](WPENGINE_INTEGRATION.md)** - WPEngine Provider Details
- WPEngine API integration
- SSH Gateway operations
- Authentication mechanisms
- Database export/import strategies
- File synchronization (rsync)
- Remote media proxy configuration
- Search-replace for multisite
- Error handling and retries

### Container Setup

**[DDEV_MULTISITE_IMPLEMENTATION.md](DDEV_MULTISITE_IMPLEMENTATION.md)** - DDEV & Multisite
- DDEV configuration generation
- WordPress multisite detection and setup
- Subdomain vs subdirectory modes
- SSL certificate automation
- Nginx configuration for media proxy
- Container lifecycle management
- Hosts file management
- Post-start hooks

## Reading Order

For new developers or contributors, we recommend reading in this order:

1. **ARCHITECTURE.md** - Start here for the big picture
2. **COMMANDS.md** - Understand the CLI structure
3. **CONFIG_SPEC.md** - Learn the configuration system
4. **DDEV_MULTISITE_IMPLEMENTATION.md** - Container and multisite details
5. **WPENGINE_INTEGRATION.md** - Provider implementation example

## Key Concepts

### Provider Abstraction
Stax uses a provider interface pattern to support multiple hosting platforms. WPEngine is the primary implementation, with stubs for AWS and WordPress VIP.

### DDEV Integration
DDEV was chosen over Podman and Docker Compose for its WordPress-optimized configuration, multisite support, and Mac performance.

### Configuration Management
Multi-source configuration loading (global, project, environment) with Viper provides flexibility and team consistency.

### Security-First Design
Credentials stored in macOS Keychain, HTTPS enforcement, input validation, and secure defaults throughout.

## Related Documentation

- [Testing Guide](../TESTING.md) - Writing and running tests
- [Security Overview](../SECURITY.md) - Security guidelines
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute

## Contributing

When modifying Stax's architecture or implementation:

1. Update relevant technical documentation
2. Include architecture diagrams when helpful
3. Document design decisions and rationale
4. Update configuration specs if adding new options
5. Add examples to command documentation
6. Keep documentation in sync with code

## Questions?

For questions about Stax's architecture or implementation:

1. Review the relevant technical document
2. Examine the source code in `pkg/` and `cmd/`
3. Check [Contributing Guide](../CONTRIBUTING.md) for development workflow
