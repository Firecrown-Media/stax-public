package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectConfigVersion(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
		wantErr  bool
	}{
		{
			name: "versioned config",
			content: `version: 1
project:
  name: test`,
			expected: 1,
			wantErr:  false,
		},
		{
			name: "unversioned config",
			content: `project:
  name: test`,
			expected: 0, // Unversioned configs return 0
			wantErr:  false,
		},
		{
			name: "invalid yaml",
			content: `invalid: [yaml
content`,
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".stax.yml")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Test detection
			version, err := DetectConfigVersion(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectConfigVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version.Version != tt.expected {
				t.Errorf("DetectConfigVersion() version = %v, expected %v", version.Version, tt.expected)
			}
		})
	}
}

func TestNeedsMigration(t *testing.T) {
	tests := []struct {
		name     string
		version  int
		expected bool
	}{
		{
			name:     "current version",
			version:  CurrentVersion,
			expected: false,
		},
		{
			name:     "old version",
			version:  0,
			expected: true,
		},
		{
			name:     "version 1",
			version:  1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Version: tt.version}
			needs, _ := NeedsMigration(cfg)
			if needs != tt.expected {
				t.Errorf("NeedsMigration() = %v, expected %v", needs, tt.expected)
			}
		})
	}
}

func TestBackupConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".stax.yml")

	// Create test config
	content := []byte(`version: 1
project:
  name: test`)
	err := os.WriteFile(configPath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test backup
	err = BackupConfig(configPath)
	if err != nil {
		t.Fatalf("BackupConfig() error = %v", err)
	}

	// Verify backup was created
	backups, err := ListBackups(configPath)
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}

	if len(backups) != 1 {
		t.Errorf("Expected 1 backup, got %d", len(backups))
	}

	// Verify backup content
	if len(backups) > 0 {
		backupContent, err := os.ReadFile(backups[0])
		if err != nil {
			t.Fatalf("Failed to read backup: %v", err)
		}

		if string(backupContent) != string(content) {
			t.Errorf("Backup content doesn't match original")
		}
	}
}

func TestMigrateV0ToV1(t *testing.T) {
	// Create minimal v0 config
	cfg := &Config{
		Version: 0,
		Project: ProjectConfig{
			Name: "test",
		},
		ProviderConfig: map[string]any{
			"install": "testinstall",
		},
	}

	// Migrate
	migrated := migrateV0ToV1(cfg)

	// Verify version was updated
	if migrated.Version != 1 {
		t.Errorf("Expected version 1, got %d", migrated.Version)
	}

	// Verify defaults were set
	if migrated.Project.Type == "" {
		t.Error("Expected project.type to be set")
	}
	if migrated.DDEV.PHPVersion == "" {
		t.Error("Expected ddev.php_version to be set")
	}
	if migrated.WordPress.Version == "" {
		t.Error("Expected wordpress.version to be set")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Project: ProjectConfig{
					Name: "test",
					Type: "wordpress-multisite",
					Mode: "subdomain",
				},
				ProviderConfig: map[string]any{
					"install":     "testinstall",
					"environment": "production",
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			cfg: &Config{
				Project: ProjectConfig{
					Type: "wordpress",
				},
				ProviderConfig: map[string]any{
					"install":     "testinstall",
					"environment": "production",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid project type",
			cfg: &Config{
				Project: ProjectConfig{
					Name: "test",
					Type: "invalid",
				},
				ProviderConfig: map[string]any{
					"install":     "testinstall",
					"environment": "production",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			cfg: &Config{
				Project: ProjectConfig{
					Name: "test",
					Type: "wordpress",
				},
				ProviderConfig: map[string]any{
					"install":     "testinstall",
					"environment": "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMigrateConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".stax.yml")

	// Create v0 config (unversioned)
	content := `project:
  name: test-project
wpengine:
  install: testinstall
  environment: production`

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test dry-run first
	plan, err := MigrateConfig(configPath, true)
	if err != nil {
		t.Fatalf("MigrateConfig() dry-run error = %v", err)
	}

	if plan.ToVersion != CurrentVersion {
		t.Errorf("Expected migration to version %d, got %d", CurrentVersion, plan.ToVersion)
	}

	// Verify file wasn't modified
	afterDryRun, _ := os.ReadFile(configPath)
	if string(afterDryRun) != content {
		t.Error("Dry-run modified the config file")
	}

	// Test actual migration
	_, err = MigrateConfig(configPath, false)
	if err != nil {
		t.Fatalf("MigrateConfig() error = %v", err)
	}

	// Verify backup was created
	backups, err := ListBackups(configPath)
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}
	if len(backups) == 0 {
		t.Error("No backup was created")
	}

	// Verify config was updated
	version, err := DetectConfigVersion(configPath)
	if err != nil {
		t.Fatalf("DetectConfigVersion() error = %v", err)
	}
	if version.Version != CurrentVersion {
		t.Errorf("Config version = %d, expected %d", version.Version, CurrentVersion)
	}

	// Verify migrated config is valid
	cfg, err := loadConfigFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load migrated config: %v", err)
	}

	err = ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Migrated config is invalid: %v", err)
	}
}
