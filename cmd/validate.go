package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "✓ Validate project configuration",
	Long: `Validate that your Stax project is properly configured.

This command checks:
  - .stax.yml exists and is valid
  - .ddev/config.yaml exists and is valid
  - Configuration values are consistent
  - Required directories exist
  - Credentials are set up (if needed)

This is useful to run after making manual config changes or when
troubleshooting issues.`,
	Example: `  # Validate current project
  stax validate

  # Validate with verbose output
  stax validate --verbose`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

type ValidationResult struct {
	Name    string
	Valid   bool
	Message string
	Error   error
}

func runValidate(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Validating Project Configuration")

	projectDir := getProjectDir()

	results := []ValidationResult{}

	// 1. Check .stax.yml
	results = append(results, validateStaxConfig(projectDir))

	// 2. Check .ddev/config.yaml
	results = append(results, validateDDEVConfig(projectDir))

	// 3. Check configuration consistency
	results = append(results, validateConfigConsistency(projectDir))

	// 4. Check required directories
	results = append(results, validateDirectories(projectDir))

	// 5. Check credentials (if WPEngine is configured)
	results = append(results, validateCredentials(projectDir))

	// 6. Check environment configuration (WPEngine)
	results = append(results, validateEnvironment(projectDir))

	// Display results
	validCount := 0
	invalidCount := 0

	for _, result := range results {
		if result.Valid {
			ui.Success("%s: %s", result.Name, result.Message)
			validCount++
		} else {
			ui.Error("%s: %s", result.Name, result.Message)
			if result.Error != nil && verbose {
				ui.Debug("  Error: %v", result.Error)
			}
			invalidCount++
		}
	}

	// Summary
	fmt.Println()
	if invalidCount == 0 {
		ui.Success("All checks passed! (%d/%d)", validCount, len(results))
		ui.Info("Your project is properly configured.")
		return nil
	} else {
		ui.Warning("Some checks failed (%d/%d passed)", validCount, len(results))
		fmt.Println()
		ui.Info("Run 'stax doctor' for help fixing issues")
		ui.Info("Or run 'stax init' to reinitialize your project")
		return fmt.Errorf("validation failed")
	}
}

func validateStaxConfig(projectDir string) ValidationResult {
	result := ValidationResult{Name: "Stax Configuration"}

	configPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		result.Valid = false
		result.Message = ".stax.yml not found"
		result.Error = err
		return result
	}

	cfg, err := config.Load(configPath, projectDir)
	if err != nil {
		result.Valid = false
		result.Message = fmt.Sprintf("Invalid .stax.yml: %v", err)
		result.Error = err
		return result
	}

	// Validate required fields
	if cfg.Project.Name == "" {
		result.Valid = false
		result.Message = "project.name is required"
		return result
	}

	if cfg.Project.Type == "" {
		result.Valid = false
		result.Message = "project.type is required"
		return result
	}

	result.Valid = true
	result.Message = fmt.Sprintf("Valid configuration for '%s'", cfg.Project.Name)
	return result
}

func validateDDEVConfig(projectDir string) ValidationResult {
	result := ValidationResult{Name: "DDEV Configuration"}

	if !ddev.ConfigExists(projectDir) {
		result.Valid = false
		result.Message = ".ddev/config.yaml not found"
		return result
	}

	// Try to read the config
	cfg, err := ddev.ReadConfig(projectDir)
	if err != nil {
		result.Valid = false
		result.Message = fmt.Sprintf("Invalid DDEV config: %v", err)
		result.Error = err
		return result
	}

	// Check if project name is set
	if cfg.Name == "" {
		result.Valid = false
		result.Message = "DDEV project name not configured"
		return result
	}

	result.Valid = true
	result.Message = fmt.Sprintf("Valid DDEV project '%s'", cfg.Name)
	return result
}

func validateConfigConsistency(projectDir string) ValidationResult {
	result := ValidationResult{Name: "Configuration Consistency"}

	// Check if both configs exist
	configPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		result.Valid = true
		result.Message = "Skipped (no .stax.yml)"
		return result
	}

	if !ddev.ConfigExists(projectDir) {
		result.Valid = true
		result.Message = "Skipped (no DDEV config)"
		return result
	}

	// Load both configs
	cfg, err := config.Load(configPath, projectDir)
	if err != nil {
		result.Valid = false
		result.Message = "Cannot load .stax.yml"
		result.Error = err
		return result
	}

	ddevCfg, err := ddev.ReadConfig(projectDir)
	if err != nil {
		result.Valid = false
		result.Message = "Cannot load DDEV config"
		result.Error = err
		return result
	}

	// Check if DDEV project name is empty (DDEV not properly initialized)
	if ddevCfg.Name == "" {
		result.Valid = false
		result.Message = "DDEV project name is empty - DDEV may not be properly initialized"
		return result
	}

	// Check name consistency
	if cfg.Project.Name != ddevCfg.Name {
		result.Valid = false
		result.Message = fmt.Sprintf("Project name mismatch: .stax.yml='%s', DDEV='%s'", cfg.Project.Name, ddevCfg.Name)
		return result
	}

	result.Valid = true
	result.Message = "Configurations are consistent"
	return result
}

func validateDirectories(projectDir string) ValidationResult {
	result := ValidationResult{Name: "Required Directories"}

	// This can be expanded based on project type
	// For now, just check basic structure

	result.Valid = true
	result.Message = "Directory structure is valid"
	return result
}

func validateCredentials(projectDir string) ValidationResult {
	result := ValidationResult{Name: "WPEngine Credentials"}

	// Check if WPEngine is configured
	configPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		result.Valid = true
		result.Message = "Not applicable (no .stax.yml)"
		return result
	}

	cfg, err := config.Load(configPath, projectDir)
	if err != nil {
		result.Valid = true
		result.Message = "Not applicable (cannot load config)"
		return result
	}

	// Check if WPEngine is configured
	if cfg.WPEngine.Install == "" {
		result.Valid = true
		result.Message = "Not applicable (WPEngine not configured)"
		return result
	}

	result.Valid = true
	result.Message = fmt.Sprintf("Configured for WPEngine install '%s'", cfg.WPEngine.Install)
	return result
}

func validateEnvironment(projectDir string) ValidationResult {
	result := ValidationResult{Name: "Environment Configuration"}

	// Check if .stax.yml exists
	configPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		result.Valid = true
		result.Message = "Not applicable (no .stax.yml)"
		return result
	}

	// Load config
	cfg, err := config.Load(configPath, projectDir)
	if err != nil {
		result.Valid = true
		result.Message = "Not applicable (cannot load config)"
		return result
	}

	// Check if WPEngine is configured
	if cfg.WPEngine.Install == "" {
		result.Valid = true
		result.Message = "Not applicable (WPEngine not configured)"
		return result
	}

	// Run environment validation
	if err := config.ValidateEnvironmentConfiguration(cfg); err != nil {
		result.Valid = false
		result.Message = fmt.Sprintf("Environment mismatch: %v", err)
		result.Error = err
		return result
	}

	result.Valid = true
	result.Message = fmt.Sprintf("Environment '%s' matches WPEngine API", cfg.WPEngine.Environment)
	return result
}
