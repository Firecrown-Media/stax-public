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

// Composer handles composer operations
type Composer struct {
	workingDir string
}

// NewComposer creates a new Composer instance
func NewComposer(workingDir string) *Composer {
	return &Composer{
		workingDir: workingDir,
	}
}

// Install runs composer install
func (c *Composer) Install(options ComposerOptions) error {
	args := []string{"install"}

	if options.NoDev {
		args = append(args, "--no-dev")
	}

	if options.NoScripts {
		args = append(args, "--no-scripts")
	}

	if options.IgnorePlatformReqs {
		args = append(args, "--ignore-platform-reqs")
	}

	if options.PreferDist {
		args = append(args, "--prefer-dist")
	}

	if options.PreferSource {
		args = append(args, "--prefer-source")
	}

	if options.Optimize {
		args = append(args, "--optimize-autoloader")
	}

	if options.Verbose {
		args = append(args, "-v")
	}

	return c.runComposer(args, options.Timeout)
}

// Update runs composer update
func (c *Composer) Update(packages []string, options ComposerOptions) error {
	args := []string{"update"}
	args = append(args, packages...)

	if options.NoDev {
		args = append(args, "--no-dev")
	}

	if options.IgnorePlatformReqs {
		args = append(args, "--ignore-platform-reqs")
	}

	if options.PreferDist {
		args = append(args, "--prefer-dist")
	}

	if options.Verbose {
		args = append(args, "-v")
	}

	return c.runComposer(args, options.Timeout)
}

// RunScript runs a specific composer script
func (c *Composer) RunScript(scriptName string, options ComposerOptions) error {
	args := []string{"run-script", scriptName}

	if options.Verbose {
		args = append(args, "-v")
	}

	return c.runComposer(args, options.Timeout)
}

// Lint runs PHPCS via composer (if lint script exists)
func (c *Composer) Lint() error {
	scripts, err := c.ListScripts()
	if err != nil {
		return err
	}

	// Check if lint script exists
	if _, exists := scripts["lint"]; !exists {
		return fmt.Errorf("lint script not found in composer.json")
	}

	return c.RunScript("lint", ComposerOptions{})
}

// Fix runs PHPCBF via composer (if fix script exists)
func (c *Composer) Fix() error {
	scripts, err := c.ListScripts()
	if err != nil {
		return err
	}

	// Check if fix script exists
	if _, exists := scripts["fix"]; !exists {
		return fmt.Errorf("fix script not found in composer.json")
	}

	return c.RunScript("fix", ComposerOptions{})
}

// ListScripts returns available composer scripts
func (c *Composer) ListScripts() (map[string]interface{}, error) {
	composerJSON, err := c.GetComposerJSON()
	if err != nil {
		return nil, err
	}

	return composerJSON.Scripts, nil
}

// GetComposerJSON parses and returns the composer.json file
func (c *Composer) GetComposerJSON() (*ComposerJSON, error) {
	composerFile := filepath.Join(c.workingDir, "composer.json")

	data, err := os.ReadFile(composerFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read composer.json: %w", err)
	}

	var composerJSON ComposerJSON
	if err := json.Unmarshal(data, &composerJSON); err != nil {
		return nil, fmt.Errorf("failed to parse composer.json: %w", err)
	}

	return &composerJSON, nil
}

// ValidateComposer validates the composer.json file
func (c *Composer) ValidateComposer() error {
	args := []string{"validate", "--no-check-all", "--no-check-publish"}
	return c.runComposer(args, 30)
}

// GetStatus returns the status of composer dependencies
func (c *Composer) GetStatus() (*DependencyStatus, error) {
	status := &DependencyStatus{
		ConfigFile: filepath.Join(c.workingDir, "composer.json"),
		LockFile:   filepath.Join(c.workingDir, "composer.lock"),
		VendorDir:  filepath.Join(c.workingDir, "vendor"),
	}

	// Check composer.json
	if info, err := os.Stat(status.ConfigFile); err == nil {
		status.ConfigExists = true
		status.ConfigModified = info.ModTime()
	}

	// Check composer.lock
	if info, err := os.Stat(status.LockFile); err == nil {
		status.LockExists = true
		status.LockModified = info.ModTime()
	}

	// Check vendor directory
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

// runComposer executes a composer command
func (c *Composer) runComposer(args []string, timeout int) error {
	cmd := exec.Command("composer", args...)
	cmd.Dir = c.workingDir
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

// CheckComposerExists checks if composer is available
func CheckComposerExists() bool {
	cmd := exec.Command("composer", "--version")
	return cmd.Run() == nil
}

// GetComposerVersion returns the installed composer version
func GetComposerVersion() (string, error) {
	cmd := exec.Command("composer", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from output (e.g., "Composer version 2.5.8")
	version := strings.TrimSpace(string(output))
	parts := strings.Fields(version)
	if len(parts) >= 3 {
		return parts[2], nil
	}

	return version, nil
}

// RequirePackage adds a package to composer.json
func (c *Composer) RequirePackage(packageName string, version string, dev bool) error {
	args := []string{"require", packageName}

	if version != "" {
		args[1] = fmt.Sprintf("%s:%s", packageName, version)
	}

	if dev {
		args = append(args, "--dev")
	}

	return c.runComposer(args, 300)
}

// RemovePackage removes a package from composer.json
func (c *Composer) RemovePackage(packageName string, dev bool) error {
	args := []string{"remove", packageName}

	if dev {
		args = append(args, "--dev")
	}

	return c.runComposer(args, 60)
}

// DumpAutoload regenerates the autoloader
func (c *Composer) DumpAutoload(optimize bool) error {
	args := []string{"dump-autoload"}

	if optimize {
		args = append(args, "--optimize")
	}

	return c.runComposer(args, 30)
}

// ClearCache clears the composer cache
func (c *Composer) ClearCache() error {
	args := []string{"clear-cache"}
	return c.runComposer(args, 30)
}

// Diagnose runs composer diagnose
func (c *Composer) Diagnose() error {
	args := []string{"diagnose"}
	return c.runComposer(args, 60)
}
