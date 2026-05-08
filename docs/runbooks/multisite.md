Working with WordPress multisite networks using stax.

## Supported modes

| Mode | Description |
|------|-------------|
| `subdomain` | Each site on its own subdomain: `site1.mynetwork.com`, `site2.mynetwork.com` |
| `subdirectory` | Each site in a path: `mynetwork.com/site1`, `mynetwork.com/site2` |

Set the mode in `.stax.yml`:

```yaml
project:
  type: wordpress-multisite
  mode: subdomain
```

## Initialize a multisite project

Run `stax init` and choose `wordpress-multisite` when prompted for project type. The resulting `.stax.yml` will include the network section:

```yaml
version: 2
provider: wpengine
provider_config:
  install: my-network
  environment: production
project:
  name: my-project
  type: wordpress-multisite
  mode: subdomain
network:
  domain: my-project.ddev.site
  sites:
    - name: site1
      slug: site1
      title: Site One
      domain: site1.my-project.ddev.site
      provider_domain: site1.mynetwork.com
      active: true
```

Map each remote subdomain or path to its local equivalent in the `network.sites[]` array. The `provider_domain` is the production domain; `domain` is the local DDEV domain.

## Pull the database for a multisite

```bash
stax db pull
```

Works the same as single-site. URL replacement runs network-wide, then per-site for subdomain networks to handle the individual site domains mapped in `network.sites[]`.

## Network domain config

The `network.sites[]` array controls per-site URL replacement. For each active site, `stax db pull` runs:

```
wp search-replace 'https://<provider_domain>' 'https://<domain>' --all-tables
```

Keep this array up to date as sites are added or removed from the network.

## Adding a new site to the network

Use `stax wp` to run WP-CLI network commands in the DDEV container:

```bash
stax wp site create --slug=newsite --title="New Site" --email=admin@example.com
```

Then add the new site to the `network.sites[]` array in `.stax.yml` so URL replacement includes it on the next `stax db pull`.

## What not to do

- Don't manually edit `wp_blogs` or `wp_site` to add sites. Use WP-CLI via `stax wp` to avoid database inconsistencies.
- Subdomain mode requires wildcard DNS in DDEV. Add wildcard hostnames to your DDEV config or use `additional_hostnames` in `.ddev/config.yaml`. Without this, subdomain sites will return DNS errors locally.
