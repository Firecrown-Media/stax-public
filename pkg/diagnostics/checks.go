package diagnostics

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/git"
	"github.com/firecrown-media/stax/pkg/system"
)

// CheckResult represents the result of a diagnostic check
type CheckResult struct {
	Name       string
	Status     CheckStatus
	Message    string
	Suggestion string
	Details    map[string]string
	Category   string // Category for grouping checks
	CanAutoFix bool   // Whether this check can be auto-fixed
	FixApplied bool   // Whether a fix was applied
}

// CheckStatus represents the status of a check
type CheckStatus string

const (
	StatusPass    CheckStatus = "pass"
	StatusWarning CheckStatus = "warning"
	StatusFail    CheckStatus = "fail"
	StatusSkip    CheckStatus = "skip"
)

// DiagnosticReport contains all diagnostic check results
type DiagnosticReport struct {
	Checks      []CheckResult
	Summary     Summary
	ProjectPath string
	Verbose     bool
	AutoFix     bool
	Categories  map[string][]CheckResult // Checks grouped by category
}

// Summary provides a summary of check results
type Summary struct {
	Total    int
	Passed   int
	Warnings int
	Failed   int
	Skipped  int
	Fixed    int
}

// RunAllChecks runs all diagnostic checks
func RunAllChecks(projectPath string, verbose bool, autoFix bool) (*DiagnosticReport, error) {
	report := &DiagnosticReport{
		ProjectPath: projectPath,
		Checks:      []CheckResult{},
		Verbose:     verbose,
		AutoFix:     autoFix,
		Categories:  make(map[string][]CheckResult),
	}

	// System requirements checks
	report.Checks = append(report.Checks, CheckGit())
	report.Checks = append(report.Checks, CheckDocker())
	report.Checks = append(report.Checks, CheckDDEV())
	report.Checks = append(report.Checks, CheckMemory())
	report.Checks = append(report.Checks, CheckRequiredCommands())
	report.Checks = append(report.Checks, CheckGo())

	// Project configuration checks
	report.Checks = append(report.Checks, CheckStaxConfig(projectPath))
	report.Checks = append(report.Checks, CheckDDEVConfig(projectPath))

	// Credential checks
	report.Checks = append(report.Checks, CheckCredentials(projectPath))
	report.Checks = append(report.Checks, CheckSSHKey())
	report.Checks = append(report.Checks, CheckGitHubToken())

	// Network checks
	report.Checks = append(report.Checks, CheckPorts())
	report.Checks = append(report.Checks, CheckWPEngineAPI())
	report.Checks = append(report.Checks, CheckWPEngineSSH())
	report.Checks = append(report.Checks, CheckGitHubAPI())
	report.Checks = append(report.Checks, CheckInternetConnectivity())

	// Environment checks
	report.Checks = append(report.Checks, CheckDiskSpace(projectPath))

	// Service health checks - with auto-fix support
	ddevStatus := CheckDDEVStatus(projectPath)
	if autoFix && ddevStatus.CanAutoFix && (ddevStatus.Status == StatusWarning || ddevStatus.Status == StatusFail) {
		ddevStatus = FixDDEVStatus(projectPath, ddevStatus)
	}
	report.Checks = append(report.Checks, ddevStatus)

	report.Checks = append(report.Checks, CheckDatabaseConnectivity(projectPath))
	report.Checks = append(report.Checks, CheckWordPressInstallation(projectPath))

	// Group checks by category
	report.groupByCategory()

	// Calculate summary
	report.Summary = calculateSummary(report.Checks)

	return report, nil
}

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

// CheckCredentials checks if credentials are properly configured
func CheckCredentials(projectPath string) CheckResult {
	diag := credentials.RunDiagnostics()

	// Analyze diagnostic results
	if diag.OverallStatus == "error" {
		return CheckResult{
			Name:       "Credentials",
			Category:   "Credentials",
			Status:     StatusFail,
			Message:    "Credential configuration has errors",
			Suggestion: "Run: stax setup --check for details",
		}
	}

	if diag.OverallStatus == "warning" {
		return CheckResult{
			Name:       "Credentials",
			Category:   "Credentials",
			Status:     StatusWarning,
			Message:    "Credential configuration has warnings",
			Suggestion: "Run: stax setup --check for details",
		}
	}

	return CheckResult{
		Name:     "Credentials",
		Category: "Credentials",
		Status:   StatusPass,
		Message:  "All credentials are properly configured",
	}
}

// CheckSSHKey checks if SSH key is available and properly configured
func CheckSSHKey() CheckResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return CheckResult{
			Name:       "SSH Key",
			Category:   "Credentials",
			Status:     StatusFail,
			Message:    "Cannot determine home directory",
			Suggestion: "Check your system configuration",
		}
	}

	// Check for common SSH key types
	keyPaths := []string{
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
	}

	var foundKey string
	var keyInfo os.FileInfo
	for _, keyPath := range keyPaths {
		info, err := os.Stat(keyPath)
		if err == nil {
			foundKey = keyPath
			keyInfo = info
			break
		}
	}

	if foundKey == "" {
		return CheckResult{
			Name:       "SSH Key",
			Category:   "Credentials",
			Status:     StatusWarning,
			Message:    "No SSH key found",
			Suggestion: "Generate SSH key: ssh-keygen -t ed25519 -C \"your_email@example.com\"",
			Details: map[string]string{
				"checked_paths": strings.Join(keyPaths, ", "),
			},
		}
	}

	// Check permissions
	mode := keyInfo.Mode()
	if mode.Perm()&0077 != 0 {
		return CheckResult{
			Name:       "SSH Key",
			Category:   "Credentials",
			Status:     StatusWarning,
			Message:    "SSH key has insecure permissions",
			Suggestion: fmt.Sprintf("Fix permissions: chmod 600 %s", foundKey),
			Details: map[string]string{
				"path":        foundKey,
				"permissions": fmt.Sprintf("%o", mode.Perm()),
			},
		}
	}

	// Check if public key exists
	pubKeyPath := foundKey + ".pub"
	if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
		return CheckResult{
			Name:       "SSH Key",
			Category:   "Credentials",
			Status:     StatusWarning,
			Message:    "SSH public key not found",
			Suggestion: fmt.Sprintf("Generate public key: ssh-keygen -y -f %s > %s", foundKey, pubKeyPath),
			Details: map[string]string{
				"private_key": foundKey,
			},
		}
	}

	return CheckResult{
		Name:     "SSH Key",
		Category: "Credentials",
		Status:   StatusPass,
		Message:  "SSH key found and properly configured",
		Details: map[string]string{
			"path":        foundKey,
			"permissions": "0600",
			"public_key":  pubKeyPath,
		},
	}
}

// CheckGitHubToken checks if GitHub token is configured
func CheckGitHubToken() CheckResult {
	token, err := credentials.GetGitHubToken("default")
	if err != nil || token == "" {
		return CheckResult{
			Name:       "GitHub Token",
			Category:   "Credentials",
			Status:     StatusSkip,
			Message:    "GitHub token not configured (optional)",
			Suggestion: "Configure token if needed for private repos: stax setup github",
		}
	}

	return CheckResult{
		Name:     "GitHub Token",
		Category: "Credentials",
		Status:   StatusPass,
		Message:  "GitHub token configured",
		Details: map[string]string{
			"token_length": fmt.Sprintf("%d characters", len(token)),
		},
	}
}

// CheckPorts checks if required ports are available
func CheckPorts() CheckResult {
	defaultPorts := system.DefaultDDEVPorts()
	inUse, err := system.CheckRequiredPorts(defaultPorts)
	if err != nil {
		recommendations := system.RecommendedPorts(defaultPorts)
		var suggestions []string
		for original, recommended := range recommendations {
			suggestions = append(suggestions, fmt.Sprintf("Port %d -> %d", original, recommended))
		}

		return CheckResult{
			Name:       "Port Availability",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    fmt.Sprintf("Some required ports are in use: %v", inUse),
			Suggestion: fmt.Sprintf("Consider using alternative ports:\n%s", strings.Join(suggestions, "\n")),
			Details: map[string]string{
				"ports_in_use": fmt.Sprintf("%v", inUse),
			},
		}
	}

	return CheckResult{
		Name:     "Port Availability",
		Category: "Network Connectivity",
		Status:   StatusPass,
		Message:  "All required ports are available",
		Details: map[string]string{
			"checked_ports": fmt.Sprintf("%v", defaultPorts),
		},
	}
}

// CheckWPEngineAPI checks WPEngine API connectivity
func CheckWPEngineAPI() CheckResult {
	creds, err := credentials.GetWPEngineCredentials("default")
	if err != nil {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "WPEngine credentials not configured",
			Suggestion: "Configure credentials: stax setup wpengine",
		}
	}

	// Try to make a simple API request to test connectivity
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.wpengineapi.com/v1/installs", nil)
	if err != nil {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "Failed to create API request",
			Suggestion: "Check your network connection",
		}
	}

	req.SetBasicAuth(creds.APIUser, creds.APIPassword)
	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "Cannot reach WPEngine API",
			Suggestion: "Check your internet connection and firewall settings",
			Details: map[string]string{
				"error": err.Error(),
			},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusFail,
			Message:    "WPEngine API credentials are invalid",
			Suggestion: "Reconfigure credentials: stax setup wpengine",
			Details: map[string]string{
				"status_code": fmt.Sprintf("%d", resp.StatusCode),
			},
		}
	}

	if resp.StatusCode != 200 {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    fmt.Sprintf("WPEngine API returned status %d", resp.StatusCode),
			Suggestion: "Check WPEngine API status",
			Details: map[string]string{
				"status_code": fmt.Sprintf("%d", resp.StatusCode),
			},
		}
	}

	return CheckResult{
		Name:     "WPEngine API",
		Category: "Network Connectivity",
		Status:   StatusPass,
		Message:  "WPEngine API is reachable and credentials are valid",
		Details: map[string]string{
			"api_user": creds.APIUser,
		},
	}
}

// CheckWPEngineSSH checks WPEngine SSH gateway connectivity
func CheckWPEngineSSH() CheckResult {
	creds, err := credentials.GetWPEngineCredentials("default")
	if err != nil {
		return CheckResult{
			Name:       "WPEngine SSH Gateway",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "WPEngine credentials not configured",
			Suggestion: "Configure credentials: stax setup wpengine",
		}
	}

	gateway := creds.SSHGateway
	if gateway == "" {
		gateway = "ssh.wpengine.net"
	}

	// Test SSH gateway connectivity (port 22)
	conn, err := net.DialTimeout("tcp", gateway+":22", 5*time.Second)
	if err != nil {
		return CheckResult{
			Name:       "WPEngine SSH Gateway",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    fmt.Sprintf("Cannot reach SSH gateway: %s", gateway),
			Suggestion: "Check your internet connection and firewall settings",
			Details: map[string]string{
				"gateway": gateway,
				"error":   err.Error(),
			},
		}
	}
	conn.Close()

	return CheckResult{
		Name:     "WPEngine SSH Gateway",
		Category: "Network Connectivity",
		Status:   StatusPass,
		Message:  "SSH gateway is reachable",
		Details: map[string]string{
			"gateway": gateway,
			"port":    "22",
		},
	}
}

// CheckInternetConnectivity checks basic internet connectivity
func CheckInternetConnectivity() CheckResult {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Try to reach a few reliable endpoints
	endpoints := []string{
		"https://www.google.com",
		"https://www.cloudflare.com",
		"https://api.github.com",
	}

	var lastErr error
	for _, endpoint := range endpoints {
		resp, err := client.Get(endpoint)
		if err == nil {
			resp.Body.Close()
			return CheckResult{
				Name:     "Internet Connectivity",
				Category: "Network Connectivity",
				Status:   StatusPass,
				Message:  "Internet connection is working",
				Details: map[string]string{
					"tested_endpoint": endpoint,
				},
			}
		}
		lastErr = err
	}

	return CheckResult{
		Name:       "Internet Connectivity",
		Category:   "Network Connectivity",
		Status:     StatusWarning,
		Message:    "Cannot reach internet",
		Suggestion: "Check your network connection",
		Details: map[string]string{
			"error": lastErr.Error(),
		},
	}
}

// CheckDiskSpace checks available disk space with proper implementation
func CheckDiskSpace(projectPath string) CheckResult {
	if projectPath == "" || projectPath == "." {
		var err error
		projectPath, err = os.Getwd()
		if err != nil {
			return CheckResult{
				Name:       "Disk Space",
				Category:   "Environment",
				Status:     StatusWarning,
				Message:    "Cannot determine current directory",
				Suggestion: "Check your working directory",
			}
		}
	}

	var stat syscall.Statfs_t
	err := syscall.Statfs(projectPath, &stat)
	if err != nil {
		return CheckResult{
			Name:       "Disk Space",
			Category:   "Environment",
			Status:     StatusWarning,
			Message:    "Cannot check disk space",
			Suggestion: "Verify project path is accessible",
		}
	}

	// Calculate available space in GB
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	availableGB := float64(availableBytes) / (1024 * 1024 * 1024)

	// Calculate total space in GB
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	totalGB := float64(totalBytes) / (1024 * 1024 * 1024)

	// Calculate used percentage
	usedPercentage := float64(stat.Blocks-stat.Bavail) / float64(stat.Blocks) * 100

	if availableGB < 5 {
		return CheckResult{
			Name:       "Disk Space",
			Category:   "Environment",
			Status:     StatusFail,
			Message:    fmt.Sprintf("Low disk space: %.2f GB available", availableGB),
			Suggestion: "Free up disk space. At least 5GB recommended for DDEV projects",
			Details: map[string]string{
				"available": fmt.Sprintf("%.2f GB", availableGB),
				"total":     fmt.Sprintf("%.2f GB", totalGB),
				"used":      fmt.Sprintf("%.1f%%", usedPercentage),
			},
		}
	}

	if availableGB < 10 {
		return CheckResult{
			Name:       "Disk Space",
			Category:   "Environment",
			Status:     StatusWarning,
			Message:    fmt.Sprintf("Moderate disk space: %.2f GB available", availableGB),
			Suggestion: "Consider freeing up disk space. 10GB+ recommended",
			Details: map[string]string{
				"available": fmt.Sprintf("%.2f GB", availableGB),
				"total":     fmt.Sprintf("%.2f GB", totalGB),
				"used":      fmt.Sprintf("%.1f%%", usedPercentage),
			},
		}
	}

	return CheckResult{
		Name:     "Disk Space",
		Category: "Environment",
		Status:   StatusPass,
		Message:  fmt.Sprintf("%.2f GB available", availableGB),
		Details: map[string]string{
			"available": fmt.Sprintf("%.2f GB", availableGB),
			"total":     fmt.Sprintf("%.2f GB", totalGB),
			"used":      fmt.Sprintf("%.1f%%", usedPercentage),
		},
	}
}

// CheckDDEVStatus checks if DDEV project is running
func CheckDDEVStatus(projectPath string) CheckResult {
	// Check if .ddev directory exists
	ddevPath := filepath.Join(projectPath, ".ddev")
	if _, err := os.Stat(ddevPath); os.IsNotExist(err) {
		return CheckResult{
			Name:       "DDEV Status",
			Category:   "Service Health",
			Status:     StatusSkip,
			Message:    "DDEV not configured for this project",
			Suggestion: "Initialize DDEV: stax init",
		}
	}

	// Check if DDEV is running
	manager := ddev.NewManager(projectPath)
	running, err := manager.IsRunning()
	if err != nil {
		return CheckResult{
			Name:       "DDEV Status",
			Category:   "Service Health",
			Status:     StatusWarning,
			Message:    "Cannot determine DDEV status",
			Suggestion: "Check DDEV installation: ddev version",
		}
	}

	if !running {
		return CheckResult{
			Name:       "DDEV Status",
			Category:   "Service Health",
			Status:     StatusWarning,
			Message:    "DDEV project is not running",
			Suggestion: "Start DDEV: stax start",
			CanAutoFix: true,
		}
	}

	// Get detailed status
	status, err := manager.GetStatus()
	if err != nil {
		return CheckResult{
			Name:     "DDEV Status",
			Category: "Service Health",
			Status:   StatusPass,
			Message:  "DDEV project is running",
		}
	}

	return CheckResult{
		Name:     "DDEV Status",
		Category: "Service Health",
		Status:   StatusPass,
		Message:  fmt.Sprintf("DDEV project '%s' is running", status.ProjectName),
		Details: map[string]string{
			"project_name": status.ProjectName,
			"state":        status.State,
			"php_version":  status.PHPVersion,
			"db_version":   status.DBVersion,
		},
	}
}

// CheckDatabaseConnectivity checks if database is accessible
func CheckDatabaseConnectivity(projectPath string) CheckResult {
	// Check if DDEV is configured
	ddevPath := filepath.Join(projectPath, ".ddev")
	if _, err := os.Stat(ddevPath); os.IsNotExist(err) {
		return CheckResult{
			Name:     "Database Connectivity",
			Category: "Service Health",
			Status:   StatusSkip,
			Message:  "DDEV not configured",
		}
	}

	// Check if DDEV is running
	manager := ddev.NewManager(projectPath)
	running, err := manager.IsRunning()
	if err != nil || !running {
		return CheckResult{
			Name:       "Database Connectivity",
			Category:   "Service Health",
			Status:     StatusSkip,
			Message:    "DDEV is not running",
			Suggestion: "Start DDEV to check database: stax start",
		}
	}

	// Try to connect to database by running a simple query
	cmd := exec.Command("ddev", "mysql", "-e", "SELECT 1;")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CheckResult{
			Name:       "Database Connectivity",
			Category:   "Service Health",
			Status:     StatusFail,
			Message:    "Cannot connect to database",
			Suggestion: "Check DDEV database container: ddev describe",
			Details: map[string]string{
				"error": string(output),
			},
		}
	}

	return CheckResult{
		Name:     "Database Connectivity",
		Category: "Service Health",
		Status:   StatusPass,
		Message:  "Database is accessible",
	}
}

// CheckWordPressInstallation checks if WordPress is installed
func CheckWordPressInstallation(projectPath string) CheckResult {
	// Check if we're in a project directory
	if projectPath == "" || projectPath == "." {
		return CheckResult{
			Name:     "WordPress Installation",
			Category: "Service Health",
			Status:   StatusSkip,
			Message:  "Not in a project directory",
		}
	}

	// Check for wp-config.php
	wpConfigPath := filepath.Join(projectPath, "wp-config.php")
	if _, err := os.Stat(wpConfigPath); os.IsNotExist(err) {
		return CheckResult{
			Name:       "WordPress Installation",
			Category:   "Service Health",
			Status:     StatusWarning,
			Message:    "wp-config.php not found",
			Suggestion: "Install WordPress or initialize project: stax init",
		}
	}

	// Check for wp-content directory
	wpContentPath := filepath.Join(projectPath, "wp-content")
	if _, err := os.Stat(wpContentPath); os.IsNotExist(err) {
		return CheckResult{
			Name:       "WordPress Installation",
			Category:   "Service Health",
			Status:     StatusWarning,
			Message:    "wp-content directory not found",
			Suggestion: "Check WordPress installation",
		}
	}

	// Check if DDEV is running to test WP-CLI
	manager := ddev.NewManager(projectPath)
	running, _ := manager.IsRunning()
	if !running {
		return CheckResult{
			Name:     "WordPress Installation",
			Category: "Service Health",
			Status:   StatusPass,
			Message:  "WordPress files present (DDEV not running to verify)",
			Details: map[string]string{
				"wp-config":  "found",
				"wp-content": "found",
			},
		}
	}

	// Try WP-CLI to check installation
	cmd := exec.Command("ddev", "wp", "core", "is-installed")
	cmd.Dir = projectPath
	err := cmd.Run()
	if err != nil {
		return CheckResult{
			Name:       "WordPress Installation",
			Category:   "Service Health",
			Status:     StatusWarning,
			Message:    "WordPress is not fully installed",
			Suggestion: "Complete WordPress installation or run: ddev wp core install",
		}
	}

	// Get WordPress version
	cmd = exec.Command("ddev", "wp", "core", "version")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	version := "unknown"
	if err == nil {
		version = strings.TrimSpace(string(output))
	}

	return CheckResult{
		Name:     "WordPress Installation",
		Category: "Service Health",
		Status:   StatusPass,
		Message:  fmt.Sprintf("WordPress %s installed and configured", version),
		Details: map[string]string{
			"version": version,
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

// CheckGitHubAPI checks GitHub API accessibility
func CheckGitHubAPI() CheckResult {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://api.github.com")
	if err != nil {
		return CheckResult{
			Name:       "GitHub API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "Cannot reach GitHub API",
			Suggestion: "Check your internet connection and firewall settings",
			Details: map[string]string{
				"error": err.Error(),
			},
		}
	}
	defer resp.Body.Close()

	return CheckResult{
		Name:     "GitHub API",
		Category: "Network Connectivity",
		Status:   StatusPass,
		Message:  "GitHub API is accessible",
		Details: map[string]string{
			"status_code": fmt.Sprintf("%d", resp.StatusCode),
		},
	}
}

// FixDDEVStatus attempts to fix DDEV status issues
func FixDDEVStatus(projectPath string, originalCheck CheckResult) CheckResult {
	manager := ddev.NewManager(projectPath)

	// Try to start DDEV
	err := manager.Start()
	if err != nil {
		return CheckResult{
			Name:       originalCheck.Name,
			Category:   originalCheck.Category,
			Status:     StatusFail,
			Message:    "Failed to start DDEV",
			Suggestion: "Start DDEV manually: stax start",
			CanAutoFix: true,
			FixApplied: true,
			Details: map[string]string{
				"error": err.Error(),
			},
		}
	}

	// Wait for DDEV to be ready
	err = manager.WaitForReady(30 * time.Second)
	if err != nil {
		return CheckResult{
			Name:       originalCheck.Name,
			Category:   originalCheck.Category,
			Status:     StatusWarning,
			Message:    "DDEV started but not ready",
			Suggestion: "Wait for DDEV to fully start: ddev describe",
			CanAutoFix: true,
			FixApplied: true,
		}
	}

	// Get status to confirm
	status, err := manager.GetStatus()
	if err != nil {
		return CheckResult{
			Name:       originalCheck.Name,
			Category:   originalCheck.Category,
			Status:     StatusPass,
			Message:    "DDEV project started successfully",
			CanAutoFix: true,
			FixApplied: true,
		}
	}

	return CheckResult{
		Name:       originalCheck.Name,
		Category:   originalCheck.Category,
		Status:     StatusPass,
		Message:    fmt.Sprintf("DDEV project '%s' started successfully", status.ProjectName),
		CanAutoFix: true,
		FixApplied: true,
		Details: map[string]string{
			"project_name": status.ProjectName,
			"state":        status.State,
			"php_version":  status.PHPVersion,
		},
	}
}

// groupByCategory groups check results by category
func (r *DiagnosticReport) groupByCategory() {
	r.Categories = make(map[string][]CheckResult)
	for _, check := range r.Checks {
		category := check.Category
		if category == "" {
			category = "Other"
		}
		r.Categories[category] = append(r.Categories[category], check)
	}
}

// calculateSummary calculates the summary of check results
func calculateSummary(checks []CheckResult) Summary {
	summary := Summary{
		Total: len(checks),
	}

	for _, check := range checks {
		if check.FixApplied {
			summary.Fixed++
		}
		switch check.Status {
		case StatusPass:
			summary.Passed++
		case StatusWarning:
			summary.Warnings++
		case StatusFail:
			summary.Failed++
		case StatusSkip:
			summary.Skipped++
		}
	}

	return summary
}

// HasCriticalFailures returns true if there are any critical failures
func (r *DiagnosticReport) HasCriticalFailures() bool {
	return r.Summary.Failed > 0
}

// HasWarnings returns true if there are any warnings
func (r *DiagnosticReport) HasWarnings() bool {
	return r.Summary.Warnings > 0
}

// IsHealthy returns true if all checks passed
func (r *DiagnosticReport) IsHealthy() bool {
	return r.Summary.Failed == 0 && r.Summary.Warnings == 0
}
