package diagnostics

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/firecrown-media/stax/pkg/config"
)

// CheckStaxConfig checks if .stax.yml exists and is valid
func CheckStaxConfig(projectPath string) CheckResult {
	configPath := filepath.Join(projectPath, ".stax.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return CheckResult{
			Name:       "Stax Configuration",
			Category:   "Configuration",
			Status:     StatusFail,
			Message:    ".stax.yml not found",
			Suggestion: "Create configuration: stax init or stax config template > .stax.yml",
			Details: map[string]string{
				"expected_path": configPath,
			},
		}
	}

	return CheckResult{
		Name:     "Stax Configuration",
		Category: "Configuration",
		Status:   StatusPass,
		Message:  ".stax.yml found",
		Details: map[string]string{
			"path": configPath,
		},
	}
}

// CheckDDEVConfig checks if DDEV is configured for the project
func CheckDDEVConfig(projectPath string) CheckResult {
	ddevPath := filepath.Join(projectPath, ".ddev")
	if _, err := os.Stat(ddevPath); os.IsNotExist(err) {
		return CheckResult{
			Name:       "DDEV Configuration",
			Category:   "Configuration",
			Status:     StatusWarning,
			Message:    ".ddev directory not found",
			Suggestion: "Initialize DDEV: stax init or ddev config",
			Details: map[string]string{
				"expected_path": ddevPath,
			},
		}
	}

	configPath := filepath.Join(ddevPath, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return CheckResult{
			Name:       "DDEV Configuration",
			Category:   "Configuration",
			Status:     StatusWarning,
			Message:    ".ddev/config.yaml not found",
			Suggestion: "Configure DDEV: stax init or ddev config",
			Details: map[string]string{
				"expected_path": configPath,
			},
		}
	}

	return CheckResult{
		Name:     "DDEV Configuration",
		Category: "Configuration",
		Status:   StatusPass,
		Message:  "DDEV is configured",
		Details: map[string]string{
			"path": configPath,
		},
	}
}

// CheckEnvironmentMismatch checks if .stax.yml environment matches WPEngine API
func CheckEnvironmentMismatch(projectPath string) CheckResult {
	// Try to load config
	cfg, err := config.Load("", projectPath)
	if err != nil {
		return CheckResult{
			Name:     "WPEngine Environment",
			Category: "Configuration",
			Status:   StatusSkip,
			Message:  "Could not load .stax.yml",
		}
	}

	// Check for environment mismatch
	err = config.ValidateEnvironmentConfiguration(cfg)
	if err != nil {
		return CheckResult{
			Name:       "WPEngine Environment",
			Category:   "Configuration",
			Status:     StatusWarning,
			Message:    err.Error(),
			Suggestion: "Run 'stax doctor --fix' to automatically correct the environment",
			CanAutoFix: true,
			Details: map[string]string{
				"configured": providerConfigStr(cfg, "environment"),
				"install":    providerConfigStr(cfg, "install"),
			},
		}
	}

	return CheckResult{
		Name:     "WPEngine Environment",
		Category: "Configuration",
		Status:   StatusPass,
		Message:  fmt.Sprintf("Environment '%s' matches WPEngine API", providerConfigStr(cfg, "environment")),
		Details: map[string]string{
			"environment": providerConfigStr(cfg, "environment"),
			"install":     providerConfigStr(cfg, "install"),
		},
	}
}

// FixEnvironmentMismatch attempts to fix the environment mismatch
func providerConfigStr(cfg *config.Config, key string) string {
	if cfg.ProviderConfig == nil {
		return ""
	}
	v, _ := cfg.ProviderConfig[key].(string)
	return v
}

func FixEnvironmentMismatch(projectPath string, check CheckResult) CheckResult {
	cfg, err := config.Load("", projectPath)
	if err != nil {
		check.Message = fmt.Sprintf("Could not load config: %v", err)
		return check
	}

	actualEnv, err := config.FixEnvironmentMismatch(cfg, projectPath)
	if err != nil {
		check.Message = fmt.Sprintf("Could not fix environment: %v", err)
		return check
	}

	check.Status = StatusPass
	check.Message = fmt.Sprintf("Environment updated to '%s'", actualEnv)
	check.FixApplied = true
	check.Suggestion = ""

	return check
}
