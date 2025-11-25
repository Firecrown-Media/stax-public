package wpengine

import (
	"fmt"
)

// GetInstallEnvironment queries WPEngine API for install environment type
func (c *Client) GetInstallEnvironment(installName string) (string, error) {
	// Get the full install details which includes environment
	details, err := c.GetInstallByName(installName)
	if err != nil {
		return "", fmt.Errorf("failed to get install info: %w", err)
	}

	// Return the environment field
	return details.Environment, nil
}
