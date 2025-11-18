# Stax Installation Guide

This guide walks you through installing Stax and all its prerequisites. We'll go step-by-step, explaining what each tool does and why you need it.

---

## Table of Contents

- [Quick Install](#quick-install)
- [System Requirements](#system-requirements)
- [Prerequisites](#prerequisites)
  - [1. Docker Desktop](#1-docker-desktop)
  - [2. Homebrew](#2-homebrew)
  - [3. DDEV](#3-ddev)
  - [4. WP-CLI (Optional)](#4-wp-cli-optional)
- [Installing Stax](#installing-stax)
  - [Option 1: Homebrew (Recommended)](#option-1-homebrew-recommended)
  - [Option 2: Build from Source](#option-2-build-from-source)
- [Post-Installation Setup](#post-installation-setup)
- [Verifying Your Installation](#verifying-your-installation)
- [Updating Stax](#updating-stax)
- [Uninstalling Stax](#uninstalling-stax)
- [Troubleshooting](#troubleshooting)

---

## Quick Install

If you already have Docker Desktop and Homebrew installed, you can install everything with these commands:

```bash
# Install DDEV
brew install ddev/ddev/ddev

# Install Stax via Homebrew
brew tap firecrown-media/tap
brew install stax

# Configure credentials
stax setup

# Verify installation
stax doctor
```

If you're missing prerequisites or want detailed explanations, continue reading.

---

## System Requirements

### Operating System
- **macOS 12.0 (Monterey)** or later
- **Processor**: Intel or Apple Silicon (M1/M2/M3)
- Currently **macOS only** - Windows and Linux support may come in future versions

### Hardware
- **RAM**: 8GB minimum, 16GB recommended
  - WordPress + DDEV containers need memory to run smoothly
  - More RAM = better performance, especially with multiple projects
- **Disk Space**: 10GB free space minimum
  - Docker images: ~3-5GB
  - Project files and databases: ~2-5GB per project
  - Additional space for snapshots and backups
- **Network**: Stable internet connection
  - Required for initial setup and database syncing
  - Remote media proxying works better with faster connections

---

## Prerequisites

Stax builds on top of several tools that handle different parts of the development workflow. Here's what you need and why:

### 1. Docker Desktop

**What it is**: Docker Desktop provides the container runtime that DDEV uses to run WordPress, MySQL, and other services.

**Why you need it**: Containers give you isolated, reproducible environments. Docker Desktop manages all the complexity of running Linux containers on macOS.

**Installation**:

1. Download Docker Desktop from [https://www.docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop)

2. Choose the right version:
   - **Apple Silicon (M1/M2/M3)**: Download "Apple Silicon" version
   - **Intel Mac**: Download "Intel Chip" version

3. Install Docker Desktop:
   - Open the downloaded `.dmg` file
   - Drag Docker to your Applications folder
   - Open Docker Desktop from Applications
   - Follow the setup wizard

4. Configure Docker Desktop:
   - Open Docker Desktop preferences
   - **Resources > Advanced**:
     - CPUs: 4 (recommended)
     - Memory: 4GB minimum, 8GB recommended
     - Swap: 2GB
     - Disk: 64GB+
   - Enable "Use the new Virtualization framework" (Apple Silicon only)
   - Enable "VirtioFS" for better file sharing performance

5. Verify Docker is running:
   ```bash
   docker --version
   # Should show: Docker version 24.x.x or later

   docker ps
   # Should show an empty table (no running containers yet)
   ```

**Troubleshooting Docker**:
- If Docker won't start, try: Restart your Mac
- If you see "Docker daemon not running": Make sure Docker Desktop is running in your menu bar
- If installation fails: Check that you have admin rights on your Mac

### Alternative Docker Runtimes

While Stax documentation focuses on Docker Desktop, DDEV (which Stax uses internally) supports alternative Docker runtimes on macOS:

#### Docker Desktop (Recommended)

**Pros:**
- Most tested and supported by Stax team
- Best performance on Apple Silicon
- Official Docker support
- Integrated GUI for container management

**Download:** https://www.docker.com/products/docker-desktop

#### Colima (Open-Source Alternative)

**Pros:**
- Free and open-source
- Lower resource usage than Docker Desktop
- Fast startup times
- No Docker Desktop license required

**Cons:**
- Not officially tested by Stax team
- Command-line only (no GUI)
- May require additional configuration

**Installation:**
```bash
# Install Colima
brew install colima

# Start with recommended resources
colima start --cpu 4 --memory 8 --disk 100

# Verify Docker is available
docker --version
docker ps
```

**Important:** Ensure Docker socket is accessible:
```bash
# Check Docker socket
ls -la ~/.colima/default/docker.sock

# Should be linked to /var/run/docker.sock
```

#### Rancher Desktop (Alternative)

**Pros:**
- Full Docker Desktop replacement
- Includes Kubernetes
- Open-source
- Cross-platform

**Download:** https://rancherdesktop.io/

**Note:** Not officially tested by Stax team.

#### Important Notes for Alternative Runtimes

If using Colima, Rancher Desktop, or another Docker runtime:

- ✅ Ensure `docker` CLI command works: `docker ps`
- ✅ Ensure Docker socket is accessible
- ✅ Allocate sufficient resources (4 CPU, 8GB RAM minimum)
- ⚠️ Some Stax features may behave differently
- ⚠️ You're responsible for runtime-specific configuration
- ⚠️ Check [DDEV documentation](https://ddev.readthedocs.io/) for runtime-specific setup

**If you encounter issues with alternative runtimes, we recommend trying Docker Desktop for best compatibility.**

### 2. Homebrew

**What it is**: Homebrew is macOS's package manager - think of it as an app store for command-line tools.

**Why you need it**: Homebrew makes it easy to install and update DDEV, Stax, and other developer tools.

**Installation**:

1. Open Terminal (Applications > Utilities > Terminal)

2. Install Homebrew:
   ```bash
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   ```

3. Follow the prompts:
   - Press RETURN to continue
   - Enter your macOS password when asked
   - Wait for installation to complete (5-10 minutes)

4. Add Homebrew to your PATH (the installer will show you these commands):
   ```bash
   # For Apple Silicon Macs:
   echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
   eval "$(/opt/homebrew/bin/brew shellenv)"

   # For Intel Macs:
   echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile
   eval "$(/usr/local/bin/brew shellenv)"
   ```

5. Verify Homebrew installation:
   ```bash
   brew --version
   # Should show: Homebrew 4.x.x or later

   brew doctor
   # Should say: "Your system is ready to brew"
   ```

**Already have Homebrew?**
Update it first:
```bash
brew update
```

### 3. DDEV

**What it is**: DDEV is a Docker-based local development environment specifically designed for PHP applications like WordPress.

**Why you need it**: DDEV handles all the complexity of running WordPress locally:
- Manages PHP, MySQL, nginx/Apache containers
- Generates SSL certificates automatically
- Handles DNS resolution for subdomains
- Provides WP-CLI, Composer, and other tools
- Makes multisite setup easy

**Installation**:

1. Add the DDEV tap to Homebrew:
   ```bash
   brew tap ddev/ddev
   ```

2. Install DDEV:
   ```bash
   brew install ddev
   ```
   This takes 2-5 minutes.

3. Run DDEV's configuration:
   ```bash
   mkcert -install
   ```
   This installs the local certificate authority for SSL certificates. You may need to enter your password.

4. Verify DDEV installation:
   ```bash
   ddev version
   # Should show: ddev version v1.22.x or later

   ddev --help
   # Shows all DDEV commands
   ```

**DDEV Configuration** (Optional):
```bash
# Set global defaults
ddev config global --performance-mode mutagen
ddev config global --router-bind-all-interfaces
```

**Troubleshooting DDEV**:
- If `ddev version` fails: Make sure Homebrew is in your PATH
- If mkcert fails: Try `brew install mkcert` first
- For Apple Silicon: Make sure you're using Docker Desktop (not Docker Toolbox)

### 4. WP-CLI (Optional)

**What it is**: WP-CLI is the command-line interface for WordPress. It lets you manage WordPress without using the admin dashboard.

**Why you need it**: While DDEV includes WP-CLI in containers, having it installed globally can be useful for some operations.

**Installation**:

```bash
brew install wp-cli
```

**Verify**:
```bash
wp --version
# Should show: WP-CLI 2.x.x or later
```

**Note**: This is optional - Stax uses the WP-CLI inside DDEV containers, so you don't strictly need a global installation.

---

## Installing Stax

Now that prerequisites are installed, let's install Stax itself.

### Option 1: Homebrew (Recommended)

This is the easiest method and enables automatic updates.

```bash
# Add the Firecrown tap
brew tap firecrown-media/tap

# Install Stax
brew install stax

# Verify installation
stax --version
```

**Updating via Homebrew**:
```bash
brew update
brew upgrade stax
```

**For detailed Homebrew installation instructions, see [HOMEBREW_INSTALLATION.md](HOMEBREW_INSTALLATION.md).**

### Option 2: Direct Download

Download pre-built binaries from the [GitHub Releases](https://github.com/firecrown-media/stax/releases) page.

1. Go to [Releases](https://github.com/firecrown-media/stax/releases/latest)
2. Download the appropriate archive for your system:
   - macOS Intel: `stax_VERSION_Darwin_x86_64.tar.gz`
   - macOS Apple Silicon: `stax_VERSION_Darwin_arm64.tar.gz`
   - Linux: `stax_VERSION_Linux_x86_64.tar.gz`

3. Extract and install:
   ```bash
   # Extract the archive
   tar -xzf stax_*_Darwin_*.tar.gz

   # Move binary to your PATH
   sudo mv stax /usr/local/bin/

   # Verify installation
   stax --version
   ```

4. Verify the download (optional but recommended):
   ```bash
   # Download checksums file
   wget https://github.com/firecrown-media/stax/releases/download/vX.Y.Z/checksums.txt

   # Verify
   shasum -a 256 -c checksums.txt
   ```

### Option 3: Build from Source

Use this method if you want the latest development version or need to modify Stax.

**Prerequisites for building**:
- Go 1.22 or later
- Git
- Make

**Install Go** (if not already installed):
```bash
brew install go
```

**Build and install Stax**:

1. Clone the repository:
   ```bash
   cd ~/Development  # or wherever you keep code
   git clone https://github.com/firecrown-media/stax.git
   cd stax
   ```

2. Build the binary:
   ```bash
   make build
   ```
   This creates the `stax` binary in the current directory.

3. Install globally:
   ```bash
   make install
   ```
   This copies `stax` to `/usr/local/bin`.

4. Verify installation:
   ```bash
   stax --version
   which stax  # Should show: /usr/local/bin/stax
   ```

**Updating from source**:
```bash
cd ~/Development/stax
git pull origin main
make build
make install
```

---

## Post-Installation Setup

After installing Stax, you can access comprehensive documentation:

```bash
# View the manual
man stax

# Get command help
stax --help
stax init --help
```

The man page includes:
- Complete command reference
- Usage examples
- Configuration files
- Environment variables
- Troubleshooting tips

Now, let's configure your credentials.

### 1. Configure WPEngine Credentials

Stax needs your WPEngine API credentials to pull databases and sync files.

**Get your WPEngine API credentials**:

1. Log in to [WPEngine User Portal](https://my.wpengine.com/)
2. Go to **Account** > **API Access**
3. Create a new API user or use existing credentials
4. Save your username and password

**Get your WPEngine SSH key**:

1. Generate an SSH key (if you don't have one):
   ```bash
   ssh-keygen -t ed25519 -C "your_email@example.com" -f ~/.ssh/wpengine
   ```

2. Add the public key to WPEngine:
   - Copy your public key:
     ```bash
     cat ~/.ssh/wpengine.pub
     ```
   - In WPEngine portal: **Account** > **SSH Keys**
   - Click "Add SSH Key"
   - Paste your public key

3. Test SSH connection:
   ```bash
   ssh -i ~/.ssh/wpengine git@git.wpengine.com info
   # Should show your WPEngine installs
   ```

### 2. Configure GitHub Token (Optional)

If you're working with private repositories, you'll need a GitHub personal access token.

**Create a GitHub token**:

1. Go to [GitHub Settings > Developer Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Give it a name: "Stax CLI"
4. Set expiration: 90 days (or your preference)
5. Select scopes:
   - `repo` (Full control of private repositories)
   - `read:org` (Read organization data)
6. Click "Generate token"
7. **Copy the token immediately** (you won't see it again)

### 3. Configure Credentials

Stax supports multiple ways to store credentials, depending on how you installed it.

#### Credential Storage Methods

**Homebrew Installations** (CGO_ENABLED=0):
- Homebrew builds cannot use macOS Keychain
- Use Environment Variables (recommended for CI/CD)
- Use Config File (recommended for development)

**Source Builds with CGO** (CGO_ENABLED=1):
- Can use macOS Keychain for maximum security
- Run `stax setup` to store credentials interactively

#### Option 1: Environment Variables (Recommended for CI/CD)

Add these to your shell profile (`~/.zshrc` or `~/.bashrc`):

```bash
export WPENGINE_API_USER="your-api-username"
export WPENGINE_API_PASSWORD="your-api-password"
export WPENGINE_SSH_GATEWAY="ssh.wpengine.net"
export GITHUB_TOKEN="ghp_your_token_here"
```

Then reload your shell:
```bash
source ~/.zshrc
```

**Pros**:
- Works everywhere (CI/CD, Docker, scripts)
- Easy to update
- No files to manage

**Cons**:
- Visible in shell history and process lists
- Less secure than Keychain

#### Option 2: Config File (Recommended for Development)

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

**Pros**:
- Easy to manage
- Works with Homebrew installations
- Single file with all credentials

**Cons**:
- Plain text on disk
- Must be careful not to commit to git

**Important**: Add to your `.gitignore`:
```bash
echo "~/.stax/credentials.yml" >> ~/.gitignore
```

#### Option 3: macOS Keychain (Source Builds Only)

If you built Stax from source with CGO enabled, you can use the Keychain:

```bash
stax setup
```

You'll be prompted for:
- **WPEngine API Username**: Your WPEngine API username
- **WPEngine API Password**: Your WPEngine API password
- **GitHub Token**: Your personal access token (or press Enter to skip)
- **SSH Key Path**: Path to your WPEngine SSH key (usually `~/.ssh/wpengine`)

**Example session**:
```
? WPEngine API Username: myuser@firecrown.com
? WPEngine API Password: ********
? GitHub Personal Access Token (optional): ghp_xxxxxxxxxxxxx
? SSH Key for WPEngine: ~/.ssh/wpengine

✓ Validating WPEngine credentials
✓ Validating GitHub token
✓ Saving credentials to macOS Keychain

Credentials saved successfully!
```

All credentials are stored securely in macOS Keychain. They're never saved in plain text.

**Pros**:
- Maximum security
- Encrypted by macOS
- Integration with system security

**Cons**:
- Only works with CGO-enabled builds
- Not available in Homebrew installations
- Harder to use in CI/CD

#### Checking Your Setup

Run `stax setup` to see which storage method is available:

**Homebrew installations** will show:
```
⚠ macOS Keychain storage is not available in this build
  This is normal for Homebrew installations (built with CGO_ENABLED=0)

[Instructions for environment variables or config file]
```

**CGO-enabled builds** will prompt for credentials interactively.

#### Updating Credentials

**Environment Variables**: Edit your shell profile and reload
**Config File**: Edit `~/.stax/credentials.yml`
**Keychain**: Run `stax setup` again

---

## Verifying Your Installation

Let's make sure everything is working correctly.

### Run Stax Doctor

Stax includes a diagnostic tool that checks your setup:

```bash
stax doctor
```

**Expected output**:
```
Running diagnostics...

✓ Stax installed (v2.12.5)
✓ DDEV installed (v1.22.7)
✓ Docker Desktop running
✓ WPEngine credentials valid
✓ GitHub token valid (optional)
✓ SSH key configured for WPEngine
✓ Ports 80, 443, 8025 available
✓ mkcert installed for SSL certificates

All checks passed! You're ready to use Stax.
```

### What if checks fail?

**"DDEV not found"**:
```bash
# Reinstall DDEV
brew reinstall ddev
```

**"Docker not running"**:
- Open Docker Desktop from Applications
- Wait for it to start (green light in menu bar)

**"WPEngine credentials invalid"**:
```bash
# Reconfigure credentials
stax setup
```

**"Port 80 or 443 in use"**:
```bash
# Stop Apache (if running)
sudo apachectl stop

# Or check what's using the port
sudo lsof -i :80
sudo lsof -i :443
```

---

## Updating Stax

### Via Homebrew (Recommended)

```bash
brew update
brew upgrade stax
```

### Via Direct Download

1. Download the latest version from [Releases](https://github.com/firecrown-media/stax/releases/latest)
2. Follow the same installation steps as above

### From Source

```bash
cd ~/Development/stax
git pull origin main
make build
make install
```

### Check Your Version

```bash
stax --version
```

To check for available updates:
```bash
# If installed via Homebrew
brew outdated stax

# Check latest release on GitHub
open https://github.com/firecrown-media/stax/releases/latest
```

---

## Uninstalling Stax

### Remove Stax via Homebrew

```bash
brew uninstall stax
brew untap firecrown-media/tap
```

### Remove Stax installed manually

```bash
sudo rm /usr/local/bin/stax
```

### Clean up Stax data (optional)

```bash
# Remove configuration and snapshots
rm -rf ~/.stax

# Remove credentials from Keychain
stax setup --remove  # Before uninstalling Stax
```

### Uninstall DDEV and Docker (optional)

If you no longer need DDEV:
```bash
# Stop all DDEV projects
ddev poweroff

# Uninstall DDEV
brew uninstall ddev

# Remove DDEV data
rm -rf ~/.ddev

# Uninstall Docker Desktop
# Drag Docker from Applications to Trash
# Remove data:
rm -rf ~/Library/Containers/com.docker.docker
```

---

## Troubleshooting

### Common Installation Issues

#### "Command not found: brew"

**Problem**: Homebrew isn't in your PATH.

**Solution**:
```bash
# For Apple Silicon:
eval "$(/opt/homebrew/bin/brew shellenv)"

# For Intel:
eval "$(/usr/local/bin/brew shellenv)"

# Make it permanent:
# For Apple Silicon:
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile

# For Intel:
echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile
```

#### "Permission denied" when installing

**Problem**: You don't have admin rights.

**Solution**: Contact your system administrator or use `sudo`:
```bash
sudo make install
```

#### Docker Desktop won't start

**Problem**: Docker fails to start after installation.

**Solutions**:
1. Restart your Mac
2. Check disk space (need at least 10GB free)
3. Reset Docker Desktop: Settings > Troubleshoot > Reset to factory defaults
4. Reinstall Docker Desktop

#### DDEV installation fails

**Problem**: Homebrew can't install DDEV.

**Solution**:
```bash
# Update Homebrew
brew update

# Try again
brew tap ddev/ddev
brew install ddev

# If still fails, try direct download:
# https://github.com/ddev/ddev/releases
```

#### "mkcert not installed"

**Problem**: SSL certificates won't work without mkcert.

**Solution**:
```bash
brew install mkcert
brew install nss  # For Firefox support
mkcert -install
```

#### Can't connect to WPEngine

**Problem**: `stax setup` can't validate WPEngine credentials.

**Possible causes**:
1. Incorrect username or password
   - Double-check in WPEngine portal
   - Try logging in to WPEngine web interface first
2. SSH key not added to WPEngine
   - Make sure you added the public key (`.pub` file)
3. Network/firewall blocking connection
   - Try from a different network
   - Check with your IT department

**Test connection manually**:
```bash
# Test API
curl -u "username:password" https://api.wpengineapi.com/v1/installs

# Test SSH
ssh -i ~/.ssh/wpengine git@git.wpengine.com info
```

#### GitHub token doesn't work

**Problem**: GitHub authentication fails.

**Solution**:
1. Make sure you selected the `repo` scope when creating the token
2. Check token hasn't expired
3. Test it manually:
   ```bash
   curl -H "Authorization: token YOUR_TOKEN" https://api.github.com/user
   ```

### Getting Help

If you're still having trouble:

1. **Check the troubleshooting guide**: [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
2. **Run diagnostics**: `stax doctor`
3. **Check logs**: `~/.stax/logs/stax.log`
4. **Search GitHub issues**: [github.com/firecrown-media/stax/issues](https://github.com/firecrown-media/stax/issues)
5. **Ask for help**: Contact the Firecrown development team

---

## Next Steps

Now that Stax is installed, you're ready to start using it!

1. **Quick Start**: [docs/QUICK_START.md](./QUICK_START.md) - Set up your first project in 5 minutes
2. **User Guide**: [docs/USER_GUIDE.md](./USER_GUIDE.md) - Learn all the features
3. **Examples**: [docs/EXAMPLES.md](./EXAMPLES.md) - See real-world workflows

---

**Installation complete!** You're ready to start developing with Stax.
