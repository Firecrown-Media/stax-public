package config

import (
	"testing"
)

func TestGetValueByPath(t *testing.T) {
	cfg := &Config{
		Project: ProjectConfig{
			Name: "test-project",
			Type: "wordpress-multisite",
			Mode: "subdomain",
		},
		Provider: "wpengine",
		ProviderConfig: map[string]any{
			"install":     "testinstall",
			"environment": "production",
			"ssh_gateway": "ssh.wpengine.net",
		},
		DDEV: DDEVConfig{
			PHPVersion:    "8.1",
			MySQLVersion:  "8.0",
			WebserverType: "nginx-fpm",
			XdebugEnabled: false,
		},
		WordPress: WordPressConfig{
			Version:     "latest",
			Locale:      "en_US",
			TablePrefix: "wp_",
		},
	}

	tests := []struct {
		name        string
		path        string
		expected    interface{}
		expectError bool
	}{
		{
			name:     "project name",
			path:     "project.name",
			expected: "test-project",
		},
		{
			name:     "project type",
			path:     "project.type",
			expected: "wordpress-multisite",
		},
		{
			name:     "provider",
			path:     "provider",
			expected: "wpengine",
		},
		{
			name:     "ddev webserver type",
			path:     "ddev.webserver_type",
			expected: "nginx-fpm",
		},
		{
			name:     "ddev php version",
			path:     "ddev.php_version",
			expected: "8.1",
		},
		{
			name:     "ddev mysql version",
			path:     "ddev.mysql_version",
			expected: "8.0",
		},
		{
			name:     "ddev xdebug enabled",
			path:     "ddev.xdebug_enabled",
			expected: false,
		},
		{
			name:     "wordpress version",
			path:     "wordpress.version",
			expected: "latest",
		},
		{
			name:     "wordpress locale",
			path:     "wordpress.locale",
			expected: "en_US",
		},
		{
			name:        "invalid path",
			path:        "invalid.path",
			expectError: true,
		},
		{
			name:        "invalid nested path",
			path:        "project.invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := GetValueByPath(cfg, tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for path %s, got nil", tt.path)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for path %s: %v", tt.path, err)
				return
			}

			if value != tt.expected {
				t.Errorf("for path %s: expected %v, got %v", tt.path, tt.expected, value)
			}
		})
	}
}

func TestSetValueByPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		value       string
		expectError bool
	}{
		{
			name:  "set string value",
			path:  "project.name",
			value: "new-project",
		},
		{
			name:  "set string value nested",
			path:  "ddev.webserver_type",
			value: "apache-fpm",
		},
		{
			name:  "set boolean true",
			path:  "ddev.xdebug_enabled",
			value: "true",
		},
		{
			name:  "set boolean false",
			path:  "ddev.xdebug_enabled",
			value: "false",
		},
		{
			name:        "set invalid boolean",
			path:        "ddev.xdebug_enabled",
			value:       "not-a-bool",
			expectError: true,
		},
		{
			name:        "set invalid path",
			path:        "invalid.path",
			value:       "value",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Project: ProjectConfig{
					Name: "test-project",
				},
				ProviderConfig: map[string]any{
					"environment": "production",
				},
				DDEV: DDEVConfig{
					XdebugEnabled: false,
				},
			}

			err := SetValueByPath(cfg, tt.path, tt.value)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for path %s, got nil", tt.path)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for path %s: %v", tt.path, err)
				return
			}

			// Verify the value was set
			newValue, err := GetValueByPath(cfg, tt.path)
			if err != nil {
				t.Errorf("failed to get value after setting: %v", err)
				return
			}

			// For boolean values, compare as bool
			if tt.path == "ddev.xdebug_enabled" {
				expectedBool := tt.value == "true"
				if newValue != expectedBool {
					t.Errorf("expected %v, got %v", expectedBool, newValue)
				}
			} else {
				if newValue != tt.value {
					t.Errorf("expected %v, got %v", tt.value, newValue)
				}
			}
		})
	}
}

func TestToFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"name", "Name"},
		{"project", "Project"},
		{"wpengine", "WPEngine"},
		{"wordpress", "WordPress"},
		{"ddev", "DDEV"},
		{"php_version", "PHPVersion"},
		{"mysql_version", "MySQLVersion"},
		{"mysql_type", "MySQLType"},
		{"ssh_gateway", "SSHGateway"},
		{"xdebug_enabled", "XdebugEnabled"},
		{"nodejs_version", "NodeJSVersion"},
		{"table_prefix", "TablePrefix"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toFieldName(tt.input)
			if result != tt.expected {
				t.Errorf("toFieldName(%s): expected %s, got %s", tt.input, tt.expected, result)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	cfg := &Config{
		Project: ProjectConfig{
			Name: "test",
		},
		ProviderConfig: map[string]any{
			"install": "test",
		},
	}

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name: "valid path",
			path: "project.name",
		},
		{
			name: "valid nested path",
			path: "project.name",
		},
		{
			name:        "invalid path",
			path:        "invalid.path",
			expectError: true,
		},
		{
			name:        "invalid nested path",
			path:        "project.invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(cfg, tt.path)

			if tt.expectError && err == nil {
				t.Errorf("expected error for path %s", tt.path)
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for path %s: %v", tt.path, err)
			}
		})
	}
}
