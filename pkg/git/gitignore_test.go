package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWordPressGitignoreEntries(t *testing.T) {
	entries := WordPressGitignoreEntries()

	// Should have some entries
	if len(entries) == 0 {
		t.Error("Expected some gitignore entries, got none")
	}

	// Check for critical entries
	entriesStr := strings.Join(entries, "\n")
	critical := []string{".stax.yml", "wp-config.php", ".env", "node_modules/", "/vendor/"}

	for _, entry := range critical {
		if !strings.Contains(entriesStr, entry) {
			t.Errorf("Expected critical entry %q to be present", entry)
		}
	}
}

func TestEnsureGitignoreNonGitRepo(t *testing.T) {
	// Create temp directory without .git
	tmpDir, err := os.MkdirTemp("", "stax-test-nogit-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Should skip .gitignore creation
	err = EnsureGitignore(tmpDir)
	if err != nil {
		t.Errorf("Expected no error for non-git directory, got: %v", err)
	}

	// .gitignore should not be created
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); !os.IsNotExist(err) {
		t.Error("Expected .gitignore to not be created for non-git directory")
	}
}

func TestEnsureGitignoreNewFile(t *testing.T) {
	// Create temp directory with .git
	tmpDir, err := os.MkdirTemp("", "stax-test-git-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Should create .gitignore
	err = EnsureGitignore(tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// .gitignore should be created
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error("Expected .gitignore to be created")
	}

	// Check contents
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, ".stax.yml") {
		t.Error("Expected .gitignore to contain .stax.yml")
	}
	if !strings.Contains(contentStr, "wp-config.php") {
		t.Error("Expected .gitignore to contain wp-config.php")
	}
}

func TestEnsureGitignoreUpdateExisting(t *testing.T) {
	// Create temp directory with .git
	tmpDir, err := os.MkdirTemp("", "stax-test-git-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create existing .gitignore with some entries
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	existingContent := "node_modules/\n.DS_Store\n"
	if err := os.WriteFile(gitignorePath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create existing .gitignore: %v", err)
	}

	// Should update .gitignore
	err = EnsureGitignore(tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check updated contents
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	contentStr := string(content)
	// Should preserve existing entries
	if !strings.Contains(contentStr, "node_modules/") {
		t.Error("Expected .gitignore to preserve existing node_modules/ entry")
	}
	if !strings.Contains(contentStr, ".DS_Store") {
		t.Error("Expected .gitignore to preserve existing .DS_Store entry")
	}
	// Should add Stax entries
	if !strings.Contains(contentStr, ".stax.yml") {
		t.Error("Expected .gitignore to add .stax.yml")
	}
	if !strings.Contains(contentStr, "wp-config.php") {
		t.Error("Expected .gitignore to add wp-config.php")
	}
}

func TestContainsEntry(t *testing.T) {
	entries := map[string]bool{
		"node_modules/": true,
		".env":          true,
		".stax/":        true,
	}

	tests := []struct {
		name     string
		entry    string
		expected bool
	}{
		{"exact match", "node_modules/", true},
		{"directory with slash", ".stax/", true},
		{"directory without slash", ".stax", true},
		{"not present", ".gitignore", false},
		{"partial match should not work", "node", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsEntry(entries, tt.entry)
			if result != tt.expected {
				t.Errorf("containsEntry(%q) = %v, expected %v", tt.entry, result, tt.expected)
			}
		})
	}
}
