# Stax Development Patterns

This document describes the coding patterns and conventions used throughout the Stax codebase.

## Error Handling Patterns

### Enhanced Errors

Stax uses enhanced errors (defined in `pkg/errors/enhanced.go`) for user-facing errors that need actionable solutions:

```go
import "github.com/firecrown-media/stax/pkg/errors"

// Create an enhanced error with solutions
return errors.NewCredentialsNotFoundError(
    triedLocations,  // []string of what was attempted
    underlyingErr,   // Original error (can be nil)
)
```

Enhanced errors include:
- **Error codes** (e.g., `STX-001`) for easy reference
- **Detailed messages** explaining what went wrong
- **Tried locations** showing what was attempted
- **Solutions** with commands or steps to resolve

### Error Codes

| Code | Type | Description |
|------|------|-------------|
| STX-001 | ConfigNotFound | `.stax.yml` not found |
| STX-002 | CredentialsNotFound | WPEngine credentials not in keychain |
| STX-003 | SSHKeyNotFound | SSH private key not found |
| STX-004 | DDEVNotInstalled | DDEV CLI not found |
| STX-005 | DDEVNotConfigured | DDEV project not configured |
| STX-006 | CommandNotImplemented | Feature not yet implemented |
| STX-007 | InvalidConfig | Configuration validation failed |
| STX-008 | WPEngineAPI | WPEngine API error |

### Standard Error Wrapping

For internal errors, use standard Go error wrapping:

```go
if err != nil {
    return fmt.Errorf("failed to read config: %w", err)
}
```

## Credential Retrieval Patterns

### Fallback Chain

Credentials are retrieved with a fallback chain to support multiple configurations:

```go
import "github.com/firecrown-media/stax/pkg/credentials"

// Try multiple sources in order
creds, err := credentials.GetWPEngineCredentialsWithFallback(installName)
if err != nil {
    if credErr, ok := err.(*credentials.CredentialsNotFoundError); ok {
        return errors.NewCredentialsNotFoundError(credErr.Tried, credErr.LastErr)
    }
    return fmt.Errorf("failed to get credentials: %w", err)
}
```

### SSH Key Retrieval

```go
sshKey, err := credentials.GetSSHPrivateKeyWithFallback("wpengine")
if err != nil {
    if keyErr, ok := err.(*credentials.SSHKeyNotFoundError); ok {
        return errors.NewSSHKeyNotFoundError("", keyErr.Tried, keyErr.LastErr)
    }
    return fmt.Errorf("failed to get SSH key: %w", err)
}
```

### Keychain Storage

Credentials are stored in macOS Keychain:
- Service: `com.firecrownmedia.stax`
- Account pattern: `{install}.{credential-type}`

```go
// Store credential
err := credentials.Set(install, "ssh-password", password)

// Retrieve credential
password, err := credentials.Get(install, "ssh-password")

// Delete credential
err := credentials.Delete(install, "ssh-password")
```

## DDEV Integration Patterns

### Manager Pattern

All DDEV operations go through the Manager:

```go
import "github.com/firecrown-media/stax/pkg/ddev"

mgr := ddev.NewManager(projectDir)

// Check if running with retry logic
if err := mgr.RequireRunning(); err != nil {
    return err
}

// Execute commands in container
output, err := mgr.Exec([]string{"wp", "option", "get", "siteurl"}, nil)
```

### RequireRunning Pattern

For commands that need DDEV to be running, use `RequireRunning()`:

```go
mgr := ddev.NewManager(projectDir)

// This includes retry logic (3 attempts, 2s delay)
// and returns user-friendly error messages
if err := mgr.RequireRunning(); err != nil {
    return err
}

// Proceed with operation...
```

### WaitForReady Pattern

For operations that start DDEV, wait for services:

```go
if err := mgr.Start(); err != nil {
    return err
}

if err := mgr.WaitForReady(60 * time.Second); err != nil {
    return fmt.Errorf("DDEV did not become ready: %w", err)
}
```

## Provider Pattern

### Interface Implementation

All hosting providers implement the `Provider` interface:

```go
type Provider interface {
    Name() string
    Description() string
    Capabilities() ProviderCapabilities
    Authenticate(credentials map[string]string) error
    TestConnection() error
    ValidateCredentials(credentials map[string]string) error
    ListSites() ([]Site, error)
    GetSite(identifier string) (*Site, error)
    // ... more methods
}
```

### Registry Pattern

Providers register themselves at package init:

```go
func init() {
    provider.RegisterProvider("wpengine", NewWPEngineProvider())
}
```

### Capability Checking

Check provider capabilities before calling methods:

```go
caps := provider.Capabilities()
if !caps.DatabaseExport {
    return fmt.Errorf("provider does not support database export")
}
```

## UI Output Patterns

### Consistent Output Functions

Use the `pkg/ui` package for all user-facing output:

```go
import "github.com/firecrown-media/stax/pkg/ui"

ui.Success("Operation completed")
ui.Error("Something went wrong: %v", err)
ui.Warning("This might cause issues")
ui.Info("Additional information")
ui.Debug("Debug details")  // Only shown with --debug flag
```

### Quiet Mode

All output functions respect `--quiet` flag except `Error`:

```go
// These are suppressed in quiet mode:
ui.Success("Done")
ui.Info("Details")
ui.Warning("Caution")

// This always shows (errors are never suppressed):
ui.Error("Failed")
```

### Spinner for Long Operations

```go
import "github.com/firecrown-media/stax/pkg/ui"

err := ui.WithSpinner("Downloading database...", func() error {
    return downloadDatabase()
})
```

## Configuration Patterns

### Loading Configuration

```go
import "github.com/firecrown-media/stax/pkg/config"

cfg, err := config.Load(projectDir)
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
```

### Validating Config Exists

```go
staxConfigPath := filepath.Join(projectDir, ".stax.yml")
if _, err := os.Stat(staxConfigPath); os.IsNotExist(err) {
    return fmt.Errorf("no .stax.yml found in %s. Please run 'stax init' first", projectDir)
}
```

## Testing Patterns

### Table-Driven Tests

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "hello", "HELLO", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

### Test State Isolation

When testing global state (like registries), save and restore:

```go
func TestRegistryFunction(t *testing.T) {
    // Save state
    oldProviders := GetAllProviders()
    ClearRegistry()
    defer func() {
        ClearRegistry()
        for name, p := range oldProviders {
            RegisterProvider(name, p)
        }
    }()

    // Test with clean registry...
}
```

### Capturing Output

For testing UI functions:

```go
func captureOutput(fn func()) (string, string) {
    oldStdout := os.Stdout
    oldStderr := os.Stderr
    rOut, wOut, _ := os.Pipe()
    rErr, wErr, _ := os.Pipe()
    os.Stdout = wOut
    os.Stderr = wErr

    fn()

    wOut.Close()
    wErr.Close()
    os.Stdout = oldStdout
    os.Stderr = oldStderr

    var bufOut, bufErr bytes.Buffer
    io.Copy(&bufOut, rOut)
    io.Copy(&bufErr, rErr)
    return bufOut.String(), bufErr.String()
}
```

## Command Patterns

### Cobra Command Structure

```go
var myCmd = &cobra.Command{
    Use:   "my-command",
    Short: "Brief description",
    Long:  `Detailed description of what the command does.`,
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)
    myCmd.Flags().StringVar(&myFlag, "my-flag", "", "Flag description")
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Get project directory
    projectDir := getProjectDir()

    // Load config
    cfg, err := config.Load(projectDir)
    if err != nil {
        return err
    }

    // Implementation...
    ui.Success("Command completed!")
    return nil
}
```

### Interactive vs Non-Interactive

Always use safe prompts that handle non-TTY:

```go
import "github.com/firecrown-media/stax/pkg/prompts"

// This returns default in non-interactive mode
input, err := prompts.SafePromptInput("Enter value", "default", true)
```

## File Organization

### Package Structure

```
pkg/
  config/       # Configuration loading and validation
  credentials/  # Keychain credential management
  ddev/         # DDEV integration
  errors/       # Enhanced error types
  provider/     # Provider abstraction layer
  providers/    # Concrete provider implementations
    wpengine/
    aws/
    local/
  ui/           # Terminal output formatting
  wordpress/    # WP-CLI operations
  wpengine/     # WPEngine API client
```

### Test Files

Test files live alongside source files:
- `config.go` -> `config_test.go`
- `manager.go` -> `manager_test.go`

## Concurrency Patterns

### Registry Thread Safety

The provider registry uses `sync.RWMutex`:

```go
// Read operations use RLock
registry.mu.RLock()
defer registry.mu.RUnlock()

// Write operations use Lock
registry.mu.Lock()
defer registry.mu.Unlock()
```

**Important**: Don't call functions that acquire locks while holding a lock (Go mutexes are not reentrant).
