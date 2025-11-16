# File Synchronization Troubleshooting Guide

Comprehensive guide to diagnosing and fixing file sync issues with Stax, including the v2.12.5 media proxy file sync fix.

## Quick Diagnosis

Use this flowchart to quickly identify your issue:

```
Files not syncing?
│
├─ No files synced at all
│  └─ See: [Complete Sync Failure](#complete-sync-failure)
│
├─ Missing uploads directory
│  └─ See: [Missing Uploads Directory](#missing-uploads-directory)
│
├─ Missing themes/plugins
│  └─ See: [Missing Themes or Plugins](#missing-themes-or-plugins)
│
├─ Wrong files being synced
│  └─ See: [Wrong Files Being Synced](#wrong-files-being-synced)
│
└─ Sync takes too long
   └─ See: [Slow Sync Performance](#slow-sync-performance)
```

## The v2.12.5 Fix: What Changed

### What Was Broken

In Stax v2.12.4 and earlier, there was a critical bug in the `stax init` command:

**The Problem:**
- The uploads directory was ALWAYS excluded from sync
- No check of the `media.proxy_enabled` configuration setting
- Users couldn't sync media files even when media proxy was disabled

**Affected Code:**
```go
// OLD CODE (v2.12.4 and earlier) - BROKEN
filesExcludeUploads = true  // Always excluded!
```

### What's Fixed

In Stax v2.12.5+, the file sync logic now properly respects your configuration:

**The Fix:**
```go
// NEW CODE (v2.12.5+) - FIXED
if cfg.Media.ProxyEnabled {
    filesExcludeUploads = true   // Exclude only if proxy enabled
} else {
    filesExcludeUploads = false  // Include if proxy disabled
}
```

**Behavior Changes:**

| Configuration | v2.12.4 (Broken) | v2.12.5+ (Fixed) |
|---------------|------------------|------------------|
| `proxy_enabled: true` | Excludes uploads ✓ | Excludes uploads ✓ |
| `proxy_enabled: false` | Excludes uploads ✗ (WRONG!) | Includes uploads ✓ |
| No config (default) | Excludes uploads ✓ | Excludes uploads ✓ |

### How to Tell If You're Affected

You were affected by this bug if:

1. You're using Stax v2.12.4 or earlier
2. You have `media.proxy_enabled: false` in your `.stax.yml`
3. Your uploads directory is missing or empty

**Check your version:**
```bash
stax version
```

**Check your config:**
```bash
cat .stax.yml | grep proxy_enabled
```

**Check your uploads:**
```bash
ls -la public/wp-content/uploads/
```

### Migration Steps

If you were affected, follow these steps:

**Option 1: Re-run init (Recommended)**
```bash
# Update to v2.12.5+
brew upgrade stax

# Re-initialize to sync uploads
stax init --force

# Verify uploads are synced
ls -la public/wp-content/uploads/
```

**Option 2: Manual file pull**
```bash
# Pull files with uploads included
stax files pull --environment=production

# Verify uploads are present
ls -la public/wp-content/uploads/
```

## Common Issues

### Missing Uploads Directory

**Symptoms:**
- `public/wp-content/uploads/` doesn't exist or is empty
- Images broken in WordPress admin/frontend
- Media library shows no files

**Diagnosis:**

1. **Check your Stax version:**
   ```bash
   stax version
   ```
   If < v2.12.5, you may have the bug.

2. **Check media proxy configuration:**
   ```bash
   grep -A 5 "media:" .stax.yml
   ```

3. **Check if uploads directory exists:**
   ```bash
   ls -la public/wp-content/uploads/
   ```

**Solutions:**

#### Solution 1: Media Proxy Enabled (Intended Behavior)

If media proxy is enabled, uploads SHOULD be excluded:

```yaml
# .stax.yml
media:
  proxy_enabled: true  # This is correct for media proxy
```

**This is working as intended.** Media is served from remote:

```bash
# Verify media proxy is configured
stax media status

# Set up media proxy if not configured
stax media setup-proxy

# Test that media loads from remote
stax media test
```

See [Media Proxy Setup Guide](media-proxy-setup.md) for full setup.

#### Solution 2: Media Proxy Disabled (Should Include Uploads)

If media proxy is disabled, uploads SHOULD be synced:

```yaml
# .stax.yml
media:
  proxy_enabled: false  # Uploads should be synced
```

**If uploads are missing with v2.12.5+:**

```bash
# Re-pull files
stax files pull --environment=production

# Verify uploads synced
ls -la public/wp-content/uploads/
```

**If uploads still missing (v2.12.4 or earlier):**

```bash
# Upgrade Stax first
brew upgrade stax

# Then re-initialize
stax init --force
```

#### Solution 3: Missing Config (Use Default)

If no media config exists, default is proxy enabled:

```bash
# Add explicit config to .stax.yml
cat >> .stax.yml << 'EOF'

media:
  proxy_enabled: false  # Set to false to sync uploads
EOF

# Re-run init
stax init --force
```

### Missing Themes or Plugins

**Symptoms:**
- Themes directory empty or missing
- Plugins directory empty or missing
- WordPress shows "broken theme" or missing plugins

**Diagnosis:**

```bash
# Check what was synced
ls -la public/wp-content/

# Look for:
ls -la public/wp-content/themes/
ls -la public/wp-content/plugins/
ls -la public/wp-content/mu-plugins/
```

**Solutions:**

#### Check WPEngine Configuration

```bash
# Verify WPEngine install name
grep -A 3 "wpengine:" .stax.yml
```

Should show:
```yaml
wpengine:
  install: mysite        # Must match your WPEngine install
  environment: production
```

**Test SSH access:**
```bash
# Try manual SSH connection
ssh mysite@mysite.ssh.wpengine.net

# If this fails, check:
# 1. SSH key is added to WPEngine
# 2. Install name is correct
# 3. You have access to the install
```

#### Re-pull Files

```bash
# Pull files explicitly
stax files pull --environment=production

# Check what was synced
stax files status
```

#### Check Remote Source

```bash
# SSH into WPEngine and verify files exist
ssh mysite@mysite.ssh.wpengine.net

# On WPEngine server:
ls -la sites/mysite/wp-content/themes/
ls -la sites/mysite/wp-content/plugins/

# Exit
exit
```

If themes/plugins are missing on WPEngine, that's the source issue.

### Wrong Files Being Synced

**Symptoms:**
- Uploads synced when media proxy enabled
- Uploads not synced when media proxy disabled
- Unexpected files included/excluded

**Diagnosis:**

**Step 1: Check Stax version**
```bash
stax version
# Should be v2.12.5 or higher
```

**Step 2: Check configuration**
```bash
cat .stax.yml | grep -A 10 "media:"
```

**Step 3: Check what's actually synced**
```bash
find public/wp-content -type d -maxdepth 1
```

**Step 4: Review init output**
```bash
# Re-run init and watch output
stax init --force 2>&1 | tee init-output.log

# Look for:
# - "Media proxy enabled - excluding uploads directory"
# - OR "Media proxy disabled - pulling all files including uploads"
```

**Solutions:**

#### Fix 1: Update Stax Version

```bash
# If version < v2.12.5
brew upgrade stax

# Verify new version
stax version

# Re-initialize
stax init --force
```

#### Fix 2: Correct Configuration

Ensure your `.stax.yml` matches your intent:

**For media proxy (exclude uploads):**
```yaml
media:
  proxy_enabled: true  # Uploads will be excluded ✓
```

**For local media (include uploads):**
```yaml
media:
  proxy_enabled: false  # Uploads will be synced ✓
```

#### Fix 3: Manual Sync Control

Use `stax files pull` with explicit flags:

```bash
# Exclude uploads manually
stax files pull --exclude-uploads

# Include uploads manually
stax files pull --environment=production
# (no --exclude-uploads flag)
```

### Complete Sync Failure

**Symptoms:**
- No files synced at all
- Error messages during `stax init`
- "File pull failed" errors

**Diagnosis:**

```bash
# Check DDEV status
ddev status

# Check SSH connectivity
ssh mysite@mysite.ssh.wpengine.net echo "Connected"

# Check .stax.yml syntax
cat .stax.yml | grep -A 20 "wpengine:"
```

**Common Causes:**

#### Cause 1: SSH Key Not Configured

**Symptoms:**
- "Permission denied (publickey)" errors
- SSH connection fails

**Solution:**
```bash
# Generate SSH key if needed
ssh-keygen -t ed25519 -C "your-email@example.com"

# Add key to WPEngine
cat ~/.ssh/id_ed25519.pub
# Copy and add to WPEngine portal: User → SSH Keys

# Test connection
ssh mysite@mysite.ssh.wpengine.net
```

#### Cause 2: Wrong WPEngine Install Name

**Symptoms:**
- "Host not found" errors
- "Invalid install name" errors

**Solution:**
```bash
# Check install name in WPEngine portal
# Update .stax.yml:
wpengine:
  install: correct-install-name  # Must match exactly
```

#### Cause 3: Missing Dependencies

**Symptoms:**
- "rsync command not found"
- "ssh command not found"

**Solution:**
```bash
# macOS: Install Xcode Command Line Tools
xcode-select --install

# Or install via Homebrew
brew install rsync openssh
```

#### Cause 4: Firewall/Network Issues

**Symptoms:**
- Timeouts during sync
- "Connection refused" errors

**Solution:**
```bash
# Test WPEngine connectivity
curl -I https://mysite.wpengine.com

# Test SSH connectivity
ssh -v mysite@mysite.ssh.wpengine.net
# Look for connection errors in verbose output

# Try alternative network (VPN, different WiFi, etc.)
```

### Slow Sync Performance

**Symptoms:**
- Sync takes hours instead of minutes
- Progress appears stalled

**Diagnosis:**

```bash
# Check what's being synced
stax files status

# Check uploads directory size (if being synced)
du -sh public/wp-content/uploads/

# Monitor sync progress
stax files pull --verbose
```

**Solutions:**

#### Solution 1: Enable Media Proxy (Recommended)

Skip uploads sync entirely:

```yaml
# .stax.yml
media:
  proxy_enabled: true
```

```bash
# Re-initialize with proxy enabled
stax init --force

# Set up media proxy
stax media setup-proxy
```

See [Media Proxy Setup Guide](media-proxy-setup.md).

#### Solution 2: Selective Sync

Sync only recent uploads:

```bash
# Sync only current year
stax files pull --include="uploads/2024/**" --exclude="uploads/**"

# Or sync only specific month
stax files pull --include="uploads/2024/11/**" --exclude="uploads/**"
```

#### Solution 3: Increase Network Performance

```bash
# Use compression
stax files pull --compress

# Parallel transfers (if available)
stax files pull --parallel
```

## Verification Commands

### Verify Sync Completed Successfully

```bash
# Check directory structure
tree -L 3 public/wp-content/

# Verify critical directories exist
ls -la public/wp-content/themes/
ls -la public/wp-content/plugins/
ls -la public/wp-content/mu-plugins/

# Check uploads (if media proxy disabled)
ls -la public/wp-content/uploads/
```

### Verify Media Proxy Configuration

```bash
# Check status
stax media status

# Test proxy
stax media test

# Check nginx config
cat .ddev/nginx_full/media-proxy.conf
```

### Verify File Counts

```bash
# Count themes
ls -1 public/wp-content/themes/ | wc -l

# Count plugins
ls -1 public/wp-content/plugins/ | wc -l

# Count uploads (if synced)
find public/wp-content/uploads -type f | wc -l
```

## Manual Sync Commands

### Sync All Files

```bash
# Full sync including uploads
stax files pull --environment=production

# Full sync excluding uploads
stax files pull --environment=production --exclude-uploads
```

### Sync Specific Directories

```bash
# Themes only
stax files pull --include="themes/**" --exclude="**"

# Plugins only
stax files pull --include="plugins/**" --exclude="**"

# Uploads only
stax files pull --include="uploads/**" --exclude="**"

# Current year uploads only
stax files pull --include="uploads/2024/**" --exclude="uploads/**"
```

### Dry Run (Preview)

```bash
# See what would be synced without actually syncing
stax files pull --dry-run
```

## Advanced Troubleshooting

### Enable Debug Logging

```bash
# Run with verbose output
stax init --verbose

# Or set debug environment variable
export STAX_DEBUG=true
stax init
```

### Check rsync Command

```bash
# View exact rsync command being used
stax files pull --verbose 2>&1 | grep rsync
```

### Inspect .stax.yml

```bash
# Validate YAML syntax
cat .stax.yml | python -m json.tool 2>&1 || echo "Invalid YAML"

# View full config
cat .stax.yml

# View media config specifically
grep -A 15 "media:" .stax.yml
```

### Check DDEV Logs

```bash
# View web server logs
ddev logs -s web

# View error logs
ddev logs -s web | grep -i error

# View nginx logs
ddev ssh
cat /var/log/nginx/error.log
exit
```

## Common Error Messages

### "Media proxy disabled - pulling all files including uploads"

**Meaning:** Media proxy is turned off, so uploads will be synced.

**Action:**
- If you want media proxy: Set `proxy_enabled: true` in `.stax.yml`
- If you want local uploads: This is correct, let it sync

### "Excluding uploads directory (configure media proxy for remote media)"

**Meaning:** Uploads are being excluded from sync.

**In v2.12.4 and earlier:** This always appeared (bug).

**In v2.12.5+:** This appears only when `proxy_enabled: true`.

**Action:**
- If you want uploads: Set `proxy_enabled: false`
- If you want media proxy: Run `stax media setup-proxy`

### "File validation warnings detected"

**Meaning:** Some expected directories are missing or empty.

**Action:**
```bash
# Check which directories are missing
stax files status

# Re-pull files
stax files pull --environment=production

# Verify on remote server
ssh mysite@mysite.ssh.wpengine.net
ls -la sites/mysite/wp-content/
exit
```

### "File pull failed"

**Meaning:** rsync command failed during sync.

**Actions:**
1. Check SSH connectivity
2. Verify WPEngine install name
3. Check network connection
4. Review error message for specific cause

See [Complete Sync Failure](#complete-sync-failure).

## Getting Help

### Diagnostic Command

```bash
# Run Stax diagnostics
stax doctor

# This checks:
# - Stax version
# - Dependencies installed
# - DDEV status
# - Configuration validity
# - WPEngine connectivity
```

### Collect Information

Before asking for help, collect:

```bash
# Version info
stax version

# Configuration
cat .stax.yml

# Directory structure
tree -L 3 public/wp-content/

# Recent errors
stax init --verbose 2>&1 | tail -50
```

### Related Documentation

- [Media Proxy Setup Guide](media-proxy-setup.md) - Setting up media proxy
- [Media Proxy Reference](../MEDIA_PROXY.md) - Complete media proxy docs
- [General Troubleshooting](../TROUBLESHOOTING.md) - Other common issues
- [WPEngine Integration](../WPENGINE.md) - WPEngine-specific features
- [FAQ](../FAQ.md) - Frequently asked questions

## Summary Checklist

Use this checklist to verify file sync is working correctly:

### For Media Proxy Enabled (Default)

- [ ] Stax version v2.12.5 or higher
- [ ] `.stax.yml` has `media.proxy_enabled: true`
- [ ] `stax init` shows "excluding uploads directory"
- [ ] `public/wp-content/themes/` exists and has content
- [ ] `public/wp-content/plugins/` exists and has content
- [ ] `public/wp-content/uploads/` is empty or doesn't exist
- [ ] `stax media status` shows proxy enabled
- [ ] `stax media test` passes
- [ ] Images load in browser from remote source

### For Media Proxy Disabled

- [ ] Stax version v2.12.5 or higher
- [ ] `.stax.yml` has `media.proxy_enabled: false`
- [ ] `stax init` shows "pulling all files including uploads"
- [ ] `public/wp-content/themes/` exists and has content
- [ ] `public/wp-content/plugins/` exists and has content
- [ ] `public/wp-content/uploads/` exists and has content
- [ ] Images load in browser from local files
- [ ] File count matches remote (approximately)

If all items are checked, your file sync is working correctly!
