package diagnostics

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/firecrown-media/stax/pkg/ddev"
)

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
