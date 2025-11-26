package wordpress

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// CLI wraps WP-CLI operations
type CLI struct {
	ProjectDir string
	useDDEV    bool
}

// NewCLI creates a new WordPress CLI wrapper
func NewCLI(projectDir string) *CLI {
	return &CLI{
		ProjectDir: projectDir,
		useDDEV:    true,
	}
}

// IsInstalled checks if WP-CLI is available
func IsInstalled() bool {
	_, err := exec.LookPath("wp")
	return err == nil
}

// Exec executes a WP-CLI command
func (c *CLI) Exec(args ...string) (string, error) {
	return c.ExecuteWithOutput(args...)
}

// Execute executes a WP-CLI command
func (c *CLI) Execute(args ...string) error {
	_, err := c.ExecuteWithOutput(args...)
	return err
}

// ExecuteWithOutput executes a WP-CLI command and returns the output
func (c *CLI) ExecuteWithOutput(args ...string) (string, error) {
	var cmd *exec.Cmd

	if c.useDDEV {
		// Prepend 'wp' to args
		ddevArgs := append([]string{"wp"}, args...)
		cmd = exec.Command("ddev", ddevArgs...)
	} else {
		cmd = exec.Command("wp", args...)
	}

	cmd.Dir = c.ProjectDir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("wp-cli command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), nil
}

// SearchReplace runs search-replace on the database
func (c *CLI) SearchReplace(old, new string, network bool) error {
	args := []string{"search-replace", old, new, "--skip-columns=guid"}
	if network {
		args = append(args, "--network")
	}
	return c.Execute(args...)
}

// GetSites returns a list of sites in the multisite network
func (c *CLI) GetSites() ([]Site, error) {
	output, err := c.ExecuteWithOutput("site", "list", "--field=url")
	if err != nil {
		return nil, err
	}

	var sites []Site
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, url := range lines {
		if url != "" {
			sites = append(sites, Site{
				ID:  i + 1,
				URL: strings.TrimSpace(url),
			})
		}
	}

	return sites, nil
}

// GetSiteURL gets the site URL for a specific blog ID
func (c *CLI) GetSiteURL(blogID int) (string, error) {
	args := []string{"option", "get", "siteurl"}
	if blogID > 1 {
		args = append(args, fmt.Sprintf("--url=%d", blogID))
	}

	output, err := c.ExecuteWithOutput(args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// GetOption gets a WordPress option value
func (c *CLI) GetOption(option string, blogID int) (string, error) {
	args := []string{"option", "get", option}
	if blogID > 0 {
		args = append(args, fmt.Sprintf("--blog=%d", blogID))
	}

	output, err := c.ExecuteWithOutput(args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// SetOption sets a WordPress option value
func (c *CLI) SetOption(option, value string, blogID int) error {
	args := []string{"option", "set", option, value}
	if blogID > 0 {
		args = append(args, fmt.Sprintf("--blog=%d", blogID))
	}

	return c.Execute(args...)
}

// FlushCache flushes the WordPress object cache
func (c *CLI) FlushCache() error {
	return c.Execute("cache", "flush")
}

// GetTablePrefix gets the WordPress table prefix
func (c *CLI) GetTablePrefix() (string, error) {
	output, err := c.ExecuteWithOutput("config", "get", "table_prefix")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// CoreVersion gets the WordPress core version
func (c *CLI) CoreVersion() (string, error) {
	output, err := c.ExecuteWithOutput("core", "version")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// ImportDatabase imports a SQL file into the database
func (c *CLI) ImportDatabase(sqlFile string) error {
	var cmd *exec.Cmd

	if c.useDDEV {
		cmd = exec.Command("ddev", "import-db", "--src="+sqlFile)
	} else {
		cmd = exec.Command("wp", "db", "import", sqlFile)
	}

	cmd.Dir = c.ProjectDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("database import failed: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}

// ExportDatabase exports the database to a SQL file
func (c *CLI) ExportDatabase(destination string) error {
	args := []string{"db", "export", destination}
	return c.Execute(args...)
}

// Query executes a SQL query
func (c *CLI) Query(query string) (string, error) {
	return c.ExecuteWithOutput("db", "query", query)
}

// CLISite represents a WordPress site from CLI output
type CLISite struct {
	ID     int
	URL    string
	Domain string
	Path   string
}

// VerifySiteURL verifies that the siteurl option matches the expected URL
// Returns true if the URL matches, false if it doesn't, and the actual URL found
func (c *CLI) VerifySiteURL(expectedURL string) (bool, string, error) {
	actualURL, err := c.ExecuteWithOutput("option", "get", "siteurl")
	if err != nil {
		return false, "", fmt.Errorf("failed to get siteurl: %w", err)
	}

	actualURL = strings.TrimSpace(actualURL)
	expectedURL = strings.TrimSpace(expectedURL)

	// Normalize URLs for comparison (remove trailing slashes)
	actualURL = strings.TrimRight(actualURL, "/")
	expectedURL = strings.TrimRight(expectedURL, "/")

	matches := actualURL == expectedURL
	return matches, actualURL, nil
}
