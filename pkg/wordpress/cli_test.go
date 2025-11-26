package wordpress

import (
	"os"
	"testing"
)

func TestNewCLI(t *testing.T) {
	projectDir := "/test/project"
	cli := NewCLI(projectDir)

	if cli == nil {
		t.Fatal("NewCLI() returned nil")
	}

	if cli.ProjectDir != projectDir {
		t.Errorf("expected ProjectDir %q, got %q", projectDir, cli.ProjectDir)
	}

	if !cli.useDDEV {
		t.Error("expected useDDEV to be true by default")
	}
}

func TestIsInstalled(t *testing.T) {
	// This test checks if WP-CLI is in PATH
	installed := IsInstalled()

	// Just verify the function executes without error
	t.Logf("WP-CLI installed: %v", installed)
}

func TestCLI_SearchReplace(t *testing.T) {
	if os.Getenv("RUN_WP_TESTS") != "true" {
		t.Skip("Skipping WP-CLI integration test (set RUN_WP_TESTS=true to run)")
	}

	tests := []struct {
		name    string
		old     string
		new     string
		network bool
		wantErr bool
	}{
		{
			name:    "simple search replace",
			old:     "https://old.com",
			new:     "https://new.local",
			network: false,
			wantErr: true, // Will fail without actual WordPress install
		},
		{
			name:    "network search replace",
			old:     "https://old.com",
			new:     "https://new.local",
			network: true,
			wantErr: true, // Will fail without actual WordPress install
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cli := NewCLI(tmpDir)

			err := cli.SearchReplace(tt.old, tt.new, tt.network)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchReplace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCLI_GetSites(t *testing.T) {
	if os.Getenv("RUN_WP_TESTS") != "true" {
		t.Skip("Skipping WP-CLI integration test (set RUN_WP_TESTS=true to run)")
	}

	tmpDir := t.TempDir()
	cli := NewCLI(tmpDir)

	// This will fail without a WordPress multisite installation
	sites, err := cli.GetSites()
	if err == nil {
		t.Error("expected GetSites() to fail without WordPress install")
	}

	if len(sites) != 0 {
		t.Errorf("expected 0 sites, got %d", len(sites))
	}
}

func TestCLI_GetSiteURL(t *testing.T) {
	if os.Getenv("RUN_WP_TESTS") != "true" {
		t.Skip("Skipping WP-CLI integration test (set RUN_WP_TESTS=true to run)")
	}

	tests := []struct {
		name    string
		blogID  int
		wantErr bool
	}{
		{
			name:    "get main site URL",
			blogID:  1,
			wantErr: true, // Will fail without actual WordPress install
		},
		{
			name:    "get subsite URL",
			blogID:  2,
			wantErr: true, // Will fail without actual WordPress install
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cli := NewCLI(tmpDir)

			url, err := cli.GetSiteURL(tt.blogID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSiteURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && url == "" {
				t.Error("expected non-empty URL")
			}
		})
	}
}

// Unit tests that don't require WordPress or WP-CLI
func TestCLIConstruction(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
	}{
		{
			name:       "create CLI with valid directory",
			projectDir: "/test/project",
		},
		{
			name:       "create CLI with empty directory",
			projectDir: "",
		},
		{
			name:       "create CLI with relative path",
			projectDir: "./test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI(tt.projectDir)

			if cli == nil {
				t.Fatal("NewCLI() returned nil")
			}

			if cli.ProjectDir != tt.projectDir {
				t.Errorf("ProjectDir = %q, want %q", cli.ProjectDir, tt.projectDir)
			}

			if !cli.useDDEV {
				t.Error("expected useDDEV to be true")
			}
		})
	}
}

func TestSiteStruct(t *testing.T) {
	site := Site{
		ID:  1,
		URL: "https://example.local",
	}

	if site.ID != 1 {
		t.Errorf("expected ID 1, got %d", site.ID)
	}

	if site.URL != "https://example.local" {
		t.Errorf("expected URL 'https://example.local', got %q", site.URL)
	}
}

func TestVerifySiteURL(t *testing.T) {
	if os.Getenv("RUN_WP_TESTS") != "true" {
		t.Skip("Skipping WP-CLI integration test (set RUN_WP_TESTS=true to run)")
	}

	tests := []struct {
		name        string
		expectedURL string
		wantMatch   bool
		wantErr     bool
	}{
		{
			name:        "verify URL matches",
			expectedURL: "https://example.local",
			wantMatch:   false, // Will not match without real WordPress
			wantErr:     true,  // Will error without real WordPress
		},
		{
			name:        "verify URL with trailing slash",
			expectedURL: "https://example.local/",
			wantMatch:   false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cli := NewCLI(tmpDir)

			matches, actualURL, err := cli.VerifySiteURL(tt.expectedURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySiteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if matches != tt.wantMatch {
					t.Errorf("VerifySiteURL() matches = %v, want %v (actual URL: %s)", matches, tt.wantMatch, actualURL)
				}
			}
		})
	}
}
