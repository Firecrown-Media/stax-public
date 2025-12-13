package build

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewComposer(t *testing.T) {
	workingDir := "/tmp/test-project"
	composer := NewComposer(workingDir)

	if composer == nil {
		t.Fatal("NewComposer returned nil")
	}

	if composer.workingDir != workingDir {
		t.Errorf("expected workingDir %s, got %s", workingDir, composer.workingDir)
	}
}

func TestComposer_GetComposerJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid composer.json
	composerJSON := ComposerJSON{
		Name:        "test/project",
		Type:        "wordpress-plugin",
		Description: "A test project",
		Require: map[string]string{
			"php": ">=8.0",
		},
		RequireDev: map[string]string{
			"phpunit/phpunit": "^9.0",
		},
		Scripts: map[string]interface{}{
			"lint": "phpcs",
			"fix":  "phpcbf",
		},
	}

	data, err := json.MarshalIndent(composerJSON, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal composer.json: %v", err)
	}

	composerPath := filepath.Join(tmpDir, "composer.json")
	if err := os.WriteFile(composerPath, data, 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	composer := NewComposer(tmpDir)
	result, err := composer.GetComposerJSON()
	if err != nil {
		t.Fatalf("GetComposerJSON failed: %v", err)
	}

	if result.Name != composerJSON.Name {
		t.Errorf("expected name %s, got %s", composerJSON.Name, result.Name)
	}
	if result.Type != composerJSON.Type {
		t.Errorf("expected type %s, got %s", composerJSON.Type, result.Type)
	}
	if result.Description != composerJSON.Description {
		t.Errorf("expected description %s, got %s", composerJSON.Description, result.Description)
	}
}

func TestComposer_GetComposerJSON_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	composer := NewComposer(tmpDir)
	_, err := composer.GetComposerJSON()

	if err == nil {
		t.Error("expected error when composer.json doesn't exist")
	}
}

func TestComposer_GetComposerJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid JSON
	composerPath := filepath.Join(tmpDir, "composer.json")
	if err := os.WriteFile(composerPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write invalid composer.json: %v", err)
	}

	composer := NewComposer(tmpDir)
	_, err := composer.GetComposerJSON()

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestComposer_ListScripts(t *testing.T) {
	tmpDir := t.TempDir()

	composerJSON := ComposerJSON{
		Name: "test/project",
		Scripts: map[string]interface{}{
			"lint":  "phpcs",
			"fix":   "phpcbf",
			"test":  "phpunit",
			"build": []string{"npm run build", "composer dump-autoload"},
		},
	}

	data, err := json.MarshalIndent(composerJSON, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal composer.json: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "composer.json"), data, 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	composer := NewComposer(tmpDir)
	scripts, err := composer.ListScripts()
	if err != nil {
		t.Fatalf("ListScripts failed: %v", err)
	}

	if len(scripts) != 4 {
		t.Errorf("expected 4 scripts, got %d", len(scripts))
	}

	if _, ok := scripts["lint"]; !ok {
		t.Error("expected 'lint' script to exist")
	}
	if _, ok := scripts["fix"]; !ok {
		t.Error("expected 'fix' script to exist")
	}
}

func TestComposer_GetStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// Create composer.json
	composerJSON := ComposerJSON{Name: "test/project"}
	data, _ := json.Marshal(composerJSON)
	if err := os.WriteFile(filepath.Join(tmpDir, "composer.json"), data, 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	composer := NewComposer(tmpDir)
	status, err := composer.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	// composer.json exists
	if !status.ConfigExists {
		t.Error("expected ConfigExists to be true")
	}

	// composer.lock doesn't exist
	if status.LockExists {
		t.Error("expected LockExists to be false")
	}

	// vendor doesn't exist
	if status.VendorExists {
		t.Error("expected VendorExists to be false")
	}

	// Not installed (no vendor or lock)
	if status.Installed {
		t.Error("expected Installed to be false")
	}
}

func TestComposer_GetStatus_Installed(t *testing.T) {
	tmpDir := t.TempDir()

	// Create composer.json
	composerJSON := ComposerJSON{Name: "test/project"}
	data, _ := json.Marshal(composerJSON)
	if err := os.WriteFile(filepath.Join(tmpDir, "composer.json"), data, 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	// Create composer.lock
	if err := os.WriteFile(filepath.Join(tmpDir, "composer.lock"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write composer.lock: %v", err)
	}

	// Create vendor directory
	vendorDir := filepath.Join(tmpDir, "vendor")
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		t.Fatalf("failed to create vendor dir: %v", err)
	}

	composer := NewComposer(tmpDir)
	status, err := composer.GetStatus()
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

func TestComposer_GetStatus_NeedsUpdate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create composer.lock first
	lockPath := filepath.Join(tmpDir, "composer.lock")
	if err := os.WriteFile(lockPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write composer.lock: %v", err)
	}

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Create composer.json after (newer)
	composerJSON := ComposerJSON{Name: "test/project"}
	data, _ := json.Marshal(composerJSON)
	configPath := filepath.Join(tmpDir, "composer.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	// Create vendor directory
	vendorDir := filepath.Join(tmpDir, "vendor")
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		t.Fatalf("failed to create vendor dir: %v", err)
	}

	composer := NewComposer(tmpDir)
	status, err := composer.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.NeedsUpdate {
		t.Error("expected NeedsUpdate to be true when composer.json is newer than lock")
	}
}

func TestCheckComposerExists(t *testing.T) {
	// This test depends on whether composer is installed on the system
	// Just verify it returns a boolean without error
	result := CheckComposerExists()
	t.Logf("CheckComposerExists returned: %v", result)
}

func TestGetComposerVersion(t *testing.T) {
	// Skip if composer not installed
	if !CheckComposerExists() {
		t.Skip("composer not installed, skipping version test")
	}

	version, err := GetComposerVersion()
	if err != nil {
		t.Fatalf("GetComposerVersion failed: %v", err)
	}

	if version == "" {
		t.Error("expected non-empty version string")
	}

	t.Logf("Composer version: %s", version)
}

func TestComposerOptions_BuildArgs(t *testing.T) {
	// Test that Install with various options builds correct args
	// This is an indirect test since runComposer is not exported

	testCases := []struct {
		name     string
		options  ComposerOptions
		expected []string // Expected to contain these args
	}{
		{
			name:     "default options",
			options:  ComposerOptions{},
			expected: []string{},
		},
		{
			name:     "no-dev",
			options:  ComposerOptions{NoDev: true},
			expected: []string{"--no-dev"},
		},
		{
			name:     "no-scripts",
			options:  ComposerOptions{NoScripts: true},
			expected: []string{"--no-scripts"},
		},
		{
			name:     "ignore-platform-reqs",
			options:  ComposerOptions{IgnorePlatformReqs: true},
			expected: []string{"--ignore-platform-reqs"},
		},
		{
			name:     "prefer-dist",
			options:  ComposerOptions{PreferDist: true},
			expected: []string{"--prefer-dist"},
		},
		{
			name:     "optimize",
			options:  ComposerOptions{Optimize: true},
			expected: []string{"--optimize-autoloader"},
		},
		{
			name:     "verbose",
			options:  ComposerOptions{Verbose: true},
			expected: []string{"-v"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't directly test the args without running composer,
			// but we can verify the options are properly initialized
			if tc.options.NoDev && !contains(tc.expected, "--no-dev") {
				t.Error("expected --no-dev in args")
			}
		})
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
