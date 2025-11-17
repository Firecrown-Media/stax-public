package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/providers/wpengine"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	listOutputFormat string
	listProvider     string
	listFilter       string
	listEnvironment  string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "✓ List available WPEngine installs",
	Long: `List all WPEngine installations available to your account.

This command works globally without requiring a stax.yml file.
Use it to discover install names before creating your project configuration.`,
	Example: `  # List all installs
  stax list

  # List with JSON output
  stax list --output=json

  # Filter by name
  stax list --filter="client.*"

  # Filter by environment
  stax list --environment=production

  # Combined filters
  stax list --filter="fs.*" --environment=staging --output=yaml`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listOutputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	listCmd.Flags().StringVarP(&listProvider, "provider", "p", "wpengine", "Provider to list from")
	listCmd.Flags().StringVarP(&listFilter, "filter", "f", "", "Filter by install name (regex)")
	listCmd.Flags().StringVarP(&listEnvironment, "environment", "e", "", "Filter by environment")
}

func runList(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Listing WPEngine Installs")

	// 1. Load credentials
	creds, err := loadGlobalCredentials()
	if err != nil {
		return err
	}

	// 2. Create provider
	p, err := createWPEngineProvider(creds)
	if err != nil {
		return err
	}
	defer p.Close()

	// 3. List and filter sites
	sites, err := listAndFilterSites(p, listFilter, listEnvironment)
	if err != nil {
		return err
	}

	// 4. Output based on format
	switch listOutputFormat {
	case "json":
		return outputSitesJSON(sites)
	case "yaml":
		return outputSitesYAML(sites)
	default:
		return outputSitesTable(sites)
	}
}

// loadGlobalCredentials loads WPEngine credentials without requiring stax.yml
func loadGlobalCredentials() (*credentials.WPEngineCredentials, error) {
	// Try to load credentials using "global" as the install identifier
	// This works with all three storage methods:
	// 1. Environment variables (no install name needed)
	// 2. Credentials file (no install name needed)
	// 3. Keychain (uses "global" as account name)

	creds, err := credentials.GetWPEngineCredentialsWithFallback("global")
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	// Validate required fields
	if creds.APIUser == "" || creds.APIPassword == "" {
		return nil, fmt.Errorf("incomplete credentials: api_user and api_password are required")
	}

	return creds, nil
}

// createWPEngineProvider creates a WPEngine provider instance without full config
func createWPEngineProvider(creds *credentials.WPEngineCredentials) (*wpengine.WPEngineProvider, error) {
	// Create provider instance
	p := &wpengine.WPEngineProvider{}

	// Authenticate with credentials
	// For listing, we need to provide install but it won't be used for the API call
	credMap := map[string]string{
		"api_user":     creds.APIUser,
		"api_password": creds.APIPassword,
		"install":      "temp", // Required by ValidateCredentials but not used for ListSites
		"ssh_key":      "",     // Empty to skip SSH client creation
	}

	if creds.SSHGateway != "" {
		credMap["ssh_gateway"] = creds.SSHGateway
	} else {
		credMap["ssh_gateway"] = "ssh.wpengine.net"
	}

	if err := p.Authenticate(credMap); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Test connection
	if err := p.TestConnection(); err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	return p, nil
}

// listAndFilterSites retrieves and filters sites
func listAndFilterSites(p provider.Provider, filterName, filterEnv string) ([]provider.Site, error) {
	// Get all sites
	sites, err := p.ListSites()
	if err != nil {
		return nil, fmt.Errorf("failed to list sites: %w", err)
	}

	// Apply filters
	var filtered []provider.Site

	// Compile regex filter if provided
	var nameRegex *regexp.Regexp
	if filterName != "" {
		nameRegex, err = regexp.Compile(filterName)
		if err != nil {
			return nil, fmt.Errorf("invalid filter regex: %w", err)
		}
	}

	for _, site := range sites {
		// Filter by name
		if nameRegex != nil && !nameRegex.MatchString(site.Name) {
			continue
		}

		// Filter by environment
		if filterEnv != "" && site.Environment != filterEnv {
			continue
		}

		filtered = append(filtered, site)
	}

	return filtered, nil
}

// outputSitesTable displays sites in table format
func outputSitesTable(sites []provider.Site) error {
	if len(sites) == 0 {
		ui.Warning("No installs found matching your criteria")
		fmt.Println()
		fmt.Println("Tips:")
		fmt.Println("  - Check your filter/environment flags")
		fmt.Println("  - Verify your WPEngine account has install access")
		fmt.Println("  - Run without filters: stax list")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "INSTALL NAME\tENVIRONMENT\tPRIMARY DOMAIN\tPHP\tSTATUS")
	fmt.Fprintln(w, "------------\t-----------\t--------------\t---\t------")

	// Rows
	for _, site := range sites {
		phpVersion := site.Metadata["php_version"]
		if phpVersion == "" {
			phpVersion = "unknown"
		}

		status := site.Status
		if status == "" {
			status = "active"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			site.Name,
			site.Environment,
			site.PrimaryDomain,
			phpVersion,
			status,
		)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "Total: %d installs\n", len(sites))

	return nil
}

// outputSitesJSON displays sites in JSON format
func outputSitesJSON(sites []provider.Site) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sites)
}

// outputSitesYAML displays sites in YAML format
func outputSitesYAML(sites []provider.Site) error {
	data, err := yaml.Marshal(sites)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
