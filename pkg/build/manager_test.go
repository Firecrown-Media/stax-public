package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	projectPath := "/tmp/test-project"
	manager := NewManager(projectPath)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.projectPath != projectPath {
		t.Errorf("expected projectPath %s, got %s", projectPath, manager.projectPath)
	}

	if manager.verbose {
		t.Error("expected verbose to be false by default")
	}
}

func TestManager_SetVerbose(t *testing.T) {
	manager := NewManager("/tmp/test")

	manager.SetVerbose(true)
	if !manager.verbose {
		t.Error("expected verbose to be true after SetVerbose(true)")
	}

	manager.SetVerbose(false)
	if manager.verbose {
		t.Error("expected verbose to be false after SetVerbose(false)")
	}
}

func TestManager_DetectBuildScripts(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create scripts directory structure
	scriptsDir := filepath.Join(tmpDir, "scripts")
	buildDir := filepath.Join(scriptsDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}

	// Create main build script
	mainScript := filepath.Join(scriptsDir, "build.sh")
	if err := os.WriteFile(mainScript, []byte("#!/bin/bash\necho 'build'"), 0755); err != nil {
		t.Fatalf("failed to create main script: %v", err)
	}

	// Create numbered build scripts
	if err := os.WriteFile(filepath.Join(buildDir, "10-mu-plugins.sh"), []byte("#!/bin/bash\necho 'mu-plugins'"), 0755); err != nil {
		t.Fatalf("failed to create mu-plugins script: %v", err)
	}
	if err := os.WriteFile(filepath.Join(buildDir, "20-theme.sh"), []byte("#!/bin/bash\necho 'theme'"), 0755); err != nil {
		t.Fatalf("failed to create theme script: %v", err)
	}

	manager := NewManager(tmpDir)
	scripts, err := manager.DetectBuildScripts()
	if err != nil {
		t.Fatalf("DetectBuildScripts failed: %v", err)
	}

	if len(scripts) != 3 {
		t.Errorf("expected 3 scripts, got %d", len(scripts))
	}

	// Verify scripts are sorted by order
	if len(scripts) >= 3 {
		if scripts[0].Name != "build.sh" {
			t.Errorf("expected first script to be build.sh, got %s", scripts[0].Name)
		}
		if scripts[1].Name != "10-mu-plugins.sh" {
			t.Errorf("expected second script to be 10-mu-plugins.sh, got %s", scripts[1].Name)
		}
		if scripts[2].Name != "20-theme.sh" {
			t.Errorf("expected third script to be 20-theme.sh, got %s", scripts[2].Name)
		}
	}
}

func TestManager_DetectBuildScripts_NoScripts(t *testing.T) {
	tmpDir := t.TempDir()

	manager := NewManager(tmpDir)
	scripts, err := manager.DetectBuildScripts()
	if err != nil {
		t.Fatalf("DetectBuildScripts failed: %v", err)
	}

	if len(scripts) != 0 {
		t.Errorf("expected 0 scripts for empty project, got %d", len(scripts))
	}
}

func TestManager_Clean(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure that would be cleaned
	vendorDir := filepath.Join(tmpDir, "wp-content", "mu-plugins", "firecrown", "vendor")
	nodeModulesDir := filepath.Join(tmpDir, "wp-content", "themes", "firecrown-parent", "node_modules")
	buildDir := filepath.Join(tmpDir, "wp-content", "themes", "firecrown-parent", "build")

	for _, dir := range []string{vendorDir, nodeModulesDir, buildDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
		// Add a file to each directory
		testFile := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	manager := NewManager(tmpDir)
	if err := manager.Clean(); err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	// Verify directories were removed
	for _, dir := range []string{vendorDir, nodeModulesDir, buildDir} {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed", dir)
		}
	}
}

func TestManager_GenerateBuildScript(t *testing.T) {
	tmpDir := t.TempDir()

	manager := NewManager(tmpDir)
	if err := manager.GenerateBuildScript(); err != nil {
		t.Fatalf("GenerateBuildScript failed: %v", err)
	}

	// Verify main build script was created
	mainScript := filepath.Join(tmpDir, "scripts", "build.sh")
	if _, err := os.Stat(mainScript); os.IsNotExist(err) {
		t.Error("expected main build script to be created")
	}

	// Verify it's executable
	info, err := os.Stat(mainScript)
	if err != nil {
		t.Fatalf("failed to stat main script: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("expected main build script to be executable")
	}

	// Verify individual build scripts were created
	muPluginsScript := filepath.Join(tmpDir, "scripts", "build", "10-mu-plugins.sh")
	if _, err := os.Stat(muPluginsScript); os.IsNotExist(err) {
		t.Error("expected 10-mu-plugins.sh to be created")
	}

	themeScript := filepath.Join(tmpDir, "scripts", "build", "20-theme.sh")
	if _, err := os.Stat(themeScript); os.IsNotExist(err) {
		t.Error("expected 20-theme.sh to be created")
	}
}

func TestManager_GenerateBuildScript_DoesNotOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing build script with custom content
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	mainScript := filepath.Join(scriptsDir, "build.sh")
	customContent := "#!/bin/bash\necho 'custom build'"
	if err := os.WriteFile(mainScript, []byte(customContent), 0755); err != nil {
		t.Fatalf("failed to create custom script: %v", err)
	}

	manager := NewManager(tmpDir)
	if err := manager.GenerateBuildScript(); err != nil {
		t.Fatalf("GenerateBuildScript failed: %v", err)
	}

	// Verify custom content was preserved
	data, err := os.ReadFile(mainScript)
	if err != nil {
		t.Fatalf("failed to read script: %v", err)
	}

	if string(data) != customContent {
		t.Error("expected existing build script content to be preserved")
	}
}

func TestManager_RunBuildScript_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	manager := NewManager(tmpDir)
	_, err := manager.RunBuildScript()

	if err == nil {
		t.Error("expected error when build script doesn't exist")
	}
}

func TestScriptInfo_Order(t *testing.T) {
	// Test that script order parsing works correctly
	testCases := []struct {
		filename      string
		expectedOrder int
	}{
		{"build.sh", 0},          // main script
		{"10-mu-plugins.sh", 10}, // numbered
		{"20-theme.sh", 20},      // numbered
		{"05-early.sh", 5},       // low number
		{"99-last.sh", 99},       // high number
		{"custom.sh", 999},       // unnumbered
	}

	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "scripts", "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}

	// Create all test scripts
	for _, tc := range testCases {
		var path string
		if tc.filename == "build.sh" {
			path = filepath.Join(tmpDir, "scripts", tc.filename)
		} else {
			path = filepath.Join(buildDir, tc.filename)
		}
		if err := os.WriteFile(path, []byte("#!/bin/bash\necho test"), 0755); err != nil {
			t.Fatalf("failed to create script %s: %v", tc.filename, err)
		}
	}

	manager := NewManager(tmpDir)
	scripts, err := manager.DetectBuildScripts()
	if err != nil {
		t.Fatalf("DetectBuildScripts failed: %v", err)
	}

	// Verify order is ascending
	for i := 1; i < len(scripts); i++ {
		if scripts[i].Order < scripts[i-1].Order {
			t.Errorf("scripts not sorted by order: %s (order %d) comes after %s (order %d)",
				scripts[i].Name, scripts[i].Order, scripts[i-1].Name, scripts[i-1].Order)
		}
	}
}
