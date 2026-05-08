package wpengine

import (
	"strings"
	"testing"
)

func TestBuildExportCommand_ExtraFlags(t *testing.T) {
	cmd, err := buildExportCommand("wp_", DatabaseOptions{
		ExtraFlags: []string{"--hex-blob", "--quote-names", "--default-character-set=utf8mb4"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--hex-blob", "--quote-names", "--default-character-set=utf8mb4"} {
		if !strings.Contains(cmd, flag) {
			t.Errorf("expected command to contain %q, got: %s", flag, cmd)
		}
	}
}

func TestBuildExportCommand_NoExtraFlags(t *testing.T) {
	cmd, err := buildExportCommand("wp_", DatabaseOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cmd, "--add-drop-table") {
		t.Errorf("expected command to contain --add-drop-table, got: %s", cmd)
	}
	if !strings.HasSuffix(cmd, " -") {
		t.Errorf("expected command to end with ' -' (stdout), got: %s", cmd)
	}
}
