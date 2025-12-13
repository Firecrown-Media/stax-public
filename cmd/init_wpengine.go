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

	// Transform to picker format
	installDetails := make([]prompts.WPEngineInstallWithDetails, len(installs))
	for i, install := range installs {
		installDetails[i] = prompts.WPEngineInstallWithDetails{
			Name:         install.Name,
			Environment:  install.Environment,
			PHPVersion:   install.PHPVersion,
			MySQLVersion: "", // Not available from ListInstalls API - would require 60+ individual API calls
		}
	}

	// Show picker
	selected, err := prompts.WPEngineInstallPickerPrompt(installDetails)
	if err != nil {
		return nil, err
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
