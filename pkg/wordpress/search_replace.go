package wordpress

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SearchReplaceOptions represents options for search-replace operations
type SearchReplaceOptions struct {
	Network     bool
	SkipColumns []string
	SkipTables  []string
	DryRun      bool
	URL         string
}

// SearchReplacePair represents a search-replace operation
type SearchReplacePair struct {
	Old string
	New string
	URL string
}

// MultisiteSearchReplaceConfig represents configuration for multisite search-replace
type MultisiteSearchReplaceConfig struct {
	Network NetworkReplace
	Sites   []SiteReplace
}

// NetworkReplace represents network-wide search-replace
type NetworkReplace struct {
	Old string
	New string
}

// SiteReplace represents site-specific search-replace
type SiteReplace struct {
	Old    string
	New    string
	URL    string
	BlogID int
}

// SearchReplaceResult contains the results of a search-replace operation
type SearchReplaceResult struct {
	Replacements int
	Output       string
	Success      bool
}

// SearchReplace performs a single search-replace operation
func (c *CLI) SearchReplaceWithOptions(old, new string, options SearchReplaceOptions) error {
	args := []string{"search-replace", old, new}

	// Always use --all-tables to ensure comprehensive replacement
	args = append(args, "--all-tables")

	// Skip GUID column by default
	skipColumns := options.SkipColumns
	if len(skipColumns) == 0 {
		skipColumns = []string{"guid"}
	}
	args = append(args, "--skip-columns="+joinStrings(skipColumns, ","))

	// Add flags to avoid DNS resolution issues
	// These flags prevent WP-CLI from loading themes/plugins which may trigger DNS lookups
	args = append(args, "--skip-themes", "--skip-plugins")

	// Network-wide
	if options.Network {
		args = append(args, "--network")
	}

	// Specific site
	if options.URL != "" {
		args = append(args, "--url="+options.URL)
	}

	// Dry run
	if options.DryRun {
		args = append(args, "--dry-run")
	}

	// Execute and capture output
	output, err := c.ExecuteWithOutput(args...)
	if err != nil {
		return fmt.Errorf("search-replace command failed: %w", err)
	}

	// Parse the output to verify replacements were made
	result := parseSearchReplaceOutput(output)

	if !result.Success {
		return fmt.Errorf("search-replace completed but reported failure: %s", output)
	}

	if result.Replacements == 0 {
		return fmt.Errorf("search-replace made 0 replacements - URLs may not have been updated (output: %s)", output)
	}

	return nil
}

// parseSearchReplaceOutput parses the output from wp search-replace and extracts statistics
// Example output: "Success: Made 42 replacements."
// Or for errors: "Error: ..." or "Warning: ..."
func parseSearchReplaceOutput(output string) SearchReplaceResult {
	result := SearchReplaceResult{
		Output:  output,
		Success: false,
	}

	// Check for success message
	if strings.Contains(output, "Success:") {
		result.Success = true
	}

	// Check for error indicators
	if strings.Contains(output, "Error:") || strings.Contains(output, "Fatal error") {
		result.Success = false
		return result
	}

	// Parse replacement count using regex
	// Match patterns like "Made 42 replacements" or "Made 1 replacement"
	re := regexp.MustCompile(`Made (\d+) replacement`)
	matches := re.FindStringSubmatch(output)

	if len(matches) >= 2 {
		count, err := strconv.Atoi(matches[1])
		if err == nil {
			result.Replacements = count
		}
	}

	return result
}

// MultisiteSearchReplace performs search-replace for all sites in a multisite network
func (c *CLI) MultisiteSearchReplace(config MultisiteSearchReplaceConfig) error {
	// Network-wide replacement
	if config.Network.Old != "" && config.Network.New != "" {
		opts := SearchReplaceOptions{
			Network:     true,
			SkipColumns: []string{"guid"},
		}
		if err := c.SearchReplaceWithOptions(config.Network.Old, config.Network.New, opts); err != nil {
			return fmt.Errorf("network search-replace failed: %w", err)
		}
	}

	// Site-specific replacements
	for _, site := range config.Sites {
		opts := SearchReplaceOptions{
			Network:     false,
			SkipColumns: []string{"guid"},
			URL:         site.URL,
		}
		if err := c.SearchReplaceWithOptions(site.Old, site.New, opts); err != nil {
			return fmt.Errorf("search-replace failed for %s: %w", site.URL, err)
		}
	}

	return nil
}

// PreviewSearchReplace shows what would be replaced without making changes
func (c *CLI) PreviewSearchReplace(old, new string) (string, error) {
	args := []string{"search-replace", old, new, "--dry-run"}
	return c.ExecuteWithOutput(args...)
}

// GetSubsites gets all subsites from the database
func (c *CLI) GetSubsites() ([]Site, error) {
	return c.GetSites()
}

// BuildFirecrownSearchReplaceConfig builds the standard Firecrown multisite search-replace configuration
func BuildFirecrownSearchReplaceConfig(localDomain string) MultisiteSearchReplaceConfig {
	return MultisiteSearchReplaceConfig{
		Network: NetworkReplace{
			Old: "fsmultisite.wpenginepowered.com",
			New: localDomain,
		},
		Sites: []SiteReplace{
			{
				Old: "flyingmag.com",
				New: "flyingmag." + localDomain,
				URL: "flyingmag.com",
			},
			{
				Old: "planeandpilotmag.com",
				New: "planeandpilot." + localDomain,
				URL: "planeandpilotmag.com",
			},
			{
				Old: "finescale.com",
				New: "finescale." + localDomain,
				URL: "finescale.com",
			},
			{
				Old: "avweb.com",
				New: "avweb." + localDomain,
				URL: "avweb.com",
			},
		},
	}
}

// Helper function to join strings
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}
	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}
