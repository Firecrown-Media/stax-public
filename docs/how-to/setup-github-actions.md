# How to Set Up GitHub Actions for WPEngine Deployment

This guide shows you how to configure automated deployments to WPEngine using GitHub Actions.

## Prerequisites

- A Git repository for your WordPress project
- WPEngine SSH Gateway access
- GitHub repository with admin access

## Steps

### 1. Generate the Workflow

From your project directory:

```bash
stax actions setup
```

Or with specific branch configuration:

```bash
stax actions setup \
  --production main \
  --staging develop \
  --prod-install mysite-prod \
  --stage-install mysite-staging
```

This creates:
- `.github/workflows/deploy.yml` - Deployment workflow
- `.github/CODEOWNERS` template (if it doesn't exist)

### 2. Add GitHub Secrets

The workflow requires your WPEngine SSH private key. Add it to GitHub:

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `WPE_SSHG_KEY_PRIVATE`
5. Value: Your WPEngine SSH private key (entire contents including headers)

**Getting your WPEngine SSH key:**
1. Log in to [WPEngine User Portal](https://my.wpengine.com/)
2. Go to **SSH Gateway**
3. Generate or copy your private key

### 3. Configure Branch Protection (Recommended)

Protect your deployment branches:

1. Go to **Settings** → **Branches**
2. Add rule for `main`:
   - ✅ Require pull request before merging
   - ✅ Require approvals (at least 1)
   - ✅ Require status checks to pass
   - ✅ Require conversation resolution

3. Add rule for `develop` (if using staging):
   - ✅ Require status checks to pass

### 4. Test the Deployment

Create a test commit and push:

```bash
# Make a small change
echo "# Test" >> README.md
git add README.md
git commit -m "test: verify GitHub Actions deployment"
git push
```

Monitor the deployment:
1. Go to **Actions** tab in GitHub
2. Click on the running workflow
3. Watch the deployment logs

## Workflow Configuration

The generated workflow deploys:
- `main` branch → Production environment
- `develop` branch → Staging environment (if configured)

### Customizing the Workflow

Edit `.github/workflows/deploy.yml` to:

**Change build commands:**
```yaml
- name: Build assets
  run: |
    npm ci
    npm run build
    composer install --no-dev
```

**Add additional checks:**
```yaml
- name: Run tests
  run: npm test

- name: Run PHP linting
  run: composer lint
```

**Change deployed directories:**
```yaml
SRC_PATH: "wp-content/themes/my-theme/"
REMOTE_PATH: "wp-content/themes/my-theme/"
```

## Troubleshooting

### Deployment fails with SSH error

- Verify the `WPE_SSHG_KEY_PRIVATE` secret is correctly set
- Ensure the key has proper line breaks (not all on one line)
- Check that your WPEngine user has SSH access

### Files not deploying

- Check the `SRC_PATH` and `REMOTE_PATH` in the workflow
- Verify files aren't excluded by the `FLAGS` rsync options
- Check the workflow logs for specific errors

### Workflow not triggering

- Ensure you're pushing to the correct branch (`main` or `develop`)
- Check that the workflow file is valid YAML
- Verify the workflow is enabled in GitHub Actions settings

## Next Steps

- [Configure branch protection](configure-branch-protection.md) for code review workflows
- [Set up CODEOWNERS](configure-branch-protection.md#codeowners) for automatic reviewers

## See Also

- [Reference: GitHub Actions Workflows](../reference/github-actions.md)
- [Explanation: Git Workflow](../explanation/git-workflow.md)
