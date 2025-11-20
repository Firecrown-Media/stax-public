# Changelog

## [2.14.6](https://github.com/Firecrown-Media/stax/compare/v2.14.5...v2.14.6) (2025-11-20)


### Bug Fixes

* **diagnostics:** improve SSH key detection across all fallback locations ([#116](https://github.com/Firecrown-Media/stax/issues/116)) ([e01881b](https://github.com/Firecrown-Media/stax/commit/e01881bad4c8b482c7001da39be10d0c7444bed3))

## [2.14.5](https://github.com/Firecrown-Media/stax/compare/v2.14.4...v2.14.5) (2025-11-20)


### Performance Improvements

* dramatically improve WPEngine install picker performance ([#114](https://github.com/Firecrown-Media/stax/issues/114)) ([5c7cc2c](https://github.com/Firecrown-Media/stax/commit/5c7cc2c56613483820a6cf431048e59e18442fef))

## [2.14.4](https://github.com/Firecrown-Media/stax/compare/v2.14.3...v2.14.4) (2025-11-20)


### Bug Fixes

* implement database pull functionality and install-specific SSH gateway ([#112](https://github.com/Firecrown-Media/stax/issues/112)) ([ad703a2](https://github.com/Firecrown-Media/stax/commit/ad703a209b8f703c7ef83baa0ecbe72b7aaf4604))

## [2.14.3](https://github.com/Firecrown-Media/stax/compare/v2.14.2...v2.14.3) (2025-11-18)


### Bug Fixes

* eliminate all hardcoded 'public' DocRoot references in cmd/init.go ([#110](https://github.com/Firecrown-Media/stax/issues/110)) ([4fd1a1f](https://github.com/Firecrown-Media/stax/commit/4fd1a1f2d25b799997925dd75154a7744239d475))

## [2.14.2](https://github.com/Firecrown-Media/stax/compare/v2.14.1...v2.14.2) (2025-11-18)


### Bug Fixes

* **ci:** update Homebrew tap workflow to use stax-public repository ([#108](https://github.com/Firecrown-Media/stax/issues/108)) ([81254ff](https://github.com/Firecrown-Media/stax/commit/81254ff04d7f613c5c62d76a40fa5f040fa6561a))

## [2.14.1](https://github.com/Firecrown-Media/stax/compare/v2.14.0...v2.14.1) (2025-11-18)


### Bug Fixes

* **ddev:** change default DocRoot from 'public' to '.' for WordPress ([#106](https://github.com/Firecrown-Media/stax/issues/106)) ([db2ed34](https://github.com/Firecrown-Media/stax/commit/db2ed3433bdbb801878c9fc425dce4d617241eb2))

## [2.14.0](https://github.com/Firecrown-Media/stax/compare/v2.13.3...v2.14.0) (2025-11-18)


### Features

* UX improvements phase 2 - documentation, .gitignore, and Docker alternatives ([#104](https://github.com/Firecrown-Media/stax/issues/104)) ([a931342](https://github.com/Firecrown-Media/stax/commit/a931342200e0248e8b2e9e32dbc1e6aa4b96ef8d))

## [2.13.3](https://github.com/Firecrown-Media/stax/compare/v2.13.2...v2.13.3) (2025-11-18)


### Bug Fixes

* **status:** parse DDEV v1.24.4 JSON structure correctly ([1e84de5](https://github.com/Firecrown-Media/stax/commit/1e84de585db50dc387f8074b74780802352bb2cb))

## [2.13.2](https://github.com/Firecrown-Media/stax/compare/v2.13.1...v2.13.2) (2025-11-17)


### Bug Fixes

* **setup:** accept numeric input for storage method selection ([#101](https://github.com/Firecrown-Media/stax/issues/101)) ([c5c5aa3](https://github.com/Firecrown-Media/stax/commit/c5c5aa39039fb8a1f820001ef823feb45f7cdacc))

## [2.13.1](https://github.com/Firecrown-Media/stax/compare/v2.13.0...v2.13.1) (2025-11-17)


### Bug Fixes

* **credentials:** use fallback credential loading in init and list commands ([#99](https://github.com/Firecrown-Media/stax/issues/99)) ([f5b38b2](https://github.com/Firecrown-Media/stax/commit/f5b38b2b23540bc4cdfdc092cb0dde7d29ec9f3f))

## [2.13.0](https://github.com/Firecrown-Media/stax/compare/v2.12.8...v2.13.0) (2025-11-17)


### Features

* **ux:** implement credential storage fallback and WPEngine install picker ([#97](https://github.com/Firecrown-Media/stax/issues/97)) ([527567e](https://github.com/Firecrown-Media/stax/commit/527567e78031976759e5ca309ad2737ea6c31b81))

## [2.12.8](https://github.com/Firecrown-Media/stax/compare/v2.12.7...v2.12.8) (2025-11-17)


### Bug Fixes

* resolve DDEV status check race conditions in init and db commands ([#95](https://github.com/Firecrown-Media/stax/issues/95)) ([0af3e43](https://github.com/Firecrown-Media/stax/commit/0af3e43c37f9a451e24ed5cf2aaeff0de9daddc0))

## [2.12.7](https://github.com/Firecrown-Media/stax/compare/v2.12.6...v2.12.7) (2025-11-17)


### Bug Fixes

* **ci:** upgrade golangci-lint to v1.62.2 and Go to 1.23 for compatibility ([#92](https://github.com/Firecrown-Media/stax/issues/92)) ([20bcbda](https://github.com/Firecrown-Media/stax/commit/20bcbda0de4d09f8402c93aa943cacbd312fe05a))

## [2.12.6](https://github.com/Firecrown-Media/stax/compare/v2.12.5...v2.12.6) (2025-11-17)


### Bug Fixes

* **db:** resolve DDEV detection failure with absolute path handling ([#90](https://github.com/Firecrown-Media/stax/issues/90)) ([6fb3e8f](https://github.com/Firecrown-Media/stax/commit/6fb3e8fcec463a30005e90482d54c6966a3ce3f6))

## [2.12.5](https://github.com/Firecrown-Media/stax/compare/v2.12.4...v2.12.5) (2025-11-16)


### Bug Fixes

* **init:** respect media proxy config when pulling files ([#88](https://github.com/Firecrown-Media/stax/issues/88)) ([37d21bb](https://github.com/Firecrown-Media/stax/commit/37d21bbc1f26eadd9f99e23ed6bdadb3d18b1925))

## [2.12.4](https://github.com/Firecrown-Media/stax/compare/v2.12.3...v2.12.4) (2025-11-16)


### Bug Fixes

* update version command to reference public mirror repository ([80b89c4](https://github.com/Firecrown-Media/stax/commit/80b89c47607905431eb4cff30e1dc18d5d6ed3f1))

## [2.12.3](https://github.com/Firecrown-Media/stax/compare/v2.12.2...v2.12.3) (2025-11-16)


### Bug Fixes

* **release:** use HOMEBREW_TAP_TOKEN for cross-repo GitHub releases ([97a2da4](https://github.com/Firecrown-Media/stax/commit/97a2da49d1b04a3ccccc50f6a84c150f570c4988))

## [2.12.2](https://github.com/Firecrown-Media/stax/compare/v2.12.1...v2.12.2) (2025-11-16)


### Bug Fixes

* **release:** use HOMEBREW_TAP_TOKEN for stax-public releases ([731ddd7](https://github.com/Firecrown-Media/stax/commit/731ddd7a0d92e4340fc80178111d413f072bf3c2))

## [2.12.1](https://github.com/Firecrown-Media/stax/compare/v2.12.0...v2.12.1) (2025-11-16)


### Bug Fixes

* **sync:** exclude .github/workflows from public mirror ([5dc6692](https://github.com/Firecrown-Media/stax/commit/5dc66922959054e38cf753d4cced6f158ea859a1))
* **sync:** recreate tag pointing to cleaned commit ([76305b5](https://github.com/Firecrown-Media/stax/commit/76305b5cd4f61496e3c4789473e6345224330147))
* **sync:** use git rm to remove workflows from index ([3bc9a75](https://github.com/Firecrown-Media/stax/commit/3bc9a757a765639ca25aea1ecbdaa85cc600ba7c))

## [2.12.0](https://github.com/Firecrown-Media/stax/compare/v2.11.0...v2.12.0) (2025-11-16)


### Features

* implement hybrid public mirror for private development ([#81](https://github.com/Firecrown-Media/stax/issues/81)) ([6a1bbf6](https://github.com/Firecrown-Media/stax/commit/6a1bbf6a294b17ccc9e9934173c7d5db33664770))

## [2.11.0](https://github.com/Firecrown-Media/stax/compare/v2.10.0...v2.11.0) (2025-11-16)


### Features

* **ddev:** enable DNS resolution to skip hosts file updates ([#79](https://github.com/Firecrown-Media/stax/issues/79)) ([da7069b](https://github.com/Firecrown-Media/stax/commit/da7069bf0fa36f0eb3e8cf961de53925a1402c5f))

## [2.10.0](https://github.com/Firecrown-Media/stax/compare/v2.9.0...v2.10.0) (2025-11-16)


### Features

* **db:** implement comprehensive snapshot functionality ([#77](https://github.com/Firecrown-Media/stax/issues/77)) ([889cd8b](https://github.com/Firecrown-Media/stax/commit/889cd8bc0df4fc4b1f09da6b587eb0baae6095a9)), closes [#17](https://github.com/Firecrown-Media/stax/issues/17)


### Bug Fixes

* **config:** implement Build config merging in loader ([#74](https://github.com/Firecrown-Media/stax/issues/74)) ([2c27595](https://github.com/Firecrown-Media/stax/commit/2c27595374b9e123645a1b0bc852227c7ccbb79c)), closes [#10](https://github.com/Firecrown-Media/stax/issues/10)

## [2.9.0](https://github.com/Firecrown-Media/stax/compare/v2.8.0...v2.9.0) (2025-11-15)


### Features

* **config:** implement Phase 10 advanced configuration management ([#72](https://github.com/Firecrown-Media/stax/issues/72)) ([2681fdc](https://github.com/Firecrown-Media/stax/commit/2681fdc77ff95e22b2d220e5c9c89a071cfbeefd))

## [2.8.0](https://github.com/Firecrown-Media/stax/compare/v2.7.0...v2.8.0) (2025-11-15)


### Features

* **doctor:** implement Phase 11 enhanced diagnostics ([#70](https://github.com/Firecrown-Media/stax/issues/70)) ([3540d27](https://github.com/Firecrown-Media/stax/commit/3540d27c473e9156c5925e5fe208dd5d3c4d079c))

## [2.7.0](https://github.com/Firecrown-Media/stax/compare/v2.6.0...v2.7.0) (2025-11-15)


### Features

* **db,files:** implement Phases 8 & 9 push capabilities ([#68](https://github.com/Firecrown-Media/stax/issues/68)) ([ef89e8d](https://github.com/Firecrown-Media/stax/commit/ef89e8debc82f147cddb1c55568860dd0bc7287c))

## [2.6.0](https://github.com/Firecrown-Media/stax/compare/v2.5.0...v2.6.0) (2025-11-15)


### Features

* Phase 7 - Enhanced File Operations ([#66](https://github.com/Firecrown-Media/stax/issues/66)) ([ed2d84e](https://github.com/Firecrown-Media/stax/commit/ed2d84e8d385b59506ef37258e9e53fb0fef9dfa))

## [2.5.0](https://github.com/Firecrown-Media/stax/compare/v2.4.0...v2.5.0) (2025-11-15)


### Features

* Phase 12 - WordPress Core Download & wp-config Generation ([#64](https://github.com/Firecrown-Media/stax/issues/64)) ([e0e9cbd](https://github.com/Firecrown-Media/stax/commit/e0e9cbd93f53a7b1957bd7a25b342cf422a47268))

## [2.4.0](https://github.com/Firecrown-Media/stax/compare/v2.3.0...v2.4.0) (2025-11-15)


### Features

* **db:** implement automatic database import and URL search-replace ([#62](https://github.com/Firecrown-Media/stax/issues/62)) ([c8bc93f](https://github.com/Firecrown-Media/stax/commit/c8bc93f6f881deaff3c8fb3ca50a7614c201fb00))

## [2.3.0](https://github.com/Firecrown-Media/stax/compare/v2.2.0...v2.3.0) (2025-11-13)


### Features

* trigger release for Phase 6-11 completion ([4edea49](https://github.com/Firecrown-Media/stax/commit/4edea4919cf454e961b58c73cfd316bc4e777239))

## [2.2.0](https://github.com/Firecrown-Media/stax/compare/v2.1.1...v2.2.0) (2025-11-12)


### Features

* **doctor:** enhance diagnostics and add global WPEngine discovery ([d610297](https://github.com/Firecrown-Media/stax/commit/d61029709d67c7c11a03fff17ff4ee4b8666d8c6))
* **init:** implement complete interactive and non-interactive project initialization ([#47](https://github.com/Firecrown-Media/stax/issues/47)) ([cbcf2e3](https://github.com/Firecrown-Media/stax/commit/cbcf2e36a7711923e5d5e68e95b77b43ec777dcb))
* **media:** implement media proxy configuration commands ([fc11f20](https://github.com/Firecrown-Media/stax/commit/fc11f20cee8505ebac84461900894d99d2efbf2b))
* **ui:** add command status indicators and enhanced version command ([#45](https://github.com/Firecrown-Media/stax/issues/45)) ([609599b](https://github.com/Firecrown-Media/stax/commit/609599be4351c7c2f54aa20b77f6a7058cfdd390))

## [2.1.1](https://github.com/Firecrown-Media/stax/compare/v2.1.0...v2.1.1) (2025-11-11)


### Bug Fixes

* **ci:** update release-please workflow to v4 configuration ([5b87221](https://github.com/Firecrown-Media/stax/commit/5b87221c092b76e3b07219d83403dd2564d2488a))

## [1.1.0](https://github.com/Firecrown-Media/stax/compare/v1.0.0...v1.1.0) (2025-11-10)


### Features

* Enhanced error messaging and UX improvements ([#34](https://github.com/Firecrown-Media/stax/issues/34)) ([84ef266](https://github.com/Firecrown-Media/stax/commit/84ef2662d90c9dc62ce67408634dacdd60300c68))

## [1.0.0](https://github.com/Firecrown-Media/stax/compare/v0.5.0...v1.0.0) (2025-11-10)


### ⚠ BREAKING CHANGES

* Default project type changed from wordpress-multisite to wordpress

### Documentation

* clarify single site and multisite support ([3a2414b](https://github.com/Firecrown-Media/stax/commit/3a2414b294503ebb852d76d7edc380ddf6a78db3))

## [0.5.0](https://github.com/Firecrown-Media/stax/compare/v0.4.2...v0.5.0) (2025-11-10)


### Features

* **list:** add global list command for WPEngine installs ([#30](https://github.com/Firecrown-Media/stax/issues/30)) ([45f73c2](https://github.com/Firecrown-Media/stax/commit/45f73c224feb70037f91fc74a5f34a9099c14bfd))

## [0.4.2](https://github.com/Firecrown-Media/stax/compare/v0.4.1...v0.4.2) (2025-11-10)


### Bug Fixes

* credential storage for CGO-disabled builds (Homebrew) ([#28](https://github.com/Firecrown-Media/stax/issues/28)) ([3379b3a](https://github.com/Firecrown-Media/stax/commit/3379b3ae3fafedeb17457dd7bfe3706446de63e7)), closes [#27](https://github.com/Firecrown-Media/stax/issues/27)

## [0.4.1](https://github.com/Firecrown-Media/stax/compare/v0.4.0...v0.4.1) (2025-11-09)


### Bug Fixes

* resolve platform-specific keychain build issues for releases ([#25](https://github.com/Firecrown-Media/stax/issues/25)) ([e39f9ab](https://github.com/Firecrown-Media/stax/commit/e39f9ab1e638ff8fc7d585b5395767104634cad2))

## [0.4.0](https://github.com/Firecrown-Media/stax/compare/v0.3.0...v0.4.0) (2025-11-09)


### Features

* complete codebase refactor with build system, tests, and release automation ([e09284f](https://github.com/Firecrown-Media/stax/commit/e09284f0f73dc00ce77eb92202b3764bd663f34c))


### Bug Fixes

* disable CGO for Darwin builds to enable cross-compilation ([d65db36](https://github.com/Firecrown-Media/stax/commit/d65db36556b401216cb7f9f1b13510daf2e6a245))

## [0.3.0](https://github.com/Firecrown-Media/stax/compare/v0.2.0...v0.3.0) (2025-11-09)


### Features

* complete codebase refactor with build system, tests, and CI fixes ([9de3d0d](https://github.com/Firecrown-Media/stax/commit/9de3d0dd4bfb4b7862055d2a7c72d02e6004d4f2))


### Bug Fixes

* update GoReleaser action to v6 for version 2 config support ([6865c4f](https://github.com/Firecrown-Media/stax/commit/6865c4fe7d901f090daa29e7cc4cbbbe3f1a73ef))

## [0.2.0](https://github.com/Firecrown-Media/stax/compare/v0.1.1...v0.2.0) (2025-11-09)


### Features

* add automated release process and reorganize documentation ([6d605c5](https://github.com/Firecrown-Media/stax/commit/6d605c51038b0c765000d52f66b57d9e9267f97e))


### Bug Fixes

* add wp_blogs table to mock database dump for multisite tests ([07faa2d](https://github.com/Firecrown-Media/stax/commit/07faa2d2a20533ade31b6a632c3c8d612e370b1d))
* correct mock types and remove omitempty from Build config booleans ([e39a618](https://github.com/Firecrown-Media/stax/commit/e39a6185ec5c4e7046e2b38bb781b87f315bdf1a)), closes [#9](https://github.com/Firecrown-Media/stax/issues/9)
* format code with gofmt to pass CI checks ([c2654f7](https://github.com/Firecrown-Media/stax/commit/c2654f73db9a20c8966797a955611766722d5616))
* format test files with gofmt ([559c111](https://github.com/Firecrown-Media/stax/commit/559c1111fc37c059df89ef8fa24778f938e88359))
* update .gitignore to include pkg/build directory ([bac9821](https://github.com/Firecrown-Media/stax/commit/bac982110fa36d62974cb47d5d95fdd680dad3b4))
