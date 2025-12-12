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

func TestValidateEnvironmentConfiguration_BothEmpty(t *testing.T) {
	cfg := &Config{
		WPEngine: WPEngineConfig{
			Install:     "",
			Environment: "",
		},
	}

	err := ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		t.Errorf("Expected nil when both empty, got: %v", err)
	}
}

func TestValidateEnvironmentConfiguration_NilConfig(t *testing.T) {
	// Test that nil config doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ValidateEnvironmentConfiguration panicked with nil WPEngine: %v", r)
		}
	}()

	cfg := &Config{}
	err := ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		t.Errorf("Expected nil for empty config, got: %v", err)
	}
}

func TestValidateEnvironmentConfiguration_GracefulAPIFailure(t *testing.T) {
	// This test verifies that validation is skipped gracefully when:
	// - Install is configured
	// - Environment is configured
	// - But API is unavailable (no credentials or API error)

	cfg := &Config{
		WPEngine: WPEngineConfig{
			Install:     "testinstall-that-does-not-exist",
			Environment: "production",
		},
	}

	// Without credentials, this should skip validation gracefully
	err := ValidateEnvironmentConfiguration(cfg)

	// Should be nil because it gracefully skips when credentials unavailable
	if err != nil {
		t.Errorf("Expected nil (skip validation), got error: %v", err)
	}
}
