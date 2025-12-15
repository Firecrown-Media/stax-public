package ddev

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Manager handles DDEV operations
type Manager struct {
	ProjectDir string
}

// NewManager creates a new DDEV manager
func NewManager(projectDir string) *Manager {
	return &Manager{
		ProjectDir: projectDir,
	}
}

// IsInstalled checks if DDEV is installed
func IsInstalled() bool {
	_, err := exec.LookPath("ddev")
	return err == nil
}

// GetVersion returns the DDEV version
func GetVersion() (string, error) {
	cmd := exec.Command("ddev", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get DDEV version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// IsRunning checks if a DDEV project is running
func (m *Manager) IsRunning() (bool, error) {
	cmd := exec.Command("ddev", "describe", "-j")
	cmd.Dir = m.ProjectDir

	output, err := cmd.Output()
	if err != nil {
		// If describe fails, project is not running
		return false, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return false, fmt.Errorf("failed to parse describe output: %w", err)
	}

	// Extract the 'raw' object which contains the actual project data
	raw, ok := result["raw"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected JSON format from DDEV: 'raw' field not found or invalid")
	}

	status, ok := raw["status"].(string)
	if !ok {
		return false, nil
	}

	return status == "running", nil
}

// Start starts the DDEV environment
func (m *Manager) Start() error {
	cmd := exec.Command("ddev", "start")
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start DDEV: %w", err)
	}

	return nil
}

// Stop stops the DDEV environment
func (m *Manager) Stop() error {
	cmd := exec.Command("ddev", "stop")
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop DDEV: %w", err)
	}

	return nil
}

// Restart restarts the DDEV environment
func (m *Manager) Restart() error {
	cmd := exec.Command("ddev", "restart")
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart DDEV: %w", err)
	}

	return nil
}

// Delete removes the DDEV project
func (m *Manager) Delete(removeData bool) error {
	args := []string{"delete", "-y"}
	if removeData {
		args = append(args, "--omit-snapshot")
	}

	cmd := exec.Command("ddev", args...)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete DDEV project: %w", err)
	}

	return nil
}

// GetStatus returns the detailed status of the DDEV environment
func (m *Manager) GetStatus() (*DDEVStatus, error) {
	cmd := exec.Command("ddev", "describe", "-j")
	cmd.Dir = m.ProjectDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get DDEV status: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse describe output: %w", err)
	}

	// Extract the 'raw' object which contains the actual project data
	raw, ok := result["raw"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected JSON format from DDEV: 'raw' field not found or invalid")
	}

	status := &DDEVStatus{
		ProjectName: getStringValue(raw, "name"),
		Type:        getStringValue(raw, "type"),
		Location:    getStringValue(raw, "shortroot"),
		State:       getStringValue(raw, "status"),
		URLs:        getURLs(raw),
		PHPVersion:  getStringValue(raw, "php_version"),
		RouterHTTP:  getStringValue(raw, "router_http_port"),
		RouterHTTPS: getStringValue(raw, "router_https_port"),
	}

	// Parse database version
	if dbType, ok := raw["dbinfo"].(map[string]interface{}); ok {
		status.DBVersion = getStringValue(dbType, "version")
	}

	// Parse services
	status.Services = parseServices(raw)

	return status, nil
}

// Describe returns detailed project information
func (m *Manager) Describe() (*ProjectInfo, error) {
	cmd := exec.Command("ddev", "describe", "-j")
	cmd.Dir = m.ProjectDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to describe DDEV project: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse describe output: %w", err)
	}

	// Extract the 'raw' object which contains the actual project data
	raw, ok := result["raw"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected JSON format from DDEV: 'raw' field not found or invalid")
	}

	urls := getURLs(raw)
	status := getStringValue(raw, "status")
	location := getStringValue(raw, "shortroot")

	info := &ProjectInfo{
		Name:            getStringValue(raw, "name"),
		Type:            getStringValue(raw, "type"),
		Location:        location,
		AppRoot:         location,
		URLs:            urls,
		PrimaryURL:      getPrimaryURL(urls),
		PHPVersion:      getStringValue(raw, "php_version"),
		RouterHTTPPort:  getStringValue(raw, "router_http_port"),
		RouterHTTPSPort: getStringValue(raw, "router_https_port"),
		Status:          status,
		Running:         status == "running",
		Healthy:         status == "running", // Simplified for now
		Hostnames:       getHostnames(raw),
		Services:        parseServices(raw),
		Router:          "ddev-router",
		RouterStatus:    status,
		Webserver:       "nginx-fpm",
		XdebugEnabled:   getBoolValue(raw, "xdebug_enabled"),
		MailhogURL:      getMailhogURL(urls),
	}

	// Parse database info
	if dbInfo, ok := raw["dbinfo"].(map[string]interface{}); ok {
		info.DatabaseType = getStringValue(dbInfo, "type")
		info.DatabaseVersion = getStringValue(dbInfo, "version")
	}

	return info, nil
}

// Exec executes a command in the DDEV container
func (m *Manager) Exec(command []string, options *ExecOptions) error {
	if options == nil {
		options = &ExecOptions{Service: "web"}
	}
	if options.Service == "" {
		options.Service = "web"
	}

	args := []string{"exec"}

	if options.Service != "" && options.Service != "web" {
		args = append(args, "-s", options.Service)
	}

	if options.Dir != "" {
		args = append(args, "-d", options.Dir)
	}

	args = append(args, command...)

	cmd := exec.Command("ddev", args...)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set environment variables
	if len(options.Environment) > 0 {
		cmd.Env = append(os.Environ(), options.Environment...)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

// ExecCommand executes a DDEV command (like "xdebug on")
func (m *Manager) ExecCommand(args ...string) error {
	cmd := exec.Command("ddev", args...)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute DDEV command: %w", err)
	}

	return nil
}

// Logs retrieves or tails logs from DDEV services
func (m *Manager) Logs(options *LogOptions) error {
	if options == nil {
		options = &LogOptions{Follow: false}
	}

	args := []string{"logs"}

	if options.Service != "" {
		args = append(args, "-s", options.Service)
	}

	if options.Follow {
		args = append(args, "-f")
	}

	if options.Tail > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", options.Tail))
	}

	if options.Timestamps {
		args = append(args, "--timestamps")
	}

	if options.Since != "" {
		args = append(args, "--since", options.Since)
	}

	cmd := exec.Command("ddev", args...)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}

// SSH opens an SSH session to the DDEV web container
func (m *Manager) SSH(service string) error {
	if service == "" {
		service = "web"
	}

	args := []string{"ssh"}
	if service != "web" {
		args = append(args, "-s", service)
	}

	cmd := exec.Command("ddev", args...)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open SSH session: %w", err)
	}

	return nil
}

// ImportDB imports a database file
func (m *Manager) ImportDB(dbPath string) error {
	cmd := exec.Command("ddev", "import-db", "--file="+dbPath)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to import database: %w", err)
	}

	return nil
}

// ExportDB exports the database to a file
func (m *Manager) ExportDB(outputPath string) error {
	cmd := exec.Command("ddev", "export-db", "-f", outputPath)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to export database: %w", err)
	}

	return nil
}

// Snapshot creates a database snapshot
func (m *Manager) Snapshot(name string) error {
	args := []string{"snapshot"}
	if name != "" {
		args = append(args, "-n", name)
	}

	cmd := exec.Command("ddev", args...)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	return nil
}

// RestoreSnapshot restores a database snapshot
func (m *Manager) RestoreSnapshot(name string) error {
	cmd := exec.Command("ddev", "snapshot", "restore", name)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore snapshot: %w", err)
	}

	return nil
}

// Config initializes DDEV configuration (runs ddev config)
func (m *Manager) Config(options ConfigOptions) error {
	args := []string{"config"}

	if options.ProjectName != "" {
		args = append(args, "--project-name="+options.ProjectName)
	}
	if options.Type != "" {
		args = append(args, "--project-type="+options.Type)
	}
	if options.DocRoot != "" {
		args = append(args, "--docroot="+options.DocRoot)
	}
	if options.PHPVersion != "" {
		args = append(args, "--php-version="+options.PHPVersion)
	}
	if options.DatabaseType != "" && options.DatabaseVersion != "" {
		args = append(args, fmt.Sprintf("--database=%s:%s", options.DatabaseType, options.DatabaseVersion))
	}

	cmd := exec.Command("ddev", args...)
	cmd.Dir = m.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure DDEV: %w", err)
	}

	return nil
}

// RequireRunning checks if DDEV is running with retry logic and returns a
// user-friendly error if not. This is useful for commands that require DDEV
// to be running before proceeding.
func (m *Manager) RequireRunning() error {
	var running bool
	var err error

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		running, err = m.IsRunning()
		if err == nil && running {
			return nil
		}
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to check DDEV status in %s: %w\n\nTry running: cd %s && ddev describe", m.ProjectDir, err, m.ProjectDir)
	}
	return fmt.Errorf("DDEV is not running.\n\nProject directory: %s\nPlease run: cd %s && stax start", m.ProjectDir, m.ProjectDir)
}

// WaitForReady waits for DDEV services to be ready
func (m *Manager) WaitForReady(timeout time.Duration) error {
	start := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			running, err := m.IsRunning()
			if err != nil {
				return err
			}
			if running {
				// Give it a bit more time for services to fully initialize
				time.Sleep(2 * time.Second)
				return nil
			}

			if time.Since(start) > timeout {
				return fmt.Errorf("timeout waiting for DDEV to be ready")
			}
		}
	}
}

// WaitForDatabaseReady waits for the database to be ready for queries
// by attempting a simple WP-CLI database command with retries.
// This should be called after ImportDB() to ensure database is fully initialized.
func (m *Manager) WaitForDatabaseReady(timeout time.Duration) error {
	start := time.Now()
	interval := 500 * time.Millisecond
	maxInterval := 2 * time.Second

	for {
		// Try a simple database query via WP-CLI
		cmd := exec.Command("ddev", "wp", "db", "check", "--quiet")
		cmd.Dir = m.ProjectDir

		if err := cmd.Run(); err == nil {
			return nil // Database is ready
		}

		// Check timeout
		if time.Since(start) > timeout {
			return fmt.Errorf("database not ready after %v - the database container may still be initializing. Try running 'ddev restart' and try again", timeout)
		}

		// Exponential backoff up to maxInterval
		time.Sleep(interval)
		if interval < maxInterval {
			interval = interval * 2
			if interval > maxInterval {
				interval = maxInterval
			}
		}
	}
}

// Helper functions

func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBoolValue(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getPrimaryURL(urls []string) string {
	if len(urls) > 0 {
		return urls[0]
	}
	return ""
}

func getMailhogURL(urls []string) string {
	for _, url := range urls {
		if containsSubstring(url, "8025") {
			return url
		}
	}
	return ""
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && indexOfSubstring(s, substr) >= 0
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func getURLs(result map[string]interface{}) []string {
	urls := []string{}
	if httpURL, ok := result["httpurl"].(string); ok && httpURL != "" {
		urls = append(urls, httpURL)
	}
	if httpsURL, ok := result["httpsurl"].(string); ok && httpsURL != "" {
		urls = append(urls, httpsURL)
	}
	if urlList, ok := result["urls"].([]interface{}); ok {
		for _, u := range urlList {
			if urlStr, ok := u.(string); ok {
				urls = append(urls, urlStr)
			}
		}
	}
	return urls
}

func getHostnames(result map[string]interface{}) []string {
	hostnames := []string{}
	if hnList, ok := result["hostnames"].([]interface{}); ok {
		for _, hn := range hnList {
			if hnStr, ok := hn.(string); ok {
				hostnames = append(hostnames, hnStr)
			}
		}
	}
	return hostnames
}

func parseServices(result map[string]interface{}) []ServiceStatus {
	services := []ServiceStatus{}

	// DDEV doesn't provide detailed service info in JSON,
	// so we return basic info
	services = append(services, ServiceStatus{
		Name:  "web",
		State: getStringValue(result, "status"),
	})

	services = append(services, ServiceStatus{
		Name:  "db",
		State: getStringValue(result, "status"),
	})

	return services
}
