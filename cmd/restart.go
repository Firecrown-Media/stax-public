package cmd

import (
	"fmt"

	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	restartBuild  bool
	restartXdebug bool
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "✓ Restart the DDEV environment",
	Long: `Restart the DDEV environment (stop followed by start).

This is equivalent to running 'stax stop' followed by 'stax start',
but is more efficient as it uses DDEV's built-in restart capability.`,
	Example: `  # Basic restart
  stax restart

  # Restart with build
  stax restart --build

  # Restart with Xdebug
  stax restart --xdebug`,
	RunE: runRestart,
}

func init() {
	rootCmd.AddCommand(restartCmd)

	restartCmd.Flags().BoolVar(&restartBuild, "build", false, "run build process after restart")
	restartCmd.Flags().BoolVar(&restartXdebug, "xdebug", false, "enable Xdebug")
}

func runRestart(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Restarting DDEV Environment")

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

	// TODO: Run ddev restart
	// TODO: Enable Xdebug if requested
	// TODO: Run build process if requested

	// Check if project is configured
	if !ddev.IsConfigured(projectDir) {
		return fmt.Errorf("DDEV is not configured for this project. Run: stax init")
	}

	// Restart DDEV
	spinner := ui.NewSpinner("Restarting DDEV containers")
	spinner.Start()

	if err := ddev.Restart(projectDir); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to restart environment: %w", err)
	}

	spinner.Success("DDEV containers restarted")

	// Wait for services
	ui.Info("Waiting for services to be ready...")
	if err := waitForServices(projectDir); err != nil {
		ui.Warning("Services may not be fully ready yet")
	}

	// Enable Xdebug if requested
	if restartXdebug {
		spinner = ui.NewSpinner("Enabling Xdebug")
		spinner.Start()

		if err := ddev.EnableXdebug(projectDir); err != nil {
			spinner.Stop()
			ui.Warning("Failed to enable Xdebug: %v", err))
		} else {
			spinner.Success("Xdebug enabled")
		}
	}

	// Run build if requested
	if restartBuild {
		spinner = ui.NewSpinner("Running build process")
		spinner.Start()

		if err := runBuildProcess(projectDir); err != nil {
			spinner.Stop()
			ui.Warning("Build process failed: %v", err))
		} else {
			spinner.Success("Build completed")
		}
	}

	// Get and display status
	status, err := ddev.GetStatus(projectDir)
	if err != nil {
		ui.Warning("Could not retrieve environment status")
	} else {
		fmt.Println()
		ui.Success("Environment restarted successfully!")
		fmt.Println()
		fmt.Printf("Primary URL: %s\n", status.PrimaryURL)
	}

	return nil
}
