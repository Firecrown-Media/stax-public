# GoReleaser Cross-Repository Release Fix

## Problem Diagnosis

### Error Message
```
POST https://api.github.com/repos/firecrown-media/stax-public/releases: 403 Resource not accessible by personal access token
```

### Root Cause
The `HOMEBREW_TAP_TOKEN` Personal Access Token (PAT) lacks the required permissions to create releases in the `Firecrown-Media/stax-public` repository.

### Why This Happens
GoReleaser performs two distinct operations that require different repository access:

1. **Homebrew Formula Update** → `Firecrown-Media/homebrew-stax` repository
   - Configured in `.goreleaser.yml` under `brews.repository`
   - Uses `HOMEBREW_TAP_TOKEN` explicitly via `token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"`

2. **GitHub Release Creation** → `Firecrown-Media/stax-public` repository
   - Configured in `.goreleaser.yml` under `release.github`
   - Uses `GITHUB_TOKEN` environment variable (currently set to `HOMEBREW_TAP_TOKEN`)

The current token was likely created with permissions only for `homebrew-stax`, not `stax-public`.

## Solution: Update Token Permissions

### Step 1: Update the Fine-Grained Personal Access Token

Go to GitHub Settings → Developer Settings → Personal Access Tokens → Fine-grained tokens → `HOMEBREW_TAP_TOKEN`

**Required Repository Access:**
- `Firecrown-Media/homebrew-stax` (existing)
- `Firecrown-Media/stax-public` (ADD THIS)

**Required Permissions for BOTH repositories:**

| Permission | Access Level | Purpose |
|------------|--------------|---------|
| **Contents** | Read and write | Create releases, push formula updates |
| **Metadata** | Read-only | Repository metadata (automatic) |

### Step 2: Verify Token Configuration

The token should have:
- Expiration: Set appropriately (recommend 90 days or 1 year)
- Repository access: Only selected repositories
  - `Firecrown-Media/homebrew-stax`
  - `Firecrown-Media/stax-public`

### Step 3: Update GitHub Secret (if token was regenerated)

If you regenerated the token:
1. Go to `Firecrown-Media/stax` → Settings → Secrets and variables → Actions
2. Update `HOMEBREW_TAP_TOKEN` with the new token value

## Configuration Verification

### Current GoReleaser Configuration (Correct)

The `.goreleaser.yml` is correctly configured:

```yaml
# Homebrew formula publishing
brews:
  - name: stax
    repository:
      owner: Firecrown-Media
      name: homebrew-stax
      branch: main
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"  # Uses HOMEBREW_TAP_TOKEN

# GitHub release publishing
release:
  github:
    owner: firecrown-media
    name: stax-public  # Releases go to stax-public, not stax
```

### Current Workflow Configuration (Correct)

The `.github/workflows/release-please.yml` is correctly configured:

```yaml
- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v6
  with:
    args: release --clean
  env:
    GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}      # Used for stax-public releases
    HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }} # Used for homebrew-stax formula
```

## Why This Approach Works

1. **Single Token, Multiple Repositories:**
   - The `HOMEBREW_TAP_TOKEN` is configured with access to both repositories
   - GoReleaser uses `GITHUB_TOKEN` env var for release operations
   - GoReleaser uses `HOMEBREW_TAP_TOKEN` from config for Homebrew operations

2. **No Configuration Changes Needed:**
   - The `.goreleaser.yml` and workflow files are already correct
   - Only the token permissions need updating

3. **Security Best Practice:**
   - Fine-grained PAT with minimal scope (only 2 specific repositories)
   - Read/write access only to Contents (not admin, code, or other permissions)

## Testing the Fix

After updating token permissions:

1. **Local Validation:**
   ```bash
   goreleaser check --config .goreleaser.yml
   ```

2. **Test Release:**
   - Create a test commit and push to main
   - Release-please will create a PR
   - Merge the PR to trigger release
   - Monitor workflow: `gh run watch`

3. **Verify Success:**
   - Check release created in `Firecrown-Media/stax-public`
   - Check formula updated in `Firecrown-Media/homebrew-stax`

## Alternative Solutions (Not Recommended)

### Option A: Use Classic PAT Instead of Fine-Grained
Classic PATs have broader permissions and work across all repositories in an organization. However, they're less secure and GitHub is phasing them out.

### Option B: Use Two Separate Tokens
Create separate tokens for releases vs. Homebrew:
- `GITHUB_TOKEN` → for stax-public releases
- `HOMEBREW_TAP_TOKEN` → for homebrew-stax formula

This adds complexity without security benefits since both need similar permissions.

### Option C: Use GitHub Actions GITHUB_TOKEN
The default `GITHUB_TOKEN` provided by GitHub Actions cannot access other repositories, so this won't work for cross-repo releases.

## Summary

**The Fix:** Update the `HOMEBREW_TAP_TOKEN` fine-grained PAT to include `Firecrown-Media/stax-public` repository access with "Contents: Read and write" permissions.

**Why It Works:** GoReleaser needs to perform two operations:
1. Create releases in `stax-public` (uses `GITHUB_TOKEN` env var)
2. Update formula in `homebrew-stax` (uses `HOMEBREW_TAP_TOKEN` from config)

By giving the token access to both repositories and using it for both operations, we maintain security while enabling cross-repository automation.

**No Code Changes Required:** The configuration files are already correct. Only the token permissions need updating.
