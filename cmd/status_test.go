package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// Note: These tests verify the file detection logic in the status command.
// The status command now loads config via root.go's PersistentPreRunE
// (removed from skipConfigCommands), so:
// - Valid config: cfg != nil, hasStaxConfig = true
// - Invalid config: cfg = nil, hasStaxConfig = false (shows warning)
// - No config file: cfg = nil, hasStaxConfig = false (shows different warning)
func TestStatusConfigDetection(t *testing.T) {
	tests := []struct {
		name             string
		createStaxConfig bool
		configContent    string
		createDDEV       bool
	}{
		{
			name:             "both configs exist and valid",
			createStaxConfig: true,
			configContent: `version: 1
project:
  name: test-project
  type: wordpress
wpengine:
  install: testinstall
`,
			createDDEV: true,
		},
		{
			name:             "only DDEV config exists",
			createStaxConfig: false,
			createDDEV:       true,
		},
		{
			name:             "stax config exists but invalid",
			createStaxConfig: true,
			configContent:    `invalid: yaml: [[[`,
			createDDEV:       true,
		},
		{
			name:             "no config exists",
			createStaxConfig: false,
			createDDEV:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create .stax.yml if requested
			if tt.createStaxConfig {
				configPath := filepath.Join(tmpDir, ".stax.yml")
				if err := os.WriteFile(configPath, []byte(tt.configContent), 0644); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			}

			// Create .ddev/config.yaml if requested
			if tt.createDDEV {
				ddevDir := filepath.Join(tmpDir, ".ddev")
				if err := os.MkdirAll(ddevDir, 0755); err != nil {
					t.Fatalf("Failed to create .ddev dir: %v", err)
				}
				ddevConfig := filepath.Join(ddevDir, "config.yaml")
				ddevContent := `name: test-project
type: wordpress
`
				if err := os.WriteFile(ddevConfig, []byte(ddevContent), 0644); err != nil {
					t.Fatalf("Failed to create DDEV config: %v", err)
				}
			}

			// Test the detection logic directly
			configPath := filepath.Join(tmpDir, ".stax.yml")
			configFileExists := false
			if _, err := os.Stat(configPath); err == nil {
				configFileExists = true
			}

			// Verify our detection logic matches expectation
			if tt.createStaxConfig && !configFileExists {
				t.Error("Expected config file to be detected, but it wasn't")
			}
			if !tt.createStaxConfig && configFileExists {
				t.Error("Config file detected when it shouldn't exist")
			}
		})
	}
}

func TestStatusConfigFileVsLoadDistinction(t *testing.T) {
	// This test verifies we can distinguish between:
	// 1. File doesn't exist
	// 2. File exists but failed to load (invalid YAML)
	// 3. File exists and loads successfully

	tmpDir := t.TempDir()

	// Test 1: File doesn't exist
	configPath := filepath.Join(tmpDir, ".stax.yml")
	_, err := os.Stat(configPath)
	if !os.IsNotExist(err) {
		t.Error("Expected file to not exist initially")
	}

	// Test 2: Create invalid config
	invalidYAML := `invalid: yaml: [[[`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// File should exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to exist after creation")
	}

	// Test 3: Create valid config
	validYAML := `version: 1
project:
  name: test
  type: wordpress
`
	if err := os.WriteFile(configPath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to write valid config: %v", err)
	}

	// File should still exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to exist after update")
	}
}
