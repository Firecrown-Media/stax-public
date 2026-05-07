package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigVersion represents the configuration version metadata
type ConfigVersion struct {
	Version  int    `yaml:"version"`
	Schema   string `yaml:"schema,omitempty"`
	Migrated string `yaml:"migrated,omitempty"` // Timestamp of last migration
}

// MigrationChange represents a single change during migration
type MigrationChange struct {
	Type        string // "rename", "add", "remove", "update", "restructure"
	Description string
	OldPath     string
	NewPath     string
	OldValue    interface{}
	NewValue    interface{}
}

// MigrationPlan represents a planned migration
type MigrationPlan struct {
	FromVersion    int
	ToVersion      int
	Changes        []MigrationChange
	RequiresBackup bool
}

const (
	CurrentVersion = 1
	CurrentSchema  = "v1"
)

// DetectConfigVersion reads a config file and returns its version information
func DetectConfigVersion(configPath string) (*ConfigVersion, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse just the version fields
	var version ConfigVersion
	if err := yaml.Unmarshal(data, &version); err != nil {
		return nil, fmt.Errorf("failed to parse config version: %w", err)
	}

	// Check if version field exists in the file by looking for "version:" in the content
	hasVersionField := false
	lines := string(data)
	if yaml.Unmarshal([]byte(lines), &map[string]interface{}{}) == nil {
		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err == nil {
			if _, ok := raw["version"]; ok {
				hasVersionField = true
			}
		}
	}

	// If version is 0 and no version field exists, it's an unversioned config
	// Otherwise, keep the parsed version (which could be 0, 1, etc.)
	if version.Version == 0 && !hasVersionField {
		// Leave as 0 to indicate unversioned
		version.Schema = "v0"
	} else if version.Version == 0 && hasVersionField {
		// Explicit version: 0
		version.Schema = "v0"
	} else if version.Schema == "" {
		// Infer schema from version if not set
		version.Schema = fmt.Sprintf("v%d", version.Version)
	}

	return &version, nil
}

// NeedsMigration checks if a config needs migration and returns the target version
func NeedsMigration(cfg *Config) (bool, string) {
	if cfg.Version == 0 || cfg.Version < CurrentVersion {
		return true, fmt.Sprintf("v%d → v%d", cfg.Version, CurrentVersion)
	}
	return false, ""
}

// GetMigrationPlan analyzes a config and returns a migration plan
func GetMigrationPlan(configPath string) (*MigrationPlan, error) {
	version, err := DetectConfigVersion(configPath)
	if err != nil {
		return nil, err
	}

	// Check if migration is needed
	if version.Version >= CurrentVersion {
		return nil, fmt.Errorf("config is already at version %d (latest)", CurrentVersion)
	}

	// Load the config to analyze
	cfg, err := loadConfigFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config for analysis: %w", err)
	}

	plan := &MigrationPlan{
		FromVersion:    version.Version,
		ToVersion:      CurrentVersion,
		Changes:        []MigrationChange{},
		RequiresBackup: true,
	}

	// Build migration plan based on version
	// For now, we'll add example migrations that might be common
	// These can be expanded as the config schema evolves

	// Example: If migrating from version 0 (unversioned) to 1
	if version.Version == 0 {
		plan.Changes = append(plan.Changes, MigrationChange{
			Type:        "add",
			Description: "Add version field",
			NewPath:     "version",
			NewValue:    1,
		})

		// Check for deprecated fields that might exist in old configs
		if hasDeprecatedFields(cfg) {
			plan.Changes = append(plan.Changes, getDeprecatedFieldMigrations(cfg)...)
		}
	}

	// Future: Add more version-specific migrations here
	// Example:
	// if version.Version == 1 && CurrentVersion >= 2 {
	//     plan.Changes = append(plan.Changes, migrateV1ToV2Changes(cfg)...)
	// }

	return plan, nil
}

// hasDeprecatedFields checks if config has any deprecated fields
func hasDeprecatedFields(cfg *Config) bool {
	// Example checks for deprecated fields
	// This is where you'd add checks for old field names
	// For now, we'll return false as current config is version 1
	return false
}

// getDeprecatedFieldMigrations returns migration changes for deprecated fields
func getDeprecatedFieldMigrations(cfg *Config) []MigrationChange {
	changes := []MigrationChange{}

	// Example: If old configs had a "provider" field that's now removed
	// changes = append(changes, MigrationChange{
	//     Type:        "remove",
	//     Description: "Remove deprecated 'provider' field",
	//     OldPath:     "provider",
	// })

	return changes
}

// MigrateConfig migrates a configuration file to the latest version
func MigrateConfig(configPath string, dryRun bool) (*MigrationPlan, error) {
	// Get migration plan
	plan, err := GetMigrationPlan(configPath)
	if err != nil {
		return nil, err
	}

	if dryRun {
		// Just return the plan without executing
		return plan, nil
	}

	// Create backup before migration
	if plan.RequiresBackup {
		if err := BackupConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Load the config
	cfg, err := loadConfigFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Execute migration based on version path
	migratedCfg := cfg
	for v := plan.FromVersion; v < plan.ToVersion; v++ {
		migratedCfg = executeMigration(migratedCfg, v, v+1)
	}

	// Update version and add migration timestamp
	migratedCfg.Version = plan.ToVersion

	// Save migrated config
	if err := Save(migratedCfg, configPath); err != nil {
		return nil, fmt.Errorf("failed to save migrated config: %w", err)
	}

	// Validate migrated config
	if err := ValidateConfig(migratedCfg); err != nil {
		return nil, fmt.Errorf("migrated config failed validation: %w", err)
	}

	return plan, nil
}

// executeMigration executes a migration from one version to another
func executeMigration(cfg *Config, fromVersion, toVersion int) *Config {
	switch {
	case fromVersion == 0 && toVersion == 1:
		return migrateV0ToV1(cfg)
	// Future migrations:
	// case fromVersion == 1 && toVersion == 2:
	//     return migrateV1ToV2(cfg)
	default:
		return cfg
	}
}

// migrateV0ToV1 migrates from unversioned to version 1
func migrateV0ToV1(cfg *Config) *Config {
	// Set version
	cfg.Version = 1

	// Ensure all required fields have defaults
	if cfg.Project.Type == "" {
		cfg.Project.Type = "wordpress-multisite"
	}
	if cfg.Project.Mode == "" {
		cfg.Project.Mode = "subdomain"
	}

	// Set sensible defaults for DDEV if not set
	if cfg.DDEV.PHPVersion == "" {
		cfg.DDEV.PHPVersion = "8.1"
	}
	if cfg.DDEV.MySQLVersion == "" {
		cfg.DDEV.MySQLVersion = "8.0"
	}
	if cfg.DDEV.MySQLType == "" {
		cfg.DDEV.MySQLType = "mysql"
	}
	if cfg.DDEV.WebserverType == "" {
		cfg.DDEV.WebserverType = "nginx-fpm"
	}

	// Set provider defaults
	if cfg.Provider == "" {
		cfg.Provider = "wpengine"
	}
	if cfg.ProviderConfig == nil {
		cfg.ProviderConfig = make(map[string]any)
	}
	if _, ok := cfg.ProviderConfig["environment"]; !ok {
		cfg.ProviderConfig["environment"] = "production"
	}
	if _, ok := cfg.ProviderConfig["ssh_gateway"]; !ok {
		cfg.ProviderConfig["ssh_gateway"] = "ssh.wpengine.net"
	}

	// Set WordPress defaults
	if cfg.WordPress.Version == "" {
		cfg.WordPress.Version = "latest"
	}
	if cfg.WordPress.Locale == "" {
		cfg.WordPress.Locale = "en_US"
	}
	if cfg.WordPress.TablePrefix == "" {
		cfg.WordPress.TablePrefix = "wp_"
	}

	return cfg
}

// Example future migration (template for when version 2 is needed)
// func migrateV1ToV2(cfg *Config) *Config {
//     // Example: Rename field
//     // if cfg.OldField != "" {
//     //     cfg.NewField = cfg.OldField
//     //     cfg.OldField = ""
//     // }
//
//     // Example: Restructure nested config
//     // cfg.NewSection = NewSectionType{
//     //     Field: cfg.OldSection.Field,
//     // }
//
//     // Update version
//     cfg.Version = 2
//     return cfg
// }

// BackupConfig creates a backup of the config file
func BackupConfig(configPath string) error {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup.%s", configPath, timestamp)

	// Read original file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Write backup
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// GetBackupPath returns the backup path for a config file
func GetBackupPath(configPath string) string {
	return configPath + ".backup"
}

// ListBackups returns a list of backup files for a config
func ListBackups(configPath string) ([]string, error) {
	dir := filepath.Dir(configPath)
	base := filepath.Base(configPath)
	pattern := filepath.Join(dir, base+".backup.*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

// ValidateConfig performs validation on a config
func ValidateConfig(cfg *Config) error {
	errors := []string{}

	// Check required fields
	if cfg.Project.Name == "" {
		errors = append(errors, "project.name is required")
	}
	if cfg.Project.Type == "" {
		errors = append(errors, "project.type is required")
	}
	// Validate enums
	validProjectTypes := map[string]bool{
		"wordpress":           true,
		"wordpress-multisite": true,
	}
	if !validProjectTypes[cfg.Project.Type] {
		errors = append(errors, "project.type must be 'wordpress' or 'wordpress-multisite'")
	}

	validModes := map[string]bool{
		"single":       true,
		"subdomain":    true,
		"subdirectory": true,
	}
	if cfg.Project.Mode != "" && !validModes[cfg.Project.Mode] {
		errors = append(errors, "project.mode must be 'single', 'subdomain', or 'subdirectory'")
	}

	if env, ok := cfg.ProviderConfig["environment"]; ok {
		validEnvironments := map[string]bool{
			"production":  true,
			"staging":     true,
			"development": true,
		}
		if envStr, ok := env.(string); ok && !validEnvironments[envStr] {
			errors = append(errors, "provider_config.environment must be 'production', 'staging', or 'development'")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	return nil
}
