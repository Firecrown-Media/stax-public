package wordpress

import (
	"testing"
)

func TestParseSearchReplaceOutput(t *testing.T) {
	tests := []struct {
		name             string
		output           string
		wantReplacements int
		wantSuccess      bool
	}{
		{
			name:             "successful replacement with multiple changes",
			output:           "Success: Made 42 replacements.",
			wantReplacements: 42,
			wantSuccess:      true,
		},
		{
			name:             "successful replacement with single change",
			output:           "Success: Made 1 replacement.",
			wantReplacements: 1,
			wantSuccess:      true,
		},
		{
			name:             "zero replacements",
			output:           "Success: Made 0 replacements.",
			wantReplacements: 0,
			wantSuccess:      true,
		},
		{
			name:             "error in output",
			output:           "Error: The database could not be accessed.",
			wantReplacements: 0,
			wantSuccess:      false,
		},
		{
			name:             "fatal error in output",
			output:           "Fatal error: Cannot connect to database",
			wantReplacements: 0,
			wantSuccess:      false,
		},
		{
			name: "typical wp search-replace output with table info",
			output: `+---------------------+-----------------------+--------------+------+
| Table               | Column                | Replacements | Type |
+---------------------+-----------------------+--------------+------+
| wp_options          | option_value          | 15           | SQL  |
| wp_posts            | post_content          | 27           | SQL  |
+---------------------+-----------------------+--------------+------+
Success: Made 42 replacements.`,
			wantReplacements: 42,
			wantSuccess:      true,
		},
		{
			name:             "empty output",
			output:           "",
			wantReplacements: 0,
			wantSuccess:      false,
		},
		{
			name:             "warning but successful",
			output:           "Warning: Some theme files could not be loaded.\nSuccess: Made 15 replacements.",
			wantReplacements: 15,
			wantSuccess:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSearchReplaceOutput(tt.output)

			if result.Success != tt.wantSuccess {
				t.Errorf("parseSearchReplaceOutput() success = %v, want %v", result.Success, tt.wantSuccess)
			}

			if result.Replacements != tt.wantReplacements {
				t.Errorf("parseSearchReplaceOutput() replacements = %v, want %v", result.Replacements, tt.wantReplacements)
			}

			if result.Output != tt.output {
				t.Errorf("parseSearchReplaceOutput() output not preserved")
			}
		})
	}
}

func TestSearchReplaceOptions(t *testing.T) {
	tests := []struct {
		name string
		opts SearchReplaceOptions
	}{
		{
			name: "network-wide replacement",
			opts: SearchReplaceOptions{
				Network:     true,
				SkipColumns: []string{"guid"},
				DryRun:      false,
			},
		},
		{
			name: "single-site replacement",
			opts: SearchReplaceOptions{
				Network:     false,
				SkipColumns: []string{"guid", "post_name"},
				DryRun:      true,
			},
		},
		{
			name: "specific site replacement",
			opts: SearchReplaceOptions{
				Network:     false,
				URL:         "example.local",
				SkipColumns: []string{"guid"},
				DryRun:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the struct is properly constructed
			if tt.opts.Network && tt.opts.URL != "" {
				t.Log("Note: Network and URL flags are both set, which may cause unexpected behavior")
			}
		})
	}
}

func TestSearchReplaceResult(t *testing.T) {
	result := SearchReplaceResult{
		Replacements: 42,
		Output:       "Success: Made 42 replacements.",
		Success:      true,
	}

	if result.Replacements != 42 {
		t.Errorf("expected 42 replacements, got %d", result.Replacements)
	}

	if !result.Success {
		t.Error("expected success to be true")
	}

	if result.Output == "" {
		t.Error("expected non-empty output")
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name      string
		strings   []string
		separator string
		want      string
	}{
		{
			name:      "join with comma",
			strings:   []string{"guid", "post_name"},
			separator: ",",
			want:      "guid,post_name",
		},
		{
			name:      "join with pipe",
			strings:   []string{"a", "b", "c"},
			separator: "|",
			want:      "a|b|c",
		},
		{
			name:      "single string",
			strings:   []string{"guid"},
			separator: ",",
			want:      "guid",
		},
		{
			name:      "empty slice",
			strings:   []string{},
			separator: ",",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinStrings(tt.strings, tt.separator)
			if got != tt.want {
				t.Errorf("joinStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultisiteSearchReplaceConfig(t *testing.T) {
	config := MultisiteSearchReplaceConfig{
		Network: NetworkReplace{
			Old: "https://old.wpengine.com",
			New: "https://local.ddev.site",
		},
		Sites: []SiteReplace{
			{
				Old:    "https://site1.old.com",
				New:    "https://site1.local",
				URL:    "site1.local",
				BlogID: 2,
			},
			{
				Old:    "https://site2.old.com",
				New:    "https://site2.local",
				URL:    "site2.local",
				BlogID: 3,
			},
		},
	}

	if config.Network.Old != "https://old.wpengine.com" {
		t.Errorf("unexpected network old URL: %s", config.Network.Old)
	}

	if len(config.Sites) != 2 {
		t.Errorf("expected 2 sites, got %d", len(config.Sites))
	}

	if config.Sites[0].BlogID != 2 {
		t.Errorf("expected blog ID 2, got %d", config.Sites[0].BlogID)
	}
}

func TestBuildFirecrownSearchReplaceConfig(t *testing.T) {
	localDomain := "local.ddev.site"
	config := BuildFirecrownSearchReplaceConfig(localDomain)

	if config.Network.Old != "fsmultisite.wpenginepowered.com" {
		t.Errorf("unexpected network old URL: %s", config.Network.Old)
	}

	if config.Network.New != localDomain {
		t.Errorf("unexpected network new URL: %s", config.Network.New)
	}

	if len(config.Sites) != 4 {
		t.Errorf("expected 4 sites in Firecrown config, got %d", len(config.Sites))
	}

	// Verify one of the site configs
	foundFlying := false
	for _, site := range config.Sites {
		if site.Old == "flyingmag.com" {
			foundFlying = true
			expectedNew := "flyingmag." + localDomain
			if site.New != expectedNew {
				t.Errorf("expected flying mag new URL to be %s, got %s", expectedNew, site.New)
			}
		}
	}

	if !foundFlying {
		t.Error("expected to find flyingmag.com in site configs")
	}
}
