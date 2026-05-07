package cmd

import (
	staxinit "github.com/firecrown-media/stax/pkg/init"
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

	initCmd.Flags().StringVar(&initName, "name", "", "project name (default: current directory name)")
	initCmd.Flags().StringVar(&initType, "type", "wordpress", "project type (wordpress, wordpress-multisite)")
	initCmd.Flags().StringVar(&initMode, "mode", "subdomain", "multisite mode (subdomain, subdirectory)")
	initCmd.Flags().StringVar(&initPHPVersion, "php", "8.1", "PHP version")
	initCmd.Flags().StringVar(&initMySQLVersion, "mysql", "8.0", "MySQL version")

	initCmd.Flags().StringVar(&initRepo, "repo", "", "GitHub repository URL")
	initCmd.Flags().StringVar(&initBranch, "branch", "main", "repository branch")

	initCmd.Flags().StringVar(&initWPEngineInstall, "install", "", "WPEngine install name")
	initCmd.Flags().StringVar(&initWPEngineEnv, "environment", "production", "WPEngine environment (production, staging, development)")

	initCmd.Flags().BoolVar(&initInteractive, "interactive", prompts.IsInteractive(), "enable interactive prompts (default: auto-detect TTY)")
	initCmd.Flags().BoolVar(&initStart, "start", false, "start DDEV after initialization")
	initCmd.Flags().BoolVar(&initPullDB, "pull-db", false, "pull database after initialization")
	initCmd.Flags().BoolVar(&initPullFiles, "pull-files", false, "pull files after initialization")
	initCmd.Flags().BoolVar(&initSkipDB, "skip-db", false, "skip database operations")
	initCmd.Flags().BoolVar(&initSkipFiles, "skip-files", false, "skip file operations")
	initCmd.Flags().BoolVar(&initSkipWordPress, "skip-wordpress", false, "skip WordPress core download")
	initCmd.Flags().StringVar(&initWordPressVersion, "wordpress-version", "", "WordPress version to download (default: latest)")

	initCmd.Flags().BoolVar(&initFromDDEV, "from-ddev", false, "import existing DDEV project")
	initCmd.Flags().BoolVar(&initTemplate, "template", false, "generate .stax.yml template to stdout")
	initCmd.Flags().BoolVar(&initShowExample, "show-example", false, "show example configuration with comments")
}

func runInit(cmd *cobra.Command, args []string) error {
	if initTemplate {
		return staxinit.GenerateTemplate()
	}

	if initShowExample {
		return staxinit.ShowExample()
	}

	ui.PrintHeader("Initializing Stax Project")

	projectDir := getProjectDir()

	if initFromDDEV {
		return staxinit.RunFromDDEV(projectDir)
	}

	return staxinit.Run(staxinit.Options{
		Name:             initName,
		Type:             initType,
		Mode:             initMode,
		PHPVersion:       initPHPVersion,
		MySQLVersion:     initMySQLVersion,
		Repo:             initRepo,
		Branch:           initBranch,
		WPEngineInstall:  initWPEngineInstall,
		WPEngineEnv:      initWPEngineEnv,
		Interactive:      initInteractive,
		SkipDB:           initSkipDB,
		SkipFiles:        initSkipFiles,
		Start:            initStart,
		PullDB:           initPullDB,
		PullFiles:        initPullFiles,
		SkipWordPress:    initSkipWordPress,
		WordPressVersion: initWordPressVersion,
		ProjectDir:       projectDir,
	})
}
