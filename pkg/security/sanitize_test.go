package security

import (
	"fmt"
	"strings"
	"testing"
)

func TestSanitizeForShell(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		// Valid cases
		{"simple path", "path/to/file", false},
		{"with hyphen", "my-file", false},
		{"with underscore", "my_file", false},
		{"with dot", "file.txt", false},
		{"absolute path", "/var/www/html", false},

		// Invalid cases
		{"with space", "my file", true},
		{"with semicolon", "file;rm", true},
		{"with pipe", "file|cat", true},
		{"with ampersand", "file&malicious", true},
		{"with dollar", "file$variable", true},
		{"with backtick", "file`cmd`", true},
		{"with parenthesis", "file()", true},
		{"with angle bracket", "file<>", true},
		{"with quote", "file'", true},
		{"with double quote", "file\"", true},
		{"with backslash", "file\\path", true},
		{"with newline", "file\n", true},
		{"with tab", "file\t", true},
		{"with asterisk", "file*", true},
		{"with question mark", "file?", true},
		{"with bracket", "file[", true},
		{"with brace", "file{", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SanitizeForShell(tt.input)
			if tt.shouldErr && err == nil {
				t.Errorf("SanitizeForShell(%q) expected error, got nil", tt.input)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("SanitizeForShell(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestSanitizeCommandArgs(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		shouldErr bool
	}{
		// Valid cases (alphabetic characters are allowed, but not all special chars)
		{"simple args", []string{"arg1", "arg2"}, false},
		{"path args", []string{"/path/to/file"}, false},

		// Invalid cases
		{"with semicolon", []string{"arg;malicious"}, true},
		{"with pipe", []string{"arg|cat"}, true},
		{"with backtick", []string{"arg`cmd`"}, true},
		{"command substitution", []string{"$(whoami)"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SanitizeCommandArgs(tt.input)
			if tt.shouldErr && err == nil {
				t.Errorf("SanitizeCommandArgs(%v) expected error, got nil", tt.input)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("SanitizeCommandArgs(%v) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestRemoveSensitiveData(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		shouldNotContain string
	}{
		{
			"password in connection string",
			"Error: connection to mysql://user:password123@localhost failed",
			"password123",
		},
		{
			"password parameter",
			"Error: password=secret123 failed",
			"secret123",
		},
		{
			"token parameter",
			"Error: token=abc123def failed",
			"abc123def",
		},
		{
			"api_key parameter",
			"Error: api_key=mykey123 failed",
			"mykey123",
		},
		{
			"Authorization header",
			"Error: Authorization: Bearer secrettoken failed",
			"Bearer secrettoken",
		},
		{
			"mysql connection string",
			"Error: mysql://user:pass123@localhost/db",
			"pass123",
		},
		{
			"postgres connection string",
			"Error: postgres://user:secret@localhost/db",
			"secret",
		},
		{
			"SSH private key",
			"Error: key content -----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKC...\n-----END RSA PRIVATE KEY----- found",
			"MIIEowIBAAKC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveSensitiveData(tt.input)
			if strings.Contains(result, tt.shouldNotContain) {
				t.Errorf("RemoveSensitiveData() leaked sensitive data: %q still in output", tt.shouldNotContain)
			}
			if !strings.Contains(result, "***") && tt.shouldNotContain != "" {
				t.Errorf("RemoveSensitiveData() didn't redact anything, expected redaction")
			}
		})
	}
}

func TestSanitizeLogMessage(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantLen   int
	}{
		{"short message", "Error occurred", 100, 14},
		{"long message truncated", strings.Repeat("a", 200), 50, 50 + len("... [truncated]")},
		{"with password", "Error: password=secret", 100, -1}, // Don't check length, just that it's sanitized
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeLogMessage(tt.input, tt.maxLength)
			if tt.wantLen > 0 && len(result) != tt.wantLen {
				t.Errorf("SanitizeLogMessage() length = %d, want %d", len(result), tt.wantLen)
			}
			if strings.Contains(tt.input, "password") && strings.Contains(result, "secret") {
				t.Errorf("SanitizeLogMessage() leaked password")
			}
		})
	}
}

func TestSanitizeErrorForUser(t *testing.T) {
	tests := []struct {
		name             string
		input            error
		shouldNotContain []string
	}{
		{
			"error with password",
			fmt.Errorf("authentication failed with password=secret123"),
			[]string{"secret123"},
		},
		{
			"error with path",
			fmt.Errorf("failed to read /Users/user/stax/config.yaml"),
			[]string{"/Users/user/stax"},
		},
		{
			"nil error",
			nil,
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeErrorForUser(tt.input)
			for _, secret := range tt.shouldNotContain {
				if strings.Contains(result, secret) {
					t.Errorf("SanitizeErrorForUser() leaked: %q", secret)
				}
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
		expected  string
	}{
		// Valid cases
		{"simple", "file.txt", false, "file.txt"},
		{"with hyphen", "my-file.txt", false, "my-file.txt"},
		{"with underscore", "my_file.txt", false, "my_file.txt"},
		{"with dots", "file.tar.gz", false, "file.tar.gz"},

		// Invalid cases
		{"empty", "", true, ""},
		{"with slash", "path/file.txt", false, "pathfile.txt"},      // Slashes are removed
		{"with backslash", "path\\file.txt", false, "pathfile.txt"}, // Backslashes are removed
		{"double dot", "..", true, ""},
		{"single dot", ".", true, ""},
		{"tilde", "~", true, ""},
		{"too long", strings.Repeat("a", 256), true, ""},
		{"with space", "my file.txt", true, ""},
		{"with special char", "file@name.txt", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeFilename(tt.input)
			if tt.shouldErr && err == nil {
				t.Errorf("SanitizeFilename(%q) expected error, got nil", tt.input)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("SanitizeFilename(%q) unexpected error: %v", tt.input, err)
			}
			if !tt.shouldErr && result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeWPCLIArgs(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		shouldErr bool
	}{
		// Valid cases
		{"simple args", []string{"option", "get", "siteurl"}, false},
		{"with flags", []string{"search-replace", "old", "new", "--skip-columns=guid"}, false},
		{"with equals", []string{"--url=http://example.com"}, false},

		// Invalid cases
		{"command substitution", []string{"$(whoami)"}, true},
		{"backtick", []string{"`id`"}, true},
		{"semicolon", []string{"arg;malicious"}, true},
		{"pipe", []string{"arg|cat"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SanitizeWPCLIArgs(tt.input)
			if tt.shouldErr && err == nil {
				t.Errorf("SanitizeWPCLIArgs(%v) expected error, got nil", tt.input)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("SanitizeWPCLIArgs(%v) unexpected error: %v", tt.input, err)
			}
		})
	}
}

// Test shell metacharacter detection
func TestContainsShellMetachars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Contains metacharacters
		{"semicolon", "test;cmd", true},
		{"pipe", "test|cmd", true},
		{"ampersand", "test&cmd", true},
		{"dollar", "test$var", true},
		{"backtick", "test`cmd`", true},
		{"parenthesis", "test()", true},
		{"angle bracket", "test<file", true},
		{"newline", "test\n", true},
		{"asterisk", "test*", true},
		{"question", "test?", true},
		{"bracket", "test[0]", true},
		{"brace", "test{}", true},
		{"backslash", "test\\cmd", true},
		{"single quote", "test'", true},
		{"double quote", "test\"", true},

		// No metacharacters
		{"simple", "test", false},
		{"with hyphen", "test-file", false},
		{"with underscore", "test_file", false},
		{"with dot", "test.txt", false},
		{"with slash", "test/path", false},
		{"with colon", "test:value", false},
		{"with equals", "test=value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsShellMetachars(tt.input)
			if result != tt.expected {
				t.Errorf("containsShellMetachars(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Fuzzing-style tests for credential leakage
func TestCredentialLeakageFuzzing(t *testing.T) {
	credentials := []struct {
		pattern string
		secret  string
	}{
		{"password=%s", "secret123"},
		{"pass=%s", "mypass"},
		{"token=%s", "abc123"},
		{"api_key=%s", "key123"},
		{"Authorization: Bearer %s", "token123"},
	}

	for _, cred := range credentials {
		t.Run(cred.pattern, func(t *testing.T) {
			input := fmt.Sprintf(cred.pattern, cred.secret)
			result := RemoveSensitiveData(input)
			if strings.Contains(result, cred.secret) {
				t.Errorf("RemoveSensitiveData() leaked credential: %q in %q", cred.secret, result)
			}
		})
	}
}

// Test that sanitization preserves functionality
func TestSanitizationPreservesValidInput(t *testing.T) {
	validPaths := []string{
		"/var/www/html",
		"path/to/file.txt",
		"./relative/path",
		"filename.tar.gz",
	}

	for _, path := range validPaths {
		t.Run(path, func(t *testing.T) {
			result, err := SanitizeForShell(path)
			if err != nil {
				t.Errorf("SanitizeForShell(%q) rejected valid path: %v", path, err)
			}
			if result != path {
				t.Errorf("SanitizeForShell(%q) = %q, want unchanged", path, result)
			}
		})
	}
}

// Benchmark sanitization functions
func BenchmarkSanitizeForShell(b *testing.B) {
	input := "/path/to/some/file.txt"
	for i := 0; i < b.N; i++ {
		_, _ = SanitizeForShell(input)
	}
}

func BenchmarkRemoveSensitiveData(b *testing.B) {
	input := "Error: failed to connect with password=secret123 and token=abc456"
	for i := 0; i < b.N; i++ {
		RemoveSensitiveData(input)
	}
}

func BenchmarkSanitizeWPCLIArgs(b *testing.B) {
	args := []string{"search-replace", "http://old.com", "http://new.com", "--skip-columns=guid"}
	for i := 0; i < b.N; i++ {
		_, _ = SanitizeWPCLIArgs(args)
	}
}
