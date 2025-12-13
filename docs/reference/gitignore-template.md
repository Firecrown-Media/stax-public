# WordPress .gitignore Template

This is the default `.gitignore` template that Stax generates when you run `stax repo init`.

## Template

```gitignore
# WordPress Core (don't commit - install via composer or download)
/wp-admin/
/wp-includes/
/wp-*.php
/index.php
/license.txt
/readme.html
/xmlrpc.php

# WPEngine specific
mysql.sql
.smushit-status
.wpengine-conf/

# wp-content exceptions
/wp-content/uploads/
/wp-content/cache/
/wp-content/upgrade/
/wp-content/backup*/
/wp-content/backups/
/wp-content/blogs.dir/
/wp-content/debug.log
/wp-content/advanced-cache.php
/wp-content/object-cache.php
/wp-content/wp-cache-config.php

# Build artifacts in themes/plugins
/wp-content/themes/*/node_modules/
/wp-content/plugins/*/node_modules/
/wp-content/themes/*/.sass-cache/
/wp-content/plugins/*/.sass-cache/

# Dependency directories
/vendor/
/node_modules/

# Environment and local config
.env
.env.*
wp-config-local.php
*.local.php

# IDE and editor files
.idea/
.vscode/
*.swp
*.swo
.DS_Store
Thumbs.db

# Stax local files
.stax.local.yml

# DDEV (local development)
.ddev/
!.ddev/config.yaml
!.ddev/commands/

# Logs
*.log
logs/
```

## Explanation

### WordPress Core Files

```gitignore
/wp-admin/
/wp-includes/
/wp-*.php
```

WordPress core files should not be committed. They should be:
- Downloaded fresh during deployment
- Installed via Composer (recommended for production)
- Managed by WPEngine (they update core automatically)

### WPEngine Files

```gitignore
mysql.sql
.smushit-status
.wpengine-conf/
```

These files are created by WPEngine:
- `mysql.sql` - Database dump, should never be committed (security risk)
- `.smushit-status` - Image optimization status
- `.wpengine-conf/` - WPEngine configuration

### wp-content Directories

```gitignore
/wp-content/uploads/
/wp-content/cache/
```

- `uploads/` - Media files are large and belong on the server/CDN
- `cache/` - Generated files, environment-specific

### Build Artifacts

```gitignore
/wp-content/themes/*/node_modules/
```

Build dependencies should be installed during deployment, not committed.

### DDEV Configuration

```gitignore
.ddev/
!.ddev/config.yaml
!.ddev/commands/
```

Only commit DDEV configuration that should be shared with the team:
- `config.yaml` - Main DDEV configuration
- `commands/` - Custom DDEV commands

## Customization

### Adding Files to Track

If you need to track files that are normally ignored, use negation:

```gitignore
# Ignore all uploads
/wp-content/uploads/

# But track specific files
!/wp-content/uploads/.htaccess
!/wp-content/uploads/index.php
```

### Project-Specific Additions

Add to the end of your `.gitignore`:

```gitignore
# Project-specific ignores
/wp-content/themes/my-theme/dist/
/wp-content/plugins/my-plugin/build/
```

## See Also

- [How to: Initialize Git Repository](../how-to/init-git-repo.md)
- [Explanation: Git Workflow](../explanation/git-workflow.md)
