package diagnostics

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/git"
	"github.com/firecrown-media/stax/pkg/system"
)

// CheckGit checks if Git is installed and configured
func CheckGit() CheckResult {
	if !git.IsGitAvailable() {
		return CheckResult{
			Name:       "Git Installation",
			Category:   "System Requirements",
			Status:     StatusFail,
			Message:    "Git is not installed",
			Suggestion: "Install Git: https://git-scm.com/downloads",
		}
	}

	version, err := git.GetGitVersion()
	if err != nil {
		return CheckResult{
			Name:       "Git Installation",
			Category:   "System Requirements",
			Status:     StatusWarning,
			Message:    "Git is installed but version could not be determined",
			Suggestion: "Verify Git installation: git --version",
		}
	}

	return CheckResult{
		Name:     "Git Installation",
		Category: "System Requirements",
		Status:   StatusPass,
		Message:  fmt.Sprintf("Git version %s installed", version),
		Details: map[string]string{
			"version": version,
		},
	}
}

// CheckDocker checks if Docker is installed and running
func CheckDocker() CheckResult {
	info, err := system.GetDockerInfo()
	if err != nil {
		return CheckResult{
			Name:       "Docker",
			Category:   "System Requirements",
			Status:     StatusFail,
			Message:    "Failed to get Docker information",
			Suggestion: "Verify Docker installation",
		}
	}

	if !info.Installed {
		return CheckResult{
			Name:       "Docker",
			Category:   "System Requirements",
			Status:     StatusFail,
			Message:    "Docker is not installed",
			Suggestion: "Install Docker Desktop: https://www.docker.com/products/docker-desktop",
		}
	}

	if !info.Running {
		return CheckResult{
			Name:       "Docker",
			Category:   "System Requirements",
			Status:     StatusFail,
			Message:    "Docker is installed but not running",
			Suggestion: "Start Docker Desktop application",
			CanAutoFix: true,
			Details: map[string]string{
				"version": info.Version,
			},
		}
	}

	details := map[string]string{
		"version": info.Version,
		"running": "yes",
	}

	if info.ComposeInstalled {
		details["compose_version"] = info.ComposeVersion
	}

	return CheckResult{
		Name:     "Docker",
		Category: "System Requirements",
		Status:   StatusPass,
		Message:  fmt.Sprintf("Docker %s is running", info.Version),
		Details:  details,
	}
}

// CheckDDEV checks if DDEV is installed and configured
func CheckDDEV() CheckResult {
	if !ddev.IsInstalled() {
		return CheckResult{
			Name:       "DDEV",
			Category:   "System Requirements",
			Status:     StatusFail,
			Message:    "DDEV is not installed",
			Suggestion: "Install DDEV: https://ddev.readthedocs.io/en/stable/users/install/",
		}
	}

	version, err := ddev.GetVersion()
	if err != nil {
		return CheckResult{
			Name:       "DDEV",
			Category:   "System Requirements",
			Status:     StatusWarning,
			Message:    "DDEV is installed but version could not be determined",
			Suggestion: "Verify DDEV installation: ddev version",
		}
	}

	return CheckResult{
		Name:     "DDEV",
		Category: "System Requirements",
		Status:   StatusPass,
		Message:  fmt.Sprintf("DDEV version %s installed", version),
		Details: map[string]string{
			"version": version,
		},
	}
}

// CheckGo checks if Go is installed (optional, for development)
func CheckGo() CheckResult {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return CheckResult{
			Name:       "Go Installation",
			Category:   "System Requirements",
			Status:     StatusSkip,
			Message:    "Go is not installed (optional for development)",
			Suggestion: "Install Go if you plan to develop Stax: https://golang.org/dl/",
		}
	}

	version := strings.TrimSpace(string(output))
	return CheckResult{
		Name:     "Go Installation",
		Category: "System Requirements",
		Status:   StatusPass,
		Message:  version,
		Details: map[string]string{
			"version": strings.TrimPrefix(version, "go version "),
		},
	}
}

// CheckMemory checks available system memory
func CheckMemory() CheckResult {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return CheckResult{
			Name:       "Available Memory",
			Category:   "System Requirements",
			Status:     StatusWarning,
			Message:    "Cannot determine available memory",
			Suggestion: "Check system configuration",
		}
	}

	var totalBytes int64
	// If parsing fails, totalBytes remains 0 and will trigger the warning below
	_, _ = fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &totalBytes)
	totalGB := float64(totalBytes) / (1024 * 1024 * 1024)

	if totalGB < 4 {
		return CheckResult{
			Name:       "Available Memory",
			Category:   "System Requirements",
			Status:     StatusFail,
			Message:    fmt.Sprintf("Low memory: %.1f GB total", totalGB),
			Suggestion: "DDEV requires at least 4GB of RAM. Consider upgrading your system.",
			Details: map[string]string{
				"total": fmt.Sprintf("%.1f GB", totalGB),
			},
		}
	}

	if totalGB < 8 {
		return CheckResult{
			Name:       "Available Memory",
			Category:   "System Requirements",
			Status:     StatusWarning,
			Message:    fmt.Sprintf("%.1f GB total memory", totalGB),
			Suggestion: "8GB+ recommended for optimal performance",
			Details: map[string]string{
				"total": fmt.Sprintf("%.1f GB", totalGB),
			},
		}
	}

	return CheckResult{
		Name:     "Available Memory",
		Category: "System Requirements",
		Status:   StatusPass,
		Message:  fmt.Sprintf("%.1f GB total memory", totalGB),
		Details: map[string]string{
			"total": fmt.Sprintf("%.1f GB", totalGB),
		},
	}
}

// CheckRequiredCommands checks if required command-line tools are available
func CheckRequiredCommands() CheckResult {
	requiredCmds := []string{"git", "ssh", "rsync"}
	missing := []string{}

	for _, cmd := range requiredCmds {
		if _, err := exec.LookPath(cmd); err != nil {
			missing = append(missing, cmd)
		}
	}

	if len(missing) > 0 {
		return CheckResult{
			Name:       "Required Commands",
			Category:   "System Requirements",
			Status:     StatusFail,
			Message:    fmt.Sprintf("Missing required commands: %s", strings.Join(missing, ", ")),
			Suggestion: fmt.Sprintf("Install missing commands: %s", strings.Join(missing, ", ")),
			Details: map[string]string{
				"missing": strings.Join(missing, ", "),
			},
		}
	}

	return CheckResult{
		Name:     "Required Commands",
		Category: "System Requirements",
		Status:   StatusPass,
		Message:  "All required commands available",
		Details: map[string]string{
			"commands": strings.Join(requiredCmds, ", "),
		},
	}
}
