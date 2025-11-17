# Stax Troubleshooting Guide

Common problems and their solutions.

---

## Table of Contents

- [Getting Help](#getting-help)
- [Installation Issues](#installation-issues)
- [List Command Issues](#list-command-issues)
- [Database Problems](#database-problems)
- [DDEV and Container Issues](#ddev-and-container-issues)
- [Network and Connectivity](#network-and-connectivity)
- [SSL Certificate Errors](#ssl-certificate-errors)
- [Multisite Issues](#multisite-issues)
- [Build and Development](#build-and-development)
- [Performance Issues](#performance-issues)
- [WPEngine Integration](#wpengine-integration)
- [Common Error Messages](#common-error-messages)

---

## Getting Help

Before diving into specific issues, here's how to diagnose problems:

### Run Stax Doctor

```bash
stax doctor
```

This checks:
- Stax installation
- DDEV installation and version
- Docker Desktop status
- Credentials validity
- Port availability
- SSL certificates
- Common configuration issues

**Expected output when healthy**:
```
🩺 Running diagnostics...

✓ Stax installed (v2.12.5)
✓ DDEV installed (v1.22.7)
✓ Docker Desktop running
✓ WPEngine credentials valid
✓ GitHub token valid
✓ Ports 80, 443, 8025 available
✓ SSL certificates valid

All checks passed!
```

### Check Logs

**Stax logs**:
```bash
# View Stax logs
cat ~/.stax/logs/stax.log

# Follow Stax logs
tail -f ~/.stax/logs/stax.log
```

**Container logs**:
```bash
# All logs
stax logs -f

# Web container only
stax logs --service=web

# Database container
stax logs --service=db

# Last 500 lines
stax logs --tail=500
```

### Check Status

```bash
stax status
```

Shows:
- Container health
- URLs
- Configuration
- Database info

---

## Installation Issues

### "Command not found: stax"

**Problem**: Shell can't find the `stax` command.

**Causes and solutions**:

**1. Stax not installed**:
```bash
# Install via Homebrew
brew install stax

# Or build from source
cd ~/path/to/stax
make install
```

**2. Stax not in PATH**:
```bash
# Check where stax is
which stax

# Should show: /usr/local/bin/stax

# If not, add to PATH
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**3. Installed with brew but PATH not updated**:
```bash
# For Apple Silicon
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc

# For Intel
echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zshrc

source ~/.zshrc
```

### "DDEV not found"

**Problem**: Stax can't find DDEV.

**Solution**:
```bash
# Install DDEV
brew tap ddev/ddev
brew install ddev

# Verify
ddev version

# If installed but not found, check PATH
which ddev
```

### "Docker daemon not running"

**Problem**: Docker Desktop isn't running.

**Solution**:
1. Open Docker Desktop from Applications
2. Wait for it to start (green icon in menu bar)
3. Try again:
   ```bash
   docker ps
   ```

**If Docker won't start**:
1. Restart your Mac
2. Check disk space (need 10GB+ free)
3. Reset Docker: Settings → Troubleshoot → Reset to factory defaults
4. Reinstall Docker Desktop

### "Permission denied" installing Stax

**Problem**: Don't have permission to write to `/usr/local/bin`.

**Solution**:
```bash
# Use sudo
sudo make install

# Or install to home directory
mkdir -p ~/bin
cp stax ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### mkcert installation fails

**Problem**: Can't install mkcert (needed for SSL).

**Solution**:
```bash
# Install mkcert
brew install mkcert

# Install certificates
mkcert -install

# For Firefox support
brew install nss
```

---

## List Command Issues

### "WPEngine credentials not found"

**Symptom**:
```
Error: WPEngine credentials not found

Please configure your credentials using one of these methods:
...
```

**Solution**:

Stax supports multiple credential storage methods:

**Option 1: Interactive Setup** (Recommended for Development):
```bash
# Configure credentials interactively
stax setup
```
Note: Keychain storage is only available when building from source with CGO enabled. Homebrew builds use config file storage instead.

**Option 2: Environment Variables** (Recommended for CI/CD):
```bash
# Set in your shell profile (~/.zshrc or ~/.bashrc)
export WPENGINE_API_USER="your-username"
export WPENGINE_API_PASSWORD="your-password"
```

**Option 3: Config File** (Alternative):
```bash
# Create credentials file
mkdir -p ~/.stax
cat > ~/.stax/credentials.yml <<EOF
wpengine:
  api_user: your-username
  api_password: your-password
EOF
chmod 600 ~/.stax/credentials.yml
```

**For detailed credential setup instructions**, see [WPENGINE.md](WPENGINE.md#getting-your-wpengine-api-credentials)

### "Authentication failed"

**Symptom**:
```
Error: authentication failed: invalid credentials
```

**Causes and solutions**:

**1. Wrong API credentials**:
```bash
# Reconfigure with correct credentials
stax setup

# Verify credentials in WPEngine portal
# Go to: Account Settings > API Access
```

**2. API access not enabled**:
Your account must have API access enabled (requires Owner role):
- Log in to [WPEngine User Portal](https://my.wpengine.com)
- Go to Account Settings > API Access
- Enable API access (Owner role required)
- Create new API credentials if needed
- See [Enabling WPEngine API](https://wpengine.com/support/enabling-wp-engine-api/) for official instructions

**3. Insufficient account permissions**:
Your user role may not have the necessary permissions:
- Check your role at [WPEngine User Portal](https://my.wpengine.com)
- You need Owner role to enable API, or appropriate role for access
- See [WPEngine User Roles](https://wpengine.com/support/users/) for role descriptions
- Contact your account owner if you need additional permissions

**4. Incorrect username format**:
```bash
# Should be your email or API username
# NOT your WPEngine login username
stax setup
# Enter: yourname@company.com (not just "yourname")
```

### "Connection test failed"

**Symptom**:
```
Error: connection test failed: timeout
```

**Solutions**:

**1. Network issue**:
```bash
# Check internet connection
ping api.wpengine.com

# Check if behind VPN/proxy
# Try without VPN
```

**2. Firewall blocking**:
```bash
# Check if firewall allows outbound HTTPS
curl -I https://api.wpengine.com
```

**3. WPEngine API down**:
```bash
# Check WPEngine status page
# https://status.wpengine.com
```

### "No installs found"

**Symptom**:
```
No installs found matching your criteria

Tips:
  - Check your filter/environment flags
  - Verify your WPEngine account has install access
  - Run without filters: stax list
```

**Causes and solutions**:

**1. Filters too restrictive**:
```bash
# Remove filters
stax list

# Check filter syntax
stax list --filter=".*"  # Should match all

# Try simpler filter
stax list --filter="myinstall"
```

**2. No installs in account**:
- Verify you have access to WPEngine installs
- Check with WPEngine account admin
- Ensure API user has correct permissions

**3. Wrong API credentials**:
```bash
# Verify credentials
stax setup
```

### API Access Permissions Issues

**Symptom**: Can't enable API access or can't access certain installs even with valid credentials.

**Understanding WPEngine User Roles**:

WPEngine has different user roles with varying permissions. See [WPEngine User Roles](https://wpengine.com/support/users/) for complete details.

**Common scenarios**:

**1. Can't enable API access**:
- **Cause**: Only users with the Owner role can enable API access
- **Solution**:
  - Contact your account owner to enable API access
  - Or have them upgrade your role to Owner
  - Check your role at [WPEngine User Portal](https://my.wpengine.com)

**2. API enabled but can't access installs**:
- **Cause**: Your user role (Full User or Partial User) may not have access to specific installs
- **Solution**:
  - Check which installs you have access to at [WPEngine User Portal](https://my.wpengine.com)
  - Contact your account owner to grant access to specific installs
  - Partial Users only have access to installs they've been explicitly granted

**3. Need to verify your permissions**:
```bash
# Test your access
stax list

# This will show all installs you have access to
# If you don't see an install, you don't have permission to access it
```

**4. Requesting access**:
- Go to [WPEngine User Portal](https://my.wpengine.com)
- Check your current role and install access
- Contact your account owner or administrator
- Provide specific install names you need access to
- Reference the [WPEngine Users Guide](https://wpengine.com/support/users/) for role requirements

**Role summary**:
- **Owner**: Full access, can enable API, manage users
- **Full User**: Access to all installs, but can't enable API
- **Partial User**: Access only to specific installs, can't enable API

### List command is slow

**Symptom**: `stax list` takes more than 10 seconds.

**Causes and solutions**:

**1. Many installs**:
```bash
# Filter to reduce results
stax list --filter="client.*"
stax list --environment=production
```

**2. Network latency**:
```bash
# Check network speed
curl -w "@-" -o /dev/null -s https://api.wpengine.com <<EOF
    time_namelookup:  %{time_namelookup}\n
       time_connect:  %{time_connect}\n
    time_appconnect:  %{time_appconnect}\n
      time_redirect:  %{time_redirect}\n
   time_starttransfer:  %{time_starttransfer}\n
                      ----------\n
          time_total:  %{time_total}\n
EOF
```

**3. API rate limiting**:
- Wait a few minutes and try again
- WPEngine may be rate limiting your requests

### Output format issues

**JSON output malformed**:
```bash
# Ensure no other output is mixed in
stax list --output=json 2>/dev/null

# Validate JSON
stax list --output=json | jq .
```

**YAML output not parsing**:
```bash
# Ensure no other output
stax list --output=yaml 2>/dev/null

# Validate YAML
stax list --output=yaml | yq .
```

**Table format misaligned**:
- This is usually due to very long domain names
- Use JSON or YAML for machine-readable output
- Filter to reduce results: `stax list --filter="short-name"`

---

## Database Problems

### Database import fails

**Symptom**:
```
Error: Failed to import database
```

**Causes and solutions**:

**1. WPEngine connection failed**:
```bash
# Test connection
ssh -i ~/.ssh/wpengine git@git.wpengine.com info

# If fails, check:
# - SSH key is added to WPEngine
# - SSH key path is correct in stax setup
stax setup
```

**2. Database too large**:
```bash
# Import only essential data
stax db pull --skip-logs --skip-transients --skip-spam
```

**3. MySQL crash**:
```bash
# Restart database
stax restart

# Check database logs
stax logs --service=db
```

**4. Disk space full**:
```bash
# Check disk space
df -h

# Clean up Docker
docker system prune -a

# Clean up old snapshots
stax db list
stax db delete-snapshot old-snapshot-name
```

### Search-replace doesn't work

**Symptom**: URLs still show production domains after import.

**Solution**:
```bash
# Manual search-replace
stax wp search-replace \
  'production.com' \
  'local.local' \
  --all-tables

# For multisite, do each site
stax wp search-replace \
  'site1.com' \
  'site1.local.local' \
  --url=site1.com

# Flush cache
stax wp cache flush --network
```

### Can't connect to database

**Symptom**:
```
Error establishing a database connection
```

**Solutions**:

**1. Container not running**:
```bash
stax status
# If stopped:
stax start
```

**2. Database crashed**:
```bash
stax logs --service=db
# Check for errors

# Restart
stax restart
```

**3. wp-config.php wrong**:
```bash
stax ssh
cat wp-config.php | grep DB_HOST
# Should be: db

cat wp-config.php | grep DB_NAME
# Should be: db

exit
```

**4. Corrupted database**:
```bash
# Restore from snapshot
stax db list
stax db restore latest

# Or pull fresh from WPEngine
stax db pull
```

### Database snapshot fails

**Symptom**:
```
Error: Failed to create snapshot
```

**Solutions**:

**1. Disk space full**:
```bash
df -h

# Clean up old snapshots
stax db list
stax db delete-snapshot old-snapshot
```

**2. Permissions issue**:
```bash
# Check snapshots directory
ls -la ~/.stax/snapshots/

# Fix permissions
chmod -R 755 ~/.stax/snapshots/
```

**3. Database locked**:
```bash
# Wait for other operations to complete
# Or restart
stax restart
```

---

## DDEV and Container Issues

### Containers won't start

**Symptom**:
```
Failed to start containers
```

**Solutions**:

**1. Port conflicts**:
```bash
# Check what's using ports
sudo lsof -i :80
sudo lsof -i :443

# Stop conflicting services
sudo apachectl stop  # Apache
# Or stop other DDEV projects
ddev poweroff
```

**2. Docker out of resources**:
```bash
# Increase Docker Desktop resources
# Settings → Resources → Advanced
# Memory: 4GB minimum, 8GB recommended
# CPUs: 2 minimum, 4 recommended
```

**3. Corrupted containers**:
```bash
# Remove and recreate
stax stop
ddev delete -Oy
stax start
```

**4. DDEV version too old**:
```bash
# Update DDEV
brew update
brew upgrade ddev

ddev version
# Should be 1.22+
```

### Containers start but site won't load

**Symptom**: Containers running but site shows "502 Bad Gateway" or doesn't load.

**Solutions**:

**1. Check container health**:
```bash
stax status
# All containers should show (healthy)

# If unhealthy, check logs
stax logs --service=web
```

**2. nginx configuration error**:
```bash
stax ssh
nginx -t
# Should show: syntax is ok

# If errors, check config
cat /etc/nginx/sites-enabled/wordpress.conf
exit
```

**3. PHP crashed**:
```bash
stax logs --service=web | grep php
# Look for errors

# Restart
stax restart
```

**4. File permissions**:
```bash
stax ssh
ls -la
# web user should own files

# Fix if needed (rarely needed)
sudo chown -R www-data:www-data .
exit
```

### Container keeps restarting

**Symptom**: Container starts then stops repeatedly.

**Solution**:
```bash
# Check logs for errors
stax logs --service=web
stax logs --service=db

# Common causes:
# - PHP syntax error in wp-config.php
# - Database corruption
# - Out of memory

# Reset environment
stax stop
ddev delete -Oy
stax init
```

---

## Network and Connectivity

### Can't access site (ERR_NAME_NOT_RESOLVED)

**Symptom**: Browser can't resolve domain.

**Solutions**:

**1. Router container not running**:
```bash
docker ps | grep ddev-router
# Should show running router

# If not:
ddev poweroff
stax start
```

**2. Wrong domain**:
```bash
# Check configured domain
stax config get network.domain

# Should match what you're accessing
# e.g., my-project.local

# Update browser URL to match
```

**3. DNS cache**:
```bash
# Flush DNS cache (macOS)
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder

# Restart browser
```

### Subdomain not accessible

**Symptom**: Main site works, but `site1.my-project.local` doesn't.

**Solutions**:

**1. Missing from DDEV config**:
```bash
# Check DDEV config
cat .ddev/config.yaml | grep additional_fqdns

# Should include:
# - "*.my-project.local"
# - site1.my-project.local

# If missing, regenerate
stax restart
```

**2. Site doesn't exist in WordPress**:
```bash
stax wp site list
# Should show the site

# If missing, create it
stax wp site create --slug=site1
```

**3. Wrong URL in database**:
```bash
stax wp option get siteurl --url=site1.my-project.local
# Should be: https://site1.my-project.local

# If wrong, fix it
stax wp search-replace \
  'wrong-domain.com' \
  'site1.my-project.local' \
  --url=site1.my-project.local
```

### Slow site performance

**Symptom**: Site loads very slowly locally.

**Solutions**:

**1. Enable Mutagen** (better file sync):
```bash
# Stop project
stax stop

# Enable mutagen in DDEV
echo "mutagen_enabled: true" >> .ddev/config.yaml

# Restart
stax start
```

**2. Increase Docker resources**:
- Docker Desktop → Settings → Resources
- Memory: 8GB
- CPUs: 4
- Swap: 2GB

**3. Disable Xdebug** (if enabled):
```bash
stax ssh
ddev xdebug off
exit
```

**4. Skip remote media** (if proxying is slow):
- Download uploads folder instead of proxying
- Or use local dummy images

**5. Optimize database**:
```bash
# Skip unnecessary tables
stax db pull --skip-logs --skip-transients
```

---

## SSL Certificate Errors

### Browser shows "Not secure"

**Symptom**: Browser warns about certificate.

**Solutions**:

**1. Accept the certificate** (easiest):
- Chrome: Click "Advanced" → "Proceed to site (unsafe)"
- Firefox: Click "Advanced" → "Accept Risk"
- Safari: Click "Show Details" → "visit this website"

**2. Trust mkcert CA**:
```bash
# Reinstall certificates
mkcert -install

# Restart browser
```

**3. Regenerate certificates**:
```bash
stax restart
```

**4. For Firefox** (needs extra setup):
```bash
brew install nss
mkcert -install
```

### Certificate expired

**Symptom**:
```
NET::ERR_CERT_DATE_INVALID
```

**Solution**:
```bash
# Regenerate certificates
ddev delete -Oy
stax start

# Or manually
stax ssh
rm -rf /etc/ssl/certs/master.*
exit
stax restart
```

### Mixed content warnings

**Symptom**: Page loads but some assets load over HTTP instead of HTTPS.

**Solution**:
```bash
# Update URLs in database
stax wp search-replace 'http://' 'https://' --dry-run
# Check output, then run for real:
stax wp search-replace 'http://' 'https://'

# Clear cache
stax wp cache flush
```

---

## Multisite Issues

See [MULTISITE.md](./MULTISITE.md) for detailed multisite troubleshooting.

**Quick fixes**:

### Can't access network admin

```bash
# Make yourself super admin
stax wp super-admin add your-username
```

### Site shows wrong content

```bash
# Fix site URL
stax wp option update siteurl 'https://correct-domain.local' \
  --url=current-domain.com

# Run search-replace
stax wp search-replace \
  'wrong-domain.com' \
  'correct-domain.local' \
  --url=wrong-domain.com
```

### New site not working

```bash
# Verify site exists
stax wp site list

# Check site URL
stax wp option get siteurl --url=site.domain.local

# Add to .stax.yml
# ... add site configuration ...

# Restart
stax restart
```

---

## Build and Development

### Build fails

**Symptom**:
```
Build failed with errors
```

**Solutions**:

**1. Dependencies not installed**:
```bash
stax ssh
composer install
npm install
exit

stax build
```

**2. Build script error**:
```bash
# Check build logs
stax logs -f

# Run build manually to see errors
stax ssh
bash scripts/build.sh
# Look for errors
exit
```

**3. Node/PHP version mismatch**:
```bash
# Check versions
stax ssh
node --version
php --version
exit

# Update in .stax.yml if needed
stax config set ddev.php_version 8.2
stax config set ddev.nodejs_version 20
stax restart
```

### Linter failures

**Symptom**:
```
Linting failed: X errors
```

**Solution**:
```bash
# Auto-fix (when possible)
stax lint --fix

# Or fix manually
# Read error output
stax lint

# Fix each error in your editor
# Run again
stax lint
```

### Watch mode not detecting changes

**Symptom**: Files change but build doesn't trigger.

**Solutions**:

**1. Restart watch**:
```bash
# Stop (Ctrl+C)
# Start again
stax dev
```

**2. Check watch paths**:
```bash
# Edit .stax.yml
build:
  watch:
    paths:
      - wp-content/themes/*/assets/**
      - wp-content/mu-plugins/*/assets/**
```

**3. File system events** (macOS):
```bash
# Install fswatch
brew install fswatch

# Or use mutagen
echo "mutagen_enabled: true" >> .ddev/config.yaml
stax restart
```

---

## Performance Issues

### High CPU usage

**Causes**:
- Docker Desktop
- File sync (especially without Mutagen)
- Build processes running

**Solutions**:

**1. Enable Mutagen**:
```bash
echo "mutagen_enabled: true" >> .ddev/config.yaml
stax restart
```

**2. Stop watch mode** when not needed:
```bash
# If stax dev is running
# Press Ctrl+C
```

**3. Stop unused projects**:
```bash
ddev poweroff
```

**4. Limit Docker resources**:
- Docker Desktop → Settings → Resources
- Don't give Docker ALL your CPU/RAM
- Leave some for macOS

### High RAM usage

**Solutions**:

**1. Stop containers** when not using:
```bash
stax stop
```

**2. Reduce Docker memory allocation**:
- Docker Desktop → Settings → Resources
- Memory: 4-6GB (instead of 8GB)

**3. Clean up**:
```bash
# Remove old containers
docker system prune -a

# Remove old snapshots
stax db list
stax db delete-snapshot old-snapshot
```

### Disk space running out

**Solutions**:

**1. Clean up Docker**:
```bash
docker system prune -a --volumes
```

**2. Clean up snapshots**:
```bash
stax db list
stax db delete-snapshot old-snapshot-name
```

**3. Clean up DDEV snapshots**:
```bash
ddev snapshot --cleanup
```

**4. Check disk usage**:
```bash
# Overall
df -h

# Docker
docker system df

# Stax snapshots
du -sh ~/.stax/snapshots/*
```

---

## WPEngine Integration

### "Keychain storage not available" error

**Symptom**:
```
⚠ macOS Keychain storage is not available in this build
Error: keychain storage is not supported - please use environment variables or config files
```

**Explanation**:
This is normal for Homebrew installations. Homebrew builds use `CGO_ENABLED=0` for better compatibility and can't access macOS Keychain APIs.

**Solutions**:

**Option 1: Use Environment Variables** (Recommended for CI/CD):

Add to your `~/.zshrc` or `~/.bashrc`:
```bash
export WPENGINE_API_USER="your-api-username"
export WPENGINE_API_PASSWORD="your-api-password"
export WPENGINE_SSH_GATEWAY="ssh.wpengine.net"
export GITHUB_TOKEN="ghp_your_token_here"
```

Reload your shell:
```bash
source ~/.zshrc
```

**Option 2: Use Config File** (Recommended for Development):

Create `~/.stax/credentials.yml`:
```yaml
wpengine:
  api_user: "your-api-username"
  api_password: "your-api-password"
  ssh_gateway: "ssh.wpengine.net"

github:
  token: "ghp_your_token_here"

ssh:
  private_key_path: "~/.ssh/wpengine"
```

Secure the file:
```bash
chmod 600 ~/.stax/credentials.yml
```

**Option 3: Build from Source with CGO**:

If you need Keychain support:
```bash
# Clone the repository
git clone https://github.com/firecrown-media/stax.git
cd stax

# Build with CGO enabled
CGO_ENABLED=1 go build -o stax .

# Install
sudo mv stax /usr/local/bin/

# Now stax setup will work with Keychain
stax setup
```

**Verify your setup**:
```bash
# Check if credentials are loaded
stax doctor

# Test WPEngine connection
ssh -i ~/.ssh/wpengine git@git.wpengine.com info
```

### WPEngine authentication fails

**Symptom**:
```
Error: WPEngine authentication failed
```

**Solutions**:

**1. Reconfigure credentials**:
```bash
stax setup
# Enter correct credentials
```

**2. Test manually**:
```bash
# Test SSH
ssh -i ~/.ssh/wpengine git@git.wpengine.com info

# Should list your installs
```

**3. Check SSH key**:
```bash
# Verify key exists
ls -la ~/.ssh/wpengine*

# Verify public key is in WPEngine
# WPEngine portal → SSH Keys
# Should see your key
```

**4. Generate new key**:
```bash
ssh-keygen -t ed25519 -f ~/.ssh/wpengine

# Add public key to WPEngine
cat ~/.ssh/wpengine.pub
# Copy and paste in WPEngine portal

# Update Stax
stax setup
```

### SSH connections fail with Warp terminal

**Symptom**:
- SSH connections to WPEngine hang or timeout
- `stax db pull` operations fail during SSH phase
- Manual `ssh` commands to `*.ssh.wpengine.net` don't work
- Operations that work in other terminals fail in Warp

**Affected environments**:
- Warp terminal (macOS/Linux)
- All SSH operations to WPEngine sidecars (`*.ssh.wpengine.net`)
- Commands: `stax db pull`, `stax files pull`, manual SSH

**Root cause**:
Compatibility issue between Warp's SSH handling and WPEngine's SSH sidecar infrastructure.

**Solution: Use alternative terminal**

Switch to a different terminal for WPEngine SSH operations:

**macOS**:
- Terminal.app (built-in)
- iTerm2
- Alacritty
- VS Code integrated terminal

**Linux**:
- GNOME Terminal
- Konsole
- Alacritty
- Tmux

**Verification**:
```bash
# In alternative terminal (not Warp)
stax db pull

# Should work normally
```

**Why this happens**:
- Warp implements custom SSH handling that conflicts with WPEngine sidecars
- The issue is specific to WPEngine SSH infrastructure
- Other SSH connections (GitHub, AWS, etc.) work fine in Warp
- This affects ~10-20% of users who use Warp

**Tracking**:
- Issue documented in GitHub #4
- Warp team has been notified
- Workaround is stable and reliable

### Can't pull database from WPEngine

**Symptom**:
```
Error: Failed to pull database from WPEngine
```

**Solutions**:

**1. Check environment exists**:
```bash
# List environments
stax wp env list

# Try specific environment
stax db pull --environment=production
stax db pull --environment=staging
```

**2. Network issues**:
```bash
# Test connection
ping ssh.wpengineapi.net

# Try again
stax db pull
```

**3. Database too large**:
```bash
# Skip non-essential data
stax db pull \
  --skip-logs \
  --skip-transients \
  --skip-spam \
  --exclude-tables=wp_actionscheduler_logs
```

### WPEngine file sync fails

**Symptom**:
```
Error: File sync failed
```

**Solutions**:

**1. Check SSH connection**:
```bash
ssh -i ~/.ssh/wpengine git@git.wpengine.com info
```

**2. Try manual sync**:
```bash
# Sync specific directory
stax ssh
rsync -avz -e "ssh -i ~/.ssh/wpengine" \
  git@git.wpengine.com:/sites/myinstall/wp-content/uploads/ \
  wp-content/uploads/
exit
```

**3. Sync smaller directory**:
```bash
# Instead of all uploads:
stax wpe sync wp-content/uploads/2024/
```

---

## Common Error Messages

### "Port 80 is already in use"

**Solution**:
```bash
# Find what's using it
sudo lsof -i :80

# Usually Apache
sudo apachectl stop

# Or another DDEV project
ddev poweroff

# Restart
stax start
```

### "No space left on device"

**Solution**:
```bash
# Clean up Docker
docker system prune -a --volumes

# Clean up snapshots
stax db list
stax db delete-snapshot old-snapshot

# Check disk space
df -h
```

### "Error establishing a database connection"

**Solution**:
```bash
# Restart database
stax restart

# Check database is running
stax status

# Check wp-config.php
stax ssh
cat wp-config.php | grep DB_
# Should show:
# DB_HOST = 'db'
# DB_NAME = 'db'
exit
```

### "413 Request Entity Too Large"

**Cause**: Trying to upload large file.

**Solution**:
```bash
# Increase upload limit
stax ssh
echo "upload_max_filesize = 100M" >> /etc/php/php.ini
echo "post_max_size = 100M" >> /etc/php/php.ini
exit

stax restart
```

### "Maximum execution time exceeded"

**Solution**:
```bash
# Increase max execution time
stax ssh
echo "max_execution_time = 300" >> /etc/php/php.ini
exit

stax restart
```

### "Memory limit exhausted"

**Solution**:
```bash
# Increase PHP memory limit
stax ssh
echo "memory_limit = 512M" >> /etc/php/php.ini
exit

stax restart
```

---

## Still Need Help?

If your issue isn't covered here:

1. **Check docs**:
   - [User Guide](./USER_GUIDE.md)
   - [Multisite Guide](./MULTISITE.md)
   - [WPEngine Guide](./WPENGINE.md)
   - [FAQ](./FAQ.md)

2. **Run diagnostics**:
   ```bash
   stax doctor
   ```

3. **Check logs**:
   ```bash
   stax logs -f
   cat ~/.stax/logs/stax.log
   ```

4. **Search GitHub issues**:
   - [github.com/firecrown-media/stax/issues](https://github.com/firecrown-media/stax/issues)

5. **Create a new issue**:
   Include:
   - Stax version (`stax --version`)
   - DDEV version (`ddev version`)
   - macOS version
   - What you were trying to do
   - Full error message
   - Output of `stax doctor`
   - Relevant logs

6. **Contact the team**:
   - Internal: Slack #dev-tools channel
   - Email: dev@firecrown.com

---

**Most issues can be solved by**: Restarting (`stax restart`) or running diagnostics (`stax doctor`).
