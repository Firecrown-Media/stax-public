# Release Documentation

This directory contains all documentation related to the Stax release process.

## Quick Start

**For first-time release:**
1. Follow [AUTOMATED_RELEASE_PROCESS.md](AUTOMATED_RELEASE_PROCESS.md) - Step-by-step guide
2. Use [QUICK_REFERENCE.md](QUICK_REFERENCE.md) - Quick command reference

**For regular development:**
1. Read [FEATURE_BRANCH_WORKFLOW.md](FEATURE_BRANCH_WORKFLOW.md) - Feature branch process
2. Use conventional commits (`feat:`, `fix:`, etc.)
3. Merge PR to main → Release Please handles the rest!

## Documents

### Primary Process

**[FEATURE_BRANCH_WORKFLOW.md](FEATURE_BRANCH_WORKFLOW.md)** ⭐ NEW
- Complete workflow from feature → release
- Branch naming conventions
- Conventional commit examples
- PR creation and merging
- Multiple features workflow
- Hotfix process
- Best practices
- **Read this first for day-to-day development**

**[AUTOMATED_RELEASE_PROCESS.md](AUTOMATED_RELEASE_PROCESS.md)** ⭐ 
- How Release Please works
- Commit message format
- Version bumping strategy
- Release workflow steps
- Common scenarios
- Troubleshooting
- **Read this to understand the automation**

### Quick Reference

**[QUICK_REFERENCE.md](QUICK_REFERENCE.md)**
- Quick command reference
- One-liner commands for common tasks
- Automated and manual release commands
- Verification commands
- Emergency rollback commands
- Conventional commits cheatsheet

## Workflow Overview

### Daily Development

```bash
# 1. Create feature branch
git checkout -b feature/your-feature

# 2. Develop with conventional commits
git commit -m "feat: add new capability"

# 3. Push and create PR
git push -u origin feature/your-feature
gh pr create --base main --title "Your Feature"

# 4. After review, merge to main
gh pr merge --squash
```

### Automated Release

```bash
# 5. Release Please creates Release PR automatically
gh pr list | grep "chore(main): release"

# 6. Review and merge Release PR
gh pr view <PR-NUMBER>
gh pr merge <PR-NUMBER> --merge

# 7. Release is automatically created!
gh release view
```

## Key Concepts

### Conventional Commits

| Type | Version Bump | Example |
|------|--------------|---------|
| `feat:` | Minor (1.0.0 → 1.1.0) | `feat: add AWS provider` |
| `fix:` | Patch (1.0.0 → 1.0.1) | `fix: handle timeouts` |
| `feat!:` | Major (1.0.0 → 2.0.0) | `feat!: redesign API` |
| `docs:` | Patch | `docs: update guide` |
| `refactor:` | Patch | `refactor: simplify code` |
| `test:` | None | `test: add unit tests` |
| `chore:` | None | `chore: update deps` |

### Branch Strategy

- `main` - Protected, production-ready code
- `feature/*` - New features
- `fix/*` - Bug fixes
- `hotfix/*` - Emergency fixes
- `docs/*` - Documentation updates

### Release Types

**Automated (Recommended):**
- Triggered by merging to main
- Version determined by commit types
- Changelog generated automatically
- Consistent and predictable

**Manual (Emergency Only):**
- Direct git tag creation
- For hotfixes that can't wait
- Bypasses Release Please
- Use sparingly

## Common Tasks

### Create a Feature

```bash
git checkout -b feature/name
git commit -m "feat: description"
git push -u origin feature/name
gh pr create --base main
gh pr merge --squash
```

### Create a Fix

```bash
git checkout -b fix/name
git commit -m "fix: description"
git push -u origin fix/name
gh pr create --base main
gh pr merge --squash
```

### Emergency Hotfix

```bash
git checkout -b hotfix/name
git commit -m "fix: critical issue"
git push -u origin hotfix/name
gh pr create --base main
gh pr merge --squash

# Skip Release Please for immediate release
git tag -a v1.0.1 -m "Hotfix"
git push origin v1.0.1
```

### Monitor Release

```bash
# Check for Release PR
gh pr list | grep "release"

# View release status
gh run list --workflow=release-please.yml

# Verify release
gh release view
```

## Timeline

| Phase | Duration | Automatic? |
|-------|----------|------------|
| Feature development | Hours-Days | Manual |
| PR review | Hours-Days | Manual |
| Merge to main | Seconds | Manual |
| Release PR creation | 30 seconds | ✅ Automatic |
| Release PR review | Minutes | Manual |
| Release creation | 30 seconds | ✅ Automatic |
| Binary builds | 3-5 minutes | ✅ Automatic |
| Homebrew update | 30 seconds | ✅ Automatic |

## Getting Help

**For questions about:**
- Feature branches → Read [FEATURE_BRANCH_WORKFLOW.md](FEATURE_BRANCH_WORKFLOW.md)
- Release Please → Read [AUTOMATED_RELEASE_PROCESS.md](AUTOMATED_RELEASE_PROCESS.md)
- Quick commands → Read [QUICK_REFERENCE.md](QUICK_REFERENCE.md)

## Related Documentation

- [Main Documentation](../README.md)
- [Technical Architecture](../technical/README.md)

---

**Last Updated:** 2025-11-09
**Status:** Ready for v1.0.0
**Workflow:** Feature branches → main → automated release
