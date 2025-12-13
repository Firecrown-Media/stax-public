# How to Initialize a Git Repository for an Existing Site

This guide shows you how to set up a Git repository for a WordPress site that already exists on WPEngine but doesn't have version control.

## Prerequisites

- Stax installed and configured
- WPEngine credentials set up (`stax setup`)
- Git installed
- GitHub CLI (`gh`) installed (optional, for automatic repo creation)

## Steps

### 1. Create Your Project Directory

```bash
mkdir my-wordpress-site
cd my-wordpress-site
```

### 2. Initialize the Repository

```bash
stax repo init --install mysite-prod
```

This command will:
1. Initialize a Git repository
2. Generate a WordPress-appropriate `.gitignore`
3. Sync files from WPEngine:
   - `wp-content/themes/`
   - `wp-content/plugins/`
   - `wp-content/mu-plugins/`
4. Create an initial commit

### 3. (Optional) Create a GitHub Repository

If you have the GitHub CLI installed:

```bash
stax repo init --install mysite-prod --github myorg/my-wordpress-site
```

This additionally:
- Creates a private GitHub repository
- Adds the remote
- Pushes the initial commit

### 4. Verify the Setup

```bash
# Check Git status
git status

# View the .gitignore
cat .gitignore

# List synced files
ls -la wp-content/themes/
ls -la wp-content/plugins/
```

## Customizing What Gets Synced

By default, Stax syncs:
- `wp-content/themes/`
- `wp-content/plugins/`
- `wp-content/mu-plugins/`

To customize:

```bash
stax repo init --install mysite-prod \
  --sync-dirs wp-content/themes,wp-content/plugins
```

## What's Excluded

The generated `.gitignore` excludes:
- WordPress core files (`wp-admin/`, `wp-includes/`)
- Uploads directory (`wp-content/uploads/`)
- Database dumps (`mysql.sql`)
- Build artifacts (`node_modules/`)
- Environment files (`.env`)
- DDEV local files (`.ddev/`)

## Next Steps

After initializing the repository:

1. [Set up GitHub Actions](setup-github-actions.md) for automated deployments
2. [Configure branch protection](configure-branch-protection.md) for code review
3. [Initialize Stax](../tutorials/first-project.md) for local development

## See Also

- [Reference: WordPress .gitignore Template](../reference/gitignore-template.md)
- [Explanation: Git Workflow](../explanation/git-workflow.md)
