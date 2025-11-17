package build

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NPM handles npm operations
type NPM struct {
	workingDir string
}

// NewNPM creates a new NPM instance
func NewNPM(workingDir string) *NPM {
	return &NPM{
		workingDir: workingDir,
	}
}

// Install runs npm install
func (n *NPM) Install(options NPMOptions) error {
	// Clean node_modules if requested
	if options.Clean {
		nodeModulesPath := filepath.Join(n.workingDir, "node_modules")
		if _, err := os.Stat(nodeModulesPath); err == nil {
			if err := os.RemoveAll(nodeModulesPath); err != nil {
				return fmt.Errorf("failed to remove node_modules: %w", err)
			}
		}
	}

	args := []string{"install"}

	if options.Production {
		args = append(args, "--production")
	}

	if options.LegacyPeerDeps {
		args = append(args, "--legacy-peer-deps")
	}

	if options.Verbose {
		args = append(args, "--verbose")
	}

	return n.runNPM(args, options.Timeout)
}

// Build runs npm run build
func (n *NPM) Build(options NPMOptions) error {
	args := []string{"run", "build"}

	if options.Verbose {
		args = append(args, "--verbose")
	}

	return n.runNPM(args, options.Timeout)
}

// Start runs npm start (for development mode)
func (n *NPM) Start(background bool, options NPMOptions) error {
	args := []string{"start"}

	if options.Verbose {
		args = append(args, "--verbose")
	}

	if background {
		return n.runNPMBackground(args)
	}

	return n.runNPM(args, options.Timeout)
}

// RunScript runs a specific npm script
func (n *NPM) RunScript(scriptName string, options NPMOptions) error {
	args := []string{"run", scriptName}

	if options.Verbose {
		args = append(args, "--verbose")
	}

	return n.runNPM(args, options.Timeout)
}

// ListScripts returns available npm scripts
func (n *NPM) ListScripts() (map[string]string, error) {
	packageJSON, err := n.GetPackageJSON()
	if err != nil {
		return nil, err
	}

	return packageJSON.Scripts, nil
}

// GetPackageJSON parses and returns the package.json file
func (n *NPM) GetPackageJSON() (*PackageJSON, error) {
	packageFile := filepath.Join(n.workingDir, "package.json")

	data, err := os.ReadFile(packageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var packageJSON PackageJSON
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	return &packageJSON, nil
}

// ValidatePackage validates the package.json file
func (n *NPM) ValidatePackage() error {
	_, err := n.GetPackageJSON()
	return err
}

// CleanNodeModules removes the node_modules directory
func (n *NPM) CleanNodeModules() error {
	nodeModulesPath := filepath.Join(n.workingDir, "node_modules")

	if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
		return nil // Already clean
	}

	return os.RemoveAll(nodeModulesPath)
}

// GetStatus returns the status of npm dependencies
func (n *NPM) GetStatus() (*DependencyStatus, error) {
	status := &DependencyStatus{
		ConfigFile: filepath.Join(n.workingDir, "package.json"),
		LockFile:   filepath.Join(n.workingDir, "package-lock.json"),
		VendorDir:  filepath.Join(n.workingDir, "node_modules"),
	}

	// Check package.json
	if info, err := os.Stat(status.ConfigFile); err == nil {
		status.ConfigExists = true
		status.ConfigModified = info.ModTime()
	}

	// Check package-lock.json
	if info, err := os.Stat(status.LockFile); err == nil {
		status.LockExists = true
		status.LockModified = info.ModTime()
	}

	// Check node_modules directory
	if info, err := os.Stat(status.VendorDir); err == nil {
		status.VendorExists = true
		status.VendorModified = info.ModTime()
	}

	// Determine if installed
	status.Installed = status.VendorExists && status.LockExists

	// Determine if needs update
	if status.ConfigExists && status.LockExists {
		status.NeedsUpdate = status.ConfigModified.After(status.LockModified)
	}

	if status.LockExists && status.VendorExists {
		if status.LockModified.After(status.VendorModified) {
			status.NeedsUpdate = true
		}
	}

	return status, nil
}

// runNPM executes an npm command
func (n *NPM) runNPM(args []string, timeout int) error {
	cmd := exec.Command("npm", args...)
	cmd.Dir = n.workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set timeout if specified
	if timeout > 0 {
		timer := time.AfterFunc(time.Duration(timeout)*time.Second, func() {
			// Best-effort process kill on timeout - ignore errors as process may have already exited
			_ = cmd.Process.Kill()
		})
		defer timer.Stop()
	}

	return cmd.Run()
}

// runNPMBackground executes an npm command in the background
func (n *NPM) runNPMBackground(args []string) error {
	cmd := exec.Command("npm", args...)
	cmd.Dir = n.workingDir

	// Create log file for background output
	logFile := filepath.Join(n.workingDir, ".npm-start.log")
	f, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	cmd.Stdout = f
	cmd.Stderr = f

	if err := cmd.Start(); err != nil {
		f.Close()
		return err
	}

	// Write PID to file for later management
	pidFile := filepath.Join(n.workingDir, ".npm-start.pid")
	pidContent := fmt.Sprintf("%d", cmd.Process.Pid)
	if err := os.WriteFile(pidFile, []byte(pidContent), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// StopBackground stops a background npm process
func (n *NPM) StopBackground() error {
	pidFile := filepath.Join(n.workingDir, ".npm-start.pid")

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("no background process found (PID file missing)")
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return fmt.Errorf("invalid PID file: %w", err)
	}

	// Kill the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	// Remove PID file
	os.Remove(pidFile)

	return nil
}

// CheckNPMExists checks if npm is available
func CheckNPMExists() bool {
	cmd := exec.Command("npm", "--version")
	return cmd.Run() == nil
}

// GetNPMVersion returns the installed npm version
func GetNPMVersion() (string, error) {
	cmd := exec.Command("npm", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetNodeVersion returns the installed node version
func GetNodeVersion() (string, error) {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// InstallPackage installs a specific npm package
func (n *NPM) InstallPackage(packageName string, version string, dev bool) error {
	args := []string{"install", packageName}

	if version != "" {
		args[1] = fmt.Sprintf("%s@%s", packageName, version)
	}

	if dev {
		args = append(args, "--save-dev")
	}

	return n.runNPM(args, 300)
}

// UninstallPackage removes a specific npm package
func (n *NPM) UninstallPackage(packageName string) error {
	args := []string{"uninstall", packageName}
	return n.runNPM(args, 60)
}

// Update runs npm update
func (n *NPM) Update(packages []string, options NPMOptions) error {
	args := []string{"update"}
	args = append(args, packages...)

	if options.Verbose {
		args = append(args, "--verbose")
	}

	return n.runNPM(args, options.Timeout)
}

// Outdated checks for outdated packages
func (n *NPM) Outdated() error {
	args := []string{"outdated"}
	return n.runNPM(args, 30)
}

// Audit runs npm audit
func (n *NPM) Audit() error {
	args := []string{"audit"}
	return n.runNPM(args, 30)
}

// AuditFix runs npm audit fix
func (n *NPM) AuditFix() error {
	args := []string{"audit", "fix"}
	return n.runNPM(args, 120)
}

// ClearCache clears the npm cache
func (n *NPM) ClearCache() error {
	args := []string{"cache", "clean", "--force"}
	return n.runNPM(args, 30)
}
