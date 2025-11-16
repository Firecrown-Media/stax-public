# Media Proxy Setup Guide

Step-by-step guide to configuring Stax media proxy for efficient WordPress development without downloading gigabytes of media files.

## What You'll Achieve

By the end of this guide, you'll have:
- Media proxy configured to serve files from WPEngine/CDN
- Local WordPress site with zero media downloads
- Faster project initialization
- Significant disk space savings

## Prerequisites

Before you begin:

- Stax installed ([Installation Guide](../INSTALLATION.md))
- WPEngine hosting account with site access
- (Optional) BunnyCDN or other CDN URL
- Existing `.stax.yml` configuration file

## Understanding Media Proxy

Media proxy allows your local environment to serve media files from a remote source (WPEngine or CDN) without downloading them. When enabled:

- **Uploads directory is excluded** from sync
- **Media served on-demand** from remote server
- **Disk space saved:** 10GB-200GB typical
- **Faster init:** Minutes instead of hours

When disabled:
- **All files synced** including uploads
- **Media stored locally**
- **Better for:** Offline work, media editing

## Step 1: Configure .stax.yml

Open your `.stax.yml` file and add or update the media proxy configuration:

### Basic Configuration (WPEngine Only)

```yaml
wpengine:
  install: mysite              # Your WPEngine install name
  environment: production

media:
  proxy_enabled: true          # Enable media proxy
```

With this minimal config, Stax will automatically use `https://mysite.wpengine.com` as the media source.

### Advanced Configuration (with CDN)

```yaml
wpengine:
  install: mysite
  environment: production

media:
  proxy_enabled: true

  # Optional: BunnyCDN configuration
  bunnycdn:
    hostname: mysite.b-cdn.net
    storage_zone: mysite-storage
    api_key: ${BUNNYCDN_API_KEY}  # Use environment variable

  # Optional: Custom proxy settings
  proxy:
    cache:
      enabled: true
      ttl: 30d                # Cache files for 30 days
      max_size: 10g          # Maximum 10GB cache
```

### Configuration with Custom URLs

```yaml
media:
  proxy_enabled: true
  proxy:
    remote_url: https://cdn.mysite.com        # Primary CDN
    fallback_url: https://mysite.wpengine.com # Fallback if CDN fails
    cache:
      enabled: true
      ttl: 30d
      max_size: 10g
```

## Step 2: Initialize Your Project

Now initialize your project with media proxy:

```bash
# If this is a new project
stax init

# If you already have a project, reinitialize to apply changes
stax init --force
```

**What happens:**
1. Stax reads your `.stax.yml` configuration
2. Configures DDEV with nginx media proxy
3. Syncs themes, plugins, mu-plugins (but NOT uploads)
4. Sets up proxy to serve media from remote

**Expected output:**
```
Pulling Files
Media proxy enabled - excluding uploads directory
Syncing: themes, plugins, mu-plugins
This may take several minutes...
Files pulled successfully
File validation passed
  - themes: present
  - plugins: present
  - uploads: excluded (media proxy enabled)
```

## Step 3: Set Up nginx Proxy Configuration

The `stax init` command should have configured nginx automatically, but you can also set it up manually or update it:

```bash
# Auto-detect settings from .stax.yml
stax media setup-proxy

# Or specify URLs explicitly
stax media setup-proxy --cdn=https://mysite.b-cdn.net

# Or use WPEngine only
stax media setup-proxy --url=https://mysite.wpengine.com
```

**What this does:**
1. Generates `.ddev/nginx_full/media-proxy.conf`
2. Configures proxy to remote media source
3. Sets up caching (if enabled)
4. Validates nginx configuration
5. Restarts DDEV to apply changes

## Step 4: Verify Media Proxy is Working

### Check Configuration Status

```bash
stax media status
```

**Expected output:**
```
Media Proxy Status

Configuration
  Proxy Enabled:   ✓ Yes
  CDN Hostname:    mysite.b-cdn.net
  WPEngine:        mysite.wpengine.com
  Cache Enabled:   ✓ Enabled
  Cache Max Size:  10g

Nginx Configuration
  Config File:     ✓ Exists
  Location:        .ddev/nginx_full/media-proxy.conf
  Validation:      ✓ Valid

DDEV Status
  Status:          ✓ Running
  Primary URL:     https://my-site.ddev.site
```

### Test Media Proxy

```bash
stax media test
```

This validates:
- nginx configuration exists and is valid
- DDEV is running
- Proxy sources are configured correctly

### Manual Browser Verification

1. **Start your site:**
   ```bash
   stax start
   ```

2. **Open in browser:**
   ```bash
   open https://your-site.ddev.site
   ```

3. **Check DevTools:**
   - Open browser DevTools (F12)
   - Go to Network tab
   - Navigate to a page with images
   - Click on an image request
   - Look for these headers:
     - `X-Proxy-Source: cdn` or `X-Proxy-Source: wpengine`
     - `X-Cache-Status: HIT` or `X-Cache-Status: MISS`

4. **Verify file path:**
   Images should load from paths like:
   ```
   https://your-site.ddev.site/wp-content/uploads/2024/11/image.jpg
   ```
   But be served from:
   ```
   https://mysite.b-cdn.net/wp-content/uploads/2024/11/image.jpg
   ```

### Check Local Files

Verify uploads directory is NOT present locally:

```bash
ls -la public/wp-content/
```

You should see:
- `themes/` ✓
- `plugins/` ✓
- `mu-plugins/` ✓
- `uploads/` ✗ (should NOT exist, or be empty)

## Step 5: Optional - Configure Caching

Caching significantly improves performance by storing frequently-accessed files locally.

### Enable Caching

Already enabled by default, but you can customize:

```bash
# Cache for 30 days (default)
stax media setup-proxy --cache-ttl=30d

# Cache for 7 days (less disk usage)
stax media setup-proxy --cache-ttl=7d

# Disable caching (always fetch fresh)
stax media setup-proxy --no-cache
```

### Monitor Cache Performance

```bash
# Check cache status
stax media status
# Look for: Cache Size: X.X GB

# View cache hit rate
# Open site and check DevTools Network tab
# Look for X-Cache-Status header:
#   - MISS: First request (fetched from remote)
#   - HIT: Cached (served from local cache)
```

### Clear Cache

If needed, clear the cache:

```bash
ddev ssh
sudo rm -rf /var/cache/nginx/media/*
exit
stax restart
```

## Common Issues and Quick Fixes

### Issue: Images Not Loading

**Symptoms:** Broken image icons in browser

**Solutions:**

1. **Check nginx config exists:**
   ```bash
   ls -la .ddev/nginx_full/media-proxy.conf
   ```
   If missing: `stax media setup-proxy`

2. **Verify DDEV is running:**
   ```bash
   stax status
   ```
   If stopped: `stax start`

3. **Test remote URL:**
   ```bash
   curl -I https://mysite.b-cdn.net/wp-content/uploads/test.jpg
   ```
   Should return 200 OK

4. **Check nginx logs:**
   ```bash
   ddev logs -s web | grep -i error
   ```

### Issue: Slow Image Loading

**Symptoms:** Images take 10-30 seconds to load

**Solutions:**

1. **Enable caching:**
   ```bash
   stax media setup-proxy --cache
   stax restart
   ```

2. **Preload cache by browsing site:**
   Open your site and navigate through pages. Second loads will be much faster.

3. **Use faster CDN:**
   ```bash
   stax media setup-proxy --cdn=https://mysite.b-cdn.net
   ```

### Issue: Wrong Files Being Proxied

**Symptoms:** Some files load locally, others from remote

**Solutions:**

Check your nginx configuration:
```bash
cat .ddev/nginx_full/media-proxy.conf
```

Should include:
```nginx
location ~ ^/wp-content/uploads/(.*)$ {
    try_files $uri @proxy_media;
}
```

This tries local file first, then proxies if not found.

### Issue: "Media proxy disabled" During Init

**Symptoms:** `stax init` says "Media proxy disabled - pulling all files including uploads"

**Cause:** `media.proxy_enabled: false` in `.stax.yml`

**Solution:**

Update `.stax.yml`:
```yaml
media:
  proxy_enabled: true  # Change from false to true
```

Then re-run:
```bash
stax init --force
```

## Next Steps

### Optimize Performance

1. **Hybrid approach** - Download frequently-used files, proxy the rest:
   ```bash
   # Download current year only
   stax files pull --include="uploads/2024/**"
   ```
   nginx will serve 2024 from disk (fast) and proxy older content.

2. **Increase cache size** if you have disk space:
   ```yaml
   media:
     proxy:
       cache:
         max_size: 20g  # Instead of default 10g
   ```

3. **Use regional CDN** for better latency:
   ```yaml
   media:
     proxy:
       remote_url: https://cdn-us-west.mysite.com
   ```

### Learn More

- [Media Proxy Reference](../MEDIA_PROXY.md) - Complete documentation
- [File Sync Troubleshooting](troubleshooting-file-sync.md) - Detailed sync help
- [Performance Optimization](../MEDIA_PROXY.md#performance) - Advanced performance tips
- [WPEngine Integration](../WPENGINE.md) - WPEngine-specific features

### Advanced Configuration

- [Custom nginx Configuration](../MEDIA_PROXY.md#custom-nginx-configuration)
- [Multisite Media Proxy](../MEDIA_PROXY.md#per-site-proxies-multisite)
- [Conditional Proxying](../MEDIA_PROXY.md#conditional-proxying)
- [Debugging](../MEDIA_PROXY.md#debugging)

## Summary

You've successfully configured media proxy! Your local environment now:

- Serves media from remote WPEngine/CDN
- Saves 10GB-200GB of disk space
- Initializes in minutes instead of hours
- Caches frequently-used files for better performance

**Quick reference:**
- Check status: `stax media status`
- Test config: `stax media test`
- Update proxy: `stax media setup-proxy`
- View cache: Check DevTools Network tab for `X-Cache-Status` header

## Troubleshooting

If you encounter issues not covered here, see:
- [File Sync Troubleshooting Guide](troubleshooting-file-sync.md)
- [Media Proxy Troubleshooting](../MEDIA_PROXY.md#troubleshooting)
- [General Troubleshooting](../TROUBLESHOOTING.md)
