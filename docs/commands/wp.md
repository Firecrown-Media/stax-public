Run WP-CLI commands inside the DDEV container.

Stax does not expose a dedicated `stax wp` wrapper. WP-CLI runs directly through DDEV:

```bash
ddev wp <wp-cli-args>
```

DDEV makes `wp` available inside the web container. All WP-CLI commands work the same way — just prefix with `ddev`.

## Examples

```bash
# Flush the object cache
ddev wp cache flush

# List active plugins
ddev wp plugin list --status=active

# Export the database
ddev wp db export backup.sql

# Create a user
ddev wp user create editor editor@example.com --role=editor

# Run search-replace
ddev wp search-replace 'https://old-domain.com' 'https://new-domain.ddev.site' --all-tables

# Check the site URL
ddev wp option get siteurl
```

## Path notes

Commands run inside the container. File paths in WP-CLI arguments must use container-internal paths, not host paths. The web root is `/var/www/html`.

```bash
# Export to the web root (accessible from host at the project directory)
ddev wp db export /var/www/html/backup.sql
```

## Multisite

For multisite networks, pass `--url=<site-domain>` to target a specific subsite:

```bash
ddev wp --url=subsite.my-project.ddev.site option get siteurl
```

For network-wide operations:

```bash
ddev wp site list --network
ddev wp plugin list --network
```
