package ddev

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		name    string
		options ConfigOptions
		want    *DDEVConfig
		wantErr bool
	}{
		{
			name: "default configuration with DNS enabled",
			options: ConfigOptions{
				ProjectName:        "testproject",
				Type:               "wordpress",
				PHPVersion:         "8.1",
				DatabaseType:       "mysql",
				DatabaseVersion:    "8.0",
				UseDNSWhenPossible: true,
			},
			want: &DDEVConfig{
				Name:               "testproject",
				Type:               "wordpress",
				DocRoot:            ".",
				PHPVersion:         "8.1",
				UseDNSWhenPossible: true,
				Database: DatabaseConfig{
					Type:    "mysql",
					Version: "8.0",
				},
			},
			wantErr: false,
		},
		{
			name: "DNS disabled when explicitly set to false",
			options: ConfigOptions{
				ProjectName:        "testproject",
				Type:               "wordpress",
				PHPVersion:         "8.2",
				DatabaseType:       "mariadb",
				DatabaseVersion:    "10.6",
				UseDNSWhenPossible: false,
			},
			want: &DDEVConfig{
				Name:               "testproject",
				Type:               "wordpress",
				DocRoot:            ".",
				PHPVersion:         "8.2",
				UseDNSWhenPossible: false,
				Database: DatabaseConfig{
					Type:    "mariadb",
					Version: "10.6",
				},
			},
			wantErr: false,
		},
		{
			name: "multisite with DNS enabled",
			options: ConfigOptions{
				ProjectName:         "multisite",
				Type:                "wordpress",
				PHPVersion:          "8.1",
				DatabaseType:        "mysql",
				DatabaseVersion:     "8.0",
				UseDNSWhenPossible:  true,
				AdditionalHostnames: []string{"site1", "site2"},
			},
			want: &DDEVConfig{
				Name:                "multisite",
				Type:                "wordpress",
				DocRoot:             ".",
				PHPVersion:          "8.1",
				UseDNSWhenPossible:  true,
				AdditionalHostnames: []string{"site1", "site2"},
				Database: DatabaseConfig{
					Type:    "mysql",
					Version: "8.0",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			got, err := GenerateConfig(tmpDir, tt.options)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil && tt.want != nil {
				t.Error("GenerateConfig() returned nil config")
				return
			}

			// Check key fields
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.PHPVersion != tt.want.PHPVersion {
				t.Errorf("PHPVersion = %v, want %v", got.PHPVersion, tt.want.PHPVersion)
			}
			if got.UseDNSWhenPossible != tt.want.UseDNSWhenPossible {
				t.Errorf("UseDNSWhenPossible = %v, want %v", got.UseDNSWhenPossible, tt.want.UseDNSWhenPossible)
			}
			if got.Database.Type != tt.want.Database.Type {
				t.Errorf("Database.Type = %v, want %v", got.Database.Type, tt.want.Database.Type)
			}
			if got.Database.Version != tt.want.Database.Version {
				t.Errorf("Database.Version = %v, want %v", got.Database.Version, tt.want.Database.Version)
			}

			// Check additional hostnames if specified
			if len(tt.want.AdditionalHostnames) > 0 {
				if len(got.AdditionalHostnames) != len(tt.want.AdditionalHostnames) {
					t.Errorf("AdditionalHostnames length = %v, want %v", len(got.AdditionalHostnames), len(tt.want.AdditionalHostnames))
				}
			}
		})
	}
}

func TestWriteAndReadConfig(t *testing.T) {
	tmpDir := t.TempDir()

	original := &DDEVConfig{
		Name:               "testproject",
		Type:               "wordpress",
		DocRoot:            ".",
		PHPVersion:         "8.1",
		UseDNSWhenPossible: true,
		Database: DatabaseConfig{
			Type:    "mysql",
			Version: "8.0",
		},
		RouterHTTPPort:  "80",
		RouterHTTPSPort: "443",
		ComposerVersion: "2",
	}

	// Write config
	err := WriteConfig(tmpDir, original)
	if err != nil {
		t.Fatalf("WriteConfig() failed: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".ddev", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Read config back
	got, err := ReadConfig(tmpDir)
	if err != nil {
		t.Fatalf("ReadConfig() failed: %v", err)
	}

	// Verify UseDNSWhenPossible is preserved
	if got.UseDNSWhenPossible != original.UseDNSWhenPossible {
		t.Errorf("UseDNSWhenPossible = %v, want %v", got.UseDNSWhenPossible, original.UseDNSWhenPossible)
	}

	// Verify other key fields
	if got.Name != original.Name {
		t.Errorf("Name = %v, want %v", got.Name, original.Name)
	}
	if got.PHPVersion != original.PHPVersion {
		t.Errorf("PHPVersion = %v, want %v", got.PHPVersion, original.PHPVersion)
	}
}

func TestUseDNSWhenPossibleYAMLMarshaling(t *testing.T) {
	config := &DDEVConfig{
		Name:               "test",
		Type:               "wordpress",
		DocRoot:            ".",
		PHPVersion:         "8.1",
		UseDNSWhenPossible: true,
		Database: DatabaseConfig{
			Type:    "mysql",
			Version: "8.0",
		},
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("yaml.Marshal() failed: %v", err)
	}

	// Verify the YAML contains use_dns_when_possible: true
	yamlStr := string(data)
	if !contains(yamlStr, "use_dns_when_possible: true") {
		t.Errorf("YAML does not contain 'use_dns_when_possible: true'\nGot:\n%s", yamlStr)
	}

	// Unmarshal and verify
	var unmarshaled DDEVConfig
	if err := yaml.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("yaml.Unmarshal() failed: %v", err)
	}

	if !unmarshaled.UseDNSWhenPossible {
		t.Error("UseDNSWhenPossible was not unmarshaled correctly")
	}
}

func TestGetDefaultConfigOptions(t *testing.T) {
	opts := GetDefaultConfigOptions("testproject")

	if opts.ProjectName != "testproject" {
		t.Errorf("ProjectName = %v, want testproject", opts.ProjectName)
	}

	if !opts.UseDNSWhenPossible {
		t.Error("Expected UseDNSWhenPossible to be true by default")
	}

	if opts.PHPVersion != "8.1" {
		t.Errorf("PHPVersion = %v, want 8.1", opts.PHPVersion)
	}

	if opts.Type != "wordpress" {
		t.Errorf("Type = %v, want wordpress", opts.Type)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *DDEVConfig
		wantErr bool
	}{
		{
			name: "valid config with DNS enabled",
			config: &DDEVConfig{
				Name:               "testproject",
				Type:               "wordpress",
				PHPVersion:         "8.1",
				UseDNSWhenPossible: true,
				Database: DatabaseConfig{
					Type:    "mysql",
					Version: "8.0",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with DNS disabled",
			config: &DDEVConfig{
				Name:               "testproject",
				Type:               "wordpress",
				PHPVersion:         "8.1",
				UseDNSWhenPossible: false,
				Database: DatabaseConfig{
					Type:    "mysql",
					Version: "8.0",
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			config: &DDEVConfig{
				Type:               "wordpress",
				PHPVersion:         "8.1",
				UseDNSWhenPossible: true,
				Database: DatabaseConfig{
					Type:    "mysql",
					Version: "8.0",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid PHP version",
			config: &DDEVConfig{
				Name:               "test",
				Type:               "wordpress",
				PHPVersion:         "8",
				UseDNSWhenPossible: true,
				Database: DatabaseConfig{
					Type:    "mysql",
					Version: "8.0",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Should not exist initially
	if ConfigExists(tmpDir) {
		t.Error("ConfigExists() should return false for non-existent config")
	}

	// Create config
	config := &DDEVConfig{
		Name:               "test",
		Type:               "wordpress",
		DocRoot:            ".",
		PHPVersion:         "8.1",
		UseDNSWhenPossible: true,
		Database: DatabaseConfig{
			Type:    "mysql",
			Version: "8.0",
		},
	}

	if err := WriteConfig(tmpDir, config); err != nil {
		t.Fatalf("WriteConfig() failed: %v", err)
	}

	// Should exist now
	if !ConfigExists(tmpDir) {
		t.Error("ConfigExists() should return true after config is written")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
