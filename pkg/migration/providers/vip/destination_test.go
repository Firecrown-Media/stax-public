package vip_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/firecrown-media/stax/pkg/migration"
	vipdest "github.com/firecrown-media/stax/pkg/migration/providers/vip"
)

func TestVIPDestination_Audit_PHPCSNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dest := vipdest.NewVIPDestination("")
	_, err := dest.Audit("/some/path", migration.AuditOptions{})
	if err == nil {
		t.Fatal("expected error when phpcs not in PATH")
	}
	if !strings.Contains(err.Error(), "phpcs not found") {
		t.Errorf("expected 'phpcs not found' in error, got: %v", err)
	}
}

func TestVIPDestination_ValidateDatabase_VIPCLINotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dest := vipdest.NewVIPDestination("")
	err := dest.ValidateDatabase("/some/file.sql")
	if err == nil {
		t.Fatal("expected error when vip CLI not in PATH")
	}
	if !strings.Contains(err.Error(), "VIP CLI not found") {
		t.Errorf("expected 'VIP CLI not found' in error, got: %v", err)
	}
}

func TestVIPDestination_ImportDatabase_VIPCLINotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dest := vipdest.NewVIPDestination("")
	err := dest.ImportDatabase("/some/file.sql", migration.ImportOptions{})
	if err == nil {
		t.Fatal("expected error when vip CLI not in PATH")
	}
	if !strings.Contains(err.Error(), "VIP CLI not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVIPDestination_ImportMedia_VIPCLINotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dest := vipdest.NewVIPDestination("")
	err := dest.ImportMedia(migration.ImportOptions{})
	if err == nil {
		t.Fatal("expected error when vip CLI not in PATH")
	}
	if !strings.Contains(err.Error(), "VIP CLI not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVIPDestination_Audit_PHPCSStderrSurfacedOnFatal(t *testing.T) {
	// Simulate phpcs crashing during startup (missing standard, memory
	// exhaustion, autoload failure) — it writes a fatal to stderr and exits
	// non-zero with empty stdout. The previous implementation used
	// exec.Cmd.Output() which dropped stderr, swallowed the ExitError, and
	// silently produced an empty audit report. This test asserts the stderr
	// is surfaced as an error.
	tmpBin := t.TempDir()
	phpcsPath := filepath.Join(tmpBin, "phpcs")
	script := "#!/bin/sh\necho 'PHP Fatal error: Class WordPressCS\\\\WordPress\\\\Sniff not found' >&2\nexit 2\n"
	if err := os.WriteFile(phpcsPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmpBin)

	wpContent := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wpContent, "plugins"), 0755); err != nil {
		t.Fatal(err)
	}

	dest := vipdest.NewVIPDestination("")
	_, err := dest.Audit(wpContent, migration.AuditOptions{})
	if err == nil {
		t.Fatal("expected error when phpcs writes fatal to stderr with empty stdout")
	}
	if !strings.Contains(err.Error(), "PHP Fatal") {
		t.Errorf("expected phpcs stderr in error, got: %v", err)
	}
}

func TestVIPDestination_Audit_ParsesValidJSONDespiteNonZeroExit(t *testing.T) {
	// Real-world phpcs exits non-zero when it finds violations but still
	// produces valid JSON on stdout. The audit must parse it and not treat
	// the non-zero exit as a failure.
	tmpBin := t.TempDir()
	phpcsPath := filepath.Join(tmpBin, "phpcs")
	// Emit valid phpcs JSON via printf (a /bin/sh builtin) so the script does
	// not depend on $PATH, which this test clears.
	script := `#!/bin/sh
printf '%s\n' '{"totals":{"errors":1,"warnings":0,"fixable":0},"files":{"/tmp/foo.php":{"errors":1,"warnings":0,"messages":[{"message":"x","source":"y","severity":5,"type":"ERROR","line":1,"column":1,"fixable":false}]}}}'
exit 2
`
	if err := os.WriteFile(phpcsPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmpBin)

	wpContent := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wpContent, "plugins"), 0755); err != nil {
		t.Fatal(err)
	}

	dest := vipdest.NewVIPDestination("")
	report, err := dest.Audit(wpContent, migration.AuditOptions{})
	if err != nil {
		t.Fatalf("unexpected error parsing valid JSON despite non-zero phpcs exit: %v", err)
	}
	if report.TotalErrors != 1 {
		t.Errorf("expected 1 error in report, got %d", report.TotalErrors)
	}
}

func TestVIPDestination_CompareFiles(t *testing.T) {
	// Build a fake WPEngine wp-content and VIP repo layout.
	wpeDir := t.TempDir()
	vipDir := t.TempDir()

	// Both have plugins/hello-world; only WPE has plugins/wpe-only; only VIP has plugins/vip-only.
	if err := os.MkdirAll(filepath.Join(wpeDir, "plugins", "hello-world"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(wpeDir, "plugins", "wpe-only"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(vipDir, "plugins", "hello-world"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(vipDir, "plugins", "vip-only"), 0755); err != nil {
		t.Fatal(err)
	}

	dest := vipdest.NewVIPDestination(vipDir)
	result, err := dest.CompareFiles(wpeDir)
	if err != nil {
		t.Fatalf("CompareFiles returned error: %v", err)
	}

	if !containsItem(result.MissingFromVIP, "plugins/wpe-only") {
		t.Errorf("expected plugins/wpe-only in MissingFromVIP, got %v", result.MissingFromVIP)
	}
	if !containsItem(result.MissingFromWPE, "plugins/vip-only") {
		t.Errorf("expected plugins/vip-only in MissingFromWPE, got %v", result.MissingFromWPE)
	}
	if containsItem(result.MissingFromVIP, "plugins/hello-world") {
		t.Errorf("plugins/hello-world should not be in MissingFromVIP")
	}
	if containsItem(result.MissingFromWPE, "plugins/hello-world") {
		t.Errorf("plugins/hello-world should not be in MissingFromWPE")
	}
}

func TestVIPDestination_CompareFiles_EmptyDirs(t *testing.T) {
	// Both sides empty — should return empty result with no error.
	dest := vipdest.NewVIPDestination(t.TempDir())
	result, err := dest.CompareFiles(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.MissingFromVIP) != 0 || len(result.MissingFromWPE) != 0 {
		t.Errorf("expected empty result for empty dirs, got MissingFromVIP=%v MissingFromWPE=%v",
			result.MissingFromVIP, result.MissingFromWPE)
	}
}

func containsItem(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
