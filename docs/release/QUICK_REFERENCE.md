# Release Quick Reference

Quick commands for the Stax release process. For detailed documentation, see [AUTOMATED_RELEASE_PROCESS.md](AUTOMATED_RELEASE_PROCESS.md).

## Automated Release (Recommended)

```bash
# 1. Develop with conventional commits
git commit -m "feat: your feature description"
git push origin main

# 2. Wait for Release Please PR
gh pr list | grep "chore(main): release"

# 3. Review and merge Release PR
gh pr merge <PR-NUMBER> --merge

# Release is automatically created!
```

## Manual Release (Emergency Only)

```bash
# Determine current version
git describe --tags --abbrev=0

# Create and push tag
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

## Local Testing

```bash
# Validate GoReleaser config
make release-check

# Build snapshot locally
make release-snapshot

# Test dry-run
make release-dry-run

# Check current version
make version
```

## Verification

```bash
# Check GitHub release created
gh release view

# Check Homebrew formula updated
brew info Firecrown-Media/stax/stax

# Test installation
brew upgrade stax && stax --version
```

## Conventional Commits

| Type | Version Bump | Example |
|------|--------------|---------|
| `feat:` | Minor (1.0.0 -> 1.1.0) | `feat: add AWS provider` |
| `fix:` | Patch (1.0.0 -> 1.0.1) | `fix: handle timeouts` |
| `feat!:` | Major (1.0.0 -> 2.0.0) | `feat!: redesign API` |
| `docs:` | Patch | `docs: update guide` |
| `refactor:` | Patch | `refactor: simplify code` |
| `test:` | None | `test: add unit tests` |
| `chore:` | None | `chore: update deps` |

## Branch Strategy

- `main` - Protected, production-ready code
- `feature/*` - New features
- `fix/*` - Bug fixes
- `hotfix/*` - Emergency fixes
- `docs/*` - Documentation updates

## Troubleshooting

### Release Failed

```bash
# Delete bad tag
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3

# Delete GitHub release
gh release delete v1.2.3 --yes
```

### Homebrew Not Updated

```bash
# Check secret exists
gh secret list --repo firecrown-media/stax | grep HOMEBREW_TAP_TOKEN

# Check GoReleaser logs in GitHub Actions
```

### Rollback Release

```bash
# Delete tag and release
git push origin :refs/tags/v1.2.3
gh release delete v1.2.3 --repo firecrown-media/stax --yes
```

## Timeline

| Phase | Duration | Automatic? |
|-------|----------|------------|
| Feature development | Hours-Days | Manual |
| PR review | Hours-Days | Manual |
| Merge to main | Seconds | Manual |
| Release PR creation | 30 seconds | Automatic |
| Release PR review | Minutes | Manual |
| Release creation | 30 seconds | Automatic |
| Binary builds | 3-5 minutes | Automatic |
| Homebrew update | 30 seconds | Automatic |

## Related Documentation

- [Automated Release Process](AUTOMATED_RELEASE_PROCESS.md) - Full automation details
- [Feature Branch Workflow](FEATURE_BRANCH_WORKFLOW.md) - Development workflow
- [Release Documentation Overview](README.md) - Release docs hub
