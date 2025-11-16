package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	versionShort bool
	versionJSON  bool
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information and feature status",
	Long: `Display comprehensive version information including:
  - Version number, git commit, and build date
  - List of implemented features
  - Partially implemented features with workarounds
  - Planned features under development
  - Known limitations

This command helps you understand what features are available
in your current version of Stax.`,
	Example: `  # Show full version information
  stax version

  # Show short version only
  stax version --short`,
	RunE: runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "show only version number")
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "output as JSON")
}

func runVersion(cmd *cobra.Command, args []string) error {
	// Short version output
	if versionShort {
		fmt.Printf("stax version %s\n", Version)
		return nil
	}

	// JSON output
	if versionJSON {
		return outputVersionJSON()
	}

	// Full version output
	return outputVersionFull()
}

func outputVersionFull() error {
	ui.PrintHeader(fmt.Sprintf("Stax CLI Version %s", Version))
	fmt.Println()

	// Version information
	ui.Section("Version Information")
	fmt.Printf("  Version:     %s\n", Version)
	fmt.Printf("  Git Commit:  %s\n", GitCommit)
	fmt.Printf("  Build Date:  %s\n", BuildDate)
	fmt.Println()

	// Implemented Features
	ui.Section("✓ Implemented Features")
	implementedFeatures := []string{
		"WPEngine site discovery and listing",
		"Database pull/sync from WPEngine",
		"File synchronization from WPEngine",
		"DDEV environment management (start/stop/restart)",
		"Environment status monitoring",
		"Project build system (Composer + NPM)",
		"Development mode with file watching",
		"PHP code quality checks (PHPCS/PHPCBF)",
		"Project configuration validation",
		"System diagnostics and health checks",
	}
	for _, feature := range implementedFeatures {
		fmt.Printf("  ✓ %s\n", feature)
	}
	fmt.Println()

	// Partially Implemented
	ui.Section("⚠ Partially Implemented")
	partialFeatures := []struct {
		name       string
		workaround string
	}{
		{
			name:       "Credential storage via macOS Keychain",
			workaround: "Keychain unavailable in Homebrew builds (use file or env storage)",
		},
	}
	for _, feature := range partialFeatures {
		fmt.Printf("  ⚠ %s\n", feature.name)
		fmt.Printf("    Workaround: %s\n", feature.workaround)
	}
	fmt.Println()

	// Coming Soon
	ui.Section("🚧 Coming Soon")
	plannedFeatures := []string{
		"Full project initialization from templates",
		"Multi-provider support (Kinsta, Pantheon, etc.)",
		"Advanced configuration management",
		"Database push to WPEngine",
		"File push to WPEngine",
		"Automated deployment workflows",
		"Plugin and theme scaffolding",
	}
	for _, feature := range plannedFeatures {
		fmt.Printf("  🚧 %s\n", feature)
	}
	fmt.Println()

	// Known Limitations
	ui.Section("Known Limitations")
	limitations := []string{
		"Keychain integration requires CGO (not available in Homebrew builds)",
		"WPEngine is currently the only supported hosting provider",
		"Multisite support is limited to subdomain and subdirectory modes",
		"Remote media proxying requires manual BunnyCDN configuration",
	}
	for _, limitation := range limitations {
		fmt.Printf("  • %s\n", limitation)
	}
	fmt.Println()

	// System Information
	ui.Section("System Information")
	fmt.Println("  For detailed system diagnostics, run: stax doctor")
	fmt.Println("  For credential status, run: stax setup --check")
	fmt.Println()

	// Footer
	ui.Info("For more information, visit: https://github.com/Firecrown-Media/stax-public")

	return nil
}

func outputVersionJSON() error {
	versionInfo := map[string]interface{}{
		"version":    Version,
		"git_commit": GitCommit,
		"build_date": BuildDate,
		"features": map[string]interface{}{
			"implemented": []string{
				"wpengine_site_discovery",
				"database_pull",
				"file_sync",
				"ddev_management",
				"status_monitoring",
				"build_system",
				"development_mode",
				"code_quality",
				"config_validation",
				"diagnostics",
			},
			"partial": []map[string]string{
				{
					"name":       "keychain_storage",
					"workaround": "use file or environment variable storage",
				},
			},
			"planned": []string{
				"project_init",
				"multi_provider",
				"config_management",
				"database_push",
				"file_push",
				"deployment",
				"scaffolding",
			},
		},
		"limitations": []string{
			"keychain_requires_cgo",
			"wpengine_only",
			"limited_multisite",
			"manual_bunnycdn",
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(versionInfo)
}
