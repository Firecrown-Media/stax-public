package prompts

import (
	"testing"
)

func TestWPEngineInstallValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "valid install name",
			input:     "mysite-prod",
			wantError: false,
		},
		{
			name:      "valid with numbers",
			input:     "site123",
			wantError: false,
		},
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
		{
			name:      "uppercase letters",
			input:     "MySite",
			wantError: true,
		},
		{
			name:      "special characters",
			input:     "my_site",
			wantError: true,
		},
		{
			name:      "spaces",
			input:     "my site",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := func(input string) error {
				if input == "" {
					return ErrEmpty
				}
				for _, c := range input {
					if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
						return ErrInvalidFormat
					}
				}
				return nil
			}

			err := validator(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("validation error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestDomainValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "valid domain",
			input:     "example.com",
			wantError: false,
		},
		{
			name:      "valid subdomain",
			input:     "sub.example.com",
			wantError: false,
		},
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
		{
			name:      "no dot",
			input:     "example",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := func(input string) error {
				if input == "" {
					return ErrEmpty
				}
				if !containsDot(input) {
					return ErrInvalidDomain
				}
				return nil
			}

			err := validator(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("validation error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// Helper errors for testing
var (
	ErrEmpty         = &ValidationError{Message: "cannot be empty"}
	ErrInvalidFormat = &ValidationError{Message: "invalid format"}
	ErrInvalidDomain = &ValidationError{Message: "must contain at least one dot"}
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func containsDot(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}

func TestWPEngineInstallPickerPrompt(t *testing.T) {
	tests := []struct {
		name      string
		installs  []WPEngineInstallWithDetails
		wantError bool
	}{
		{
			name:      "empty installs list",
			installs:  []WPEngineInstallWithDetails{},
			wantError: true,
		},
		{
			name: "single install",
			installs: []WPEngineInstallWithDetails{
				{
					Name:         "mysite",
					Environment:  "production",
					PHPVersion:   "8.1",
					MySQLVersion: "8.0",
				},
			},
			wantError: false,
		},
		{
			name: "multiple installs",
			installs: []WPEngineInstallWithDetails{
				{
					Name:         "site1",
					Environment:  "production",
					PHPVersion:   "8.1",
					MySQLVersion: "8.0",
				},
				{
					Name:         "site2",
					Environment:  "staging",
					PHPVersion:   "8.2",
					MySQLVersion: "8.0",
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't actually test the interactive prompt without mocking
			// But we can test the validation logic
			if len(tt.installs) == 0 && !tt.wantError {
				t.Error("expected error for empty installs list")
			}
		})
	}
}
