Proxy media requests from your local DDEV environment to the WPEngine or BunnyCDN origin so you don't need to download the uploads directory.

## What the proxy does

When the proxy is active, any request to `/wp-content/uploads/` in your local environment is forwarded to the configured remote URL. Images and other media load from production automatically — no local copy needed. This eliminates the need to run `stax files pull` on the uploads directory, which is often several gigabytes.

## Setup

```bash
stax media setup
```

This writes an nginx configuration file to `.ddev/nginx_full/media-proxy.conf` and restarts DDEV to apply it. The file proxies all requests under `/wp-content/uploads/` to the origin specified in `.stax.yml`.

## Verify

```bash
stax media status
```

Shows whether the proxy configuration file exists and whether the proxy is active.

## Required config in .stax.yml

```yaml
media:
  proxy_enabled: true
  primary_source: https://my-install.wpengine.com
```

Replace `my-install.wpengine.com` with your WPEngine install URL. If you're using BunnyCDN, set `primary_source` to your CDN hostname instead.

## Test it

After running `stax start`, open any media URL from your local site in a browser — for example `https://my-project.ddev.site/wp-content/uploads/2024/01/image.jpg`. The image should load without any local copy of the file.

If it doesn't load, check:

1. `stax media status` — confirm the proxy is active
2. The browser network tab — a `502` means the origin URL is unreachable; a `404` means the file doesn't exist on the origin

## What not to do

- Do not set `primary_source` to a staging URL for a production site. Staging environments often have different SSL certificates and CORS headers, which can cause mixed-content errors in the browser.
