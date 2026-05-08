Read and validate the `.stax.yml` configuration.

## stax config get

Print a configuration value by dot-path key.

```bash
stax config get <key>
```

```bash
stax config get provider
stax config get project.name
stax config get provider_config.install
```

## stax config set

Set a configuration value by dot-path key.

```bash
stax config set <key> <value>
```

```bash
stax config set project.name my-project
stax config set provider_config.environment staging
```

## stax config list

Print all configuration values.

```bash
stax config list
```

## stax config show

Show the full current configuration in YAML format.

```bash
stax config show
```

## stax config validate

Check that `.stax.yml` is well-formed and all required fields are present.

```bash
stax config validate
```

Returns a non-zero exit code if validation fails, making it suitable for CI checks.

## stax config template

Generate a configuration template with all fields and their defaults.

```bash
stax config template
```

Useful for initializing a new project manually or reviewing all available options.

## stax config migrate

Migrate a v1 configuration file to the current v2 schema.

```bash
stax config migrate
```

Rewrites `.stax.yml` in place. The v1 `wpengine:` block is replaced with `provider: wpengine` and `provider_config:`.

---

## .stax.yml reference

Stax requires `version: 2`. V1 files are rejected — run `stax config migrate` to upgrade.

```yaml
version: 2

# Project metadata
project:
  name: my-project             # used for the DDEV site name and local URL
  type: wordpress              # wordpress | wordpress-multisite
  mode: single                 # single | subdomain | subdirectory

# Provider selection
provider: wpengine

# Provider-specific settings
provider_config:
  install: my-install          # WPEngine install name
  environment: production      # production | staging
  ssh_gateway: ""              # optional: override ssh.wpengine.net

# Migration destination (required for stax migrate commands)
migration:
  destination: vip             # vip

# DDEV configuration
ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  webserver_type: nginx-fpm
  nodejs_version: "20"
  composer_version: "2"

# Multisite network (wordpress-multisite only)
network:
  domain: my-project.ddev.site
  sites:
    - name: site1
      slug: site1
      title: Site One
      domain: site1.my-project.ddev.site
      provider_domain: site1.mynetwork.com
      active: true

# Remote media proxy
media:
  proxy_enabled: true
  primary_source: https://my-install.wpengine.com
  wpengine_fallback: true
  cache:
    enabled: true
    directory: .stax/media-cache
    max_size: 1GB
    ttl: 86400

# Database snapshot settings
snapshots:
  directory: ~/.stax/snapshots
  auto_snapshot_before_pull: true
  compression: gzip
  retention:
    auto: 7      # days
    manual: 30   # days

# Performance tuning
performance:
  rsync_bandwidth_limit: 0    # KB/s; 0 = unlimited
```
