package config

import (
	"fmt"
	"path/filepath"

	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/wpengine"
)

// ValidateEnvironmentConfiguration checks if .stax.yml environment matches WPEngine API
// This is a non-fatal validation that helps users catch configuration mismatches early.
// Returns nil if environment matches or validation should be skipped.
// Returns an error with solution steps if mismatch detected.
func ValidateEnvironmentConfiguration(cfg *Config) error {
	// Skip validation if not a wpengine provider
	if cfg.Provider != "wpengine" {
		return nil
	}

	installVal, hasInstall := cfg.ProviderConfig["install"]
	install, _ := installVal.(string)
	if !hasInstall || install == "" {
		return nil
	}

	environmentVal, hasEnv := cfg.ProviderConfig["environment"]
	environment, _ := environmentVal.(string)
	if !hasEnv || environment == "" {
		return nil
	}

	// Get credentials - skip validation if not available
	// This allows the command to work even if credentials aren't set up yet
	creds, err := credentials.GetWPEngineCredentialsWithFallback(install)
	if err != nil {
		// Skip validation if no credentials (user may not have setup yet)
		return nil
	}

	// Create API client
	client := wpengine.NewClient(creds.APIUser, creds.APIPassword, install)

	// Query actual environment from WPEngine API
	actualEnv, err := client.GetInstallEnvironment(install)
	if err != nil {
		// Non-fatal: API might be unavailable or install might not exist yet
		// Return nil to allow operations to continue
		return nil
	}

	// Compare configured environment with actual environment
	if environment != actualEnv {
		return errors.NewWithSolution(
			"Environment mismatch detected",
			fmt.Sprintf(".stax.yml has 'environment: %s' but WPEngine API reports '%s'",
				environment, actualEnv),
			errors.Solution{
				Description: "Update .stax.yml to match WPEngine environment",
				Steps: []string{
					"Run 'stax doctor --fix' to automatically update .stax.yml",
					fmt.Sprintf("Or manually edit .stax.yml and change 'provider_config.environment' to '%s'", actualEnv),
					"Verify your WPEngine install environment at my.wpengine.com",
					"This mismatch won't prevent operations, but may cause confusion",
				},
			},
		)
	}

	return nil
}

// FixEnvironmentMismatch updates .stax.yml to match the actual WPEngine environment.
// Returns the corrected environment name on success, or an error if the fix fails.
func FixEnvironmentMismatch(cfg *Config, projectDir string) (string, error) {
	installVal, _ := cfg.ProviderConfig["install"]
	install, _ := installVal.(string)

	// Get credentials
	creds, err := credentials.GetWPEngineCredentialsWithFallback(install)
	if err != nil {
		return "", fmt.Errorf("credentials not available: %w", err)
	}

	// Query actual environment from WPEngine API
	client := wpengine.NewClient(creds.APIUser, creds.APIPassword, install)
	actualEnv, err := client.GetInstallEnvironment(install)
	if err != nil {
		return "", fmt.Errorf("failed to query WPEngine API: %w", err)
	}

	// Check if there's actually a mismatch
	currentEnv, _ := cfg.ProviderConfig["environment"].(string)
	if currentEnv == actualEnv {
		return actualEnv, nil // No fix needed
	}

	// Update config
	if cfg.ProviderConfig == nil {
		cfg.ProviderConfig = make(map[string]any)
	}
	cfg.ProviderConfig["environment"] = actualEnv

	// Save config
	configPath := filepath.Join(projectDir, ".stax.yml")
	if err := Save(cfg, configPath); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return actualEnv, nil
}
