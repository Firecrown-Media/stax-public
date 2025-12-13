package provider

import (
	"fmt"
	"os"
	"testing"
)

func TestProviderConfig_Fields(t *testing.T) {
	config := ProviderConfig{
		Name: "wpengine",
		Credentials: map[string]string{
			"api_token": "test-token",
			"ssh_key":   "/path/to/key",
		},
		Options: map[string]string{
			"timeout": "30",
		},
	}

	if config.Name != "wpengine" {
		t.Errorf("expected Name to be wpengine, got %s", config.Name)
	}
	if config.Credentials["api_token"] != "test-token" {
		t.Errorf("unexpected Credentials: %v", config.Credentials)
	}
	if config.Options["timeout"] != "30" {
		t.Errorf("unexpected Options: %v", config.Options)
	}
}

func TestNewProvider_EmptyName(t *testing.T) {
	config := ProviderConfig{
		Name: "",
	}

	_, err := NewProvider(config)
	if err == nil {
		t.Error("expected error for empty provider name")
	}
}

func TestNewProvider_UnregisteredProvider(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	config := ProviderConfig{
		Name: "nonexistent",
	}

	_, err := NewProvider(config)
	if err == nil {
		t.Error("expected error for unregistered provider")
	}
}

func TestNewProvider_WithRegisteredProvider(t *testing.T) {
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

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	if provider.Name() != "test-provider" {
		t.Errorf("expected provider name test-provider, got %s", provider.Name())
	}
}

func TestNewProvider_WithCredentials(t *testing.T) {
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
		Credentials: map[string]string{
			"api_key": "secret-key",
		},
	}

	_, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	// Verify authenticate was called
	if !mock.authenticateCalled {
		t.Error("expected Authenticate to be called")
	}
	if mock.lastCredentials["api_key"] != "secret-key" {
		t.Errorf("unexpected credentials passed: %v", mock.lastCredentials)
	}
}

func TestNewProvider_AuthenticationFails(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register a mock provider that fails authentication
	mock := newMockProvider("test-provider")
	mock.authenticateError = fmt.Errorf("invalid credentials")
	if err := RegisterProvider("test-provider", mock); err != nil {
		t.Fatalf("failed to register mock provider: %v", err)
	}

	config := ProviderConfig{
		Name: "test-provider",
		Credentials: map[string]string{
			"api_key": "bad-key",
		},
	}

	_, err := NewProvider(config)
	if err == nil {
		t.Error("expected error when authentication fails")
	}
}

func TestNewProviderFromName_EmptyName(t *testing.T) {
	_, err := NewProviderFromName("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestNewProviderFromName_Success(t *testing.T) {
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

	provider, err := NewProviderFromName("test-provider")
	if err != nil {
		t.Fatalf("NewProviderFromName failed: %v", err)
	}

	if provider.Name() != "test-provider" {
		t.Errorf("expected provider name test-provider, got %s", provider.Name())
	}

	// Verify authenticate was NOT called (no credentials provided)
	if mock.authenticateCalled {
		t.Error("Authenticate should not be called without credentials")
	}
}

func TestDetectProviderFromConfig_ExplicitName(t *testing.T) {
	config := map[string]interface{}{
		"name": "my-provider",
	}

	name, err := DetectProviderFromConfig(config)
	if err != nil {
		t.Fatalf("DetectProviderFromConfig failed: %v", err)
	}

	if name != "my-provider" {
		t.Errorf("expected name my-provider, got %s", name)
	}
}

func TestDetectProviderFromConfig_KnownProvider(t *testing.T) {
	// Test each known provider
	knownProviders := []string{"wpengine", "aws", "wordpress-vip", "local"}

	for _, provider := range knownProviders {
		config := map[string]interface{}{
			provider: map[string]string{"key": "value"},
		}

		name, err := DetectProviderFromConfig(config)
		if err != nil {
			t.Fatalf("DetectProviderFromConfig failed for %s: %v", provider, err)
		}

		if name != provider {
			t.Errorf("expected name %s, got %s", provider, name)
		}
	}
}

func TestDetectProviderFromConfig_Default(t *testing.T) {
	config := map[string]interface{}{
		"some_other_key": "value",
	}

	name, err := DetectProviderFromConfig(config)
	if err != nil {
		t.Fatalf("DetectProviderFromConfig failed: %v", err)
	}

	// Should return default provider
	expected := GetDefaultProvider()
	if name != expected {
		t.Errorf("expected default provider %s, got %s", expected, name)
	}
}

func TestResolveProvider_ExplicitName(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register a provider
	mock := newMockProvider("explicit-provider")
	if err := RegisterProvider("explicit-provider", mock); err != nil {
		t.Fatalf("failed to register mock provider: %v", err)
	}

	name, err := ResolveProvider("explicit-provider", nil, "")
	if err != nil {
		t.Fatalf("ResolveProvider failed: %v", err)
	}

	if name != "explicit-provider" {
		t.Errorf("expected explicit-provider, got %s", name)
	}
}

func TestResolveProvider_ExplicitNameNotFound(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	_, err := ResolveProvider("nonexistent", nil, "")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestResolveProvider_EnvironmentVariable(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
		os.Unsetenv("STAX_PROVIDER")
	}()

	// Register a provider
	mock := newMockProvider("env-provider")
	if err := RegisterProvider("env-provider", mock); err != nil {
		t.Fatalf("failed to register mock provider: %v", err)
	}

	// Set environment variable
	os.Setenv("STAX_PROVIDER", "env-provider")

	name, err := ResolveProvider("", nil, "")
	if err != nil {
		t.Fatalf("ResolveProvider failed: %v", err)
	}

	if name != "env-provider" {
		t.Errorf("expected env-provider, got %s", name)
	}
}

func TestResolveProvider_GlobalDefault(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
		os.Unsetenv("STAX_PROVIDER")
	}()

	// Register a provider
	mock := newMockProvider("global-default")
	if err := RegisterProvider("global-default", mock); err != nil {
		t.Fatalf("failed to register mock provider: %v", err)
	}

	// Unset environment variable to test next priority
	os.Unsetenv("STAX_PROVIDER")

	name, err := ResolveProvider("", nil, "global-default")
	if err != nil {
		t.Fatalf("ResolveProvider failed: %v", err)
	}

	if name != "global-default" {
		t.Errorf("expected global-default, got %s", name)
	}
}

func TestValidateProviderConfig_EmptyName(t *testing.T) {
	config := ProviderConfig{Name: ""}
	err := ValidateProviderConfig(config)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestValidateProviderConfig_UnknownProvider(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	config := ProviderConfig{Name: "unknown"}
	err := ValidateProviderConfig(config)
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestValidateProviderConfig_ValidConfig(t *testing.T) {
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

	err := ValidateProviderConfig(config)
	if err != nil {
		t.Errorf("ValidateProviderConfig should succeed: %v", err)
	}
}

func TestGetProviderCapabilities(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register a mock provider with specific capabilities
	mock := newMockProvider("test-provider")
	mock.capabilities = ProviderCapabilities{
		Authentication: true,
		SiteManagement: true,
		DatabaseExport: true,
		Deployment:     false,
	}
	if err := RegisterProvider("test-provider", mock); err != nil {
		t.Fatalf("failed to register mock provider: %v", err)
	}

	caps, err := GetProviderCapabilities("test-provider")
	if err != nil {
		t.Fatalf("GetProviderCapabilities failed: %v", err)
	}

	if !caps.Authentication {
		t.Error("expected Authentication to be true")
	}
	if !caps.SiteManagement {
		t.Error("expected SiteManagement to be true")
	}
	if caps.Deployment {
		t.Error("expected Deployment to be false")
	}
}

func TestSupportsCapability(t *testing.T) {
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
	mock.capabilities = ProviderCapabilities{
		Authentication: true,
		SiteManagement: true,
		Deployment:     false,
	}
	if err := RegisterProvider("test-provider", mock); err != nil {
		t.Fatalf("failed to register mock provider: %v", err)
	}

	tests := []struct {
		capability string
		expected   bool
	}{
		{"authentication", true},
		{"site_management", true},
		{"deployment", false},
		{"database_export", false},
	}

	for _, tt := range tests {
		t.Run(tt.capability, func(t *testing.T) {
			result, err := SupportsCapability("test-provider", tt.capability)
			if err != nil {
				t.Fatalf("SupportsCapability failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v for %s, got %v", tt.expected, tt.capability, result)
			}
		})
	}
}

func TestSupportsCapability_AllCapabilities(t *testing.T) {
	// Save and restore registry state
	oldProviders := GetAllProviders()
	ClearRegistry()
	defer func() {
		ClearRegistry()
		for name, p := range oldProviders {
			RegisterProvider(name, p)
		}
	}()

	// Register a mock provider with all capabilities enabled
	mock := newMockProvider("full-provider")
	mock.capabilities = ProviderCapabilities{
		Authentication:  true,
		SiteManagement:  true,
		DatabaseExport:  true,
		DatabaseImport:  true,
		FileSync:        true,
		Deployment:      true,
		Environments:    true,
		Backups:         true,
		RemoteExecution: true,
		MediaManagement: true,
		SSHAccess:       true,
		APIAccess:       true,
		Scaling:         true,
		Monitoring:      true,
		Logging:         true,
	}
	if err := RegisterProvider("full-provider", mock); err != nil {
		t.Fatalf("failed to register mock provider: %v", err)
	}

	allCapabilities := []string{
		"authentication", "site_management", "database_export", "database_import",
		"file_sync", "deployment", "environments", "backups",
		"remote_execution", "media_management", "ssh_access", "api_access",
		"scaling", "monitoring", "logging",
	}

	for _, cap := range allCapabilities {
		result, err := SupportsCapability("full-provider", cap)
		if err != nil {
			t.Errorf("SupportsCapability failed for %s: %v", cap, err)
		}
		if !result {
			t.Errorf("expected %s to be supported", cap)
		}
	}
}

func TestSupportsCapability_UnknownCapability(t *testing.T) {
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
		t.Fatalf("failed to register mock provider: %v", err)
	}

	_, err := SupportsCapability("test-provider", "unknown_capability")
	if err == nil {
		t.Error("expected error for unknown capability")
	}
}

func TestCreateProviderFromResolution(t *testing.T) {
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

	provider, err := CreateProviderFromResolution(
		"test-provider",
		nil,
		"",
		map[string]string{"key": "value"},
	)

	if err != nil {
		t.Fatalf("CreateProviderFromResolution failed: %v", err)
	}

	if provider.Name() != "test-provider" {
		t.Errorf("expected test-provider, got %s", provider.Name())
	}

	// Verify credentials were passed
	if mock.lastCredentials["key"] != "value" {
		t.Errorf("unexpected credentials: %v", mock.lastCredentials)
	}
}
