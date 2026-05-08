package migration

import (
	"fmt"
	"sync"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/provider"
)

// SourceFactory creates a Source for a given provider name.
type SourceFactory func(p provider.Provider, cfg *config.Config) Source

// DestinationFactory creates a Destination for a given destination name.
type DestinationFactory func(repoPath string) Destination

var (
	mu           sync.RWMutex
	sources      = map[string]SourceFactory{}
	destinations = map[string]DestinationFactory{}
)

// RegisterSource registers a SourceFactory under the given provider name.
// Call this from provider package init() functions.
// Panics if name is empty or factory is nil (programming error at init time).
func RegisterSource(name string, factory SourceFactory) {
	if name == "" || factory == nil {
		panic(fmt.Sprintf("migration.RegisterSource: invalid registration (name=%q, factory nil=%v)", name, factory == nil))
	}
	mu.Lock()
	defer mu.Unlock()
	sources[name] = factory
}

// RegisterDestination registers a DestinationFactory under the given name.
// Call this from destination package init() functions.
// Panics if name is empty or factory is nil (programming error at init time).
func RegisterDestination(name string, factory DestinationFactory) {
	if name == "" || factory == nil {
		panic(fmt.Sprintf("migration.RegisterDestination: invalid registration (name=%q, factory nil=%v)", name, factory == nil))
	}
	mu.Lock()
	defer mu.Unlock()
	destinations[name] = factory
}

// NewSource returns a Source for the given provider name.
func NewSource(name string, p provider.Provider, cfg *config.Config) (Source, error) {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := sources[name]
	if !ok {
		return nil, fmt.Errorf("no migration source registered for provider %q", name)
	}
	return factory(p, cfg), nil
}

// NewDestination returns a Destination for the given name.
// repoPath is the local path to the VIP repo checkout (used by CompareFiles).
func NewDestination(name string, repoPath string) (Destination, error) {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := destinations[name]
	if !ok {
		return nil, fmt.Errorf("no migration destination registered for %q", name)
	}
	return factory(repoPath), nil
}

// ResetRegistryForTesting clears all registered factories.
// Use in t.Cleanup to prevent test pollution across test files that share this package.
func ResetRegistryForTesting() {
	mu.Lock()
	defer mu.Unlock()
	sources = map[string]SourceFactory{}
	destinations = map[string]DestinationFactory{}
}
