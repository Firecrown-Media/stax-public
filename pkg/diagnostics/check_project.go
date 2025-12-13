package diagnostics

import (
	"os"
	"path/filepath"
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
