# Internal Documentation

This directory contains internal documentation not intended for end users, including AI assistant context and development tools.

## Overview

Files in this directory are:
- Excluded from user-facing documentation
- Used for development and AI assistance
- Maintained separately from public docs
- Not included in releases

## Documents in This Section

### AI Assistant Context

**[claude.md](claude.md)** - AI Assistant Context File
- Complete project overview for AI assistants
- Current project state and status
- Technology stack and architecture
- Project structure and organization
- Key architectural concepts
- Important files and their purposes
- Common development tasks
- Development guidelines and standards
- Current known issues
- Next steps and priorities
- Working with specialized subagents

## Purpose

The AI assistant context file (claude.md) provides:

1. **Project Understanding** - Complete context for AI assistants working on Stax
2. **Architecture Overview** - High-level understanding of system design
3. **Development Guidance** - Best practices and conventions
4. **Current Status** - Up-to-date project state and priorities
5. **Code Organization** - Understanding of package structure
6. **Common Tasks** - Frequently needed development operations
7. **Known Issues** - Awareness of current problems and limitations

## When to Update

Update the AI context file when:

- Major architectural changes occur
- New packages or components are added
- Project status changes significantly
- New development patterns are adopted
- Known issues are resolved or new ones discovered
- Development priorities shift
- Documentation structure changes

## Not User-Facing

These files are intentionally excluded from:

- User documentation navigation
- Release documentation
- Installation guides
- API reference
- Command help text

## Git Ignore

The `.gitignore` file should exclude:

```gitignore
# AI assistant files (keep in repo for development)
.claude/
docs/.internal/

# But include these for team collaboration
!docs/.internal/README.md
!docs/.internal/claude.md
```

## Related Documentation

For user-facing documentation, see:
- [Main Documentation Index](../README.md)
- [Quick Start Guide](../QUICK_START.md)
- [User Guide](../USER_GUIDE.md)
- [Architecture Documentation](../technical/ARCHITECTURE.md)

For development documentation, see:
- [Contributing Guide](../CONTRIBUTING.md)
- [Testing Guide](../TESTING.md)
- [Architecture](../technical/ARCHITECTURE.md)

## Questions?

These files are for internal development use. For questions about:

- **Using Stax:** See [User Guide](../USER_GUIDE.md)
- **Contributing:** See [Contributing Guide](../CONTRIBUTING.md)
- **Development:** See [Architecture](../technical/ARCHITECTURE.md)
