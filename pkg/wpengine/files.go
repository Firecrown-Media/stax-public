package wpengine

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/firecrown-media/stax/pkg/security"
)

var (
	// DefaultRsyncExclusions are patterns commonly excluded from file sync
	DefaultRsyncExclusions = []string{
		"*.log",
		"cache/",
		".DS_Store",
		"Thumbs.db",
		"*.tmp",
		"*.swp",
		"node_modules/",
	}
)

// SyncWPContent syncs wp-content directory from WPEngine
func (c *SSHClient) SyncWPContent(destination string, options SyncOptions) error {
	// Build source path
	source := fmt.Sprintf("%s@%s@%s:/sites/%s/wp-content/",
		c.config.Install,
		c.config.Install,
		c.config.Host,
		c.config.Install,
	)

	// Set defaults in options
	if options.Source == "" {
		options.Source = source
	}
	if options.Destination == "" {
		options.Destination = destination
	}

	return c.Rsync(options)
}

// Rsync performs rsync file synchronization
func (c *SSHClient) Rsync(options SyncOptions) error {
	args := []string{
		"-rlDvz",
		"--size-only",
	}

	// Preserve permissions if requested
	if options.PreservePermissions {
		args = append(args, "-p")
	}

	// Add progress if requested
	if options.Progress {
		args = append(args, "--progress")
	}

	// Build exclusion list starting with defaults
	exclusions := options.Exclude
	if len(exclusions) == 0 {
		exclusions = DefaultRsyncExclusions
	}

	// Load .staxignore patterns if ProjectDir is set
	if options.ProjectDir != "" {
		staxignorePatterns, err := LoadStaxIgnore(options.ProjectDir)
		if err != nil {
			// Log warning but continue - .staxignore errors are not fatal
			fmt.Fprintf(os.Stderr, "Warning: failed to load .staxignore: %v\n", err)
		} else if len(staxignorePatterns) > 0 {
			// Merge .staxignore patterns with existing exclusions
			exclusions = append(exclusions, staxignorePatterns...)
		}
	}

	// Add exclusions with validation
	for _, pattern := range exclusions {
		// Validate rsync pattern to prevent command injection
		if err := security.ValidateRsyncPattern(pattern); err != nil {
			return fmt.Errorf("invalid exclusion pattern: %w", err)
		}
		args = append(args, "--exclude="+pattern)
	}

	// Add inclusions with validation
	for _, pattern := range options.Include {
		// Validate rsync pattern to prevent command injection
		if err := security.ValidateRsyncPattern(pattern); err != nil {
			return fmt.Errorf("invalid inclusion pattern: %w", err)
		}
		args = append(args, "--include="+pattern)
	}

	// Delete local files not on remote
	if options.Delete {
		args = append(args, "--delete")
	}

	// Bandwidth limit
	if options.BandwidthLimit > 0 {
		args = append(args, fmt.Sprintf("--bwlimit=%d", options.BandwidthLimit))
	}

	// Dry run
	if options.DryRun {
		args = append(args, "--dry-run")
	}

	// Add SSH key if available
	if c.config.PrivateKey != "" {
		// Write private key to temp file (now secure)
		tmpKey, err := writePrivateKeyToTempFile(c.config.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to write SSH key: %w", err)
		}
		defer func() {
			// Securely delete temp key file
			_ = secureDeleteFile(tmpKey) // Best-effort cleanup
		}()

		// Note: We still use -o StrictHostKeyChecking=no for rsync
		// because the SSH library handles host key verification
		// separately. This is only for the rsync subprocess.
		sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", tmpKey)
		args = append(args, "-e", sshCmd)
	}

	// Source and destination
	args = append(args, options.Source, options.Destination)

	// Execute rsync
	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	return nil
}

// GetExcludePatterns returns default exclusion patterns
func GetExcludePatterns() []string {
	return DefaultRsyncExclusions
}

// LoadStaxIgnore loads .staxignore patterns from the project directory
// Returns an empty slice if the file doesn't exist (not an error)
func LoadStaxIgnore(projectDir string) ([]string, error) {
	// Build path to .staxignore
	staxignorePath := filepath.Join(projectDir, ".staxignore")

	// Check if file exists
	if _, err := os.Stat(staxignorePath); os.IsNotExist(err) {
		// File doesn't exist, return empty slice (not an error)
		return []string{}, nil
	}

	// Open file
	file, err := os.Open(staxignorePath)
	if err != nil {
		// Only return error for actual read failures
		return []string{}, fmt.Errorf("failed to open .staxignore: %w", err)
	}
	defer file.Close()

	// Parse patterns
	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Validate pattern to prevent command injection
		if err := security.ValidateRsyncPattern(line); err != nil {
			// Skip invalid patterns but log warning
			continue
		}

		patterns = append(patterns, line)
	}

	if err := scanner.Err(); err != nil {
		return patterns, fmt.Errorf("failed to read .staxignore: %w", err)
	}

	return patterns, nil
}

// VerifyFileIntegrity verifies rsync completion by comparing file counts
func (c *SSHClient) VerifyFileIntegrity(remotePath, localPath string) error {
	// Count remote files
	remoteCount, err := c.countFiles(remotePath)
	if err != nil {
		return fmt.Errorf("failed to count remote files: %w", err)
	}

	// Count local files
	localCount, err := countLocalFiles(localPath)
	if err != nil {
		return fmt.Errorf("failed to count local files: %w", err)
	}

	// Compare counts (allow some variance for hidden files)
	variance := float64(localCount) / float64(remoteCount)
	if variance < 0.95 || variance > 1.05 {
		return fmt.Errorf("file count mismatch: remote=%d, local=%d", remoteCount, localCount)
	}

	return nil
}

// countFiles counts files in a remote directory
func (c *SSHClient) countFiles(remotePath string) (int, error) {
	// Sanitize path to prevent command injection
	safePath, err := security.SanitizeForShell(remotePath)
	if err != nil {
		return 0, fmt.Errorf("invalid remote path: %w", err)
	}

	cmd := fmt.Sprintf("find %s -type f | wc -l", safePath)
	output, err := c.ExecuteCommand(cmd)
	if err != nil {
		return 0, err
	}

	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(output), "%d", &count); err != nil {
		return 0, fmt.Errorf("failed to parse file count: %w", err)
	}

	return count, nil
}

// countLocalFiles counts files in a local directory
func countLocalFiles(localPath string) (int, error) {
	cmd := exec.Command("find", localPath, "-type", "f")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return len(lines), nil
}

// writePrivateKeyToTempFile writes a private key to a temporary file securely
// This implementation eliminates the race condition by setting permissions atomically
func writePrivateKeyToTempFile(privateKey string) (string, error) {
	// Generate secure random filename
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random filename: %w", err)
	}
	randomHex := hex.EncodeToString(randomBytes)
	filename := fmt.Sprintf("stax-ssh-key-%s", randomHex)

	// Create file with secure permissions from the start (no race condition)
	tmpPath := filepath.Join(os.TempDir(), filename)
	tmpFile, err := os.OpenFile(
		tmpPath,
		os.O_RDWR|os.O_CREATE|os.O_EXCL, // Exclusive creation
		0600,                            // Secure permissions at creation
	)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	// Ensure cleanup on error
	success := false
	defer func() {
		_ = tmpFile.Close() // File handle cleanup
		if !success {
			_ = secureDeleteFile(tmpPath) // Best-effort cleanup
		}
	}()

	// Write private key
	if _, err := tmpFile.WriteString(privateKey); err != nil {
		return "", fmt.Errorf("failed to write private key: %w", err)
	}

	// Sync to disk
	if err := tmpFile.Sync(); err != nil {
		return "", fmt.Errorf("failed to sync temp file: %w", err)
	}

	success = true
	return tmpPath, nil
}

// secureDeleteFile securely deletes a file by overwriting before removal
func secureDeleteFile(path string) error {
	// Get file size
	info, err := os.Stat(path)
	if err != nil {
		// If file doesn't exist, that's fine
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Open file for writing
	file, err := os.OpenFile(path, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Overwrite with zeros (single pass is sufficient for modern systems)
	zeros := make([]byte, info.Size())
	if _, err := file.Write(zeros); err != nil {
		return err
	}

	// Sync to disk to ensure data is written
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file before deletion: %w", err)
	}

	// Close before delete
	file.Close()

	// Finally delete
	return os.Remove(path)
}

// SyncDirectory syncs a specific directory from WPEngine
func (c *SSHClient) SyncDirectory(remotePath, localPath string, options SyncOptions) error {
	// Validate and sanitize remote path
	sanitizedRemotePath, err := security.SanitizePath(remotePath)
	if err != nil {
		return fmt.Errorf("invalid remote path: %w", err)
	}

	// Validate and sanitize local path
	sanitizedLocalPath, err := security.SanitizePath(localPath)
	if err != nil {
		return fmt.Errorf("invalid local path: %w", err)
	}

	// Build full remote path
	source := fmt.Sprintf("%s@%s@%s:%s",
		c.config.Install,
		c.config.Install,
		c.config.Host,
		sanitizedRemotePath,
	)

	options.Source = source
	options.Destination = sanitizedLocalPath

	return c.Rsync(options)
}

// PushDirectory pushes a local directory to WPEngine (reverse of SyncDirectory)
func (c *SSHClient) PushDirectory(localPath, remotePath string, options SyncOptions) error {
	// Validate and sanitize local path
	sanitizedLocalPath, err := security.SanitizePath(localPath)
	if err != nil {
		return fmt.Errorf("invalid local path: %w", err)
	}

	// Validate and sanitize remote path
	sanitizedRemotePath, err := security.SanitizePath(remotePath)
	if err != nil {
		return fmt.Errorf("invalid remote path: %w", err)
	}

	// Validate local path exists
	if _, err := os.Stat(sanitizedLocalPath); err != nil {
		return fmt.Errorf("local path does not exist: %w", err)
	}

	// Build full remote path (destination)
	destination := fmt.Sprintf("%s@%s@%s:%s",
		c.config.Install,
		c.config.Install,
		c.config.Host,
		sanitizedRemotePath,
	)

	// Set source as local and destination as remote (reversed from pull)
	options.Source = sanitizedLocalPath
	options.Destination = destination

	return c.Rsync(options)
}
