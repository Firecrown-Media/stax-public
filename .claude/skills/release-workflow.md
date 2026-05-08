# Release Workflow Skill

This skill automates the complete release workflow for stax, ensuring consistent release practices, proper versioning, and documentation maintenance.

## Purpose

Use this skill whenever making code changes that need to be released. It ensures:
- Feature branch workflow is followed
- Release-please handles versioning automatically
- Homebrew formula is updated
- Public mirror is synced
- Documentation is updated and organized
- Redundant or outdated information is removed

## When to Use

Invoke this skill when:
- Implementing new features
- Fixing bugs
- Making any code changes that need to be released
- Updating dependencies
- Refactoring code

## Workflow Steps

### 1. Pre-Development Planning

Before writing code:
- Identify what needs to be changed
- Determine the type of change (feat/fix/chore/docs)
- Check for related documentation that may need updates

### 2. Feature Branch Creation

```bash
# Create a descriptive feature branch
git checkout -b <type>/<description>

# Examples:
git checkout -b feat/add-search-command
git checkout -b fix/database-timeout-error
git checkout -b chore/update-dependencies
```

Branch naming convention:
- `feat/` - New features
- `fix/` - Bug fixes
- `chore/` - Maintenance tasks
- `docs/` - Documentation updates
- `refactor/` - Code refactoring

### 3. Implementation

- Make code changes
- Run tests: `go test ./...`
- Build to verify: `go build`
- Run linters if configured: `golangci-lint run`

### 4. Documentation Updates

Check and update these files as needed:

**Required for features:**
- `README.md` - User-facing documentation
- `docs/GETTING_STARTED.md` - Quick start guide
- `docs/IMPLEMENTATION_ROADMAP.md` - Track completed phases
- Command help text in `cmd/*.go` files

**Optional:**
- `CHANGELOG.md` - Manual changes (release-please handles most)
- Example configurations
- Architecture diagrams

**Documentation Cleanup:**
- Remove outdated information
- Consolidate duplicate content
- Fix broken links
- Update version numbers
- Remove completed TODOs

### 5. Commit with Conventional Commits

Use conventional commit format for release-please automation:

```bash
git add <files>
git commit -m "<type>: <description>

<optional body>

<optional footer>
"
```

**Commit types:**
- `feat:` - New feature (triggers MINOR version bump: 1.2.0 → 1.3.0)
- `fix:` - Bug fix (triggers PATCH version bump: 1.2.0 → 1.2.1)
- `chore:` - Maintenance (no version bump)
- `docs:` - Documentation only (no version bump)
- `refactor:` - Code refactoring (no version bump)
- `test:` - Test changes (no version bump)
- `perf:` - Performance improvements (triggers PATCH)

**Breaking changes:**
- Add `BREAKING CHANGE:` in footer (triggers MAJOR version bump: 1.2.0 → 2.0.0)
- Or add `!` after type: `feat!:` or `fix!:`

**Examples:**
```bash
# Feature (minor bump)
git commit -m "feat: add search command for finding files

Implements grep-based search across WordPress files with
support for regex patterns and file type filtering.

Related: #45"

# Bug fix (patch bump)
git commit -m "fix: resolve database connection timeout

Increases timeout from 30s to 60s and adds retry logic
for transient network errors.

Fixes: #123"

# Breaking change (major bump)
git commit -m "feat!: migrate to new config format

BREAKING CHANGE: .stax.yml format has changed.
Old configs need to be migrated with 'stax config migrate'

Related: #67"
```

### 6. Create Pull Request

```bash
# Push feature branch
git push -u origin <branch-name>

# Create PR with detailed description
gh pr create --title "<type>: <description>" --body "<detailed-body>"
```

**PR Description Template:**
```markdown
## Summary
Brief overview of changes

## Problem
What issue does this solve?

## Solution
How does this change solve the problem?

## Changes
- List specific changes made
- File-by-file if major refactor

## Testing
- ✅ Tests pass
- ✅ Code compiles
- ✅ Manual testing completed

## Documentation
- ✅ README updated
- ✅ Help text updated
- ✅ Examples added/updated

## Breaking Changes
List any breaking changes (if applicable)
```

### 7. Merge and Release

```bash
# Option 1: Auto-merge when checks pass
gh pr merge <pr-number> --squash --auto

# Option 2: Manual merge
gh pr merge <pr-number> --squash
```

**After merge:**
1. Release-please automatically creates/updates a release PR
2. Merge the release PR (it will have title like "chore(main): release X.Y.Z")
3. Release workflow automatically:
   - Creates GitHub release
   - Builds binaries with GoReleaser
   - Syncs to public mirror (stax-public)
   - Updates Homebrew formula

### 8. Verify Release

```bash
# Check release was created
gh release list --limit 3

# Verify public mirror
curl -I https://github.com/Firecrown-Media/stax-public/releases/tag/vX.Y.Z

# Verify Homebrew formula updated
curl https://raw.githubusercontent.com/Firecrown-Media/homebrew-stax/main/Formula/stax.rb | grep "version"

# Update local installation
brew update && brew upgrade stax
stax --version
```

### 9. Post-Release Documentation Review

After each release, review and update:

**Documentation Organization:**
- [ ] Remove completed roadmap items
- [ ] Archive old documentation
- [ ] Update version numbers in examples
- [ ] Check for broken internal links
- [ ] Consolidate duplicate information
- [ ] Update screenshots/examples if UI changed

**Cleanup Tasks:**
```bash
# Remove merged feature branches
git branch -d <branch-name>
git push origin --delete <branch-name>

# Update local main
git checkout main && git pull

# Clean up old releases if needed
gh release list --limit 20
```

## Troubleshooting

### Release-please PR not created
- Check commit message format (must use conventional commits)
- Ensure commits are on `main` branch
- Wait a few minutes for GitHub Actions to run

### Homebrew formula not updated
- Check workflow runs: `gh run list --workflow=update-homebrew-tap.yml`
- Verify release published: `gh release view vX.Y.Z`
- Manually update if needed (see docs)

### Public mirror not synced
- Check workflow runs: `gh run list --workflow=sync-public-mirror.yml`
- Verify PAT token is valid
- Manually trigger if needed

### Build failed
- Check GoReleaser config: `.goreleaser.yml`
- Verify Go version matches
- Check for compilation errors: `go build`

## Best Practices

1. **Small, focused changes** - One feature/fix per PR
2. **Meaningful commit messages** - Future you will thank you
3. **Test before pushing** - Run `go test ./...` and `go build`
4. **Update docs with code** - Don't leave docs for later
5. **Review your own PR** - Catch issues before others see them
6. **Keep branches up to date** - Rebase or merge main regularly
7. **Clean up after release** - Delete merged branches

## Documentation Maintenance Checklist

For each release, ensure:

- [ ] **README.md** - Updated with new features/changes
- [ ] **CHANGELOG.md** - Auto-generated by release-please
- [ ] **Command help** - Updated in `cmd/*.go` files
- [ ] **Examples** - Working and up-to-date
- [ ] **Roadmap** - Mark completed items
- [ ] **Getting Started** - Reflects current workflow
- [ ] **No broken links** - All references valid
- [ ] **No duplicate info** - Consolidated where possible
- [ ] **No outdated info** - Removed or archived
- [ ] **Version numbers** - Updated in examples

## Quick Reference

```bash
# Create feature branch
git checkout -b feat/my-feature

# Make changes, commit with conventional commits
git commit -m "feat: add new feature"

# Push and create PR
git push -u origin feat/my-feature
gh pr create --title "feat: add new feature" --body "Description"

# Merge PR (triggers release-please)
gh pr merge <number> --squash

# Merge release-please PR (triggers release)
gh pr merge <number> --squash

# Verify and upgrade
gh release list --limit 3
brew update && brew upgrade stax
stax --version
```

## Related Files

- `.github/workflows/release-please.yml` - Release automation
- `.github/workflows/update-homebrew-tap.yml` - Homebrew updates
- `.github/workflows/sync-public-mirror.yml` - Public mirror sync
- `.goreleaser.yml` - Binary build configuration
- `release-please-config.json` - Release-please config
- `.release-please-manifest.json` - Version tracking

## Success Criteria

A successful release includes:
- ✅ Feature branch merged to main
- ✅ Release-please PR created and merged
- ✅ GitHub release published
- ✅ Binaries built and uploaded
- ✅ Public mirror synced
- ✅ Homebrew formula updated
- ✅ Local installation upgraded
- ✅ Documentation updated
- ✅ No outdated/duplicate docs
- ✅ All tests passing
- ✅ Version verified

---

**Last Updated:** 2025-11-18 (v2.14.3)
