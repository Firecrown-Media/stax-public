package credentials

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSSHPrivateKeyWithFallback(t *testing.T) {
	// Save original env vars
	origStaxSSH := os.Getenv("STAX_SSH_PRIVATE_KEY")
	origWPESSH := os.Getenv("WPENGINE_SSH_KEY")
	origHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("STAX_SSH_PRIVATE_KEY", origStaxSSH)
		os.Setenv("WPENGINE_SSH_KEY", origWPESSH)
		os.Setenv("HOME", origHome)
	}()

	// Clean environment
	os.Unsetenv("STAX_SSH_PRIVATE_KEY")
	os.Unsetenv("WPENGINE_SSH_KEY")

	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		cleanup     func(t *testing.T, path string)
		expectError bool
	}{
		{
			name: "fallback to environment variable STAX_SSH_PRIVATE_KEY",
			setup: func(t *testing.T) string {
				// Create temp SSH key
				tmpDir := t.TempDir()
				keyPath := filepath.Join(tmpDir, "id_rsa")
				keyContent := "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----\n"
				if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
					t.Fatalf("failed to create test key: %v", err)
				}
				os.Setenv("STAX_SSH_PRIVATE_KEY", keyPath)
				return keyPath
			},
			cleanup: func(t *testing.T, path string) {
				os.Unsetenv("STAX_SSH_PRIVATE_KEY")
			},
			expectError: false,
		},
		{
			name: "fallback to environment variable WPENGINE_SSH_KEY",
			setup: func(t *testing.T) string {
				// Create temp SSH key
				tmpDir := t.TempDir()
				keyPath := filepath.Join(tmpDir, "id_rsa")
				keyContent := "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----\n"
				if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
					t.Fatalf("failed to create test key: %v", err)
				}
				os.Setenv("WPENGINE_SSH_KEY", keyPath)
				return keyPath
			},
			cleanup: func(t *testing.T, path string) {
				os.Unsetenv("WPENGINE_SSH_KEY")
			},
			expectError: false,
		},
		{
			name: "no key found returns error with tried locations",
			setup: func(t *testing.T) string {
				// Set HOME to a temp directory with no SSH keys to ensure isolation
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				// Create empty .ssh directory to ensure no default keys are found
				os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0700)
				return tmpDir
			},
			cleanup: func(t *testing.T, path string) {
				// HOME will be restored when test completes via defer in test setup
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPath := tt.setup(t)
			defer tt.cleanup(t, keyPath)

			key, err := GetSSHPrivateKeyWithFallback("test")

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				// Check that error contains tried locations
				if keyErr, ok := err.(*SSHKeyNotFoundError); ok {
					if len(keyErr.Tried) == 0 {
						t.Errorf("expected Tried locations, got empty")
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if key == "" {
					t.Errorf("expected key, got empty string")
				}
			}
		})
	}
}

func TestGetWPEngineCredentialsWithFallback(t *testing.T) {
	// Save original env vars
	origAPIUser := os.Getenv("WPENGINE_API_USER")
	origAPIPassword := os.Getenv("WPENGINE_API_PASSWORD")
	origHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("WPENGINE_API_USER", origAPIUser)
		os.Setenv("WPENGINE_API_PASSWORD", origAPIPassword)
		os.Setenv("HOME", origHome)
	}()

	// Clean environment
	os.Unsetenv("WPENGINE_API_USER")
	os.Unsetenv("WPENGINE_API_PASSWORD")

	tests := []struct {
		name        string
		setup       func(t *testing.T)
		cleanup     func(t *testing.T)
		expectError bool
	}{
		{
			name: "fallback to environment variables",
			setup: func(t *testing.T) {
				os.Setenv("WPENGINE_API_USER", "testuser")
				os.Setenv("WPENGINE_API_PASSWORD", "testpass")
			},
			cleanup: func(t *testing.T) {
				os.Unsetenv("WPENGINE_API_USER")
				os.Unsetenv("WPENGINE_API_PASSWORD")
			},
			expectError: false,
		},
		{
			name: "no credentials found returns error with tried locations",
			setup: func(t *testing.T) {
				// Set HOME to a temp directory with no credentials file to ensure isolation
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				// Create empty .stax directory to ensure no credentials file is found
				os.MkdirAll(filepath.Join(tmpDir, ".stax"), 0700)
			},
			cleanup: func(t *testing.T) {
				// HOME will be restored when test completes via defer in test setup
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			defer tt.cleanup(t)

			creds, err := GetWPEngineCredentialsWithFallback("testinstall")

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				// Check that error contains tried locations
				if credErr, ok := err.(*CredentialsNotFoundError); ok {
					if len(credErr.Tried) == 0 {
						t.Errorf("expected Tried locations, got empty")
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if creds == nil {
					t.Errorf("expected credentials, got nil")
				}
				if creds != nil && creds.APIUser == "" {
					t.Errorf("expected APIUser, got empty string")
				}
			}
		})
	}
}

func TestGetGitHubTokenWithFallback(t *testing.T) {
	// Save original env var
	origToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", origToken)

	// Clean environment
	os.Unsetenv("GITHUB_TOKEN")

	tests := []struct {
		name        string
		setup       func(t *testing.T)
		cleanup     func(t *testing.T)
		expectError bool
	}{
		{
			name: "fallback to environment variable",
			setup: func(t *testing.T) {
				os.Setenv("GITHUB_TOKEN", "ghp_test_token")
			},
			cleanup: func(t *testing.T) {
				os.Unsetenv("GITHUB_TOKEN")
			},
			expectError: false,
		},
		{
			name: "no token found returns error with tried locations",
			setup: func(t *testing.T) {
				// No setup - clean environment
			},
			cleanup:     func(t *testing.T) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			defer tt.cleanup(t)

			token, err := GetGitHubTokenWithFallback("testorg")

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				// Check that error contains tried locations
				if tokenErr, ok := err.(*GitHubTokenNotFoundError); ok {
					if len(tokenErr.Tried) == 0 {
						t.Errorf("expected Tried locations, got empty")
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if token == "" {
					t.Errorf("expected token, got empty string")
				}
			}
		})
	}
}

func TestValidateSSHKey(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "valid RSA private key",
			content:  "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----\n",
			expected: true,
		},
		{
			name:     "valid OpenSSH private key",
			content:  "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAACm...\n-----END OPENSSH PRIVATE KEY-----\n",
			expected: true,
		},
		{
			name:     "valid EC private key",
			content:  "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEII...\n-----END EC PRIVATE KEY-----\n",
			expected: true,
		},
		{
			name:     "invalid file - not a private key",
			content:  "This is not a private key",
			expected: false,
		},
		{
			name:     "empty file",
			content:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			keyPath := filepath.Join(tmpDir, "test_key")
			if err := os.WriteFile(keyPath, []byte(tt.content), 0600); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			result := validateSSHKey(keyPath)
			if result != tt.expected {
				t.Errorf("validateSSHKey() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expand tilde to home",
			input:    "~",
			expected: home,
		},
		{
			name:     "expand tilde with path",
			input:    "~/.ssh/id_rsa",
			expected: filepath.Join(home, ".ssh/id_rsa"),
		},
		{
			name:     "no expansion for absolute path",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "no expansion for relative path",
			input:    "relative/path",
			expected: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCredentialsNotFoundError(t *testing.T) {
	err := &CredentialsNotFoundError{
		Install: "testinstall",
		Tried: []string{
			"macOS Keychain",
			"Environment variables",
			"Credentials file",
		},
		LastErr: os.ErrNotExist,
	}

	errMsg := err.Error()

	// Check that error message contains install name
	if !contains(errMsg, "testinstall") {
		t.Errorf("error message should contain install name")
	}

	// Check that error message contains tried locations
	if !contains(errMsg, "macOS Keychain") {
		t.Errorf("error message should contain tried location: macOS Keychain")
	}

	// Check that error message contains last error
	if !contains(errMsg, os.ErrNotExist.Error()) {
		t.Errorf("error message should contain last error")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
