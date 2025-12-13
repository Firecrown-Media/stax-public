package diagnostics

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/firecrown-media/stax/pkg/credentials"
)

// CheckCredentials checks if credentials are properly configured
func CheckCredentials(projectPath string) CheckResult {
	diag := credentials.RunDiagnostics()

	// Analyze diagnostic results
	if diag.OverallStatus == "error" {
		return CheckResult{
			Name:       "Credentials",
			Category:   "Credentials",
			Status:     StatusFail,
			Message:    "Credential configuration has errors",
			Suggestion: "Run: stax setup --check for details",
		}
	}

	if diag.OverallStatus == "warning" {
		return CheckResult{
			Name:       "Credentials",
			Category:   "Credentials",
			Status:     StatusWarning,
			Message:    "Credential configuration has warnings",
			Suggestion: "Run: stax setup --check for details",
		}
	}

	return CheckResult{
		Name:     "Credentials",
		Category: "Credentials",
		Status:   StatusPass,
		Message:  "All credentials are properly configured",
	}
}

// CheckSSHKey checks if SSH key is available and properly configured
func CheckSSHKey() CheckResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return CheckResult{
			Name:       "SSH Key",
			Category:   "Credentials",
			Status:     StatusFail,
			Message:    "Cannot determine home directory",
			Suggestion: "Check your system configuration",
		}
	}

	// Check for common SSH key types
	keyPaths := []string{
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
	}

	var foundKey string
	var keyInfo os.FileInfo
	for _, keyPath := range keyPaths {
		info, err := os.Stat(keyPath)
		if err == nil {
			foundKey = keyPath
			keyInfo = info
			break
		}
	}

	if foundKey == "" {
		return CheckResult{
			Name:       "SSH Key",
			Category:   "Credentials",
			Status:     StatusWarning,
			Message:    "No SSH key found",
			Suggestion: "Generate SSH key: ssh-keygen -t ed25519 -C \"your_email@example.com\"",
			Details: map[string]string{
				"checked_paths": strings.Join(keyPaths, ", "),
			},
		}
	}

	// Check permissions
	mode := keyInfo.Mode()
	if mode.Perm()&0077 != 0 {
		return CheckResult{
			Name:       "SSH Key",
			Category:   "Credentials",
			Status:     StatusWarning,
			Message:    "SSH key has insecure permissions",
			Suggestion: fmt.Sprintf("Fix permissions: chmod 600 %s", foundKey),
			Details: map[string]string{
				"path":        foundKey,
				"permissions": fmt.Sprintf("%o", mode.Perm()),
			},
		}
	}

	// Check if public key exists
	pubKeyPath := foundKey + ".pub"
	if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
		return CheckResult{
			Name:       "SSH Key",
			Category:   "Credentials",
			Status:     StatusWarning,
			Message:    "SSH public key not found",
			Suggestion: fmt.Sprintf("Generate public key: ssh-keygen -y -f %s > %s", foundKey, pubKeyPath),
			Details: map[string]string{
				"private_key": foundKey,
			},
		}
	}

	return CheckResult{
		Name:     "SSH Key",
		Category: "Credentials",
		Status:   StatusPass,
		Message:  "SSH key found and properly configured",
		Details: map[string]string{
			"path":        foundKey,
			"permissions": "0600",
			"public_key":  pubKeyPath,
		},
	}
}

// CheckGitHubToken checks if GitHub token is configured
func CheckGitHubToken() CheckResult {
	token, err := credentials.GetGitHubToken("default")
	if err != nil || token == "" {
		return CheckResult{
			Name:       "GitHub Token",
			Category:   "Credentials",
			Status:     StatusSkip,
			Message:    "GitHub token not configured (optional)",
			Suggestion: "Configure token if needed for private repos: stax setup github",
		}
	}

	return CheckResult{
		Name:     "GitHub Token",
		Category: "Credentials",
		Status:   StatusPass,
		Message:  "GitHub token configured",
		Details: map[string]string{
			"token_length": fmt.Sprintf("%d characters", len(token)),
		},
	}
}
