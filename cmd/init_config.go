package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ddev"
	"github.com/firecrown-media/stax/pkg/errors"
	"github.com/firecrown-media/stax/pkg/prompts"
	"github.com/firecrown-media/stax/pkg/ui"
)

// gatherProjectConfiguration collects all project configuration from flags or prompts
func gatherProjectConfiguration(projectDir string) (*config.Config, error) {
	ui.Section("Project Configuration")

	cfg := config.Defaults()

	// Project name
	defaultName := filepath.Base(projectDir)
	if initName != "" {
		cfg.Project.Name = initName
	} else if initInteractive {
		name, err := prompts.SafePromptInput("Project name", defaultName, false)
		if err != nil {
			return nil, err
		}
		cfg.Project.Name = name
	} else {
		cfg.Project.Name = defaultName
	}

	// Project type
	if initType != "" {
		cfg.Project.Type = initType
		if strings.Contains(initType, "multisite") {
			if initMode != "" {
				cfg.Project.Mode = initMode
			}
		} else {
			cfg.Project.Mode = "single"
		}
	} else if initInteractive {
		projectType, err := prompts.SafeProjectTypePrompt()
		if err != nil {
			return nil, err
		}
		cfg.Project.Type = projectType

		if isMultisite(projectType) {
			mode, err := promptMultisiteModeSafe()
			if err != nil {
				return nil, err
			}
			cfg.Project.Mode = mode
		} else {
			cfg.Project.Mode = "single"
		}
	}

	// Set initial DDEV configuration from flags
	cfg.DDEV.PHPVersion = initPHPVersion
	cfg.DDEV.MySQLVersion = initMySQLVersion

	// WPEngine configuration (may auto-populate PHP/MySQL versions)
	if err := gatherWPEngineConfiguration(cfg); err != nil {
		return nil, err
	}

	// DDEV configuration - prompt only in interactive mode and if not already set by WPEngine picker
	if initInteractive {
		// Only prompt for PHP version if not already set by WPEngine picker
		if cfg.DDEV.PHPVersion == initPHPVersion {
			phpVersion, err := prompts.SafePromptInput("PHP version", cfg.DDEV.PHPVersion, false)
			if err != nil {
				return nil, err
			}
			cfg.DDEV.PHPVersion = phpVersion
		}

		// Only prompt for MySQL version if not already set by WPEngine picker
		if cfg.DDEV.MySQLVersion == initMySQLVersion {
			mysqlVersion, err := prompts.SafePromptInput("MySQL version", cfg.DDEV.MySQLVersion, false)
			if err != nil {
				return nil, err
			}
			cfg.DDEV.MySQLVersion = mysqlVersion
		}
	}

	// Repository configuration
	if err := gatherRepositoryConfiguration(cfg); err != nil {
		return nil, err
	}

	// Network domain for multisite
	if isMultisite(cfg.Project.Type) {
		if initInteractive {
			defaultDomain := fmt.Sprintf("%s.ddev.site", cfg.Project.Name)
			domain, err := prompts.SafePromptInput("Primary domain", defaultDomain, false)
			if err != nil {
				return nil, err
			}
			cfg.Network.Domain = domain
		} else {
			cfg.Network.Domain = fmt.Sprintf("%s.ddev.site", cfg.Project.Name)
		}
	}

	return cfg, nil
}

// gatherRepositoryConfiguration collects repository configuration from flags or prompts
func gatherRepositoryConfiguration(cfg *config.Config) error {
	ui.Section("Repository Configuration")

	cloneRepo := false
	if initRepo != "" {
		cloneRepo = true
	} else if initInteractive {
		var err error
		cloneRepo, err = prompts.SafePromptConfirm("Clone from Git repository?", false)
		if err != nil {
			return err
		}
	}

	if !cloneRepo {
		ui.Info("Skipping repository cloning")
		return nil
	}

	// Repository URL
	if initRepo != "" {
		cfg.Repository.URL = initRepo
	} else if initInteractive {
		repoURL, err := prompts.SafeRepositoryPrompt("")
		if err != nil {
			return err
		}
		cfg.Repository.URL = repoURL
	}

	// Branch
	if initBranch != "" {
		cfg.Repository.Branch = initBranch
	} else if initInteractive {
		branch, err := prompts.SafePromptInput("Repository branch", "main", false)
		if err != nil {
			return err
		}
		cfg.Repository.Branch = branch
	}

	return nil
}

// checkExistingConfiguration checks for existing .stax.yml and DDEV configs
func checkExistingConfiguration(projectDir string) error {
	configPath := filepath.Join(projectDir, ".stax.yml")
	if _, err := os.Stat(configPath); err == nil {
		if initInteractive {
			overwrite, err := prompts.SafePromptConfirm(".stax.yml already exists. Overwrite?", false)
			if err != nil {
				return err
			}
			if !overwrite {
				return fmt.Errorf("initialization cancelled by user")
			}
		} else {
			return errors.NewWithSolution(
				"Configuration already exists",
				".stax.yml already exists in this directory",
				errors.Solution{
					Description: "Choose an action",
					Steps: []string{
						"Run with --interactive to confirm overwrite",
						"Remove .stax.yml manually and try again",
						"Use a different directory",
					},
				},
			)
		}
	}

	// Check for existing DDEV config
	if ddev.IsConfigured(projectDir) {
		ui.Warning("DDEV configuration already exists")
		if initInteractive {
			overwrite, err := prompts.SafePromptConfirm("Overwrite existing DDEV configuration?", false)
			if err != nil {
				return err
			}
			if !overwrite {
				ui.Info("Will preserve existing DDEV configuration")
			}
		}
	}

	return nil
}

// promptProjectType prompts for WordPress project type selection
func promptProjectType() (string, error) {
	options := []string{
		"wordpress",
		"wordpress-multisite",
	}

	idx, selected, err := prompts.PromptSelect("Select project type:", options, 0)
	if err != nil {
		return "", err
	}

	ui.Info("Selected: %s", selected)
	return options[idx], nil
}

// promptMultisiteMode prompts for multisite mode selection
func promptMultisiteMode() (string, error) {
	options := []string{
		"subdomain",
		"subdirectory",
	}

	idx, selected, err := prompts.PromptSelect("Select multisite mode:", options, 0)
	if err != nil {
		return "", err
	}

	ui.Info("Selected: %s", selected)
	return options[idx], nil
}

// promptMultisiteModeSafe prompts for multisite mode with safe fallback for non-interactive mode
func promptMultisiteModeSafe() (string, error) {
	options := []string{
		"subdomain",
		"subdirectory",
	}

	idx, selected, err := prompts.SafePromptSelect("Select multisite mode:", options, 0)
	if err != nil {
		return "", err
	}

	ui.Info("Selected: %s", selected)
	return options[idx], nil
}
