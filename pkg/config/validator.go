package config

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationSeverity represents the severity level of a validation issue
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field    string
	Message  string
	Severity ValidationSeverity
	Fix      string // Suggested fix
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
	Infos    []ValidationError
}

// Validate validates the configuration
func Validate(cfg *Config) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
		Infos:    []ValidationError{},
	}

	// Validate required fields
	validateRequired(cfg, result)

	// Validate field formats
	validateFormats(cfg, result)

	// Validate cross-field constraints
	validateConstraints(cfg, result)

	// Validate version compatibility
	validateVersions(cfg, result)

	// Validate WPEngine configuration
	validateWPEngine(cfg, result)

	// Validate Network configuration
	validateNetwork(cfg, result)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}

// ValidateProjectSection validates the project configuration section
func ValidateProjectSection(cfg *Config) []ValidationError {
	var errors []ValidationError

	if cfg.Project.Name == "" {
		errors = append(errors, ValidationError{
			Field:    "project.name",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: project:\n      name: my-project",
		})
	}

	if cfg.Project.Type == "" {
		errors = append(errors, ValidationError{
			Field:    "project.type",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: project:\n      type: wordpress-multisite",
		})
	} else {
		validTypes := []string{"wordpress", "wordpress-multisite"}
		if !contains(validTypes, cfg.Project.Type) {
			errors = append(errors, ValidationError{
				Field:    "project.type",
				Message:  fmt.Sprintf("must be 'wordpress' or 'wordpress-multisite', got '%s'", cfg.Project.Type),
				Severity: SeverityError,
				Fix:      fmt.Sprintf("Change '%s' to 'wordpress' or 'wordpress-multisite'", cfg.Project.Type),
			})
		}
	}

	if cfg.Project.Mode == "" {
		errors = append(errors, ValidationError{
			Field:    "project.mode",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: project:\n      mode: subdomain",
		})
	}

	return errors
}

// ValidateProviderSection validates the provider configuration section
func ValidateProviderSection(cfg *Config) []ValidationError {
	var errors []ValidationError

	install, _ := cfg.ProviderConfig["install"].(string)
	if install == "" {
		errors = append(errors, ValidationError{
			Field:    "provider_config.install",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: provider_config:\n      install: myinstall",
		})
	}

	environment, _ := cfg.ProviderConfig["environment"].(string)
	if environment == "" {
		errors = append(errors, ValidationError{
			Field:    "provider_config.environment",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: provider_config:\n      environment: production",
		})
	} else {
		validEnvs := []string{"production", "staging", "development"}
		if !contains(validEnvs, environment) {
			errors = append(errors, ValidationError{
				Field:    "provider_config.environment",
				Message:  fmt.Sprintf("must be one of: %s", strings.Join(validEnvs, ", ")),
				Severity: SeverityError,
				Fix:      "Change to 'production', 'staging', or 'development'",
			})
		}
	}

	return errors
}

// ValidateWPEngineSection is an alias for ValidateProviderSection for backwards compatibility.
func ValidateWPEngineSection(cfg *Config) []ValidationError {
	return ValidateProviderSection(cfg)
}

// ValidateNetworkSection validates the network configuration section
func ValidateNetworkSection(cfg *Config) []ValidationError {
	var errors []ValidationError

	if cfg.Project.Type == "wordpress-multisite" && cfg.Network.Domain == "" {
		errors = append(errors, ValidationError{
			Field:    "network.domain",
			Message:  "is required for multisite projects",
			Severity: SeverityError,
			Fix:      "Add: network:\n      domain: example.local",
		})
	}

	return errors
}

// validateRequired checks that required fields are present
func validateRequired(cfg *Config, result *ValidationResult) {
	if cfg.Project.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "project.name",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: project:\n      name: my-project",
		})
	}

	if cfg.Project.Type == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "project.type",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: project:\n      type: wordpress-multisite",
		})
	}

	if cfg.Project.Mode == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "project.mode",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: project:\n      mode: subdomain",
		})
	}

	install, _ := cfg.ProviderConfig["install"].(string)
	if install == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "provider_config.install",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: provider_config:\n      install: myinstall",
		})
	}

	environment, _ := cfg.ProviderConfig["environment"].(string)
	if environment == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "provider_config.environment",
			Message:  "is required",
			Severity: SeverityError,
			Fix:      "Add: provider_config:\n      environment: production",
		})
	}

	// Network domain only required for multisite
	if cfg.Project.Type == "wordpress-multisite" && cfg.Network.Domain == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "network.domain",
			Message:  "is required for multisite projects",
			Severity: SeverityError,
			Fix:      "Add: network:\n      domain: example.local",
		})
	}
}

// validateFormats checks that field formats are valid
func validateFormats(cfg *Config, result *ValidationResult) {
	// Validate project name (alphanumeric, hyphens, underscores)
	if cfg.Project.Name != "" {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, cfg.Project.Name)
		if !matched {
			result.Errors = append(result.Errors, ValidationError{
				Field:    "project.name",
				Message:  "must contain only alphanumeric characters, hyphens, and underscores",
				Severity: SeverityError,
				Fix:      fmt.Sprintf("Remove special characters from '%s'", cfg.Project.Name),
			})
		}
	}

	// Validate project type
	validTypes := []string{"wordpress", "wordpress-multisite"}
	if cfg.Project.Type != "" && !contains(validTypes, cfg.Project.Type) {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "project.type",
			Message:  fmt.Sprintf("must be 'wordpress' or 'wordpress-multisite', got '%s'", cfg.Project.Type),
			Severity: SeverityError,
			Fix:      fmt.Sprintf("Change '%s' to 'wordpress' or 'wordpress-multisite'", cfg.Project.Type),
		})
	}

	// Validate project mode
	validModes := []string{"subdomain", "subdirectory", "single"}
	if cfg.Project.Mode != "" && !contains(validModes, cfg.Project.Mode) {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "project.mode",
			Message:  fmt.Sprintf("must be one of: %s", strings.Join(validModes, ", ")),
			Severity: SeverityError,
			Fix:      fmt.Sprintf("Change '%s' to 'subdomain', 'subdirectory', or 'single'", cfg.Project.Mode),
		})
	}

	// Validate provider_config environment
	validEnvs := []string{"production", "staging", "development"}
	if env, _ := cfg.ProviderConfig["environment"].(string); env != "" && !contains(validEnvs, env) {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "provider_config.environment",
			Message:  fmt.Sprintf("must be one of: %s", strings.Join(validEnvs, ", ")),
			Severity: SeverityError,
			Fix:      fmt.Sprintf("Change '%s' to 'production', 'staging', or 'development'", env),
		})
	}

	// Validate domain format
	if cfg.Network.Domain != "" {
		if !isValidDomain(cfg.Network.Domain) {
			result.Errors = append(result.Errors, ValidationError{
				Field:    "network.domain",
				Message:  "is not a valid domain format",
				Severity: SeverityError,
				Fix:      fmt.Sprintf("Use a valid domain format like 'example.local' instead of '%s'", cfg.Network.Domain),
			})
		}
	}

	// Validate site domains
	for i, site := range cfg.Network.Sites {
		if site.Domain != "" && !isValidDomain(site.Domain) {
			result.Errors = append(result.Errors, ValidationError{
				Field:    fmt.Sprintf("network.sites[%d].domain", i),
				Message:  "is not a valid domain format",
				Severity: SeverityError,
				Fix:      fmt.Sprintf("Use a valid domain format like 'site.example.local'"),
			})
		}
	}

	// Validate PHP version
	validPHPVersions := []string{"7.4", "8.0", "8.1", "8.2", "8.3"}
	if cfg.DDEV.PHPVersion != "" && !contains(validPHPVersions, cfg.DDEV.PHPVersion) {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "ddev.php_version",
			Message:  fmt.Sprintf("must be one of: %s", strings.Join(validPHPVersions, ", ")),
			Severity: SeverityError,
			Fix:      "Use a supported PHP version: 7.4, 8.0, 8.1, 8.2, or 8.3",
		})
	}
}

// validateConstraints checks cross-field constraints
func validateConstraints(cfg *Config, result *ValidationResult) {
	// If mode is subdomain, ensure site domains are subdomains of network domain
	// Allow first site to be the root domain (primary site)
	if cfg.Project.Mode == "subdomain" && cfg.Network.Domain != "" {
		for i, site := range cfg.Network.Sites {
			if site.Domain != "" {
				// Allow the first site to be the root domain or a subdomain
				isRootDomain := site.Domain == cfg.Network.Domain
				isSubdomain := strings.HasSuffix(site.Domain, "."+cfg.Network.Domain)

				if !isRootDomain && !isSubdomain {
					result.Errors = append(result.Errors, ValidationError{
						Field:    fmt.Sprintf("network.sites[%d].domain", i),
						Message:  fmt.Sprintf("must be a subdomain of %s or the root domain", cfg.Network.Domain),
						Severity: SeverityError,
						Fix:      fmt.Sprintf("Change to a subdomain like 'site%d.%s' or use '%s'", i+1, cfg.Network.Domain, cfg.Network.Domain),
					})
				}
			}
		}
	}

	// Check for duplicate site domains
	domainMap := make(map[string]bool)
	for i, site := range cfg.Network.Sites {
		if site.Domain != "" {
			if domainMap[site.Domain] {
				result.Errors = append(result.Errors, ValidationError{
					Field:    fmt.Sprintf("network.sites[%d].domain", i),
					Message:  fmt.Sprintf("duplicate domain '%s'", site.Domain),
					Severity: SeverityError,
					Fix:      "Use a unique domain for each site",
				})
			}
			domainMap[site.Domain] = true
		}
	}

	// Check for duplicate site slugs
	slugMap := make(map[string]bool)
	for i, site := range cfg.Network.Sites {
		if site.Slug != "" {
			if slugMap[site.Slug] {
				result.Errors = append(result.Errors, ValidationError{
					Field:    fmt.Sprintf("network.sites[%d].slug", i),
					Message:  fmt.Sprintf("duplicate slug '%s'", site.Slug),
					Severity: SeverityError,
					Fix:      "Use a unique slug for each site",
				})
			}
			slugMap[site.Slug] = true
		}
	}
}

// validateVersions checks version compatibility
func validateVersions(cfg *Config, result *ValidationResult) {
	// Check if PHP version is older than recommended
	if cfg.DDEV.PHPVersion == "7.4" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "ddev.php_version",
			Message:  "PHP 7.4 is end-of-life, consider upgrading to 8.1 or later",
			Severity: SeverityWarning,
			Fix:      "Update to PHP 8.1, 8.2, or 8.3",
		})
	}

	// Check if Xdebug is enabled (performance warning)
	if cfg.DDEV.XdebugEnabled {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "ddev.xdebug_enabled",
			Message:  "Xdebug is enabled, which may impact performance",
			Severity: SeverityWarning,
			Fix:      "Disable Xdebug when not actively debugging: xdebug_enabled: false",
		})
	}

	// Check MySQL version
	if cfg.DDEV.MySQLVersion == "5.7" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "ddev.mysql_version",
			Message:  "MySQL 5.7 is approaching end-of-life",
			Severity: SeverityWarning,
			Fix:      "Consider upgrading to MySQL 8.0",
		})
	}
}

// validateWPEngine validates WPEngine-specific configuration
func validateWPEngine(cfg *Config, result *ValidationResult) {
	sshGateway, _ := cfg.ProviderConfig["ssh_gateway"].(string)
	if sshGateway == "" || sshGateway == "ssh.wpengine.net" {
		result.Infos = append(result.Infos, ValidationError{
			Field:    "provider_config.ssh_gateway",
			Message:  "using default SSH gateway (ssh.wpengine.net)",
			Severity: SeverityInfo,
			Fix:      "This is fine - the default gateway will be used",
		})
	}

	install, _ := cfg.ProviderConfig["install"].(string)
	if install != "" && strings.Contains(install, ".wpengine.com") {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "provider_config.install",
			Message:  "should be the install name only, not the full domain",
			Severity: SeverityError,
			Fix:      fmt.Sprintf("Use '%s' instead of '%s'", strings.Split(install, ".")[0], install),
		})
	}

	if cfg.Snapshots.AutoSnapshotBeforePull {
		result.Infos = append(result.Infos, ValidationError{
			Field:    "snapshots.auto_snapshot_before_pull",
			Message:  "automatic snapshots enabled - backups will be created before database pulls",
			Severity: SeverityInfo,
			Fix:      "",
		})
	}
}

// validateNetwork validates network-specific configuration
func validateNetwork(cfg *Config, result *ValidationResult) {
	// Warn if multisite but no sites configured
	if cfg.Project.Type == "wordpress-multisite" && len(cfg.Network.Sites) == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "network.sites",
			Message:  "no sites configured - sites will be auto-detected from database",
			Severity: SeverityWarning,
			Fix:      "This is fine if sites already exist in the database",
		})
	}

	// Check if network title is set for multisite
	if cfg.Project.Type == "wordpress-multisite" && cfg.Network.Title == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "network.title",
			Message:  "network title not set",
			Severity: SeverityWarning,
			Fix:      "Add: network:\n      title: My Network",
		})
	}

	// Check if admin email is set
	if cfg.Project.Type == "wordpress-multisite" && cfg.Network.AdminEmail == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "network.admin_email",
			Message:  "admin email not set",
			Severity: SeverityWarning,
			Fix:      "Add: network:\n      admin_email: admin@example.local",
		})
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isValidDomain(domain string) bool {
	// Simple domain validation regex
	matched, _ := regexp.MatchString(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}$`, domain)
	return matched
}
