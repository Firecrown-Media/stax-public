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
	// Skip validation if no install configured
	if cfg.WPEngine.Install == "" {
		return nil
	}

	// Skip validation if no environment configured
	if cfg.WPEngine.Environment == "" {
		return nil
	}

	// Get credentials - skip validation if not available
	// This allows the command to work even if credentials aren't set up yet
	creds, err := credentials.GetWPEngineCredentialsWithFallback(cfg.WPEngine.Install)
	if err != nil {
		// Skip validation if no credentials (user may not have setup yet)
		return nil
	}

	// Create API client
	client := wpengine.NewClient(creds.APIUser, creds.APIPassword, cfg.WPEngine.Install)

	// Query actual environment from WPEngine API
	actualEnv, err := client.GetInstallEnvironment(cfg.WPEngine.Install)
	if err != nil {
		// Non-fatal: API might be unavailable or install might not exist yet
		// Return nil to allow operations to continue
		return nil
	}

	// Compare configured environment with actual environment
	if cfg.WPEngine.Environment != actualEnv {
		return errors.NewWithSolution(
			"Environment mismatch detected",
			fmt.Sprintf(".stax.yml has 'environment: %s' but WPEngine API reports '%s'",
				cfg.WPEngine.Environment, actualEnv),
			errors.Solution{
				Description: "Update .stax.yml to match WPEngine environment",
				Steps: []string{
					"Run 'stax doctor --fix' to automatically update .stax.yml",
					fmt.Sprintf("Or manually edit .stax.yml and change 'wpengine.environment' to '%s'", actualEnv),
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
	// Get credentials
	creds, err := credentials.GetWPEngineCredentialsWithFallback(cfg.WPEngine.Install)
	if err != nil {
		return "", fmt.Errorf("credentials not available: %w", err)
	}

	// Query actual environment from WPEngine API
	client := wpengine.NewClient(creds.APIUser, creds.APIPassword, cfg.WPEngine.Install)
	actualEnv, err := client.GetInstallEnvironment(cfg.WPEngine.Install)
	if err != nil {
		return "", fmt.Errorf("failed to query WPEngine API: %w", err)
	}

	// Check if there's actually a mismatch
	if cfg.WPEngine.Environment == actualEnv {
		return actualEnv, nil // No fix needed
	}

	// Update config
	cfg.WPEngine.Environment = actualEnv

	// Save config
	configPath := filepath.Join(projectDir, ".stax.yml")
	if err := Save(cfg, configPath); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return actualEnv, nil
}
