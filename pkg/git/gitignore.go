package git

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/firecrown-media/stax/pkg/ui"
)

// WordPressGitignoreEntries returns recommended .gitignore entries for WordPress projects
func WordPressGitignoreEntries() []string {
	return []string{
		"# WordPress Core (install via composer or download)",
		"/wp-admin/",
		"/wp-includes/",
		"/wp-*.php",
		"/index.php",
		"/license.txt",
		"/readme.html",
		"/xmlrpc.php",
		"",
		"# WPEngine specific",
		"mysql.sql",
		"*.sql",
		".smushit-status",
		".wpengine-conf/",
		"",
		"# WordPress uploads and generated content",
		"/wp-content/uploads/",
		"/wp-content/cache/",
		"/wp-content/upgrade/",
		"/wp-content/backup*/",
		"/wp-content/backups/",
		"/wp-content/blogs.dir/",
		"/wp-content/debug.log",
		"/wp-content/advanced-cache.php",
		"/wp-content/object-cache.php",
		"/wp-content/wp-cache-config.php",
		"/wp-content/index.php",
		"/wp-content/languages/",
		"",
		"# WordPress config",
		"wp-config.php",
		"wp-config-local.php",
		"wp-config-ddev.php",
		"",
		"# Build artifacts in themes/plugins",
		"/wp-content/themes/*/node_modules/",
		"/wp-content/plugins/*/node_modules/",
		"/wp-content/themes/*/.sass-cache/",
		"/wp-content/plugins/*/.sass-cache/",
		"",
		"# Dependency directories",
		"/vendor/",
		"/node_modules/",
		"composer.lock",
		"",
		"# Environment files",
		".env",
		".env.*",
		"*.local.php",
		"",
		"# Stax files (do not commit)",
		".stax.yml",
		".stax.yaml",
		"stax.yaml",
		".stax/",
		".stax.local.yml",
		"",
		"# DDEV (local development)",
		".ddev/",
		"!.ddev/config.yaml",
		"!.ddev/commands/",
		"",
		"# Logs",
		"*.log",
		"logs/",
		"",
		"# Editor files",
		".vscode/",
		".idea/",
		"*.swp",
		"*.swo",
		"*~",
		"",
		"# OS files",
		".DS_Store",
		".DS_Store?",
		"._*",
		".Spotlight-V100",
		".Trashes",
		"Thumbs.db",
	}
}

// EnsureGitignore creates or updates .gitignore with WordPress and Stax entries
func EnsureGitignore(projectDir string) error {
	gitignorePath := filepath.Join(projectDir, ".gitignore")

	// Check if project is a git repository
	gitDir := filepath.Join(projectDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Not a git repository, skip .gitignore creation
		ui.Debug("Not a git repository, skipping .gitignore creation")
		return nil
	}

	// Check if .gitignore exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// Create new .gitignore
		return createGitignore(gitignorePath)
	}

	// .gitignore exists, update it
	return updateGitignore(gitignorePath)
}

// createGitignore creates a new .gitignore file with recommended entries
func createGitignore(gitignorePath string) error {
	ui.Info("Creating .gitignore with WordPress and Stax entries...")

	file, err := os.Create(gitignorePath)
	if err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	defer file.Close()

	entries := WordPressGitignoreEntries()
	for _, entry := range entries {
		if _, err := file.WriteString(entry + "\n"); err != nil {
			return fmt.Errorf("failed to write to .gitignore: %w", err)
		}
	}

	ui.Success(".gitignore created successfully")
	return nil
}

// updateGitignore adds missing Stax-specific entries to existing .gitignore
func updateGitignore(gitignorePath string) error {
	// Read existing entries
	existingEntries, err := readGitignoreEntries(gitignorePath)
	if err != nil {
		return err
	}

	// Determine which Stax entries are missing
	staxEntries := getStaxSpecificEntries()
	missingEntries := []string{}

	for _, entry := range staxEntries {
		// Skip comments and empty lines
		if strings.HasPrefix(entry, "#") || entry == "" {
			continue
		}

		if !containsEntry(existingEntries, entry) {
			missingEntries = append(missingEntries, entry)
		}
	}

	// If no missing entries, we're done
	if len(missingEntries) == 0 {
		ui.Debug(".gitignore already contains Stax entries")
		return nil
	}

	// Append missing entries
	ui.Info(fmt.Sprintf("Adding %d Stax-specific entries to .gitignore...", len(missingEntries)))

	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer file.Close()

	// Add a section header
	if _, err := file.WriteString("\n# Stax configuration (added by stax init)\n"); err != nil {
		return fmt.Errorf("failed to write to .gitignore: %w", err)
	}

	// Add missing entries
	for _, entry := range missingEntries {
		if _, err := file.WriteString(entry + "\n"); err != nil {
			return fmt.Errorf("failed to write to .gitignore: %w", err)
		}
	}

	ui.Success(fmt.Sprintf("Added %d entries to .gitignore", len(missingEntries)))
	return nil
}

// readGitignoreEntries reads all non-comment, non-empty entries from .gitignore
func readGitignoreEntries(gitignorePath string) (map[string]bool, error) {
	entries := make(map[string]bool)

	file, err := os.Open(gitignorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .gitignore: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			entries[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .gitignore: %w", err)
	}

	return entries, nil
}

// getStaxSpecificEntries returns entries that Stax wants in .gitignore
func getStaxSpecificEntries() []string {
	return []string{
		"# Stax configuration",
		".stax.yml",
		".stax.yaml",
		"stax.yaml",
		".stax/",
		"",
		"# Environment files",
		".env",
		".env.local",
		"",
		"# WordPress config (may contain credentials)",
		"wp-config.php",
		"wp-config-local.php",
	}
}

// containsEntry checks if a .gitignore entry already exists
func containsEntry(entries map[string]bool, entry string) bool {
	// Check exact match
	if entries[entry] {
		return true
	}

	// Check if entry is a directory pattern (with or without trailing slash)
	entryWithSlash := strings.TrimSuffix(entry, "/") + "/"
	entryWithoutSlash := strings.TrimSuffix(entry, "/")

	return entries[entryWithSlash] || entries[entryWithoutSlash]
}
