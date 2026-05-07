package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/system"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	startBuild     bool
	startXdebug    bool
	startSkipHooks bool
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "✓ Start the DDEV environment",
	Long: `Start the DDEV environment for the current project.

This command starts all DDEV containers (web, database, router) and
optionally runs the build process and enables Xdebug.

Prerequisites:
  - Docker must be running
  - DDEV must be installed
  - Project must be initialized (stax init or .ddev/config.yaml exists)`,
	Aliases: []string{"up"},
	Example: `  # Basic start
  stax start

  # Start with Xdebug enabled
  stax start --xdebug

  # Start and rebuild
  stax start --build

  # Start without running hooks
  stax start --skip-hooks`,
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolVar(&startBuild, "build", false, "run build process after start")
	startCmd.Flags().BoolVar(&startXdebug, "xdebug", false, "enable Xdebug")
	startCmd.Flags().BoolVar(&startSkipHooks, "skip-hooks", false, "skip running project hooks")
}

func runStart(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Starting DDEV Environment")

	// Get project directory
	projectDir := getProjectDir()

	// Check if we have .stax.yml config
	hasStaxConfig := false
	if cfg != nil {
		hasStaxConfig = true
	}

	// Check if we have DDEV config
	hasDDEVConfig := ddev.IsConfigured(projectDir)

	if !hasStaxConfig && !hasDDEVConfig {
		return errors.NewWithSolution(
			"No project configuration found",
			"Neither .stax.yml nor .ddev/config.yaml exists",
			errors.Solution{
				Description: "Initialize your project",
				Steps: []string{
					"Run 'stax init' to set up a new Stax project",
					"Or run 'ddev config' if you just want basic DDEV",
				},
			},
		)
	}

	if !hasStaxConfig {
		ui.Warning("Using DDEV configuration only (no .stax.yml found)")
		ui.Info("Run 'stax init' to enable Stax features like WPEngine sync")
	}

	// TODO: Check if DDEV is installed
	// TODO: Run ddev start
	// TODO: Enable Xdebug if requested
	// TODO: Run build process if requested
	// TODO: Display environment URLs

	// 1. Validate prerequisites
	if err := validateStartPrerequisites(projectDir); err != nil {
		return err
	}

	// 2. Check if already running
	status, err := ddev.GetStatus(projectDir)
	if err == nil && status.Running {
		ui.Info("Environment is already running")
		ui.Info("  Primary URL: %s", status.PrimaryURL))
		return nil
	}

	// 3. Start DDEV
	spinner := ui.NewSpinner("Starting DDEV containers")
	spinner.Start()

	if err := ddev.Start(projectDir); err != nil {
		spinner.Stop()
		return errors.NewWithSolution(
			"Failed to start DDEV",
			err.Error(),
			errors.Solution{
				Description: "Troubleshooting steps",
				Steps: []string{
					"1. Verify Docker is running: docker info",
					"2. Check DDEV status: ddev describe",
					"3. Try stopping first: stax stop",
					"4. Check for port conflicts: stax doctor",
				},
			},
		)
	}

	spinner.Success("DDEV containers started")

	// 4. Wait for services to be ready
	ui.Info("Waiting for services to be ready...")
	if err := waitForServices(projectDir); err != nil {
		ui.Warning("Services may not be fully ready yet")
		ui.Info("You can check status with: stax status")
	}

	// 5. Enable Xdebug if requested
	if startXdebug {
		spinner = ui.NewSpinner("Enabling Xdebug")
		spinner.Start()

		if err := ddev.EnableXdebug(projectDir); err != nil {
			spinner.Stop()
			ui.Warning("Failed to enable Xdebug: %v", err))
		} else {
			spinner.Success("Xdebug enabled")
		}
	}

	// 6. Run build if requested
	if startBuild {
		spinner = ui.NewSpinner("Running build process")
		spinner.Start()

		if err := runBuildProcess(projectDir); err != nil {
			spinner.Stop()
			ui.Warning("Build process failed: %v", err))
		} else {
			spinner.Success("Build completed")
		}
	}

	// 7. Run project hooks unless skipped
	if !startSkipHooks {
		if err := runProjectHooks(projectDir, "post-start"); err != nil {
			ui.Warning("Post-start hooks failed: %v", err))
		}
	}

	// 8. Show environment info
	status, err = ddev.GetStatus(projectDir)
	if err != nil {
		return fmt.Errorf("failed to get status after start: %w", err)
	}

	fmt.Println()
	ui.Success("Environment started successfully!")
	fmt.Println()
	displayEnvironmentInfo(status)

	return nil
}

// validateStartPrerequisites checks all prerequisites before starting
func validateStartPrerequisites(projectDir string) error {
	// Check Docker
	if !system.IsDockerAvailable() {
		return errors.NewWithSolution(
			"Docker is not installed",
			"Docker is required to run DDEV",
			errors.Solution{
				Description: "Install Docker Desktop",
				Steps: []string{
					"Visit: https://www.docker.com/products/docker-desktop",
					"Download and install Docker Desktop for your platform",
					"Start Docker Desktop",
					"Run: stax doctor to verify installation",
				},
			},
		)
	}

	if !system.IsDockerRunning() {
		return errors.NewWithSolution(
			"Docker is not running",
			"Docker Desktop needs to be started",
			errors.Solution{
				Description: "Start Docker Desktop",
				Steps: []string{
					"Launch Docker Desktop application",
					"Wait for Docker to fully start (whale icon in menu bar/system tray)",
					"Run: docker info to verify",
					"Then retry: stax start",
				},
			},
		)
	}

	// Check DDEV
	if !ddev.IsInstalled() {
		return errors.NewWithSolution(
			"DDEV is not installed",
			"DDEV is required to manage the development environment",
			errors.Solution{
				Description: "Install DDEV",
				Steps: []string{
					"Visit: https://ddev.readthedocs.io/en/stable/users/install/",
					"Follow installation instructions for your platform",
					"Run: stax doctor to verify installation",
				},
			},
		)
	}

	// Check DDEV configuration
	if !ddev.IsConfigured(projectDir) {
		return errors.NewWithSolution(
			"DDEV is not configured",
			fmt.Sprintf("No DDEV configuration found in %s", projectDir),
			errors.Solution{
				Description: "Initialize the project",
				Steps: []string{
					"Run: stax init",
					"Or manually configure DDEV: ddev config --project-type=wordpress",
					"Then retry: stax start",
				},
			},
		)
	}

	return nil
}

// waitForServices waits for all services to be ready
func waitForServices(projectDir string) error {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		status, err := ddev.GetStatus(projectDir)
		if err == nil && status.Running && status.Healthy {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("services did not become healthy within timeout")
}

// runBuildProcess runs the build process (if configured)
func runBuildProcess(projectDir string) error {
	// Check if package.json exists
	packageJSON := fmt.Sprintf("%s/package.json", projectDir)
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found")
	}

	// Run npm install first
	if err := ddev.Exec(projectDir, "npm", "install"); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	// Run build
	if err := ddev.Exec(projectDir, "npm", "run", "build"); err != nil {
		return fmt.Errorf("npm run build failed: %w", err)
	}

	return nil
}

// runProjectHooks runs project-specific hooks
func runProjectHooks(projectDir, hookName string) error {
	// Check for .stax/hooks directory
	hooksDir := fmt.Sprintf("%s/.stax/hooks", projectDir)
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return nil // No hooks directory, skip silently
	}

	// Check for specific hook script
	hookScript := fmt.Sprintf("%s/%s.sh", hooksDir, hookName)
	if _, err := os.Stat(hookScript); os.IsNotExist(err) {
		return nil // No hook script, skip silently
	}

	// Run the hook
	ui.Info("Running %s hook...", hookName))
	if err := ddev.Exec(projectDir, "bash", hookScript); err != nil {
		return fmt.Errorf("hook failed: %w", err)
	}

	return nil
}

// displayEnvironmentInfo shows environment information after start
func displayEnvironmentInfo(status *ddev.ProjectInfo) {
	fmt.Println("Environment Information:")
	fmt.Printf("  Project:     %s\n", status.Name)
	fmt.Printf("  Status:      %s\n", status.Status)
	fmt.Printf("  Primary URL: %s\n", status.PrimaryURL)

	if len(status.URLs) > 1 {
		fmt.Println("  Other URLs:")
		for _, url := range status.URLs[1:] {
			fmt.Printf("    - %s\n", url)
		}
	}

	fmt.Printf("  PHP Version: %s\n", status.PHPVersion)
	fmt.Printf("  Database:    %s %s\n", status.DatabaseType, status.DatabaseVersion)

	if status.MailhogURL != "" {
		fmt.Printf("  Mailhog:     %s\n", status.MailhogURL)
	}

	fmt.Println()
	ui.Info("Useful commands:")
	ui.Info("  stax status       - Show detailed status")
	ui.Info("  stax db pull      - Pull database from WPEngine")
	ui.Info("  stax stop         - Stop environment")
	ui.Info("  stax doctor       - Run health checks")
}
