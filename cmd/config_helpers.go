package cmd

import "github.com/firecrown-media/stax/pkg/config"

// wpeInstall returns the WPEngine install name from provider_config.
func wpeInstall(cfg *config.Config) string {
	v, _ := cfg.ProviderConfig["install"].(string)
	return v
}

// wpeEnv returns the WPEngine environment from provider_config.
func wpeEnv(cfg *config.Config) string {
	v, _ := cfg.ProviderConfig["environment"].(string)
	return v
}

// wpeSSHGateway returns the WPEngine SSH gateway from provider_config, defaulting to ssh.wpengine.net.
func wpeSSHGateway(cfg *config.Config) string {
	if v, ok := cfg.ProviderConfig["ssh_gateway"].(string); ok && v != "" {
		return v
	}
	return "ssh.wpengine.net"
}
