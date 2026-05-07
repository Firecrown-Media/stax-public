# Changelog

## [2.21.0](https://github.com/Firecrown-Media/stax/compare/v2.20.1...v2.21.0) (2026-05-07)


### Features

* add decodeProviderConfig and AuthenticateFromConfig to WPEngine provider ([50c1a6c](https://github.com/Firecrown-Media/stax/commit/50c1a6c7566e61a6ebbf158e5904b4d75ee4a370))
* add Provider and ProviderConfig fields to Config struct (v2) ([2a4752e](https://github.com/Firecrown-Media/stax/commit/2a4752edfabe75db62690b8c435e781f32cbc03c))
* add v2 schema version validation and simplify mergeConfigs ([d48e4b8](https://github.com/Firecrown-Media/stax/commit/d48e4b813ea467154aaa904488b155a7b33bca70))
* create pkg/database service with Pull, Push, URL helpers ([03aac13](https://github.com/Firecrown-Media/stax/commit/03aac1343be4fdb5ed8191e0e08d0544b8ec4dd4))
* create pkg/files service with Pull, Push, BuildSyncOptions ([3a08cf0](https://github.com/Firecrown-Media/stax/commit/3a08cf0b9cc33f58fb51d0d2f3886ec593ca1582))


### Bug Fixes

* move FixEnvironmentMismatch doc comment to correct position ([cf19d8a](https://github.com/Firecrown-Media/stax/commit/cf19d8aee60d49245226e764fac4d71f4582483a))

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
