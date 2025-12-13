package provider

import (
	"io"
	"strings"
	"testing"
)

// mockProvider is a test implementation of the Provider interface
type mockProvider struct {
	name         string
	description  string
	capabilities ProviderCapabilities
	sites        []Site

	// Track method calls for verification
	authenticateCalled    bool
	testConnectionCalled  bool
	validateCredsCalled   bool
	lastCredentials       map[string]string
	authenticateError     error
	testConnectionError   error
	validateCredsError    error
}

func newMockProvider(name string) *mockProvider {
	return &mockProvider{
		name:        name,
		description: "Mock provider for testing",
		capabilities: ProviderCapabilities{
			Authentication: true,
			SiteManagement: true,
			DatabaseExport: true,
			DatabaseImport: true,
			FileSync:       true,
		},
		sites: []Site{
			{
				ID:            "site-1",
				Name:          "test-site",
				PrimaryDomain: "test.example.com",
				Environment:   "production",
				Status:        "active",
				Provider:      name,
			},
		},
	}
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Description() string {
	return m.description
}

func (m *mockProvider) Capabilities() ProviderCapabilities {
	return m.capabilities
}

func (m *mockProvider) Authenticate(credentials map[string]string) error {
	m.authenticateCalled = true
	m.lastCredentials = credentials
	return m.authenticateError
}

func (m *mockProvider) TestConnection() error {
	m.testConnectionCalled = true
	return m.testConnectionError
}

func (m *mockProvider) ValidateCredentials(credentials map[string]string) error {
	m.validateCredsCalled = true
	m.lastCredentials = credentials
	return m.validateCredsError
}

func (m *mockProvider) ListSites() ([]Site, error) {
	return m.sites, nil
}

func (m *mockProvider) GetSite(identifier string) (*Site, error) {
	for _, site := range m.sites {
		if site.ID == identifier || site.Name == identifier {
			return &site, nil
		}
	}
	return nil, nil
}

func (m *mockProvider) GetSiteMetadata(site *Site) (*SiteMetadata, error) {
	return &SiteMetadata{
		Site:             site,
		PHPVersion:       "8.1",
		MySQLVersion:     "8.0",
		WordPressVersion: "6.4",
		Domains:          []string{site.PrimaryDomain},
	}, nil
}

func (m *mockProvider) ExportDatabase(site *Site, options DatabaseExportOptions) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("-- SQL dump")), nil
}

func (m *mockProvider) ImportDatabase(site *Site, data io.Reader, options DatabaseImportOptions) error {
	return nil
}

func (m *mockProvider) GetDatabaseCredentials(site *Site) (*DatabaseCredentials, error) {
	return &DatabaseCredentials{
		Host:     "localhost",
		Port:     3306,
		Database: "wordpress",
		Username: "wp_user",
		Password: "wp_pass",
		SSL:      false,
	}, nil
}

func (m *mockProvider) SyncFiles(site *Site, destination string, options SyncOptions) error {
	return nil
}

func (m *mockProvider) DownloadFile(site *Site, remotePath string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("file content")), nil
}

func (m *mockProvider) UploadFile(site *Site, localPath, remotePath string) error {
	return nil
}

func (m *mockProvider) GetPHPVersion(site *Site) (string, error) {
	return "8.1", nil
}

func (m *mockProvider) GetMySQLVersion(site *Site) (string, error) {
	return "8.0", nil
}

func (m *mockProvider) GetWordPressVersion(site *Site) (string, error) {
	return "6.4", nil
}

// TestProviderInterface verifies the mock implements Provider
func TestProviderInterface(t *testing.T) {
	var _ Provider = (*mockProvider)(nil)
}

func TestSite_Fields(t *testing.T) {
	site := Site{
		ID:            "site-123",
		Name:          "my-site",
		PrimaryDomain: "mysite.com",
		Environment:   "production",
		Status:        "active",
		Provider:      "wpengine",
		Metadata: map[string]string{
			"region": "us-east-1",
		},
	}

	if site.ID != "site-123" {
		t.Errorf("expected ID to be site-123, got %s", site.ID)
	}
	if site.Name != "my-site" {
		t.Errorf("expected Name to be my-site, got %s", site.Name)
	}
	if site.PrimaryDomain != "mysite.com" {
		t.Errorf("expected PrimaryDomain to be mysite.com, got %s", site.PrimaryDomain)
	}
	if site.Environment != "production" {
		t.Errorf("expected Environment to be production, got %s", site.Environment)
	}
	if site.Status != "active" {
		t.Errorf("expected Status to be active, got %s", site.Status)
	}
	if site.Provider != "wpengine" {
		t.Errorf("expected Provider to be wpengine, got %s", site.Provider)
	}
	if site.Metadata["region"] != "us-east-1" {
		t.Errorf("expected Metadata[region] to be us-east-1, got %s", site.Metadata["region"])
	}
}

func TestSiteMetadata_Fields(t *testing.T) {
	site := &Site{ID: "test", Name: "test-site"}
	metadata := SiteMetadata{
		Site:             site,
		PHPVersion:       "8.2",
		MySQLVersion:     "8.0.33",
		WordPressVersion: "6.4.2",
		DiskUsage: DiskUsage{
			Used:  1024 * 1024 * 500, // 500MB
			Total: 1024 * 1024 * 1024, // 1GB
		},
		Domains:   []string{"primary.com", "alias.com"},
		Features:  []string{"ssl", "cdn", "backups"},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-06-01T00:00:00Z",
	}

	if metadata.PHPVersion != "8.2" {
		t.Errorf("expected PHPVersion to be 8.2, got %s", metadata.PHPVersion)
	}
	if metadata.DiskUsage.Used != 1024*1024*500 {
		t.Errorf("unexpected DiskUsage.Used: %d", metadata.DiskUsage.Used)
	}
	if len(metadata.Domains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(metadata.Domains))
	}
	if len(metadata.Features) != 3 {
		t.Errorf("expected 3 features, got %d", len(metadata.Features))
	}
}

func TestProviderCapabilities_Fields(t *testing.T) {
	caps := ProviderCapabilities{
		Authentication: true,
		SiteManagement: true,
		DatabaseExport: true,
		DatabaseImport: true,
		FileSync:       true,
		Deployment:     true,
		Environments:   true,
		Backups:        true,
		RemoteExecution: true,
		MediaManagement: true,
		SSHAccess:      true,
		APIAccess:      true,
		Scaling:        false,
		Monitoring:     false,
		Logging:        false,
	}

	if !caps.Authentication {
		t.Error("expected Authentication to be true")
	}
	if !caps.Deployment {
		t.Error("expected Deployment to be true")
	}
	if caps.Scaling {
		t.Error("expected Scaling to be false")
	}
}

func TestDatabaseExportOptions_Fields(t *testing.T) {
	opts := DatabaseExportOptions{
		ExcludeTables:  []string{"wp_sessions", "wp_statistics"},
		SkipLogs:       true,
		SkipTransients: true,
		SkipSpam:       true,
		Compress:       true,
		IncludePrefix:  true,
	}

	if len(opts.ExcludeTables) != 2 {
		t.Errorf("expected 2 excluded tables, got %d", len(opts.ExcludeTables))
	}
	if !opts.SkipLogs {
		t.Error("expected SkipLogs to be true")
	}
	if !opts.Compress {
		t.Error("expected Compress to be true")
	}
}

func TestDatabaseImportOptions_Fields(t *testing.T) {
	opts := DatabaseImportOptions{
		DropExisting:  true,
		SearchReplace: []string{"old.com", "new.com"},
		SkipErrors:    false,
	}

	if !opts.DropExisting {
		t.Error("expected DropExisting to be true")
	}
	if len(opts.SearchReplace) != 2 {
		t.Errorf("expected 2 search-replace items, got %d", len(opts.SearchReplace))
	}
}

func TestDatabaseCredentials_Fields(t *testing.T) {
	creds := DatabaseCredentials{
		Host:     "db.example.com",
		Port:     3306,
		Database: "wordpress_db",
		Username: "wp_user",
		Password: "secret123",
		SSL:      true,
	}

	if creds.Host != "db.example.com" {
		t.Errorf("unexpected Host: %s", creds.Host)
	}
	if creds.Port != 3306 {
		t.Errorf("expected Port to be 3306, got %d", creds.Port)
	}
	if !creds.SSL {
		t.Error("expected SSL to be true")
	}
}

func TestSyncOptions_Fields(t *testing.T) {
	opts := SyncOptions{
		Source:         "/wp-content/uploads/",
		Destination:    "/local/uploads/",
		Include:        []string{"*.jpg", "*.png"},
		Exclude:        []string{"*.tmp", "cache/*"},
		Delete:         true,
		DryRun:         false,
		BandwidthLimit: 5000,
		Progress:       true,
	}

	if opts.Source != "/wp-content/uploads/" {
		t.Errorf("unexpected Source: %s", opts.Source)
	}
	if len(opts.Include) != 2 {
		t.Errorf("expected 2 include patterns, got %d", len(opts.Include))
	}
	if len(opts.Exclude) != 2 {
		t.Errorf("expected 2 exclude patterns, got %d", len(opts.Exclude))
	}
	if opts.BandwidthLimit != 5000 {
		t.Errorf("expected BandwidthLimit to be 5000, got %d", opts.BandwidthLimit)
	}
}

func TestEnvironment_Fields(t *testing.T) {
	env := Environment{
		Name:         "staging",
		URL:          "https://staging.example.com",
		Status:       "active",
		IsDefault:    false,
		LastDeployAt: "2024-06-01T12:00:00Z",
	}

	if env.Name != "staging" {
		t.Errorf("expected Name to be staging, got %s", env.Name)
	}
	if env.IsDefault {
		t.Error("expected IsDefault to be false")
	}
}

func TestDeployOptions_Fields(t *testing.T) {
	opts := DeployOptions{
		Branch:      "main",
		Commit:      "abc123",
		Message:     "Deploy v1.2.3",
		Environment: "production",
		Metadata: map[string]string{
			"triggered_by": "ci",
		},
	}

	if opts.Branch != "main" {
		t.Errorf("expected Branch to be main, got %s", opts.Branch)
	}
	if opts.Metadata["triggered_by"] != "ci" {
		t.Errorf("unexpected Metadata: %v", opts.Metadata)
	}
}

func TestDeployment_Fields(t *testing.T) {
	deployment := Deployment{
		ID:         "deploy-123",
		Status:     "completed",
		Branch:     "main",
		Commit:     "abc123def",
		Message:    "Feature release",
		DeployedAt: "2024-06-01T12:00:00Z",
		DeployedBy: "user@example.com",
	}

	if deployment.ID != "deploy-123" {
		t.Errorf("unexpected ID: %s", deployment.ID)
	}
	if deployment.Status != "completed" {
		t.Errorf("expected Status to be completed, got %s", deployment.Status)
	}
}

func TestBackup_Fields(t *testing.T) {
	backup := Backup{
		ID:          "backup-456",
		Type:        "manual",
		Description: "Pre-update backup",
		Size:        1024 * 1024 * 100, // 100MB
		CreatedAt:   "2024-06-01T12:00:00Z",
		Status:      "completed",
		ExpiresAt:   "2024-07-01T12:00:00Z",
	}

	if backup.ID != "backup-456" {
		t.Errorf("unexpected ID: %s", backup.ID)
	}
	if backup.Type != "manual" {
		t.Errorf("expected Type to be manual, got %s", backup.Type)
	}
	if backup.Size != 1024*1024*100 {
		t.Errorf("unexpected Size: %d", backup.Size)
	}
}

func TestMediaOptions_Fields(t *testing.T) {
	opts := MediaOptions{
		CDNEnabled: true,
		CDNDomain:  "cdn.example.com",
		CacheTTL:   86400,
		Excludes:   []string{"/wp-admin/*", "/private/*"},
	}

	if !opts.CDNEnabled {
		t.Error("expected CDNEnabled to be true")
	}
	if opts.CacheTTL != 86400 {
		t.Errorf("expected CacheTTL to be 86400, got %d", opts.CacheTTL)
	}
	if len(opts.Excludes) != 2 {
		t.Errorf("expected 2 excludes, got %d", len(opts.Excludes))
	}
}

func TestMigrationOptions_Fields(t *testing.T) {
	opts := MigrationOptions{
		IncludeDatabase: true,
		IncludeFiles:    true,
		IncludeMedia:    true,
		ExcludePlugins:  []string{"debug-bar", "query-monitor"},
		ExcludeThemes:   []string{"twentytwenty"},
		DryRun:          false,
	}

	if !opts.IncludeDatabase {
		t.Error("expected IncludeDatabase to be true")
	}
	if len(opts.ExcludePlugins) != 2 {
		t.Errorf("expected 2 excluded plugins, got %d", len(opts.ExcludePlugins))
	}
}
