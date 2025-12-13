# Tutorial: WPEngine to GitHub Workflow

This tutorial walks you through setting up a complete WordPress development workflow:
- Local development environment with DDEV
- Git repository with proper WordPress .gitignore
- GitHub Actions for automated deployment to WPEngine

## Prerequisites

Before starting, ensure you have:

```bash
# Install Stax via Homebrew
brew install firecrown-media/stax/stax

# Install DDEV
brew install ddev/ddev/ddev

# Install GitHub CLI
brew install gh

# Ensure Docker Desktop is running
open -a Docker
```

Verify your setup:
```bash
stax doctor
```

## Step 1: Configure Credentials

Run the setup wizard to configure your WPEngine credentials:

```bash
stax setup
```

When prompted:
- **WPEngine API Username**: Your WPEngine portal email
- **WPEngine API Password**: Your WPEngine API password (input is hidden)
- **GitHub Token**: Optional - for automated repo creation
- **SSH Key**: Path to your WPEngine SSH key (default: ~/.ssh/id_rsa)

> **Security Note**: Credentials are stored securely in macOS Keychain or `~/.stax/credentials.yml`. Never commit credentials to version control.

## Step 2: Initialize Git Repository

For a site on WPEngine that doesn't have a GitHub repository yet:

```bash
# Create project directory
mkdir my-wordpress-site
cd my-wordpress-site

# Initialize repository and sync from WPEngine
stax repo init --install mysite-prod --github myorg/my-wordpress-site --private
```

This command:
1. Syncs themes, plugins, and mu-plugins from WPEngine
2. Creates a comprehensive WordPress .gitignore
3. Creates the initial Git commit
4. Creates a private GitHub repository
5. Pushes to GitHub

### What Gets Synced

By default, these directories are synced:
- `wp-content/themes/`
- `wp-content/plugins/`
- `wp-content/mu-plugins/`

### What Gets Excluded

The generated .gitignore excludes:
- WordPress core files (wp-admin/, wp-includes/, etc.)
- Database dumps (mysql.sql, *.sql)
- Uploads directory (wp-content/uploads/)
- Cache and backup directories
- Environment files (.env, wp-config.php)
- DDEV local configuration

## Step 3: Set Up Local Development

Initialize the local DDEV environment:

```bash
stax init --install mysite-prod --start
```

This will:
1. Create `.stax.yml` configuration
2. Configure DDEV for WordPress with matching PHP/MySQL versions
3. Start DDEV containers
4. Pull the database from WPEngine
5. Automatically run URL search-replace

Access your local site:
```bash
ddev launch
```

## Step 4: Configure GitHub Actions

Set up automated deployment to WPEngine:

```bash
stax actions setup --production main --staging develop
```

This creates `.github/workflows/deploy.yml` with:
- Deployment on push to main (production) and develop (staging)
- Node.js build step for theme/plugin assets
- WPEngine's official deployment action

### Add the WPEngine SSH Secret

1. Go to your GitHub repository → Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `WPE_SSHG_KEY_PRIVATE`
4. Value: Your WPEngine SSH private key (entire contents)

Get your key from: https://my.wpengine.com/ → SSH Gateway

### Commit the Workflow

```bash
git add .github/
git commit -m "chore: add GitHub Actions deployment workflow"
git push
```

## Step 5: Development Workflow

### Daily Development

```bash
# Start local environment
stax start

# Pull latest database from WPEngine
stax db pull

# Pull latest files (themes/plugins)
stax files pull

# Work on your code...

# Commit and push
git add .
git commit -m "feat: add new feature"
git push
```

### Deployment

Pushes to configured branches trigger automatic deployment:
- Push to `main` → Deploys to production WPEngine install
- Push to `develop` → Deploys to staging WPEngine install

## Common Issues and Solutions

### Search-Replace Shows 0 Replacements

If the database pull shows "0 replacements", the URL pattern may not match. Configure the exact URL in `.stax.yml`:

```yaml
wpengine:
  install: mysite-prod
  environment: staging
  domains:
    staging:
      primary: mysitestage.wpengine.com  # Set exact URL
```

### Environment Mismatch Warning

If you see "Environment mismatch detected", run:

```bash
stax doctor --fix
```

This automatically updates `.stax.yml` to match the WPEngine environment.

### mysql.sql in Files

Database dumps are now excluded by default. If you have an old `mysql.sql` file:

```bash
rm wp-content/mysql.sql
echo "mysql.sql" >> .gitignore
```

Or add to `.staxignore` for file sync exclusion.

## Project Structure

After setup, your project should look like:

```
my-wordpress-site/
├── .ddev/
│   └── config.yaml          # DDEV configuration
├── .github/
│   ├── workflows/
│   │   └── deploy.yml       # Deployment workflow
│   └── CODEOWNERS           # Code ownership
├── .gitignore               # WordPress-optimized
├── .stax.yml                # Stax configuration
├── wp-content/
│   ├── themes/              # Your themes (tracked)
│   ├── plugins/             # Your plugins (tracked)
│   └── mu-plugins/          # Must-use plugins (tracked)
└── README.md
```

## Next Steps

- **Set up branch protection**: Require PRs for production deployments
- **Configure CODEOWNERS**: Assign reviewers for different parts of the codebase
- **Add CI checks**: Lint, test, and validate before deployment
- **Set up media proxy**: Avoid syncing large upload directories

```bash
stax media setup
```

## Reference Commands

| Command | Description |
|---------|-------------|
| `stax doctor` | Check system health and fix issues |
| `stax start` | Start local DDEV environment |
| `stax stop` | Stop local environment |
| `stax db pull` | Pull database from WPEngine |
| `stax db push` | Push database to WPEngine |
| `stax files pull` | Sync files from WPEngine |
| `stax status` | Show environment status |
| `stax shell` | SSH into DDEV container |

## Further Reading

- [WPEngine GitHub Action Documentation](https://github.com/wpengine/github-action-wpe-site-deploy)
- [DDEV Documentation](https://ddev.readthedocs.io/)
- [Stax Command Reference](../reference/commands.md)
