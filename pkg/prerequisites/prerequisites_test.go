package prerequisites

import (
	"fmt"
	"testing"
)

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Docker version format",
			input:    "Docker version 24.0.6, build ed223bc",
			expected: "24.0.6",
		},
		{
			name:     "DDEV version format",
			input:    "ddev version v1.22.4",
			expected: "1.22.4",
		},
		{
			name:     "Git version format",
			input:    "git version 2.42.0",
			expected: "2.42.0",
		},
		{
			name:     "GitHub CLI version format",
			input:    "gh version 2.40.1 (2023-12-13)",
			expected: "2.40.1",
		},
		{
			name:     "Simple version",
			input:    "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "Version with v prefix",
			input:    "v3.4.5",
			expected: "3.4.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersion(tt.input)
			if result != tt.expected {
				t.Errorf("extractVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsVersionSufficient(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		minVersion string
		expected   bool
	}{
		{
			name:       "equal versions",
			version:    "1.2.3",
			minVersion: "1.2.3",
			expected:   true,
		},
		{
			name:       "higher major version",
			version:    "2.0.0",
			minVersion: "1.9.9",
			expected:   true,
		},
		{
			name:       "higher minor version",
			version:    "1.3.0",
			minVersion: "1.2.9",
			expected:   true,
		},
		{
			name:       "higher patch version",
			version:    "1.2.4",
			minVersion: "1.2.3",
			expected:   true,
		},
		{
			name:       "lower major version",
			version:    "1.0.0",
			minVersion: "2.0.0",
			expected:   false,
		},
		{
			name:       "lower minor version",
			version:    "1.2.0",
			minVersion: "1.3.0",
			expected:   false,
		},
		{
			name:       "lower patch version",
			version:    "1.2.2",
			minVersion: "1.2.3",
			expected:   false,
		},
		{
			name:       "version with v prefix",
			version:    "v1.22.4",
			minVersion: "1.22.0",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVersionSufficient(tt.version, tt.minVersion)
			if result != tt.expected {
				t.Errorf("isVersionSufficient(%q, %q) = %v, want %v",
					tt.version, tt.minVersion, result, tt.expected)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "simple version",
			input:    "1.2.3",
			expected: []int{1, 2, 3},
		},
		{
			name:     "version with v prefix",
			input:    "v1.22.4",
			expected: []int{1, 22, 4},
		},
		{
			name:     "version with extra text",
			input:    "version 24.0.6",
			expected: []int{24, 0, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersion(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseVersion(%q) returned %d parts, want %d",
					tt.input, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseVersion(%q)[%d] = %d, want %d",
						tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestDefaultDependencies(t *testing.T) {
	deps := DefaultDependencies()

	if len(deps) == 0 {
		t.Error("DefaultDependencies() returned empty list")
	}

	// Check that we have the expected dependencies
	names := make(map[string]bool)
	for _, dep := range deps {
		names[dep.Name] = true
	}

	expected := []string{"Docker", "DDEV", "Git", "GitHub CLI"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("DefaultDependencies() missing %s", name)
		}
	}
}

func TestRequiredOnly(t *testing.T) {
	deps := RequiredOnly()

	for _, dep := range deps {
		if !dep.Required {
			t.Errorf("RequiredOnly() returned non-required dependency: %s", dep.Name)
		}
	}

	// Should have at least Docker, DDEV, Git
	if len(deps) < 3 {
		t.Errorf("RequiredOnly() returned %d dependencies, expected at least 3", len(deps))
	}
}

func TestGetDependency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"by name", "Docker", "Docker"},
		{"by command", "docker", "Docker"},
		{"case insensitive", "DOCKER", "Docker"},
		{"ddev by name", "DDEV", "DDEV"},
		{"ddev by command", "ddev", "DDEV"},
		{"not found", "nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDependency(tt.input)
			if tt.expected == "" {
				if result != nil {
					t.Errorf("GetDependency(%q) = %s, want nil", tt.input, result.Name)
				}
			} else {
				if result == nil {
					t.Errorf("GetDependency(%q) = nil, want %s", tt.input, tt.expected)
				} else if result.Name != tt.expected {
					t.Errorf("GetDependency(%q).Name = %s, want %s", tt.input, result.Name, tt.expected)
				}
			}
		})
	}
}

func TestFilterFailed(t *testing.T) {
	results := []CheckResult{
		{Dependency: Dependency{Name: "A"}, Installed: true},
		{Dependency: Dependency{Name: "B"}, Installed: false},
		{Dependency: Dependency{Name: "C"}, Installed: true},
	}

	failed := FilterFailed(results)

	if len(failed) != 1 {
		t.Errorf("FilterFailed() returned %d results, want 1", len(failed))
	}

	if len(failed) > 0 && failed[0].Dependency.Name != "B" {
		t.Errorf("FilterFailed()[0].Name = %s, want B", failed[0].Dependency.Name)
	}
}

func TestHasFailedRequired(t *testing.T) {
	tests := []struct {
		name     string
		results  []CheckResult
		expected bool
	}{
		{
			name: "no failures",
			results: []CheckResult{
				{Dependency: Dependency{Name: "A", Required: true}, Installed: true},
			},
			expected: false,
		},
		{
			name: "required failed",
			results: []CheckResult{
				{Dependency: Dependency{Name: "A", Required: true}, Installed: false},
			},
			expected: true,
		},
		{
			name: "optional failed",
			results: []CheckResult{
				{Dependency: Dependency{Name: "A", Required: false}, Installed: false},
			},
			expected: false,
		},
		{
			name: "mixed",
			results: []CheckResult{
				{Dependency: Dependency{Name: "A", Required: true}, Installed: true},
				{Dependency: Dependency{Name: "B", Required: false}, Installed: false},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasFailedRequired(tt.results)
			if result != tt.expected {
				t.Errorf("HasFailedRequired() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCheckResult_OK tests the OK method
func TestCheckResult_OK(t *testing.T) {
	tests := []struct {
		name     string
		result   CheckResult
		expected bool
	}{
		{
			name:     "installed no error",
			result:   CheckResult{Installed: true, Error: nil},
			expected: true,
		},
		{
			name:     "not installed",
			result:   CheckResult{Installed: false, Error: nil},
			expected: false,
		},
		{
			name:     "installed with error",
			result:   CheckResult{Installed: true, Error: fmt.Errorf("some error")},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.OK() != tt.expected {
				t.Errorf("CheckResult.OK() = %v, want %v", tt.result.OK(), tt.expected)
			}
		})
	}
}
