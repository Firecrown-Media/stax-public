# Token Permissions Checklist

## Current Situation

**Error:** `403 Resource not accessible by personal access token`

**Reason:** The HOMEBREW_TAP_TOKEN doesn't have permission to create releases in `Firecrown-Media/stax-public`

## Fix Checklist

### [ ] 1. Navigate to GitHub Token Settings

1. Go to: https://github.com/settings/tokens?type=beta
2. Find: `HOMEBREW_TAP_TOKEN` (or whatever you named it)
3. Click: "Edit" or "Regenerate token" if needed

### [ ] 2. Update Repository Access

**Current Repository Access (likely):**
- ✓ Firecrown-Media/homebrew-stax

**Required Repository Access:**
- ✓ Firecrown-Media/homebrew-stax
- ✓ Firecrown-Media/stax-public ← **ADD THIS**

### [ ] 3. Verify Permissions for BOTH Repositories

For each repository, ensure these permissions:

| Permission Category | Access Level | Status |
|--------------------|--------------|--------|
| Contents | Read and write | [ ] |
| Metadata | Read-only | [ ] (automatic) |

**DO NOT** grant these (not needed):
- ❌ Administration
- ❌ Actions
- ❌ Deployments
- ❌ Environments
- ❌ Issues
- ❌ Pull requests
- ❌ Webhooks

### [ ] 4. Save Token Settings

- Click "Update token" or "Generate token"
- If regenerated, copy the new token value

### [ ] 5. Update GitHub Secret (if token was regenerated)

1. Go to: https://github.com/Firecrown-Media/stax/settings/secrets/actions
2. Edit: `HOMEBREW_TAP_TOKEN`
3. Paste: New token value
4. Click: "Update secret"

### [ ] 6. Test the Fix

Trigger a test release:

```bash
# Option A: Wait for next release
# Release-please will create a PR, merge it to trigger release

# Option B: Manually trigger (if workflow supports)
gh workflow run release-please.yml

# Option C: Monitor next release
gh run list --workflow=release-please.yml --limit 1
gh run watch  # Watch the latest run
```

### [ ] 7. Verify Success

**Check Release Created:**
```bash
gh release list --repo Firecrown-Media/stax-public --limit 1
```

**Check Formula Updated:**
```bash
gh api repos/Firecrown-Media/homebrew-stax/contents/Formula/stax.rb | jq -r '.content' | base64 -d | head -20
```

## Quick Verification Commands

**Test token has access to stax-public:**
```bash
# This should return repository info, not 403/404
gh api repos/Firecrown-Media/stax-public -H "Authorization: token YOUR_TOKEN_HERE"
```

**Check current releases in stax-public:**
```bash
gh release list --repo Firecrown-Media/stax-public
```

**Check current formula in homebrew-stax:**
```bash
gh repo view Firecrown-Media/homebrew-stax
```

## Expected Outcome

After fixing token permissions, the release workflow should:

1. ✓ Build binaries for all platforms
2. ✓ Create archives
3. ✓ Generate checksums
4. ✓ **Create release in Firecrown-Media/stax-public** ← Currently failing here
5. ✓ Update formula in Firecrown-Media/homebrew-stax
6. ✓ Upload artifacts

## Token Scope Summary

```
HOMEBREW_TAP_TOKEN (Fine-grained PAT)
├── Repository Access
│   ├── Firecrown-Media/homebrew-stax
│   │   └── Permissions: Contents (Read/Write)
│   └── Firecrown-Media/stax-public ← MUST ADD THIS
│       └── Permissions: Contents (Read/Write)
└── Expiration: [Set appropriately]
```

## Workflow Flow

```
GitHub Actions Workflow (release-please.yml)
├── GITHUB_TOKEN = secrets.HOMEBREW_TAP_TOKEN
├── HOMEBREW_TAP_TOKEN = secrets.HOMEBREW_TAP_TOKEN
└── Runs GoReleaser
    ├── Release to stax-public (uses GITHUB_TOKEN env var)
    │   └── Requires: Contents (Read/Write) on stax-public
    └── Update Homebrew formula (uses HOMEBREW_TAP_TOKEN from config)
        └── Requires: Contents (Read/Write) on homebrew-stax
```

## Troubleshooting

**If still getting 403:**
- Verify token hasn't expired
- Check token is for correct GitHub account/organization
- Ensure repository names are exact (case-sensitive)
- Verify token was saved to GitHub secret correctly

**If getting different error:**
- Check workflow logs: `gh run view --log-failed`
- Validate GoReleaser config: `goreleaser check`
- Check repository visibility (public vs private)

**If Homebrew update fails but release succeeds:**
- Token has access to stax-public but not homebrew-stax
- Add homebrew-stax to repository access list
