package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	statusJSON bool
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "✓ Show environment status",
	Long: `Show detailed status information about the DDEV environment,
including container health, URLs, configuration, database info, and more.`,
	Aliases: []string{"s"},
	Example: `  # Show status
  stax status

  # Show status as JSON
  stax status --json`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output as JSON")
}

func runStatus(cmd *cobra.Command, args []string) error {
	projectDir := getProjectDir()

	// Explicitly check for .stax.yml file existence
	configPath := filepath.Join(projectDir, ".stax.yml")
	configFileExists := false
	if _, err := os.Stat(configPath); err == nil {
		configFileExists = true
	}

	// Check if config was successfully loaded
	hasStaxConfig := (cfg != nil)

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

	// Improved messaging based on file existence vs. config loading
	if configFileExists && !hasStaxConfig {
		ui.Warning(".stax.yml exists but could not be loaded")
		ui.Info("Run 'stax validate' to check for configuration errors")
		ui.Info("Using DDEV configuration for now")
		fmt.Println()
	} else if !configFileExists {
		ui.Warning("Using DDEV configuration only (no .stax.yml found)")
		ui.Info("Run 'stax init' to enable Stax features like WPEngine sync")
		fmt.Println()
	}

	// Get status
	status, err := ddev.GetStatus(projectDir)
	if err != nil {
		return fmt.Errorf("failed to get environment status: %w", err)
	}

	ui.PrintHeader("Environment Status")
	fmt.Println()

	// Project Information
	ui.Section("Project Information")
	fmt.Printf("  Name:        %s\n", status.Name)
	fmt.Printf("  Type:        %s\n", status.Type)
	fmt.Printf("  Location:    %s\n", status.AppRoot)
	fmt.Printf("  Status:      %s\n", getStatusIndicator(status))
	fmt.Println()

	// URLs
	ui.Section("URLs")
	fmt.Printf("  Primary:     %s\n", status.PrimaryURL)
	if len(status.URLs) > 1 {
		fmt.Println("  Additional:")
		for _, url := range status.URLs[1:] {
			fmt.Printf("    - %s\n", url)
		}
	}
	if status.MailhogURL != "" {
		fmt.Printf("  Mailhog:     %s\n", status.MailhogURL)
	}
	fmt.Println()

	// Container Status
	ui.Section("Containers")
	fmt.Printf("  Web:         %s\n", getContainerStatus(status.Running))
	fmt.Printf("  Database:    %s\n", getContainerStatus(status.Running))
	fmt.Printf("  Router:      %s\n", getContainerStatus(status.Running))
	fmt.Println()

	// Configuration
	ui.Section("Configuration")
	fmt.Printf("  PHP Version: %s\n", status.PHPVersion)
	fmt.Printf("  Database:    %s %s\n", status.DatabaseType, status.DatabaseVersion)
	fmt.Printf("  Webserver:   %s\n", status.Webserver)
	if status.XdebugEnabled {
		fmt.Println("  Xdebug:      ✓ Enabled")
	} else {
		fmt.Println("  Xdebug:      ✗ Disabled")
	}
	fmt.Println()

	// Router Status
	if status.Router != "" {
		ui.Section("Router")
		fmt.Printf("  Status:      %s\n", status.RouterStatus)
		if status.RouterHTTPPort != "" {
			fmt.Printf("  HTTP Port:   %s\n", status.RouterHTTPPort)
		}
		if status.RouterHTTPSPort != "" {
			fmt.Printf("  HTTPS Port:  %s\n", status.RouterHTTPSPort)
		}
		fmt.Println()
	}

	// Project Details
	if cfg != nil {
		ui.Section("Stax Configuration")
		fmt.Printf("  Provider:    wpengine\n")
		if cfg.WPEngine.Install != "" {
			fmt.Printf("  Install:     %s\n", cfg.WPEngine.Install)
		}
		if cfg.WPEngine.Environment != "" {
			fmt.Printf("  Environment: %s\n", cfg.WPEngine.Environment)
		}
		fmt.Println()
	}

	// Quick actions
	if status.Running {
		ui.Info("Quick commands:")
		ui.Info("  stax stop      - Stop environment")
		ui.Info("  stax restart   - Restart environment")
		ui.Info("  stax db pull   - Pull database from WPEngine")
	} else {
		ui.Info("Environment is stopped. Run: stax start")
	}

	return nil
}

// getStatusIndicator returns a colored status indicator
func getStatusIndicator(status *ddev.ProjectInfo) string {
	if !status.Running {
		return "⚫ Stopped"
	}
	if status.Healthy {
		return "🟢 Running (Healthy)"
	}
	return "🟡 Running (Starting up...)"
}

// getContainerStatus returns container status indicator
func getContainerStatus(running bool) string {
	if running {
		return "🟢 Running"
	}
	return "⚫ Stopped"
}
