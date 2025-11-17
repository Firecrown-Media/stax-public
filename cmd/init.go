package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/git"
	"github.com/firecrown-media/stax/pkg/prompts"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/providers/wpengine"
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
	initCmd.Flags().BoolVar(&initInteractive, "interactive", true, "enable interactive prompts")
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

	// Step 8: Start DDEV if requested
	shouldStart := initStart
	if initInteractive && !initStart {
		var err error
		shouldStart, err = prompts.PromptConfirm("Start DDEV now?", true)
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

func gatherProjectConfiguration(projectDir string) (*config.Config, error) {
	ui.Section("Project Configuration")

	cfg := config.Defaults()

	// Project name
	defaultName := filepath.Base(projectDir)
	if initName != "" {
		cfg.Project.Name = initName
	} else if initInteractive {
		name, err := prompts.PromptInput("Project name", defaultName)
		if err != nil {
			return nil, err
		}
		cfg.Project.Name = name
	} else {
		cfg.Project.Name = defaultName
	}

	// Project type
	if initType != "" {
		cfg.Project.Type = initType
		if strings.Contains(initType, "multisite") {
			if initMode != "" {
				cfg.Project.Mode = initMode
			}
		} else {
			cfg.Project.Mode = "single"
		}
	} else if initInteractive {
		projectType, err := promptProjectType()
		if err != nil {
			return nil, err
		}
		cfg.Project.Type = projectType

		if isMultisite(projectType) {
			mode, err := promptMultisiteMode()
			if err != nil {
				return nil, err
			}
			cfg.Project.Mode = mode
		} else {
			cfg.Project.Mode = "single"
		}
	}

	// DDEV configuration
	cfg.DDEV.PHPVersion = initPHPVersion
	cfg.DDEV.MySQLVersion = initMySQLVersion

	if initInteractive {
		phpVersion, err := prompts.PromptInput("PHP version", initPHPVersion)
		if err != nil {
			return nil, err
		}
		cfg.DDEV.PHPVersion = phpVersion

		mysqlVersion, err := prompts.PromptInput("MySQL version", initMySQLVersion)
		if err != nil {
			return nil, err
		}
		cfg.DDEV.MySQLVersion = mysqlVersion
	}

	// WPEngine configuration
	if err := gatherWPEngineConfiguration(cfg); err != nil {
		return nil, err
	}

	// Repository configuration
	if err := gatherRepositoryConfiguration(cfg); err != nil {
		return nil, err
	}

	// Network domain for multisite
	if isMultisite(cfg.Project.Type) {
		if initInteractive {
			defaultDomain := fmt.Sprintf("%s.ddev.site", cfg.Project.Name)
			domain, err := prompts.PromptInput("Primary domain", defaultDomain)
			if err != nil {
				return nil, err
			}
			cfg.Network.Domain = domain
		} else {
			cfg.Network.Domain = fmt.Sprintf("%s.ddev.site", cfg.Project.Name)
		}
	}

	return cfg, nil
}

func gatherWPEngineConfiguration(cfg *config.Config) error {
	ui.Section("WPEngine Integration")

	setupWPEngine := false
	if initWPEngineInstall != "" {
		setupWPEngine = true
	} else if initInteractive {
		var err error
		setupWPEngine, err = prompts.PromptConfirm("Set up WPEngine integration?", true)
		if err != nil {
			return err
		}
	}

	if !setupWPEngine {
		ui.Info("Skipping WPEngine integration")
		return nil
	}

	// Load credentials to get available installations
	var installName string
	if initWPEngineInstall != "" {
		installName = initWPEngineInstall
	} else if initInteractive {
		// Try to list available installations
		creds, err := credentials.GetWPEngineCredentials("global")
		if err == nil {
			// Show available installations
			p, err := createWPEngineProviderForListing(creds)
			if err == nil {
				sites, err := p.ListSites()
				if err == nil && len(sites) > 0 {
					ui.Info("Available WPEngine installations:")
					for i, site := range sites {
						if i < 10 { // Limit to first 10
							ui.Info("  - %s (%s)", site.Name, site.Environment)
						}
					}
					fmt.Println()
				}
			}
		}

		// Prompt for install name
		name, err := prompts.PromptInput("WPEngine install name", "")
		if err != nil {
			return err
		}
		installName = name
	}

	cfg.WPEngine.Install = installName

	// Environment
	if initWPEngineEnv != "" {
		cfg.WPEngine.Environment = initWPEngineEnv
	} else if initInteractive {
		env, err := prompts.EnvironmentPrompt(cfg.WPEngine.Environment)
		if err != nil {
			return err
		}
		cfg.WPEngine.Environment = env
	}

	return nil
}

func gatherRepositoryConfiguration(cfg *config.Config) error {
	ui.Section("Repository Configuration")

	cloneRepo := false
	if initRepo != "" {
		cloneRepo = true
	} else if initInteractive {
		var err error
		cloneRepo, err = prompts.PromptConfirm("Clone from Git repository?", false)
		if err != nil {
			return err
		}
	}

	if !cloneRepo {
		ui.Info("Skipping repository cloning")
		return nil
	}

	// Repository URL
	if initRepo != "" {
		cfg.Repository.URL = initRepo
	} else if initInteractive {
		repoURL, err := prompts.RepositoryPrompt("")
		if err != nil {
			return err
		}
		cfg.Repository.URL = repoURL
	}

	// Branch
	if initBranch != "" {
		cfg.Repository.Branch = initBranch
	} else if initInteractive {
		branch, err := prompts.PromptInput("Repository branch", "main")
		if err != nil {
			return err
		}
		cfg.Repository.Branch = branch
	}

	return nil
}

func checkExistingConfiguration(projectDir string) error {
	configPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(configPath); err == nil {
		if initInteractive {
			overwrite, err := prompts.PromptConfirm(".stax.yml already exists. Overwrite?", false)
			if err != nil {
				return err
			}
			if !overwrite {
				return fmt.Errorf("initialization cancelled by user")
			}
		} else {
			return errors.NewWithSolution(
				"Configuration already exists",
				".stax.yml already exists in this directory",
				errors.Solution{
					Description: "Choose an action",
					Steps: []string{
						"Run with --interactive to confirm overwrite",
						"Remove .stax.yml manually and try again",
						"Use a different directory",
					},
				},
			)
		}
	}

	// Check for existing DDEV config
	if ddev.IsConfigured(projectDir) {
		ui.Warning("DDEV configuration already exists")
		if initInteractive {
			overwrite, err := prompts.PromptConfirm("Overwrite existing DDEV configuration?", false)
			if err != nil {
				return err
			}
			if !overwrite {
				ui.Info("Will preserve existing DDEV configuration")
			}
		}
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

func generateDDEVConfig(projectDir string, cfg *config.Config) error {
	ui.Section("Generating DDEV Configuration")

	// Check if config already exists
	if ddev.IsConfigured(projectDir) {
		ui.Info("DDEV configuration already exists, skipping generation")
		return nil
	}

	// Prepare DDEV config options
	options := ddev.ConfigOptions{
		ProjectName:        cfg.Project.Name,
		Type:               mapProjectTypeToDDEV(cfg.Project.Type),
		DocRoot:            "public",
		PHPVersion:         cfg.DDEV.PHPVersion,
		DatabaseType:       cfg.DDEV.MySQLType,
		DatabaseVersion:    cfg.DDEV.MySQLVersion,
		RouterHTTPPort:     cfg.DDEV.RouterHTTPPort,
		RouterHTTPSPort:    cfg.DDEV.RouterHTTPSPort,
		XdebugEnabled:      cfg.DDEV.XdebugEnabled,
		UseDNSWhenPossible: cfg.DDEV.UseDNSWhenPossible,
		MutagenEnabled:     cfg.DDEV.MutagenEnabled,
		ComposerVersion:    cfg.DDEV.ComposerVersion,
		NodeJSVersion:      cfg.DDEV.NodeJSVersion,
	}

	// Add additional hostnames for multisite
	if isMultisite(cfg.Project.Type) {
		options.AdditionalHostnames = generateMultisiteHostnames(cfg)
	}

	// Generate config
	ddevConfig, err := ddev.GenerateConfig(projectDir, options)
	if err != nil {
		return fmt.Errorf("failed to generate DDEV config: %w", err)
	}

	// Write config
	if err := ddev.WriteConfig(projectDir, ddevConfig); err != nil {
		return fmt.Errorf("failed to write DDEV config: %w", err)
	}

	ui.Success("Generated DDEV configuration")
	return nil
}

func generateMultisiteNginxConfig(projectDir string, cfg *config.Config) error {
	ui.Section("Generating Multisite Configuration")

	// Create nginx configuration directory
	nginxDir := filepath.Join(projectDir, ".ddev", "nginx_full")
	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		return fmt.Errorf("failed to create nginx directory: %w", err)
	}

	// Generate multisite nginx config based on mode
	var nginxConfig string
	if cfg.Project.Mode == "subdomain" {
		nginxConfig = generateSubdomainNginxConfig(cfg)
	} else {
		nginxConfig = generateSubdirectoryNginxConfig(cfg)
	}

	// Write config
	configPath := filepath.Join(nginxDir, "multisite.conf")
	if err := os.WriteFile(configPath, []byte(nginxConfig), 0644); err != nil {
		return fmt.Errorf("failed to write nginx config: %w", err)
	}

	ui.Success("Generated multisite nginx configuration")
	return nil
}

func shouldPullDatabase(cfg *config.Config) bool {
	if initSkipDB {
		return false
	}

	if initPullDB {
		return true
	}

	if cfg.WPEngine.Install == "" {
		return false
	}

	if initInteractive {
		pull, err := prompts.PromptConfirm("Pull database from WPEngine now?", false)
		if err != nil {
			return false
		}
		return pull
	}

	return false
}

func shouldPullFiles(cfg *config.Config) bool {
	if initSkipFiles {
		return false
	}

	if initPullFiles {
		return true
	}

	if cfg.WPEngine.Install == "" {
		return false
	}

	if initInteractive {
		pull, err := prompts.PromptConfirm("Pull files from WPEngine now?", false)
		if err != nil {
			return false
		}
		return pull
	}

	return false
}

// hasWordPressCore checks if WordPress core files are present in the docroot
func hasWordPressCore(projectDir string) bool {
	// Read DDEV config to get docroot
	ddevConfig, err := ddev.ReadConfig(projectDir)
	docroot := "public" // default
	if err == nil && ddevConfig.DocRoot != "" {
		docroot = ddevConfig.DocRoot
	}

	// Check for wp-includes/version.php as indicator of WordPress core
	versionPath := filepath.Join(projectDir, docroot, "wp-includes", "version.php")
	if _, err := os.Stat(versionPath); err == nil {
		return true
	}

	// Also check for wp-load.php in docroot directory
	loadPath := filepath.Join(projectDir, docroot, "wp-load.php")
	if _, err := os.Stat(loadPath); err == nil {
		return true
	}

	return false
}

// downloadWordPressCore downloads WordPress core files via DDEV to the correct docroot
func downloadWordPressCore(projectDir string, version string) error {
	// Check if DDEV is running
	mgr := ddev.NewManager(projectDir)
	running, err := mgr.IsRunning()
	if err != nil || !running {
		return fmt.Errorf("DDEV must be running to download WordPress core")
	}

	// Read DDEV config to get docroot location
	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read DDEV config: %w", err)
	}

	// Get docroot (default to "public" if not specified)
	docroot := ddevConfig.DocRoot
	if docroot == "" {
		docroot = "public"
	}

	// Ensure docroot directory exists
	docrootPath := filepath.Join(projectDir, docroot)
	if err := os.MkdirAll(docrootPath, 0755); err != nil {
		return fmt.Errorf("failed to create docroot directory: %w", err)
	}

	// Build WP-CLI command with --path to specify docroot
	args := []string{"wp", "core", "download", fmt.Sprintf("--path=%s", docroot)}
	if version != "" && version != "latest" {
		args = append(args, fmt.Sprintf("--version=%s", version))
		ui.Info("Downloading WordPress %s to %s/...", version, docroot)
	} else {
		ui.Info("Downloading latest WordPress to %s/...", docroot)
	}

	// Execute download
	spinner := ui.NewSpinner("Downloading WordPress core...")
	spinner.Start()

	if err := mgr.Exec(args, nil); err != nil {
		spinner.Error("Failed to download WordPress core")
		return fmt.Errorf("WordPress download failed: %w\n\nNote: WordPress should be installed to %s/ directory", err, docroot)
	}

	spinner.Success(fmt.Sprintf("WordPress core downloaded successfully to %s/", docroot))
	return nil
}

// hasWordPressConfig checks if wp-config.php exists in the docroot
func hasWordPressConfig(projectDir string) bool {
	// Read DDEV config to get docroot
	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		// If we can't read config, check default location
		configPath := filepath.Join(projectDir, "public", "wp-config.php")
		_, err := os.Stat(configPath)
		return err == nil
	}

	docroot := ddevConfig.DocRoot
	if docroot == "" {
		docroot = "public"
	}

	configPath := filepath.Join(projectDir, docroot, "wp-config.php")
	_, err = os.Stat(configPath)
	return err == nil
}

// generateWordPressConfig creates wp-config.php via DDEV and configures multisite if needed
func generateWordPressConfig(projectDir string, cfg *config.Config) error {
	ui.Section("Generating wp-config.php")

	// Check if DDEV is running
	mgr := ddev.NewManager(projectDir)
	running, err := mgr.IsRunning()
	if err != nil || !running {
		return fmt.Errorf("DDEV must be running to generate wp-config.php")
	}

	// Read DDEV config to get docroot location
	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read DDEV config: %w", err)
	}

	// Get docroot (default to "public" if not specified)
	docroot := ddevConfig.DocRoot
	if docroot == "" {
		docroot = "public"
	}

	// DDEV database defaults
	dbName := "db"
	dbUser := "db"
	dbPass := "db"
	dbHost := "db"

	// Build WP-CLI command to create wp-config.php in the correct docroot
	args := []string{
		"wp", "config", "create",
		fmt.Sprintf("--path=%s", docroot),
		fmt.Sprintf("--dbname=%s", dbName),
		fmt.Sprintf("--dbuser=%s", dbUser),
		fmt.Sprintf("--dbpass=%s", dbPass),
		fmt.Sprintf("--dbhost=%s", dbHost),
		"--skip-check", // Skip DB connection check (DB may not have tables yet)
	}

	// Execute config creation
	spinner := ui.NewSpinner("Creating wp-config.php...")
	spinner.Start()

	if err := mgr.Exec(args, nil); err != nil {
		spinner.Error("Failed to generate wp-config.php")
		return err
	}

	spinner.Success("wp-config.php created successfully")

	// For multisite, add additional configuration
	if isMultisite(cfg.Project.Type) {
		if err := configureMultisite(projectDir, cfg, mgr); err != nil {
			ui.Warning(fmt.Sprintf("Multisite configuration warning: %v", err))
			ui.Info("You may need to manually configure multisite in wp-config.php")
		}
	}

	ui.Success("wp-config.php generated successfully")
	return nil
}

// configureMultisite adds multisite constants to wp-config.php
func configureMultisite(projectDir string, cfg *config.Config, mgr *ddev.Manager) error {
	ui.Info("Adding multisite configuration...")

	// Determine subdomain vs subdirectory
	subdomainInstall := "true"
	if cfg.Project.Mode == "subdirectory" {
		subdomainInstall = "false"
	}

	// Get the domain from config
	domain := cfg.Network.Domain
	if domain == "" {
		domain = fmt.Sprintf("%s.ddev.site", cfg.Project.Name)
	}

	// Add multisite constants via wp-cli
	multisiteConstants := []struct {
		name  string
		value string
		raw   bool
	}{
		{"MULTISITE", "true", true},
		{"SUBDOMAIN_INSTALL", subdomainInstall, true},
		{"DOMAIN_CURRENT_SITE", fmt.Sprintf("'%s'", domain), false},
		{"PATH_CURRENT_SITE", "'/'", false},
		{"SITE_ID_CURRENT_SITE", "1", true},
		{"BLOG_ID_CURRENT_SITE", "1", true},
	}

	for _, constant := range multisiteConstants {
		args := []string{"wp", "config", "set", constant.name, constant.value}
		if constant.raw {
			args = append(args, "--raw")
		}

		if err := mgr.Exec(args, &ddev.ExecOptions{Service: "web"}); err != nil {
			return fmt.Errorf("failed to set %s: %w", constant.name, err)
		}
	}

	ui.Success("Multisite configuration added")
	return nil
}
func pullDatabase(projectDir string, cfg *config.Config) error {
	ui.Section("Pulling Database")
	ui.Info("This may take several minutes...")

	// Verify WPEngine configuration exists
	if cfg.WPEngine.Install == "" {
		return fmt.Errorf("WPEngine install not configured")
	}

	// Save current directory and change to project directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			ui.Debug("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}

	// Set the environment for database pull
	dbEnvironment = cfg.WPEngine.Environment
	if dbEnvironment == "" {
		dbEnvironment = "production"
	}

	// Call the existing database pull function
	if err := runDBPull(nil, nil); err != nil {
		return fmt.Errorf("database pull failed: %w\n\nYou can try manually: stax db pull --environment=%s", err, dbEnvironment)
	}

	ui.Success("Database pulled successfully")
	return nil
}

// pullFiles pulls WordPress files from WPEngine.
// If media proxy is enabled in the config, it excludes the uploads directory
// and only syncs themes, plugins, and mu-plugins. Otherwise, it syncs all files
// including uploads.
//
// Media proxy configuration is found in cfg.Media.ProxyEnabled. When enabled,
// media files are served from the remote WPEngine server on-demand rather than
// being stored locally, significantly reducing sync time and disk usage.
func pullFiles(projectDir string, cfg *config.Config) error {
	ui.Section("Pulling Files")

	// Verify WPEngine configuration exists
	if cfg.WPEngine.Install == "" {
		return fmt.Errorf("WPEngine install not configured")
	}

	// Save current directory and change to project directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			ui.Debug("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}

	// Set the environment for file pull
	filesEnvironment = cfg.WPEngine.Environment
	if filesEnvironment == "" {
		filesEnvironment = "production"
	}

	// Only exclude uploads if media proxy is enabled in config
	if cfg.Media.ProxyEnabled {
		filesExcludeUploads = true
		ui.Info("Media proxy enabled - excluding uploads directory")
		ui.Info("Syncing: themes, plugins, mu-plugins")
		ui.Info("This may take several minutes...")
	} else {
		filesExcludeUploads = false
		ui.Info("Media proxy disabled - pulling all files including uploads")
		ui.Info("Syncing: themes, plugins, mu-plugins, uploads")
		ui.Info("This may take longer due to media files...")
	}

	// Call the existing file pull function
	if err := runFilesPull(nil, nil); err != nil {
		// Provide helpful error message with next steps
		if cfg.Media.ProxyEnabled {
			return fmt.Errorf("file pull failed: %w\n\nYou can try manually: stax files pull --environment=%s --exclude-uploads", err, filesEnvironment)
		}
		return fmt.Errorf("file pull failed: %w\n\nYou can try manually: stax files pull --environment=%s", err, filesEnvironment)
	}

	ui.Success("Files pulled successfully")

	// Validate that critical directories exist and have content
	if err := validatePulledFiles(projectDir, cfg); err != nil {
		ui.Warning("File validation warnings detected:")
		ui.Warning(err.Error())
		ui.Info("\nNext steps:")
		ui.Info("  1. Check your WPEngine install has themes and plugins")
		ui.Info("  2. Verify SSH access: stax files pull --environment=%s", filesEnvironment)
		ui.Info("  3. Check .stax.yml configuration")
	}

	return nil
}

// validatePulledFiles checks that critical WordPress directories exist and have content.
// This helps catch sync issues early and provides clear feedback to users.
func validatePulledFiles(projectDir string, cfg *config.Config) error {
	wpContentDir := filepath.Join(projectDir, "public", "wp-content")
	var warnings []string

	// Check themes directory
	themesDir := filepath.Join(wpContentDir, "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		warnings = append(warnings, "themes directory not found")
	} else {
		// Check if themes directory has content
		entries, err := os.ReadDir(themesDir)
		if err == nil && len(entries) == 0 {
			warnings = append(warnings, "themes directory is empty")
		}
	}

	// Check plugins directory
	pluginsDir := filepath.Join(wpContentDir, "plugins")
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		warnings = append(warnings, "plugins directory not found")
	}

	// Check mu-plugins directory (optional, so only warn if expected)
	muPluginsDir := filepath.Join(wpContentDir, "mu-plugins")
	if _, err := os.Stat(muPluginsDir); os.IsNotExist(err) {
		ui.Info("Note: mu-plugins directory not found (this is optional)")
	}

	// Check uploads directory only if media proxy is disabled
	if !cfg.Media.ProxyEnabled {
		uploadsDir := filepath.Join(wpContentDir, "uploads")
		if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
			warnings = append(warnings, "uploads directory not found (media proxy disabled)")
		}
	}

	if len(warnings) > 0 {
		return fmt.Errorf("%s", strings.Join(warnings, "; "))
	}

	ui.Success("File validation passed")
	ui.Info("  - themes: present")
	ui.Info("  - plugins: present")

	if cfg.Media.ProxyEnabled {
		ui.Info("  - uploads: excluded (media proxy enabled)")
	} else {
		ui.Info("  - uploads: synced")
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

	if cfg.WPEngine.Install != "" && !initPullDB {
		ui.ProgressMsg("stax db pull       - Pull database from WPEngine")
	}

	if cfg.WPEngine.Install != "" && !initPullFiles {
		ui.ProgressMsg("stax files pull    - Pull files from WPEngine")
	}

	ui.ProgressMsg("stax status        - View environment status")

	fmt.Println()
	ui.Success("Your site will be available at: https://%s.ddev.site", cfg.Project.Name)
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

		cfg.WPEngine.Install = install
		cfg.WPEngine.Environment = env
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

func generateTemplate() error {
	cfg := config.Defaults()
	cfg.Project.Name = "example-project"
	cfg.WPEngine.Install = "example-install"
	cfg.Network.Domain = "example.ddev.site"

	data, err := cfg.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func showExample() error {
	example := `# Stax Configuration File
# Version: 1

version: 1

# Project metadata
project:
  name: myproject
  type: wordpress-multisite  # wordpress, wordpress-multisite
  mode: subdomain            # subdomain, subdirectory, single
  description: My WordPress project

# WPEngine integration
wpengine:
  install: myinstall
  environment: production    # production, staging, development
  ssh_gateway: ssh.wpengine.net
  backup:
    auto_snapshot: true
    skip_logs: true
    skip_transients: true
    skip_spam: true

# Network configuration (for multisite)
network:
  domain: myproject.ddev.site
  title: My Network
  sites: []

# DDEV configuration
ddev:
  php_version: "8.1"
  mysql_version: "8.0"
  mysql_type: mysql
  webserver_type: nginx-fpm
  router_http_port: "80"
  router_https_port: "443"
  nfs_mount_enabled: false
  mutagen_enabled: true      # Enable on macOS for better performance
  xdebug_enabled: false
  nodejs_version: "20"
  composer_version: "2"

# Repository configuration
repository:
  url: https://github.com/org/repo.git
  branch: main
  private: true
  depth: 1
  submodules: false

# WordPress configuration
wordpress:
  version: latest
  locale: en_US
  table_prefix: wp_

# Media proxy configuration
media:
  proxy_enabled: true
  wpengine_fallback: true
  cache:
    enabled: true
    directory: .stax/media-cache
    max_size: 1GB
    ttl: 86400

# Logging configuration
logging:
  level: info
  file: ~/.stax/logs/stax.log
  format: json

# Snapshot configuration
snapshots:
  directory: ~/.stax/snapshots
  auto_snapshot_before_pull: true
  auto_snapshot_before_import: true
  retention:
    auto: 7    # days
    manual: 30 # days

# Performance configuration
performance:
  parallel_downloads: 4
  rsync_bandwidth_limit: 0
  database_import_batch_size: 1000
`

	fmt.Println(example)
	return nil
}

// Helper functions

func isMultisite(projectType string) bool {
	return strings.Contains(projectType, "multisite")
}

func mapProjectTypeToDDEV(projectType string) string {
	switch {
	case strings.Contains(projectType, "multisite"):
		return "wordpress"
	default:
		return projectType
	}
}

func mapDDEVTypeToStax(ddevType string) string {
	// DDEV doesn't distinguish multisite, so default to single
	return "wordpress"
}

func promptProjectType() (string, error) {
	options := []string{
		"wordpress",
		"wordpress-multisite",
	}

	idx, selected, err := prompts.PromptSelect("Select project type:", options, 0)
	if err != nil {
		return "", err
	}

	ui.Info("Selected: %s", selected)
	return options[idx], nil
}

func promptMultisiteMode() (string, error) {
	options := []string{
		"subdomain",
		"subdirectory",
	}

	idx, selected, err := prompts.PromptSelect("Select multisite mode:", options, 0)
	if err != nil {
		return "", err
	}

	ui.Info("Selected: %s", selected)
	return options[idx], nil
}

func generateMultisiteHostnames(cfg *config.Config) []string {
	// For subdomain multisite, generate wildcard hostname
	if cfg.Project.Mode == "subdomain" {
		return []string{
			fmt.Sprintf("*.%s.ddev.site", cfg.Project.Name),
		}
	}
	return []string{}
}

func generateSubdomainNginxConfig(cfg *config.Config) string {
	return fmt.Sprintf(`# WordPress Multisite (subdomain) configuration
# Generated by Stax

# Handle wildcard subdomains for multisite
server_name_in_redirect off;

# Multisite subdomain rewrite rules
if (!-e $request_filename) {
    rewrite /wp-admin$ $scheme://$host$uri/ permanent;
    rewrite ^(/[^/]+)?(/wp-.*) $2 last;
    rewrite ^(/[^/]+)?(/.*\.php) $2 last;
}

# Additional multisite configuration
location / {
    try_files $uri $uri/ /index.php?$args;
}

# Domain: %s
`, cfg.Network.Domain)
}

func generateSubdirectoryNginxConfig(cfg *config.Config) string {
	return fmt.Sprintf(`# WordPress Multisite (subdirectory) configuration
# Generated by Stax

# Multisite subdirectory rewrite rules
if (!-e $request_filename) {
    rewrite /wp-admin$ $scheme://$host$uri/ permanent;
    rewrite ^(/[^/]+)?(/wp-.*) $2 last;
    rewrite ^(/[^/]+)?(/.*\.php) $2 last;
}

# Additional multisite configuration
location / {
    try_files $uri $uri/ /index.php?$args;
}

# Domain: %s
`, cfg.Network.Domain)
}

func createWPEngineProviderForListing(creds *credentials.WPEngineCredentials) (provider.Provider, error) {
	p := &wpengine.WPEngineProvider{}

	credMap := map[string]string{
		"api_user":     creds.APIUser,
		"api_password": creds.APIPassword,
		"install":      "temp",
		"ssh_gateway":  "ssh.wpengine.net",
		"ssh_key":      "",
	}

	if err := p.Authenticate(credMap); err != nil {
		return nil, err
	}

	return p, nil
}
