package provider

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.GetProviderName() != "test-provider" {
		t.Errorf("expected provider name test-provider, got %s", manager.GetProviderName())
	}
}

func TestManager_GetProvider(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	provider := manager.GetProvider()
	if provider == nil {
		t.Fatal("GetProvider returned nil")
	}

	if provider.Name() != "test-provider" {
		t.Errorf("expected provider name test-provider, got %s", provider.Name())
	}
}

func TestManager_GetProviderName(t *testing.T) {
	mock := newMockProvider("my-provider")
	manager := NewManager(mock)

	if manager.GetProviderName() != "my-provider" {
		t.Errorf("expected my-provider, got %s", manager.GetProviderName())
	}
}

func TestNewManagerFromConfig(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register a mock provider
	mock := newMockProvider("test-provider")
	if err := RegisterProvider("test-provider", mock); err != nil {
		t.Fatalf("failed to register mock provider: %v", err)
	}

	config := ProviderConfig{
		Name: "test-provider",
	}

	manager, err := NewManagerFromConfig(config)
	if err != nil {
		t.Fatalf("NewManagerFromConfig failed: %v", err)
	}

	if manager.GetProviderName() != "test-provider" {
		t.Errorf("expected test-provider, got %s", manager.GetProviderName())
	}
}

func TestNewManagerFromConfig_InvalidConfig(t *testing.T) {
	config := ProviderConfig{
		Name: "", // Invalid - empty name
	}

	_, err := NewManagerFromConfig(config)
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestManager_SwitchProvider(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register two mock providers
	mock1 := newMockProvider("provider-1")
	mock2 := newMockProvider("provider-2")
	if err := RegisterProvider("provider-1", mock1); err != nil {
		t.Fatalf("failed to register provider-1: %v", err)
	}
	if err := RegisterProvider("provider-2", mock2); err != nil {
		t.Fatalf("failed to register provider-2: %v", err)
	}

	manager := NewManager(mock1)
	if manager.GetProviderName() != "provider-1" {
		t.Fatalf("expected provider-1, got %s", manager.GetProviderName())
	}

	// Switch to provider-2
	err := manager.SwitchProvider(ProviderConfig{Name: "provider-2"})
	if err != nil {
		t.Fatalf("SwitchProvider failed: %v", err)
	}

	if manager.GetProviderName() != "provider-2" {
		t.Errorf("expected provider-2 after switch, got %s", manager.GetProviderName())
	}
}

func TestManager_SwitchProvider_InvalidProvider(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	mock := newMockProvider("provider-1")
	if err := RegisterProvider("provider-1", mock); err != nil {
		t.Fatalf("failed to register provider-1: %v", err)
	}

	manager := NewManager(mock)

	err := manager.SwitchProvider(ProviderConfig{Name: "nonexistent"})
	if err == nil {
		t.Error("expected error when switching to nonexistent provider")
	}

	// Verify original provider is still active
	if manager.GetProviderName() != "provider-1" {
		t.Errorf("expected provider-1 to remain active, got %s", manager.GetProviderName())
	}
}

func TestManager_TestCurrentProvider(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	err := manager.TestCurrentProvider()
	if err != nil {
		t.Errorf("TestCurrentProvider failed: %v", err)
	}

	if !mock.testConnectionCalled {
		t.Error("expected TestConnection to be called")
	}
}

func TestManager_TestCurrentProvider_NoProvider(t *testing.T) {
	manager := &Manager{
		currentProvider: nil,
	}

	err := manager.TestCurrentProvider()
	if err == nil {
		t.Error("expected error when no provider configured")
	}
}

func TestManager_TestCurrentProvider_Fails(t *testing.T) {
	mock := newMockProvider("test-provider")
	mock.testConnectionError = fmt.Errorf("connection failed")
	manager := NewManager(mock)

	err := manager.TestCurrentProvider()
	if err == nil {
		t.Error("expected error when connection test fails")
	}
}

func TestManager_GetCurrentProviderInfo(t *testing.T) {
	mock := newMockProvider("test-provider")
	mock.description = "Test provider description"
	mock.capabilities = ProviderCapabilities{
		Authentication: true,
		SiteManagement: true,
	}
	manager := NewManager(mock)

	info, err := manager.GetCurrentProviderInfo()
	if err != nil {
		t.Fatalf("GetCurrentProviderInfo failed: %v", err)
	}

	if info.Name != "test-provider" {
		t.Errorf("expected name test-provider, got %s", info.Name)
	}
	if info.Description != "Test provider description" {
		t.Errorf("unexpected description: %s", info.Description)
	}
	if !info.Capabilities.Authentication {
		t.Error("expected Authentication capability to be true")
	}
}

func TestManager_GetCurrentProviderInfo_NoProvider(t *testing.T) {
	manager := &Manager{
		currentProvider: nil,
	}

	_, err := manager.GetCurrentProviderInfo()
	if err == nil {
		t.Error("expected error when no provider configured")
	}
}

func TestManager_ListSites(t *testing.T) {
	mock := newMockProvider("test-provider")
	mock.sites = []Site{
		{ID: "site-1", Name: "Site One"},
		{ID: "site-2", Name: "Site Two"},
	}
	manager := NewManager(mock)

	sites, err := manager.ListSites()
	if err != nil {
		t.Fatalf("ListSites failed: %v", err)
	}

	if len(sites) != 2 {
		t.Errorf("expected 2 sites, got %d", len(sites))
	}
}

func TestManager_ListSites_NoProvider(t *testing.T) {
	manager := &Manager{
		currentProvider: nil,
	}

	_, err := manager.ListSites()
	if err == nil {
		t.Error("expected error when no provider configured")
	}
}

func TestManager_GetSite(t *testing.T) {
	mock := newMockProvider("test-provider")
	mock.sites = []Site{
		{ID: "site-1", Name: "Site One"},
	}
	manager := NewManager(mock)

	site, err := manager.GetSite("site-1")
	if err != nil {
		t.Fatalf("GetSite failed: %v", err)
	}

	if site == nil {
		t.Fatal("GetSite returned nil")
	}
	if site.ID != "site-1" {
		t.Errorf("expected site ID site-1, got %s", site.ID)
	}
}

func TestManager_GetSite_NoProvider(t *testing.T) {
	manager := &Manager{
		currentProvider: nil,
	}

	_, err := manager.GetSite("site-1")
	if err == nil {
		t.Error("expected error when no provider configured")
	}
}

func TestManager_ExportDatabase(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	site := &Site{ID: "site-1", Name: "Test Site"}
	options := DatabaseExportOptions{SkipLogs: true}

	reader, err := manager.ExportDatabase(site, options)
	if err != nil {
		t.Fatalf("ExportDatabase failed: %v", err)
	}
	defer reader.Close()

	// Read the content
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read export: %v", err)
	}

	if !strings.Contains(string(data), "SQL dump") {
		t.Errorf("unexpected export content: %s", string(data))
	}
}

func TestManager_ExportDatabase_NoProvider(t *testing.T) {
	manager := &Manager{
		currentProvider: nil,
	}

	site := &Site{ID: "site-1"}
	_, err := manager.ExportDatabase(site, DatabaseExportOptions{})
	if err == nil {
		t.Error("expected error when no provider configured")
	}
}

func TestManager_ImportDatabase(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	site := &Site{ID: "site-1", Name: "Test Site"}
	data := strings.NewReader("-- SQL dump content")
	options := DatabaseImportOptions{DropExisting: true}

	err := manager.ImportDatabase(site, data, options)
	if err != nil {
		t.Errorf("ImportDatabase failed: %v", err)
	}
}

func TestManager_ImportDatabase_NoProvider(t *testing.T) {
	manager := &Manager{
		currentProvider: nil,
	}

	site := &Site{ID: "site-1"}
	data := strings.NewReader("-- SQL")
	err := manager.ImportDatabase(site, data, DatabaseImportOptions{})
	if err == nil {
		t.Error("expected error when no provider configured")
	}
}

func TestManager_SyncFiles(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	site := &Site{ID: "site-1", Name: "Test Site"}
	options := SyncOptions{Destination: "/local/path"}

	err := manager.SyncFiles(site, "/local/path", options)
	if err != nil {
		t.Errorf("SyncFiles failed: %v", err)
	}
}

func TestManager_SyncFiles_NoProvider(t *testing.T) {
	manager := &Manager{
		currentProvider: nil,
	}

	site := &Site{ID: "site-1"}
	err := manager.SyncFiles(site, "/local/path", SyncOptions{})
	if err == nil {
		t.Error("expected error when no provider configured")
	}
}

func TestManager_Deploy_NotSupported(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	site := &Site{ID: "site-1"}
	options := DeployOptions{Branch: "main"}

	_, err := manager.Deploy(site, options)
	if err == nil {
		t.Error("expected error when provider doesn't support deployment")
	}
}

func TestManager_ListEnvironments_NotSupported(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	site := &Site{ID: "site-1"}
	_, err := manager.ListEnvironments(site)
	if err == nil {
		t.Error("expected error when provider doesn't support environment management")
	}
}

func TestManager_CreateBackup_NotSupported(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	site := &Site{ID: "site-1"}
	_, err := manager.CreateBackup(site, "test backup")
	if err == nil {
		t.Error("expected error when provider doesn't support backups")
	}
}

func TestManager_ExecuteWPCLI_NotSupported(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	site := &Site{ID: "site-1"}
	_, err := manager.ExecuteWPCLI(site, []string{"option", "get", "siteurl"})
	if err == nil {
		t.Error("expected error when provider doesn't support remote execution")
	}
}

func TestManager_GetMediaURL_NotSupported(t *testing.T) {
	mock := newMockProvider("test-provider")
	manager := NewManager(mock)

	site := &Site{ID: "site-1"}
	_, err := manager.GetMediaURL(site)
	if err == nil {
		t.Error("expected error when provider doesn't support media management")
	}
}

func TestManager_MigrateSite_NoSourceProvider(t *testing.T) {
	targetMock := newMockProvider("target-provider")

	manager := &Manager{
		currentProvider: nil,
	}

	options := MigrateOptions{
		SourceSite:     &Site{ID: "site-1"},
		TargetProvider: targetMock,
	}

	err := manager.MigrateSite(options)
	if err == nil {
		t.Error("expected error when no source provider")
	}
}

func TestManager_MigrateSite_NoTargetProvider(t *testing.T) {
	sourceMock := newMockProvider("source-provider")
	manager := NewManager(sourceMock)

	options := MigrateOptions{
		SourceSite:     &Site{ID: "site-1"},
		TargetProvider: nil,
	}

	err := manager.MigrateSite(options)
	if err == nil {
		t.Error("expected error when no target provider")
	}
}

func TestManager_MigrateSite_NoSourceSite(t *testing.T) {
	sourceMock := newMockProvider("source-provider")
	targetMock := newMockProvider("target-provider")
	manager := NewManager(sourceMock)

	options := MigrateOptions{
		SourceSite:     nil,
		TargetProvider: targetMock,
	}

	err := manager.MigrateSite(options)
	if err == nil {
		t.Error("expected error when no source site")
	}
}

func TestCompareProviders(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register two mock providers with different capabilities
	mock1 := newMockProvider("provider-1")
	mock1.capabilities = ProviderCapabilities{
		Authentication: true,
		SiteManagement: true,
		DatabaseExport: true,
		Deployment:     true,
	}

	mock2 := newMockProvider("provider-2")
	mock2.capabilities = ProviderCapabilities{
		Authentication: true,
		SiteManagement: true,
		DatabaseExport: false,
		Deployment:     false,
		Backups:        true,
	}

	if err := RegisterProvider("provider-1", mock1); err != nil {
		t.Fatalf("failed to register provider-1: %v", err)
	}
	if err := RegisterProvider("provider-2", mock2); err != nil {
		t.Fatalf("failed to register provider-2: %v", err)
	}

	comparison, err := CompareProviders("provider-1", "provider-2")
	if err != nil {
		t.Fatalf("CompareProviders failed: %v", err)
	}

	if comparison.Provider1 != "provider-1" {
		t.Errorf("expected Provider1 to be provider-1, got %s", comparison.Provider1)
	}
	if comparison.Provider2 != "provider-2" {
		t.Errorf("expected Provider2 to be provider-2, got %s", comparison.Provider2)
	}

	// Check shared features
	sharedContains := func(feature string) bool {
		for _, f := range comparison.SharedFeatures {
			if f == feature {
				return true
			}
		}
		return false
	}

	if !sharedContains("authentication") {
		t.Error("expected authentication in shared features")
	}
	if !sharedContains("site_management") {
		t.Error("expected site_management in shared features")
	}
	if sharedContains("deployment") {
		t.Error("deployment should not be in shared features")
	}
}

func TestCompareProviders_ProviderNotFound(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	_, err := CompareProviders("nonexistent-1", "nonexistent-2")
	if err == nil {
		t.Error("expected error for nonexistent providers")
	}
}

func TestGetProviderRecommendation(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register providers with different capabilities
	basic := newMockProvider("basic")
	basic.capabilities = ProviderCapabilities{
		Authentication: true,
		SiteManagement: true,
	}

	full := newMockProvider("full")
	full.capabilities = ProviderCapabilities{
		Authentication:  true,
		SiteManagement:  true,
		DatabaseExport:  true,
		DatabaseImport:  true,
		Deployment:      true,
		Backups:         true,
		RemoteExecution: true,
	}

	if err := RegisterProvider("basic", basic); err != nil {
		t.Fatalf("failed to register basic: %v", err)
	}
	if err := RegisterProvider("full", full); err != nil {
		t.Fatalf("failed to register full: %v", err)
	}

	// Ask for provider with many capabilities
	requirements := []string{
		"authentication", "database_export", "deployment", "backups",
	}

	recommendation, err := GetProviderRecommendation(requirements)
	if err != nil {
		t.Fatalf("GetProviderRecommendation failed: %v", err)
	}

	if recommendation != "full" {
		t.Errorf("expected full provider recommendation, got %s", recommendation)
	}
}

func TestGetProviderRecommendation_NoMatch(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// No providers registered
	_, err := GetProviderRecommendation([]string{"authentication"})
	if err == nil {
		t.Error("expected error when no providers match")
	}
}

func TestProviderInfo_Fields(t *testing.T) {
	info := ProviderInfo{
		Name:        "test-provider",
		Description: "Test provider",
		Capabilities: ProviderCapabilities{
			Authentication: true,
		},
		IsDefault: true,
	}

	if info.Name != "test-provider" {
		t.Errorf("unexpected Name: %s", info.Name)
	}
	if !info.IsDefault {
		t.Error("expected IsDefault to be true")
	}
}

func TestProviderComparison_Fields(t *testing.T) {
	comparison := ProviderComparison{
		Provider1: "p1",
		Provider2: "p2",
		Capabilities1: ProviderCapabilities{
			Authentication: true,
		},
		Capabilities2: ProviderCapabilities{
			Authentication: true,
			Deployment:     true,
		},
		SharedFeatures: []string{"authentication"},
	}

	if comparison.Provider1 != "p1" {
		t.Errorf("unexpected Provider1: %s", comparison.Provider1)
	}
	if !comparison.Capabilities2.Deployment {
		t.Error("expected Capabilities2.Deployment to be true")
	}
	if len(comparison.SharedFeatures) != 1 {
		t.Errorf("expected 1 shared feature, got %d", len(comparison.SharedFeatures))
	}
}

func TestMigrateOptions_Fields(t *testing.T) {
	site := &Site{ID: "site-1"}
	target := newMockProvider("target")

	opts := MigrateOptions{
		SourceSite:      site,
		TargetProvider:  target,
		TargetSiteName:  "new-site",
		IncludeDatabase: true,
		IncludeFiles:    true,
		DryRun:          false,
	}

	if opts.SourceSite.ID != "site-1" {
		t.Errorf("unexpected SourceSite ID: %s", opts.SourceSite.ID)
	}
	if opts.TargetSiteName != "new-site" {
		t.Errorf("unexpected TargetSiteName: %s", opts.TargetSiteName)
	}
	if !opts.IncludeDatabase {
		t.Error("expected IncludeDatabase to be true")
	}
}
