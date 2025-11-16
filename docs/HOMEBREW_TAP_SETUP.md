# Homebrew Tap Setup Guide

This guide walks through setting up the `homebrew-tap` repository for automated Homebrew formula distribution.

## Overview

The Homebrew tap (`firecrown-media/homebrew-tap`) is a separate Git repository that contains the Homebrew formula for Stax. When you run a release, GoReleaser automatically updates the formula in this repository.

## Repository Setup

### Step 1: Create the Tap Repository

1. **Create a new GitHub repository**:
   - Repository name: `homebrew-tap`
   - Owner: `firecrown-media`
   - Description: "Homebrew tap for Firecrown Media tools"
   - Visibility: Public
   - Initialize with README: Yes

2. **Clone the repository locally**:
   ```bash
   git clone https://github.com/firecrown-media/homebrew-tap.git
   cd homebrew-tap
   ```

### Step 2: Create Directory Structure

Create the required directory structure:

```bash
# Create Formula directory
mkdir -p Formula

# Create initial structure
touch Formula/.gitkeep
```

### Step 3: Create Initial README

Create or update `README.md`:

```markdown
# Firecrown Media Homebrew Tap

Official Homebrew tap for Firecrown Media tools.

## Installation

### Stax

Stax is a powerful CLI tool for WordPress multisite development workflows.

**Install:**
```bash
brew tap firecrown-media/tap
brew install stax
```

**Update:**
```bash
brew update
brew upgrade stax
```

**Uninstall:**
```bash
brew uninstall stax
brew untap firecrown-media/tap
```

## Available Formulas

- **stax** - WordPress multisite development CLI tool

## Support

- [Stax Documentation](https://github.com/firecrown-media/stax)
- [Report Issues](https://github.com/firecrown-media/stax/issues)

## About

This tap is maintained by Firecrown Media and automatically updated by GoReleaser during releases.
```

### Step 4: Commit and Push

```bash
git add .
git commit -m "Initial tap setup"
git push origin main
```

## GitHub Token Setup

GoReleaser needs a GitHub Personal Access Token to push formula updates to the tap repository.

### Step 1: Create Personal Access Token

1. Go to GitHub Settings: https://github.com/settings/tokens

2. Click "Generate new token (classic)"

3. Configure the token:
   - **Note**: "Stax Homebrew Tap Updates"
   - **Expiration**: No expiration (or long duration)
   - **Scopes**: Select:
     - `repo` (Full control of private repositories)
       - `repo:status`
       - `repo_deployment`
       - `public_repo`
       - `repo:invite`
       - `security_events`

4. Click "Generate token"

5. **IMPORTANT**: Copy the token immediately - you won't see it again!

Example token format: `ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

### Step 2: Add Token to Stax Repository Secrets

1. Go to the Stax repository settings:
   ```
   https://github.com/firecrown-media/stax/settings/secrets/actions
   ```

2. Click "New repository secret"

3. Add the secret:
   - **Name**: `HOMEBREW_TAP_TOKEN`
   - **Value**: (paste the token you just created)

4. Click "Add secret"

### Step 3: Verify Token Permissions

The token must have:
- Write access to `firecrown-media/homebrew-tap`
- Ability to create commits and push to the repository
- Access to public repositories at minimum

## Initial Formula Creation

GoReleaser will create the formula automatically on first release, but you can create an initial version manually.

### Manual Initial Formula (Optional)

Create `Formula/stax.rb`:

```ruby
# typed: false
# frozen_string_literal: true

class Stax < Formula
  desc "Powerful CLI tool for WordPress multisite development workflows"
  homepage "https://github.com/firecrown-media/stax"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/firecrown-media/stax/releases/download/v0.1.0/stax_0.1.0_Darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER"
    end
    if Hardware::CPU.intel?
      url "https://github.com/firecrown-media/stax/releases/download/v0.1.0/stax_0.1.0_Darwin_x86_64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  depends_on "ddev" => :optional
  depends_on "docker" => :optional

  def install
    bin.install "stax"
  end

  test do
    system "#{bin}/stax", "--version"
  end
end
```

**Note**: This is just a placeholder. GoReleaser will replace it with the correct version, URLs, and checksums on the first release.

## Testing the Tap

### Test Locally

Before the first release, test that users can add the tap:

```bash
# Add the tap (won't have formulas yet)
brew tap firecrown-media/tap

# Verify it was added
brew tap | grep firecrown-media
```

### Test Formula Installation (After First Release)

After the first release is published:

```bash
# Update Homebrew
brew update

# Install Stax
brew install firecrown-media/stax/stax

# Verify installation
stax --version

# Check formula info
brew info stax
```

## Automated Updates

Once set up, GoReleaser automatically handles formula updates:

### What Happens on Each Release

1. **Tag is pushed** to firecrown-media/stax
2. **GitHub Actions triggers** release workflow
3. **GoReleaser runs** and:
   - Builds binaries for all platforms
   - Creates GitHub release with assets
   - Calculates SHA256 checksums
   - Updates `Formula/stax.rb` in homebrew-tap
   - Commits changes with message: "stax: update to vX.Y.Z"
   - Pushes to homebrew-tap repository

### Formula Update Example

When v1.2.3 is released, GoReleaser creates this commit in homebrew-tap:

```
commit abc123...
Author: goreleaser[bot]
Date:   Mon Jan 1 12:00:00 2024 +0000

    stax: update to v1.2.3
```

The formula will be updated with:
- New version number
- New download URLs
- New SHA256 checksums
- Any dependency changes

## Maintenance

### Monitoring

Monitor the tap repository for:

1. **Successful updates**: Check that each release creates a commit
2. **Formula validation**: Homebrew may report issues
3. **User issues**: Check for installation problems

### Manual Formula Fixes

If you need to manually fix the formula:

```bash
# Clone the tap
git clone https://github.com/firecrown-media/homebrew-tap.git
cd homebrew-tap

# Edit the formula
nano Formula/stax.rb

# Test locally
brew install --build-from-source Formula/stax.rb

# Commit and push
git add Formula/stax.rb
git commit -m "stax: fix formula issue"
git push origin main
```

### Validate Formula

Use Homebrew's audit tools:

```bash
# Audit the formula
brew audit --strict firecrown-media/stax/stax

# Test installation
brew test firecrown-media/stax/stax

# Test style
brew style firecrown-media/stax/stax
```

## Troubleshooting

### GoReleaser Can't Push to Tap

**Symptoms**: Release succeeds but formula isn't updated.

**Checks**:
1. Verify `HOMEBREW_TAP_TOKEN` secret exists in stax repository
2. Check token has `repo` scope
3. Verify token hasn't expired
4. Check GoReleaser logs in GitHub Actions

**Solution**:
```bash
# Create new token with repo scope
# Update HOMEBREW_TAP_TOKEN secret
# Re-run release workflow
```

### Formula Installation Fails

**Symptoms**: Users can't install via Homebrew.

**Checks**:
1. Test formula syntax: `brew audit firecrown-media/stax/stax`
2. Verify URLs are accessible
3. Check SHA256 checksums match
4. Test installation locally

**Solution**:
```bash
# Download the binary
curl -L https://github.com/firecrown-media/stax/releases/download/v1.2.3/stax_1.2.3_Darwin_x86_64.tar.gz -o stax.tar.gz

# Calculate correct checksum
shasum -a 256 stax.tar.gz

# Update formula with correct checksum
# Commit and push
```

### Wrong Version in Formula

**Symptoms**: Formula shows old version after release.

**Checks**:
1. Check GoReleaser successfully ran
2. Verify commit was pushed to homebrew-tap
3. Check users ran `brew update`

**Solution**:
```bash
# Users should run:
brew update
brew upgrade stax
```

### Multiple Tap Locations

**Symptoms**: Users have the tap at different URLs.

**Note**: Homebrew taps must follow the naming convention:
- Repository: `homebrew-tap` (or `homebrew-<name>`)
- Tap name: `firecrown-media/tap` (or `firecrown-media/<name>`)

**Correct usage**:
```bash
brew tap firecrown-media/tap  # Correct
brew tap firecrown-media/homebrew-tap  # Also works but verbose
```

## Security Considerations

### Token Security

- **Never commit** the `HOMEBREW_TAP_TOKEN` to any repository
- **Rotate tokens** periodically (recommended: annually)
- **Use minimal permissions**: Only `repo` scope needed
- **Monitor usage**: Check GitHub token usage regularly

### Formula Security

- **Review commits**: Monitor automated commits to homebrew-tap
- **Verify checksums**: Ensure SHA256 checksums match releases
- **Audit dependencies**: Keep formula dependencies up to date
- **Sign releases**: Consider signing binaries (future enhancement)

## Advanced Configuration

### Multiple Tools in One Tap

To add more tools to the tap:

```bash
# Add new formula
touch Formula/another-tool.rb

# GoReleaser in another-tool project would target same tap
# Each tool gets its own formula file
```

### Custom Tap Name

To use a different tap name:

1. Repository name: `homebrew-<custom-name>`
2. Users tap with: `brew tap firecrown-media/<custom-name>`

Update `.goreleaser.yml`:
```yaml
brews:
  - name: stax
    repository:
      owner: firecrown-media
      name: homebrew-<custom-name>  # Change here
```

### Private Tap

For private formulas:

1. Make homebrew-tap repository private
2. Users install with authentication:
   ```bash
   brew tap firecrown-media/tap https://github.com/firecrown-media/homebrew-tap
   ```

## Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
- [GoReleaser Homebrew Documentation](https://goreleaser.com/customization/homebrew/)
- [GitHub Personal Access Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)

## Summary Checklist

Setup checklist for new tap:

- [ ] Create `homebrew-tap` repository on GitHub
- [ ] Add README explaining the tap
- [ ] Create `Formula/` directory
- [ ] Generate GitHub Personal Access Token with `repo` scope
- [ ] Add `HOMEBREW_TAP_TOKEN` secret to stax repository
- [ ] Test tap can be added: `brew tap firecrown-media/tap`
- [ ] Create first release to test automation
- [ ] Verify formula was auto-created
- [ ] Test installation: `brew install stax`
- [ ] Document for users

---

**The tap repository should require minimal manual maintenance once properly configured.**
