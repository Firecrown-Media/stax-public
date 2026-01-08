package cmd

import (
	"fmt"
	"strings"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/firecrown-media/stax/pkg/wordpress"
)

// getWPEngineURL returns the WPEngine URL for the current environment
func getWPEngineURL(cfg *config.Config) string {
	// Determine the environment URL pattern based on the environment
	environment := cfg.WPEngine.Environment
	install := cfg.WPEngine.Install

	// Check if custom domain is configured in config
	if environment == "production" {
		// For production, check if a custom primary domain is configured
		if cfg.WPEngine.Domains.Production.Primary != "" {
			return "https://" + cfg.WPEngine.Domains.Production.Primary
		}
		// Default production URL pattern
		return fmt.Sprintf("https://%s.wpengine.com", install)
	} else if environment == "staging" {
		// For staging, check if a custom primary domain is configured
		if cfg.WPEngine.Domains.Staging.Primary != "" {
			return "https://" + cfg.WPEngine.Domains.Staging.Primary
		}
		// WPEngine staging installs often have environment suffix in the name
		// and use .wpengine.com (not .wpengineurl.com)
		// Examples: mysitestage, mysite-staging, mysitestg
		if strings.HasSuffix(install, "stg") ||
			strings.HasSuffix(install, "stage") ||
			strings.HasSuffix(install, "staging") ||
			strings.Contains(install, "stag") {
			return fmt.Sprintf("https://%s.wpengine.com", install)
		}
		// Default staging URL pattern (older wpengineurl.com pattern)
		return fmt.Sprintf("https://%s.wpengine.com", install)
	} else if environment == "development" {
		// Development environment pattern
		return fmt.Sprintf("https://%s.wpengine.com", install)
	}

	// Fallback to .wpengine.com pattern
	return fmt.Sprintf("https://%s.wpengine.com", install)
}

// getPossibleWPEngineURLs returns possible URL patterns for this install
// This is used to try multiple URL patterns when search-replace doesn't find matches
func getPossibleWPEngineURLs(cfg *config.Config) []string {
	install := cfg.WPEngine.Install
	urls := []string{}

	// Custom domain first (most specific)
	if cfg.WPEngine.Domains.Production.Primary != "" {
		urls = append(urls, "https://"+cfg.WPEngine.Domains.Production.Primary)
	}
	if cfg.WPEngine.Domains.Staging.Primary != "" {
		urls = append(urls, "https://"+cfg.WPEngine.Domains.Staging.Primary)
	}

	// Common WPEngine patterns - try multiple domain patterns
	urls = append(urls,
		fmt.Sprintf("https://%s.wpengine.com", install),
		fmt.Sprintf("https://%s.wpengineurl.com", install),
		fmt.Sprintf("https://%s.wpenginepowered.com", install),
	)

	return urls
}

// getDDEVURL returns the local DDEV URL
func getDDEVURL(cfg *config.Config) string {
	// Local DDEV URLs always follow the pattern: {project-name}.ddev.site
	return fmt.Sprintf("https://%s.ddev.site", cfg.Project.Name)
}

// runSearchReplace executes wp search-replace via DDEV and verifies the results
func runSearchReplace(projectDir string, from, to string, cfg *config.Config) error {
	// Create WordPress CLI wrapper
	cli := wordpress.NewCLI(projectDir)

	ui.Info(fmt.Sprintf("Replacing URLs: %s -> %s", from, to))

	// Check if this is a multisite installation
	isMultisite := cfg.Project.Type == "wordpress-multisite"

	if isMultisite {
		// For multisite, use network-wide search-replace
		ui.Info("Detected multisite installation - running network-wide search-replace")

		opts := wordpress.SearchReplaceOptions{
			Network:     true,
			SkipColumns: []string{"guid"},
			DryRun:      false,
		}

		if err := cli.SearchReplaceWithOptions(from, to, opts); err != nil {
			return fmt.Errorf("multisite search-replace failed: %w", err)
		}

		// For multisite, we may also need to replace subdomain URLs
		if cfg.Project.Mode == "subdomain" && len(cfg.Network.Sites) > 0 {
			ui.Info("Running additional search-replace for subdomain sites")

			// For each site, perform URL replacement if configured
			for _, site := range cfg.Network.Sites {
				if !site.Active {
					continue
				}

				if site.WPEngineDomain != "" && site.Domain != "" {
					siteFrom := "https://" + site.WPEngineDomain
					siteTo := "https://" + site.Domain

					ui.Info(fmt.Sprintf("Site %s: %s -> %s", site.Name, siteFrom, siteTo))

					siteOpts := wordpress.SearchReplaceOptions{
						Network:     false,
						URL:         site.Domain,
						SkipColumns: []string{"guid"},
						DryRun:      false,
					}

					if err := cli.SearchReplaceWithOptions(siteFrom, siteTo, siteOpts); err != nil {
						ui.Warning(fmt.Sprintf("Search-replace failed for site %s: %v", site.Name, err))
						// Continue with other sites even if one fails
					}
				}
			}
		}
	} else {
		// For single-site, use standard search-replace
		ui.Info("Detected single-site installation - running standard search-replace")

		opts := wordpress.SearchReplaceOptions{
			Network:     false,
			SkipColumns: []string{"guid"},
			DryRun:      false,
		}

		if err := cli.SearchReplaceWithOptions(from, to, opts); err != nil {
			return fmt.Errorf("search-replace failed: %w", err)
		}
	}

	// CRITICAL: Verify that the siteurl was actually changed
	ui.Info("Verifying URL replacement...")
	matches, actualURL, err := cli.VerifySiteURL(to)
	if err != nil {
		return fmt.Errorf("failed to verify siteurl: %w", err)
	}

	if !matches {
		return fmt.Errorf("URL replacement verification failed: expected siteurl to be '%s' but found '%s'. The search-replace may have completed without errors but did not update the URLs correctly", to, actualURL)
	}

	ui.Success(fmt.Sprintf("Verified: siteurl is correctly set to %s", to))

	return nil
}
