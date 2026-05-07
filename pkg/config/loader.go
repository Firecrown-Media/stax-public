package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load loads configuration from multiple sources and merges them
func Load(cfgFile string, projectDir string) (*Config, error) {
	cfg := Defaults()

	// Determine project directory
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Load global config first (~/.stax/config.yml)
	globalCfg, err := loadGlobalConfig()
	if err == nil {
		// Merge global config into defaults
		cfg = mergeConfigs(cfg, globalCfg)
	}

	// Load project config
	projectCfgFile := cfgFile
	if projectCfgFile == "" {
		projectCfgFile = filepath.Join(projectDir, ".stax.yml")
	}

	projectCfg, err := loadProjectConfig(projectCfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}

	// Merge project config (overrides global)
	cfg = mergeConfigs(cfg, projectCfg)

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	return cfg, nil
}

// loadGlobalConfig loads the global configuration file
func loadGlobalConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	globalCfgPath := filepath.Join(homeDir, ".stax", "config.yml")
	return loadConfigFile(globalCfgPath)
}

// loadProjectConfig loads the project configuration file
func loadProjectConfig(path string) (*Config, error) {
	return loadConfigFile(path)
}

// loadConfigFile loads a configuration file from disk
func loadConfigFile(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to a file
func Save(cfg *Config, path string) error {
	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// mergeConfigs merges two configs, with the second taking precedence
func mergeConfigs(base, override *Config) *Config {
	// For now, we'll do a simple override
	// In a full implementation, this would do a deep merge

	result := base

	// Override version if set
	if override.Version != 0 {
		result.Version = override.Version
	}

	// Override project config
	if override.Project.Name != "" {
		result.Project.Name = override.Project.Name
	}
	if override.Project.Type != "" {
		result.Project.Type = override.Project.Type
	}
	if override.Project.Mode != "" {
		result.Project.Mode = override.Project.Mode
	}
	if override.Project.Description != "" {
		result.Project.Description = override.Project.Description
	}

	// Override provider config
	if override.Provider != "" {
		result.Provider = override.Provider
	}
	if len(override.ProviderConfig) > 0 {
		if result.ProviderConfig == nil {
			result.ProviderConfig = make(map[string]any)
		}
		for k, v := range override.ProviderConfig {
			result.ProviderConfig[k] = v
		}
	}

	// Override DDEV config
	if override.DDEV.PHPVersion != "" {
		result.DDEV.PHPVersion = override.DDEV.PHPVersion
	}
	if override.DDEV.MySQLVersion != "" {
		result.DDEV.MySQLVersion = override.DDEV.MySQLVersion
	}
	if override.DDEV.WebserverType != "" {
		result.DDEV.WebserverType = override.DDEV.WebserverType
	}

	// Override network config
	if override.Network.Domain != "" {
		result.Network.Domain = override.Network.Domain
	}
	if override.Network.Title != "" {
		result.Network.Title = override.Network.Title
	}
	if override.Network.AdminEmail != "" {
		result.Network.AdminEmail = override.Network.AdminEmail
	}
	if len(override.Network.Sites) > 0 {
		result.Network.Sites = override.Network.Sites
	}

	// Override repository config
	if override.Repository.URL != "" {
		result.Repository.URL = override.Repository.URL
	}
	if override.Repository.Branch != "" {
		result.Repository.Branch = override.Repository.Branch
	}

	// Override Build config - Composer
	if override.Build.Composer.InstallArgs != "" {
		result.Build.Composer.InstallArgs = override.Build.Composer.InstallArgs
	}
	if override.Build.Composer.Timeout != 0 {
		result.Build.Composer.Timeout = override.Build.Composer.Timeout
	}
	result.Build.Composer.Optimize = override.Build.Composer.Optimize
	result.Build.Composer.NoDev = override.Build.Composer.NoDev
	result.Build.Composer.IgnorePlatformReqs = override.Build.Composer.IgnorePlatformReqs

	// Override Build config - NPM
	if override.Build.NPM.InstallArgs != "" {
		result.Build.NPM.InstallArgs = override.Build.NPM.InstallArgs
	}
	if override.Build.NPM.BuildCommand != "" {
		result.Build.NPM.BuildCommand = override.Build.NPM.BuildCommand
	}
	if override.Build.NPM.DevCommand != "" {
		result.Build.NPM.DevCommand = override.Build.NPM.DevCommand
	}
	if override.Build.NPM.Timeout != 0 {
		result.Build.NPM.Timeout = override.Build.NPM.Timeout
	}
	result.Build.NPM.LegacyPeerDeps = override.Build.NPM.LegacyPeerDeps

	// Override Build config - PHPCS
	if override.Build.PHPCS.Config != "" {
		result.Build.PHPCS.Config = override.Build.PHPCS.Config
	}
	if override.Build.PHPCS.Standard != "" {
		result.Build.PHPCS.Standard = override.Build.PHPCS.Standard
	}
	if override.Build.PHPCS.Extensions != "" {
		result.Build.PHPCS.Extensions = override.Build.PHPCS.Extensions
	}
	if override.Build.PHPCS.Ignore != "" {
		result.Build.PHPCS.Ignore = override.Build.PHPCS.Ignore
	}
	result.Build.PHPCS.ShowSniffs = override.Build.PHPCS.ShowSniffs

	// Override Build config - Hooks
	result.Build.Hooks.PreCommit = override.Build.Hooks.PreCommit
	result.Build.Hooks.PrePush = override.Build.Hooks.PrePush
	result.Build.Hooks.CommitMsg = override.Build.Hooks.CommitMsg

	// Override Build config - Scripts
	if override.Build.Scripts.Main != "" {
		result.Build.Scripts.Main = override.Build.Scripts.Main
	}
	if len(override.Build.Scripts.Additional) > 0 {
		result.Build.Scripts.Additional = override.Build.Scripts.Additional
	}
	if len(override.Build.Scripts.PreBuild) > 0 {
		result.Build.Scripts.PreBuild = override.Build.Scripts.PreBuild
	}
	if len(override.Build.Scripts.PostBuild) > 0 {
		result.Build.Scripts.PostBuild = override.Build.Scripts.PostBuild
	}

	return result
}

// applyEnvOverrides applies environment variable overrides
func applyEnvOverrides(cfg *Config) {
	// Check for common environment variables
	if val := os.Getenv("STAX_PROJECT_NAME"); val != "" {
		cfg.Project.Name = val
	}
	if val := os.Getenv("STAX_WPENGINE_INSTALL"); val != "" {
		if cfg.ProviderConfig == nil {
			cfg.ProviderConfig = make(map[string]any)
		}
		cfg.ProviderConfig["install"] = val
	}
	if val := os.Getenv("STAX_WPENGINE_ENVIRONMENT"); val != "" {
		if cfg.ProviderConfig == nil {
			cfg.ProviderConfig = make(map[string]any)
		}
		cfg.ProviderConfig["environment"] = val
	}
	if val := os.Getenv("STAX_DDEV_PHP_VERSION"); val != "" {
		cfg.DDEV.PHPVersion = val
	}
	if val := os.Getenv("STAX_DDEV_MYSQL_VERSION"); val != "" {
		cfg.DDEV.MySQLVersion = val
	}
	if val := os.Getenv("STAX_LOGGING_LEVEL"); val != "" {
		cfg.Logging.Level = val
	}
}

// GetGlobalConfigPath returns the path to the global config file
func GetGlobalConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".stax", "config.yml"), nil
}

// GetProjectConfigPath returns the path to the project config file
func GetProjectConfigPath(projectDir string) string {
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	return filepath.Join(projectDir, ".stax.yml")
}
