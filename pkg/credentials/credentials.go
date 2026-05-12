package credentials

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetSSHPrivateKeyWithFallback attempts to get SSH private key from multiple sources
// in the following order:
// 1. Keychain (if available)
// 2. Environment variable (STAX_SSH_PRIVATE_KEY or WPENGINE_SSH_KEY)
// 3. Credentials file (~/.stax/credentials.yml)
// 4. Default SSH key locations (~/.ssh/id_rsa, ~/.ssh/id_ed25519)
func GetSSHPrivateKeyWithFallback(service string) (string, error) {
	var lastErr error
	var tried []string

	// 1. Try keychain first (if available)
	if IsKeychainAvailable() {
		tried = append(tried, "macOS Keychain")
		key, err := GetSSHPrivateKey(service)
		if err == nil && key != "" {
			return key, nil
		}
		lastErr = err
	}

	// 2. Try environment variables
	tried = append(tried, "Environment variable STAX_SSH_PRIVATE_KEY")
	if keyPath := os.Getenv("STAX_SSH_PRIVATE_KEY"); keyPath != "" {
		expanded := expandPath(keyPath)
		if validateSSHKey(expanded) {
			key, err := os.ReadFile(expanded)
			if err == nil {
				return string(key), nil
			}
			lastErr = err
		}
	}

	tried = append(tried, "Environment variable WPENGINE_SSH_KEY")
	if keyPath := os.Getenv("WPENGINE_SSH_KEY"); keyPath != "" {
		expanded := expandPath(keyPath)
		if validateSSHKey(expanded) {
			key, err := os.ReadFile(expanded)
			if err == nil {
				return string(key), nil
			}
			lastErr = err
		}
	}

	// 3. Try credentials file
	tried = append(tried, "Credentials file ~/.stax/credentials.yml")
	credFile, err := LoadCredentialsFile()
	if err == nil && credFile.SSH.PrivateKeyPath != "" {
		keyPath := expandPath(credFile.SSH.PrivateKeyPath)
		if validateSSHKey(keyPath) {
			key, err := os.ReadFile(keyPath)
			if err == nil {
				return string(key), nil
			}
			lastErr = err
		}
	} else if err != nil {
		lastErr = err
	}

	// 4. Try default SSH key locations
	defaultPaths := []string{
		"~/.ssh/id_rsa",
		"~/.ssh/id_ed25519",
		"~/.ssh/id_ecdsa",
	}

	for _, path := range defaultPaths {
		tried = append(tried, fmt.Sprintf("Default SSH key location %s", path))
		expandedPath := expandPath(path)
		if validateSSHKey(expandedPath) {
			key, err := os.ReadFile(expandedPath)
			if err == nil {
				return string(key), nil
			}
			lastErr = err
		}
	}

	// No key found in any location
	return "", &SSHKeyNotFoundError{
		Tried:   tried,
		LastErr: lastErr,
	}
}

// GetWPEngineCredentialsWithFallback attempts to get WPEngine credentials from multiple sources
// in the following order:
// 1. Keychain (if available)
// 2. Environment variables
// 3. Credentials file (~/.stax/credentials.yml)
func GetWPEngineCredentialsWithFallback(install string) (*WPEngineCredentials, error) {
	var lastErr error
	var tried []string

	// 1. Try keychain first (if available)
	if IsKeychainAvailable() {
		tried = append(tried, "macOS Keychain")
		creds, err := GetWPEngineCredentials(install)
		if err == nil && creds != nil {
			return creds, nil
		}
		lastErr = err
	}

	// 2. Try environment variables
	tried = append(tried, "Environment variables (WPENGINE_API_USER, WPENGINE_API_PASSWORD)")
	apiUser := os.Getenv("WPENGINE_API_USER")
	apiPassword := os.Getenv("WPENGINE_API_PASSWORD")
	sshGateway := getEnvOrDefault("WPENGINE_SSH_GATEWAY", DefaultSSHGateway)
	sshUser := getEnvOrDefault("WPENGINE_SSH_USER", install)

	if apiUser != "" && apiPassword != "" {
		return &WPEngineCredentials{
			APIUser:     apiUser,
			APIPassword: apiPassword,
			SSHUser:     sshUser,
			SSHGateway:  sshGateway,
		}, nil
	}

	// 3. Try credentials file
	tried = append(tried, "Credentials file ~/.stax/credentials.yml")
	credFile, err := LoadCredentialsFile()
	if err == nil && credFile.WPEngine.APIUser != "" && credFile.WPEngine.APIPassword != "" {
		return &WPEngineCredentials{
			APIUser:     credFile.WPEngine.APIUser,
			APIPassword: credFile.WPEngine.APIPassword,
			SSHUser:     credFile.WPEngine.SSHUser,
			SSHGateway:  credFile.WPEngine.SSHGateway,
		}, nil
	} else if err != nil {
		lastErr = err
	}

	// No credentials found in any location
	return nil, &CredentialsNotFoundError{
		Install: install,
		Tried:   tried,
		LastErr: lastErr,
	}
}

// GetGitHubTokenWithFallback attempts to get GitHub token from multiple sources
// in the following order:
// 1. Keychain (if available)
// 2. Environment variable (GITHUB_TOKEN)
// 3. Credentials file (~/.stax/credentials.yml)
func GetGitHubTokenWithFallback(organization string) (string, error) {
	var lastErr error
	var tried []string

	// 1. Try keychain first (if available)
	if IsKeychainAvailable() {
		tried = append(tried, "macOS Keychain")
		token, err := GetGitHubToken(organization)
		if err == nil && token != "" {
			return token, nil
		}
		lastErr = err
	}

	// 2. Try environment variable
	tried = append(tried, "Environment variable GITHUB_TOKEN")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// 3. Try credentials file
	tried = append(tried, "Credentials file ~/.stax/credentials.yml")
	credFile, err := LoadCredentialsFile()
	if err == nil && credFile.GitHub.Token != "" {
		return credFile.GitHub.Token, nil
	} else if err != nil {
		lastErr = err
	}

	// No token found in any location
	return "", &GitHubTokenNotFoundError{
		Organization: organization,
		Tried:        tried,
		LastErr:      lastErr,
	}
}

// validateSSHKey checks if an SSH key file exists and is readable
func validateSSHKey(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return false
	}

	// Check if file is readable
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first few bytes to check if it looks like a private key
	header := make([]byte, 100)
	n, err := file.Read(header)
	if err != nil || n == 0 {
		return false
	}

	// Check for common SSH key headers
	headerStr := string(header[:n])
	return strings.Contains(headerStr, "PRIVATE KEY") ||
		strings.Contains(headerStr, "BEGIN RSA PRIVATE KEY") ||
		strings.Contains(headerStr, "BEGIN OPENSSH PRIVATE KEY") ||
		strings.Contains(headerStr, "BEGIN EC PRIVATE KEY") ||
		strings.Contains(headerStr, "BEGIN DSA PRIVATE KEY")
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			if len(path) == 1 {
				return home
			}
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// DefaultSSHGateway is the default WPEngine SSH gateway
const DefaultSSHGateway = "ssh.wpengine.net"

// Custom error types with detailed information

// SSHKeyNotFoundError represents an error when SSH key cannot be found
type SSHKeyNotFoundError struct {
	Tried   []string
	LastErr error
}

func (e *SSHKeyNotFoundError) Error() string {
	msg := "SSH private key not found in any location"
	if len(e.Tried) > 0 {
		msg += "\n\nTried:"
		for _, location := range e.Tried {
			msg += fmt.Sprintf("\n  - %s", location)
		}
	}
	if e.LastErr != nil {
		msg += fmt.Sprintf("\n\nLast error: %v", e.LastErr)
	}
	return msg
}

// CredentialsNotFoundError represents an error when WPEngine credentials cannot be found
type CredentialsNotFoundError struct {
	Install string
	Tried   []string
	LastErr error
}

func (e *CredentialsNotFoundError) Error() string {
	msg := fmt.Sprintf("WPEngine credentials not found for install '%s'", e.Install)
	if len(e.Tried) > 0 {
		msg += "\n\nTried:"
		for _, location := range e.Tried {
			msg += fmt.Sprintf("\n  - %s", location)
		}
	}
	if e.LastErr != nil {
		msg += fmt.Sprintf("\n\nLast error: %v", e.LastErr)
	}
	return msg
}

// GitHubTokenNotFoundError represents an error when GitHub token cannot be found
type GitHubTokenNotFoundError struct {
	Organization string
	Tried        []string
	LastErr      error
}

func (e *GitHubTokenNotFoundError) Error() string {
	msg := fmt.Sprintf("GitHub token not found for organization '%s'", e.Organization)
	if len(e.Tried) > 0 {
		msg += "\n\nTried:"
		for _, location := range e.Tried {
			msg += fmt.Sprintf("\n  - %s", location)
		}
	}
	if e.LastErr != nil {
		msg += fmt.Sprintf("\n\nLast error: %v", e.LastErr)
	}
	return msg
}
