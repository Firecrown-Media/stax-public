Everything you need to set up stax and pull your first WPEngine site to a local DDEV environment.

## Prerequisites

| Requirement | Notes |
|-------------|-------|
| Docker | Docker Desktop or Colima |
| DDEV | [Installation guide](https://ddev.readthedocs.io/en/stable/users/install/) |
| macOS 12+ | Keychain credential storage requires macOS |
| WPEngine account | SSH key must be configured in your WPEngine portal |

## Install stax

```bash
brew install firecrown-media/stax/stax
```

Verify the installation:

```bash
stax --version
```

## Store credentials

Run the setup command to store your WPEngine credentials:

```bash
stax setup
```

The command prompts for three values:

- **WPEngine username** — your WPEngine portal username
- **API token** — generated in the WPEngine portal under API Access
- **SSH key path** — the private key registered with WPEngine (e.g. `~/.ssh/id_ed25519`)

Credentials are stored in macOS Keychain under `com.firecrownmedia.stax`. They are never written to `.stax.yml` or any config file.

## Initialize a project

Change into an existing DDEV project directory and run:

```bash
cd my-ddev-project
stax init
```

This creates `.stax.yml` in the project root. A minimal config looks like:

```yaml
version: 2
provider: wpengine
provider_config:
  install: my-install
  environment: production
```

Replace `my-install` with your WPEngine install name. Set `environment` to `production` or `staging`.

## Pull the database

```bash
stax db pull
```

This command:

1. Exports the database from WPEngine via SSH
2. Imports it into your local DDEV environment
3. Runs URL search-replace, swapping the WPEngine URL (e.g. `my-install.wpengine.com`) for your local `.ddev.site` domain

## Pull files

```bash
stax files pull --exclude-uploads
```

The `--exclude-uploads` flag skips `wp-content/uploads`, which is typically large. Upload assets can be served directly from WPEngine by configuring the media proxy instead — see the [media proxy runbook](media-proxy.md).

## Start the site

```bash
stax start
```

Once DDEV starts, open the local `.ddev.site` URL printed in the output to view your site.

## What not to do

- Do not run `stax db push` to production without first reviewing `--dry-run` output. A push to production overwrites the live database.
- Do not put credentials in `.stax.yml`. All secrets belong in Keychain, where `stax setup` puts them. The config file is safe to commit.
