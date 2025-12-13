# Stax TODO Tracking

This document tracks outstanding TODO items in the codebase, categorized by priority and area.

## Provider Implementations (Future Work)

### AWS Provider (`pkg/providers/aws/`)
**Status**: Stub implementation only

All methods return "not yet implemented" errors:
- Authentication and credential validation
- Site listing and retrieval
- Database export/import
- File sync, upload, download
- Version queries (PHP, MySQL, WordPress)

**Effort**: Large - requires AWS SDK integration, IAM configuration, and testing with actual AWS infrastructure.

### WordPress VIP Provider (`pkg/providers/wordpress-vip/`)
**Status**: Stub implementation only

All methods return "not yet implemented" errors:
- VIP API authentication
- Site listing and retrieval
- Database operations
- File operations
- Version queries

**Effort**: Large - requires VIP CLI/API integration and VIP account for testing.

### Local Provider (`pkg/providers/local/`)
**Status**: Partial implementation

Working:
- Basic site listing (returns mock data)
- Credential and connection operations (no-op)

Not implemented:
- DDEV integration for actual local site discovery
- Database export/import via DDEV
- Version queries via DDEV

**Effort**: Medium - requires DDEV integration work.

## Command Implementations

### Start Command (`cmd/start.go`)
**Status**: Placeholder

TODOs (lines 91-95):
- Check if DDEV is installed
- Run `ddev start`
- Enable Xdebug if requested
- Run build process if requested
- Display environment URLs

**Effort**: Small - most logic exists in `pkg/ddev`, just needs wiring.

### Stop Command (`cmd/stop.go`)
**Status**: Placeholder

TODOs (lines 86-87):
- Run `ddev stop`
- Optionally remove data if requested

**Effort**: Small - straightforward DDEV wrapper.

### Restart Command (`cmd/restart.go`)
**Status**: Placeholder

TODOs (lines 76-78):
- Run `ddev restart`
- Enable Xdebug if requested
- Run build process if requested

**Effort**: Small - straightforward DDEV wrapper.

### Provider Command (`cmd/provider.go`)
**Status**: Partial implementation

TODOs:
- Line 217: Update `.stax.yml` with new provider
- Line 226: Load configuration and create provider instance
- Line 292: Implement YAML output for provider list

**Effort**: Small - mostly configuration file manipulation.

## Internal TODOs

### Provider Manager (`pkg/provider/manager.go`)
**Status**: Working with one TODO

TODO (line 229):
- Implement manual migration logic

**Effort**: Medium - requires understanding of cross-provider migration requirements.

## Priority Recommendations

### High Priority (Core Functionality)
1. **Start/Stop/Restart commands** - Essential for developer workflow
2. **Local provider DDEV integration** - Enables local-to-remote sync

### Medium Priority (Feature Expansion)
3. **Provider command completion** - Better provider management UX
4. **Provider migration logic** - Enables multi-provider workflows

### Low Priority (Future Roadmap)
5. **AWS provider** - Enterprise feature
6. **WordPress VIP provider** - Enterprise feature

## How to Track Progress

When completing a TODO:
1. Remove the TODO comment
2. Update this document
3. Add tests for the new functionality
4. Update CHANGELOG.md if user-facing

## Related Issues

Create GitHub issues for major TODO categories:
- [ ] Issue: Implement start/stop/restart commands
- [ ] Issue: Complete local provider DDEV integration
- [ ] Issue: AWS provider implementation
- [ ] Issue: WordPress VIP provider implementation
