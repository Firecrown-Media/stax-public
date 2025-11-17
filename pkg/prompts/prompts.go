package prompts

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/firecrown-media/stax/pkg/ui"
)

// PromptInput prompts for a text input with a default value
func PromptInput(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" && defaultValue != "" {
		return defaultValue, nil
	}

	return input, nil
}

// PromptConfirm prompts for a yes/no confirmation
func PromptConfirm(prompt string, defaultYes bool) (bool, error) {
	var suffix string
	if defaultYes {
		suffix = " [Y/n]: "
	} else {
		suffix = " [y/N]: "
	}

	fmt.Print(prompt + suffix)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultYes, nil
	}

	return input == "y" || input == "yes", nil
}

// PromptSelect prompts to select from a list of options
func PromptSelect(prompt string, options []string, defaultIndex int) (int, string, error) {
	fmt.Println(prompt)
	fmt.Println()

	for i, option := range options {
		marker := " "
		if i == defaultIndex {
			marker = ">"
		}
		fmt.Printf("%s %d. %s\n", marker, i+1, option)
	}
	fmt.Println()

	fmt.Printf("Select [1-%d] (default: %d): ", len(options), defaultIndex+1)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultIndex, options[defaultIndex], nil
	}

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(options) {
		return 0, "", fmt.Errorf("invalid selection: must be between 1 and %d", len(options))
	}

	index := selection - 1
	return index, options[index], nil
}

// PromptMultiSelect prompts to select multiple options from a list
func PromptMultiSelect(prompt string, options []string) ([]int, []string, error) {
	fmt.Println(prompt)
	fmt.Println()

	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}
	fmt.Println()
	fmt.Print("Select (comma-separated, e.g., 1,3,5): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return []int{}, []string{}, nil
	}

	parts := strings.Split(input, ",")
	var indices []int
	var selected []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		selection, err := strconv.Atoi(part)
		if err != nil || selection < 1 || selection > len(options) {
			return nil, nil, fmt.Errorf("invalid selection: %s (must be between 1 and %d)", part, len(options))
		}

		index := selection - 1
		indices = append(indices, index)
		selected = append(selected, options[index])
	}

	return indices, selected, nil
}

// PromptPassword prompts for a password (no echo)
func PromptPassword(prompt string) (string, error) {
	fmt.Print(prompt + ": ")

	// Note: This is a simple implementation
	// For production, consider using golang.org/x/term for proper password input
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	return strings.TrimSpace(password), nil
}

// PromptWithValidation prompts for input with custom validation
func PromptWithValidation(prompt, defaultValue string, validator func(string) error) (string, error) {
	for {
		input, err := PromptInput(prompt, defaultValue)
		if err != nil {
			return "", err
		}

		if validator != nil {
			if err := validator(input); err != nil {
				ui.Warning(err.Error())
				continue
			}
		}

		return input, nil
	}
}

// WPEngineInstallPrompt prompts for WPEngine install name with validation
func WPEngineInstallPrompt(defaultValue string) (string, error) {
	validator := func(input string) error {
		if input == "" {
			return fmt.Errorf("install name cannot be empty")
		}
		// WPEngine install names are alphanumeric, lowercase, and may contain hyphens
		for _, c := range input {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				return fmt.Errorf("install name must be lowercase alphanumeric with hyphens only")
			}
		}
		return nil
	}

	return PromptWithValidation("WPEngine install name", defaultValue, validator)
}

// EnvironmentPrompt prompts for environment selection
func EnvironmentPrompt(defaultEnv string) (string, error) {
	options := []string{"production", "staging", "development"}
	defaultIndex := 1 // staging by default

	for i, opt := range options {
		if opt == defaultEnv {
			defaultIndex = i
			break
		}
	}

	_, environment, err := PromptSelect("Select environment:", options, defaultIndex)
	return environment, err
}

// ProjectTypePrompt prompts for WordPress project type
func ProjectTypePrompt() (string, error) {
	options := []string{
		"wordpress (Single site)",
		"wordpress-multisite-subdomain (Multisite with subdomains)",
		"wordpress-multisite-subdirectory (Multisite with subdirectories)",
	}

	_, selected, err := PromptSelect("Select project type:", options, 0)
	if err != nil {
		return "", err
	}

	// Extract the actual type from the display string
	parts := strings.Split(selected, " ")
	return parts[0], nil
}

// DomainPrompt prompts for domain with validation
func DomainPrompt(defaultDomain string) (string, error) {
	validator := func(input string) error {
		if input == "" {
			return fmt.Errorf("domain cannot be empty")
		}
		// Basic domain validation
		if !strings.Contains(input, ".") {
			return fmt.Errorf("domain must contain at least one dot")
		}
		return nil
	}

	return PromptWithValidation("Primary domain", defaultDomain, validator)
}

// RepositoryPrompt prompts for Git repository URL with validation
func RepositoryPrompt(defaultRepo string) (string, error) {
	validator := func(input string) error {
		if input == "" {
			return nil // Empty is allowed (skip repository cloning)
		}
		// Basic Git URL validation
		if !strings.HasPrefix(input, "git@") && !strings.HasPrefix(input, "https://") {
			return fmt.Errorf("repository URL must start with git@ or https://")
		}
		return nil
	}

	return PromptWithValidation("Git repository URL (optional)", defaultRepo, validator)
}

// WPEngineInstallWithDetails represents an install with metadata
type WPEngineInstallWithDetails struct {
	Name         string
	Environment  string
	PHPVersion   string
	MySQLVersion string
}

// WPEngineInstallPickerPrompt shows a picker to select from WPEngine installs
func WPEngineInstallPickerPrompt(installs []WPEngineInstallWithDetails) (WPEngineInstallWithDetails, error) {
	if len(installs) == 0 {
		return WPEngineInstallWithDetails{}, fmt.Errorf("no installs available")
	}

	// Build options for picker
	options := make([]string, len(installs))
	for i, install := range installs {
		// Format: "install-name (environment) - PHP 8.1, MySQL 8.0"
		options[i] = fmt.Sprintf("%s (%s) - PHP %s, MySQL %s",
			install.Name,
			install.Environment,
			install.PHPVersion,
			install.MySQLVersion,
		)
	}

	idx, _, err := PromptSelect("Select WPEngine Install:", options, 0)
	if err != nil {
		return WPEngineInstallWithDetails{}, err
	}

	return installs[idx], nil
}

// ProgressCallback is a function type for progress updates
type ProgressCallback func(message string, percent int)

// WithProgress wraps a long-running operation with progress updates
func WithProgress(message string, operation func(ProgressCallback) error) error {
	ui.Section(message)

	err := operation(func(msg string, percent int) {
		if percent >= 0 {
			fmt.Printf("\r  %s... %d%%", msg, percent)
		} else {
			fmt.Printf("\r  %s...", msg)
		}
	})

	fmt.Println() // Clear progress line

	if err != nil {
		ui.Error(fmt.Sprintf("Failed: %v", err))
		return err
	}

	ui.Success(message)
	return nil
}
