package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TempDir creates a temporary directory for tests and returns the path
// The directory is automatically cleaned up when the test completes
func TempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "stax-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// WriteTestFile writes a test file with the given content
func WriteTestFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// ReadTestFile reads a test file and returns its content
func ReadTestFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	return string(content)
}

// AssertFileExists asserts that a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}

// AssertFileNotExists asserts that a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file to not exist: %s", path)
	}
}

// AssertFileContains asserts that a file contains the given text
func AssertFileContains(t *testing.T, path, content string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	if !strings.Contains(string(data), content) {
		t.Errorf("file %s does not contain expected content: %q", path, content)
	}
}

// AssertFileNotContains asserts that a file does not contain the given text
func AssertFileNotContains(t *testing.T, path, content string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	if strings.Contains(string(data), content) {
		t.Errorf("file %s contains unexpected content: %q", path, content)
	}
}

// AssertDirExists asserts that a directory exists
func AssertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("expected directory to exist: %s", path)
		return
	}

	if err != nil {
		t.Fatalf("failed to stat directory %s: %v", path, err)
	}

	if !info.IsDir() {
		t.Errorf("expected %s to be a directory, but it's a file", path)
	}
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNotEqual asserts that two values are not equal
func AssertNotEqual(t *testing.T, got, notWant interface{}) {
	t.Helper()

	if got == notWant {
		t.Errorf("got %v, expected to be different", got)
	}
}

// AssertError asserts that an error occurred
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()

	if err == nil {
		t.Errorf("%s: expected error but got nil", msg)
	}
}

// AssertNoError asserts that no error occurred
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()

	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// AssertErrorContains asserts that an error contains the given text
func AssertErrorContains(t *testing.T, err error, text string) {
	t.Helper()

	if err == nil {
		t.Errorf("expected error containing %q, but got nil", text)
		return
	}

	if !strings.Contains(err.Error(), text) {
		t.Errorf("error %q does not contain expected text %q", err.Error(), text)
	}
}

// MockExecCommand is a helper for mocking exec.Command
type MockExecCommand struct {
	Commands []MockCommand
}

// MockCommand represents a mocked command
type MockCommand struct {
	Cmd    string
	Args   []string
	Stdout string
	Stderr string
	Error  error
}

// Mock mocks an exec.Command call
func (m *MockExecCommand) Mock(cmd string, args []string, stdout, stderr string, err error) {
	m.Commands = append(m.Commands, MockCommand{
		Cmd:    cmd,
		Args:   args,
		Stdout: stdout,
		Stderr: stderr,
		Error:  err,
	})
}

// CommandExists checks if a command exists in PATH
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// SkipIfCommandNotFound skips the test if a command is not found
func SkipIfCommandNotFound(t *testing.T, command string) {
	t.Helper()

	if !CommandExists(command) {
		t.Skipf("skipping test: %s not found in PATH", command)
	}
}

// Chdir changes to a directory for the duration of a test
func Chdir(t *testing.T, dir string) {
	t.Helper()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", dir, err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("warning: failed to restore directory to %s: %v", oldDir, err)
		}
	})
}

// SetEnv sets an environment variable for the duration of a test
func SetEnv(t *testing.T, key, value string) {
	t.Helper()

	oldValue, hadOldValue := os.LookupEnv(key)

	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env var %s: %v", key, err)
	}

	t.Cleanup(func() {
		if hadOldValue {
			os.Setenv(key, oldValue)
		} else {
			os.Unsetenv(key)
		}
	})
}

// CreateTestProject creates a test project structure
func CreateTestProject(t *testing.T, dir string) {
	t.Helper()

	// Create common WordPress directories
	dirs := []string{
		"wp-content/themes",
		"wp-content/plugins",
		"wp-content/mu-plugins",
		"wp-content/uploads",
	}

	for _, d := range dirs {
		path := filepath.Join(dir, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", path, err)
		}
	}

	// Create a basic package.json
	packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "scripts": {
    "build": "echo 'building...'",
    "dev": "echo 'dev mode...'"
  }
}`
	WriteTestFile(t, filepath.Join(dir, "package.json"), packageJSON)

	// Create a basic composer.json
	composerJSON := `{
  "name": "test/project",
  "type": "project",
  "require": {
    "php": ">=7.4"
  }
}`
	WriteTestFile(t, filepath.Join(dir, "composer.json"), composerJSON)
}
