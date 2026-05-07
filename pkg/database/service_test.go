package database

import (
	"testing"
)

func TestGetDDEVURL(t *testing.T) {
	url := GetDDEVURL("my-project")
	if url != "https://my-project.ddev.site" {
		t.Errorf("expected 'https://my-project.ddev.site', got %q", url)
	}
}

func TestGetWPEngineURL_Production(t *testing.T) {
	url := GetWPEngineURL("mysite", "production", "")
	if url != "https://mysite.wpengine.com" {
		t.Errorf("expected 'https://mysite.wpengine.com', got %q", url)
	}
}

func TestGetWPEngineURL_CustomDomain(t *testing.T) {
	url := GetWPEngineURL("mysite", "production", "www.example.com")
	if url != "https://www.example.com" {
		t.Errorf("expected 'https://www.example.com', got %q", url)
	}
}
