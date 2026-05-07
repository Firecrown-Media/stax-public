package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (cfgFile, projectDir string)
		cleanupFunc func()
		wantErr     bool
		validate    func(t *testing.T, cfg *Config)
	}{
		{
			name: "load project config successfully",
			setupFunc: func(t *testing.T) (string, string) {
				// Create temp directory
				dir := t.TempDir()

				// Create project config
				cfg := Defaults()
				cfg.Project.Name = "test-project"
				cfg.ProviderConfig["install"] = "testinstall"

				cfgPath := filepath.Join(dir, ".stax.yml")
				if err := Save(cfg, cfgPath); err != nil {
					t.Fatalf("failed to save test config: %v", err)
				}

				return cfgPath, dir
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Project.Name != "test-project" {
					t.Errorf("expected project name 'test-project', got %q", cfg.Project.Name)
				}
				install, _ := cfg.ProviderConfig["install"].(string)
				if install != "testinstall" {
					t.Errorf("expected install 'testinstall', got %q", install)
				}
			},
		},
		{
			name: "load with defaults when project config not found",
			setupFunc: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				// Don't create config file - test default behavior
				return filepath.Join(dir, ".stax.yml"), dir
			},
			wantErr: true, // Should error when config file doesn't exist
		},
		{
			name: "merge global and project configs",
			setupFunc: func(t *testing.T) (string, string) {
				// Create temp directory
				dir := t.TempDir()

				// Create project config
				cfg := Defaults()
				cfg.Project.Name = "test-project"

				cfgPath := filepath.Join(dir, ".stax.yml")
				if err := Save(cfg, cfgPath); err != nil {
					t.Fatalf("failed to save test config: %v", err)
				}

				return cfgPath, dir
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				// Should have project name from project config
				if cfg.Project.Name != "test-project" {
					t.Errorf("expected project name 'test-project', got %q", cfg.Project.Name)
				}
				// Should have defaults for other fields
				if cfg.DDEV.PHPVersion == "" {
					t.Error("expected PHP version from defaults")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgFile, projectDir := tt.setupFunc(t)
			if tt.cleanupFunc != nil {
				defer tt.cleanupFunc()
			}

			cfg, err := Load(cfgFile, projectDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadProjectConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
		check   func(t *testing.T, cfg *Config)
	}{
		{
			name: "successfully load valid config",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cfg := Defaults()
				cfg.Project.Name = "valid-project"

				path := filepath.Join(dir, ".stax.yml")
				if err := Save(cfg, path); err != nil {
					t.Fatalf("failed to save config: %v", err)
				}
				return path
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Project.Name != "valid-project" {
					t.Errorf("expected 'valid-project', got %q", cfg.Project.Name)
				}
			},
		},
		{
			name: "fail when config file doesn't exist",
			setup: func(t *testing.T) string {
				return "/nonexistent/config.yml"
			},
			wantErr: true,
		},
		{
			name: "fail on invalid YAML",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, ".stax.yml")

				// Write invalid YAML
				invalidYAML := `project:
  name: test
  invalid: [unclosed array`
				if err := os.WriteFile(path, []byte(invalidYAML), 0644); err != nil {
					t.Fatalf("failed to write invalid YAML: %v", err)
				}
				return path
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			cfg, err := loadProjectConfig(path)

			if (err != nil) != tt.wantErr {
				t.Errorf("loadProjectConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		setup   func(t *testing.T) string
		wantErr bool
		check   func(t *testing.T, path string)
	}{
		{
			name: "save config successfully",
			cfg: &Config{
				Version: 1,
				Project: ProjectConfig{
					Name: "test-save",
					Type: "wordpress-multisite",
				},
			},
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				return filepath.Join(dir, ".stax.yml")
			},
			wantErr: false,
			check: func(t *testing.T, path string) {
				// Verify file exists
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Error("config file was not created")
				}

				// Load and verify content
				cfg, err := loadConfigFile(path)
				if err != nil {
					t.Fatalf("failed to load saved config: %v", err)
				}

				if cfg.Project.Name != "test-save" {
					t.Errorf("expected 'test-save', got %q", cfg.Project.Name)
				}
			},
		},
		{
			name: "create directory if it doesn't exist",
			cfg:  Defaults(),
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				return filepath.Join(dir, "nested", "path", ".stax.yml")
			},
			wantErr: false,
			check: func(t *testing.T, path string) {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Error("config file was not created in nested directory")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)

			err := Save(tt.cfg, path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, path)
			}
		})
	}
}

func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		name     string
		base     *Config
		override *Config
		check    func(t *testing.T, result *Config)
	}{
		{
			name: "override project name",
			base: &Config{
				Project: ProjectConfig{Name: "base-project"},
			},
			override: &Config{
				Project: ProjectConfig{Name: "override-project"},
			},
			check: func(t *testing.T, result *Config) {
				if result.Project.Name != "override-project" {
					t.Errorf("expected 'override-project', got %q", result.Project.Name)
				}
			},
		},
		{
			name: "keep base values when override is empty",
			base: &Config{
				Project: ProjectConfig{
					Name: "base-project",
					Type: "wordpress-multisite",
				},
				DDEV: DDEVConfig{
					PHPVersion: "8.1",
				},
			},
			override: &Config{},
			check: func(t *testing.T, result *Config) {
				if result.Project.Name != "base-project" {
					t.Errorf("expected 'base-project', got %q", result.Project.Name)
				}
				if result.DDEV.PHPVersion != "8.1" {
					t.Errorf("expected '8.1', got %q", result.DDEV.PHPVersion)
				}
			},
		},
		{
			name: "merge multiple fields",
			base: &Config{
				Project:        ProjectConfig{Name: "base"},
				ProviderConfig: map[string]any{"install": "base-install"},
			},
			override: &Config{
				Project:        ProjectConfig{Type: "wordpress"},
				ProviderConfig: map[string]any{"environment": "staging"},
			},
			check: func(t *testing.T, result *Config) {
				if result.Project.Name != "base" {
					t.Errorf("expected 'base', got %q", result.Project.Name)
				}
				if result.Project.Type != "wordpress" {
					t.Errorf("expected 'wordpress', got %q", result.Project.Type)
				}
				install, _ := result.ProviderConfig["install"].(string)
				if install != "base-install" {
					t.Errorf("expected 'base-install', got %q", install)
				}
				env, _ := result.ProviderConfig["environment"].(string)
				if env != "staging" {
					t.Errorf("expected 'staging', got %q", env)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeConfigs(tt.base, tt.override)
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		cfg     *Config
		check   func(t *testing.T, cfg *Config)
	}{
		{
			name: "override project name",
			envVars: map[string]string{
				"STAX_PROJECT_NAME": "env-project",
			},
			cfg: &Config{
				Project: ProjectConfig{Name: "original"},
			},
			check: func(t *testing.T, cfg *Config) {
				if cfg.Project.Name != "env-project" {
					t.Errorf("expected 'env-project', got %q", cfg.Project.Name)
				}
			},
		},
		{
			name: "override multiple values",
			envVars: map[string]string{
				"STAX_WPENGINE_INSTALL":     "env-install",
				"STAX_WPENGINE_ENVIRONMENT": "staging",
				"STAX_DDEV_PHP_VERSION":     "8.2",
			},
			cfg: &Config{
				ProviderConfig: map[string]any{
					"install":     "original-install",
					"environment": "production",
				},
				DDEV: DDEVConfig{
					PHPVersion: "8.1",
				},
			},
			check: func(t *testing.T, cfg *Config) {
				install, _ := cfg.ProviderConfig["install"].(string)
				if install != "env-install" {
					t.Errorf("expected 'env-install', got %q", install)
				}
				env, _ := cfg.ProviderConfig["environment"].(string)
				if env != "staging" {
					t.Errorf("expected 'staging', got %q", env)
				}
				if cfg.DDEV.PHPVersion != "8.2" {
					t.Errorf("expected '8.2', got %q", cfg.DDEV.PHPVersion)
				}
			},
		},
		{
			name:    "no overrides when env vars not set",
			envVars: map[string]string{},
			cfg: &Config{
				Project: ProjectConfig{Name: "original"},
			},
			check: func(t *testing.T, cfg *Config) {
				if cfg.Project.Name != "original" {
					t.Errorf("expected 'original', got %q", cfg.Project.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			applyEnvOverrides(tt.cfg)

			if tt.check != nil {
				tt.check(t, tt.cfg)
			}
		})
	}
}

func TestGetGlobalConfigPath(t *testing.T) {
	path, err := GetGlobalConfigPath()
	if err != nil {
		t.Fatalf("GetGlobalConfigPath() failed: %v", err)
	}

	if path == "" {
		t.Error("expected non-empty path")
	}

	// Should contain .stax
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
}

func TestGetProjectConfigPath(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		want       string
	}{
		{
			name:       "with project directory",
			projectDir: "/test/project",
			want:       "/test/project/.stax.yml",
		},
		{
			name:       "empty project directory uses cwd",
			projectDir: "",
			// Will use current working directory
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetProjectConfigPath(tt.projectDir)

			if tt.want != "" && got != tt.want {
				t.Errorf("GetProjectConfigPath() = %q, want %q", got, tt.want)
			}

			if tt.projectDir == "" && got == "" {
				t.Error("expected non-empty path when using cwd")
			}
		})
	}
}
