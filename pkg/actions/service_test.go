package actions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetup_CreatesWorkflowFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0755); err != nil {
		t.Fatal(err)
	}
	// Initialize a git repo so git.IsGitRepository passes
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	opts := SetupOptions{
		ProductionBranch: "main",
		StagingBranch:    "develop",
		ProdInstall:      "mysite",
		StageInstall:     "mysite-staging",
		ProjectDir:       dir,
		Force:            true,
	}

	if err := Setup(opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wfPath := filepath.Join(dir, ".github", "workflows", "deploy.yml")
	if _, err := os.Stat(wfPath); os.IsNotExist(err) {
		t.Error("expected deploy.yml to be created")
	}
}
