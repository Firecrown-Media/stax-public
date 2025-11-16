# Media Proxy Guide

Complete guide to using Stax's media proxy feature for serving remote media files without local downloads.

---

## Table of Contents

- [Overview](#overview)
- [How It Works](#how-it-works)
- [Setup](#setup)
- [Configuration](#configuration)
- [Commands](#commands)
- [Troubleshooting](#troubleshooting)
- [Performance](#performance)
- [Advanced Usage](#advanced-usage)

---

## Recent Fixes

### v2.12.5 - File Synchronization Fix (November 2024)

**Issue:** Previous versions of Stax incorrectly handled upload directory exclusion during `stax init`, causing files to be synced improperly regardless of media proxy configuration.

**The Problem:**
- The `stax init` command was hardcoded to always exclude the uploads directory
- No check of the `media.proxy_enabled` configuration setting before excluding uploads
- Users with media proxy disabled were unable to sync their media files
- Users with media proxy enabled would still have uploads excluded (correct), but the logic was wrong

**The Fix:**
The `stax init` command now properly respects the `media.proxy_enabled` setting in your `.stax.yml` configuration:

**When media proxy IS enabled** (default behavior):
- Excludes uploads directory from sync
- Syncs only: themes, plugins, mu-plugins
- Media files served from remote WPEngine/CDN on-demand
- Faster initialization, smaller disk usage
- See [Setup](#setup) for configuration

**When media proxy IS NOT enabled:**
- Includes uploads directory in sync
- Syncs: themes, plugins, mu-plugins, AND uploads
- All media files stored locally
- Longer initialization time due to media download
- Better for offline work or when you need to modify media

**Impact:**
- Users can now successfully sync media files when media proxy is disabled
- File synchronization behavior now matches configuration expectations
- Added validation to verify critical directories exist after sync
- Improved user feedback about what's being synced

**Migration Guide:**
If you initialized a project with Stax v2.12.4 or earlier and media proxy was disabled:

```bash
# Re-run init to sync missing uploads
stax init

# Or manually pull files with uploads
stax files pull --environment=production
```

If media proxy is enabled (default), no action needed - behavior is unchanged.

**See Also:**
- [Troubleshooting File Sync](guides/troubleshooting-file-sync.md) - Detailed sync troubleshooting
- [Media Proxy Setup Guide](guides/media-proxy-setup.md) - Step-by-step configuration
- [Troubleshooting](#troubleshooting) - General media proxy issues

---

## Overview

### What is Media Proxy?

Media proxy is a feature in Stax that allows your local WordPress environment to serve media files (images, videos, PDFs, etc.) from a remote source (WPEngine or CDN) without downloading them to your local machine.

Instead of storing gigabytes of media locally, nginx in your DDEV environment fetches files on-demand from production when your browser requests them.

### Why Use Media Proxy?

**The Problem:**
Modern WordPress sites often have massive media libraries:
- **E-commerce sites:** 10GB-50GB of product images
- **News/Magazine sites:** 50GB-200GB of articles and photos
- **Portfolio sites:** 20GB-100GB of high-resolution images
- **Video sites:** 100GB+ of video content

Downloading all this media for local development:
- Takes hours or days to sync
- Wastes 10GB-200GB of disk space per project
- Requires constant re-syncing as content is added
- Slows down initial project setup significantly

**The Solution:**
With media proxy enabled:
- **Zero download time** - No media files to download
- **Zero disk space used** - Files streamed from remote
- **Always up-to-date** - Shows current production media
- **Fast setup** - Project ready in minutes, not hours

### When to Use Media Proxy

**You should use media proxy when:**
- Uploads directory is large (10GB+)
- You don't need to modify media files
- You have a stable internet connection
- You want fast project initialization
- You're working on code, not content
- Multiple developers need consistent media access

**You should download files when:**
- Testing WordPress upload functionality
- Working offline frequently
- Modifying images/media locally
- Uploads directory is small (<1GB)
- Internet connection is slow/unreliable
- Need maximum performance (no network latency)

### Key Benefits

1. **Saves Time:** No waiting for multi-gigabyte downloads
2. **Saves Space:** 10GB-200GB saved per project
3. **Always Current:** See latest production media automatically
4. **Transparent:** Works seamlessly with WordPress
5. **Flexible:** Can cache frequently-used files locally
6. **Selective:** Download only specific files you need

---

## How It Works

### Technical Architecture

Stax uses nginx reverse proxy in DDEV to intercept requests for media files and fetch them from remote sources.

**Components:**
- **DDEV nginx:** Web server that handles all HTTP requests
- **nginx proxy module:** Proxies requests to remote servers
- **nginx cache:** Optionally caches proxied files locally
- **Remote source:** WPEngine or BunnyCDN hosting the media

### Request Flow (Detailed)

When your browser requests a media file, here's the complete flow:

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Browser Request                                          │
│    GET https://my-site.ddev.site/wp-content/uploads/       │
│        2024/11/logo.jpg                                     │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. nginx Receives Request                                   │
│    - Checks location ~ ^/wp-content/uploads/(.*)$          │
│    - Matches the upload path pattern                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Local Filesystem Check (try_files $uri)                 │
│    Checks: /var/www/html/wp-content/uploads/2024/11/       │
│           logo.jpg                                          │
└─────────────────────────────────────────────────────────────┘
              ↓                                ↓
         File exists?                    File not found
              ↓                                ↓
   ┌──────────────────┐         ┌──────────────────────────┐
   │ Serve from disk  │         │ Jump to @proxy_media     │
   │ (instant)        │         │ (proxy location)         │
   └──────────────────┘         └──────────────────────────┘
              ↓                                ↓
         [Browser]           ┌─────────────────────────────┐
                            │ 4. Check nginx Cache         │
                            │    - Hash the request URI    │
                            │    - Check cache directory   │
                            └─────────────────────────────┘
                                         ↓                  ↓
                                   Cache HIT          Cache MISS
                                         ↓                  ↓
                            ┌──────────────────┐  ┌─────────────────┐
                            │ Serve from cache │  │ 5. Proxy to CDN │
                            │ (< 10ms)         │  │    Primary      │
                            └──────────────────┘  └─────────────────┘
                                         ↓                  ↓
                                    [Browser]    ┌────────────────────┐
                                                │ GET https://        │
                                                │ mysite.b-cdn.net/   │
                                                │ wp-content/uploads/ │
                                                │ 2024/11/logo.jpg    │
                                                └────────────────────┘
                                                         ↓          ↓
                                                   CDN Success   CDN 404
                                                         ↓          ↓
                                         ┌──────────────────┐   ┌──────────────┐
                                         │ 6. Return image  │   │ 7. Fallback  │
                                         │    Cache it      │   │    WPEngine  │
                                         │    Send browser  │   └──────────────┘
                                         └──────────────────┘          ↓
                                                    ↓          ┌──────────────────┐
                                               [Browser]      │ GET https://      │
                                                             │ mysite.wpengine. │
                                                             │ com/wp-content/  │
                                                             │ uploads/2024/11/ │
                                                             │ logo.jpg         │
                                                             └──────────────────┘
                                                                      ↓        ↓
                                                               WPE Success  WPE 404
                                                                      ↓        ↓
                                                         ┌─────────────────┐  ┌────────┐
                                                         │ Return image    │  │ 404 to │
                                                         │ Cache it        │  │ Browser│
                                                         │ Send to browser │  └────────┘
                                                         └─────────────────┘
                                                                  ↓
                                                             [Browser]
```

### Performance Timings

**First Request (Cold - No Cache):**
1. Local file check: <1ms
2. Cache check: <1ms
3. CDN request: 100-300ms (network latency + CDN processing)
4. Total: ~100-300ms

**Second Request (Warm - Cached):**
1. Local file check: <1ms
2. Cache hit: <10ms
3. Total: ~10ms (10-30x faster!)

**Local File (Hybrid):**
1. Local file check: <1ms
2. Serve from disk: <1ms
3. Total: ~1ms (100x faster!)

### nginx Configuration Explained

Stax generates nginx configuration at `.ddev/nginx_full/media-proxy.conf`:

**1. Upload Location Block:**
```nginx
location ~ ^/wp-content/uploads/(.*)$ {
    # Try local file first, fallback to proxy
    try_files $uri @proxy_media;
}
```
This captures all requests to the uploads directory and attempts local file first.

**2. Proxy Location Block:**
```nginx
location @proxy_media {
    # Proxy to primary source (CDN)
    proxy_pass https://mysite.b-cdn.net$request_uri;

    # Handle errors (404 = try fallback)
    proxy_intercept_errors on;
    error_page 404 = @wpengine_fallback;

    # SSL configuration
    proxy_ssl_server_name on;
    proxy_ssl_verify off;  # Development only

    # Caching (if enabled)
    proxy_cache media_cache;
    proxy_cache_valid 200 30d;     # Cache successful responses 30 days
    proxy_cache_valid 404 1m;      # Cache 404s for 1 minute
    proxy_cache_key "$scheme$request_method$host$request_uri";

    # Headers
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    add_header X-Proxy-Source "cdn";
    add_header X-Cache-Status $upstream_cache_status;

    # Performance
    proxy_buffering on;
    proxy_buffer_size 4k;
    proxy_buffers 8 4k;
}
```

**3. Fallback Location Block:**
```nginx
location @wpengine_fallback {
    # Proxy to WPEngine if CDN fails
    proxy_pass https://mysite.wpengine.com$request_uri;
    proxy_set_header Host mysite.wpengine.com;

    # Same caching settings
    proxy_cache media_cache;
    proxy_cache_valid 200 30d;

    add_header X-Proxy-Source "wpengine";
}
```

**4. Cache Path Configuration:**
```nginx
# Defined at http/server level
proxy_cache_path /var/cache/nginx/media
    levels=1:2                # Directory structure: /a/bc/...
    keys_zone=media_cache:10m # 10MB shared memory (~80k keys)
    max_size=10g             # Max total cache size
    inactive=30d;            # Delete after 30 days unused
```

### Cache Mechanism

**Cache Key Generation:**

nginx generates a unique key for each request:
```nginx
proxy_cache_key "$scheme$request_method$host$request_uri";
```

Example:
- Request: `https://my-site.ddev.site/wp-content/uploads/2024/11/logo.jpg`
- Key: `httpsGETmy-site.ddev.site/wp-content/uploads/2024/11/logo.jpg`
- Hash: MD5 sum of key (e.g., `a1b2c3d4e5f6...`)

**Cache Storage:**

The hash is used to create a directory structure:
```
/var/cache/nginx/media/
├── 1/
│   └── 2a/
│       └── a1b2c3d4e5f6... (cached file + metadata)
├── 3/
│   └── f1/
│       └── b2c3d4e5f6a1...
└── ...
```

The `levels=1:2` setting creates this two-level hierarchy to prevent too many files in a single directory.

**Cache Metadata:**

Each cached file includes:
- The actual media content (image bytes, video, etc.)
- Response headers from origin server
- Cache status and expiration info
- Last access timestamp

---

## Setup

### Prerequisites

Before setting up media proxy:

1. **DDEV must be configured** - Run `stax init` or have existing `.ddev/config.yaml`
2. **WPEngine or CDN URL** - You need at least one remote media source
3. **Optional: `.stax.yml`** - For automatic configuration detection

### Quick Setup

**Basic setup (auto-detect from .stax.yml):**
```bash
stax media setup-proxy
```

Stax will:
- Read WPEngine install from `.stax.yml`
- Read BunnyCDN hostname if configured
- Generate nginx configuration
- Validate configuration
- Restart DDEV

**Setup with CDN URL:**
```bash
stax media setup-proxy --cdn=https://mysite.b-cdn.net
```

**Setup with custom WPEngine URL:**
```bash
stax media setup-proxy --url=https://mysite.wpengine.com
```

**Setup without caching:**
```bash
stax media setup-proxy --no-cache
```
Always fetches from remote (slower but uses zero disk space).

**Setup with custom cache TTL:**
```bash
stax media setup-proxy --cache-ttl=7d
```
Cache files for 7 days instead of default 30 days.

### Configuration Detection

If `.stax.yml` exists, Stax automatically detects:

**WPEngine configuration:**
```yaml
wpengine:
  install: mysite
```
Converts to: `https://mysite.wpengine.com`

**BunnyCDN configuration:**
```yaml
media:
  bunnycdn:
    hostname: mysite.b-cdn.net
```
Converts to: `https://mysite.b-cdn.net`

**Media proxy settings:**
```yaml
media:
  proxy:
    enabled: true
    cache:
      enabled: true
      ttl: 30d
      max_size: 10g
```

### Manual Configuration

You can manually configure media proxy in `.stax.yml`:

```yaml
media:
  proxy:
    enabled: true
    remote_url: https://cdn.mysite.com
    fallback_url: https://mysite.wpengine.com
    cache:
      enabled: true
      ttl: 30d
      max_size: 10g
      directory: .ddev/media-cache  # Optional custom cache location
```

After editing `.stax.yml`, run:
```bash
stax media setup-proxy
stax restart
```

### Verifying Setup

**Check status:**
```bash
stax media status
```

Output:
```
Media Proxy Status

Configuration
  Proxy Enabled:   ✓ Yes
  CDN Hostname:    mysite.b-cdn.net
  WPEngine:        mysite.wpengine.com
  WP Fallback:     ✓ Enabled
  Cache Enabled:   ✓ Enabled
  Cache Directory: .ddev/media-cache
  Cache Max Size:  10g

Nginx Configuration
  Config File:     ✓ Exists
  Location:        .ddev/nginx_full/media-proxy.conf
  Validation:      ✓ Valid

DDEV Status
  Status:          ✓ Running
  Primary URL:     https://my-site.ddev.site

Cache Status
  Cache Directory: ✓ Exists
  Location:        /Users/me/project/.ddev/media-cache
  Cache Size:      2.3 GB
```

**Test configuration:**
```bash
stax media test
```

Output:
```
Testing Media Proxy

Configuration Tests
✓ Nginx configuration file exists

Environment Tests
✓ DDEV is running

Nginx Validation
✓ Nginx configuration is valid

Proxy Source Tests
  CDN URL: https://mysite.b-cdn.net
✓ BunnyCDN configured
  WPEngine URL: https://mysite.wpengine.com
✓ WPEngine configured

All media proxy tests passed!

Manual verification steps:
  1. Visit: https://my-site.ddev.site
  2. Navigate to a page with media/images
  3. Check browser DevTools Network tab
  4. Verify images load from remote source
  5. Look for X-Proxy-Source header in response
```

---

## Configuration

### Configuration File Options

Full `.stax.yml` media proxy configuration:

```yaml
media:
  # Enable/disable media proxy
  proxy:
    enabled: true

    # Primary remote source
    remote_url: https://cdn.mysite.com

    # Fallback source (if primary fails)
    fallback_url: https://mysite.wpengine.com

    # Caching configuration
    cache:
      enabled: true
      ttl: 30d              # How long to cache files
      max_size: 10g         # Maximum cache size
      directory: .ddev/media-cache  # Cache location (optional)

  # BunnyCDN configuration (optional - for automatic URL detection)
  bunnycdn:
    hostname: mysite.b-cdn.net
    storage_zone: mysite-storage
    api_key: ${BUNNYCDN_API_KEY}  # From environment variable

# WPEngine configuration (for automatic URL detection)
wpengine:
  install: mysite
  environment: production
```

### Command-Line Options

**`stax media setup-proxy` flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `--cdn` | CDN URL for primary source | `--cdn=https://mysite.b-cdn.net` |
| `--url` | WPEngine URL (overrides config) | `--url=https://mysite.wpengine.com` |
| `--cache` | Enable caching (default: true) | `--cache=true` or `--no-cache` |
| `--cache-ttl` | Cache duration | `--cache-ttl=7d` or `--cache-ttl=24h` |

**Cache TTL formats:**
- `30d` - 30 days
- `7d` - 7 days (1 week)
- `24h` - 24 hours (1 day)
- `12h` - 12 hours
- `60m` - 60 minutes (1 hour)

### Multisite Configuration

For WordPress multisite, you can configure per-site media proxies:

```yaml
network:
  sites:
    - name: site1
      domain: site1.mynetwork.local
      wpengine_domain: site1.com
      media:
        proxy:
          enabled: true
          remote_url: https://cdn.site1.com
          fallback_url: https://site1.wpengine.com

    - name: site2
      domain: site2.mynetwork.local
      wpengine_domain: site2.com
      media:
        proxy:
          enabled: true
          remote_url: https://cdn.site2.com
          fallback_url: https://site2.wpengine.com
```

Each subsite can have:
- Different CDN URLs
- Different WPEngine sources
- Different caching settings

---

## Commands

### `stax media setup-proxy`

Configure nginx for media proxying.

**Usage:**
```bash
stax media setup-proxy [flags]
```

**Flags:**
- `--cdn=URL` - CDN URL for primary source
- `--url=URL` - WPEngine URL for fallback
- `--cache` / `--no-cache` - Enable/disable caching (default: enabled)
- `--cache-ttl=DURATION` - Cache TTL (default: 30d)

**Examples:**

```bash
# Auto-detect from .stax.yml
stax media setup-proxy

# Specify CDN
stax media setup-proxy --cdn=https://mysite.b-cdn.net

# Use WPEngine only (no CDN)
stax media setup-proxy --url=https://mysite.wpengine.com

# Disable caching
stax media setup-proxy --no-cache

# Custom cache TTL
stax media setup-proxy --cache-ttl=7d

# Combine options
stax media setup-proxy \
  --cdn=https://mysite.b-cdn.net \
  --url=https://mysite.wpengine.com \
  --cache-ttl=14d
```

**What it does:**
1. Validates DDEV is configured
2. Reads configuration from `.stax.yml` (if exists)
3. Applies command-line overrides
4. Generates `.ddev/nginx_full/media-proxy.conf`
5. Generates `.ddev/nginx_full/cache-config.conf` (if caching)
6. Validates nginx configuration syntax
7. Restarts DDEV to apply changes
8. Shows configuration summary

---

### `stax media status`

Show media proxy status and configuration.

**Usage:**
```bash
stax media status
```

**Shows:**
- Configuration from `.stax.yml`
- nginx configuration file status
- DDEV running status
- Cache directory and size
- Validation status

**Example output:**
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

Cache Status
  Cache Directory: ✓ Exists
  Location:        /path/to/project/.ddev/media-cache
  Cache Size:      2.3 GB
```

---

### `stax media test`

Test media proxy configuration and connectivity.

**Usage:**
```bash
stax media test
```

**Tests:**
1. Nginx configuration file exists
2. DDEV is running
3. Nginx configuration syntax is valid
4. Proxy sources are configured
5. (Optional) Connectivity to remote sources

**Example output:**
```
Testing Media Proxy

Configuration Tests
✓ Nginx configuration file exists

Environment Tests
✓ DDEV is running

Nginx Validation
✓ Nginx configuration is valid

Proxy Source Tests
  CDN URL: https://mysite.b-cdn.net
✓ BunnyCDN configured
  WPEngine URL: https://mysite.wpengine.com
✓ WPEngine configured

All media proxy tests passed!

Manual verification steps:
  1. Visit: https://my-site.ddev.site
  2. Navigate to a page with media/images
  3. Check browser DevTools Network tab
  4. Verify images load from remote source
  5. Look for X-Proxy-Source header in response
```

---

## Troubleshooting

### Media Not Loading

**Symptom:** Images show as broken in browser.

**Diagnostic Steps:**

1. **Check nginx configuration exists:**
   ```bash
   ls -la .ddev/nginx_full/media-proxy.conf
   ```
   If missing, run `stax media setup-proxy`.

2. **Check DDEV is running:**
   ```bash
   stax status
   ```
   If stopped, run `stax start`.

3. **Validate nginx configuration:**
   ```bash
   stax media test
   ```
   If invalid, regenerate: `stax media setup-proxy`.

4. **Check browser DevTools:**
   - Open DevTools → Network tab
   - Navigate to page with images
   - Click on failed image request
   - Check Status Code and Headers

5. **Test remote URL manually:**
   ```bash
   curl -I https://mysite.b-cdn.net/wp-content/uploads/2024/11/test.jpg
   ```
   Should return 200 OK.

**Common Causes:**

**1. nginx config not generated:**
```bash
# Fix: Generate config
stax media setup-proxy
stax restart
```

**2. Wrong remote URL:**
```bash
# Check current config
stax media status

# Fix: Set correct URL
stax media setup-proxy --cdn=https://correct-url.com
```

**3. Remote source blocked/down:**
```bash
# Test connectivity
curl -I https://mysite.b-cdn.net

# Fix: Use fallback or different source
stax media setup-proxy --url=https://mysite.wpengine.com
```

**4. SSL certificate issues:**
```nginx
# In .ddev/nginx_full/media-proxy.conf
proxy_ssl_verify off;  # Temporarily disable SSL verification
```

---

### Slow Performance

**Symptom:** Images take a long time to load.

**Diagnostic Steps:**

1. **Check if caching is enabled:**
   ```bash
   stax media status
   # Look for: Cache Enabled: ✓ Enabled
   ```

2. **Check cache hit rate:**
   - Open DevTools → Network tab
   - Look for `X-Cache-Status` header
   - First load: `MISS`
   - Subsequent loads: `HIT`

3. **Check cache size:**
   ```bash
   stax media status
   # Look for: Cache Size: X.X GB
   ```

4. **Test response times:**
   ```bash
   # First request (cold)
   time curl -o /dev/null https://my-site.ddev.site/wp-content/uploads/test.jpg

   # Second request (cached)
   time curl -o /dev/null https://my-site.ddev.site/wp-content/uploads/test.jpg
   ```

**Solutions:**

**1. Enable caching:**
```bash
stax media setup-proxy --cache
stax restart
```

**2. Increase cache TTL:**
```bash
stax media setup-proxy --cache-ttl=30d
stax restart
```

**3. Use faster CDN:**
```bash
# Switch to BunnyCDN or Cloudflare
stax media setup-proxy --cdn=https://mysite.b-cdn.net
```

**4. Download frequently-used files:**
```bash
# Download specific directory
rsync -avz wpengine:/path/to/uploads/2024/ ./wp-content/uploads/2024/
```
nginx will serve local files (faster) and proxy the rest.

---

### Cache Not Working

**Symptom:** `X-Cache-Status` always shows `MISS`.

**Diagnostic Steps:**

1. **Check cache directory exists:**
   ```bash
   ddev ssh
   ls -la /var/cache/nginx/media/
   exit
   ```

2. **Check cache configuration:**
   ```bash
   cat .ddev/nginx_full/cache-config.conf
   ```

3. **Check nginx error logs:**
   ```bash
   ddev logs -s web -f
   # Look for cache-related errors
   ```

**Common Causes:**

**1. Cache directory missing:**
```bash
# SSH into container
ddev ssh

# Create cache directory
sudo mkdir -p /var/cache/nginx/media
sudo chown -R www-data:www-data /var/cache/nginx/media
sudo chmod -R 755 /var/cache/nginx/media

exit

# Restart
stax restart
```

**2. Cache disabled in config:**
```bash
# Re-enable cache
stax media setup-proxy --cache
stax restart
```

**3. Cache size limit reached:**
```bash
# Increase cache size
stax media setup-proxy --cache-ttl=30d
# Or clear cache
ddev ssh
sudo rm -rf /var/cache/nginx/media/*
exit
```

---

### Headers Missing

**Symptom:** Can't see `X-Proxy-Source` or `X-Cache-Status` headers.

**Solutions:**

1. **Check nginx config includes add_header directives:**
   ```bash
   cat .ddev/nginx_full/media-proxy.conf | grep add_header
   ```
   Should show:
   ```nginx
   add_header X-Proxy-Source "cdn";
   add_header X-Cache-Status $upstream_cache_status;
   ```

2. **Regenerate config:**
   ```bash
   stax media setup-proxy
   stax restart
   ```

3. **Check in different browser:**
   Sometimes browser extensions hide headers. Try incognito mode.

---

### WPEngine Fallback Not Working

**Symptom:** Images fail when CDN is down, but WPEngine fallback doesn't work.

**Diagnostic Steps:**

1. **Check fallback configuration:**
   ```bash
   cat .ddev/nginx_full/media-proxy.conf | grep wpengine_fallback
   ```

2. **Test WPEngine URL:**
   ```bash
   curl -I https://mysite.wpengine.com/wp-content/uploads/test.jpg
   ```

3. **Check nginx error logs:**
   ```bash
   ddev logs -s web | grep -i error
   ```

**Solutions:**

**1. Ensure fallback is configured:**
```bash
stax media setup-proxy \
  --cdn=https://mysite.b-cdn.net \
  --url=https://mysite.wpengine.com
```

**2. Check WPEngine domain is correct:**
```yaml
# In .stax.yml
wpengine:
  install: mysite  # Should match your WPEngine install name
```

**3. Set Host header correctly:**
```nginx
# In @wpengine_fallback location
proxy_set_header Host mysite.wpengine.com;
```

---

## Performance

### Cache Strategies

**1. Long TTL for Stable Content:**
```bash
# Cache for 30 days (good for historical content)
stax media setup-proxy --cache-ttl=30d
```

**Use when:**
- Content rarely changes
- Uploads are historical (old blog posts, archives)
- Disk space is available

**2. Short TTL for Dynamic Content:**
```bash
# Cache for 1 day (good for frequently-updated content)
stax media setup-proxy --cache-ttl=1d
```

**Use when:**
- Content changes frequently
- Testing with near-live data
- Limited disk space

**3. No Cache (Always Fresh):**
```bash
# No caching (always fetch from remote)
stax media setup-proxy --no-cache
```

**Use when:**
- Minimal disk space available
- Need absolutely current media
- Testing cache-related issues

### Bandwidth Considerations

**First Page Load:**
- Without cache: Downloads all visible media (~5-20 MB typical page)
- With cache: Only uncached media
- With local files: Zero bandwidth

**Subsequent Loads:**
- With cache: Zero bandwidth (served from disk)
- Without cache: Full download each time

**Typical Bandwidth Usage:**

| Scenario | Initial Load | Cached Load | Daily Usage |
|----------|-------------|-------------|-------------|
| Blog post (10 images) | 5 MB | 0 MB | 5 MB |
| Product page (50 images) | 20 MB | 0 MB | 20 MB |
| Full site browse | 50-100 MB | 0 MB | 50-100 MB |

**With 95% cache hit rate:**
- Daily development: ~50-100 MB/day
- Weekly development: ~200-500 MB/week

### Optimizing Performance

**1. Hybrid Approach (Best Performance):**

Download frequently-used assets, proxy the rest:

```bash
# Setup proxy
stax media setup-proxy

# Download current year uploads
rsync -avz wpengine:/path/to/uploads/2024/ ./wp-content/uploads/2024/

# nginx serves 2024 from disk (fast), proxies older content
```

**2. Increase Cache Size:**
```bash
# Allow larger cache (20GB instead of 10GB)
stax media setup-proxy --cache-ttl=30d

# Manually edit .ddev/nginx_full/cache-config.conf:
# max_size=20g
```

**3. Use Nearest CDN:**
```bash
# Use geographically closer CDN
stax media setup-proxy --cdn=https://cdn-us-west.mysite.com
```

**4. Preload Cache:**

After setup, browse your site to preload cache:

```bash
# Start site
stax start

# Open in browser and navigate through pages
# Or use wget to crawl and cache
wget --mirror --no-parent https://my-site.ddev.site
```

### Monitoring Cache Performance

**Check cache hit rate:**

```bash
# SSH into container
ddev ssh

# Check nginx cache stats
grep -i "cache" /var/log/nginx/access.log | \
  grep -o "X-Cache-Status: [A-Z]*" | \
  sort | uniq -c

# Output:
#   856 X-Cache-Status: HIT
#    44 X-Cache-Status: MISS
# Hit rate: 856/(856+44) = 95%

exit
```

**Monitor cache size:**
```bash
# Check current cache size
stax media status

# Or manually
ddev ssh
du -sh /var/cache/nginx/media/
exit
```

**Clear cache statistics:**
```bash
ddev ssh
sudo rm -rf /var/cache/nginx/media/*
exit
stax restart
```

---

## Advanced Usage

### Custom nginx Configuration

If you need custom proxy behavior, you can manually edit the nginx configuration:

**1. Edit configuration:**
```bash
nano .ddev/nginx_full/media-proxy.conf
```

**2. Add custom directives:**
```nginx
location @proxy_media {
    # Custom timeout
    proxy_connect_timeout 10s;
    proxy_send_timeout 30s;
    proxy_read_timeout 30s;

    # Custom buffer sizes
    proxy_buffer_size 8k;
    proxy_buffers 16 8k;

    # Custom headers
    add_header X-Custom-Header "value";

    # Existing proxy_pass, etc...
    proxy_pass https://mysite.b-cdn.net$request_uri;
    # ...
}
```

**3. Validate and restart:**
```bash
stax media test
stax restart
```

**Note:** Custom changes will be overwritten if you run `stax media setup-proxy` again. Consider creating a separate config file that won't be overwritten:

```bash
# Create custom config
touch .ddev/nginx_full/media-proxy-custom.conf

# nginx automatically loads all .conf files
# Add your customizations there
```

### Per-Site Proxies (Multisite)

For WordPress multisite with different media sources per site:

**Configuration:**
```yaml
network:
  sites:
    - name: site1
      domain: site1.mynetwork.local
      media:
        proxy:
          enabled: true
          remote_url: https://cdn.site1.com
          fallback_url: https://site1.wpengine.com
          cache:
            enabled: true
            ttl: 30d

    - name: site2
      domain: site2.mynetwork.local
      media:
        proxy:
          enabled: true
          remote_url: https://cdn.site2.com
          fallback_url: https://site2.wpengine.com
          cache:
            enabled: false  # No cache for site2
```

**Generated nginx configuration:**
```nginx
# Site 1
location ~ ^/wp-content/uploads/sites/2/(.*)$ {
    try_files $uri @proxy_media_site1;
}

location @proxy_media_site1 {
    proxy_pass https://cdn.site1.com$request_uri;
    error_page 404 = @wpengine_fallback_site1;
    proxy_cache media_cache_site1;
    # ...
}

# Site 2
location ~ ^/wp-content/uploads/sites/3/(.*)$ {
    try_files $uri @proxy_media_site2;
}

location @proxy_media_site2 {
    proxy_pass https://cdn.site2.com$request_uri;
    # No caching for site2
    # ...
}
```

### Conditional Proxying

Proxy only specific file types:

**Edit `.ddev/nginx_full/media-proxy.conf`:**
```nginx
# Only proxy images (not videos/PDFs)
location ~ ^/wp-content/uploads/.*\.(jpg|jpeg|png|gif|webp)$ {
    try_files $uri @proxy_media;
}

# Serve videos/PDFs only if local
location ~ ^/wp-content/uploads/.*\.(mp4|pdf|zip)$ {
    try_files $uri =404;
}
```

### Debugging

**Enable nginx debug logging:**

**1. Create debug config:**
```bash
# .ddev/nginx_full/debug.conf
error_log /var/log/nginx/error.log debug;
```

**2. Restart and check logs:**
```bash
stax restart
ddev logs -s web -f | grep -i proxy
```

**3. Verbose curl testing:**
```bash
# Test with verbose output
curl -v https://my-site.ddev.site/wp-content/uploads/test.jpg

# Look for:
# - HTTP status codes
# - X-Proxy-Source header
# - X-Cache-Status header
# - Response times
```

**4. Check cache keys:**
```bash
ddev ssh

# View cache directory structure
find /var/cache/nginx/media/ -type f | head -20

# Check a specific file's metadata
cat /var/cache/nginx/media/1/2f/a1b2c3d4e5f6...

exit
```

---

## Next Steps

- **WPEngine Integration:** [WPENGINE.md](./WPENGINE.md) - Full WPEngine setup
- **Getting Started:** [GETTING_STARTED.md](./GETTING_STARTED.md) - Initial setup
- **User Guide:** [USER_GUIDE.md](./USER_GUIDE.md) - Complete features
- **Troubleshooting:** [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) - Common issues

---

**Questions or issues?** Check the [Troubleshooting](#troubleshooting) section above or contact your team lead.
