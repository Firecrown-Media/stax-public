# Contributing

Stax is developed in the private repo `github.com/Firecrown-Media/stax`. The public mirror at `github.com/Firecrown-Media/stax-public` is updated automatically — do not open PRs there.

## Dev setup

```bash
git clone https://github.com/Firecrown-Media/stax.git
cd stax
go mod download
PATH="/opt/homebrew/bin:$PATH" make build
```

Go is installed via Homebrew. If `make build` fails with "go: command not found", prefix with `PATH="/opt/homebrew/bin:$PATH"`.

## Running tests

```bash
make test          # unit tests with race detection
make lint          # golangci-lint
make fmt           # gofmt — CI fails without this, run before every commit
make vet           # go vet
make test-integration  # requires RUN_INTEGRATION_TESTS=true
```

## Branch and commit conventions

- Branch off `main`: `feat/short-description`, `fix/short-description`, `chore/short-description`
- Use conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `test:`, `refactor:`
- Keep commits focused — one logical change per commit

## Submitting a PR

Open PRs against `main` in the private repo. Include:

- What changed and why (not just what — the diff shows what)
- Any manual testing steps that aren't covered by `make test`

CI runs `make fmt`, `make vet`, `make lint`, and `make test` on every PR.

## Releases

Releases are automated via release-please. Merging a `feat:` or `fix:` commit to `main` triggers a release PR. Do not create tags or push to `stax-public` manually.
