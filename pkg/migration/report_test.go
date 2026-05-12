package migration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/firecrown-media/stax/pkg/migration"
)

func TestParsePluginHeader(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "my-plugin")
	_ = os.MkdirAll(pluginDir, 0755)
	content := "<?php\n/**\n * Plugin Name: My Plugin\n * Version: 1.2.3\n */\n"
	_ = os.WriteFile(filepath.Join(pluginDir, "my-plugin.php"), []byte(content), 0644)

	name, version := migration.ParsePluginHeader(pluginDir)
	if name != "My Plugin" {
		t.Errorf("expected 'My Plugin', got %q", name)
	}
	if version != "1.2.3" {
		t.Errorf("expected '1.2.3', got %q", version)
	}
}

func TestParsePluginHeader_FallbackToDirName(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "advanced-custom-fields")
	_ = os.MkdirAll(pluginDir, 0755)

	name, version := migration.ParsePluginHeader(pluginDir)
	if name != "Advanced Custom Fields" {
		t.Errorf("expected 'Advanced Custom Fields', got %q", name)
	}
	if version != "unknown" {
		t.Errorf("expected 'unknown', got %q", version)
	}
}

func TestParseThemeHeader(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "my-theme")
	_ = os.MkdirAll(themeDir, 0755)
	content := "/*\nTheme Name: My Theme\nVersion: 2.0.0\n*/\n"
	_ = os.WriteFile(filepath.Join(themeDir, "style.css"), []byte(content), 0644)

	name, version := migration.ParseThemeHeader(themeDir)
	if name != "My Theme" {
		t.Errorf("expected 'My Theme', got %q", name)
	}
	if version != "2.0.0" {
		t.Errorf("expected '2.0.0', got %q", version)
	}
}

func TestDetectWPEMUPlugins(t *testing.T) {
	dir := t.TempDir()
	muPluginsDir := filepath.Join(dir, "mu-plugins")
	_ = os.MkdirAll(filepath.Join(muPluginsDir, "wpe-cache-plugin"), 0755)
	_ = os.MkdirAll(filepath.Join(muPluginsDir, "force-strong-passwords"), 0755)
	_ = os.MkdirAll(filepath.Join(muPluginsDir, "my-custom-mu-plugin"), 0755)

	detected := migration.DetectWPEMUPlugins(dir)
	if len(detected) != 2 {
		t.Fatalf("expected 2 WPE plugins, got %d: %v", len(detected), detected)
	}
	names := map[string]bool{detected[0].Name: true, detected[1].Name: true}
	if !names["wpe-cache-plugin"] || !names["force-strong-passwords"] {
		t.Errorf("unexpected detected plugins: %v", detected)
	}
}

func TestBuildPluginResults(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	_ = os.MkdirAll(filepath.Join(pluginsDir, "clean-plugin"), 0755)
	_ = os.WriteFile(filepath.Join(pluginsDir, "clean-plugin", "clean-plugin.php"),
		[]byte("<?php\n/**\n * Plugin Name: Clean Plugin\n * Version: 1.0\n */\n"), 0644)
	_ = os.MkdirAll(filepath.Join(pluginsDir, "bad-plugin"), 0755)

	audit := &migration.AuditReport{
		Files: []migration.FileAuditResult{
			{
				FilePath: filepath.Join(pluginsDir, "bad-plugin", "bad.php"),
				Errors:   1,
				Messages: []migration.AuditMessage{
					{Type: "ERROR", Message: "Direct database query", Severity: 5},
				},
			},
		},
		TotalErrors: 1,
	}

	results := migration.BuildPluginResults(audit, dir)
	if len(results) != 2 {
		t.Fatalf("expected 2 plugin results, got %d", len(results))
	}
	byName := map[string]migration.PluginCompatResult{}
	for _, r := range results {
		byName[r.Name] = r
	}
	if !byName["Clean Plugin"].VIPCompatible {
		t.Error("clean-plugin should be VIP compatible")
	}
	if byName["Bad Plugin"].VIPCompatible {
		t.Error("bad-plugin should not be VIP compatible")
	}
	if byName["Bad Plugin"].Issues == "" {
		t.Error("bad-plugin should have issues listed")
	}
}
