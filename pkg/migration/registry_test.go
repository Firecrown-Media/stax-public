package migration_test

import (
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/migration"
	"github.com/firecrown-media/stax/pkg/provider"
)

type mockSource struct{}

func (m *mockSource) PullFiles(_ migration.PullOptions) error        { return nil }
func (m *mockSource) ExportDatabase(_ migration.ExportOptions) error { return nil }

type mockDest struct{}

func (m *mockDest) Audit(_ string, _ migration.AuditOptions) (*migration.AuditReport, error) {
	return &migration.AuditReport{}, nil
}
func (m *mockDest) ValidateDatabase(_ string) error                          { return nil }
func (m *mockDest) ImportDatabase(_ string, _ migration.ImportOptions) error { return nil }
func (m *mockDest) ImportMedia(_ migration.ImportOptions) error              { return nil }
func (m *mockDest) CompareFiles(_ string) (*migration.CompareResult, error) {
	return &migration.CompareResult{}, nil
}

func TestRegistry_Source(t *testing.T) {
	t.Cleanup(migration.ResetRegistryForTesting)
	migration.RegisterSource("test-provider", func(p provider.Provider, cfg *config.Config) migration.Source {
		return &mockSource{}
	})

	src, err := migration.NewSource("test-provider", nil, nil)
	if err != nil {
		t.Fatalf("NewSource returned error: %v", err)
	}
	if src == nil {
		t.Fatal("expected non-nil Source")
	}
}

func TestRegistry_Source_NotFound(t *testing.T) {
	t.Cleanup(migration.ResetRegistryForTesting)
	_, err := migration.NewSource("does-not-exist", nil, nil)
	if err == nil {
		t.Fatal("expected error for unregistered source")
	}
}

func TestRegistry_Destination(t *testing.T) {
	t.Cleanup(migration.ResetRegistryForTesting)
	migration.RegisterDestination("test-dest", func(repoPath string) migration.Destination {
		return &mockDest{}
	})

	dest, err := migration.NewDestination("test-dest", "/some/path")
	if err != nil {
		t.Fatalf("NewDestination returned error: %v", err)
	}
	if dest == nil {
		t.Fatal("expected non-nil Destination")
	}
}

func TestRegistry_Destination_NotFound(t *testing.T) {
	t.Cleanup(migration.ResetRegistryForTesting)
	_, err := migration.NewDestination("does-not-exist", "")
	if err == nil {
		t.Fatal("expected error for unregistered destination")
	}
}
