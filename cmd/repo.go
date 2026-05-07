package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/git"
	"github.com/firecrown-media/stax/pkg/prerequisites"
	"github.com/firecrown-media/stax/pkg/prompts"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/firecrown-media/stax/pkg/wpengine"
	"github.com/spf13/cobra"
)

var (
	repoInitInstall  string
	repoInitGitHub   string
	repoInitPrivate  bool
	repoInitSyncDirs []string
	repoInitBranch   string
)

// repoCmd represents the repo command group
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Git repository operations",
	Long:  `Commands for managing Git repositories for WordPress projects.`,
}

// repoInitCmd initializes a Git repository for an existing WPEngine site
var repoInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Git repository for an existing WordPress site",
	Long: `Initialize a Git repository for a WordPress site that exists on WPEngine
but doesn't yet have version control.

This command:
1. Initializes a Git repository (or uses existing)
2. Generates a WordPress-appropriate .gitignore
3. Syncs files from WPEngine (themes, plugins, mu-plugins)
4. Creates an initial commit
5. Optionally creates a GitHub repository and pushes

Example:
  # Basic initialization
  stax repo init --install mysite-prod

  # With GitHub repository creation
  stax repo init --install mysite-prod --github myorg/mysite

  # Custom directories to sync
  stax repo init --install mysite-prod --sync-dirs wp-content/themes,wp-content/plugins`,
	RunE: runRepoInit,
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoInitCmd)

	repoInitCmd.Flags().StringVar(&repoInitInstall, "install", "", "WPEngine install name (required)")
	repoInitCmd.Flags().StringVar(&repoInitGitHub, "github", "", "GitHub repository (org/repo) - creates if doesn't exist")
	repoInitCmd.Flags().BoolVar(&repoInitPrivate, "private", true, "Create private GitHub repository")
	repoInitCmd.Flags().StringSliceVar(&repoInitSyncDirs, "sync-dirs",
		[]string{"wp-content/themes", "wp-content/plugins", "wp-content/mu-plugins"},
		"Directories to sync from WPEngine")
	repoInitCmd.Flags().StringVar(&repoInitBranch, "branch", "main", "Default branch name")
}

func runRepoInit(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Initializing Git Repository")

	projectDir := getProjectDir()

	// Step 1: Check prerequisites
	if err := checkRepoPrerequisites(); err != nil {
		return err
	}

	// Step 2: Get WPEngine install name
	install := repoInitInstall
	if install == "" {
		if !prompts.IsInteractive() {
			return fmt.Errorf("--install flag is required in non-interactive mode")
		}

		var err error
		install, err = prompts.WPEngineInstallPrompt("WPEngine install name")
		if err != nil {
			return err
		}
	}

	// Step 3: Validate WPEngine credentials and access
	ui.Section("Validating WPEngine Access")
	if err := validateWPEngineAccess(install); err != nil {
		return err
	}
	ui.Success("Connected to WPEngine install: %s", install)

	// Step 4: Initialize Git repository
	ui.Section("Initializing Git Repository")
	if err := initGitRepo(projectDir); err != nil {
		return err
	}

	// Step 5: Generate .gitignore
	ui.Section("Generating .gitignore")
	if err := generateWordPressGitignore(projectDir); err != nil {
		return err
	}
	ui.Success("Created .gitignore with WordPress best practices")

	// Step 6: Sync files from WPEngine
	ui.Section("Syncing Files from WPEngine")
	if err := syncFilesFromWPEngine(projectDir, install, repoInitSyncDirs); err != nil {
		return err
	}

	// Step 7: Create initial commit
	ui.Section("Creating Initial Commit")
	if err := createInitialCommit(projectDir); err != nil {
		return err
	}
	ui.Success("Created initial commit")

	// Step 8: Create GitHub repository if requested
	if repoInitGitHub != "" {
		ui.Section("Setting Up GitHub Repository")
		if err := setupGitHubRepo(projectDir, repoInitGitHub, repoInitPrivate); err != nil {
			ui.Warning("GitHub setup failed: %v", err)
			ui.Info("You can manually create the repository and push later")
		} else {
			ui.Success("Pushed to https://github.com/%s", repoInitGitHub)
		}
	}

	// Step 9: Display next steps
	displayRepoNextSteps(projectDir, install, repoInitGitHub)

	return nil
}

func checkRepoPrerequisites() error {
	ui.Section("Checking Prerequisites")

	// Check Git
	gitDep := prerequisites.GetDependency("git")
	if gitDep == nil {
		return fmt.Errorf("internal error: git dependency not found")
	}

	result := gitDep.Check()
	if !result.OK() {
		ui.Error("Git is not installed")
		ui.Info("  Install: %s", gitDep.InstallCmd)
		ui.Info("  Docs: %s", gitDep.InstallURL)
		return fmt.Errorf("git is required")
	}
	ui.Success(result.Message)

	// Check GitHub CLI if --github is specified
	if repoInitGitHub != "" {
		ghDep := prerequisites.GetDependency("gh")
		if ghDep == nil {
			return fmt.Errorf("internal error: gh dependency not found")
		}

		result := ghDep.Check()
		if !result.OK() {
			ui.Warning("GitHub CLI (gh) is not installed")
			ui.Info("  Install: %s", ghDep.InstallCmd)
			ui.Info("  The --github option requires gh CLI")
			return fmt.Errorf("gh CLI is required for --github option")
		}
		ui.Success(result.Message)

		// Check if gh is authenticated
		authCmd := exec.Command("gh", "auth", "status")
		if err := authCmd.Run(); err != nil {
			ui.Warning("GitHub CLI is not authenticated")
			ui.Info("  Run: gh auth login")
			return fmt.Errorf("gh CLI is not authenticated - run 'gh auth login'")
		}
		ui.Success("GitHub CLI authenticated")
	}

	return nil
}

func validateWPEngineAccess(install string) error {
	// Get credentials
	creds, err := credentials.GetWPEngineCredentialsWithFallback(install)
	if err != nil {
		return fmt.Errorf("WPEngine credentials not found: %w\n  Run 'stax setup' to configure credentials", err)
	}

	// Create client and test connection
	client := wpengine.NewClient(creds.APIUser, creds.APIPassword, install)
	_, err = client.GetInstall(install)
	if err != nil {
		return fmt.Errorf("failed to access WPEngine install '%s': %w", install, err)
	}

	return nil
}

func initGitRepo(projectDir string) error {
	// Check if already a git repo
	gitDir := filepath.Join(projectDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		ui.Info("Git repository already exists")
		return nil
	}

	// Initialize git repo
	cmd := exec.Command("git", "init", "-b", repoInitBranch)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	ui.Success("Initialized git repository with branch '%s'", repoInitBranch)
	return nil
}

func generateWordPressGitignore(projectDir string) error {
	gitignorePath := filepath.Join(projectDir, ".gitignore")

	// Check if .gitignore already exists
	if _, err := os.Stat(gitignorePath); err == nil {
		ui.Info(".gitignore already exists, appending WordPress rules")
		// Read existing content
		existing, err := os.ReadFile(gitignorePath)
		if err != nil {
			return fmt.Errorf("failed to read existing .gitignore: %w", err)
		}

		// Check if it already has our marker
		if strings.Contains(string(existing), "# WordPress (Stax generated)") {
			ui.Info("WordPress rules already present in .gitignore")
			return nil
		}

		// Append our rules
		content := string(existing) + "\n" + wordpressGitignoreContent()
		if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to update .gitignore: %w", err)
		}
		return nil
	}

	// Create new .gitignore
	if err := os.WriteFile(gitignorePath, []byte(wordpressGitignoreContent()), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	return nil
}

func wordpressGitignoreContent() string {
	return `# WordPress (Stax generated)
# https://github.com/Firecrown-Media/stax

# WordPress Core (install via composer or download)
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

# wp-content directories to ignore
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

# Build artifacts
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
`
}

func syncFilesFromWPEngine(projectDir, install string, syncDirs []string) error {
	// Get SSH key
	sshKey, err := credentials.GetSSHPrivateKeyWithFallback(install)
	if err != nil {
		return fmt.Errorf("SSH key not found: %w\n  Run 'stax setup' to configure SSH key", err)
	}

	// Create SSH client
	sshConfig := wpengine.SSHConfig{
		Install:    install,
		PrivateKey: sshKey,
	}

	sshClient, err := wpengine.NewSSHClient(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer sshClient.Close()

	for _, dir := range syncDirs {
		ui.Info("Syncing %s...", dir)

		localDir := filepath.Join(projectDir, dir)

		// Create local directory if it doesn't exist
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", localDir, err)
		}

		// Sync using SyncDirectory
		syncOpts := wpengine.SyncOptions{
			Exclude: []string{
				"node_modules/",
				".git/",
				".sass-cache/",
				"*.log",
				".DS_Store",
				"mysql.sql",
			},
			Progress: true,
		}

		// Remote path is relative to /sites/{install}/
		remotePath := fmt.Sprintf("/sites/%s/%s/", install, dir)

		if err := sshClient.SyncDirectory(remotePath, localDir+"/", syncOpts); err != nil {
			ui.Warning("Failed to sync %s: %v", dir, err)
			continue
		}

		ui.Success("Synced %s", dir)
	}

	return nil
}

func createInitialCommit(projectDir string) error {
	// Add all files
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = projectDir
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Check if there are any changes to commit
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = projectDir
	output, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if len(output) == 0 {
		ui.Info("No files to commit")
		return nil
	}

	// Create commit
	commitCmd := exec.Command("git", "commit", "-m", "Initial commit from WPEngine via Stax\n\nSynced from WPEngine using stax repo init.\nFiles include themes, plugins, and mu-plugins.")
	commitCmd.Dir = projectDir
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr

	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

func setupGitHubRepo(projectDir, repoName string, private bool) error {
	// Check if remote already exists
	remoteCmd := exec.Command("git", "remote", "get-url", "origin")
	remoteCmd.Dir = projectDir
	if err := remoteCmd.Run(); err == nil {
		ui.Info("Remote 'origin' already exists")
		// Just push
		return pushToGitHub(projectDir)
	}

	// Try to create the repository
	visibility := "--private"
	if !private {
		visibility = "--public"
	}

	createCmd := exec.Command("gh", "repo", "create", repoName, visibility, "--source=.", "--push")
	createCmd.Dir = projectDir
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr

	if err := createCmd.Run(); err != nil {
		// Repo might already exist, try to add remote and push
		ui.Info("Repository may already exist, trying to add remote...")

		addRemoteCmd := exec.Command("git", "remote", "add", "origin", fmt.Sprintf("git@github.com:%s.git", repoName))
		addRemoteCmd.Dir = projectDir
		if err := addRemoteCmd.Run(); err != nil {
			return fmt.Errorf("failed to add remote: %w", err)
		}

		return pushToGitHub(projectDir)
	}

	return nil
}

func pushToGitHub(projectDir string) error {
	pushCmd := exec.Command("git", "push", "-u", "origin", repoInitBranch)
	pushCmd.Dir = projectDir
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr

	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push to GitHub: %w", err)
	}

	return nil
}

func displayRepoNextSteps(projectDir, install, github string) {
	ui.Section("Next Steps")

	fmt.Println()
	ui.Info("Your Git repository is ready! Here's what to do next:")
	fmt.Println()

	if github == "" {
		ui.Info("1. Create a GitHub repository and push:")
		ui.Info("   git remote add origin git@github.com:YOUR_ORG/YOUR_REPO.git")
		ui.Info("   git push -u origin " + repoInitBranch)
		fmt.Println()
	}

	ui.Info("2. Set up GitHub Actions for deployment:")
	ui.Info("   stax actions setup")
	fmt.Println()

	ui.Info("3. Initialize Stax for local development:")
	ui.Info("   stax init --install " + install)
	fmt.Println()

	ui.Info("4. Configure branch protection (recommended):")
	ui.Info("   See: " + git.GetBranchProtectionGuideURL())
	fmt.Println()
}
