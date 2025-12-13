package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewQuality(t *testing.T) {
	projectPath := "/tmp/test-project"
	quality := NewQuality(projectPath)

	if quality == nil {
		t.Fatal("NewQuality returned nil")
	}

	if quality.projectPath != projectPath {
		t.Errorf("expected projectPath %s, got %s", projectPath, quality.projectPath)
	}
}

func TestQuality_GetPHPCSConfig(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"phpcs.xml.dist", ".phpcs.xml.dist"},
		{"phpcs.xml", "phpcs.xml"},
		{"dot phpcs.xml", ".phpcs.xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create config file
			configPath := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(configPath, []byte("<ruleset name='test'></ruleset>"), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			quality := NewQuality(tmpDir)
			result, err := quality.GetPHPCSConfig()
			if err != nil {
				t.Fatalf("GetPHPCSConfig failed: %v", err)
			}

			if result != configPath {
				t.Errorf("expected config path %s, got %s", configPath, result)
			}
		})
	}
}

func TestQuality_GetPHPCSConfig_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	quality := NewQuality(tmpDir)
	_, err := quality.GetPHPCSConfig()

	if err == nil {
		t.Error("expected error when no config file exists")
	}
}

func TestQuality_GetPHPCSConfig_Priority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both .phpcs.xml.dist and phpcs.xml
	// .phpcs.xml.dist should take priority
	distConfig := filepath.Join(tmpDir, ".phpcs.xml.dist")
	if err := os.WriteFile(distConfig, []byte("<ruleset name='dist'></ruleset>"), 0644); err != nil {
		t.Fatalf("failed to write .phpcs.xml.dist: %v", err)
	}

	xmlConfig := filepath.Join(tmpDir, "phpcs.xml")
	if err := os.WriteFile(xmlConfig, []byte("<ruleset name='xml'></ruleset>"), 0644); err != nil {
		t.Fatalf("failed to write phpcs.xml: %v", err)
	}

	quality := NewQuality(tmpDir)
	result, err := quality.GetPHPCSConfig()
	if err != nil {
		t.Fatalf("GetPHPCSConfig failed: %v", err)
	}

	if result != distConfig {
		t.Errorf("expected .phpcs.xml.dist to take priority, got %s", result)
	}
}

func TestQuality_ValidatePHPCSConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid config
	configPath := filepath.Join(tmpDir, "phpcs.xml")
	validConfig := `<?xml version="1.0"?>
<ruleset name="Test">
    <description>Test coding standard</description>
    <rule ref="PSR12"/>
</ruleset>`

	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	quality := NewQuality(tmpDir)
	if err := quality.ValidatePHPCSConfig(); err != nil {
		t.Errorf("ValidatePHPCSConfig should pass for valid config: %v", err)
	}
}

func TestQuality_ValidatePHPCSConfig_Invalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid config (missing ruleset)
	configPath := filepath.Join(tmpDir, "phpcs.xml")
	if err := os.WriteFile(configPath, []byte("<notaruleset></notaruleset>"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	quality := NewQuality(tmpDir)
	if err := quality.ValidatePHPCSConfig(); err == nil {
		t.Error("ValidatePHPCSConfig should fail for invalid config")
	}
}

func TestQuality_ValidatePHPCSConfig_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	quality := NewQuality(tmpDir)
	if err := quality.ValidatePHPCSConfig(); err == nil {
		t.Error("ValidatePHPCSConfig should fail when no config file exists")
	}
}

func TestQuality_buildPHPCSArgs(t *testing.T) {
	tmpDir := t.TempDir()
	quality := NewQuality(tmpDir)

	tests := []struct {
		name     string
		options  PHPCSOptions
		contains []string
	}{
		{
			name:     "default options",
			options:  PHPCSOptions{},
			contains: []string{"--report=json", "."},
		},
		{
			name: "with standard",
			options: PHPCSOptions{
				Standard: "PSR12",
			},
			contains: []string{"--standard=PSR12"},
		},
		{
			name: "with config file",
			options: PHPCSOptions{
				ConfigFile: "/path/to/phpcs.xml",
			},
			contains: []string{"--standard=/path/to/phpcs.xml"},
		},
		{
			name: "with extensions",
			options: PHPCSOptions{
				Extensions: []string{"php", "inc"},
			},
			contains: []string{"--extensions=php,inc"},
		},
		{
			name: "with ignore",
			options: PHPCSOptions{
				Ignore: "vendor/*,node_modules/*",
			},
			contains: []string{"--ignore=vendor/*,node_modules/*"},
		},
		{
			name: "with report format",
			options: PHPCSOptions{
				Report: "summary",
			},
			contains: []string{"--report=summary"},
		},
		{
			name: "with show sniffs",
			options: PHPCSOptions{
				ShowSniffs: true,
			},
			contains: []string{"-s"},
		},
		{
			name: "with severity",
			options: PHPCSOptions{
				Severity: 5,
			},
			contains: []string{"--severity=5"},
		},
		{
			name: "with error severity",
			options: PHPCSOptions{
				ErrorSeverity: 3,
			},
			contains: []string{"--error-severity=3"},
		},
		{
			name: "with warning severity",
			options: PHPCSOptions{
				WarningSeverity: 7,
			},
			contains: []string{"--warning-severity=7"},
		},
		{
			name: "with specific files",
			options: PHPCSOptions{
				Files: []string{"src/", "tests/"},
			},
			contains: []string{"src/", "tests/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := quality.buildPHPCSArgs(tt.options)

			for _, expected := range tt.contains {
				found := false
				for _, arg := range args {
					if arg == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected args to contain %s, got %v", expected, args)
				}
			}
		})
	}
}

func TestQuality_parsePHPCSOutput_JSON(t *testing.T) {
	quality := NewQuality("/tmp/test")

	jsonOutput := `{
		"totals": {
			"errors": 5,
			"warnings": 3,
			"fixable": 2
		},
		"files": {
			"/path/to/file.php": {
				"errors": 3,
				"warnings": 2,
				"messages": [
					{
						"line": 10,
						"column": 1,
						"type": "ERROR",
						"message": "Missing doc comment",
						"source": "Squiz.Commenting.FileComment.Missing",
						"severity": 5,
						"fixable": false
					},
					{
						"line": 15,
						"column": 5,
						"type": "WARNING",
						"message": "Line exceeds 120 characters",
						"source": "Generic.Files.LineLength.TooLong",
						"severity": 5,
						"fixable": true
					}
				]
			}
		}
	}`

	result := quality.parsePHPCSOutput(jsonOutput, "json")

	if result.Errors != 5 {
		t.Errorf("expected 5 errors, got %d", result.Errors)
	}

	if result.Warnings != 3 {
		t.Errorf("expected 3 warnings, got %d", result.Warnings)
	}

	if result.Fixable != 2 {
		t.Errorf("expected 2 fixable, got %d", result.Fixable)
	}

	if len(result.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(result.Files))
	}

	if len(result.Files) > 0 {
		file := result.Files[0]
		if file.Errors != 3 {
			t.Errorf("expected 3 file errors, got %d", file.Errors)
		}
		if len(file.Messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(file.Messages))
		}
	}
}

func TestQuality_parsePHPCSOutput_Text(t *testing.T) {
	quality := NewQuality("/tmp/test")

	textOutput := `
FILE: /path/to/file.php
----------------------------------------------------------------------
FOUND 5 ERRORS AND 3 WARNINGS AFFECTING 4 LINES
----------------------------------------------------------------------
 10 | ERROR   | Missing doc comment
 15 | WARNING | Line exceeds 120 characters
----------------------------------------------------------------------
2 FIXABLE
`

	result := quality.parsePHPCSOutput(textOutput, "full")

	if result.Errors != 5 {
		t.Errorf("expected 5 errors, got %d", result.Errors)
	}

	if result.Warnings != 3 {
		t.Errorf("expected 3 warnings, got %d", result.Warnings)
	}

	if result.Fixable != 2 {
		t.Errorf("expected 2 fixable, got %d", result.Fixable)
	}
}

func TestQuality_FormatPHPCSResults_Success(t *testing.T) {
	quality := NewQuality("/tmp/test")

	result := &PHPCSResult{
		Success: true,
	}

	output := quality.FormatPHPCSResults(result)
	expected := "No errors or warnings found"

	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestQuality_FormatPHPCSResults_WithErrors(t *testing.T) {
	quality := NewQuality("/tmp/test")

	result := &PHPCSResult{
		Success:  false,
		Errors:   3,
		Warnings: 2,
		Fixable:  1,
		Files: []PHPCSFileResult{
			{
				File:     "/path/to/file.php",
				Errors:   2,
				Warnings: 1,
				Messages: []PHPCSMessage{
					{
						Line:    10,
						Column:  1,
						Type:    "ERROR",
						Message: "Missing doc comment",
						Source:  "Squiz.Commenting.FileComment.Missing",
						Fixable: false,
					},
					{
						Line:    15,
						Column:  5,
						Type:    "WARNING",
						Message: "Line too long",
						Fixable: true,
					},
				},
			},
		},
	}

	output := quality.FormatPHPCSResults(result)

	// Check that output contains key information
	if !strings.Contains(output, "3 error(s)") {
		t.Errorf("expected output to contain error count, got: %s", output)
	}
	if !strings.Contains(output, "2 warning(s)") {
		t.Errorf("expected output to contain warning count, got: %s", output)
	}
	if !strings.Contains(output, "1 fixable") {
		t.Errorf("expected output to contain fixable count, got: %s", output)
	}
	if !strings.Contains(output, "/path/to/file.php") {
		t.Errorf("expected output to contain file path, got: %s", output)
	}
}

func TestQuality_findPHPCS_Vendor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create vendor/bin/phpcs
	vendorBin := filepath.Join(tmpDir, "vendor", "bin")
	if err := os.MkdirAll(vendorBin, 0755); err != nil {
		t.Fatalf("failed to create vendor/bin: %v", err)
	}

	phpcsPath := filepath.Join(vendorBin, "phpcs")
	if err := os.WriteFile(phpcsPath, []byte("#!/bin/bash\necho 'phpcs'"), 0755); err != nil {
		t.Fatalf("failed to create phpcs: %v", err)
	}

	quality := NewQuality(tmpDir)
	result, err := quality.findPHPCS()

	if err != nil {
		t.Fatalf("findPHPCS failed: %v", err)
	}

	if result != phpcsPath {
		t.Errorf("expected %s, got %s", phpcsPath, result)
	}
}

func TestQuality_findPHPCBF_Vendor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create vendor/bin/phpcbf
	vendorBin := filepath.Join(tmpDir, "vendor", "bin")
	if err := os.MkdirAll(vendorBin, 0755); err != nil {
		t.Fatalf("failed to create vendor/bin: %v", err)
	}

	phpcbfPath := filepath.Join(vendorBin, "phpcbf")
	if err := os.WriteFile(phpcbfPath, []byte("#!/bin/bash\necho 'phpcbf'"), 0755); err != nil {
		t.Fatalf("failed to create phpcbf: %v", err)
	}

	quality := NewQuality(tmpDir)
	result, err := quality.findPHPCBF()

	if err != nil {
		t.Fatalf("findPHPCBF failed: %v", err)
	}

	if result != phpcbfPath {
		t.Errorf("expected %s, got %s", phpcbfPath, result)
	}
}
