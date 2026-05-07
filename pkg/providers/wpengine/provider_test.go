package wpengine

import (
	"testing"
)

func TestDecodeProviderConfig(t *testing.T) {
	raw := map[string]any{
		"install":     "mysite",
		"environment": "production",
		"ssh_gateway": "ssh.wpengine.net",
	}

	cfg, err := decodeProviderConfig(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Install != "mysite" {
		t.Errorf("expected install 'mysite', got %q", cfg.Install)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected environment 'production', got %q", cfg.Environment)
	}
	if cfg.SSHGateway != "ssh.wpengine.net" {
		t.Errorf("expected ssh_gateway 'ssh.wpengine.net', got %q", cfg.SSHGateway)
	}
}

func TestDecodeProviderConfig_MissingInstall(t *testing.T) {
	raw := map[string]any{
		"environment": "production",
	}
	_, err := decodeProviderConfig(raw)
	if err == nil {
		t.Error("expected error for missing install, got nil")
	}
}
