package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneOptions configures repository cloning
type CloneOptions struct {
	URL         string
	Destination string
	Branch      string
	Depth       int
	Quiet       bool
	Progress    func(message string)
}

// Clone clones a Git repository
func Clone(opts CloneOptions) error {
	if opts.URL == "" {
		return fmt.Errorf("repository URL is required")
	}

	if opts.Destination == "" {
		return fmt.Errorf("destination path is required")
	}

	// Check if git is available
	if !IsGitAvailable() {
		return fmt.Errorf("git is not installed or not in PATH")
	}

	// Check if destination already exists
	if exists, _ := pathExists(opts.Destination); exists {
		return fmt.Errorf("destination path already exists: %s", opts.Destination)
	}

	// Build git clone command
	args := []string{"clone"}

	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
	}

	if opts.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", opts.Depth))
	}

	if opts.Quiet {
		args = append(args, "--quiet")
	} else {
		args = append(args, "--progress")
	}

	args = append(args, opts.URL, opts.Destination)

	// Execute clone
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if opts.Progress != nil {
		opts.Progress(fmt.Sprintf("Cloning repository from %s", opts.URL))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	if opts.Progress != nil {
		opts.Progress(fmt.Sprintf("Repository cloned to %s", opts.Destination))
	}

	return nil
}

// IsGitAvailable checks if git is installed and available
func IsGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}

// GetGitVersion returns the installed git version
func GetGitVersion() (string, error) {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "git version ")

	return version, nil
}

// IsGitRepository checks if a directory is a git repository
func IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	exists, _ := pathExists(gitDir)
	return exists
}

// GetRepositoryURL returns the remote origin URL of a repository
func GetRepositoryURL(repoPath string) (string, error) {
	if !IsGitRepository(repoPath) {
		return "", fmt.Errorf("not a git repository: %s", repoPath)
	}

	cmd := exec.Command("git", "-C", repoPath, "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(repoPath string) (string, error) {
	if !IsGitRepository(repoPath) {
		return "", fmt.Errorf("not a git repository: %s", repoPath)
	}

	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Pull pulls the latest changes from remote
func Pull(repoPath string, quiet bool) error {
	if !IsGitRepository(repoPath) {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	args := []string{"-C", repoPath, "pull"}
	if quiet {
		args = append(args, "--quiet")
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	return nil
}

// Checkout checks out a specific branch
func Checkout(repoPath, branch string) error {
	if !IsGitRepository(repoPath) {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	cmd := exec.Command("git", "-C", repoPath, "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	return nil
}

// GetRemoteURL extracts the repository name from a Git URL
func GetRemoteURL(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL is empty")
	}

	// Handle SSH URLs (git@github.com:user/repo.git)
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid SSH URL format")
		}
		return strings.TrimSuffix(parts[1], ".git"), nil
	}

	// Handle HTTPS URLs (https://github.com/user/repo.git)
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")
		parts := strings.Split(url, "/")
		if len(parts) < 3 {
			return "", fmt.Errorf("invalid HTTPS URL format")
		}
		return strings.TrimSuffix(strings.Join(parts[1:], "/"), ".git"), nil
	}

	return "", fmt.Errorf("unsupported URL format")
}

// ValidateURL validates a Git repository URL
func ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL is empty")
	}

	if !strings.HasPrefix(url, "git@") && !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return fmt.Errorf("URL must start with git@, https://, or http://")
	}

	return nil
}

// pathExists checks if a path exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetBranchProtectionGuideURL returns the URL to the branch protection guide
func GetBranchProtectionGuideURL() string {
	return "https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/managing-a-branch-protection-rule"
}
