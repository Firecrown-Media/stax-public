package security

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// KnownHostsManager manages SSH known hosts using TOFU (Trust On First Use) pattern
type KnownHostsManager struct {
	knownHostsFile string
	mu             sync.RWMutex
}

// NewKnownHostsManager creates a new known hosts manager
func NewKnownHostsManager() (*KnownHostsManager, error) {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create .stax directory if it doesn't exist
	staxDir := filepath.Join(homeDir, ".stax")
	if err := os.MkdirAll(staxDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create .stax directory: %w", err)
	}

	knownHostsFile := filepath.Join(staxDir, "known_hosts")

	return &KnownHostsManager{
		knownHostsFile: knownHostsFile,
	}, nil
}

// GetHostKeyCallback returns a callback for SSH client configuration
func (m *KnownHostsManager) GetHostKeyCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return m.VerifyHostKey(hostname, remote, key)
	}
}

// VerifyHostKey verifies a host key using TOFU pattern, with fallback to ~/.ssh/known_hosts
func (m *KnownHostsManager) VerifyHostKey(hostname string, remote net.Addr, key ssh.PublicKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Normalize hostname (remove port if present)
	host := hostname
	if h, _, err := net.SplitHostPort(hostname); err == nil {
		host = h
	}

	// Try stax's own known_hosts first
	knownKey, err := m.loadHostKey(host)
	if err == nil {
		// Found in stax known_hosts — compare keys
		if !bytes.Equal(key.Marshal(), knownKey) {
			return m.handleKeyMismatch(host, key.Marshal(), knownKey)
		}
		return nil
	}

	// Not in stax known_hosts — check ~/.ssh/known_hosts as fallback
	if m.checkSystemKnownHosts(hostname, remote, key) == nil {
		// Trusted by system — save to stax known_hosts for future use
		_ = m.saveHostKey(host, key)
		return nil
	}

	// Not trusted anywhere — prompt user
	return m.handleFirstConnection(host, key)
}

// checkSystemKnownHosts checks whether a host key is already trusted in ~/.ssh/known_hosts
func (m *KnownHostsManager) checkSystemKnownHosts(hostname string, remote net.Addr, key ssh.PublicKey) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	systemKnownHosts := filepath.Join(homeDir, ".ssh", "known_hosts")
	if _, err := os.Stat(systemKnownHosts); err != nil {
		return err
	}
	cb, err := knownhosts.New(systemKnownHosts)
	if err != nil {
		return err
	}
	return cb(hostname, remote, key)
}

// handleFirstConnection prompts user to accept host key on first connection
func (m *KnownHostsManager) handleFirstConnection(hostname string, key ssh.PublicKey) error {
	fingerprint := formatFingerprint(key)

	fmt.Printf("\n")
	fmt.Printf("WARNING: The authenticity of host '%s' can't be established.\n", hostname)
	fmt.Printf("SSH key fingerprint is: %s\n", fingerprint)
	fmt.Printf("Are you sure you want to continue connecting? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "yes" {
		return fmt.Errorf("host key verification failed: user rejected host key")
	}

	// User accepted - save the key
	if err := m.saveHostKey(hostname, key); err != nil {
		return fmt.Errorf("failed to save host key: %w", err)
	}

	fmt.Printf("Host '%s' added to known hosts.\n\n", hostname)

	return nil
}

// handleKeyMismatch handles the case when a host's key has changed
func (m *KnownHostsManager) handleKeyMismatch(hostname string, newKey, oldKey []byte) error {
	oldFingerprint := formatFingerprintFromBytes(oldKey)
	newFingerprint := formatFingerprintFromBytes(newKey)

	fmt.Printf("\n")
	fmt.Printf("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@\n")
	fmt.Printf("@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @\n")
	fmt.Printf("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@\n")
	fmt.Printf("\n")
	fmt.Printf("IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!\n")
	fmt.Printf("Someone could be eavesdropping on you right now (man-in-the-middle attack)!\n")
	fmt.Printf("It is also possible that the host key has just been changed.\n")
	fmt.Printf("\n")
	fmt.Printf("Host: %s\n", hostname)
	fmt.Printf("Old fingerprint: %s\n", oldFingerprint)
	fmt.Printf("New fingerprint: %s\n", newFingerprint)
	fmt.Printf("\n")
	fmt.Printf("Known hosts file: %s\n", m.knownHostsFile)
	fmt.Printf("\n")
	fmt.Printf("Do you want to update the host key? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "yes" {
		return fmt.Errorf("host key verification failed: key mismatch")
	}

	// User accepted the new key - update it (newKey is already []byte)
	if err := m.updateHostKeyBytes(hostname, newKey); err != nil {
		return fmt.Errorf("failed to update host key: %w", err)
	}

	fmt.Printf("Host key updated for '%s'.\n\n", hostname)

	return nil
}

// loadHostKey loads a host key from the known hosts file
func (m *KnownHostsManager) loadHostKey(hostname string) ([]byte, error) {
	file, err := os.Open(m.knownHostsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("host not in known hosts")
		}
		return nil, fmt.Errorf("failed to open known hosts file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse line: hostname key-type key-data
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		if parts[0] == hostname {
			// Found the host - decode the key
			keyData, err := base64.StdEncoding.DecodeString(parts[2])
			if err != nil {
				return nil, fmt.Errorf("failed to decode host key: %w", err)
			}
			return keyData, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading known hosts file: %w", err)
	}

	return nil, fmt.Errorf("host not in known hosts")
}

// saveHostKey saves a host key to the known hosts file
func (m *KnownHostsManager) saveHostKey(hostname string, key ssh.PublicKey) error {
	// Create file if it doesn't exist
	file, err := os.OpenFile(m.knownHostsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open known hosts file: %w", err)
	}
	defer file.Close()

	// Format: hostname key-type key-data
	keyType := key.Type()
	keyData := base64.StdEncoding.EncodeToString(key.Marshal())

	line := fmt.Sprintf("%s %s %s\n", hostname, keyType, keyData)

	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("failed to write host key: %w", err)
	}

	return nil
}

// updateHostKey updates a host key in the known hosts file
func (m *KnownHostsManager) updateHostKey(hostname string, newKey ssh.PublicKey) error {
	return m.updateHostKeyBytes(hostname, newKey.Marshal())
}

// updateHostKeyBytes updates a host key from bytes
func (m *KnownHostsManager) updateHostKeyBytes(hostname string, newKeyBytes []byte) error {
	// Read all lines
	lines, err := m.readKnownHostsFile()
	if err != nil {
		return err
	}

	// Remove old entry
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 0 && parts[0] != hostname {
			filtered = append(filtered, line)
		}
	}

	// Write back
	if err := m.writeKnownHostsFile(filtered); err != nil {
		return err
	}

	// Parse the key bytes back to PublicKey to save
	// We need to know the key type, so we'll parse it
	// For simplicity, we'll write the raw bytes
	return m.saveHostKeyBytes(hostname, newKeyBytes)
}

// saveHostKeyBytes saves raw key bytes to the known hosts file
func (m *KnownHostsManager) saveHostKeyBytes(hostname string, keyBytes []byte) error {
	// Parse the key to get its type
	key, err := ssh.ParsePublicKey(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	return m.saveHostKey(hostname, key)
}

// readKnownHostsFile reads all lines from known hosts file
func (m *KnownHostsManager) readKnownHostsFile() ([]string, error) {
	file, err := os.Open(m.knownHostsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to open known hosts file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading known hosts file: %w", err)
	}

	return lines, nil
}

// writeKnownHostsFile writes lines to known hosts file
func (m *KnownHostsManager) writeKnownHostsFile(lines []string) error {
	file, err := os.OpenFile(m.knownHostsFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open known hosts file: %w", err)
	}
	defer file.Close()

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write known hosts file: %w", err)
		}
	}

	return nil
}

// RemoveHostKey removes a host key from known hosts
func (m *KnownHostsManager) RemoveHostKey(hostname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lines, err := m.readKnownHostsFile()
	if err != nil {
		return err
	}

	// Filter out the host
	filtered := make([]string, 0, len(lines))
	found := false
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 0 && parts[0] == hostname {
			found = true
			continue
		}
		filtered = append(filtered, line)
	}

	if !found {
		return fmt.Errorf("host not found in known hosts: %s", hostname)
	}

	return m.writeKnownHostsFile(filtered)
}

// formatFingerprint formats a public key fingerprint as SHA256 hash
func formatFingerprint(key ssh.PublicKey) string {
	return formatFingerprintFromBytes(key.Marshal())
}

// formatFingerprintFromBytes formats a fingerprint from key bytes
func formatFingerprintFromBytes(keyBytes []byte) string {
	hash := sha256.Sum256(keyBytes)
	encoded := base64.RawStdEncoding.EncodeToString(hash[:])
	return "SHA256:" + encoded
}

// GetKnownHostsFile returns the path to the known hosts file
func (m *KnownHostsManager) GetKnownHostsFile() string {
	return m.knownHostsFile
}

// ListKnownHosts returns a list of all known hosts
func (m *KnownHostsManager) ListKnownHosts() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lines, err := m.readKnownHostsFile()
	if err != nil {
		return nil, err
	}

	hosts := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) > 0 {
			hosts = append(hosts, parts[0])
		}
	}

	return hosts, nil
}
