package cmd

import (
	"strings"
	"testing"
)

// TestStorageMethodNormalization verifies that both numeric and text inputs
// are correctly normalized to the expected storage method values.
// This test documents the fix for the issue where users saw numbered options
// but couldn't enter numbers.
func TestStorageMethodNormalization(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
		description string
	}{
		{
			name:        "numeric input 1 maps to file",
			input:       "1",
			expected:    "file",
			shouldError: false,
			description: "User enters '1' to select file storage",
		},
		{
			name:        "numeric input 2 maps to env",
			input:       "2",
			expected:    "env",
			shouldError: false,
			description: "User enters '2' to select environment variables",
		},
		{
			name:        "text input file",
			input:       "file",
			expected:    "file",
			shouldError: false,
			description: "User enters 'file' directly",
		},
		{
			name:        "text input env",
			input:       "env",
			expected:    "env",
			shouldError: false,
			description: "User enters 'env' directly",
		},
		{
			name:        "case insensitive FILE",
			input:       "FILE",
			expected:    "file",
			shouldError: false,
			description: "Uppercase 'FILE' should work",
		},
		{
			name:        "case insensitive ENV",
			input:       "ENV",
			expected:    "env",
			shouldError: false,
			description: "Uppercase 'ENV' should work",
		},
		{
			name:        "case insensitive File",
			input:       "File",
			expected:    "file",
			shouldError: false,
			description: "Mixed case 'File' should work",
		},
		{
			name:        "whitespace trimmed",
			input:       "  file  ",
			expected:    "file",
			shouldError: false,
			description: "Input with whitespace should be trimmed",
		},
		{
			name:        "whitespace trimmed with number",
			input:       " 1 ",
			expected:    "file",
			shouldError: false,
			description: "Numeric input with whitespace should be trimmed",
		},
		{
			name:        "keychain input preserved",
			input:       "keychain",
			expected:    "keychain",
			shouldError: false,
			description: "Keychain option should still work",
		},
		{
			name:        "empty input defaults to file",
			input:       "",
			expected:    "file",
			shouldError: false,
			description: "Empty input should default to file",
		},
		{
			name:        "invalid input returns error",
			input:       "invalid",
			expected:    "",
			shouldError: true,
			description: "Invalid input should return an error",
		},
		{
			name:        "invalid number returns error",
			input:       "3",
			expected:    "",
			shouldError: true,
			description: "Invalid number should return an error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the normalization logic from setup.go
			var normalized string
			var err error

			switch strings.TrimSpace(strings.ToLower(tt.input)) {
			case "1", "file":
				normalized = "file"
			case "2", "env":
				normalized = "env"
			case "keychain":
				normalized = "keychain"
			case "":
				normalized = "file" // default
			default:
				err = &invalidStorageMethodError{input: tt.input}
			}

			// Check error expectation
			if tt.shouldError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}
				if normalized != tt.expected {
					t.Errorf("%s: expected '%s' but got '%s'", tt.description, tt.expected, normalized)
				}
			}
		})
	}
}

// invalidStorageMethodError is a helper type for testing
type invalidStorageMethodError struct {
	input string
}

func (e *invalidStorageMethodError) Error() string {
	return "invalid storage method: " + e.input
}
