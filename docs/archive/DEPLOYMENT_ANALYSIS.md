# GoReleaser Cross-Repository Release - Root Cause Analysis & Fix

## Executive Summary

**Status:** Configuration is correct. Token permissions need updating.

**Issue:** GoReleaser fails when creating releases in `stax-public` repository
**Root Cause:** `HOMEBREW_TAP_TOKEN` lacks access to `stax-public` repository
**Solution:** Add `stax-public` to token's repository access list
**Time to Fix:** 5 minutes
**Code Changes Required:** None
**Risk Level:** Zero (only updating token permissions)

---

## Detailed Analysis

### Error Message
```
POST https://api.github.com/repos/firecrown-media/stax-public/releases: 
403 Resource not accessible by personal access token
```

### Error Location
- Workflow: `.github/workflows/release-please.yml`
- Step: "Run GoReleaser"
- Phase: Publishing releases
- Timestamp: 2025-11-16T04:07:16Z (most recent failure)

### What's Working
1. ✓ Build phase completes successfully
2. ✓ Archive generation works
3. ✓ Checksum calculation succeeds
4. ✓ Homebrew formula generation succeeds
5. ✓ GoReleaser configuration is valid
6. ✓ Workflow configuration is correct
7. ✓ Token is properly passed to GoReleaser

### What's Failing
1. ✗ Creating GitHub release in `stax-public` repository
2. ✗ Uploading release assets to `stax-public`
3. ✗ (Likely) Updating Homebrew formula in `homebrew-stax` (doesn't get this far)

---

## Configuration Analysis

### File: `.goreleaser.yml`
**Status:** CORRECT - No changes needed

Key sections:
```yaml
# Line 96-99: Release configuration
release:
  github:
    owner: firecrown-media
    name: stax-public  # ← Correct target repository

# Line 70-80: Homebrew configuration
brews:
  - repository:
      owner: Firecrown-Media
      name: homebrew-stax  # ← Correct tap repository
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"  # ← Correct token reference
```

**Analysis:**
- Release target is correctly set to `stax-public`
- Homebrew target is correctly set to `homebrew-stax`
- Token templating is correct
- No configuration changes required

### File: `.github/workflows/release-please.yml`
**Status:** CORRECT - No changes needed

Key section:
```yaml
# Line 50-59: GoReleaser step
- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v6
  env:
    GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}      # For stax-public
    HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }} # For homebrew-stax
```

**Analysis:**
- Both environment variables are set correctly
- GoReleaser will use `GITHUB_TOKEN` for release operations
- GoReleaser will use `HOMEBREW_TAP_TOKEN` for Homebrew operations
- Same token used for both (this is correct and secure)
- No workflow changes required

---

## Token Requirements

### Current State (Likely)
```
HOMEBREW_TAP_TOKEN
└── Repository Access
    └── Firecrown-Media/homebrew-stax ✓
        └── Permissions: Contents (Read/Write) ✓
```

### Required State
```
HOMEBREW_TAP_TOKEN
└── Repository Access
    ├── Firecrown-Media/homebrew-stax ✓
    │   └── Permissions: Contents (Read/Write) ✓
    └── Firecrown-Media/stax-public ← ADD THIS
        └── Permissions: Contents (Read/Write) ← ADD THIS
```

### Why Both Repositories?

1. **stax-public** (line 98-99 in .goreleaser.yml)
   - Purpose: Create GitHub releases
   - Operation: POST to GitHub API to create release
   - Token used: `GITHUB_TOKEN` environment variable
   - Requires: Contents (Read/Write) permission

2. **homebrew-stax** (line 76-80 in .goreleaser.yml)
   - Purpose: Update Homebrew formula
   - Operation: Commit and push to repository
   - Token used: `HOMEBREW_TAP_TOKEN` template variable
   - Requires: Contents (Read/Write) permission

---

## Solution Steps

### 1. Navigate to Token Settings
URL: https://github.com/settings/tokens?type=beta

### 2. Edit Token
- Find your `HOMEBREW_TAP_TOKEN` (or the token name you used)
- Click "Edit"

### 3. Update Repository Access
In "Repository access" section:
- Type: "Only select repositories"
- Repositories:
  - [x] Firecrown-Media/homebrew-stax (existing)
  - [x] Firecrown-Media/stax-public (ADD THIS)

### 4. Verify Permissions
For BOTH repositories, ensure:
- Contents: Read and write
- Metadata: Read-only (automatic)

DO NOT grant:
- Administration
- Actions
- Code scanning
- Commit statuses
- Deployments
- Issues
- Pull requests
- Any other permissions

### 5. Save Token
- Click "Update token"
- If you regenerated the token, copy the new value

### 6. Update GitHub Secret (if token regenerated)
Only if you regenerated the token:
1. Go to: https://github.com/Firecrown-Media/stax/settings/secrets/actions
2. Edit: `HOMEBREW_TAP_TOKEN`
3. Paste: New token value
4. Save

---

## Verification Plan

### Immediate Verification
Cannot test immediately without triggering a release, but can verify token access:

```bash
# Verify token has access to stax-public (requires actual token)
gh api repos/Firecrown-Media/stax-public -H "Authorization: token TOKEN_HERE"
```

### Next Release Verification

After next release (automatic or manual), verify:

```bash
# 1. Check release was created in stax-public
gh release list --repo Firecrown-Media/stax-public --limit 1

# 2. Check release has assets
gh release view LATEST_VERSION --repo Firecrown-Media/stax-public --json assets

# 3. Check formula was updated in homebrew-stax
gh api repos/Firecrown-Media/homebrew-stax/commits?path=Formula/stax.rb&per_page=1

# 4. Test installation works
brew uninstall stax 2>/dev/null || true
brew install firecrown-media/stax/stax
stax --version
```

### Success Criteria

A successful release should:
1. ✓ Build all platform binaries
2. ✓ Create archives
3. ✓ Generate checksums
4. ✓ Create release in `stax-public` with tag
5. ✓ Upload 4 platform archives to release
6. ✓ Upload checksums.txt to release
7. ✓ Upload README, LICENSE, docs to release
8. ✓ Commit updated formula to `homebrew-stax`
9. ✓ Make release publicly available

---

## Risk Assessment

### Changes Required
- Update token permissions: Low risk
- Code changes: None required
- Configuration changes: None required

### Rollback Plan
If issues occur:
- Revert token permissions to previous state
- Investigate error logs
- No code/config rollback needed (nothing changed)

### Testing Impact
- No impact on development workflow
- No impact on existing releases
- Only affects future releases

---

## Additional Findings

### GoReleaser Version
- Using: `goreleaser-action@v6`
- GoReleaser version: `v2.12.7`
- Status: Latest stable version ✓

### Deprecation Warnings
1. `archives.format_overrides.format` - cosmetic, still works
2. `brews` → `homebrew_casks` - cosmetic, still works

These are low-priority and don't affect functionality. Can be addressed in future updates.

### Workflow Permissions
```yaml
permissions:
  contents: write
  pull-requests: write
```

These are correct for release-please but don't grant cross-repo access (hence need for PAT).

---

## Documentation Created

1. **GORELEASER_FIX.md** - Comprehensive explanation of issue and fix
2. **TOKEN_PERMISSIONS_CHECKLIST.md** - Step-by-step fix checklist
3. **RELEASE_FLOW_DIAGRAM.md** - Visual flow diagram and architecture
4. **QUICK_FIX_GUIDE.md** - 5-minute quick reference
5. **DEPLOYMENT_ANALYSIS.md** - This document (detailed analysis)

---

## Conclusion

**No code or configuration changes are required.** 

The GoReleaser configuration and GitHub Actions workflow are correctly set up for cross-repository releases. The only issue is that the `HOMEBREW_TAP_TOKEN` needs repository access to both `stax-public` and `homebrew-stax`.

**Next Action:** Update token permissions as outlined above.

**Expected Outcome:** Next release will successfully create releases in `stax-public` and update formulas in `homebrew-stax`.

**Time to Resolution:** 5 minutes (token update) + next release cycle (automatic)
