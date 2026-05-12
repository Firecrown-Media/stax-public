# Changelog

## [2.23.0](https://github.com/Firecrown-Media/stax/compare/v2.22.0...v2.23.0) (2026-05-12)


### Features

* add stax migrate checklist command ([#145](https://github.com/Firecrown-Media/stax/issues/145)) ([faa5c1e](https://github.com/Firecrown-Media/stax/commit/faa5c1e56917ea2b8ac96fda96bf57bc586ed566))


### Bug Fixes

* expand ~ in SSH key env var paths before stat ([d3b1404](https://github.com/Firecrown-Media/stax/commit/d3b14044453f461333624d496fc872ded40d50f4))
* merge migration.destination from project config ([51aa52f](https://github.com/Firecrown-Media/stax/commit/51aa52fbff0b08c5a8c202701e6988bebfdd0d61))

## [2.22.0](https://github.com/Firecrown-Media/stax/compare/v2.21.3...v2.22.0) (2026-05-12)


### Features

* **migrate:** add stax migrate publish command ([42da82d](https://github.com/Firecrown-Media/stax/commit/42da82d0a994238f81ea81cefd9b49a1b99446fb))
* **migration:** add Publish() to upload report to S3 and commit to VIP repo ([b8bd370](https://github.com/Firecrown-Media/stax/commit/b8bd370480c24a734a03662ff9318b68ceda2acb))
* **migration:** add report data types and plugin/theme helpers ([b9b5d99](https://github.com/Firecrown-Media/stax/commit/b9b5d9962704abe1f21cce19b8841691afbf3fb7))
* **migration:** add SQL/media analysis, enriched report template, update Report() to VIP-style output ([4ce7161](https://github.com/Firecrown-Media/stax/commit/4ce7161687a2ab9e78b0d250ba02ec728ed344d1))
* **migration:** add stax migrate command group for WPEngine → VIP pipeline ([#143](https://github.com/Firecrown-Media/stax/issues/143)) ([f54cea9](https://github.com/Firecrown-Media/stax/commit/f54cea94f04e00c7b3fdfafcca583134c3c145a3))


### Bug Fixes

* correct WPEngine SSH remote paths and known_hosts fallback ([c2e14d1](https://github.com/Firecrown-Media/stax/commit/c2e14d14f0ff259a6878afb3a810b2a4f493439c))

## [2.21.3](https://github.com/Firecrown-Media/stax/compare/v2.21.2...v2.21.3) (2026-05-07)


### Bug Fixes

* correct double-paren syntax errors and spinner printf vet warnings ([fb2d021](https://github.com/Firecrown-Media/stax/commit/fb2d02162c4205cf4b3a2b8454f3ca80d49cce24))
* drop blank identifier from map index (S1005 lint) ([2e04f92](https://github.com/Firecrown-Media/stax/commit/2e04f926b3deb86aa06cc5c0c4f06670a266ef6f))
* remaining printf vet issues — Verbose/Warning/Success with variable args ([17d0ac1](https://github.com/Firecrown-Media/stax/commit/17d0ac12b0db0d94c6bc7d6c383ad0204cbf198e))
* resolve lint and integration test failures from refactor ([1acb60c](https://github.com/Firecrown-Media/stax/commit/1acb60c4652eefe7be22ce412486728bbd4acfe1))
* restore missing closing paren in snapshot.go ([49d34e5](https://github.com/Firecrown-Media/stax/commit/49d34e5aa6681db37cb748dda6d76d64b105c2e5))
* **security:** bump golang.org/x/crypto to v0.45.0, update to Go 1.24 ([3382f69](https://github.com/Firecrown-Media/stax/commit/3382f69182f38563c6c2b650b4575263f85f32c7))
* unwrap fmt.Sprintf from ui calls — satisfies Go 1.24 printf vet check ([ee2dde1](https://github.com/Firecrown-Media/stax/commit/ee2dde191993a641cda2c1ed3568543866fdef5e))
* upgrade golangci-lint to latest via official action for Go 1.24 support ([a8400e5](https://github.com/Firecrown-Media/stax/commit/a8400e5f9a8abcd12ce9ed09c6986281fdc57432))
* wrap variable args with %s in remaining ui calls for Go 1.24 vet ([54c9e66](https://github.com/Firecrown-Media/stax/commit/54c9e66dd84b5514312ce887e518b1c960956515))

## [2.21.2](https://github.com/Firecrown-Media/stax/compare/v2.21.1...v2.21.2) (2026-05-07)


### Bug Fixes

* use git reset --hard to restore staged deletions before GoReleaser ([29c8d88](https://github.com/Firecrown-Media/stax/commit/29c8d886b606961831d65e29749d34414a67d0e4))

## [2.21.1](https://github.com/Firecrown-Media/stax/compare/v2.21.0...v2.21.1) (2026-05-07)


### Bug Fixes

* restore working directory before GoReleaser to avoid dirty state ([895d2a9](https://github.com/Firecrown-Media/stax/commit/895d2a9971cd259a3f37fd5373c60e59588238d7))

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
