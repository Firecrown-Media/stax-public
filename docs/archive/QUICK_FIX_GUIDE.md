# Quick Fix Guide - GoReleaser 403 Error

## The Problem

```
Error: POST https://api.github.com/repos/firecrown-media/stax-public/releases: 
       403 Resource not accessible by personal access token
```

## The Cause

Your `HOMEBREW_TAP_TOKEN` fine-grained PAT has access to `homebrew-stax` but NOT to `stax-public`.

## The Fix (5 minutes)

### Step 1: Edit Token
1. Go to: https://github.com/settings/tokens?type=beta
2. Click on your `HOMEBREW_TAP_TOKEN`
3. Click "Edit"

### Step 2: Add stax-public Repository
In the "Repository access" section:
- Find "Selected repositories" dropdown
- Click "Select repositories"
- Add: `Firecrown-Media/stax-public`
- Keep: `Firecrown-Media/homebrew-stax`

### Step 3: Verify Permissions
For BOTH repositories, ensure:
- Contents: Read and write ✓
- Metadata: Read-only ✓ (automatic)

### Step 4: Save
- Click "Update token" at bottom
- Done! (No need to update GitHub secrets unless you regenerated the token)

## Test the Fix

Wait for next release or manually verify:

```bash
# Check if token can access stax-public (replace YOUR_TOKEN)
gh api repos/Firecrown-Media/stax-public -H "Authorization: token YOUR_TOKEN"

# Should return repo info, not 403/404
```

## Why This Works

GoReleaser needs to:
1. Create releases in `stax-public` → Uses `GITHUB_TOKEN` env var (set to `HOMEBREW_TAP_TOKEN`)
2. Update formula in `homebrew-stax` → Uses `HOMEBREW_TAP_TOKEN` from config

Same token, two different uses, needs access to both repos.

## No Code Changes Needed

Your `.goreleaser.yml` and `.github/workflows/release-please.yml` are already correct!

## Verification After Next Release

```bash
# Check release was created
gh release list --repo Firecrown-Media/stax-public --limit 1

# Check formula was updated
gh api repos/Firecrown-Media/homebrew-stax/commits?path=Formula/stax.rb&per_page=1
```

## Still Having Issues?

Check:
- [ ] Token hasn't expired
- [ ] Repository names are exact: `stax-public` (not `Stax-Public`)
- [ ] Token saved as GitHub secret `HOMEBREW_TAP_TOKEN`
- [ ] You have admin access to both repositories
- [ ] Token is fine-grained PAT (not classic)

## Summary

**Problem:** Token permissions
**Solution:** Add `stax-public` to token's repository access
**Time:** 5 minutes
**Code changes:** None required
**Configuration changes:** None required
