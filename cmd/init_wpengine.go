package cmd

import (
	"fmt"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/prompts"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/providers/wpengine"
	"github.com/firecrown-media/stax/pkg/ui"
	wpeclient "github.com/firecrown-media/stax/pkg/wpengine"
)

// promptForWPEngineInstall fetches installs from WPEngine API and shows picker
func promptForWPEngineInstall() (*prompts.WPEngineInstallWithDetails, error) {
	// Get credentials from keychain, environment, or file
	creds, err := credentials.GetWPEngineCredentialsWithFallback("global")
	if err != nil {
		return nil, fmt.Errorf("no WPEngine credentials found: %w", err)
	}

	// Create WPEngine client
	client := wpeclient.NewClient(creds.APIUser, creds.APIPassword, "")

	// Fetch installs
	ui.Info("Fetching installs from WPEngine...")
	installs, err := client.ListInstalls()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch installs: %w", err)
	}

	if len(installs) == 0 {
		return nil, fmt.Errorf("no installs found in your WPEngine account")
	}

	// Transform to picker format (basic info only for fast display)
	installDetails := make([]prompts.WPEngineInstallWithDetails, len(installs))
	for i, install := range installs {
		installDetails[i] = prompts.WPEngineInstallWithDetails{
			Name:        install.Name,
			Environment: install.Environment,
			PHPVersion:  install.PHPVersion,
		}
	}

	// Show picker
	selected, err := prompts.WPEngineInstallPickerPrompt(installDetails)
	if err != nil {
		return nil, err
	}

	// Fetch full details for selected install to get WordPress version, MySQL version, etc.
	ui.Info("Fetching install details...")
	fullDetails, err := client.GetInstallByName(selected.Name)
	if err != nil {
		ui.Warning(fmt.Sprintf("Could not fetch full install details: %v", err))
		// Return what we have - WordPress version will remain empty
		return &selected, nil
	}

	// Populate additional fields from full details
	selected.MySQLVersion = fullDetails.MySQLVersion
	selected.WordPressVersion = fullDetails.WordPressVersion

	// Debug logging for API response
	ui.Debug(fmt.Sprintf("API returned - PHP: %s, MySQL: %s, WordPress: %s",
		fullDetails.PHPVersion, fullDetails.MySQLVersion, fullDetails.WordPressVersion))

	if fullDetails.WordPressVersion == "" {
		ui.Debug("WordPress version not available from WPEngine API - this is normal for some installs")
	}

	return &selected, nil
}

// gatherWPEngineConfiguration collects WPEngine configuration from flags or prompts
func gatherWPEngineConfiguration(cfg *config.Config) error {
	ui.Section("WPEngine Integration")

	setupWPEngine := false
	if initWPEngineInstall != "" {
		setupWPEngine = true
	} else if initInteractive {
		var err error
		setupWPEngine, err = prompts.SafePromptConfirm("Set up WPEngine integration?", true)
		if err != nil {
			return err
		}
	}

	if !setupWPEngine {
		ui.Info("Skipping WPEngine integration")
		return nil
	}

	var installName string
	var phpVersion string
	var mysqlVersion string
	var autoPopulated bool

	// Only use CLI flag if provided
	if initWPEngineInstall != "" {
		installName = initWPEngineInstall
	} else if initInteractive {
		// Ask if user wants to pick from WPEngine installs
		usePicker, err := prompts.SafePromptConfirm("Would you like to select from your WPEngine installs?", true)
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if usePicker {
			selected, err := promptForWPEngineInstall()
			if err != nil {
				ui.Warning(fmt.Sprintf("Unable to fetch installs: %v", err))
				ui.Info("Falling back to manual entry...")
			} else {
				// Auto-populate from selected install
				installName = selected.Name
				phpVersion = selected.PHPVersion
				mysqlVersion = selected.MySQLVersion
				autoPopulated = true

				ui.Success(fmt.Sprintf("Selected: %s (%s)", selected.Name, selected.Environment))
				if selected.PHPVersion != "" {
					ui.Info(fmt.Sprintf("PHP: %s", selected.PHPVersion))

					// Allow user to override PHP version if they want
					usePHP, err := prompts.SafePromptConfirm(fmt.Sprintf("Use PHP %s from WPEngine?", selected.PHPVersion), true)
					if err != nil {
						return fmt.Errorf("prompt failed: %w", err)
					}
					if !usePHP {
						phpVersion = "" // Will prompt below
					}
				}
				if selected.MySQLVersion != "" {
					ui.Info(fmt.Sprintf("MySQL: %s", selected.MySQLVersion))
				}
				// Display and use WordPress version
				if selected.WordPressVersion != "" {
					ui.Info(fmt.Sprintf("WordPress: %s", selected.WordPressVersion))
					cfg.WordPress.Version = selected.WordPressVersion
				}
			}
		}

		// Only prompt for install name if not auto-populated
		if !autoPopulated {
			name, err := prompts.SafePromptInput("WPEngine install name", "", false)
			if err != nil {
				return err
			}
			installName = name
		}
	}

	cfg.WPEngine.Install = installName

	// Update PHP and MySQL versions if auto-populated
	if autoPopulated {
		if phpVersion != "" {
			cfg.DDEV.PHPVersion = phpVersion
		}
		if mysqlVersion != "" {
			cfg.DDEV.MySQLVersion = mysqlVersion
		}
	}

	// Environment
	if initWPEngineEnv != "" {
		cfg.WPEngine.Environment = initWPEngineEnv
	} else if initInteractive {
		env, err := prompts.SafeEnvironmentPrompt(cfg.WPEngine.Environment)
		if err != nil {
			return err
		}
		cfg.WPEngine.Environment = env
	}

	// Validate environment configuration against WPEngine API
	if cfg.WPEngine.Install != "" && cfg.WPEngine.Environment != "" {
		if err := validateAndFixEnvironment(cfg); err != nil {
			// Non-fatal - just log and continue
			ui.Warning(fmt.Sprintf("Environment validation: %v", err))
		}
	}

	return nil
}

// validateAndFixEnvironment checks if configured environment matches WPEngine API
// and offers to fix it inline during init
func validateAndFixEnvironment(cfg *config.Config) error {
	// Get credentials - skip validation if not available
	creds, err := credentials.GetWPEngineCredentialsWithFallback(cfg.WPEngine.Install)
	if err != nil {
		// Can't validate without credentials
		return nil
	}

	// Query actual environment from WPEngine API
	client := wpeclient.NewClient(creds.APIUser, creds.APIPassword, cfg.WPEngine.Install)
	actualEnv, err := client.GetInstallEnvironment(cfg.WPEngine.Install)
	if err != nil {
		// Can't validate if API fails - not an error, just skip
		return nil
	}

	// No mismatch - nothing to do
	if cfg.WPEngine.Environment == actualEnv {
		return nil
	}

	// There's a mismatch - inform user and offer to fix
	ui.Warning(fmt.Sprintf("Environment mismatch: you selected '%s' but WPEngine reports '%s'",
		cfg.WPEngine.Environment, actualEnv))

	if !prompts.IsInteractive() {
		ui.Info("Run 'stax doctor --fix' to correct this automatically")
		return nil
	}

	// Offer to fix inline
	fix, promptErr := prompts.SafePromptConfirm(
		fmt.Sprintf("Update environment to '%s' to match WPEngine?", actualEnv),
		true,
	)
	if promptErr != nil {
		return promptErr
	}

	if fix {
		cfg.WPEngine.Environment = actualEnv
		ui.Success(fmt.Sprintf("Environment updated to '%s'", actualEnv))
	} else {
		ui.Info("Keeping configured environment. You can fix later with 'stax doctor --fix'")
	}

	return nil
}

// shouldPullDatabase determines if database should be pulled during init
func shouldPullDatabase(cfg *config.Config) bool {
	if initSkipDB {
		return false
	}

	if initPullDB {
		return true
	}

	if cfg.WPEngine.Install == "" {
		return false
	}

	if initInteractive {
		pull, err := prompts.SafePromptConfirm("Pull database from WPEngine now?", false)
		if err != nil {
			return false
		}
		return pull
	}

	return false
}

// shouldPullFiles determines if files should be pulled during init
func shouldPullFiles(cfg *config.Config) bool {
	if initSkipFiles {
		return false
	}

	if initPullFiles {
		return true
	}

	if cfg.WPEngine.Install == "" {
		return false
	}

	if initInteractive {
		pull, err := prompts.SafePromptConfirm("Pull files from WPEngine now?", false)
		if err != nil {
			return false
		}
		return pull
	}

	return false
}

// createWPEngineProviderForListing creates a WPEngine provider for listing installs
func createWPEngineProviderForListing(creds *credentials.WPEngineCredentials) (provider.Provider, error) {
	p := &wpengine.WPEngineProvider{}

	credMap := map[string]string{
		"api_user":     creds.APIUser,
		"api_password": creds.APIPassword,
		"install":      "temp",
		"ssh_gateway":  "ssh.wpengine.net",
		"ssh_key":      "",
	}

	if err := p.Authenticate(credMap); err != nil {
		return nil, err
	}

	return p, nil
}
