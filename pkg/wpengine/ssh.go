package wpengine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/firecrown-media/stax/pkg/security"
	"golang.org/x/crypto/ssh"
)

const (
	// DefaultSSHGateway is the default WPEngine SSH gateway hostname
	DefaultSSHGateway = "ssh.wpengine.net"

	// DefaultSSHPort is the default SSH port
	DefaultSSHPort = 22

	// DefaultSSHTimeout is the default SSH connection timeout
	DefaultSSHTimeout = 30 * time.Second
)

// SSHClient represents an SSH connection to WPEngine gateway
type SSHClient struct {
	client *ssh.Client
	config SSHConfig
}

// NewSSHClient creates a new SSH client for WPEngine
func NewSSHClient(config SSHConfig) (*SSHClient, error) {
	// Parse private key
	signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Set defaults
	if config.Host == "" {
		// WPEngine uses install-specific SSH gateways: {install}.ssh.wpengine.net
		config.Host = fmt.Sprintf("%s.ssh.wpengine.net", config.Install)
	} else if config.Host == "ssh.wpengine.net" {
		// Convert legacy generic gateway to install-specific subdomain
		config.Host = fmt.Sprintf("%s.ssh.wpengine.net", config.Install)
	}
	if config.Port == 0 {
		config.Port = DefaultSSHPort
	}

	// SSH user format: just the install name
	user := config.Install

	// Initialize known hosts manager for secure host key verification
	khManager, err := security.NewKnownHostsManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize known hosts manager: %w", err)
	}

	// Create SSH client config with proper host key verification
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: khManager.GetHostKeyCallback(),
		Timeout:         DefaultSSHTimeout,
	}

	// Connect to SSH gateway
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH gateway: %w", err)
	}

	return &SSHClient{
		client: client,
		config: config,
	}, nil
}

// SSHSource returns the full rsync-style SSH source path for a given remote path.
// remotePath may be absolute or relative to the SSH home directory.
func (c *SSHClient) SSHSource(remotePath string) string {
	return fmt.Sprintf("%s@%s:%s", c.config.Install, c.config.Host, remotePath)
}

// Close closes the SSH connection
func (c *SSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ExecuteCommand executes a command via SSH and returns the output
func (c *SSHClient) ExecuteCommand(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), nil
}

// ExecuteCommandWithOutput executes a command and streams output to given writers
func (c *SSHClient) ExecuteCommandWithOutput(cmd string, stdout, stderr io.Writer) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// GetWPCLI executes a WP-CLI command on the remote server
func (c *SSHClient) GetWPCLI(args []string) (string, error) {
	// Sanitize WP-CLI arguments to prevent command injection
	sanitizedArgs, err := security.SanitizeWPCLIArgs(args)
	if err != nil {
		return "", fmt.Errorf("invalid WP-CLI arguments: %w", err)
	}

	cmd := "wp " + strings.Join(sanitizedArgs, " ")
	return c.ExecuteCommand(cmd)
}

// DownloadFile downloads a file from the remote server
func (c *SSHClient) DownloadFile(remotePath, localPath string) error {
	// Validate and sanitize remote path to prevent path traversal
	sanitizedRemotePath, err := security.SanitizePath(remotePath)
	if err != nil {
		return fmt.Errorf("invalid remote path: %w", err)
	}

	// Sanitize for shell to prevent command injection
	safeRemotePath, err := security.SanitizeForShell(sanitizedRemotePath)
	if err != nil {
		return fmt.Errorf("remote path contains unsafe characters: %w", err)
	}

	// Validate local path
	sanitizedLocalPath, err := security.SanitizePath(localPath)
	if err != nil {
		return fmt.Errorf("invalid local path: %w", err)
	}

	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Create local file with secure permissions
	localFile, err := os.OpenFile(sanitizedLocalPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	// Setup remote command output to local file
	session.Stdout = localFile

	// Execute cat command to read remote file (path is now sanitized)
	cmd := fmt.Sprintf("cat %s", safeRemotePath)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// GetRemoteFileSize gets the size of a remote file
func (c *SSHClient) GetRemoteFileSize(remotePath string) (int64, error) {
	// Sanitize path to prevent command injection
	safePath, err := security.SanitizeForShell(remotePath)
	if err != nil {
		return 0, fmt.Errorf("invalid remote path: %w", err)
	}

	output, err := c.ExecuteCommand(fmt.Sprintf("stat -f%%z %s", safePath))
	if err != nil {
		return 0, err
	}

	var size int64
	if _, err := fmt.Sscanf(strings.TrimSpace(output), "%d", &size); err != nil {
		return 0, fmt.Errorf("failed to parse file size: %w", err)
	}

	return size, nil
}

// GetDirectorySize calculates the size of a remote directory
func (c *SSHClient) GetDirectorySize(remotePath string) (int64, error) {
	// Sanitize path to prevent command injection
	safePath, err := security.SanitizeForShell(remotePath)
	if err != nil {
		return 0, fmt.Errorf("invalid remote path: %w", err)
	}

	output, err := c.ExecuteCommand(fmt.Sprintf("du -sb %s | cut -f1", safePath))
	if err != nil {
		return 0, err
	}

	var size int64
	if _, err := fmt.Sscanf(strings.TrimSpace(output), "%d", &size); err != nil {
		return 0, fmt.Errorf("failed to parse directory size: %w", err)
	}

	return size, nil
}

// UploadFile uploads a file to the remote server via SCP
func (c *SSHClient) UploadFile(localPath, remotePath string) error {
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

	// Sanitize for shell to prevent command injection
	safeRemotePath, err := security.SanitizeForShell(sanitizedRemotePath)
	if err != nil {
		return fmt.Errorf("remote path contains unsafe characters: %w", err)
	}

	// Open local file for reading
	localFile, err := os.Open(sanitizedLocalPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get file info for size
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	// Create SSH session for upload
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Create remote file using cat
	cmd := fmt.Sprintf("cat > %s", safeRemotePath)
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Setup error capturing
	var stderr bytes.Buffer
	session.Stderr = &stderr

	// Start the remote command
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start upload command: %w", err)
	}

	// Copy file contents to remote
	written, err := io.Copy(stdin, localFile)
	if err != nil {
		stdin.Close()
		return fmt.Errorf("failed to upload file: %w", err)
	}

	// Close stdin to signal EOF
	stdin.Close()

	// Wait for the command to complete
	if err := session.Wait(); err != nil {
		return fmt.Errorf("failed to complete upload: %w (stderr: %s)", err, stderr.String())
	}

	// Verify upload size matches
	if written != fileInfo.Size() {
		return fmt.Errorf("upload size mismatch: sent %d bytes, expected %d bytes", written, fileInfo.Size())
	}

	return nil
}

// TestConnection tests the SSH connection
func (c *SSHClient) TestConnection() error {
	_, err := c.ExecuteCommand("echo 'test'")
	return err
}
