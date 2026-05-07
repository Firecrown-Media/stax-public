# Changelog

## [1.0.0] - 2026-05-06

### Changed

- Provider-agnostic config schema v2: `.stax.yml` now uses `provider: wpengine` + `provider_config:` instead of hardcoded `wpengine:` block
- `cmd/` layer is now thin wrappers — all business logic lives in `pkg/database/`, `pkg/files/`, `pkg/init/`, `pkg/actions/`
- Commands resolve the authenticated provider from the registry instead of calling `pkg/wpengine` directly
- Removed stub providers: `wordpress-vip`, `aws`, `local` (VIP has its own CLI: `vip-cli`)
- `SiteConfig.WPEngineDomain` renamed to `SiteConfig.ProviderDomain` (yaml: `provider_domain`)

### Migration

If you have an existing `.stax.yml`, update it to the v2 format:

```yaml
# Before (v1):
wpengine:
  install: my-install
  environment: production

# After (v2):
provider: wpengine
provider_config:
  install: my-install
  environment: production
```

Run `stax config migrate` to migrate automatically.
