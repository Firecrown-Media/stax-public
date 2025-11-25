package config

import (
	"testing"
)

func TestValidateEnvironmentConfiguration_NoInstall(t *testing.T) {
	cfg := &Config{
		WPEngine: WPEngineConfig{
			Install:     "",
			Environment: "production",
		},
	}

	err := ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		t.Errorf("Expected nil when no install configured, got: %v", err)
	}
}

func TestValidateEnvironmentConfiguration_NoEnvironment(t *testing.T) {
	cfg := &Config{
		WPEngine: WPEngineConfig{
			Install:     "testinstall",
			Environment: "",
		},
	}

	err := ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		t.Errorf("Expected nil when no environment configured, got: %v", err)
	}
}

func TestValidateEnvironmentConfiguration_NoCredentials(t *testing.T) {
	// This test validates that validation is skipped when credentials aren't available
	// In a real environment without credentials set up, this should return nil
	cfg := &Config{
		WPEngine: WPEngineConfig{
			Install:     "nonexistentinstall",
			Environment: "production",
		},
	}

	err := ValidateEnvironmentConfiguration(cfg)
	// Should be nil because credentials won't be found
	if err != nil {
		t.Errorf("Expected nil when credentials not available, got: %v", err)
	}
}

// Note: Testing the actual API call and mismatch detection would require:
// 1. Mocking the WPEngine API client
// 2. Setting up test credentials
// 3. Creating integration tests with a test WPEngine install
// These tests verify the basic validation logic and graceful handling of missing data
