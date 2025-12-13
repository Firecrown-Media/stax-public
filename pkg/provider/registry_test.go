package provider

import (
	"sync"
	"testing"
)

func TestRegisterProvider(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	mock := newMockProvider("test-provider")
	err := RegisterProvider("test-provider", mock)
	if err != nil {
		t.Fatalf("RegisterProvider failed: %v", err)
	}

	// Verify provider was registered
	if !ProviderExists("test-provider") {
		t.Error("expected provider to exist after registration")
	}
}

func TestRegisterProvider_EmptyName(t *testing.T) {
	mock := newMockProvider("")
	err := RegisterProvider("", mock)
	if err == nil {
		t.Error("expected error for empty provider name")
	}
}

func TestRegisterProvider_NilProvider(t *testing.T) {
	err := RegisterProvider("nil-provider", nil)
	if err == nil {
		t.Error("expected error for nil provider")
	}
}

func TestRegisterProvider_Duplicate(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	mock1 := newMockProvider("duplicate")
	mock2 := newMockProvider("duplicate")

	err := RegisterProvider("duplicate", mock1)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	err = RegisterProvider("duplicate", mock2)
	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestGetProvider(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	mock := newMockProvider("test-provider")
	if err := RegisterProvider("test-provider", mock); err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	provider, err := GetProvider("test-provider")
	if err != nil {
		t.Fatalf("GetProvider failed: %v", err)
	}

	if provider.Name() != "test-provider" {
		t.Errorf("expected test-provider, got %s", provider.Name())
	}
}

func TestGetProvider_NotFound(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	_, err := GetProvider("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestListProviders(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register multiple providers
	providers := []string{"charlie", "alpha", "beta"}
	for _, name := range providers {
		mock := newMockProvider(name)
		if err := RegisterProvider(name, mock); err != nil {
			t.Fatalf("failed to register %s: %v", name, err)
		}
	}

	list := ListProviders()

	if len(list) != 3 {
		t.Errorf("expected 3 providers, got %d", len(list))
	}

	// Verify sorted order
	expected := []string{"alpha", "beta", "charlie"}
	for i, name := range expected {
		if list[i] != name {
			t.Errorf("expected %s at position %d, got %s", name, i, list[i])
		}
	}
}

func TestListProviders_Empty(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	list := ListProviders()
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d providers", len(list))
	}
}

func TestGetAllProviders(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register providers
	mock1 := newMockProvider("provider-1")
	mock2 := newMockProvider("provider-2")
	if err := RegisterProvider("provider-1", mock1); err != nil {
		t.Fatalf("failed to register provider-1: %v", err)
	}
	if err := RegisterProvider("provider-2", mock2); err != nil {
		t.Fatalf("failed to register provider-2: %v", err)
	}

	all := GetAllProviders()

	if len(all) != 2 {
		t.Errorf("expected 2 providers, got %d", len(all))
	}

	if _, ok := all["provider-1"]; !ok {
		t.Error("expected provider-1 in result")
	}
	if _, ok := all["provider-2"]; !ok {
		t.Error("expected provider-2 in result")
	}
}

func TestGetAllProviders_ReturnsCopy(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	mock := newMockProvider("test-provider")
	if err := RegisterProvider("test-provider", mock); err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	all := GetAllProviders()

	// Modify the returned map
	delete(all, "test-provider")

	// Verify original registry is unchanged
	if !ProviderExists("test-provider") {
		t.Error("modifying returned map should not affect registry")
	}
}

func TestProviderExists(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	mock := newMockProvider("existing")
	if err := RegisterProvider("existing", mock); err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	if !ProviderExists("existing") {
		t.Error("expected existing provider to exist")
	}

	if ProviderExists("nonexistent") {
		t.Error("expected nonexistent provider to not exist")
	}
}

func TestSetDefaultProvider(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	oldDefault := registry.defaultProvider
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
		registry.defaultProvider = oldDefault
	}()

	mock := newMockProvider("new-default")
	if err := RegisterProvider("new-default", mock); err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	err := SetDefaultProvider("new-default")
	if err != nil {
		t.Fatalf("SetDefaultProvider failed: %v", err)
	}

	if GetDefaultProvider() != "new-default" {
		t.Errorf("expected new-default, got %s", GetDefaultProvider())
	}
}

func TestSetDefaultProvider_NotFound(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	err := SetDefaultProvider("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestGetDefaultProvider(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	oldDefault := registry.defaultProvider
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
		registry.defaultProvider = oldDefault
	}()

	// With no explicit default, should return constant
	result := GetDefaultProvider()
	if result != DefaultProvider {
		t.Errorf("expected %s, got %s", DefaultProvider, result)
	}
}

func TestUnregisterProvider(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	mock := newMockProvider("to-remove")
	if err := RegisterProvider("to-remove", mock); err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	if !ProviderExists("to-remove") {
		t.Fatal("provider should exist before unregister")
	}

	err := UnregisterProvider("to-remove")
	if err != nil {
		t.Fatalf("UnregisterProvider failed: %v", err)
	}

	if ProviderExists("to-remove") {
		t.Error("provider should not exist after unregister")
	}
}

func TestUnregisterProvider_NotFound(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	err := UnregisterProvider("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestClearRegistry(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	oldDefault := registry.defaultProvider
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
		registry.defaultProvider = oldDefault
	}()

	// Register some providers
	mock1 := newMockProvider("provider-1")
	mock2 := newMockProvider("provider-2")
	RegisterProvider("provider-1", mock1)
	RegisterProvider("provider-2", mock2)
	registry.defaultProvider = "provider-1"

	// Clear
	ClearRegistry()

	if len(ListProviders()) != 0 {
		t.Error("expected empty registry after clear")
	}

	if registry.defaultProvider != "" {
		t.Error("expected default provider to be cleared")
	}
}

func TestGetProviderInfo(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	mock := newMockProvider("test-provider")
	mock.description = "Test description"
	mock.capabilities = ProviderCapabilities{
		Authentication: true,
		SiteManagement: true,
	}

	if err := RegisterProvider("test-provider", mock); err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	info, err := GetProviderInfo("test-provider")
	if err != nil {
		t.Fatalf("GetProviderInfo failed: %v", err)
	}

	if info.Name != "test-provider" {
		t.Errorf("expected test-provider, got %s", info.Name)
	}
	if info.Description != "Test description" {
		t.Errorf("unexpected description: %s", info.Description)
	}
	if !info.Capabilities.Authentication {
		t.Error("expected Authentication to be true")
	}
}

func TestGetProviderInfo_NotFound(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	_, err := GetProviderInfo("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestGetAllProviderInfo(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register multiple providers
	mock1 := newMockProvider("provider-1")
	mock2 := newMockProvider("provider-2")
	if err := RegisterProvider("provider-1", mock1); err != nil {
		t.Fatalf("failed to register provider-1: %v", err)
	}
	if err := RegisterProvider("provider-2", mock2); err != nil {
		t.Fatalf("failed to register provider-2: %v", err)
	}

	infos, err := GetAllProviderInfo()
	if err != nil {
		t.Fatalf("GetAllProviderInfo failed: %v", err)
	}

	if len(infos) != 2 {
		t.Errorf("expected 2 infos, got %d", len(infos))
	}
}

func TestGetAllProviderInfo_Empty(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	infos, err := GetAllProviderInfo()
	if err != nil {
		t.Fatalf("GetAllProviderInfo failed: %v", err)
	}

	if len(infos) != 0 {
		t.Errorf("expected empty infos, got %d", len(infos))
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Test concurrent registration and access
	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mock := newMockProvider("concurrent")
			// This may fail due to duplicate, that's expected
			RegisterProvider("concurrent", mock)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ListProviders()
			GetAllProviders()
			ProviderExists("concurrent")
		}()
	}

	wg.Wait()

	// Verify at least one registration succeeded
	// (others would fail as duplicates)
	if !ProviderExists("concurrent") {
		t.Error("expected at least one concurrent registration to succeed")
	}
}

func TestDefaultProviderConstant(t *testing.T) {
	// Verify the default provider constant
	if DefaultProvider != "wpengine" {
		t.Errorf("expected DefaultProvider to be wpengine, got %s", DefaultProvider)
	}
}
