package diagnostics

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/firecrown-media/stax/pkg/credentials"
	"github.com/firecrown-media/stax/pkg/system"
)

// CheckPorts checks if required ports are available
func CheckPorts() CheckResult {
	defaultPorts := system.DefaultDDEVPorts()
	inUse, err := system.CheckRequiredPorts(defaultPorts)
	if err != nil {
		recommendations := system.RecommendedPorts(defaultPorts)
		var suggestions []string
		for original, recommended := range recommendations {
			suggestions = append(suggestions, fmt.Sprintf("Port %d -> %d", original, recommended))
		}

		return CheckResult{
			Name:       "Port Availability",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    fmt.Sprintf("Some required ports are in use: %v", inUse),
			Suggestion: fmt.Sprintf("Consider using alternative ports:\n%s", strings.Join(suggestions, "\n")),
			Details: map[string]string{
				"ports_in_use": fmt.Sprintf("%v", inUse),
			},
		}
	}

	return CheckResult{
		Name:     "Port Availability",
		Category: "Network Connectivity",
		Status:   StatusPass,
		Message:  "All required ports are available",
		Details: map[string]string{
			"checked_ports": fmt.Sprintf("%v", defaultPorts),
		},
	}
}

// CheckWPEngineAPI checks WPEngine API connectivity
func CheckWPEngineAPI() CheckResult {
	creds, err := credentials.GetWPEngineCredentials("default")
	if err != nil {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "WPEngine credentials not configured",
			Suggestion: "Configure credentials: stax setup wpengine",
		}
	}

	// Try to make a simple API request to test connectivity
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.wpengineapi.com/v1/installs", nil)
	if err != nil {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "Failed to create API request",
			Suggestion: "Check your network connection",
		}
	}

	req.SetBasicAuth(creds.APIUser, creds.APIPassword)
	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "Cannot reach WPEngine API",
			Suggestion: "Check your internet connection and firewall settings",
			Details: map[string]string{
				"error": err.Error(),
			},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusFail,
			Message:    "WPEngine API credentials are invalid",
			Suggestion: "Reconfigure credentials: stax setup wpengine",
			Details: map[string]string{
				"status_code": fmt.Sprintf("%d", resp.StatusCode),
			},
		}
	}

	if resp.StatusCode != 200 {
		return CheckResult{
			Name:       "WPEngine API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    fmt.Sprintf("WPEngine API returned status %d", resp.StatusCode),
			Suggestion: "Check WPEngine API status",
			Details: map[string]string{
				"status_code": fmt.Sprintf("%d", resp.StatusCode),
			},
		}
	}

	return CheckResult{
		Name:     "WPEngine API",
		Category: "Network Connectivity",
		Status:   StatusPass,
		Message:  "WPEngine API is reachable and credentials are valid",
		Details: map[string]string{
			"api_user": creds.APIUser,
		},
	}
}

// CheckWPEngineSSH checks WPEngine SSH gateway connectivity
func CheckWPEngineSSH() CheckResult {
	creds, err := credentials.GetWPEngineCredentials("default")
	if err != nil {
		return CheckResult{
			Name:       "WPEngine SSH Gateway",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "WPEngine credentials not configured",
			Suggestion: "Configure credentials: stax setup wpengine",
		}
	}

	gateway := creds.SSHGateway
	if gateway == "" {
		gateway = "ssh.wpengine.net"
	}

	// Test SSH gateway connectivity (port 22)
	conn, err := net.DialTimeout("tcp", gateway+":22", 5*time.Second)
	if err != nil {
		return CheckResult{
			Name:       "WPEngine SSH Gateway",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    fmt.Sprintf("Cannot reach SSH gateway: %s", gateway),
			Suggestion: "Check your internet connection and firewall settings",
			Details: map[string]string{
				"gateway": gateway,
				"error":   err.Error(),
			},
		}
	}
	conn.Close()

	return CheckResult{
		Name:     "WPEngine SSH Gateway",
		Category: "Network Connectivity",
		Status:   StatusPass,
		Message:  "SSH gateway is reachable",
		Details: map[string]string{
			"gateway": gateway,
			"port":    "22",
		},
	}
}

// CheckInternetConnectivity checks basic internet connectivity
func CheckInternetConnectivity() CheckResult {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Try to reach a few reliable endpoints
	endpoints := []string{
		"https://www.google.com",
		"https://www.cloudflare.com",
		"https://api.github.com",
	}

	var lastErr error
	for _, endpoint := range endpoints {
		resp, err := client.Get(endpoint)
		if err == nil {
			resp.Body.Close()
			return CheckResult{
				Name:     "Internet Connectivity",
				Category: "Network Connectivity",
				Status:   StatusPass,
				Message:  "Internet connection is working",
				Details: map[string]string{
					"tested_endpoint": endpoint,
				},
			}
		}
		lastErr = err
	}

	return CheckResult{
		Name:       "Internet Connectivity",
		Category:   "Network Connectivity",
		Status:     StatusWarning,
		Message:    "Cannot reach internet",
		Suggestion: "Check your network connection",
		Details: map[string]string{
			"error": lastErr.Error(),
		},
	}
}

// CheckGitHubAPI checks GitHub API accessibility
func CheckGitHubAPI() CheckResult {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://api.github.com")
	if err != nil {
		return CheckResult{
			Name:       "GitHub API",
			Category:   "Network Connectivity",
			Status:     StatusWarning,
			Message:    "Cannot reach GitHub API",
			Suggestion: "Check your internet connection and firewall settings",
			Details: map[string]string{
				"error": err.Error(),
			},
		}
	}
	defer resp.Body.Close()

	return CheckResult{
		Name:     "GitHub API",
		Category: "Network Connectivity",
		Status:   StatusPass,
		Message:  "GitHub API is accessible",
		Details: map[string]string{
			"status_code": fmt.Sprintf("%d", resp.StatusCode),
		},
	}
}
