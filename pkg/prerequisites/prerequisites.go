// Package prerequisites provides dependency checking for Stax.
// It detects whether required tools (Docker, DDEV, Git, etc.) are installed
// and provides clear installation guidance when they are missing.
package prerequisites

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Dependency represents a required system dependency
type Dependency struct {
	Name        string // Human-readable name
	Command     string // Command to check existence (e.g., "docker")
	VersionFlag string // Flag to get version (e.g., "--version")
	MinVersion  string // Minimum required version (semver-ish)
	InstallURL  string // Documentation URL
	InstallCmd  string // Suggested install command (informational)
	Required    bool   // Is this strictly required for Stax to work?
	Description string // Brief description of what this dependency does
}

// CheckResult contains the result of checking a dependency
type CheckResult struct {
	Dependency Dependency
	Installed  bool
	Version    string
	Error      error
	Message    string // Human-readable status message
}

// OK returns true if the dependency check passed
func (r *CheckResult) OK() bool {
	return r.Installed && r.Error == nil
}

// DefaultDependencies returns the list of Stax prerequisites
func DefaultDependencies() []Dependency {
	return []Dependency{
		{
			Name:        "Docker",
			Command:     "docker",
			VersionFlag: "--version",
			MinVersion:  "20.0.0",
			InstallURL:  "https://docs.docker.com/desktop/install/mac-install/",
			InstallCmd:  "Download Docker Desktop from https://docker.com/products/docker-desktop",
			Required:    true,
			Description: "Container runtime for local development environments",
		},
		{
			Name:        "DDEV",
			Command:     "ddev",
			VersionFlag: "--version",
			MinVersion:  "1.22.0",
			InstallURL:  "https://ddev.readthedocs.io/en/stable/users/install/",
			InstallCmd:  "brew install ddev/ddev/ddev",
			Required:    true,
			Description: "Local development environment manager for WordPress",
		},
		{
			Name:        "Git",
			Command:     "git",
			VersionFlag: "--version",
			MinVersion:  "2.0.0",
			InstallURL:  "https://git-scm.com/download/mac",
			InstallCmd:  "brew install git",
			Required:    true,
			Description: "Version control system",
		},
		{
			Name:        "GitHub CLI",
			Command:     "gh",
			VersionFlag: "--version",
			MinVersion:  "2.0.0",
			InstallURL:  "https://cli.github.com/",
			InstallCmd:  "brew install gh",
			Required:    false,
			Description: "GitHub command-line tool (for repo and actions setup)",
		},
	}
}

// RequiredOnly returns only the required dependencies
func RequiredOnly() []Dependency {
	var required []Dependency
	for _, dep := range DefaultDependencies() {
		if dep.Required {
			required = append(required, dep)
		}
	}
	return required
}

// Check verifies a single dependency is installed and meets version requirements
func (d *Dependency) Check() *CheckResult {
	result := &CheckResult{
		Dependency: *d,
	}

	// Check if command exists
	path, err := exec.LookPath(d.Command)
	if err != nil {
		result.Installed = false
		result.Error = fmt.Errorf("command not found: %s", d.Command)
		result.Message = fmt.Sprintf("%s is not installed", d.Name)
		return result
	}

	result.Installed = true

	// Get version
	if d.VersionFlag != "" {
		cmd := exec.Command(path, d.VersionFlag)
		output, err := cmd.Output()
		if err != nil {
			result.Error = fmt.Errorf("failed to get version: %w", err)
			result.Message = fmt.Sprintf("%s is installed but version check failed", d.Name)
			return result
		}

		version := extractVersion(string(output))
		result.Version = version

		// Compare versions if we have both
		if version != "" && d.MinVersion != "" {
			if !isVersionSufficient(version, d.MinVersion) {
				result.Error = fmt.Errorf("version %s is below minimum %s", version, d.MinVersion)
				result.Message = fmt.Sprintf("%s %s is installed but version %s or higher is required",
					d.Name, version, d.MinVersion)
				return result
			}
		}
	}

	result.Message = fmt.Sprintf("%s %s", d.Name, result.Version)
	return result
}

// CheckAll verifies all provided dependencies
func CheckAll(deps []Dependency) []CheckResult {
	results := make([]CheckResult, len(deps))
	for i, dep := range deps {
		results[i] = *dep.Check()
	}
	return results
}

// FilterFailed returns only the check results that failed
func FilterFailed(results []CheckResult) []CheckResult {
	var failed []CheckResult
	for _, r := range results {
		if !r.OK() {
			failed = append(failed, r)
		}
	}
	return failed
}

// FilterRequired returns only the results for required dependencies
func FilterRequired(results []CheckResult) []CheckResult {
	var required []CheckResult
	for _, r := range results {
		if r.Dependency.Required {
			required = append(required, r)
		}
	}
	return required
}

// HasFailedRequired returns true if any required dependency failed
func HasFailedRequired(results []CheckResult) bool {
	for _, r := range results {
		if r.Dependency.Required && !r.OK() {
			return true
		}
	}
	return false
}

// extractVersion attempts to extract a version number from command output
func extractVersion(output string) string {
	// Common version patterns:
	// "Docker version 24.0.6, build ed223bc"
	// "ddev version v1.22.4"
	// "git version 2.42.0"
	// "gh version 2.40.1 (2023-12-13)"

	patterns := []string{
		`v?(\d+\.\d+\.\d+)`,           // Standard semver
		`version\s+v?(\d+\.\d+\.\d+)`, // "version X.Y.Z"
		`(\d+\.\d+\.\d+)`,             // Any X.Y.Z
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return strings.TrimSpace(output)
}

// isVersionSufficient checks if version >= minVersion (simple semver comparison)
func isVersionSufficient(version, minVersion string) bool {
	// Parse versions into components
	vParts := parseVersion(version)
	mParts := parseVersion(minVersion)

	// Compare major.minor.patch
	for i := 0; i < 3; i++ {
		v := 0
		m := 0
		if i < len(vParts) {
			v = vParts[i]
		}
		if i < len(mParts) {
			m = mParts[i]
		}

		if v > m {
			return true
		}
		if v < m {
			return false
		}
	}

	return true // versions are equal
}

// parseVersion extracts numeric components from a version string
func parseVersion(version string) []int {
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindAllString(version, 3)

	parts := make([]int, len(matches))
	for i, m := range matches {
		var n int
		fmt.Sscanf(m, "%d", &n)
		parts[i] = n
	}

	return parts
}

// GetDependency returns a specific dependency by name
func GetDependency(name string) *Dependency {
	for _, dep := range DefaultDependencies() {
		if strings.EqualFold(dep.Name, name) || strings.EqualFold(dep.Command, name) {
			return &dep
		}
	}
	return nil
}
