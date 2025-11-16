# Public Mirror Sync Workflow

This document describes the hybrid public mirror sync workflow for Stax, which automatically syncs releases from the private development repository to the public distribution repository.

## Repository Architecture

### Private Development Repository
- **Repository**: `Firecrown-Media/stax` (private)
- **Purpose**: Development, issues, planning, sensitive files
- **Access**: Team members only
- **Contents**: Full development history, Claude artifacts, build scripts, all documentation

### Public Distribution Repository
- **Repository**: `Firecrown-Media/stax-public` (public)
- **Purpose**: Release distribution, Homebrew installation
- **Access**: Public
- **Contents**: Release artifacts only, cleaned documentation, no development history

## Workflow Overview

The sync workflow automatically runs when:
1. A new release is published in the private repository
2. Manually triggered via workflow dispatch

### What Gets Synced

**Included:**
- All source code (Go files, cmd/, pkg/, internal/)
- Build configurations (.goreleaser.yml, Makefile)
- Public documentation (docs/*)
- License and README
- Git tag for the release

**Excluded (Cleaned):**
- `.claude/` directory and all Claude artifacts
- `*.claude.md` files
- Build artifacts (`dist/`, binaries)
- `.DS_Store` files
- Cache and temporary files
- Full git history (shallow clone approach)

## Workflow Steps

### 1. Trigger Detection

The workflow determines which tag to sync:
- **On Release**: Uses the release tag from the event
- **Manual Dispatch**: Uses the provided tag or latest release
- **Fallback**: Queries GitHub API for latest release

### 2. Checkout

Checks out the repository at the specific release tag:
```yaml
- uses: actions/checkout@v4
  with:
    ref: ${{ steps.get-tag.outputs.tag }}
    fetch-depth: 0
```

### 3. Clean Sensitive Files

Removes development and sensitive files:
```bash
# Claude AI artifacts
rm -rf .claude/
rm -f .claude.md *.claude.md CLAUDE.md

# Build artifacts
rm -rf dist/
rm -f stax

# macOS artifacts
find . -name ".DS_Store" -type f -delete
```

### 4. Configure SSH

Sets up SSH authentication using the deploy key:
```bash
# Deploy key stored in: secrets.PUBLIC_MIRROR_DEPLOY_KEY
# Configured with write access to stax-public repository
```

### 5. Update README

Copies the public-facing README:
```bash
cp docs/PUBLIC_MIRROR_README.md README.md
```

### 6. Push to Public Mirror

Force pushes to the public repository:
```bash
# Push main branch (current state)
git push public HEAD:main --force

# Push tag
git push public "${TAG}" --force
```

### 7. Verify Sync

Confirms the tag and branch were successfully pushed:
- Checks for tag existence on remote
- Verifies main branch update
- Reports success/failure

### 8. Cleanup

Removes SSH credentials and temporary files.

## Manual Sync

To manually sync a specific tag:

1. Go to Actions tab in GitHub
2. Select "Sync Public Mirror" workflow
3. Click "Run workflow"
4. Enter tag name or leave empty for latest
5. Click "Run workflow" button

## GoReleaser Configuration

The `.goreleaser.yml` has been updated to publish releases directly to the public repository:

```yaml
release:
  github:
    owner: firecrown-media
    name: stax-public  # Changed from 'stax'
```

This ensures:
- Release artifacts are created in the public repository
- Homebrew formula updates reference the public repository
- Users download from the public repository

## Security Considerations

### Deploy Key Setup

The workflow uses a deploy key with limited scope:
1. **Generated**: SSH key pair created specifically for this workflow
2. **Private Key**: Stored as `PUBLIC_MIRROR_DEPLOY_KEY` secret in private repo
3. **Public Key**: Added as deploy key to stax-public with write access
4. **Scope**: Only has access to stax-public repository
5. **Rotation**: Should be rotated periodically

### File Cleaning

The workflow thoroughly cleans sensitive files:
- **Claude Artifacts**: All AI development context removed
- **Build Artifacts**: Binaries and build outputs removed
- **Local Config**: Development configurations excluded
- **Git History**: Only current state synced, not full history

### Force Push Safety

Force pushing is safe in this context because:
- Public repository is a mirror, not collaborative
- No direct development happens in public repository
- Mirror should always match release state exactly
- Tag immutability ensures release integrity

## Troubleshooting

### Sync Failed - Deploy Key Issues

**Symptom**: SSH authentication fails
```
Permission denied (publickey)
```

**Solution**:
1. Verify `PUBLIC_MIRROR_DEPLOY_KEY` secret exists
2. Check deploy key is added to stax-public repository
3. Ensure deploy key has write access enabled
4. Regenerate key pair if necessary

### Sync Failed - Tag Not Found

**Symptom**: Tag doesn't exist in source repository
```
fatal: reference is not a tree
```

**Solution**:
1. Verify tag exists: `git tag -l`
2. Check tag name format (should be v*.*.*)
3. Ensure release was properly created
4. Try manual dispatch with explicit tag name

### Public Repository Out of Sync

**Symptom**: Public repository missing recent releases

**Solution**:
1. Check workflow run history in Actions tab
2. Re-run failed workflows
3. Manual dispatch for specific missing tags
4. Verify GoReleaser configuration

### README Not Updated

**Symptom**: Public repository shows wrong README

**Solution**:
1. Verify `docs/PUBLIC_MIRROR_README.md` exists
2. Check file copy step in workflow logs
3. Manually trigger sync workflow
4. Verify git commit in workflow

## Monitoring

### Workflow Runs

Monitor sync health:
1. Go to Actions tab in private repository
2. Filter by "Sync Public Mirror" workflow
3. Check recent run status
4. Review step-by-step logs for failures

### Public Repository

Verify sync status:
1. Check latest tag: https://github.com/Firecrown-Media/stax-public/tags
2. Check main branch: https://github.com/Firecrown-Media/stax-public
3. Verify releases: https://github.com/Firecrown-Media/stax-public/releases
4. Check file contents match expectations

### Homebrew Formula

Verify Homebrew integration:
1. Check formula repository: https://github.com/Firecrown-Media/homebrew-stax
2. Verify formula references stax-public
3. Test installation: `brew install firecrown-media/stax/stax`
4. Check installed version: `stax --version`

## Maintenance

### Regular Tasks

**Weekly**:
- Review workflow run history
- Check for failed syncs
- Verify public repository is current

**Monthly**:
- Audit file cleaning rules
- Review excluded files list
- Check deploy key expiration

**Quarterly**:
- Rotate deploy key
- Review security practices
- Update documentation

### Deploy Key Rotation

To rotate the deploy key:

1. Generate new SSH key pair:
   ```bash
   ssh-keygen -t ed25519 -C "stax-public-mirror" -f stax-public-deploy-key
   ```

2. Update secret in private repository:
   - Go to Settings > Secrets > Actions
   - Update `PUBLIC_MIRROR_DEPLOY_KEY` with new private key

3. Update deploy key in public repository:
   - Go to Settings > Deploy keys
   - Remove old key
   - Add new public key with write access

4. Test sync:
   - Manually trigger workflow
   - Verify successful push

## Release Process Integration

The mirror sync is integrated into the release process:

1. **Development** (private repo):
   - Feature development
   - Testing
   - Version bump
   - Create release

2. **Automated Sync** (workflow):
   - Triggered on release
   - Clean sensitive files
   - Push to public mirror

3. **GoReleaser** (public repo):
   - Build binaries
   - Create GitHub release
   - Update Homebrew formula

4. **Distribution**:
   - Users install via Homebrew
   - Downloads from public repository
   - No access to private development

## Best Practices

### Do's
- Always test releases before publishing
- Use semantic versioning (v*.*.*)
- Keep public README user-focused
- Monitor workflow runs
- Document sync issues

### Don'ts
- Don't commit secrets to either repository
- Don't sync incomplete features
- Don't manually push to public repository
- Don't expose development artifacts
- Don't share deploy key

## References

- Workflow file: `.github/workflows/sync-public-mirror.yml`
- GoReleaser config: `.goreleaser.yml`
- Public README: `docs/PUBLIC_MIRROR_README.md`
- Private README: `README.md`
- Release workflow: `.github/workflows/release-please.yml`

## Support

For issues with the mirror sync:
1. Check this documentation
2. Review workflow logs
3. Contact DevOps team
4. Create issue in private repository
