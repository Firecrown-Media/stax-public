package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	mediaProxyURL      string
	mediaProxyCDN      string
	mediaProxyCache    bool
	mediaProxyCacheTTL string
)

// mediaCmd represents the media command group
var mediaCmd = &cobra.Command{
	Use:   "media",
	Short: "Manage media proxy configuration",
	Long: `Manage media proxy configuration for loading remote media files.

This allows you to work with production media files without downloading them
locally. The media proxy can use BunnyCDN or WPEngine as the source.`,
	Example: `  # Setup media proxy from WPEngine
  stax media setup-proxy

  # Setup media proxy from BunnyCDN
  stax media setup-proxy --cdn=https://mysite.b-cdn.net

  # Check media proxy status
  stax media status

  # Test media proxy configuration
  stax media test`,
}

var mediaSetupCmd = &cobra.Command{
	Use:   "setup-proxy",
	Short: "Configure DDEV nginx for media proxying",
	Long: `Configure DDEV nginx to proxy media files from a remote source.

This creates an nginx configuration that:
  - Tries to serve media files locally first
  - Falls back to remote CDN if file doesn't exist locally
  - Falls back to WPEngine if CDN fails
  - Optionally caches remote media files

The configuration is stored in .ddev/nginx_full/media-proxy.conf`,
	Example: `  # Setup with automatic WPEngine detection
  stax media setup-proxy

  # Setup with custom CDN URL
  stax media setup-proxy --cdn=https://mysite.b-cdn.net

  # Setup without caching
  stax media setup-proxy --no-cache

  # Setup with custom cache TTL
  stax media setup-proxy --cache-ttl=7d`,
	RunE: runMediaSetup,
}

var mediaStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show media proxy status",
	Long: `Show the current status of the media proxy configuration.

This checks:
  - Whether media proxy is configured in .stax.yml
  - Whether nginx configuration exists
  - Whether DDEV is running
  - Current proxy source (CDN/WPEngine)
  - Cache settings`,
	RunE: runMediaStatus,
}

var mediaTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test media proxy is working",
	Long: `Test that the media proxy configuration is working correctly.

This validates:
  - Nginx configuration syntax
  - DDEV is running
  - Proxy sources are reachable
  - Cache directory exists (if enabled)`,
	RunE: runMediaTest,
}

func init() {
	rootCmd.AddCommand(mediaCmd)

	// Add subcommands
	mediaCmd.AddCommand(mediaSetupCmd)
	mediaCmd.AddCommand(mediaStatusCmd)
	mediaCmd.AddCommand(mediaTestCmd)

	// Flags for setup-proxy
	mediaSetupCmd.Flags().StringVar(&mediaProxyCDN, "cdn", "", "CDN URL (e.g., https://mysite.b-cdn.net)")
	mediaSetupCmd.Flags().StringVar(&mediaProxyURL, "url", "", "WPEngine URL (auto-detected if not provided)")
	mediaSetupCmd.Flags().BoolVar(&mediaProxyCache, "cache", true, "enable local caching of proxied media")
	mediaSetupCmd.Flags().StringVar(&mediaProxyCacheTTL, "cache-ttl", "30d", "cache TTL (e.g., 7d, 24h)")
}

func runMediaSetup(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Setting Up Media Proxy")

	projectDir := getProjectDir()

	// Check if DDEV is configured
	if !ddev.IsConfigured(projectDir) {
		return errors.NewWithSolution(
			"DDEV is not configured",
			"Media proxy requires DDEV configuration",
			errors.Solution{
				Description: "Initialize DDEV first",
				Steps: []string{
					"Run: stax init",
					"Or manually configure DDEV: ddev config",
				},
			},
		)
	}

	// Load config if available
	var cdnURL, wpengineURL, wpengineHost string

	if cfg != nil {
		// Get BunnyCDN URL from config
		if cfg.Media.BunnyCDN.Hostname != "" && mediaProxyCDN == "" {
			cdnURL = fmt.Sprintf("https://%s", cfg.Media.BunnyCDN.Hostname)
			ui.Info(fmt.Sprintf("Using BunnyCDN from config: %s", cdnURL))
		}

		// Get WPEngine URL from config
		if wpeInstall(cfg) != "" && mediaProxyURL == "" {
			wpengineURL = fmt.Sprintf("https://%s.wpengine.com", wpeInstall(cfg))
			wpengineHost = fmt.Sprintf("%s.wpengine.com", wpeInstall(cfg))
			ui.Info(fmt.Sprintf("Using WPEngine from config: %s", wpengineURL))
		}
	}

	// Override with command-line flags
	if mediaProxyCDN != "" {
		cdnURL = mediaProxyCDN
	}
	if mediaProxyURL != "" {
		wpengineURL = mediaProxyURL
	}

	// Validate we have at least WPEngine URL
	if wpengineURL == "" {
		return errors.NewWithSolution(
			"No proxy source configured",
			"Either .stax.yml must have WPEngine settings or --url must be provided",
			errors.Solution{
				Description: "Provide a proxy source",
				Steps: []string{
					"Option 1: Configure .stax.yml with WPEngine install name",
					"Option 2: Run with --url flag: stax media setup-proxy --url=https://mysite.wpengine.com",
				},
			},
		)
	}

	// Build media proxy options
	options := ddev.MediaProxyOptions{
		Enabled:      true,
		CDNName:      "BunnyCDN",
		CDNURL:       cdnURL,
		WPEngineURL:  wpengineURL,
		WPEngineHost: wpengineHost,
		CacheTTL:     mediaProxyCacheTTL,
		CacheMaxSize: "10g",
		CacheEnabled: mediaProxyCache,
		ProxyHeaders: map[string]string{
			"X-Real-IP":       "$remote_addr",
			"X-Forwarded-For": "$proxy_add_x_forwarded_for",
		},
	}

	// If no CDN, use WPEngine as primary
	if cdnURL == "" {
		ui.Info("No CDN configured, using WPEngine as primary source")
		options.CDNURL = wpengineURL
		options.CDNName = "WPEngine"
	}

	// Generate nginx configuration
	spinner := ui.NewSpinner("Generating nginx media proxy configuration")
	spinner.Start()

	if err := ddev.GenerateMediaProxyConfig(projectDir, options); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to generate media proxy config: %w", err)
	}

	spinner.Success("Media proxy configuration generated")

	// Validate nginx configuration if DDEV is running
	status, err := ddev.GetStatus(projectDir)
	if err == nil && status.Running {
		spinner = ui.NewSpinner("Validating nginx configuration")
		spinner.Start()

		if err := ddev.ValidateNginxConfig(projectDir); err != nil {
			spinner.Stop()

			// Rollback: Remove invalid configuration
			ui.Warning("Nginx configuration validation failed - rolling back changes")
			if removeErr := ddev.RemoveMediaProxyConfig(projectDir); removeErr != nil {
				ui.Error(fmt.Sprintf("Failed to rollback invalid config: %v", removeErr))
			}

			return errors.NewWithSolution(
				"Nginx configuration validation failed",
				fmt.Sprintf("Generated configuration has syntax errors: %v", err),
				errors.Solution{
					Description: "The media proxy configuration could not be applied",
					Steps: []string{
						"Configuration has been rolled back automatically",
						"This may be a bug - please report to: https://github.com/Firecrown-Media/stax/issues",
						fmt.Sprintf("Include this error in your report: %v", err),
					},
				},
			)
		}
		spinner.Success("Nginx configuration is valid")

		// Restart DDEV to apply changes
		ui.Info("Restarting DDEV to apply media proxy configuration...")
		spinner = ui.NewSpinner("Restarting DDEV")
		spinner.Start()

		if err := ddev.Restart(projectDir); err != nil {
			spinner.Stop()
			ui.Warning("Failed to restart DDEV automatically")
			ui.Info("Run 'stax restart' manually to apply changes")
		} else {
			spinner.Success("DDEV restarted")
		}
	} else {
		ui.Info("DDEV is not running. Changes will take effect on next start.")
	}

	fmt.Println()
	ui.Success("Media proxy configured successfully!")
	fmt.Println()

	// Show configuration summary
	ui.Section("Configuration Summary")
	if cdnURL != "" && cdnURL != wpengineURL {
		fmt.Printf("  Primary Source:  %s\n", cdnURL)
		fmt.Printf("  Fallback Source: %s\n", wpengineURL)
	} else {
		fmt.Printf("  Source:          %s\n", wpengineURL)
	}
	fmt.Printf("  Caching:         %s\n", getBoolStatus(mediaProxyCache))
	if mediaProxyCache {
		fmt.Printf("  Cache TTL:       %s\n", mediaProxyCacheTTL)
	}
	fmt.Printf("  Config File:     .ddev/nginx_full/media-proxy.conf\n")
	fmt.Println()

	ui.Info("Next steps:")
	ui.Info("  1. Test the proxy: stax media test")
	ui.Info("  2. Check status: stax media status")
	ui.Info("  3. Visit your site and verify media loads correctly")

	return nil
}

func runMediaStatus(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Media Proxy Status")

	projectDir := getProjectDir()

	// Check if DDEV is configured
	if !ddev.IsConfigured(projectDir) {
		ui.Warning("DDEV is not configured")
		return nil
	}

	fmt.Println()

	// Check .stax.yml configuration
	ui.Section("Configuration")
	if cfg != nil && cfg.Media.ProxyEnabled {
		fmt.Println("  Proxy Enabled:   ✓ Yes")

		if cfg.Media.BunnyCDN.Hostname != "" {
			fmt.Printf("  CDN Hostname:    %s\n", cfg.Media.BunnyCDN.Hostname)
		}

		if wpeInstall(cfg) != "" {
			fmt.Printf("  WPEngine:        %s.wpengine.com\n", wpeInstall(cfg))
			fmt.Printf("  WP Fallback:     %s\n", getBoolStatus(cfg.Media.WPEngineFallback))
		}

		fmt.Printf("  Cache Enabled:   %s\n", getBoolStatus(cfg.Media.Cache.Enabled))
		if cfg.Media.Cache.Enabled {
			fmt.Printf("  Cache Directory: %s\n", cfg.Media.Cache.Directory)
			fmt.Printf("  Cache Max Size:  %s\n", cfg.Media.Cache.MaxSize)
		}
	} else {
		fmt.Println("  Proxy Enabled:   ✗ No (not in .stax.yml)")
	}
	fmt.Println()

	// Check nginx configuration
	ui.Section("Nginx Configuration")
	nginxConfigPath := filepath.Join(projectDir, ".ddev", "nginx_full", "media-proxy.conf")
	if _, err := os.Stat(nginxConfigPath); err == nil {
		fmt.Println("  Config File:     ✓ Exists")
		fmt.Printf("  Location:        %s\n", nginxConfigPath)

		// Try to validate if DDEV is running
		status, err := ddev.GetStatus(projectDir)
		if err == nil && status.Running {
			if err := ddev.ValidateNginxConfig(projectDir); err != nil {
				fmt.Println("  Validation:      ✗ Invalid")
				ui.Warning("  Nginx configuration has syntax errors")
			} else {
				fmt.Println("  Validation:      ✓ Valid")
			}
		}
	} else {
		fmt.Println("  Config File:     ✗ Not found")
		ui.Info("  Run 'stax media setup-proxy' to create configuration")
	}
	fmt.Println()

	// Check DDEV status
	ui.Section("DDEV Status")
	status, err := ddev.GetStatus(projectDir)
	if err != nil {
		fmt.Println("  Status:          ✗ Not available")
	} else {
		if status.Running {
			fmt.Println("  Status:          ✓ Running")
			fmt.Printf("  Primary URL:     %s\n", status.PrimaryURL)
		} else {
			fmt.Println("  Status:          ⚫ Stopped")
			ui.Info("  Start DDEV to use media proxy: stax start")
		}
	}
	fmt.Println()

	// Check cache directory (if caching is enabled)
	if cfg != nil && cfg.Media.Cache.Enabled {
		ui.Section("Cache Status")
		cacheDir := filepath.Join(projectDir, cfg.Media.Cache.Directory)
		if _, err := os.Stat(cacheDir); err == nil {
			fmt.Println("  Cache Directory: ✓ Exists")
			fmt.Printf("  Location:        %s\n", cacheDir)

			// Try to get directory size
			var size int64
			_ = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					size += info.Size()
				}
				return nil
			})
			if size > 0 {
				fmt.Printf("  Cache Size:      %s\n", formatBytes(size))
			}
		} else {
			fmt.Println("  Cache Directory: ⚠ Not found")
			ui.Info("  Cache directory will be created automatically when media is accessed")
		}
		fmt.Println()
	}

	return nil
}

func runMediaTest(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Testing Media Proxy")

	projectDir := getProjectDir()

	// Check if DDEV is configured
	if !ddev.IsConfigured(projectDir) {
		return errors.NewWithSolution(
			"DDEV is not configured",
			"Cannot test media proxy without DDEV",
			errors.Solution{
				Description: "Initialize DDEV first",
				Steps: []string{
					"Run: stax init",
				},
			},
		)
	}

	fmt.Println()

	// Test 1: Check nginx config exists
	ui.Section("Configuration Tests")
	nginxConfigPath := filepath.Join(projectDir, ".ddev", "nginx_full", "media-proxy.conf")
	if _, err := os.Stat(nginxConfigPath); err == nil {
		ui.Success("✓ Nginx configuration file exists")
	} else {
		ui.Error("✗ Nginx configuration file not found")
		return errors.NewWithSolution(
			"Media proxy not configured",
			"Nginx configuration file does not exist",
			errors.Solution{
				Description: "Setup media proxy first",
				Steps: []string{
					"Run: stax media setup-proxy",
				},
			},
		)
	}

	// Test 2: Check DDEV is running
	ui.Section("Environment Tests")
	status, err := ddev.GetStatus(projectDir)
	if err != nil {
		ui.Error("✗ Cannot get DDEV status")
		return err
	}

	if !status.Running {
		return errors.NewWithSolution(
			"DDEV is not running",
			"Media proxy requires DDEV to be running",
			errors.Solution{
				Description: "Start DDEV",
				Steps: []string{
					"Run: stax start",
				},
			},
		)
	}
	ui.Success("✓ DDEV is running")

	// Test 3: Validate nginx configuration
	ui.Section("Nginx Validation")
	if err := ddev.ValidateNginxConfig(projectDir); err != nil {
		ui.Error("✗ Nginx configuration is invalid")
		ui.Warning(fmt.Sprintf("Error: %v", err))
		return errors.NewWithSolution(
			"Nginx configuration is invalid",
			err.Error(),
			errors.Solution{
				Description: "Fix the configuration",
				Steps: []string{
					"1. Check .ddev/nginx_full/media-proxy.conf for syntax errors",
					"2. Regenerate config: stax media setup-proxy",
					"3. Restart DDEV: stax restart",
				},
			},
		)
	}
	ui.Success("✓ Nginx configuration is valid")

	// Test 4: Check proxy sources
	if cfg != nil {
		ui.Section("Proxy Source Tests")

		if cfg.Media.BunnyCDN.Hostname != "" {
			cdnURL := fmt.Sprintf("https://%s", cfg.Media.BunnyCDN.Hostname)
			ui.Info(fmt.Sprintf("  CDN URL: %s", cdnURL))
			ui.Success("✓ BunnyCDN configured")
		}

		if wpeInstall(cfg) != "" {
			wpengineURL := fmt.Sprintf("https://%s.wpengine.com", wpeInstall(cfg))
			ui.Info(fmt.Sprintf("  WPEngine URL: %s", wpengineURL))
			ui.Success("✓ WPEngine configured")
		}
	}

	fmt.Println()
	ui.Success("All media proxy tests passed!")
	fmt.Println()

	ui.Info("Manual verification steps:")
	ui.Info(fmt.Sprintf("  1. Visit: %s", status.PrimaryURL))
	ui.Info("  2. Navigate to a page with media/images")
	ui.Info("  3. Check browser DevTools Network tab")
	ui.Info("  4. Verify images load from remote source")
	ui.Info("  5. Look for X-Proxy-Source header in response")

	return nil
}

// getBoolStatus returns a formatted string for boolean status
func getBoolStatus(enabled bool) string {
	if enabled {
		return "✓ Enabled"
	}
	return "✗ Disabled"
}

// formatBytes formats bytes to human-readable size
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
