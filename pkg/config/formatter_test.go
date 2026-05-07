package config

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFormatConfig(t *testing.T) {
	cfg := &Config{
		Version: 2,
		Project: ProjectConfig{
			Name: "test-project",
			Type: "wordpress-multisite",
			Mode: "subdomain",
		},
		Provider: "wpengine",
		ProviderConfig: map[string]any{
			"install":     "testinstall",
			"environment": "production",
		},
		DDEV: DDEVConfig{
			PHPVersion:   "8.1",
			MySQLVersion: "8.0",
		},
	}

	tests := []struct {
		name        string
		format      string
		expectError bool
		validate    func(t *testing.T, output string)
	}{
		{
			name:   "json format",
			format: "json",
			validate: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("failed to parse JSON: %v", err)
				}
				if result["Version"] != float64(2) {
					t.Errorf("expected version 2, got %v", result["Version"])
				}
			},
		},
		{
			name:   "yaml format",
			format: "yaml",
			validate: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("failed to parse YAML: %v", err)
				}
				if v, _ := result["version"].(int); v != 2 {
					t.Errorf("expected version 2, got %v", result["version"])
				}
			},
		},
		{
			name:   "pretty format",
			format: "pretty",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "Project Configuration") {
					t.Error("expected 'Project Configuration' in pretty output")
				}
				if !strings.Contains(output, "test-project") {
					t.Error("expected project name in pretty output")
				}
			},
		},
		{
			name:   "empty format defaults to pretty",
			format: "",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "Project Configuration") {
					t.Error("expected 'Project Configuration' in pretty output")
				}
			},
		},
		{
			name:        "invalid format",
			format:      "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := FormatConfig(cfg, tt.format)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestFormatPretty(t *testing.T) {
	cfg := &Config{
		Project: ProjectConfig{
			Name:        "test-project",
			Type:        "wordpress-multisite",
			Mode:        "subdomain",
			Description: "Test project description",
		},
		Provider: "wpengine",
		ProviderConfig: map[string]any{
			"install":      "testinstall",
			"environment":  "production",
			"account_name": "testaccount",
			"ssh_gateway":  "ssh.wpengine.net",
		},
		Network: NetworkConfig{
			Domain: "test.local",
			Title:  "Test Network",
		},
		DDEV: DDEVConfig{
			PHPVersion:    "8.1",
			MySQLVersion:  "8.0",
			MySQLType:     "mysql",
			WebserverType: "nginx-fpm",
			NodeJSVersion: "20",
			XdebugEnabled: false,
		},
		WordPress: WordPressConfig{
			Version:     "latest",
			Locale:      "en_US",
			TablePrefix: "wp_",
		},
	}

	output := FormatPretty(cfg)

	// Check for section headers
	sections := []string{
		"Project Configuration",
		"Provider Configuration",
		"Network Configuration",
		"DDEV Configuration",
		"WordPress Configuration",
	}

	for _, section := range sections {
		if !strings.Contains(output, section) {
			t.Errorf("expected section '%s' in output", section)
		}
	}

	// Check for specific values
	values := []string{
		"test-project",
		"wordpress-multisite",
		"subdomain",
		"testinstall",
		"production",
		"test.local",
		"8.1",
		"8.0",
		"latest",
		"en_US",
	}

	for _, value := range values {
		if !strings.Contains(output, value) {
			t.Errorf("expected value '%s' in output", value)
		}
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "string value",
			value:    "test",
			expected: "test",
		},
		{
			name:     "int value",
			value:    42,
			expected: "42",
		},
		{
			name:     "bool true",
			value:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			value:    false,
			expected: "false",
		},
		{
			name:     "nil value",
			value:    nil,
			expected: "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValue(tt.value)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMaskSensitiveValues(t *testing.T) {
	cfg := &Config{
		Credentials: CredentialsConfig{
			WPEngine: CredentialRef{
				KeychainService: "stax",
				KeychainAccount: "wpengine-api-key-12345678",
			},
			GitHub: CredentialRef{
				KeychainService: "stax",
				KeychainAccount: "github-token-87654321",
			},
		},
	}

	masked := MaskSensitiveValues(cfg)

	// Check that sensitive values are masked
	if !strings.Contains(masked.Credentials.WPEngine.KeychainAccount, "****") {
		t.Error("expected WPEngine keychain account to be masked")
	}

	if !strings.Contains(masked.Credentials.GitHub.KeychainAccount, "****") {
		t.Error("expected GitHub keychain account to be masked")
	}

	// Verify original is not modified
	if strings.Contains(cfg.Credentials.WPEngine.KeychainAccount, "****") {
		t.Error("original config should not be modified")
	}
}
