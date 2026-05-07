package staxinit

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/database"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/files"
	"github.com/firecrown-media/stax/pkg/git"
	"github.com/firecrown-media/stax/pkg/prompts"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/providers/wpengine"
	"github.com/firecrown-media/stax/pkg/ui"
	wpeclient "github.com/firecrown-media/stax/pkg/wpengine"
)

// Options holds all parameters needed to run stax init.
type Options struct {
	Name             string
	Type             string
	Mode             string
	PHPVersion       string
	MySQLVersion     string
	Repo             string
	Branch           string
	WPEngineInstall  string
	WPEngineEnv      string
	Interactive      bool
	SkipDB           bool
	SkipFiles        bool
	FromDDEV         bool
	Start            bool
	PullDB           bool
	PullFiles        bool
	SkipWordPress    bool
	WordPressVersion string
	ProjectDir       string
}

// ValidateOptions checks that required fields are present.
func ValidateOptions(opts Options) error {
	if opts.Name == "" {
		return fmt.Errorf("project name is required (--name)")
	}
	return nil
}

// Run executes the full init workflow.
func Run(opts Options) error {
	ui.Section("Checking prerequisites")
	if err := checkPrerequisites(opts); err != nil {
		return err
	}

	cfg, err := gatherProjectConfiguration(opts.ProjectDir, opts)
	if err != nil {
		return err
	}

	if err := checkExistingConfiguration(opts.ProjectDir, opts.Interactive); err != nil {
		return err
	}

	if cfg.Repository.URL != "" {
		if err := cloneRepository(opts.ProjectDir, cfg); err != nil {
			return err
		}
	}

	if err := generateDDEVConfig(opts.ProjectDir, cfg); err != nil {
		return err
	}

	if isMultisite(cfg.Project.Type) {
		if err := generateMultisiteNginxConfig(opts.ProjectDir, cfg); err != nil {
			return err
		}
	}

	configPath := filepath.Join(opts.ProjectDir, ".stax.yml")
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	ui.Success("Created .stax.yml")

	if err := git.EnsureGitignore(opts.ProjectDir); err != nil {
		ui.Warning("Failed to manage .gitignore: %v", err)
	}

	shouldStart := opts.Start
	if opts.Interactive && !opts.Start {
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

		mgr := ddev.NewManager(opts.ProjectDir)
		if err := mgr.Start(); err != nil {
			spinner.Error("Failed to start DDEV")
			return err
		}
		spinner.Success("DDEV started successfully")

		ui.Info("Waiting for services to be ready...")
		waitSpinner := ui.NewSpinner("Checking service status...")
		waitSpinner.Start()
		if err := mgr.WaitForReady(2 * time.Minute); err != nil {
			waitSpinner.Stop()
			ui.Warning("Services may not be fully ready yet: %v", err)
			ui.Info("Continuing with setup...")
		} else {
			waitSpinner.Success("Services are ready")
		}

		if !opts.SkipWordPress {
			if hasWordPressCore(opts.ProjectDir) {
				ui.Section("WordPress Core")
				ui.Info("WordPress core files already exist, skipping download")
			} else {
				ui.Section("Setting Up WordPress")

				version := "latest"
				if opts.WordPressVersion != "" {
					version = opts.WordPressVersion
					cfg.WordPress.Version = version
				} else if cfg.WordPress.Version != "" {
					version = cfg.WordPress.Version
				}

				if err := downloadWordPressCore(opts.ProjectDir, version); err != nil {
					ui.Warning("Failed to download WordPress core: %v", err)
					ui.Info("You can download manually: ddev wp core download")
				}
			}
		} else {
			ui.Info("Skipping WordPress core download (--skip-wordpress flag set)")
		}

		if !opts.SkipWordPress && !hasWordPressConfig(opts.ProjectDir) {
			if err := generateWordPressConfig(opts.ProjectDir, cfg); err != nil {
				ui.Warning("Failed to generate wp-config.php: %v", err)
				ui.Info("You can create manually: ddev wp config create --dbname=db --dbuser=db --dbpass=db --dbhost=db")
			}
		}
	}

	if shouldPullDatabase(cfg, opts) {
		if err := pullDatabase(opts.ProjectDir, cfg); err != nil {
			ui.Warning("Database pull failed: %v", err)
		}
	}

	if shouldPullFiles(cfg, opts) {
		if err := pullFiles(opts.ProjectDir, cfg); err != nil {
			ui.Warning("File pull failed: %v", err)
		}
	}

	printSuccessSummary(opts.ProjectDir, cfg, opts)

	return nil
}

// RunFromDDEV imports an existing DDEV project into Stax.
func RunFromDDEV(projectDir string) error {
	ui.Info("Importing existing DDEV project...")

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

	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read DDEV config: %w", err)
	}

	ui.Success("Found DDEV configuration")

	cfg := config.Defaults()
	cfg.Project.Name = ddevConfig.Name
	cfg.Project.Type = mapDDEVTypeToStax(ddevConfig.Type)
	cfg.DDEV.PHPVersion = ddevConfig.PHPVersion
	cfg.DDEV.MySQLVersion = ddevConfig.Database.Version
	cfg.DDEV.MySQLType = ddevConfig.Database.Type

	ui.Info("\nOptional: Configure WPEngine integration")
	fmt.Print("Add WPEngine integration? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	addWPEngine := (response == "y" || response == "yes")

	if addWPEngine {
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

// GenerateTemplate outputs a template .stax.yml configuration to stdout.
func GenerateTemplate() error {
	cfg := config.Defaults()
	cfg.Project.Name = "example-project"
	cfg.ProviderConfig["install"] = "example-install"
	cfg.Network.Domain = "example.ddev.site"

	data, err := cfg.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// ShowExample outputs an example configuration with comments to stdout.
func ShowExample() error {
	example := `# Stax Configuration File
# Version: 2

version: 2

# Project metadata
project:
  name: myproject
  type: wordpress-multisite  # wordpress, wordpress-multisite
  mode: subdomain            # subdomain, subdirectory, single
  description: My WordPress project

# Provider integration
provider: wpengine
provider_config:
  install: myinstall
  environment: production    # production, staging, development
  ssh_gateway: ssh.wpengine.net

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

// ---- internal helpers ----

func providerConfigStr(cfg *config.Config, key string) string {
	if cfg.ProviderConfig == nil {
		return ""
	}
	v, _ := cfg.ProviderConfig[key].(string)
	return v
}

func checkPrerequisites(opts Options) error {
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

	if opts.Repo != "" || (opts.Interactive && !opts.FromDDEV) {
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

func gatherProjectConfiguration(projectDir string, opts Options) (*config.Config, error) {
	ui.Section("Project Configuration")

	cfg := config.Defaults()

	defaultName := filepath.Base(projectDir)
	if opts.Name != "" {
		cfg.Project.Name = opts.Name
	} else if opts.Interactive {
		name, err := prompts.SafePromptInput("Project name", defaultName, false)
		if err != nil {
			return nil, err
		}
		cfg.Project.Name = name
	} else {
		cfg.Project.Name = defaultName
	}

	if opts.Type != "" {
		cfg.Project.Type = opts.Type
		if strings.Contains(opts.Type, "multisite") {
			if opts.Mode != "" {
				cfg.Project.Mode = opts.Mode
			}
		} else {
			cfg.Project.Mode = "single"
		}
	} else if opts.Interactive {
		projectType, err := prompts.SafeProjectTypePrompt()
		if err != nil {
			return nil, err
		}
		cfg.Project.Type = projectType

		if isMultisite(projectType) {
			mode, err := promptMultisiteModeSafe()
			if err != nil {
				return nil, err
			}
			cfg.Project.Mode = mode
		} else {
			cfg.Project.Mode = "single"
		}
	}

	cfg.DDEV.PHPVersion = opts.PHPVersion
	cfg.DDEV.MySQLVersion = opts.MySQLVersion

	if err := gatherWPEngineConfiguration(cfg, opts); err != nil {
		return nil, err
	}

	if opts.Interactive {
		if cfg.DDEV.PHPVersion == opts.PHPVersion {
			phpVersion, err := prompts.SafePromptInput("PHP version", cfg.DDEV.PHPVersion, false)
			if err != nil {
				return nil, err
			}
			cfg.DDEV.PHPVersion = phpVersion
		}

		if cfg.DDEV.MySQLVersion == opts.MySQLVersion {
			mysqlVersion, err := prompts.SafePromptInput("MySQL version", cfg.DDEV.MySQLVersion, false)
			if err != nil {
				return nil, err
			}
			cfg.DDEV.MySQLVersion = mysqlVersion
		}
	}

	if err := gatherRepositoryConfiguration(cfg, opts); err != nil {
		return nil, err
	}

	if isMultisite(cfg.Project.Type) {
		if opts.Interactive {
			defaultDomain := fmt.Sprintf("%s.ddev.site", cfg.Project.Name)
			domain, err := prompts.SafePromptInput("Primary domain", defaultDomain, false)
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

func gatherRepositoryConfiguration(cfg *config.Config, opts Options) error {
	ui.Section("Repository Configuration")

	cloneRepo := false
	if opts.Repo != "" {
		cloneRepo = true
	} else if opts.Interactive {
		var err error
		cloneRepo, err = prompts.SafePromptConfirm("Clone from Git repository?", false)
		if err != nil {
			return err
		}
	}

	if !cloneRepo {
		ui.Info("Skipping repository cloning")
		return nil
	}

	if opts.Repo != "" {
		cfg.Repository.URL = opts.Repo
	} else if opts.Interactive {
		repoURL, err := prompts.SafeRepositoryPrompt("")
		if err != nil {
			return err
		}
		cfg.Repository.URL = repoURL
	}

	if opts.Branch != "" {
		cfg.Repository.Branch = opts.Branch
	} else if opts.Interactive {
		branch, err := prompts.SafePromptInput("Repository branch", "main", false)
		if err != nil {
			return err
		}
		cfg.Repository.Branch = branch
	}

	return nil
}

func checkExistingConfiguration(projectDir string, interactive bool) error {
	configPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(configPath); err == nil {
		if interactive {
			overwrite, err := prompts.SafePromptConfirm(".stax.yml already exists. Overwrite?", false)
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

	if ddev.IsConfigured(projectDir) {
		ui.Warning("DDEV configuration already exists")
		if interactive {
			overwrite, err := prompts.SafePromptConfirm("Overwrite existing DDEV configuration?", false)
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
		Quiet:       false,
	}

	if err := git.Clone(opts); err != nil {
		spinner.Error("Failed to clone repository")
		return err
	}

	spinner.Success("Repository cloned successfully")
	return nil
}

func pullDatabase(projectDir string, cfg *config.Config) error {
	ui.Section("Pulling Database")
	ui.Info("This may take several minutes...")

	if providerConfigStr(cfg, "install") == "" {
		return fmt.Errorf("WPEngine install not configured")
	}

	env := providerConfigStr(cfg, "environment")
	if env == "" {
		env = "production"
	}

	p, err := provider.NewAuthenticatedProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to authenticate provider: %w", err)
	}

	if err := database.Pull(p, cfg, database.PullOptions{
		ProjectDir:  projectDir,
		Environment: env,
	}); err != nil {
		return fmt.Errorf("database pull failed: %w\n\nYou can try manually: stax db pull --environment=%s", err, env)
	}

	ui.Success("Database pulled successfully")
	return nil
}

func pullFiles(projectDir string, cfg *config.Config) error {
	ui.Section("Pulling Files")

	if providerConfigStr(cfg, "install") == "" {
		return fmt.Errorf("WPEngine install not configured")
	}

	env := providerConfigStr(cfg, "environment")
	if env == "" {
		env = "production"
	}

	excludeUploads := cfg.Media.ProxyEnabled
	if excludeUploads {
		ui.Info("Media proxy enabled - excluding uploads directory")
		ui.Info("Syncing: themes, plugins, mu-plugins")
	} else {
		ui.Info("Media proxy disabled - pulling all files including uploads")
		ui.Info("Syncing: themes, plugins, mu-plugins, uploads")
	}
	ui.Info("This may take several minutes...")

	fp, err := provider.NewAuthenticatedProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to authenticate provider: %w", err)
	}

	if err := files.Pull(fp, cfg, files.SyncFlags{
		Environment:    env,
		ExcludeUploads: excludeUploads,
		ProjectDir:     projectDir,
	}); err != nil {
		if excludeUploads {
			return fmt.Errorf("file pull failed: %w\n\nYou can try manually: stax files pull --environment=%s --exclude-uploads", err, env)
		}
		return fmt.Errorf("file pull failed: %w\n\nYou can try manually: stax files pull --environment=%s", err, env)
	}

	ui.Success("Files pulled successfully")

	if err := validatePulledFiles(projectDir, cfg); err != nil {
		ui.Warning("File validation warnings detected:")
		ui.Warning(err.Error()
		ui.Info("\nNext steps:")
		ui.Info("  1. Check your WPEngine install has themes and plugins")
		ui.Info("  2. Verify SSH access: stax files pull --environment=%s", env)
		ui.Info("  3. Check .stax.yml configuration")
	}

	return nil
}

func validatePulledFiles(projectDir string, cfg *config.Config) error {
	ddevConfig, err := ddev.ReadConfig(projectDir)
	docroot := "."
	if err == nil && ddevConfig.DocRoot != "" {
		docroot = ddevConfig.DocRoot
	}

	wpContentDir := filepath.Join(projectDir, docroot, "wp-content")
	var warnings []string

	themesDir := filepath.Join(wpContentDir, "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		warnings = append(warnings, "themes directory not found")
	} else {
		entries, err := os.ReadDir(themesDir)
		if err == nil && len(entries) == 0 {
			warnings = append(warnings, "themes directory is empty")
		}
	}

	pluginsDir := filepath.Join(wpContentDir, "plugins")
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		warnings = append(warnings, "plugins directory not found")
	}

	muPluginsDir := filepath.Join(wpContentDir, "mu-plugins")
	if _, err := os.Stat(muPluginsDir); os.IsNotExist(err) {
		ui.Info("Note: mu-plugins directory not found (this is optional)")
	}

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

func shouldPullDatabase(cfg *config.Config, opts Options) bool {
	if opts.SkipDB {
		return false
	}
	if opts.PullDB {
		return true
	}
	if providerConfigStr(cfg, "install") == "" {
		return false
	}
	if opts.Interactive {
		pull, err := prompts.SafePromptConfirm("Pull database from WPEngine now?", false)
		if err != nil {
			return false
		}
		return pull
	}
	return false
}

func shouldPullFiles(cfg *config.Config, opts Options) bool {
	if opts.SkipFiles {
		return false
	}
	if opts.PullFiles {
		return true
	}
	if providerConfigStr(cfg, "install") == "" {
		return false
	}
	if opts.Interactive {
		pull, err := prompts.SafePromptConfirm("Pull files from WPEngine now?", false)
		if err != nil {
			return false
		}
		return pull
	}
	return false
}

func printSuccessSummary(projectDir string, cfg *config.Config, opts Options) {
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

	if !opts.Start {
		ui.ProgressMsg("stax start         - Start DDEV environment")
	}

	if providerConfigStr(cfg, "install") != "" && !opts.PullDB {
		ui.ProgressMsg("stax db pull       - Pull database from WPEngine")
	}

	if providerConfigStr(cfg, "install") != "" && !opts.PullFiles {
		ui.ProgressMsg("stax files pull    - Pull files from WPEngine")
	}

	ui.ProgressMsg("stax status        - View environment status")

	fmt.Println()
	ui.Success("Your site will be available at: https://%s.ddev.site", cfg.Project.Name)
}

// ---- DDEV helpers (from cmd/init_ddev.go) ----

func generateDDEVConfig(projectDir string, cfg *config.Config) error {
	ui.Section("Generating DDEV Configuration")

	if ddev.IsConfigured(projectDir) {
		ui.Info("DDEV configuration already exists, skipping generation")
		return nil
	}

	options := ddev.ConfigOptions{
		ProjectName:        cfg.Project.Name,
		Type:               mapProjectTypeToDDEV(cfg.Project.Type),
		DocRoot:            ".",
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

	if isMultisite(cfg.Project.Type) {
		options.AdditionalHostnames = generateMultisiteHostnames(cfg)
	}

	ddevConfig, err := ddev.GenerateConfig(projectDir, options)
	if err != nil {
		return fmt.Errorf("failed to generate DDEV config: %w", err)
	}

	if err := ddev.WriteConfig(projectDir, ddevConfig); err != nil {
		return fmt.Errorf("failed to write DDEV config: %w", err)
	}

	ui.Success("Generated DDEV configuration")
	return nil
}

func generateMultisiteNginxConfig(projectDir string, cfg *config.Config) error {
	ui.Section("Generating Multisite Configuration")

	nginxDir := filepath.Join(projectDir, ".ddev", "nginx_full")
	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		return fmt.Errorf("failed to create nginx directory: %w", err)
	}

	var nginxConfig string
	if cfg.Project.Mode == "subdomain" {
		nginxConfig = generateSubdomainNginxConfig(cfg)
	} else {
		nginxConfig = generateSubdirectoryNginxConfig(cfg)
	}

	configPath := filepath.Join(nginxDir, "multisite.conf")
	if err := os.WriteFile(configPath, []byte(nginxConfig), 0644); err != nil {
		return fmt.Errorf("failed to write nginx config: %w", err)
	}

	ui.Success("Generated multisite nginx configuration")
	return nil
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

func mapProjectTypeToDDEV(projectType string) string {
	switch {
	case strings.Contains(projectType, "multisite"):
		return "wordpress"
	default:
		return projectType
	}
}

func mapDDEVTypeToStax(_ string) string {
	return "wordpress"
}

func generateMultisiteHostnames(cfg *config.Config) []string {
	if cfg.Project.Mode == "subdomain" {
		return []string{
			fmt.Sprintf("*.%s.ddev.site", cfg.Project.Name),
		}
	}
	return []string{}
}

func isMultisite(projectType string) bool {
	return strings.Contains(projectType, "multisite")
}

// ---- WordPress helpers (from cmd/init_wordpress.go) ----

func hasWordPressCore(projectDir string) bool {
	ddevConfig, err := ddev.ReadConfig(projectDir)
	docroot := "."
	if err == nil && ddevConfig.DocRoot != "" {
		docroot = ddevConfig.DocRoot
	}

	versionPath := filepath.Join(projectDir, docroot, "wp-includes", "version.php")
	if _, err := os.Stat(versionPath); err == nil {
		return true
	}

	loadPath := filepath.Join(projectDir, docroot, "wp-load.php")
	if _, err := os.Stat(loadPath); err == nil {
		return true
	}

	return false
}

func downloadWordPressCore(projectDir string, version string) error {
	mgr := ddev.NewManager(projectDir)
	running, err := mgr.IsRunning()
	if err != nil || !running {
		return fmt.Errorf("DDEV must be running to download WordPress core")
	}

	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read DDEV config: %w", err)
	}

	docroot := ddevConfig.DocRoot
	if docroot == "" {
		docroot = "."
	}

	docrootPath := filepath.Join(projectDir, docroot)
	if err := os.MkdirAll(docrootPath, 0755); err != nil {
		return fmt.Errorf("failed to create docroot directory: %w", err)
	}

	args := []string{"wp", "core", "download", fmt.Sprintf("--path=%s", docroot)}
	if version != "" && version != "latest" {
		args = append(args, fmt.Sprintf("--version=%s", version))
		ui.Info("Downloading WordPress %s to %s/...", version, docroot)
	} else {
		ui.Info("Downloading latest WordPress to %s/...", docroot)
	}

	spinner := ui.NewSpinner("Downloading WordPress core...")
	spinner.Start()

	if err := mgr.Exec(args, nil); err != nil {
		spinner.Error("Failed to download WordPress core")
		return fmt.Errorf("WordPress download failed: %w\n\nNote: WordPress should be installed to %s/ directory", err, docroot)
	}

	spinner.Success(fmt.Sprintf("WordPress core downloaded successfully to %s/", docroot))
	return nil
}

func hasWordPressConfig(projectDir string) bool {
	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		configPath := filepath.Join(projectDir, "wp-config.php")
		_, err := os.Stat(configPath)
		return err == nil
	}

	docroot := ddevConfig.DocRoot
	if docroot == "" {
		docroot = "."
	}

	configPath := filepath.Join(projectDir, docroot, "wp-config.php")
	_, err = os.Stat(configPath)
	return err == nil
}

func generateWordPressConfig(projectDir string, cfg *config.Config) error {
	ui.Section("Generating wp-config.php")

	mgr := ddev.NewManager(projectDir)
	running, err := mgr.IsRunning()
	if err != nil || !running {
		return fmt.Errorf("DDEV must be running to generate wp-config.php")
	}

	ddevConfig, err := ddev.ReadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read DDEV config: %w", err)
	}

	docroot := ddevConfig.DocRoot
	if docroot == "" {
		docroot = "."
	}

	args := []string{
		"wp", "config", "create",
		fmt.Sprintf("--path=%s", docroot),
		"--dbname=db",
		"--dbuser=db",
		"--dbpass=db",
		"--dbhost=db",
		"--skip-check",
	}

	spinner := ui.NewSpinner("Creating wp-config.php...")
	spinner.Start()

	if err := mgr.Exec(args, nil); err != nil {
		spinner.Error("Failed to generate wp-config.php")
		return err
	}

	spinner.Success("wp-config.php created successfully")

	if isMultisite(cfg.Project.Type) {
		if err := configureMultisite(projectDir, cfg, mgr); err != nil {
			ui.Warning("Multisite configuration warning: %v", err)
			ui.Info("You may need to manually configure multisite in wp-config.php")
		}
	}

	ui.Success("wp-config.php generated successfully")
	return nil
}

func configureMultisite(projectDir string, cfg *config.Config, mgr *ddev.Manager) error {
	ui.Info("Adding multisite configuration...")

	subdomainInstall := "true"
	if cfg.Project.Mode == "subdirectory" {
		subdomainInstall = "false"
	}

	domain := cfg.Network.Domain
	if domain == "" {
		domain = fmt.Sprintf("%s.ddev.site", cfg.Project.Name)
	}

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

// ---- WPEngine helpers (from cmd/init_wpengine.go) ----

func promptForWPEngineInstall() (*prompts.WPEngineInstallWithDetails, error) {
	creds, err := credentials.GetWPEngineCredentialsWithFallback("global")
	if err != nil {
		return nil, fmt.Errorf("no WPEngine credentials found: %w", err)
	}

	client := wpeclient.NewClient(creds.APIUser, creds.APIPassword, "")

	ui.Info("Fetching installs from WPEngine...")
	installs, err := client.ListInstalls()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch installs: %w", err)
	}

	if len(installs) == 0 {
		return nil, fmt.Errorf("no installs found in your WPEngine account")
	}

	installDetails := make([]prompts.WPEngineInstallWithDetails, len(installs))
	for i, install := range installs {
		installDetails[i] = prompts.WPEngineInstallWithDetails{
			Name:        install.Name,
			Environment: install.Environment,
			PHPVersion:  install.PHPVersion,
		}
	}

	selected, err := prompts.WPEngineInstallPickerPrompt(installDetails)
	if err != nil {
		return nil, err
	}

	ui.Info("Fetching install details...")
	fullDetails, err := client.GetInstallByName(selected.Name)
	if err != nil {
		ui.Warning("Could not fetch full install details: %v", err)
		return &selected, nil
	}

	selected.MySQLVersion = fullDetails.MySQLVersion
	selected.WordPressVersion = fullDetails.WordPressVersion

	ui.Debug("API returned - PHP: %s, MySQL: %s, WordPress: %s", fullDetails.PHPVersion, fullDetails.MySQLVersion, fullDetails.WordPressVersion)

	if fullDetails.WordPressVersion == "" {
		ui.Debug("WordPress version not available from WPEngine API - this is normal for some installs")
	}

	return &selected, nil
}

func gatherWPEngineConfiguration(cfg *config.Config, opts Options) error {
	ui.Section("WPEngine Integration")

	setupWPEngine := false
	if opts.WPEngineInstall != "" {
		setupWPEngine = true
	} else if opts.Interactive {
		var err error
		setupWPEngine, err = prompts.SafePromptConfirm("Set up WPEngine integration?", true)
		if err != nil {
			return err
		}
	}

	if !setupWPEngine {
		ui.Info("Skipping WPEngine integration")
		return nil
	}

	var installName string
	var phpVersion string
	var mysqlVersion string
	var autoPopulated bool

	if opts.WPEngineInstall != "" {
		installName = opts.WPEngineInstall
	} else if opts.Interactive {
		usePicker, err := prompts.SafePromptConfirm("Would you like to select from your WPEngine installs?", true)
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if usePicker {
			selected, err := promptForWPEngineInstall()
			if err != nil {
				ui.Warning("Unable to fetch installs: %v", err)
				ui.Info("Falling back to manual entry...")
			} else {
				installName = selected.Name
				phpVersion = selected.PHPVersion
				mysqlVersion = selected.MySQLVersion
				autoPopulated = true

				ui.Success("Selected: %s (%s)", selected.Name, selected.Environment)
				if selected.PHPVersion != "" {
					ui.Info("PHP: %s", selected.PHPVersion)

					usePHP, err := prompts.SafePromptConfirm(fmt.Sprintf("Use PHP %s from WPEngine?", selected.PHPVersion), true)
					if err != nil {
						return fmt.Errorf("prompt failed: %w", err)
					}
					if !usePHP {
						phpVersion = ""
					}
				}
				if selected.MySQLVersion != "" {
					ui.Info("MySQL: %s", selected.MySQLVersion)
				}
				if selected.WordPressVersion != "" {
					ui.Info("WordPress: %s", selected.WordPressVersion)
					cfg.WordPress.Version = selected.WordPressVersion
				}
			}
		}

		if !autoPopulated {
			name, err := prompts.SafePromptInput("WPEngine install name", "", false)
			if err != nil {
				return err
			}
			installName = name
		}
	}

	cfg.ProviderConfig["install"] = installName

	if autoPopulated {
		if phpVersion != "" {
			cfg.DDEV.PHPVersion = phpVersion
		}
		if mysqlVersion != "" {
			cfg.DDEV.MySQLVersion = mysqlVersion
		}
	}

	if opts.WPEngineEnv != "" {
		cfg.ProviderConfig["environment"] = opts.WPEngineEnv
	} else if opts.Interactive {
		env, err := prompts.SafeEnvironmentPrompt(providerConfigStr(cfg, "environment"))
		if err != nil {
			return err
		}
		cfg.ProviderConfig["environment"] = env
	}

	if providerConfigStr(cfg, "install") != "" && providerConfigStr(cfg, "environment") != "" {
		if err := validateAndFixEnvironment(cfg); err != nil {
			ui.Warning("Environment validation: %v", err)
		}
	}

	return nil
}

func validateAndFixEnvironment(cfg *config.Config) error {
	install := providerConfigStr(cfg, "install")
	creds, err := credentials.GetWPEngineCredentialsWithFallback(install)
	if err != nil {
		return nil
	}

	client := wpeclient.NewClient(creds.APIUser, creds.APIPassword, install)
	actualEnv, err := client.GetInstallEnvironment(install)
	if err != nil {
		return nil
	}

	if providerConfigStr(cfg, "environment") == actualEnv {
		return nil
	}

	ui.Warning("Environment mismatch: you selected '%s' but WPEngine reports '%s'", providerConfigStr(cfg, "environment"), actualEnv)

	if !prompts.IsInteractive() {
		ui.Info("Run 'stax doctor --fix' to correct this automatically")
		return nil
	}

	fix, promptErr := prompts.SafePromptConfirm(
		fmt.Sprintf("Update environment to '%s' to match WPEngine?", actualEnv),
		true,
	)
	if promptErr != nil {
		return promptErr
	}

	if fix {
		cfg.ProviderConfig["environment"] = actualEnv
		ui.Success("Environment updated to '%s'", actualEnv)
	} else {
		ui.Info("Keeping configured environment. You can fix later with 'stax doctor --fix'")
	}

	return nil
}

// createWPEngineProviderForListing creates a WPEngine provider for listing installs.
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

// ---- prompt helpers (from cmd/init_config.go) ----

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

func promptMultisiteModeSafe() (string, error) {
	options := []string{
		"subdomain",
		"subdirectory",
	}

	idx, selected, err := prompts.SafePromptSelect("Select multisite mode:", options, 0)
	if err != nil {
		return "", err
	}

	ui.Info("Selected: %s", selected)
	return options[idx], nil
}
