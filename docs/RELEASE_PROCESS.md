# Stax Release Process

## Overview

Stax uses automated releases via GoReleaser and GitHub Actions. Releases are triggered by pushing version tags.

## Versioning

Stax follows [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality (backward compatible)
- **PATCH** version for bug fixes (backward compatible)

## Release Steps

### Option 1: Automated Version Bump (Recommended)

1. Go to GitHub Actions in the repository
2. Select the "Version Bump" workflow
3. Click "Run workflow"
4. Select version bump type:
   - **patch** - Bug fixes (1.0.0 → 1.0.1)
   - **minor** - New features (1.0.0 → 1.1.0)
   - **major** - Breaking changes (1.0.0 → 2.0.0)
5. Click "Run workflow"

The workflow will:
- Calculate the new version number
- Create and push a new git tag
- Automatically trigger the release workflow

### Option 2: Manual Tag

1. Determine the new version number following semantic versioning

2. Create an annotated tag:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   ```

3. Push the tag to trigger the release:
   ```bash
   git push origin v1.2.3
   ```

## What Happens During Release

When a tag is pushed, GitHub Actions automatically:

1. **Runs Tests**
   - Unit tests with race detection
   - Security tests
   - Integration tests

2. **Builds Binaries**
   - macOS (Intel and Apple Silicon)
   - Linux (amd64 and arm64)
   - All binaries are stripped and optimized

3. **Creates Archives**
   - tar.gz archives for each platform
   - Includes README, LICENSE, and documentation

4. **Generates Checksums**
   - SHA256 checksums for all artifacts
   - Published in checksums.txt

5. **Creates GitHub Release**
   - Release notes with changelog
   - Uploads all binaries and archives
   - Tags as prerelease if version contains pre-release identifier

6. **Updates Homebrew Formula**
   - Pushes updated formula to homebrew-tap repository
   - Users can install/update via `brew upgrade stax`

7. **Generates Changelog**
   - Automatically created from git commits
   - Excludes docs, tests, and chore commits

## Release Checklist

Before creating a release, ensure:

- [ ] All tests are passing on main branch
- [ ] Security vulnerabilities have been addressed
- [ ] Documentation is up to date
- [ ] CHANGELOG.md has been updated (if maintained manually)
- [ ] Breaking changes are documented in upgrade guide
- [ ] Version number follows semantic versioning
- [ ] No open critical bugs
- [ ] Code review completed for all changes

## Post-Release Verification

After the release is published:

1. **Verify GitHub Release**
   ```bash
   # Check release page
   open https://github.com/firecrown-media/stax/releases/latest
   ```

2. **Test Homebrew Installation**
   ```bash
   # Update local tap
   brew update

   # Install or upgrade
   brew upgrade stax

   # Verify version
   stax --version
   ```

3. **Test Binary Downloads**
   - Download binaries from GitHub releases
   - Verify checksums match
   - Test binary execution

4. **Verify Homebrew Formula**
   ```bash
   # Check formula was updated
   brew info stax
   ```

## Rollback Procedure

If a release has critical issues:

### 1. Delete the Problematic Tag

```bash
# Delete local tag
git tag -d v1.2.3

# Delete remote tag
git push origin :refs/tags/v1.2.3
```

### 2. Delete GitHub Release

1. Go to GitHub Releases page
2. Find the problematic release
3. Click "Delete" and confirm

### 3. Revert Homebrew Formula

```bash
# Clone homebrew-tap
git clone https://github.com/firecrown-media/homebrew-tap.git
cd homebrew-tap

# Revert the commit
git revert HEAD

# Push the revert
git push origin main
```

### 4. Fix Issues and Re-release

1. Fix the issues in the main branch
2. Create a new patch release (e.g., v1.2.4)
3. Document the fix in release notes

## Troubleshooting

### Release Workflow Failed

**Check the logs:**
```bash
# Go to GitHub Actions
open https://github.com/firecrown-media/stax/actions
```

**Common issues:**

1. **Tests Failed**
   - Fix failing tests
   - Delete tag and re-release

2. **Build Failed**
   - Check Go version compatibility
   - Verify dependencies are available

3. **GoReleaser Configuration Error**
   - Validate .goreleaser.yml syntax
   - Test locally with `goreleaser release --snapshot`

### Homebrew Formula Not Updated

**Check HOMEBREW_TAP_TOKEN:**
```bash
# Verify secret is configured in GitHub
# Settings → Secrets → Actions → HOMEBREW_TAP_TOKEN
```

**Check repository permissions:**
- Token must have `repo` scope
- Token must have write access to homebrew-tap repository

**Manual formula update:**
```bash
# Clone homebrew-tap
git clone https://github.com/firecrown-media/homebrew-tap.git
cd homebrew-tap

# Update Formula/stax.rb manually
# Commit and push
git add Formula/stax.rb
git commit -m "stax: update to v1.2.3"
git push
```

### Binary Doesn't Work

**Test locally before releasing:**
```bash
# Install GoReleaser
brew install goreleaser

# Test release build
goreleaser build --snapshot --clean

# Test the binary
./dist/stax_darwin_amd64_v1/stax --version
```

**Check build flags:**
- Verify ldflags in .goreleaser.yml
- Ensure version variables are set correctly

### Checksums Don't Match

**Verify download integrity:**
```bash
# Download binary
wget https://github.com/firecrown-media/stax/releases/download/v1.2.3/stax_1.2.3_Darwin_x86_64.tar.gz

# Download checksums
wget https://github.com/firecrown-media/stax/releases/download/v1.2.3/checksums.txt

# Verify
shasum -a 256 -c checksums.txt
```

## Testing Releases Locally

Before pushing a tag, test the release process locally:

### 1. Install GoReleaser

```bash
brew install goreleaser
```

### 2. Build Snapshot

```bash
# Build without publishing
goreleaser release --snapshot --clean
```

### 3. Test Binaries

```bash
# Test macOS Intel
./dist/stax_darwin_amd64_v1/stax --version

# Test macOS ARM
./dist/stax_darwin_arm64/stax --version

# Test Linux
./dist/stax_linux_amd64_v1/stax --version
```

### 4. Verify Archives

```bash
# Check archive contents
tar -tzf dist/stax_1.2.3_Darwin_x86_64.tar.gz
```

## Release Cadence

**Recommended schedule:**

- **Patch releases**: As needed for bug fixes (weekly/bi-weekly)
- **Minor releases**: Monthly or when significant features are ready
- **Major releases**: Quarterly or when breaking changes are necessary

## Maintenance Releases

For older versions:

1. Create a maintenance branch:
   ```bash
   git checkout -b v1.x
   git push origin v1.x
   ```

2. Cherry-pick fixes:
   ```bash
   git cherry-pick <commit-hash>
   ```

3. Tag and release:
   ```bash
   git tag -a v1.2.4 -m "Release v1.2.4 (maintenance)"
   git push origin v1.2.4
   ```

## Emergency Releases

For critical security fixes:

1. **Create hotfix branch**
   ```bash
   git checkout -b hotfix/security-fix main
   ```

2. **Apply fix and test thoroughly**
   ```bash
   # Make changes
   make test
   make test-security
   ```

3. **Fast-track release**
   ```bash
   # Merge to main
   git checkout main
   git merge hotfix/security-fix

   # Tag and push immediately
   git tag -a v1.2.4 -m "Security fix"
   git push origin main
   git push origin v1.2.4
   ```

4. **Notify users**
   - Update release notes with security details
   - Send notification to users
   - Update documentation

## Release Notes Template

Use this template for release notes:

```markdown
## Stax v1.2.3

### Highlights
- Major new feature or improvement

### New Features
- Feature 1 description
- Feature 2 description

### Improvements
- Improvement 1
- Improvement 2

### Bug Fixes
- Fix for issue #123
- Fix for issue #456

### Security
- Security fix details (if applicable)

### Breaking Changes
- Breaking change description (if any)

### Installation

**Homebrew (macOS):**
```bash
brew install firecrown-media/stax/stax
```

**Direct Download:**
Download binaries from the [releases page](https://github.com/firecrown-media/stax/releases/tag/v1.2.3)

### Full Changelog
https://github.com/firecrown-media/stax/compare/v1.2.2...v1.2.3
```

## Automation Secrets

Required GitHub secrets:

1. **GITHUB_TOKEN** (automatic)
   - Provided by GitHub Actions
   - Used for creating releases and uploading assets

2. **HOMEBREW_TAP_TOKEN** (manual setup required)
   - Personal Access Token with `repo` scope
   - Add in: Settings → Secrets → Actions
   - Used to push formula updates to homebrew-tap

## Support

For release process questions:
- Check [GitHub Actions logs](https://github.com/firecrown-media/stax/actions)
- Review [GoReleaser documentation](https://goreleaser.com/intro/)
- Contact the development team
