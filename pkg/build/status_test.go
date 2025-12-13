package build

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStatusChecker(t *testing.T) {
	projectPath := "/tmp/test-project"
	checker := NewStatusChecker(projectPath)

	if checker == nil {
		t.Fatal("NewStatusChecker returned nil")
	}

	if checker.projectPath != projectPath {
		t.Errorf("expected projectPath %s, got %s", projectPath, checker.projectPath)
	}
}

func TestStatusChecker_GetStatus_NoBuildScript(t *testing.T) {
	tmpDir := t.TempDir()

	checker := NewStatusChecker(tmpDir)
	status, err := checker.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.BuildScriptExists {
		t.Error("expected BuildScriptExists to be false")
	}

	// Should have "No build script found" in reasons
	found := false
	for _, reason := range status.Reasons {
		if reason == "No build script found" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'No build script found' in reasons")
	}
}

func TestStatusChecker_GetStatus_WithBuildScript(t *testing.T) {
	tmpDir := t.TempDir()

	// Create scripts/build.sh
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	buildScript := filepath.Join(scriptsDir, "build.sh")
	if err := os.WriteFile(buildScript, []byte("#!/bin/bash\necho 'building'"), 0755); err != nil {
		t.Fatalf("failed to create build script: %v", err)
	}

	checker := NewStatusChecker(tmpDir)
	status, err := checker.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.BuildScriptExists {
		t.Error("expected BuildScriptExists to be true")
	}
}

func TestStatusChecker_GetStatus_CustomBuildScripts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create scripts/build/ directory with custom scripts
	buildDir := filepath.Join(tmpDir, "scripts", "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}

	scripts := []string{"10-mu-plugins.sh", "20-parent-theme.sh", "30-child-theme.sh"}
	for _, script := range scripts {
		scriptPath := filepath.Join(buildDir, script)
		if err := os.WriteFile(scriptPath, []byte("#!/bin/bash"), 0755); err != nil {
			t.Fatalf("failed to create script %s: %v", script, err)
		}
	}

	checker := NewStatusChecker(tmpDir)
	status, err := checker.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if len(status.CustomBuildScripts) != 3 {
		t.Errorf("expected 3 custom build scripts, got %d", len(status.CustomBuildScripts))
	}
}

func TestStatusChecker_NeedsBuild_EmptyProject(t *testing.T) {
	tmpDir := t.TempDir()

	checker := NewStatusChecker(tmpDir)
	needsBuild, reasons := checker.NeedsBuild()

	// Empty project should need build (missing artifacts)
	if !needsBuild {
		t.Error("expected NeedsBuild to be true for empty project")
	}

	// Should have at least one reason
	if len(reasons) == 0 {
		t.Error("expected at least one reason for needing build")
	}
}

func TestStatusChecker_GetLastBuildTime_NoArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	checker := NewStatusChecker(tmpDir)
	_, err := checker.GetLastBuildTime()

	if err == nil {
		t.Error("expected error when no build artifacts exist")
	}
}

func TestStatusChecker_GetLastBuildTime_WithArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a build artifact
	artifactPath := filepath.Join(tmpDir, "wp-content", "mu-plugins", "firecrown", "vendor")
	if err := os.MkdirAll(artifactPath, 0755); err != nil {
		t.Fatalf("failed to create artifact directory: %v", err)
	}

	// Touch the directory to set modification time
	now := time.Now()
	if err := os.Chtimes(artifactPath, now, now); err != nil {
		t.Fatalf("failed to set modification time: %v", err)
	}

	checker := NewStatusChecker(tmpDir)
	buildTime, err := checker.GetLastBuildTime()

	if err != nil {
		t.Fatalf("GetLastBuildTime failed: %v", err)
	}

	// Build time should be close to now
	if buildTime.Before(now.Add(-time.Second)) {
		t.Errorf("expected build time around %v, got %v", now, buildTime)
	}
}

func TestStatusChecker_buildArtifactsExist_AllPresent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create all required artifacts
	artifacts := []string{
		filepath.Join(tmpDir, "wp-content", "mu-plugins", "firecrown", "vendor"),
		filepath.Join(tmpDir, "wp-content", "themes", "firecrown-parent", "build"),
		filepath.Join(tmpDir, "wp-content", "themes", "firecrown-parent", "node_modules"),
	}

	for _, artifact := range artifacts {
		if err := os.MkdirAll(artifact, 0755); err != nil {
			t.Fatalf("failed to create artifact %s: %v", artifact, err)
		}
	}

	checker := NewStatusChecker(tmpDir)
	if !checker.buildArtifactsExist() {
		t.Error("expected buildArtifactsExist to return true")
	}
}

func TestStatusChecker_buildArtifactsExist_Missing(t *testing.T) {
	tmpDir := t.TempDir()

	checker := NewStatusChecker(tmpDir)
	if checker.buildArtifactsExist() {
		t.Error("expected buildArtifactsExist to return false for empty project")
	}
}

func TestStatusChecker_sourceFilesNewer_ZeroTime(t *testing.T) {
	tmpDir := t.TempDir()

	checker := NewStatusChecker(tmpDir)

	// Zero time should return true (always needs build)
	if !checker.sourceFilesNewer(time.Time{}) {
		t.Error("expected sourceFilesNewer to return true for zero time")
	}
}

func TestStatusChecker_sourceFilesNewer_OldBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	sourcePath := filepath.Join(tmpDir, "wp-content", "mu-plugins", "firecrown")
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	composerJSON := filepath.Join(sourcePath, "composer.json")
	if err := os.WriteFile(composerJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create composer.json: %v", err)
	}

	checker := NewStatusChecker(tmpDir)

	// Use a build time from the past
	oldBuildTime := time.Now().Add(-24 * time.Hour)

	if !checker.sourceFilesNewer(oldBuildTime) {
		t.Error("expected sourceFilesNewer to return true when source files are newer")
	}
}

func TestStatusChecker_CompareTimestamps(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	source := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(source, []byte("source"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Create dest file (older than source)
	dest := filepath.Join(tmpDir, "dest.txt")

	// First, create dest before source to make it older
	if err := os.WriteFile(dest, []byte("dest"), 0644); err != nil {
		t.Fatalf("failed to create dest file: %v", err)
	}

	// Touch source to make it newer
	time.Sleep(10 * time.Millisecond)
	now := time.Now()
	if err := os.Chtimes(source, now, now); err != nil {
		t.Fatalf("failed to touch source: %v", err)
	}

	checker := NewStatusChecker(tmpDir)
	newer, err := checker.CompareTimestamps(source, dest)
	if err != nil {
		t.Fatalf("CompareTimestamps failed: %v", err)
	}

	if !newer {
		t.Error("expected source to be newer than dest")
	}
}

func TestStatusChecker_CompareTimestamps_DestMissing(t *testing.T) {
	tmpDir := t.TempDir()

	source := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(source, []byte("source"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	dest := filepath.Join(tmpDir, "nonexistent.txt")

	checker := NewStatusChecker(tmpDir)
	newer, err := checker.CompareTimestamps(source, dest)
	if err != nil {
		t.Fatalf("CompareTimestamps failed: %v", err)
	}

	// Should return true when dest doesn't exist
	if !newer {
		t.Error("expected true when dest doesn't exist")
	}
}

func TestStatusChecker_CompareTimestamps_SourceMissing(t *testing.T) {
	tmpDir := t.TempDir()

	source := filepath.Join(tmpDir, "nonexistent.txt")
	dest := filepath.Join(tmpDir, "dest.txt")

	if err := os.WriteFile(dest, []byte("dest"), 0644); err != nil {
		t.Fatalf("failed to create dest file: %v", err)
	}

	checker := NewStatusChecker(tmpDir)
	_, err := checker.CompareTimestamps(source, dest)

	if err == nil {
		t.Error("expected error when source doesn't exist")
	}
}

func TestStatusChecker_IsBuildRunning_NoLock(t *testing.T) {
	tmpDir := t.TempDir()

	checker := NewStatusChecker(tmpDir)
	if checker.IsBuildRunning() {
		t.Error("expected IsBuildRunning to return false when no lock file")
	}
}

func TestStatusChecker_IsBuildRunning_WithLock(t *testing.T) {
	tmpDir := t.TempDir()

	// Create lock file
	lockDir := filepath.Join(tmpDir, ".stax")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		t.Fatalf("failed to create .stax dir: %v", err)
	}

	lockFile := filepath.Join(lockDir, "build.lock")
	if err := os.WriteFile(lockFile, []byte("12345"), 0644); err != nil {
		t.Fatalf("failed to create lock file: %v", err)
	}

	checker := NewStatusChecker(tmpDir)
	if !checker.IsBuildRunning() {
		t.Error("expected IsBuildRunning to return true when lock file exists")
	}
}

func TestStatusChecker_CreateBuildLock(t *testing.T) {
	tmpDir := t.TempDir()

	checker := NewStatusChecker(tmpDir)
	if err := checker.CreateBuildLock(); err != nil {
		t.Fatalf("CreateBuildLock failed: %v", err)
	}

	// Verify lock file exists
	lockFile := filepath.Join(tmpDir, ".stax", "build.lock")
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Error("expected lock file to be created")
	}

	// Verify lock file contains PID
	data, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("failed to read lock file: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected lock file to contain PID")
	}
}

func TestStatusChecker_RemoveBuildLock(t *testing.T) {
	tmpDir := t.TempDir()

	// Create lock file first
	lockDir := filepath.Join(tmpDir, ".stax")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		t.Fatalf("failed to create .stax dir: %v", err)
	}

	lockFile := filepath.Join(lockDir, "build.lock")
	if err := os.WriteFile(lockFile, []byte("12345"), 0644); err != nil {
		t.Fatalf("failed to create lock file: %v", err)
	}

	checker := NewStatusChecker(tmpDir)
	if err := checker.RemoveBuildLock(); err != nil {
		t.Fatalf("RemoveBuildLock failed: %v", err)
	}

	// Verify lock file was removed
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("expected lock file to be removed")
	}
}

func TestStatusChecker_ValidateBuild_AllPresent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create all required build output
	paths := []string{
		filepath.Join(tmpDir, "wp-content", "mu-plugins", "firecrown", "vendor"),
		filepath.Join(tmpDir, "wp-content", "themes", "firecrown-parent", "build"),
		filepath.Join(tmpDir, "wp-content", "themes", "firecrown-child", "build"),
	}

	for _, path := range paths {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create path %s: %v", path, err)
		}
	}

	// Create autoload.php
	autoload := filepath.Join(tmpDir, "wp-content", "mu-plugins", "firecrown", "vendor", "autoload.php")
	if err := os.WriteFile(autoload, []byte("<?php"), 0644); err != nil {
		t.Fatalf("failed to create autoload.php: %v", err)
	}

	// Create compiled CSS
	cssPath := filepath.Join(tmpDir, "wp-content", "themes", "firecrown-parent", "build", "scripts.css")
	if err := os.WriteFile(cssPath, []byte("/* css */"), 0644); err != nil {
		t.Fatalf("failed to create CSS file: %v", err)
	}

	checker := NewStatusChecker(tmpDir)
	if err := checker.ValidateBuild(); err != nil {
		t.Errorf("ValidateBuild should pass when all files present: %v", err)
	}
}

func TestStatusChecker_ValidateBuild_MissingAutoload(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories but not autoload.php
	vendorPath := filepath.Join(tmpDir, "wp-content", "mu-plugins", "firecrown", "vendor")
	if err := os.MkdirAll(vendorPath, 0755); err != nil {
		t.Fatalf("failed to create vendor path: %v", err)
	}

	checker := NewStatusChecker(tmpDir)
	err := checker.ValidateBuild()

	if err == nil {
		t.Error("expected error when autoload.php is missing")
	}
}

func TestStatusChecker_GetBuildDuration(t *testing.T) {
	checker := NewStatusChecker("/tmp/test")

	duration := checker.GetBuildDuration()

	// Should return some non-zero duration
	if duration == 0 {
		t.Error("expected non-zero build duration")
	}

	// Should return 3 minutes based on current implementation
	expected := 3 * time.Minute
	if duration != expected {
		t.Errorf("expected %v, got %v", expected, duration)
	}
}
