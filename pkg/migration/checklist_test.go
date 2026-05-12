package migration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/firecrown-media/stax/pkg/migration"
)

func TestWriteChecklist_AllSections(t *testing.T) {
	data := migration.ChecklistData{
		Install:      "mysite",
		Destination:  "vip",
		Domain:       "mysite.com",
		GeneratedAt:  "2026-05-12 10:00:00 CDT",
		ReportPath:   ".stax/mysite-migration-report.md",
		ReportExists: false,
		SQLPath:      ".stax/mysite-export.sql",
		SQLExists:    false,
		VIPCommitSHA: "",
		PullDone:     false,
		ExportDone:   false,
		ReportDone:   false,
		PublishDone:  false,
	}
	outputPath := filepath.Join(t.TempDir(), "checklist.md")
	if err := migration.WriteChecklist(data, outputPath); err != nil {
		t.Fatalf("WriteChecklist() failed: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read checklist: %v", err)
	}
	s := string(content)
	for _, want := range []string{
		"Migration Checklist: mysite",
		"mysite.com",
		"Pre-migration steps",
		"QA Checklist",
		"DNS Cutover",
		"Post-launch Validation",
		"Sign-off",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("checklist missing section %q", want)
		}
	}
}

func TestWriteChecklist_PreCheckedItems(t *testing.T) {
	data := migration.ChecklistData{
		Install:      "mysite",
		Destination:  "vip",
		Domain:       "mysite.com",
		GeneratedAt:  "2026-05-12 10:00:00 CDT",
		ReportPath:   ".stax/mysite-migration-report.md",
		ReportExists: true,
		SQLPath:      ".stax/mysite-export.sql",
		SQLExists:    true,
		VIPCommitSHA: "abc1234",
		PullDone:     true,
		ExportDone:   true,
		ReportDone:   true,
		PublishDone:  true,
	}
	outputPath := filepath.Join(t.TempDir(), "checklist.md")
	if err := migration.WriteChecklist(data, outputPath); err != nil {
		t.Fatalf("WriteChecklist() failed: %v", err)
	}
	content, _ := os.ReadFile(outputPath)
	s := string(content)

	checks := []struct {
		substr string
		want   bool
	}{
		{"[x] `stax migrate pull`", true},
		{"[x] `stax migrate export`", true},
		{"[x] `stax migrate audit`", true},
		{"[x] `stax migrate compare`", true},
		{"[x] `stax migrate report`", true},
		{"[x] `stax migrate publish`", true},
		{"[ ] `stax migrate import`", true}, // import is never pre-checked
		{"abc1234", true},
	}
	for _, c := range checks {
		got := strings.Contains(s, c.substr)
		if got != c.want {
			t.Errorf("expected Contains(%q)=%v, got %v", c.substr, c.want, got)
		}
	}
}

func TestWriteChecklist_UncheckedWhenNoArtifacts(t *testing.T) {
	data := migration.ChecklistData{
		Install:     "mysite",
		Destination: "vip",
		Domain:      "mysite.com",
		GeneratedAt: "2026-05-12 10:00:00 CDT",
	}
	outputPath := filepath.Join(t.TempDir(), "checklist.md")
	if err := migration.WriteChecklist(data, outputPath); err != nil {
		t.Fatalf("WriteChecklist() failed: %v", err)
	}
	content, _ := os.ReadFile(outputPath)
	s := string(content)

	for _, step := range []string{"pull", "export", "audit", "compare", "import", "report", "publish"} {
		substr := "[ ] `stax migrate " + step + "`"
		if !strings.Contains(s, substr) {
			t.Errorf("step %q should be unchecked when no artifacts, missing %q", step, substr)
		}
	}
}
