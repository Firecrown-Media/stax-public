package provider

import (
	"fmt"
	"sort"
	"sync"
)

var (
	// Global provider registry
	registry = &providerRegistry{
		providers: make(map[string]Provider),
	}

	// DefaultProvider is the default provider name (wpengine for Firecrown)
	DefaultProvider = "wpengine"
)

// providerRegistry manages registered providers
type providerRegistry struct {
	mu              sync.RWMutex
	providers       map[string]Provider
	defaultProvider string
}

// RegisterProvider registers a provider with the global registry
// This should be called from provider package init() functions
func RegisterProvider(name string, provider Provider) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	// Check if provider already registered
	if _, exists := registry.providers[name]; exists {
		return fmt.Errorf("provider %s is already registered", name)
	}

	registry.providers[name] = provider
	return nil
}

// GetProvider retrieves a provider by name from the global registry
func GetProvider(name string) (Provider, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	provider, exists := registry.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// ListProviders returns a sorted list of all registered provider names
func ListProviders() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	names := make([]string, 0, len(registry.providers))
	for name := range registry.providers {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// GetAllProviders returns all registered providers
func GetAllProviders() map[string]Provider {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	// Return a copy to prevent external modification
	providers := make(map[string]Provider, len(registry.providers))
	for name, provider := range registry.providers {
		providers[name] = provider
	}

	return providers
}

// ProviderExists checks if a provider with the given name is registered
func ProviderExists(name string) bool {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	_, exists := registry.providers[name]
	return exists
}

// SetDefaultProvider sets the default provider name
func SetDefaultProvider(name string) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	// Check existence without calling ProviderExists to avoid deadlock
	// (ProviderExists also tries to acquire a lock)
	if _, exists := registry.providers[name]; !exists {
		return fmt.Errorf("cannot set default: provider %s not found", name)
	}

	registry.defaultProvider = name
	return nil
}

// GetDefaultProvider returns the default provider name
// Falls back to DefaultProvider constant if not explicitly set
func GetDefaultProvider() string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	if registry.defaultProvider != "" {
		return registry.defaultProvider
	}

	return DefaultProvider
}

// UnregisterProvider removes a provider from the registry
// This is primarily useful for testing
func UnregisterProvider(name string) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.providers[name]; !exists {
		return fmt.Errorf("provider %s not registered", name)
	}

	delete(registry.providers, name)
	return nil
}

// ClearRegistry removes all providers from the registry
// This is primarily useful for testing
func ClearRegistry() {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.providers = make(map[string]Provider)
	registry.defaultProvider = ""
}

// ProviderInfo contains metadata about a registered provider
type ProviderInfo struct {
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Capabilities ProviderCapabilities `json:"capabilities"`
	IsDefault    bool                 `json:"is_default"`
}

// GetProviderInfo returns information about a specific provider
func GetProviderInfo(name string) (*ProviderInfo, error) {
	provider, err := GetProvider(name)
	if err != nil {
		return nil, err
	}

	return &ProviderInfo{
		Name:         provider.Name(),
		Description:  provider.Description(),
		Capabilities: provider.Capabilities(),
		IsDefault:    name == GetDefaultProvider(),
	}, nil
}

// GetAllProviderInfo returns information about all registered providers
func GetAllProviderInfo() ([]*ProviderInfo, error) {
	names := ListProviders()
	infos := make([]*ProviderInfo, 0, len(names))

	for _, name := range names {
		info, err := GetProviderInfo(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get info for provider %s: %w", name, err)
		}
		infos = append(infos, info)
	}

	return infos, nil
}
