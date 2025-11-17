package credentials

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// KeychainUnavailableError represents keychain unavailability
type KeychainUnavailableError struct {
	Operation string
}

func (e *KeychainUnavailableError) Error() string {
	return fmt.Sprintf("keychain storage not available for %s operation", e.Operation)
}

// IsKeychainUnavailable checks if error is keychain unavailable
func IsKeychainUnavailable(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*KeychainUnavailableError)
	return ok
}

// IsKeychainAvailable checks if keychain storage is available
func IsKeychainAvailable() bool {
	// Try a test operation to see if keychain works
	err := SetWPEngineCredentials("__test__", &WPEngineCredentials{
		APIUser: "test",
	})

	// Clean up test entry if it succeeded
	if err == nil {
		// Ignore cleanup errors - test entry may have already been deleted
		_ = DeleteWPEngineCredentials("__test__")
		return true
	}

	// Check if it's the "not supported" error
	return !IsKeychainUnavailable(err)
}

// GetCredentialsStorageInstructions returns helpful instructions
func GetCredentialsStorageInstructions() string {
	return `
Keychain storage is not available in this build of Stax.

You have two options to store your credentials:

OPTION 1: Environment Variables (Recommended for CI/CD)
--------------------------------------------------------
Add these to your shell profile (~/.zshrc or ~/.bashrc):

    export WPENGINE_API_USER="your-api-username"
    export WPENGINE_API_PASSWORD="your-api-password"
    export WPENGINE_SSH_GATEWAY="ssh.wpengine.net"
    export GITHUB_TOKEN="ghp_your_token_here"

Then reload your shell:
    source ~/.zshrc

OPTION 2: Config File (Recommended for Development)
---------------------------------------------------
Create ~/.stax/credentials.yml with:

    wpengine:
      api_user: "your-api-username"
      api_password: "your-api-password"
      ssh_gateway: "ssh.wpengine.net"

    github:
      token: "ghp_your_token_here"

Secure the file:
    chmod 600 ~/.stax/credentials.yml

SECURITY NOTE:
--------------
- Add ~/.stax/credentials.yml to your .gitignore
- Never commit credentials to version control
- For maximum security, build Stax from source with CGO enabled

To check which storage method is active:
    stax setup --check
`
}

// CredentialsFile represents the credentials file structure
type CredentialsFile struct {
	WPEngine WPEngineCredentialsFile `yaml:"wpengine"`
	GitHub   GitHubCredentialsFile   `yaml:"github"`
	SSH      SSHCredentialsFile      `yaml:"ssh"`
}

// WPEngineCredentialsFile for file storage
type WPEngineCredentialsFile struct {
	APIUser     string `yaml:"api_user"`
	APIPassword string `yaml:"api_password"`
	SSHUser     string `yaml:"ssh_user"`
	SSHGateway  string `yaml:"ssh_gateway"`
}

// GitHubCredentialsFile for file storage
type GitHubCredentialsFile struct {
	Token string `yaml:"token"`
}

// SSHCredentialsFile for file storage
type SSHCredentialsFile struct {
	PrivateKeyPath string `yaml:"private_key_path"`
}

// GetCredentialsFilePath returns the path to credentials file
func GetCredentialsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".stax", "credentials.yml"), nil
}

// LoadCredentialsFile loads credentials from file
func LoadCredentialsFile() (*CredentialsFile, error) {
	path, err := GetCredentialsFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds CredentialsFile
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	return &creds, nil
}

// SaveCredentialsFile saves credentials to file
func SaveCredentialsFile(creds *CredentialsFile) error {
	path, err := GetCredentialsFilePath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	data, err := yaml.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// getEnvOrDefault gets environment variable or returns default
func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
