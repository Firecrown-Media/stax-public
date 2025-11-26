package prompts

import (
	"os"
	"testing"
)

// TestIsInteractive tests TTY detection
func TestIsInteractive(t *testing.T) {
	// Note: This test will return false when running in CI/CD or test runners
	// since stdin won't be a TTY in those contexts
	result := IsInteractive()

	// We can't assert a specific value since it depends on test environment
	// But we can verify the function doesn't panic
	t.Logf("IsInteractive returned: %v", result)

	// Test that it handles stdin properly
	if os.Stdin == nil {
		t.Error("os.Stdin should not be nil")
	}
}

// TestSafePromptInput_NonInteractive tests safe prompts in non-interactive mode
func TestSafePromptInput_NonInteractive(t *testing.T) {
	// This test assumes we're running in non-interactive mode (test runner)
	// If running in interactive terminal, skip
	if IsInteractive() {
		t.Skip("Skipping non-interactive test in interactive terminal")
	}

	tests := []struct {
		name         string
		prompt       string
		defaultValue string
		required     bool
		wantValue    string
		wantError    bool
	}{
		{
			name:         "with default, not required",
			prompt:       "Test prompt",
			defaultValue: "default",
			required:     false,
			wantValue:    "default",
			wantError:    false,
		},
		{
			name:         "without default, not required",
			prompt:       "Test prompt",
			defaultValue: "",
			required:     false,
			wantValue:    "",
			wantError:    false,
		},
		{
			name:         "without default, required",
			prompt:       "Test prompt",
			defaultValue: "",
			required:     true,
			wantValue:    "",
			wantError:    true,
		},
		{
			name:         "with default, required",
			prompt:       "Test prompt",
			defaultValue: "default",
			required:     true,
			wantValue:    "default",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafePromptInput(tt.prompt, tt.defaultValue, tt.required)

			if (err != nil) != tt.wantError {
				t.Errorf("SafePromptInput() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if got != tt.wantValue {
				t.Errorf("SafePromptInput() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

// TestSafePromptConfirm_NonInteractive tests safe confirm prompts in non-interactive mode
func TestSafePromptConfirm_NonInteractive(t *testing.T) {
	if IsInteractive() {
		t.Skip("Skipping non-interactive test in interactive terminal")
	}

	tests := []struct {
		name       string
		prompt     string
		defaultYes bool
		want       bool
	}{
		{
			name:       "default yes",
			prompt:     "Confirm?",
			defaultYes: true,
			want:       true,
		},
		{
			name:       "default no",
			prompt:     "Confirm?",
			defaultYes: false,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafePromptConfirm(tt.prompt, tt.defaultYes)
			if err != nil {
				t.Errorf("SafePromptConfirm() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("SafePromptConfirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSafePromptSelect_NonInteractive tests safe select prompts in non-interactive mode
func TestSafePromptSelect_NonInteractive(t *testing.T) {
	if IsInteractive() {
		t.Skip("Skipping non-interactive test in interactive terminal")
	}

	options := []string{"option1", "option2", "option3"}

	tests := []struct {
		name         string
		prompt       string
		options      []string
		defaultIndex int
		wantIndex    int
		wantValue    string
		wantError    bool
	}{
		{
			name:         "valid default index",
			prompt:       "Select",
			options:      options,
			defaultIndex: 1,
			wantIndex:    1,
			wantValue:    "option2",
			wantError:    false,
		},
		{
			name:         "first option",
			prompt:       "Select",
			options:      options,
			defaultIndex: 0,
			wantIndex:    0,
			wantValue:    "option1",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndex, gotValue, err := SafePromptSelect(tt.prompt, tt.options, tt.defaultIndex)

			if (err != nil) != tt.wantError {
				t.Errorf("SafePromptSelect() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if gotIndex != tt.wantIndex {
				t.Errorf("SafePromptSelect() index = %v, want %v", gotIndex, tt.wantIndex)
			}

			if gotValue != tt.wantValue {
				t.Errorf("SafePromptSelect() value = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

// TestSafeEnvironmentPrompt_NonInteractive tests safe environment prompts
func TestSafeEnvironmentPrompt_NonInteractive(t *testing.T) {
	if IsInteractive() {
		t.Skip("Skipping non-interactive test in interactive terminal")
	}

	tests := []struct {
		name       string
		defaultEnv string
		want       string
	}{
		{
			name:       "default production",
			defaultEnv: "production",
			want:       "production",
		},
		{
			name:       "default staging",
			defaultEnv: "staging",
			want:       "staging",
		},
		{
			name:       "default development",
			defaultEnv: "development",
			want:       "development",
		},
		{
			name:       "empty default uses staging",
			defaultEnv: "",
			want:       "staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeEnvironmentPrompt(tt.defaultEnv)
			if err != nil {
				t.Errorf("SafeEnvironmentPrompt() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("SafeEnvironmentPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
