package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/prompts"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/firecrown-media/stax/pkg/wpengine"
	"github.com/spf13/cobra"
)

var (
	wpengineListEnvironment string
	wpengineListJSON        bool
	wpengineInfoJSON        bool
)

// wpengineCmd represents the global wpengine command group
var wpengineCmd = &cobra.Command{
	Use:   "wpengine",
	Short: "Global WPEngine discovery and management commands",
	Long: `Global WPEngine discovery and management commands that work without .stax.yml.

These commands allow you to:
  - List all available WPEngine installations
  - View detailed information about specific installations
  - Interactively select and configure installations

These commands work globally and do not require a .stax.yml configuration file.`,
}

// wpengineListCmd lists all WPEngine installations
var wpengineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all WPEngine installations",
	Long: `List all WPEngine installations available to your account.

This command connects to the WPEngine API and retrieves all installations
you have access to, displaying key information about each one.

By default, output is displayed in a table format. Use --json for machine-readable output.`,
	Example: `  # List all installations
  stax wpengine list

  # List only production environments
  stax wpengine list --environment=production

  # Output as JSON
  stax wpengine list --json`,
	RunE: runWPEngineList,
}

// wpengineInfoCmd shows detailed information about a specific installation
var wpengineInfoCmd = &cobra.Command{
	Use:   "info <install>",
	Short: "Get detailed installation information",
	Long: `Get detailed information about a specific WPEngine installation.

This command retrieves comprehensive details about an installation including:
  - Name and environment
  - PHP and MySQL versions
  - Disk usage statistics
  - Configured domains
  - SSH connection details
  - Available environments`,
	Example: `  # Get installation information
  stax wpengine info mywordpresssite

  # Output as JSON
  stax wpengine info mywordpresssite --json`,
	Args: cobra.ExactArgs(1),
	RunE: runWPEngineInfo,
}

// wpengineSelectCmd provides interactive installation selection
var wpengineSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Interactive installation selector",
	Long: `Interactive installation selector and configuration generator.

This command provides an interactive wizard that:
  1. Lists all available WPEngine installations
  2. Allows you to select an installation
  3. Prompts for project details
  4. Creates or updates .stax.yml in the current directory
  5. Optionally initializes DDEV configuration

This is a quick way to set up Stax for an existing WPEngine site.`,
	Example: `  # Start interactive selection wizard
  stax wpengine select`,
	RunE: runWPEngineSelect,
}

func init() {
	rootCmd.AddCommand(wpengineCmd)

	// Add subcommands
	wpengineCmd.AddCommand(wpengineListCmd)
	wpengineCmd.AddCommand(wpengineInfoCmd)
	wpengineCmd.AddCommand(wpengineSelectCmd)

	// List command flags
	wpengineListCmd.Flags().StringVar(&wpengineListEnvironment, "environment", "", "filter by environment (production, staging, development)")
	wpengineListCmd.Flags().BoolVar(&wpengineListJSON, "json", false, "output as JSON")

	// Info command flags
	wpengineInfoCmd.Flags().BoolVar(&wpengineInfoJSON, "json", false, "output as JSON")
}

// runWPEngineList lists all WPEngine installations
func runWPEngineList(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("WPEngine Installations")

	// Get credentials
	creds, err := credentials.GetWPEngineCredentialsWithFallback("")
	if err != nil {
		return handleCredentialsError(err)
	}

	// Create WPEngine client
	client := wpengine.NewClient(creds.APIUser, creds.APIPassword, "")

	// Show spinner while fetching
	spinner := ui.NewSpinner("Fetching installations from WPEngine API...")
	spinner.Start()

	// Fetch installations
	installs, err := client.ListInstalls()
	spinner.Stop()

	if err != nil {
		ui.Error("Failed to fetch installations")
		return fmt.Errorf("failed to list installations: %w", err)
	}

	if len(installs) == 0 {
		ui.Warning("No installations found")
		return nil
	}

	// Filter by environment if specified
	if wpengineListEnvironment != "" {
		filtered := []wpengine.Install{}
		for _, install := range installs {
			if strings.EqualFold(install.Environment, wpengineListEnvironment) {
				filtered = append(filtered, install)
			}
		}
		installs = filtered

		if len(installs) == 0 {
			ui.Warning("No installations found for environment: %s", wpengineListEnvironment)
			return nil
		}
	}

	// Output results
	if wpengineListJSON {
		return outputInstallsJSON(installs)
	}

	return outputInstallsTable(installs)
}

// runWPEngineInfo shows detailed information about an installation
func runWPEngineInfo(cmd *cobra.Command, args []string) error {
	installName := args[0]

	ui.PrintHeader(fmt.Sprintf("WPEngine Installation: %s", installName))

	// Get credentials
	creds, err := credentials.GetWPEngineCredentialsWithFallback("")
	if err != nil {
		return handleCredentialsError(err)
	}

	// Create WPEngine client
	client := wpengine.NewClient(creds.APIUser, creds.APIPassword, installName)

	// Show spinner while fetching
	spinner := ui.NewSpinner("Fetching installation details...")
	spinner.Start()

	// Fetch installation details
	details, err := client.GetInstallByName(installName)
	spinner.Stop()

	if err != nil {
		ui.Error("Failed to fetch installation details")
		return fmt.Errorf("failed to get installation info: %w", err)
	}

	// Output results
	if wpengineInfoJSON {
		return outputInstallDetailsJSON(details)
	}

	return outputInstallDetailsTable(details, creds)
}

// runWPEngineSelect runs interactive installation selection
func runWPEngineSelect(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("WPEngine Installation Selector")

	// Get credentials
	creds, err := credentials.GetWPEngineCredentialsWithFallback("")
	if err != nil {
		return handleCredentialsError(err)
	}

	// Create WPEngine client
	client := wpengine.NewClient(creds.APIUser, creds.APIPassword, "")

	// Show spinner while fetching
	spinner := ui.NewSpinner("Fetching installations from WPEngine API...")
	spinner.Start()

	// Fetch installations
	installs, err := client.ListInstalls()
	spinner.Stop()

	if err != nil {
		ui.Error("Failed to fetch installations")
		return fmt.Errorf("failed to list installations: %w", err)
	}

	if len(installs) == 0 {
		ui.Warning("No installations found")
		return nil
	}

	ui.Success("Found %d installation(s)", len(installs))
	fmt.Println()

	// Step 1: Select installation
	ui.Section("Step 1: Select Installation")

	installOptions := make([]string, len(installs))
	for i, install := range installs {
		installOptions[i] = fmt.Sprintf("%s (%s) - PHP %s", install.Name, install.Environment, install.PHPVersion)
	}

	selectedIdx, _, err := prompts.SafePromptSelect("Select an installation:", installOptions, 0)
	if err != nil {
		return err
	}

	selectedInstall := installs[selectedIdx]
	ui.Info("Selected: %s", selectedInstall.Name)

	// Step 2: Get project details
	ui.Section("Step 2: Project Configuration")

	// Get current directory name as default project name
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defaultName := currentDir[strings.LastIndex(currentDir, "/")+1:]

	projectName, err := prompts.SafePromptInput("Project name", defaultName, false)
	if err != nil {
		return err
	}

	// Step 3: Select environment
	ui.Section("Step 3: Environment Selection")

	environment, err := prompts.SafeEnvironmentPrompt(selectedInstall.Environment)
	if err != nil {
		return err
	}

	// Step 4: Create or update .stax.yml
	ui.Section("Step 4: Configuration")

	configPath := ".stax.yml"
	var cfg *config.Config

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		overwrite, err := prompts.SafePromptConfirm(".stax.yml already exists. Overwrite?", false)
		if err != nil {
			return err
		}
		if !overwrite {
			ui.Info("Configuration cancelled")
			return nil
		}
		// Load existing config to preserve settings
		cfg, err = config.Load(configPath, "")
		if err != nil {
			// If load fails, start with defaults
			cfg = config.Defaults()
		}
	} else {
		cfg = config.Defaults()
	}

	// Update config with selected values
	cfg.Project.Name = projectName
	cfg.WPEngine.Install = selectedInstall.Name
	cfg.WPEngine.Environment = environment

	// Set PHP version from install details if available
	if selectedInstall.PHPVersion != "" {
		cfg.DDEV.PHPVersion = selectedInstall.PHPVersion
	}

	// Save configuration
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	ui.Success("Created .stax.yml")

	// Step 5: Optionally initialize DDEV
	ui.Section("Step 5: DDEV Setup")

	initDDEV, err := prompts.SafePromptConfirm("Initialize DDEV configuration?", true)
	if err != nil {
		return err
	}

	if initDDEV {
		ui.Info("To complete DDEV setup, run: stax init --from-ddev")
	}

	// Print success summary
	fmt.Println()
	ui.PrintHeader("Configuration Complete!")
	fmt.Println()

	ui.Success("Created:")
	ui.Info("  - .stax.yml")
	fmt.Println()

	ui.Section("Next Steps:")
	if initDDEV {
		ui.ProgressMsg("stax init --from-ddev  - Complete DDEV initialization")
	} else {
		ui.ProgressMsg("stax start             - Start DDEV environment")
	}
	ui.ProgressMsg("stax db pull           - Pull database from WPEngine")
	ui.ProgressMsg("stax files pull        - Pull files from WPEngine")
	ui.ProgressMsg("stax status            - View environment status")

	return nil
}

// outputInstallsTable outputs installations in table format
func outputInstallsTable(installs []wpengine.Install) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "NAME\tENVIRONMENT\tPHP VERSION\tPRIMARY DOMAIN")
	fmt.Fprintln(w, "----\t-----------\t-----------\t--------------")

	// Data rows
	for _, install := range installs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			install.Name,
			install.Environment,
			install.PHPVersion,
			install.PrimaryDomain,
		)
	}

	fmt.Println()
	ui.Success("Found %d installation(s)", len(installs))

	return nil
}

// outputInstallsJSON outputs installations in JSON format
func outputInstallsJSON(installs []wpengine.Install) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(installs)
}

// outputInstallDetailsTable outputs installation details in table format
func outputInstallDetailsTable(details *wpengine.InstallDetails, creds *credentials.WPEngineCredentials) error {
	// Basic information
	ui.Section("Basic Information")
	fmt.Printf("  Name:               %s\n", details.Name)
	fmt.Printf("  Environment:        %s\n", details.Environment)
	fmt.Printf("  Primary Domain:     %s\n", details.PrimaryDomain)
	fmt.Println()

	// Versions
	ui.Section("Software Versions")
	fmt.Printf("  PHP Version:        %s\n", details.PHPVersion)
	fmt.Printf("  MySQL Version:      %s\n", details.MySQLVersion)
	fmt.Printf("  WordPress Version:  %s\n", details.WordPressVersion)
	fmt.Println()

	// Disk usage
	ui.Section("Disk Usage")
	usedGB := float64(details.DiskUsage.Used) / (1024 * 1024 * 1024)
	totalGB := float64(details.DiskUsage.Total) / (1024 * 1024 * 1024)
	usagePercent := 0.0
	if details.DiskUsage.Total > 0 {
		usagePercent = (float64(details.DiskUsage.Used) / float64(details.DiskUsage.Total)) * 100
	}
	fmt.Printf("  Used:               %.2f GB\n", usedGB)
	fmt.Printf("  Total:              %.2f GB\n", totalGB)
	fmt.Printf("  Usage:              %.1f%%\n", usagePercent)
	fmt.Println()

	// Domains
	if len(details.Domains) > 0 {
		ui.Section("Configured Domains")
		for _, domain := range details.Domains {
			fmt.Printf("  - %s\n", domain)
		}
		fmt.Println()
	}

	// SSH connection details
	ui.Section("SSH Connection")
	sshUser := creds.SSHUser
	if sshUser == "" {
		sshUser = details.Name
	}
	sshGateway := creds.SSHGateway
	if sshGateway == "" {
		sshGateway = "ssh.wpengine.net"
	}
	fmt.Printf("  User:               %s\n", sshUser)
	fmt.Printf("  Gateway:            %s\n", sshGateway)
	fmt.Printf("  Connection String:  ssh %s@%s\n", sshUser, sshGateway)
	fmt.Println()

	// Available environments
	ui.Section("Available Environments")
	ui.Info("  To pull from different environments, update .stax.yml:")
	fmt.Println()
	fmt.Println("  wpengine:")
	fmt.Printf("    install: %s\n", details.Name)
	fmt.Println("    environment: production  # or staging, development")

	return nil
}

// outputInstallDetailsJSON outputs installation details in JSON format
func outputInstallDetailsJSON(details *wpengine.InstallDetails) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(details)
}

// handleCredentialsError provides helpful error messages for credentials issues
func handleCredentialsError(err error) error {
	ui.Error("WPEngine credentials not found")
	fmt.Println()

	ui.Info("Stax requires WPEngine API credentials to access your installations.")
	fmt.Println()

	ui.Section("To configure credentials, run:")
	ui.Info("  stax setup")
	fmt.Println()

	ui.Section("Or set environment variables:")
	ui.Info("  export WPENGINE_API_USER=\"your-api-username\"")
	ui.Info("  export WPENGINE_API_PASSWORD=\"your-api-password\"")
	fmt.Println()

	return err
}
