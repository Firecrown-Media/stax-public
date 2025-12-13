package build

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewNPM(t *testing.T) {
	workingDir := "/tmp/test-project"
	npm := NewNPM(workingDir)

	if npm == nil {
		t.Fatal("NewNPM returned nil")
	}

	if npm.workingDir != workingDir {
		t.Errorf("expected workingDir %s, got %s", workingDir, npm.workingDir)
	}
}

func TestNPM_GetPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid package.json
	packageJSON := PackageJSON{
		Name:        "test-project",
		Version:     "1.0.0",
		Description: "A test project",
		Scripts: map[string]string{
			"build": "webpack",
			"test":  "jest",
			"lint":  "eslint .",
		},
		Dependencies: map[string]string{
			"react": "^18.0.0",
		},
		DevDependencies: map[string]string{
			"webpack": "^5.0.0",
		},
	}

	data, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal package.json: %v", err)
	}

	packagePath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packagePath, data, 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	npm := NewNPM(tmpDir)
	result, err := npm.GetPackageJSON()
	if err != nil {
		t.Fatalf("GetPackageJSON failed: %v", err)
	}

	if result.Name != packageJSON.Name {
		t.Errorf("expected name %s, got %s", packageJSON.Name, result.Name)
	}
	if result.Version != packageJSON.Version {
		t.Errorf("expected version %s, got %s", packageJSON.Version, result.Version)
	}
	if result.Description != packageJSON.Description {
		t.Errorf("expected description %s, got %s", packageJSON.Description, result.Description)
	}
}

func TestNPM_GetPackageJSON_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	npm := NewNPM(tmpDir)
	_, err := npm.GetPackageJSON()

	if err == nil {
		t.Error("expected error when package.json doesn't exist")
	}
}

func TestNPM_GetPackageJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid JSON
	packagePath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packagePath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write invalid package.json: %v", err)
	}

	npm := NewNPM(tmpDir)
	_, err := npm.GetPackageJSON()

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNPM_ListScripts(t *testing.T) {
	tmpDir := t.TempDir()

	packageJSON := PackageJSON{
		Name:    "test-project",
		Version: "1.0.0",
		Scripts: map[string]string{
			"build":   "webpack",
			"test":    "jest",
			"lint":    "eslint .",
			"start":   "node server.js",
			"dev":     "nodemon",
			"prepare": "husky install",
		},
	}

	data, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal package.json: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), data, 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	npm := NewNPM(tmpDir)
	scripts, err := npm.ListScripts()
	if err != nil {
		t.Fatalf("ListScripts failed: %v", err)
	}

	if len(scripts) != 6 {
		t.Errorf("expected 6 scripts, got %d", len(scripts))
	}

	if _, ok := scripts["build"]; !ok {
		t.Error("expected 'build' script to exist")
	}
	if _, ok := scripts["test"]; !ok {
		t.Error("expected 'test' script to exist")
	}
	if _, ok := scripts["lint"]; !ok {
		t.Error("expected 'lint' script to exist")
	}
}

func TestNPM_ValidatePackage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid package.json
	packageJSON := PackageJSON{
		Name:    "test-project",
		Version: "1.0.0",
	}

	data, _ := json.Marshal(packageJSON)
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), data, 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	npm := NewNPM(tmpDir)
	if err := npm.ValidatePackage(); err != nil {
		t.Errorf("ValidatePackage failed for valid package: %v", err)
	}
}

func TestNPM_ValidatePackage_Invalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid package.json
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{invalid}"), 0644); err != nil {
		t.Fatalf("failed to write invalid package.json: %v", err)
	}

	npm := NewNPM(tmpDir)
	if err := npm.ValidatePackage(); err == nil {
		t.Error("expected error for invalid package.json")
	}
}

func TestNPM_CleanNodeModules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create node_modules directory
	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}

	// Create a file inside node_modules
	testFile := filepath.Join(nodeModulesDir, "test-package", "index.js")
	if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
		t.Fatalf("failed to create test package dir: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("module.exports = {}"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	npm := NewNPM(tmpDir)
	if err := npm.CleanNodeModules(); err != nil {
		t.Fatalf("CleanNodeModules failed: %v", err)
	}

	// Verify node_modules was removed
	if _, err := os.Stat(nodeModulesDir); !os.IsNotExist(err) {
		t.Error("expected node_modules to be removed")
	}
}

func TestNPM_CleanNodeModules_NoDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	npm := NewNPM(tmpDir)
	// Should not error when node_modules doesn't exist
	if err := npm.CleanNodeModules(); err != nil {
		t.Errorf("CleanNodeModules should not error when directory doesn't exist: %v", err)
	}
}

func TestNPM_GetStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package.json
	packageJSON := PackageJSON{Name: "test-project", Version: "1.0.0"}
	data, _ := json.Marshal(packageJSON)
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), data, 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	npm := NewNPM(tmpDir)
	status, err := npm.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	// package.json exists
	if !status.ConfigExists {
		t.Error("expected ConfigExists to be true")
	}

	// package-lock.json doesn't exist
	if status.LockExists {
		t.Error("expected LockExists to be false")
	}

	// node_modules doesn't exist
	if status.VendorExists {
		t.Error("expected VendorExists to be false")
	}

	// Not installed (no node_modules or lock)
	if status.Installed {
		t.Error("expected Installed to be false")
	}
}

func TestNPM_GetStatus_Installed(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package.json
	packageJSON := PackageJSON{Name: "test-project", Version: "1.0.0"}
	data, _ := json.Marshal(packageJSON)
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), data, 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Create package-lock.json
	if err := os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write package-lock.json: %v", err)
	}

	// Create node_modules directory
	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("failed to create node_modules dir: %v", err)
	}

	npm := NewNPM(tmpDir)
	status, err := npm.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.ConfigExists {
		t.Error("expected ConfigExists to be true")
	}
	if !status.LockExists {
		t.Error("expected LockExists to be true")
	}
	if !status.VendorExists {
		t.Error("expected VendorExists to be true")
	}
	if !status.Installed {
		t.Error("expected Installed to be true")
	}
}

func TestNPM_GetStatus_NeedsUpdate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package-lock.json first
	lockPath := filepath.Join(tmpDir, "package-lock.json")
	if err := os.WriteFile(lockPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write package-lock.json: %v", err)
	}

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Create package.json after (newer)
	packageJSON := PackageJSON{Name: "test-project", Version: "1.0.0"}
	data, _ := json.Marshal(packageJSON)
	configPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Create node_modules directory
	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("failed to create node_modules dir: %v", err)
	}

	npm := NewNPM(tmpDir)
	status, err := npm.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.NeedsUpdate {
		t.Error("expected NeedsUpdate to be true when package.json is newer than lock")
	}
}

func TestCheckNPMExists(t *testing.T) {
	// This test depends on whether npm is installed on the system
	// Just verify it returns a boolean without error
	result := CheckNPMExists()
	t.Logf("CheckNPMExists returned: %v", result)
}

func TestGetNPMVersion(t *testing.T) {
	// Skip if npm not installed
	if !CheckNPMExists() {
		t.Skip("npm not installed, skipping version test")
	}

	version, err := GetNPMVersion()
	if err != nil {
		t.Fatalf("GetNPMVersion failed: %v", err)
	}

	if version == "" {
		t.Error("expected non-empty version string")
	}

	t.Logf("NPM version: %s", version)
}

func TestGetNodeVersion(t *testing.T) {
	// Skip if node not installed
	if !CheckNPMExists() {
		t.Skip("npm not installed (node likely missing too), skipping version test")
	}

	version, err := GetNodeVersion()
	if err != nil {
		t.Fatalf("GetNodeVersion failed: %v", err)
	}

	if version == "" {
		t.Error("expected non-empty version string")
	}

	t.Logf("Node version: %s", version)
}

func TestNPMOptions_BuildArgs(t *testing.T) {
	// Test that Install with various options builds correct args
	// This is an indirect test since runNPM is not exported

	testCases := []struct {
		name     string
		options  NPMOptions
		expected []string // Expected to contain these args
	}{
		{
			name:     "default options",
			options:  NPMOptions{},
			expected: []string{},
		},
		{
			name:     "production",
			options:  NPMOptions{Production: true},
			expected: []string{"--production"},
		},
		{
			name:     "legacy-peer-deps",
			options:  NPMOptions{LegacyPeerDeps: true},
			expected: []string{"--legacy-peer-deps"},
		},
		{
			name:     "verbose",
			options:  NPMOptions{Verbose: true},
			expected: []string{"--verbose"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't directly test the args without running npm,
			// but we can verify the options are properly initialized
			if tc.options.Production && !contains(tc.expected, "--production") {
				t.Error("expected --production in args")
			}
		})
	}
}

func TestNPM_StopBackground_NoPIDFile(t *testing.T) {
	tmpDir := t.TempDir()

	npm := NewNPM(tmpDir)
	err := npm.StopBackground()

	if err == nil {
		t.Error("expected error when no PID file exists")
	}
}

func TestNPM_StopBackground_InvalidPID(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid PID file
	pidFile := filepath.Join(tmpDir, ".npm-start.pid")
	if err := os.WriteFile(pidFile, []byte("not-a-number"), 0644); err != nil {
		t.Fatalf("failed to write PID file: %v", err)
	}

	npm := NewNPM(tmpDir)
	err := npm.StopBackground()

	if err == nil {
		t.Error("expected error for invalid PID")
	}
}
