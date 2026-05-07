package config

import (
	"testing"
)

func TestValidateEnvironmentConfiguration_NoInstall(t *testing.T) {
	cfg := &Config{
		Provider:       "wpengine",
		ProviderConfig: map[string]any{"environment": "production"},
	}

	err := ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		t.Errorf("Expected nil when no install configured, got: %v", err)
	}
}

func TestValidateEnvironmentConfiguration_NoEnvironment(t *testing.T) {
	cfg := &Config{
		Provider:       "wpengine",
		ProviderConfig: map[string]any{"install": "testinstall"},
	}

	err := ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		t.Errorf("Expected nil when no environment configured, got: %v", err)
	}
}

func TestValidateEnvironmentConfiguration_NoCredentials(t *testing.T) {
	cfg := &Config{
		Provider: "wpengine",
		ProviderConfig: map[string]any{
			"install":     "nonexistentinstall",
			"environment": "production",
		},
	}

	err := ValidateEnvironmentConfiguration(cfg)
	// Should be nil because credentials won't be found
	if err != nil {
		t.Errorf("Expected nil when credentials not available, got: %v", err)
	}
}

func TestValidateEnvironmentConfiguration_BothEmpty(t *testing.T) {
	cfg := &Config{
		Provider:       "wpengine",
		ProviderConfig: map[string]any{},
	}

	err := ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		t.Errorf("Expected nil when both empty, got: %v", err)
	}
}

func TestValidateEnvironmentConfiguration_NilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ValidateEnvironmentConfiguration panicked with empty config: %v", r)
		}
	}()

	cfg := &Config{}
	err := ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		t.Errorf("Expected nil for empty config, got: %v", err)
	}
}

func TestValidateEnvironmentConfiguration_GracefulAPIFailure(t *testing.T) {
	cfg := &Config{
		Provider: "wpengine",
		ProviderConfig: map[string]any{
			"install":     "testinstall-that-does-not-exist",
			"environment": "production",
		},
	}

	err := ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		t.Errorf("Expected nil (skip validation), got error: %v", err)
	}
}
