package provider

import (
	"fmt"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/credentials"
)

// configAuthenticator is satisfied by providers that can accept a full
// ProviderConfig map rather than individual credential strings.
type configAuthenticator interface {
	AuthenticateFromConfig(providerConfig map[string]any, credentials map[string]string) error
}

// NewAuthenticatedProvider resolves cfg.Provider from the registry, fetches
// credentials from the keychain/env/file fallback, and authenticates.
func NewAuthenticatedProvider(cfg *config.Config) (Provider, error) {
	providerName := cfg.Provider
	if providerName == "" {
		providerName = "wpengine"
	}

	p, err := GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("unknown provider %q: %w", providerName, err)
	}

	install := ""
	if v, ok := cfg.ProviderConfig["install"]; ok {
		install, _ = v.(string)
	}

	creds, err := credentials.GetWPEngineCredentialsWithFallback(install)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	sshKey, _ := credentials.GetSSHPrivateKeyWithFallback(providerName)

	credMap := map[string]string{
		"api_user":     creds.APIUser,
		"api_password": creds.APIPassword,
		"ssh_key":      sshKey,
	}

	if ca, ok := p.(configAuthenticator); ok {
		if err := ca.AuthenticateFromConfig(cfg.ProviderConfig, credMap); err != nil {
			return nil, fmt.Errorf("provider authentication failed: %w", err)
		}
	} else {
		if err := p.Authenticate(credMap); err != nil {
			return nil, fmt.Errorf("provider authentication failed: %w", err)
		}
	}

	return p, nil
}
