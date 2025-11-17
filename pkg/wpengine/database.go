package wpengine

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/firecrown-media/stax/pkg/security"
	"golang.org/x/crypto/ssh"
)

var (
	// DefaultExcludedTables are tables commonly excluded from database exports
	DefaultExcludedTables = []string{
		"actionscheduler_logs",
		"actionscheduler_actions",
		"wc_admin_notes",
		"wc_admin_note_actions",
	}
)

// GetTablePrefix detects the WordPress table prefix from the database
func (c *SSHClient) GetTablePrefix() (string, error) {
	// Use WP-CLI's config command instead of direct SQL query for better security
	cmd := `wp config get table_prefix`
	output, err := c.ExecuteCommand(cmd)
	if err != nil {
		// Fallback to query method if config command fails
		return c.getTablePrefixViaQuery()
	}

	prefix := strings.TrimSpace(output)
	if prefix == "" {
		return "", fmt.Errorf("table prefix is empty")
	}

	// Validate prefix format to prevent SQL injection
	if err := security.ValidateTablePrefix(prefix); err != nil {
		return "", fmt.Errorf("invalid table prefix: %w", err)
	}

	return prefix, nil
}

// getTablePrefixViaQuery is a fallback method using database query
func (c *SSHClient) getTablePrefixViaQuery() (string, error) {
	cmd := `wp db query "SHOW TABLES LIKE '%_options'" --skip-column-names`
	output, err := c.ExecuteCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to detect table prefix: %w", err)
	}

	tableName := strings.TrimSpace(output)
	if tableName == "" {
		return "", fmt.Errorf("no options table found")
	}

	// Extract prefix from table name (e.g., "wp_options" -> "wp_")
	if !strings.Contains(tableName, "_options") {
		return "", fmt.Errorf("unexpected table name format: %s", tableName)
	}

	prefix := strings.TrimSuffix(tableName, "options")

	// Validate prefix format
	if err := security.ValidateTablePrefix(prefix); err != nil {
		return "", fmt.Errorf("invalid table prefix: %w", err)
	}

	return prefix, nil
}

// GenerateExcludePattern generates the table exclusion pattern for wp db export
func GenerateExcludePattern(prefix string, options DatabaseOptions) (string, error) {
	// Validate prefix first
	if err := security.ValidateTablePrefix(prefix); err != nil {
		return "", fmt.Errorf("invalid table prefix: %w", err)
	}

	var tables []string

	// Add default exclusions
	if options.SkipLogs {
		tables = append(tables,
			prefix+"actionscheduler_logs",
			prefix+"actionscheduler_actions",
		)
	}

	if options.SkipSpam {
		// Note: Can't exclude rows via WP-CLI, only tables
		// Spam filtering would need to be done post-export
	}

	// Add user-specified exclusions with validation
	for _, table := range options.ExcludeTables {
		// Validate table name
		tableName := table
		if !strings.HasPrefix(table, prefix) {
			tableName = prefix + table
		}

		// Validate the full table name
		if err := security.ValidateTableName(tableName); err != nil {
			return "", fmt.Errorf("invalid table name %q: %w", tableName, err)
		}

		tables = append(tables, tableName)
	}

	if len(tables) == 0 {
		return "", nil
	}

	return strings.Join(tables, ","), nil
}

// ExportDatabase exports the database from WPEngine
func (c *SSHClient) ExportDatabase(options DatabaseOptions) (io.ReadCloser, error) {
	// Detect table prefix
	prefix, err := c.GetTablePrefix()
	if err != nil {
		return nil, err
	}

	// Build export command
	cmd := "wp db export --add-drop-table"

	// Add table exclusions with validation
	excludePattern, err := GenerateExcludePattern(prefix, options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate exclusion pattern: %w", err)
	}
	if excludePattern != "" {
		cmd += fmt.Sprintf(" --exclude_tables=%s", excludePattern)
	}

	// Export to stdout
	cmd += " -"

	// Create SSH session for streaming
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := session.Start(cmd); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start export command: %w", err)
	}

	// Return a ReadCloser that closes both the pipe and the session
	return &exportReadCloser{
		reader:  stdout,
		session: session,
	}, nil
}

// exportReadCloser wraps an io.Reader and closes the SSH session when done
type exportReadCloser struct {
	reader  io.Reader
	session *ssh.Session
}

func (e *exportReadCloser) Read(p []byte) (n int, err error) {
	return e.reader.Read(p)
}

func (e *exportReadCloser) Close() error {
	if e.session != nil {
		// Wait for command to complete - ignore error as we're already closing
		_ = e.session.Wait()
		return e.session.Close()
	}
	return nil
}

// CalculateExportSize estimates the database export size
func (c *SSHClient) CalculateExportSize() (int64, error) {
	// Query total database size
	cmd := `wp db query "SELECT SUM(data_length + index_length) as size FROM information_schema.TABLES WHERE table_schema = DATABASE()" --skip-column-names`
	output, err := c.ExecuteCommand(cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate database size: %w", err)
	}

	var size int64
	if _, err := fmt.Sscanf(strings.TrimSpace(output), "%d", &size); err != nil {
		return 0, fmt.Errorf("failed to parse database size: %w", err)
	}

	return size, nil
}

// StreamDatabase streams a database from a reader to a destination file
func StreamDatabase(reader io.Reader, destination string) (int64, error) {
	// Create destination file
	file, err := os.Create(destination)
	if err != nil {
		return 0, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer file.Close()

	// Copy data
	written, err := io.Copy(file, reader)
	if err != nil {
		return 0, fmt.Errorf("failed to stream database: %w", err)
	}

	return written, nil
}

// GetDatabaseName gets the WordPress database name
func (c *SSHClient) GetDatabaseName() (string, error) {
	cmd := `wp db query "SELECT DATABASE()" --skip-column-names`
	output, err := c.ExecuteCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to get database name: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// GetTableCount gets the number of tables in the database
func (c *SSHClient) GetTableCount() (int, error) {
	cmd := `wp db query "SELECT COUNT(*) FROM information_schema.TABLES WHERE table_schema = DATABASE()" --skip-column-names`
	output, err := c.ExecuteCommand(cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to get table count: %w", err)
	}

	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(output), "%d", &count); err != nil {
		return 0, fmt.Errorf("failed to parse table count: %w", err)
	}

	return count, nil
}

// ImportDatabase imports a database file on the remote server
func (c *SSHClient) ImportDatabase(remotePath string) error {
	// Validate and sanitize remote path
	sanitizedPath, err := security.SanitizePath(remotePath)
	if err != nil {
		return fmt.Errorf("invalid remote path: %w", err)
	}

	// Sanitize for shell
	safePath, err := security.SanitizeForShell(sanitizedPath)
	if err != nil {
		return fmt.Errorf("remote path contains unsafe characters: %w", err)
	}

	// Import database using wp db import
	cmd := fmt.Sprintf("wp db import %s", safePath)
	output, err := c.ExecuteCommand(cmd)
	if err != nil {
		return fmt.Errorf("database import failed: %w", err)
	}

	// Check for success message
	if !strings.Contains(output, "Success") {
		return fmt.Errorf("database import did not report success: %s", output)
	}

	return nil
}

// RemoveFile removes a file from the remote server
func (c *SSHClient) RemoveFile(remotePath string) error {
	// Validate and sanitize remote path
	sanitizedPath, err := security.SanitizePath(remotePath)
	if err != nil {
		return fmt.Errorf("invalid remote path: %w", err)
	}

	// Sanitize for shell
	safePath, err := security.SanitizeForShell(sanitizedPath)
	if err != nil {
		return fmt.Errorf("remote path contains unsafe characters: %w", err)
	}

	// Remove the file
	cmd := fmt.Sprintf("rm -f %s", safePath)
	_, err = c.ExecuteCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}
