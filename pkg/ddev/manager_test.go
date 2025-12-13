package ddev

import (
	"os"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	projectDir := "/test/project"
	manager := NewManager(projectDir)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.ProjectDir != projectDir {
		t.Errorf("expected ProjectDir %q, got %q", projectDir, manager.ProjectDir)
	}
}

func TestIsInstalled(t *testing.T) {
	// This test checks if DDEV is in PATH
	// It will pass if DDEV is installed, skip if not
	installed := IsInstalled()

	// We can't assert a specific value since it depends on the environment
	// Just verify the function executes without error
	t.Logf("DDEV installed: %v", installed)
}

func TestGetVersion(t *testing.T) {
	if !IsInstalled() {
		t.Skip("DDEV not installed, skipping version test")
	}

	version, err := GetVersion()
	if err != nil {
		t.Fatalf("GetVersion() failed: %v", err)
	}

	if version == "" {
		t.Error("expected non-empty version string")
	}

	t.Logf("DDEV version: %s", version)
}

func TestManager_IsRunning(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	manager := NewManager(tmpDir)

	// Without a DDEV project, this should return false
	running, err := manager.IsRunning()
	if err != nil {
		// Error is acceptable for non-existent project
		t.Logf("IsRunning() error (expected for non-existent project): %v", err)
	}

	// Should not be running since there's no project
	if running {
		t.Error("expected IsRunning() to return false for non-existent project")
	}
}

func TestManager_Start(t *testing.T) {
	if !IsInstalled() {
		t.Skip("DDEV not installed, skipping start test")
	}

	if os.Getenv("RUN_DDEV_TESTS") != "true" {
		t.Skip("Skipping DDEV integration test (set RUN_DDEV_TESTS=true to run)")
	}

	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// This will fail without a proper DDEV project, which is expected
	err := manager.Start()
	if err == nil {
		t.Error("expected Start() to fail for non-existent project")
	}
}

func TestManager_Stop(t *testing.T) {
	if !IsInstalled() {
		t.Skip("DDEV not installed, skipping stop test")
	}

	if os.Getenv("RUN_DDEV_TESTS") != "true" {
		t.Skip("Skipping DDEV integration test (set RUN_DDEV_TESTS=true to run)")
	}

	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// This will fail without a proper DDEV project
	err := manager.Stop()
	if err == nil {
		t.Error("expected Stop() to fail for non-existent project")
	}
}

func TestManager_Restart(t *testing.T) {
	if !IsInstalled() {
		t.Skip("DDEV not installed, skipping restart test")
	}

	if os.Getenv("RUN_DDEV_TESTS") != "true" {
		t.Skip("Skipping DDEV integration test (set RUN_DDEV_TESTS=true to run)")
	}

	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// This will fail without a proper DDEV project
	err := manager.Restart()
	if err == nil {
		t.Error("expected Restart() to fail for non-existent project")
	}
}

// Table-driven tests for manager operations
func TestManagerOperations(t *testing.T) {
	if !IsInstalled() {
		t.Skip("DDEV not installed, skipping manager operation tests")
	}

	tests := []struct {
		name      string
		operation func(*Manager) error
		wantErr   bool
	}{
		{
			name: "start non-existent project",
			operation: func(m *Manager) error {
				return m.Start()
			},
			wantErr: true,
		},
		{
			name: "stop non-existent project",
			operation: func(m *Manager) error {
				return m.Stop()
			},
			wantErr: true,
		},
		{
			name: "restart non-existent project",
			operation: func(m *Manager) error {
				return m.Restart()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if os.Getenv("RUN_DDEV_TESTS") != "true" {
				t.Skip("Skipping DDEV integration test (set RUN_DDEV_TESTS=true to run)")
			}

			tmpDir := t.TempDir()
			manager := NewManager(tmpDir)

			err := tt.operation(manager)
			if (err != nil) != tt.wantErr {
				t.Errorf("operation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_RequireRunning(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	manager := NewManager(tmpDir)

	// Without a DDEV project, this should return an error
	err := manager.RequireRunning()
	if err == nil {
		t.Error("expected RequireRunning() to return error for non-existent project")
	}

	// Error message should mention DDEV is not running
	if err != nil {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "DDEV is not running") && !strings.Contains(errMsg, "failed to check DDEV status") {
			t.Errorf("expected error message to mention DDEV status, got: %s", errMsg)
		}
	}
}

func TestManager_RequireRunning_ErrorMessage(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	err := manager.RequireRunning()
	if err == nil {
		t.Fatal("expected error")
	}

	// Error should include project directory for context
	errMsg := err.Error()
	if !strings.Contains(errMsg, tmpDir) {
		t.Errorf("error message should include project directory, got: %s", errMsg)
	}
}

// Mock tests that don't require DDEV to be installed
func TestManagerWithMock(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		wantNil    bool
	}{
		{
			name:       "create manager with valid directory",
			projectDir: "/test/project",
			wantNil:    false,
		},
		{
			name:       "create manager with empty directory",
			projectDir: "",
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.projectDir)

			if (manager == nil) != tt.wantNil {
				t.Errorf("NewManager() nil = %v, wantNil %v", manager == nil, tt.wantNil)
			}

			if manager != nil && manager.ProjectDir != tt.projectDir {
				t.Errorf("ProjectDir = %q, want %q", manager.ProjectDir, tt.projectDir)
			}
		})
	}
}
