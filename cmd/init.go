package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/git"
	"github.com/firecrown-media/stax/pkg/prompts"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	initName             string
	initType             string
	initMode             string
	initPHPVersion       string
	initMySQLVersion     string
	initRepo             string
	initBranch           string
	initWPEngineInstall  string
	initWPEngineEnv      string
	initInteractive      bool
	initSkipDB           bool
	initSkipFiles        bool
	initFromDDEV         bool
	initTemplate         bool
	initShowExample      bool
	initStart            bool
	initPullDB           bool
	initPullFiles        bool
	initSkipWordPress    bool
	initWordPressVersion string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Stax project",
	Long: `Initialize a new Stax project in the current directory.

This command can either:
  - Set up a new project from scratch (interactive or non-interactive)
  - Import an existing DDEV project (--from-ddev)
  - Generate configuration templates (--template, --show-example)

For new projects, this will:
  - Create a .stax.yml configuration file
  - Optionally configure WPEngine integration
  - Clone the GitHub repository (if specified)
  - Generate DDEV configuration
  - Start DDEV containers (optional)
  - Download WordPress core files (optional)
  - Generate wp-config.php (optional)
  - Pull database and files from WPEngine (optional)

By default, this command runs in interactive mode, prompting for all
required information. You can skip prompts by providing all flags.`,
	Example: `  # Interactive mode (default)
  stax init

  # Import existing DDEV project
  stax init --from-ddev

  # Non-interactive with all flags
  stax init \
    --name=myproject \
    --type=wordpress-multisite \
    --mode=subdomain \
    --php=8.1 \
    --mysql=8.0 \
    --repo=https://github.com/org/repo.git \
    --branch=main \
    --install=myinstall \
    --environment=staging \
    --start \
    --pull-db

  # Initialize with specific WordPress version
  stax init --start --wordpress-version=6.4.2

  # Skip WordPress core download
  stax init --start --skip-wordpress

  # Generate template configuration
  stax init --template

  # Show example configuration
  stax init --show-example`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Project configuration flags
	initCmd.Flags().StringVar(&initName, "name", "", "project name (default: current directory name)")
	initCmd.Flags().StringVar(&initType, "type", "wordpress", "project type (wordpress, wordpress-multisite)")
	initCmd.Flags().StringVar(&initMode, "mode", "subdomain", "multisite mode (subdomain, subdirectory)")
	initCmd.Flags().StringVar(&initPHPVersion, "php", "8.1", "PHP version")
	initCmd.Flags().StringVar(&initMySQLVersion, "mysql", "8.0", "MySQL version")

	// Repository flags
	initCmd.Flags().StringVar(&initRepo, "repo", "", "GitHub repository URL")
	initCmd.Flags().StringVar(&initBranch, "branch", "main", "repository branch")

	// WPEngine flags
	initCmd.Flags().StringVar(&initWPEngineInstall, "install", "", "WPEngine install name")
	initCmd.Flags().StringVar(&initWPEngineEnv, "environment", "production", "WPEngine environment (production, staging, development)")

	// Behavior flags
	// Default to auto-detecting TTY - will be set to true if stdin is a terminal
	// Users can explicitly override with --interactive=true or --interactive=false
	initCmd.Flags().BoolVar(&initInteractive, "interactive", prompts.IsInteractive(), "enable interactive prompts (default: auto-detect TTY)")
	initCmd.Flags().BoolVar(&initStart, "start", false, "start DDEV after initialization")
	initCmd.Flags().BoolVar(&initPullDB, "pull-db", false, "pull database after initialization")
	initCmd.Flags().BoolVar(&initPullFiles, "pull-files", false, "pull files after initialization")
	initCmd.Flags().BoolVar(&initSkipDB, "skip-db", false, "skip database operations")
	initCmd.Flags().BoolVar(&initSkipFiles, "skip-files", false, "skip file operations")
	initCmd.Flags().BoolVar(&initSkipWordPress, "skip-wordpress", false, "skip WordPress core download")
	initCmd.Flags().StringVar(&initWordPressVersion, "wordpress-version", "", "WordPress version to download (default: latest)")

	// Special modes
	initCmd.Flags().BoolVar(&initFromDDEV, "from-ddev", false, "import existing DDEV project")
	initCmd.Flags().BoolVar(&initTemplate, "template", false, "generate .stax.yml template to stdout")
	initCmd.Flags().BoolVar(&initShowExample, "show-example", false, "show example configuration with comments")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Handle special modes first
	if initTemplate {
		return generateTemplate()
	}

	if initShowExample {
		return showExample()
	}

	ui.PrintHeader("Initializing Stax Project")

	projectDir := getProjectDir()

	// Check if importing from existing DDEV
	if initFromDDEV {
		return runInitFromDDEV(projectDir)
	}

	// Run full initialization
	return runFullInit(projectDir)
}

func runFullInit(projectDir string) error {
	// Step 1: Check prerequisites
	if err := checkPrerequisites(); err != nil {
		return err
	}

	// Step 2: Gather project configuration
	cfg, err := gatherProjectConfiguration(projectDir)
	if err != nil {
		return err
	}

	// Step 3: Check for existing configuration
	if err := checkExistingConfiguration(projectDir); err != nil {
		return err
	}

	// Step 4: Clone repository if specified
	if cfg.Repository.URL != "" {
		if err := cloneRepository(projectDir, cfg); err != nil {
			return err
		}
	}

	// Step 5: Generate DDEV configuration
	if err := generateDDEVConfig(projectDir, cfg); err != nil {
		return err
	}

	// Step 6: Generate multisite nginx config if needed
	if isMultisite(cfg.Project.Type) {
		if err := generateMultisiteNginxConfig(projectDir, cfg); err != nil {
			return err
		}
	}

	// Step 7: Save Stax configuration
	configPath := filepath.Join(projectDir, ".stax.yml")
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	ui.Success("Created .stax.yml")

	// Step 7a: Ensure .gitignore exists with WordPress and Stax entries
	if err := git.EnsureGitignore(projectDir); err != nil {
		ui.Warning(fmt.Sprintf("Failed to manage .gitignore: %v", err))
		// Don't fail init if .gitignore management fails
	}

	// Step 8: Start DDEV if requested
	shouldStart := initStart
	if initInteractive && !initStart {
		var err error
		shouldStart, err = prompts.SafePromptConfirm("Start DDEV now?", true)
		if err != nil {
			return err
		}
	}

	if shouldStart {
		ui.Section("Starting DDEV")
		spinner := ui.NewSpinner("Starting DDEV containers...")
		spinner.Start()

		mgr := ddev.NewManager(projectDir)
		if err := mgr.Start(); err != nil {
			spinner.Error("Failed to start DDEV")
			return err
		}
		spinner.Success("DDEV started successfully")

		// Wait for DDEV to be ready
		ui.Info("Waiting for services to be ready...")
		waitSpinner := ui.NewSpinner("Checking service status...")
		waitSpinner.Start()
		if err := mgr.WaitForReady(2 * time.Minute); err != nil {
			waitSpinner.Stop()
			ui.Warning(fmt.Sprintf("Services may not be fully ready yet: %v", err))
			ui.Info("Continuing with setup...")
		} else {
			waitSpinner.Success("Services are ready")
		}

		// Step 8a: Download WordPress core if needed
		if !initSkipWordPress {
			if hasWordPressCore(projectDir) {
				ui.Section("WordPress Core")
				ui.Info("WordPress core files already exist, skipping download")
			} else {
				ui.Section("Setting Up WordPress")

				// Determine version to download
				version := "latest"
				if initWordPressVersion != "" {
					version = initWordPressVersion
					cfg.WordPress.Version = version
				} else if cfg.WordPress.Version != "" {
					version = cfg.WordPress.Version
				}

				if err := downloadWordPressCore(projectDir, version); err != nil {
					ui.Warning(fmt.Sprintf("Failed to download WordPress core: %v", err))
					ui.Info("You can download manually: ddev wp core download")
				}
			}
		} else {
			ui.Info("Skipping WordPress core download (--skip-wordpress flag set)")
		}

		// Step 8b: Generate wp-config.php if needed
		if !initSkipWordPress && !hasWordPressConfig(projectDir) {
			if err := generateWordPressConfig(projectDir, cfg); err != nil {
				ui.Warning(fmt.Sprintf("Failed to generate wp-config.php: %v", err))
				ui.Info("You can create manually: ddev wp config create --dbname=db --dbuser=db --dbpass=db --dbhost=db")
			}
		}
	}

	// Step 9: Pull database if requested
	if shouldPullDatabase(cfg) {
		if err := pullDatabase(projectDir, cfg); err != nil {
			ui.Warning("Database pull failed: %v", err)
		}
	}

	// Step 10: Pull files if requested
	if shouldPullFiles(cfg) {
		if err := pullFiles(projectDir, cfg); err != nil {
			ui.Warning("File pull failed: %v", err)
		}
	}

	// Print success summary
	printSuccessSummary(projectDir, cfg)

	return nil
}

func checkPrerequisites() error {
	ui.Section("Checking prerequisites")

	// Check if DDEV is installed
	if !ddev.IsInstalled() {
		return errors.NewWithSolution(
			"DDEV is not installed",
			"Stax requires DDEV to be installed",
			errors.Solution{
				Description: "Install DDEV",
				Steps: []string{
					"Visit https://ddev.readthedocs.io/en/stable/users/install/",
					"Follow the installation instructions for your platform",
					"Run 'ddev version' to verify installation",
				},
			},
		)
	}
	ui.Success("DDEV is installed")

	// Check if Git is available (if repository will be cloned)
	if initRepo != "" || (initInteractive && !initFromDDEV) {
		if !git.IsGitAvailable() {
			return errors.NewWithSolution(
				"Git is not installed",
				"Git is required for repository cloning",
				errors.Solution{
					Description: "Install Git",
					Steps: []string{
						"Visit https://git-scm.com/downloads",
						"Follow the installation instructions for your platform",
						"Run 'git --version' to verify installation",
					},
				},
			)
		}
		ui.Success("Git is installed")
	}

	return nil
}

func cloneRepository(projectDir string, cfg *config.Config) error {
	ui.Section("Cloning Repository")

	spinner := ui.NewSpinner("Cloning repository...")
	spinner.Start()

	opts := git.CloneOptions{
		URL:         cfg.Repository.URL,
		Destination: projectDir,
		Branch:      cfg.Repository.Branch,
		Depth:       cfg.Repository.Depth,
		Quiet:       !verbose,
	}

	if err := git.Clone(opts); err != nil {
		spinner.Error("Failed to clone repository")
		return err
	}

	spinner.Success("Repository cloned successfully")
	return nil
}

func runInitFromDDEV(projectDir string) error {
	ui.Info("Importing existing DDEV project...")

	// Check if DDEV config exists
	if !ddev.IsConfigured(projectDir) {
		return errors.NewWithSolution(
			"No DDEV configuration found",
			"Cannot import from DDEV - no .ddev/config.yaml exists",
			errors.Solution{
				Description: "Initialize DDEV first",
				Steps: []string{
					"Run 'ddev config --project-type=wordpress' to set up DDEV",
					"Then run 'stax init --from-ddev' again",
				},
			},
		)
	}

	// Check if .stax.yml already exists
	configPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(configPath); err == nil {
		ui.Warning(".stax.yml already exists")
		fmt.Print("Overwrite? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			ui.Info("Import cancelled")
			return nil
		}
	}

	// Read DDEV config to get basic info
	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read DDEV config: %w", err)
	}

	ui.Success("Found DDEV configuration")

	// Create Stax config from DDEV config
	cfg := config.Defaults()
	cfg.Project.Name = ddevConfig.Name
	cfg.Project.Type = mapDDEVTypeToStax(ddevConfig.Type)
	cfg.DDEV.PHPVersion = ddevConfig.PHPVersion
	cfg.DDEV.MySQLVersion = ddevConfig.Database.Version
	cfg.DDEV.MySQLType = ddevConfig.Database.Type

	// Prompt for optional WPEngine integration
	ui.Info("\nOptional: Configure WPEngine integration")
	fmt.Print("Add WPEngine integration? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	addWPEngine := (response == "y" || response == "yes")

	if addWPEngine {
		// Prompt for WPEngine details
		fmt.Print("WPEngine install name: ")
		install, _ := reader.ReadString('\n')
		install = strings.TrimSpace(install)

		fmt.Print("WPEngine environment (production/staging/development) [production]: ")
		env, _ := reader.ReadString('\n')
		env = strings.TrimSpace(env)
		if env == "" {
			env = "production"
		}

		cfg.ProviderConfig["install"] = install
		cfg.ProviderConfig["environment"] = env
	}

	// Save .stax.yml
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.Success("Created .stax.yml from DDEV configuration")
	ui.Info("\nYour DDEV project now has Stax features enabled!")
	ui.Info("Run 'stax status' to see your environment")

	if addWPEngine {
		ui.Info("Run 'stax db pull' to sync your database from WPEngine")
	}

	return nil
}

func printSuccessSummary(projectDir string, cfg *config.Config) {
	ui.PrintHeader("Project Initialized Successfully!")

	fmt.Println()
	ui.Success("Created:")
	ui.Info("  - .stax.yml")
	ui.Info("  - .ddev/config.yaml")
	if isMultisite(cfg.Project.Type) {
		ui.Info("  - .ddev/nginx_full/multisite.conf")
	}

	fmt.Println()
	ui.Section("Next Steps:")

	if !initStart {
		ui.ProgressMsg("stax start         - Start DDEV environment")
	}

	if wpeInstall(cfg) != "" && !initPullDB {
		ui.ProgressMsg("stax db pull       - Pull database from WPEngine")
	}

	if wpeInstall(cfg) != "" && !initPullFiles {
		ui.ProgressMsg("stax files pull    - Pull files from WPEngine")
	}

	ui.ProgressMsg("stax status        - View environment status")

	fmt.Println()
	ui.Success("Your site will be available at: https://%s.ddev.site", cfg.Project.Name)
}
