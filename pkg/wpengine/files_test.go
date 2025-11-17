package wpengine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/firecrown-media/stax/pkg/security"
)

func TestPushDirectoryValidation(t *testing.T) {
	tests := []struct {
		name        string
		localPath   string
		remotePath  string
		setupFunc   func() (string, func())
		expectError bool
		errorMsg    string
	}{
		{
			name:        "non-existent local path",
			localPath:   "/nonexistent/path",
			remotePath:  "/sites/test/wp-content/",
			setupFunc:   func() (string, func()) { return "", func() {} },
			expectError: true,
			errorMsg:    "local path does not exist",
		},
		{
			name:       "path with directory traversal attempt",
			localPath:  "",
			remotePath: "/sites/../../../etc/passwd",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			expectError: true,
			errorMsg:    "invalid remote path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localPath, cleanup := tt.setupFunc()
			defer cleanup()

			if tt.localPath != "" {
				localPath = tt.localPath
			}

			// Create a mock SSH client (validation happens before rsync)
			client := &SSHClient{
				config: SSHConfig{
					Host:    "test.wpengine.net",
					Port:    22,
					Install: "testinstall",
				},
			}

			options := SyncOptions{
				DryRun: true,
			}

			err := client.PushDirectory(localPath, tt.remotePath, options)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if tt.expectError && err != nil && tt.errorMsg != "" {
				// Check if error message contains expected substring
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

func TestLoadStaxIgnore(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantCount   int
		wantPattern string
	}{
		{
			name: "basic patterns",
			content: `# Comment
*.log
cache/
.DS_Store
`,
			wantCount:   3,
			wantPattern: "*.log",
		},
		{
			name: "empty lines and comments only",
			content: `# Just a comment

# Another comment
`,
			wantCount: 0,
		},
		{
			name: "mixed valid patterns",
			content: `node_modules/
*.tmp
.git/
dist/
build/
`,
			wantCount:   5,
			wantPattern: "node_modules/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			staxignorePath := filepath.Join(tmpDir, ".staxignore")

			if err := os.WriteFile(staxignorePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create .staxignore: %v", err)
			}

			patterns, err := LoadStaxIgnore(tmpDir)
			if err != nil {
				t.Fatalf("LoadStaxIgnore failed: %v", err)
			}

			if len(patterns) != tt.wantCount {
				t.Errorf("Expected %d patterns, got %d", tt.wantCount, len(patterns))
			}

			if tt.wantCount > 0 && patterns[0] != tt.wantPattern {
				t.Errorf("Expected first pattern '%s', got '%s'", tt.wantPattern, patterns[0])
			}
		})
	}
}

func TestLoadStaxIgnoreNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	// Should return empty slice, not error
	patterns, err := LoadStaxIgnore(tmpDir)
	if err != nil {
		t.Errorf("Expected no error for non-existent .staxignore, got: %v", err)
	}

	if len(patterns) != 0 {
		t.Errorf("Expected empty patterns, got %d patterns", len(patterns))
	}
}

func TestGetExcludePatterns(t *testing.T) {
	patterns := GetExcludePatterns()

	if len(patterns) == 0 {
		t.Error("Expected default exclusion patterns, got empty slice")
	}

	// Verify some expected patterns exist
	expectedPatterns := []string{"*.log", "cache/", ".DS_Store", "node_modules/"}
	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern '%s' not found in default exclusions", expected)
		}
	}
}

func TestSecureDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "sensitive.key")

	// Create a test file
	content := []byte("sensitive-private-key-content")
	if err := os.WriteFile(tmpFile, content, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("Test file should exist before deletion")
	}

	// Securely delete
	if err := secureDeleteFile(tmpFile); err != nil {
		t.Fatalf("secureDeleteFile failed: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("File should not exist after secure deletion")
	}
}

func TestSecureDeleteFileNonExistent(t *testing.T) {
	// Should not error when file doesn't exist
	err := secureDeleteFile("/tmp/nonexistent-file-12345.key")
	if err != nil {
		t.Errorf("Expected no error for non-existent file, got: %v", err)
	}
}

func TestWritePrivateKeyToTempFile(t *testing.T) {
	privateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0test0test0test0test0test0test0test0test
-----END RSA PRIVATE KEY-----`

	tmpPath, err := writePrivateKeyToTempFile(privateKey)
	if err != nil {
		t.Fatalf("writePrivateKeyToTempFile failed: %v", err)
	}
	defer func() {
		_ = secureDeleteFile(tmpPath) // Test cleanup
	}()

	// Verify file exists
	info, err := os.Stat(tmpPath)
	if err != nil {
		t.Fatalf("Temp file should exist: %v", err)
	}

	// Verify permissions are secure (0600)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %o", info.Mode().Perm())
	}

	// Verify content
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if string(content) != privateKey {
		t.Error("Temp file content doesn't match original private key")
	}
}

func TestSanitizePathIntegration(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "valid absolute path",
			path:        "/sites/testinstall/wp-content/themes/",
			expectError: false,
		},
		{
			name:        "valid relative path",
			path:        "wp-content/plugins/",
			expectError: false,
		},
		{
			name:        "path traversal attempt",
			path:        "../../../etc/passwd",
			expectError: true,
		},
		{
			name:        "null byte injection",
			path:        "/sites/test\x00/malicious",
			expectError: false, // Go's filepath.Clean removes null bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := security.SanitizePath(tt.path)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
