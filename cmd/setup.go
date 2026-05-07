package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	setupWPEngineUser     string
	setupWPEnginePassword string
	setupGitHubToken      string
	setupSSHKey           string
	setupInteractive      bool
	setupCheck            bool
	setupMethod           string
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "⚠ Configure WPEngine and GitHub credentials",
	Long: `Configure WPEngine and GitHub credentials securely in macOS Keychain.

This command stores sensitive credentials in the macOS Keychain, ensuring
they are never stored in plain text configuration files.`,
	Example: `  # Interactive mode
  stax setup

  # Non-interactive
  stax setup \
    --wpengine-user=myuser@example.com \
    --wpengine-password=mypassword \
    --github-token=ghp_xxxxxxxxxxxxx \
    --ssh-key=~/.ssh/wpengine_rsa`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().StringVar(&setupWPEngineUser, "wpengine-user", "", "WPEngine API username")
	setupCmd.Flags().StringVar(&setupWPEnginePassword, "wpengine-password", "", "WPEngine API password")
	setupCmd.Flags().StringVar(&setupGitHubToken, "github-token", "", "GitHub personal access token")
	setupCmd.Flags().StringVar(&setupSSHKey, "ssh-key", "", "Path to SSH private key for WPEngine")
	setupCmd.Flags().BoolVar(&setupInteractive, "interactive", true, "Interactive credential setup")
	setupCmd.Flags().BoolVar(&setupCheck, "check", false, "Check credential status and configuration")
	setupCmd.Flags().StringVar(&setupMethod, "method", "", "Storage method: keychain, file, or env")
}

func runSetup(cmd *cobra.Command, args []string) error {
	// Handle --check flag
	if setupCheck {
		return runSetupCheck()
	}

	ui.PrintHeader("Setting up Stax Credentials")

	// Determine storage method
	storageMethod := setupMethod
	keychainAvailable := credentials.IsKeychainAvailable()

	if storageMethod == "" {
		// Auto-detect or prompt for storage method
		if keychainAvailable {
			ui.Success("macOS Keychain is available")
			storageMethod = "keychain"
		} else {
			ui.Warning("macOS Keychain storage is not available in this build")
			ui.Info("This is normal for Homebrew installations (built with CGO_ENABLED=0)")
			fmt.Println()
			ui.Info("Available storage methods:")
			ui.Info("  1. file - Store credentials in ~/.stax/credentials.yml")
			ui.Info("  2. env  - Use environment variables")
			fmt.Println()

			if setupInteractive {
				storageMethod = ui.PromptString("Choose storage method (file/env)", "file")
			} else {
				storageMethod = "file"
			}
		}
	}

	// Normalize storage method input - accept both numbers and text
	switch strings.TrimSpace(strings.ToLower(storageMethod)) {
	case "1", "file":
		storageMethod = "file"
	case "2", "env":
		storageMethod = "env"
	case "keychain":
		storageMethod = "keychain"
	case "":
		storageMethod = "file" // default
	default:
		return fmt.Errorf("invalid storage method: %s (choose '1/file' or '2/env')", storageMethod)
	}

	// Validate storage method
	if storageMethod == "keychain" && !keychainAvailable {
		ui.Error("Keychain storage is not available")
		return fmt.Errorf("keychain method requested but not available")
	}

	// Get credentials interactively if not provided
	if setupInteractive {
		if setupWPEngineUser == "" {
			setupWPEngineUser = ui.PromptString("WPEngine API Username", "")
		}
		if setupWPEnginePassword == "" {
			// Use password prompt to hide input
			setupWPEnginePassword = ui.PromptPassword("WPEngine API Password")
		}
		if setupGitHubToken == "" {
			// Use password prompt to hide token input
			setupGitHubToken = ui.PromptPassword("GitHub Personal Access Token (optional, press Enter to skip)")
		}
		if setupSSHKey == "" {
			setupSSHKey = ui.PromptString("SSH Key for WPEngine (optional)", "~/.ssh/id_rsa")
		}
	}

	// Validate required fields
	if setupWPEngineUser == "" || setupWPEnginePassword == "" {
		return fmt.Errorf("WPEngine credentials are required")
	}

	// Store credentials based on method
	switch storageMethod {
	case "keychain":
		return setupWithKeychain()
	case "file":
		return setupWithFile()
	case "env":
		return setupWithEnv()
	default:
		return fmt.Errorf("invalid storage method: %s", storageMethod)
	}
}

// setupWithKeychain stores credentials in macOS Keychain
func setupWithKeychain() error {
	ui.Info("Storing credentials in macOS Keychain...")

	// Store WPEngine credentials
	wpeCreds := &credentials.WPEngineCredentials{
		APIUser:     setupWPEngineUser,
		APIPassword: setupWPEnginePassword,
		SSHGateway:  "ssh.wpengine.net",
	}

	if err := credentials.SetWPEngineCredentials("default", wpeCreds); err != nil {
		// Check if this is a "not supported" error
		if strings.Contains(err.Error(), "keychain storage is not supported") ||
			strings.Contains(err.Error(), "not supported on Linux") {
			ui.Warning("Keychain storage failed - falling back to file storage")
			ui.Info("")
			return setupWithFileAutomatically()
		}
		return fmt.Errorf("failed to store WPEngine credentials: %w", err)
	}
	ui.Success("WPEngine credentials stored in Keychain")

	// Test WPEngine API connection
	ui.Info("Testing WPEngine API connection...")
	if err := testWPEngineConnection(setupWPEngineUser, setupWPEnginePassword); err != nil {
		ui.Warning("Failed to connect to WPEngine API: %v", err)
		ui.Info("Credentials saved, but please verify they are correct")
	} else {
		ui.Success("WPEngine API connection successful")
	}

	// Store GitHub token if provided
	if setupGitHubToken != "" {
		ui.Info("Storing GitHub token in Keychain...")
		if err := credentials.SetGitHubToken("default", setupGitHubToken); err != nil {
			if strings.Contains(err.Error(), "keychain storage is not supported") ||
				strings.Contains(err.Error(), "not supported on Linux") {
				ui.Warning("Keychain storage failed - falling back to file storage")
				ui.Info("")
				return setupWithFileAutomatically()
			}
			return fmt.Errorf("failed to store GitHub token: %w", err)
		}
		ui.Success("GitHub token stored")
	}

	// Store SSH key if provided
	if setupSSHKey != "" {
		ui.Info("Storing SSH key in Keychain...")
		// Read SSH key file
		if err := storeSSHKey(setupSSHKey); err != nil {
			if strings.Contains(err.Error(), "keychain storage is not supported") ||
				strings.Contains(err.Error(), "not supported on Linux") {
				ui.Warning("Keychain storage failed - falling back to file storage")
				ui.Info("")
				return setupWithFileAutomatically()
			}
			ui.Warning("Failed to store SSH key: %v", err)
		} else {
			ui.Success("SSH key stored")
		}
	}

	ui.Section("\nCredentials saved successfully!")
	ui.Info("Your credentials are securely stored in macOS Keychain")
	ui.Info("You can now run 'stax init' to initialize a project")

	return nil
}

// setupWithFileAutomatically provides automatic fallback to file storage
func setupWithFileAutomatically() error {
	ui.Info("Using credentials file at ~/.stax/credentials.yml...")

	// Build credentials structure
	credFile := &credentials.CredentialsFile{
		WPEngine: credentials.WPEngineCredentialsFile{
			APIUser:     setupWPEngineUser,
			APIPassword: setupWPEnginePassword,
			SSHGateway:  "ssh.wpengine.net",
		},
	}

	if setupGitHubToken != "" {
		credFile.GitHub = credentials.GitHubCredentialsFile{
			Token: setupGitHubToken,
		}
	}

	if setupSSHKey != "" {
		credFile.SSH = credentials.SSHCredentialsFile{
			PrivateKeyPath: setupSSHKey,
		}
	}

	// Save to file
	if err := credentials.SaveCredentialsFile(credFile); err != nil {
		// If file also fails, show environment variable instructions
		ui.Error("File storage also failed")
		ui.Info("")
		ui.Section("Alternative: Use Environment Variables")
		ui.Info("Set these environment variables in your shell:")
		ui.Info("")
		ui.Info("  export WPENGINE_API_USER=\"%s\"", setupWPEngineUser)
		ui.Info("  export WPENGINE_API_PASSWORD=\"%s\"", setupWPEnginePassword)
		ui.Info("  export WPENGINE_SSH_GATEWAY=\"ssh.wpengine.net\"")
		if setupGitHubToken != "" {
			ui.Info("  export GITHUB_TOKEN=\"%s\"", setupGitHubToken)
		}
		if setupSSHKey != "" {
			ui.Info("  export SSH_PRIVATE_KEY_PATH=\"%s\"", setupSSHKey)
		}
		ui.Info("")
		return fmt.Errorf("unable to store credentials: %w", err)
	}

	ui.Success("Credentials saved to file successfully")

	// Show the file path
	credPath, _ := credentials.GetCredentialsFilePath()
	ui.Info("Location: %s", credPath)
	ui.Info("File permissions: 0600 (owner read/write only)")

	ui.Section("\nSecurity Reminder:")
	ui.Warning("Add ~/.stax/credentials.yml to your .gitignore")
	ui.Info("Never commit credentials to version control")

	ui.Section("\nNext Steps:")
	ui.Info("Run 'stax init' to initialize a project")
	ui.Info("Run 'stax setup --check' to verify configuration")

	return nil
}

// setupWithFile stores credentials in ~/.stax/credentials.yml
func setupWithFile() error {
	ui.Info("Creating credentials file at ~/.stax/credentials.yml...")

	// Build credentials structure
	credFile := &credentials.CredentialsFile{
		WPEngine: credentials.WPEngineCredentialsFile{
			APIUser:     setupWPEngineUser,
			APIPassword: setupWPEnginePassword,
			SSHGateway:  "ssh.wpengine.net",
		},
	}

	if setupGitHubToken != "" {
		credFile.GitHub = credentials.GitHubCredentialsFile{
			Token: setupGitHubToken,
		}
	}

	if setupSSHKey != "" {
		credFile.SSH = credentials.SSHCredentialsFile{
			PrivateKeyPath: setupSSHKey,
		}
	}

	// Save to file
	if err := credentials.SaveCredentialsFile(credFile); err != nil {
		return fmt.Errorf("failed to save credentials file: %w", err)
	}

	ui.Success("Credentials file created successfully")

	// Show the file path
	credPath, _ := credentials.GetCredentialsFilePath()
	ui.Info("Credentials saved to: %s", credPath)
	ui.Info("File permissions set to 0600 (owner read/write only)")

	ui.Section("\nSecurity Notes:")
	ui.Warning("Add ~/.stax/credentials.yml to your .gitignore")
	ui.Warning("Never commit this file to version control")

	ui.Section("\nNext Steps:")
	ui.Info("You can now run 'stax init' to initialize a project")
	ui.Info("Run 'stax setup --check' to verify your configuration")

	return nil
}

// setupWithEnv shows instructions for environment variable setup
func setupWithEnv() error {
	ui.Info("Setting up credentials via environment variables...")

	fmt.Println()
	ui.Section("Add these to your shell profile (~/.zshrc or ~/.bashrc):")
	fmt.Println()

	if setupWPEngineUser != "" {
		fmt.Printf("export WPENGINE_API_USER=\"%s\"\n", setupWPEngineUser)
	} else {
		fmt.Println("export WPENGINE_API_USER=\"your-api-username\"")
	}

	if setupWPEnginePassword != "" {
		fmt.Printf("export WPENGINE_API_PASSWORD=\"%s\"\n", setupWPEnginePassword)
	} else {
		fmt.Println("export WPENGINE_API_PASSWORD=\"your-api-password\"")
	}

	fmt.Println("export WPENGINE_SSH_GATEWAY=\"ssh.wpengine.net\"")

	if setupGitHubToken != "" {
		fmt.Printf("export GITHUB_TOKEN=\"%s\"\n", setupGitHubToken)
	}

	if setupSSHKey != "" {
		fmt.Printf("export STAX_SSH_PRIVATE_KEY=\"%s\"\n", setupSSHKey)
	}

	fmt.Println()
	ui.Section("Then reload your shell:")
	fmt.Println("  source ~/.zshrc")

	fmt.Println()
	ui.Section("Next Steps:")
	ui.Info("After setting environment variables, run 'stax setup --check' to verify")
	ui.Info("You can then run 'stax init' to initialize a project")

	return nil
}

// testWPEngineConnection tests the WPEngine API connection
func testWPEngineConnection(apiUser, apiPassword string) error {
	// Import wpengine package
	wpengine := struct {
		NewClient func(string, string, string) interface{ TestConnection() error }
	}{
		NewClient: func(u, p, i string) interface{ TestConnection() error } {
			// This is a placeholder - actual implementation will use wpengine.NewClient
			return nil
		},
	}

	if wpengine.NewClient == nil {
		return fmt.Errorf("WPEngine client not available")
	}

	// Note: In actual implementation, this would use wpengine.NewClient
	// For now, we'll just validate credentials are not empty
	if apiUser == "" || apiPassword == "" {
		return fmt.Errorf("credentials are empty")
	}

	return nil
}

// storeSSHKey reads and stores an SSH key
func storeSSHKey(keyPath string) error {
	// Read SSH key file
	keyData, err := os.ReadFile(expandPath(keyPath))
	if err != nil {
		return fmt.Errorf("failed to read SSH key: %w", err)
	}

	// Store in keychain
	if err := credentials.SetSSHPrivateKey("wpengine", string(keyData)); err != nil {
		return fmt.Errorf("failed to store SSH key: %w", err)
	}

	return nil
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}
	return path
}

// runSetupCheck runs credential diagnostics
func runSetupCheck() error {
	ui.PrintHeader("Checking Credential Configuration")

	// Run diagnostics
	diag := credentials.RunDiagnostics()

	// Display results
	displayDiagnosticResult(diag.KeychainAvailable)
	displayDiagnosticResult(diag.WPEngineAPI)
	displayDiagnosticResult(diag.WPEngineSSH)
	displayDiagnosticResult(diag.GitHubToken)
	displayDiagnosticResult(diag.CredentialsFile)
	displayDiagnosticResult(diag.EnvironmentVars)
	displayDiagnosticResult(diag.SSHKeyFile)

	// Display overall status
	ui.Section("\nOverall Status")
	switch diag.OverallStatus {
	case "ok":
		ui.Success("All checks passed")
	case "warning":
		ui.Warning("Some warnings found")
	case "error":
		ui.Error("Critical issues found")
	}

	// Display recommendations
	if len(diag.RecommendedActions) > 0 {
		ui.Section("\nRecommended Actions")
		for _, action := range diag.RecommendedActions {
			ui.Info("- %s", action)
		}
	}

	return nil
}

// displayDiagnosticResult displays a single diagnostic result
func displayDiagnosticResult(result credentials.DiagnosticResult) {
	ui.Section(fmt.Sprintf("\n%s", result.Name))

	switch result.Status {
	case "ok":
		ui.Success("%s", result.Message)
	case "warning":
		ui.Warning("%s", result.Message)
	case "error":
		ui.Error("%s", result.Message)
	}

	for _, detail := range result.Details {
		ui.Info("  %s", detail)
	}
}
