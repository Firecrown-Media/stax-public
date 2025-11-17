package system

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// DockerInfo contains Docker system information
type DockerInfo struct {
	Installed        bool
	Version          string
	Running          bool
	ComposeVersion   string
	ComposeInstalled bool
}

// IsDockerAvailable checks if Docker is installed and available
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}

// IsDockerRunning checks if Docker daemon is running
func IsDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
}

// GetDockerVersion returns the Docker version
func GetDockerVersion() (string, error) {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get docker version: %w", err)
	}

	version := strings.TrimSpace(string(output))

	// Extract version number (e.g., "Docker version 24.0.5, build ced0996")
	re := regexp.MustCompile(`Docker version ([0-9.]+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return version, nil
}

// IsDockerComposeAvailable checks if Docker Compose is installed
func IsDockerComposeAvailable() bool {
	// Try docker compose (v2)
	cmd := exec.Command("docker", "compose", "version")
	if cmd.Run() == nil {
		return true
	}

	// Try docker-compose (v1)
	cmd = exec.Command("docker-compose", "--version")
	return cmd.Run() == nil
}

// GetDockerComposeVersion returns the Docker Compose version
func GetDockerComposeVersion() (string, error) {
	// Try docker compose (v2) first
	cmd := exec.Command("docker", "compose", "version")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))

		// Extract version number (e.g., "Docker Compose version v2.20.0")
		re := regexp.MustCompile(`version v?([0-9.]+)`)
		matches := re.FindStringSubmatch(version)
		if len(matches) > 1 {
			return "v2." + matches[1], nil
		}

		return version, nil
	}

	// Try docker-compose (v1)
	cmd = exec.Command("docker-compose", "--version")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("docker compose is not installed")
	}

	version := strings.TrimSpace(string(output))

	// Extract version number (e.g., "docker-compose version 1.29.2")
	re := regexp.MustCompile(`version ([0-9.]+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) > 1 {
		return "v1." + matches[1], nil
	}

	return version, nil
}

// GetDockerInfo returns comprehensive Docker system information
func GetDockerInfo() (*DockerInfo, error) {
	info := &DockerInfo{}

	// Check if Docker is installed
	info.Installed = IsDockerAvailable()
	if !info.Installed {
		return info, nil
	}

	// Get Docker version
	version, err := GetDockerVersion()
	if err == nil {
		info.Version = version
	}

	// Check if Docker is running
	info.Running = IsDockerRunning()

	// Check if Docker Compose is installed
	info.ComposeInstalled = IsDockerComposeAvailable()
	if info.ComposeInstalled {
		composeVersion, err := GetDockerComposeVersion()
		if err == nil {
			info.ComposeVersion = composeVersion
		}
	}

	return info, nil
}

// GetRunningContainers returns a list of running Docker containers
func GetRunningContainers() ([]string, error) {
	if !IsDockerRunning() {
		return nil, fmt.Errorf("docker is not running")
	}

	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var containers []string
	for _, line := range lines {
		if line != "" {
			containers = append(containers, line)
		}
	}

	return containers, nil
}

// IsContainerRunning checks if a specific container is running
func IsContainerRunning(containerName string) (bool, error) {
	if !IsDockerRunning() {
		return false, fmt.Errorf("docker is not running")
	}

	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check container status: %w", err)
	}

	return strings.TrimSpace(string(output)) != "", nil
}

// GetDockerDiskUsage returns Docker disk usage information
func GetDockerDiskUsage() (map[string]string, error) {
	if !IsDockerRunning() {
		return nil, fmt.Errorf("docker is not running")
	}

	cmd := exec.Command("docker", "system", "df", "--format", "{{.Type}}\t{{.Size}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	usage := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) == 2 {
			usage[parts[0]] = parts[1]
		}
	}

	return usage, nil
}

// ValidateDockerRequirements checks if Docker meets minimum requirements
func ValidateDockerRequirements() error {
	if !IsDockerAvailable() {
		return fmt.Errorf("docker is not installed")
	}

	if !IsDockerRunning() {
		return fmt.Errorf("docker is not running")
	}

	// Get Docker version
	version, err := GetDockerVersion()
	if err != nil {
		return fmt.Errorf("failed to get docker version: %w", err)
	}

	// Check minimum version (20.10.0 or higher)
	if !isVersionAtLeast(version, "20.10.0") {
		return fmt.Errorf("docker version %s is too old (minimum: 20.10.0)", version)
	}

	// Check Docker Compose
	if !IsDockerComposeAvailable() {
		return fmt.Errorf("docker compose is not installed")
	}

	return nil
}

// isVersionAtLeast checks if version is at least minVersion
func isVersionAtLeast(version, minVersion string) bool {
	// Simple version comparison (works for semver)
	vParts := strings.Split(version, ".")
	minParts := strings.Split(minVersion, ".")

	for i := 0; i < len(minParts) && i < len(vParts); i++ {
		var v, minV int
		// If parsing fails, treat that part as 0 (malformed version)
		_, _ = fmt.Sscanf(vParts[i], "%d", &v)
		_, _ = fmt.Sscanf(minParts[i], "%d", &minV)

		if v > minV {
			return true
		}
		if v < minV {
			return false
		}
	}

	return len(vParts) >= len(minParts)
}
