# stax migrate Commands Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the `stax migrate` command group with WPEngine→VIP support, following the provider-agnostic interface design from the migration pipeline spec.

**Architecture:** `pkg/migration/` defines `Source` and `Destination` interfaces with a registry that resolves provider names to implementations at runtime. `WPEngineSource` wraps the existing files and database services; `VIPDestination` shells out to `phpcs` and the VIP CLI. `cmd/migrate.go` is a thin Cobra wrapper that delegates to `pkg/migration/service.go`.

**Tech Stack:** Go 1.24, Cobra, existing `pkg/files`, `pkg/database`, `pkg/provider` packages, `phpcs` (external), VIP CLI (external)

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `pkg/config/config.go` | Add `MigrationConfig` struct and `Migration` field to `Config` |
| Modify | `pkg/wpengine/types.go` | Add `ExtraFlags []string` to `DatabaseOptions` |
| Modify | `pkg/wpengine/database.go` | Append `ExtraFlags` to the `wp db export` command |
| Modify | `pkg/provider/interface.go` | Add `ExtraFlags []string` to `DatabaseExportOptions` |
| Modify | `pkg/providers/wpengine/provider.go` | Pass `ExtraFlags` from `DatabaseExportOptions` to `DatabaseOptions` |
| Create | `pkg/migration/interfaces.go` | `Source`, `Destination` interfaces + option/result types |
| Create | `pkg/migration/registry.go` | Factory registry mapping provider names to Source/Destination |
| Create | `pkg/migration/registry_test.go` | Registry unit tests |
| Create | `pkg/migration/service.go` | `Pull`, `Export`, `Audit`, `Compare`, `Import`, `Report`, `Status` |
| Create | `pkg/migration/service_test.go` | Service unit tests using mock Source/Destination |
| Create | `pkg/migration/providers/wpengine/source.go` | `WPEngineSource` implementing `Source` |
| Create | `pkg/migration/providers/wpengine/source_test.go` | WPEngineSource unit tests |
| Create | `pkg/migration/providers/vip/destination.go` | `VIPDestination` implementing `Destination` |
| Create | `pkg/migration/providers/vip/destination_test.go` | VIPDestination unit tests |
| Create | `cmd/migrate.go` | Cobra command group: pull, export, audit, compare, import, report, status |

---

## Task 1: Add MigrationConfig to pkg/config

**Files:**
- Modify: `pkg/config/config.go`
- Test: `pkg/config/config_test.go`

- [ ] **Step 1: Write the failing test**

Add to `pkg/config/config_test.go`:

```go
func TestConfig_MigrationDestination(t *testing.T) {
	yaml := `
version: 2
provider: wpengine
provider_config:
  install: mysite
  environment: production
migration:
  destination: vip
`
	var cfg Config
	if err := loadYAML([]byte(yaml), &cfg); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Migration.Destination != "vip" {
		t.Errorf("expected migration.destination 'vip', got %q", cfg.Migration.Destination)
	}
}

func TestConfig_MigrationDestination_Empty(t *testing.T) {
	yaml := `
version: 2
provider: wpengine
provider_config:
  install: mysite
`
	var cfg Config
	if err := loadYAML([]byte(yaml), &cfg); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Migration.Destination != "" {
		t.Errorf("expected empty migration.destination, got %q", cfg.Migration.Destination)
	}
}
```

Note: check existing `config_test.go` for the `loadYAML` helper name. If the package uses `yaml.Unmarshal` directly in tests, use that:
```go
var cfg Config
if err := yaml_pkg.Unmarshal([]byte(yamlStr), &cfg); err != nil { ... }
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/geoff/_projects/fc/stax
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/config/... -run TestConfig_Migration -v
```

Expected: FAIL — `cfg.Migration` field does not exist.

- [ ] **Step 3: Add MigrationConfig struct and field**

In `pkg/config/config.go`, after the `PerformanceConfig` struct definition and before the closing of the file, add:

```go
// MigrationConfig holds settings for stax migrate commands.
type MigrationConfig struct {
	Destination string `yaml:"destination"` // e.g. "vip"
}
```

Then add the field to the `Config` struct after `Performance PerformanceConfig`:

```go
// Migration configuration for stax migrate commands.
Migration MigrationConfig `yaml:"migration,omitempty"`
```

- [ ] **Step 4: Run test to verify it passes**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/config/... -run TestConfig_Migration -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/config/config.go pkg/config/config_test.go
git commit -m "feat(config): add MigrationConfig with destination field"
```

---

## Task 2: Add ExtraFlags to the database export pipeline

**Files:**
- Modify: `pkg/wpengine/types.go`
- Modify: `pkg/wpengine/database.go`
- Modify: `pkg/provider/interface.go`
- Modify: `pkg/providers/wpengine/provider.go`

- [ ] **Step 1: Write the failing test**

Add to `pkg/wpengine/database.go`-adjacent test. Create `pkg/wpengine/database_test.go` if it doesn't exist:

```go
package wpengine

import (
	"strings"
	"testing"
)

func TestBuildExportCommand_ExtraFlags(t *testing.T) {
	cmd := buildExportCommand("wp_", DatabaseOptions{
		ExtraFlags: []string{"--hex-blob", "--quote-names", "--default-character-set=utf8mb4"},
	})
	for _, flag := range []string{"--hex-blob", "--quote-names", "--default-character-set=utf8mb4"} {
		if !strings.Contains(cmd, flag) {
			t.Errorf("expected command to contain %q, got: %s", flag, cmd)
		}
	}
}
```

Note: `buildExportCommand` is an extracted helper you'll create in Step 3. If a test for an unexported helper is awkward, write an integration-style test instead that checks the assembled string through the exported logic.

- [ ] **Step 2: Run test to verify it fails**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/wpengine/... -run TestBuildExportCommand -v
```

Expected: FAIL — `buildExportCommand` does not exist.

- [ ] **Step 3: Add ExtraFlags to types and refactor ExportDatabase**

In `pkg/wpengine/types.go`, add `ExtraFlags` to `DatabaseOptions`:

```go
type DatabaseOptions struct {
	ExcludeTables  []string
	SkipLogs       bool
	SkipTransients bool
	SkipSpam       bool
	Compress       bool
	ExtraFlags     []string // additional flags passed to wp db export
}
```

In `pkg/wpengine/database.go`, extract the command-building logic into a helper and use it:

```go
// buildExportCommand assembles the wp db export command string.
// prefix is the WordPress table prefix (for exclusion patterns).
func buildExportCommand(prefix string, options DatabaseOptions) string {
	cmd := "wp db export --add-drop-table"

	excludePattern, _ := GenerateExcludePattern(prefix, options)
	if excludePattern != "" {
		cmd += fmt.Sprintf(" --exclude_tables=%s", excludePattern)
	}

	for _, flag := range options.ExtraFlags {
		cmd += " " + flag
	}

	cmd += " -"
	return cmd
}
```

Then update `ExportDatabase` to call `buildExportCommand`:

```go
func (c *SSHClient) ExportDatabase(options DatabaseOptions) (io.ReadCloser, error) {
	prefix, err := c.GetTablePrefix()
	if err != nil {
		return nil, err
	}

	cmd := buildExportCommand(prefix, options)

	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := session.Start(cmd); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start export command: %w", err)
	}

	return &exportReadCloser{reader: stdout, session: session}, nil
}
```

- [ ] **Step 4: Add ExtraFlags to provider.DatabaseExportOptions**

In `pkg/provider/interface.go`, add to `DatabaseExportOptions`:

```go
type DatabaseExportOptions struct {
	ExcludeTables  []string `json:"exclude_tables"`
	SkipLogs       bool     `json:"skip_logs"`
	SkipTransients bool     `json:"skip_transients"`
	SkipSpam       bool     `json:"skip_spam"`
	Compress       bool     `json:"compress"`
	IncludePrefix  bool     `json:"include_prefix"`
	ExtraFlags     []string `json:"extra_flags"` // passed to wp db export
}
```

- [ ] **Step 5: Pass ExtraFlags through WPEngineProvider**

In `pkg/providers/wpengine/provider.go`, update `ExportDatabase`:

```go
func (p *WPEngineProvider) ExportDatabase(site *provider.Site, options provider.DatabaseExportOptions) (io.ReadCloser, error) {
	if p.sshClient == nil {
		return nil, fmt.Errorf("SSH client not configured (SSH key required)")
	}

	wpOptions := wpengine.DatabaseOptions{
		ExcludeTables:  options.ExcludeTables,
		SkipLogs:       options.SkipLogs,
		SkipTransients: options.SkipTransients,
		SkipSpam:       options.SkipSpam,
		Compress:       options.Compress,
		ExtraFlags:     options.ExtraFlags,
	}

	return p.sshClient.ExportDatabase(wpOptions)
}
```

- [ ] **Step 6: Run tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/wpengine/... ./pkg/provider/... ./pkg/providers/... -v
```

Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add pkg/wpengine/types.go pkg/wpengine/database.go pkg/provider/interface.go pkg/providers/wpengine/provider.go
git commit -m "feat(database): add ExtraFlags to database export for VIP-compatible mysqldump options"
```

---

## Task 3: Create pkg/migration interfaces and registry

**Files:**
- Create: `pkg/migration/interfaces.go`
- Create: `pkg/migration/registry.go`
- Test: `pkg/migration/registry_test.go`

- [ ] **Step 1: Create interfaces.go**

```go
// pkg/migration/interfaces.go
package migration

// PullOptions configures a file pull from the source provider.
type PullOptions struct {
	ExcludeUploads bool
	ThemesOnly     bool
	PluginsOnly    bool
	MuPluginsOnly  bool
	DryRun         bool
	ProjectDir     string
}

// ExportOptions configures a database export from the source provider.
type ExportOptions struct {
	OutputPath string // local file path where the SQL dump is written
}

// AuditOptions configures a phpcs compatibility audit.
type AuditOptions struct {
	Severity int // minimum phpcs severity (1–5); 0 defaults to 1
}

// AuditMessage is a single phpcs finding.
type AuditMessage struct {
	Line     int
	Column   int
	Severity int
	Type     string // "ERROR" or "WARNING"
	Message  string
	Source   string // phpcs sniff identifier
}

// FileAuditResult holds phpcs findings for a single file.
type FileAuditResult struct {
	FilePath string
	Errors   int
	Warnings int
	Messages []AuditMessage
}

// AuditReport aggregates phpcs findings across all scanned paths.
type AuditReport struct {
	Files         []FileAuditResult
	TotalErrors   int
	TotalWarnings int
	GeneratedAt   string
}

// ImportOptions configures a VIP import operation.
type ImportOptions struct {
	DryRun bool
	Slug   string // VIP environment slug (passed to --slug flag)
}

// CompareResult holds the file diff between WPEngine and a VIP repo.
type CompareResult struct {
	MissingFromVIP []string // present in WPEngine wp-content, absent from VIP repo
	MissingFromWPE []string // present in VIP repo, absent from WPEngine wp-content
	GeneratedAt    string
}

// Source is the platform being migrated away from.
type Source interface {
	PullFiles(opts PullOptions) error
	ExportDatabase(opts ExportOptions) error
}

// Destination is the platform being migrated to.
type Destination interface {
	Audit(localPath string, opts AuditOptions) (*AuditReport, error)
	ValidateDatabase(path string) error
	ImportDatabase(path string, opts ImportOptions) error
	ImportMedia(opts ImportOptions) error
	CompareFiles(localPath string) (*CompareResult, error)
}
```

This file is pure type definitions — no logic to test. Verify it compiles:

```bash
PATH="/opt/homebrew/bin:$PATH" go build ./pkg/migration/...
```

- [ ] **Step 2: Write the failing registry test**

Create `pkg/migration/registry_test.go`:

```go
package migration_test

import (
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/migration"
	"github.com/firecrown-media/stax/pkg/provider"
)

type mockSource struct{}

func (m *mockSource) PullFiles(_ migration.PullOptions) error        { return nil }
func (m *mockSource) ExportDatabase(_ migration.ExportOptions) error { return nil }

type mockDest struct{}

func (m *mockDest) Audit(_ string, _ migration.AuditOptions) (*migration.AuditReport, error) {
	return &migration.AuditReport{}, nil
}
func (m *mockDest) ValidateDatabase(_ string) error                          { return nil }
func (m *mockDest) ImportDatabase(_ string, _ migration.ImportOptions) error { return nil }
func (m *mockDest) ImportMedia(_ migration.ImportOptions) error              { return nil }
func (m *mockDest) CompareFiles(_ string) (*migration.CompareResult, error) {
	return &migration.CompareResult{}, nil
}

func TestRegistry_Source(t *testing.T) {
	migration.RegisterSource("test-provider", func(p provider.Provider, cfg *config.Config) migration.Source {
		return &mockSource{}
	})

	src, err := migration.NewSource("test-provider", nil, nil)
	if err != nil {
		t.Fatalf("NewSource returned error: %v", err)
	}
	if src == nil {
		t.Fatal("expected non-nil Source")
	}
}

func TestRegistry_Source_NotFound(t *testing.T) {
	_, err := migration.NewSource("does-not-exist", nil, nil)
	if err == nil {
		t.Fatal("expected error for unregistered source")
	}
}

func TestRegistry_Destination(t *testing.T) {
	migration.RegisterDestination("test-dest", func(repoPath string) migration.Destination {
		return &mockDest{}
	})

	dest, err := migration.NewDestination("test-dest", "/some/path")
	if err != nil {
		t.Fatalf("NewDestination returned error: %v", err)
	}
	if dest == nil {
		t.Fatal("expected non-nil Destination")
	}
}

func TestRegistry_Destination_NotFound(t *testing.T) {
	_, err := migration.NewDestination("does-not-exist", "")
	if err == nil {
		t.Fatal("expected error for unregistered destination")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestRegistry -v
```

Expected: FAIL — `RegisterSource`, `NewSource`, etc. not defined.

- [ ] **Step 4: Create registry.go**

```go
// pkg/migration/registry.go
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
func RegisterSource(name string, factory SourceFactory) {
	mu.Lock()
	defer mu.Unlock()
	sources[name] = factory
}

// RegisterDestination registers a DestinationFactory under the given name.
// Call this from destination package init() functions.
func RegisterDestination(name string, factory DestinationFactory) {
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
```

- [ ] **Step 5: Run registry tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestRegistry -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/migration/interfaces.go pkg/migration/registry.go pkg/migration/registry_test.go
git commit -m "feat(migration): add Source/Destination interfaces and provider registry"
```

---

## Task 4: Implement WPEngineSource

**Files:**
- Create: `pkg/migration/providers/wpengine/source.go`
- Test: `pkg/migration/providers/wpengine/source_test.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/migration/providers/wpengine/source_test.go`:

```go
package wpengine_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/migration"
	wpe "github.com/firecrown-media/stax/pkg/migration/providers/wpengine"
	"github.com/firecrown-media/stax/pkg/provider"
)

// stubProvider satisfies provider.Provider for tests; only ExportDatabase matters.
type stubProvider struct {
	exportErr    error
	exportOutput string
}

func (s *stubProvider) Name() string                     { return "wpengine" }
func (s *stubProvider) Description() string              { return "" }
func (s *stubProvider) Capabilities() provider.ProviderCapabilities { return provider.ProviderCapabilities{} }
func (s *stubProvider) Authenticate(_ map[string]string) error       { return nil }
func (s *stubProvider) TestConnection() error                        { return nil }
func (s *stubProvider) ValidateCredentials(_ map[string]string) error { return nil }
func (s *stubProvider) ListSites() ([]provider.Site, error)          { return nil, nil }
func (s *stubProvider) GetSite(_ string) (*provider.Site, error)     { return nil, nil }
func (s *stubProvider) GetSiteMetadata(_ *provider.Site) (*provider.SiteMetadata, error) { return nil, nil }
func (s *stubProvider) ExportDatabase(_ *provider.Site, _ provider.DatabaseExportOptions) (io.ReadCloser, error) {
	if s.exportErr != nil {
		return nil, s.exportErr
	}
	return io.NopCloser(strings.NewReader(s.exportOutput)), nil
}
func (s *stubProvider) ImportDatabase(_ *provider.Site, _ io.Reader, _ provider.DatabaseImportOptions) error { return nil }
func (s *stubProvider) GetDatabaseCredentials(_ *provider.Site) (*provider.DatabaseCredentials, error) { return nil, nil }
func (s *stubProvider) SyncFiles(_ *provider.Site, _ string, _ provider.SyncOptions) error  { return nil }
func (s *stubProvider) DownloadFile(_ *provider.Site, _ string) (io.ReadCloser, error)       { return nil, nil }
func (s *stubProvider) UploadFile(_ *provider.Site, _, _ string) error                      { return nil }
func (s *stubProvider) GetPHPVersion(_ *provider.Site) (string, error)                      { return "", nil }
func (s *stubProvider) GetMySQLVersion(_ *provider.Site) (string, error)                    { return "", nil }
func (s *stubProvider) GetWordPressVersion(_ *provider.Site) (string, error)                { return "", nil }

func minimalCfg() *config.Config {
	return &config.Config{
		Provider: "wpengine",
		ProviderConfig: map[string]any{
			"install":     "mysite",
			"environment": "production",
		},
	}
}

func TestWPEngineSource_ExportDatabase_WritesFile(t *testing.T) {
	p := &stubProvider{exportOutput: "-- SQL DUMP\nCREATE TABLE test;"}
	src := wpe.NewWPEngineSource(p, minimalCfg())

	outPath := t.TempDir() + "/export.sql"
	err := src.ExportDatabase(migration.ExportOptions{OutputPath: outPath})
	if err != nil {
		t.Fatalf("ExportDatabase failed: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	if !strings.Contains(string(data), "CREATE TABLE test") {
		t.Errorf("output file missing expected content, got: %s", string(data))
	}
}

func TestWPEngineSource_ExportDatabase_ProviderError(t *testing.T) {
	p := &stubProvider{exportErr: errors.New("SSH connection failed")}
	src := wpe.NewWPEngineSource(p, minimalCfg())

	err := src.ExportDatabase(migration.ExportOptions{OutputPath: t.TempDir() + "/out.sql"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "database export failed") {
		t.Errorf("unexpected error: %v", err)
	}
}
```

Add `"os"` import to the test file.

- [ ] **Step 2: Run test to verify it fails**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/providers/wpengine/... -v
```

Expected: FAIL — package does not exist.

- [ ] **Step 3: Implement WPEngineSource**

Create `pkg/migration/providers/wpengine/source.go`:

```go
package wpengine

import (
	"fmt"
	"io"
	"os"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/files"
	"github.com/firecrown-media/stax/pkg/migration"
	"github.com/firecrown-media/stax/pkg/provider"
)

func init() {
	migration.RegisterSource("wpengine", func(p provider.Provider, cfg *config.Config) migration.Source {
		return NewWPEngineSource(p, cfg)
	})
}

// WPEngineSource implements migration.Source for WPEngine installs.
type WPEngineSource struct {
	provider provider.Provider
	cfg      *config.Config
}

// NewWPEngineSource constructs a WPEngineSource.
func NewWPEngineSource(p provider.Provider, cfg *config.Config) *WPEngineSource {
	return &WPEngineSource{provider: p, cfg: cfg}
}

// PullFiles downloads wp-content from WPEngine (uploads excluded by default).
func (s *WPEngineSource) PullFiles(opts migration.PullOptions) error {
	return files.Pull(s.provider, s.cfg, files.SyncFlags{
		ExcludeUploads: true,
		ThemesOnly:     opts.ThemesOnly,
		PluginsOnly:    opts.PluginsOnly,
		MuPluginsOnly:  opts.MuPluginsOnly,
		DryRun:         opts.DryRun,
		ProjectDir:     opts.ProjectDir,
	})
}

// ExportDatabase exports the WPEngine database with VIP-compatible mysqldump flags
// and writes the SQL dump to opts.OutputPath.
func (s *WPEngineSource) ExportDatabase(opts migration.ExportOptions) error {
	install := providerConfigString(s.cfg.ProviderConfig, "install")
	site := &provider.Site{Name: install, Provider: s.cfg.Provider}

	reader, err := s.provider.ExportDatabase(site, provider.DatabaseExportOptions{
		ExtraFlags: []string{
			"--hex-blob",
			"--quote-names",
			"--default-character-set=utf8mb4",
		},
	})
	if err != nil {
		return fmt.Errorf("database export failed: %w", err)
	}
	defer reader.Close()

	f, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", opts.OutputPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("failed to write database export: %w", err)
	}
	return nil
}

func providerConfigString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
```

- [ ] **Step 4: Run tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/providers/wpengine/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/migration/providers/wpengine/source.go pkg/migration/providers/wpengine/source_test.go
git commit -m "feat(migration): implement WPEngineSource for file pull and database export"
```

---

## Task 5: Implement VIPDestination

**Files:**
- Create: `pkg/migration/providers/vip/destination.go`
- Test: `pkg/migration/providers/vip/destination_test.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/migration/providers/vip/destination_test.go`:

```go
package vip_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/firecrown-media/stax/pkg/migration"
	vipdest "github.com/firecrown-media/stax/pkg/migration/providers/vip"
)

func TestVIPDestination_Audit_PHPCSNotFound(t *testing.T) {
	// Override PATH to ensure phpcs cannot be found.
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", t.TempDir())
	defer os.Setenv("PATH", origPath)

	dest := vipdest.NewVIPDestination("")
	_, err := dest.Audit("/some/path", migration.AuditOptions{})
	if err == nil {
		t.Fatal("expected error when phpcs not in PATH")
	}
	if !containsStr(err.Error(), "phpcs not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestVIPDestination_ValidateDatabase_VIPCLINotFound(t *testing.T) {
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", t.TempDir())
	defer os.Setenv("PATH", origPath)

	dest := vipdest.NewVIPDestination("")
	err := dest.ValidateDatabase("/some/file.sql")
	if err == nil {
		t.Fatal("expected error when vip CLI not in PATH")
	}
	if !containsStr(err.Error(), "VIP CLI not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestVIPDestination_CompareFiles(t *testing.T) {
	// Build a fake WPEngine wp-content and VIP repo layout.
	wpeDir := t.TempDir()
	vipDir := t.TempDir()

	// Both have plugins/hello-world; only WPE has plugins/wpe-only; only VIP has plugins/vip-only.
	os.MkdirAll(filepath.Join(wpeDir, "plugins", "hello-world"), 0755)
	os.MkdirAll(filepath.Join(wpeDir, "plugins", "wpe-only"), 0755)
	os.MkdirAll(filepath.Join(vipDir, "plugins", "hello-world"), 0755)
	os.MkdirAll(filepath.Join(vipDir, "plugins", "vip-only"), 0755)

	dest := vipdest.NewVIPDestination(vipDir)
	result, err := dest.CompareFiles(wpeDir)
	if err != nil {
		t.Fatalf("CompareFiles returned error: %v", err)
	}

	if !containsItem(result.MissingFromVIP, "plugins/wpe-only") {
		t.Errorf("expected plugins/wpe-only in MissingFromVIP, got %v", result.MissingFromVIP)
	}
	if !containsItem(result.MissingFromWPE, "plugins/vip-only") {
		t.Errorf("expected plugins/vip-only in MissingFromWPE, got %v", result.MissingFromWPE)
	}
	if containsItem(result.MissingFromVIP, "plugins/hello-world") {
		t.Errorf("plugins/hello-world should not be missing from VIP")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexStr(s, sub) >= 0)
}

func indexStr(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func containsItem(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/providers/vip/... -v
```

Expected: FAIL — package does not exist.

- [ ] **Step 3: Implement VIPDestination**

Create `pkg/migration/providers/vip/destination.go`:

```go
package vip

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/firecrown-media/stax/pkg/migration"
)

func init() {
	migration.RegisterDestination("vip", func(repoPath string) migration.Destination {
		return NewVIPDestination(repoPath)
	})
}

// VIPDestination implements migration.Destination for WordPress VIP.
type VIPDestination struct {
	repoPath string // local path to VIP repo checkout
}

// NewVIPDestination constructs a VIPDestination.
// repoPath is the local VIP repo checkout used by CompareFiles.
func NewVIPDestination(repoPath string) *VIPDestination {
	return &VIPDestination{repoPath: repoPath}
}

// Audit runs phpcs with the WordPress-VIP-Go ruleset against plugins, themes,
// and client-mu-plugins under localPath.
func (d *VIPDestination) Audit(localPath string, opts migration.AuditOptions) (*migration.AuditReport, error) {
	if _, err := exec.LookPath("phpcs"); err != nil {
		return nil, fmt.Errorf("phpcs not found: install with 'composer global require automattic/vip-coding-standards'")
	}

	severity := opts.Severity
	if severity == 0 {
		severity = 1
	}

	targets := []string{
		filepath.Join(localPath, "plugins"),
		filepath.Join(localPath, "themes"),
		filepath.Join(localPath, "client-mu-plugins"),
	}

	report := &migration.AuditReport{
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	for _, target := range targets {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			continue
		}

		args := []string{
			"--standard=WordPress-VIP-Go",
			fmt.Sprintf("--severity=%d", severity),
			"--report=json",
			target,
		}

		out, err := exec.Command("phpcs", args...).Output()
		// phpcs exits non-zero when violations exist but still produces JSON output.
		if err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				return nil, fmt.Errorf("phpcs failed on %s: %w", target, err)
			}
		}

		results, err := parsePHPCSReport(out)
		if err != nil {
			return nil, fmt.Errorf("failed to parse phpcs output for %s: %w", target, err)
		}
		report.Files = append(report.Files, results...)
	}

	for _, f := range report.Files {
		report.TotalErrors += f.Errors
		report.TotalWarnings += f.Warnings
	}

	return report, nil
}

// ValidateDatabase runs vip import validate-sql against the SQL file at path.
func (d *VIPDestination) ValidateDatabase(path string) error {
	if _, err := exec.LookPath("vip"); err != nil {
		return fmt.Errorf("VIP CLI not found: install with 'npm install -g @automattic/vip'")
	}
	out, err := exec.Command("vip", "import", "validate-sql", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("database validation failed: %w\n%s", err, string(out))
	}
	return nil
}

// ImportDatabase runs vip import sql on the SQL file at path.
func (d *VIPDestination) ImportDatabase(path string, opts migration.ImportOptions) error {
	if _, err := exec.LookPath("vip"); err != nil {
		return fmt.Errorf("VIP CLI not found: install with 'npm install -g @automattic/vip'")
	}
	args := []string{"import", "sql", path}
	if opts.Slug != "" {
		args = append(args, "--slug="+opts.Slug)
	}
	out, err := exec.Command("vip", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("database import failed: %w\n%s", err, string(out))
	}
	return nil
}

// ImportMedia runs vip import media.
func (d *VIPDestination) ImportMedia(opts migration.ImportOptions) error {
	if _, err := exec.LookPath("vip"); err != nil {
		return fmt.Errorf("VIP CLI not found: install with 'npm install -g @automattic/vip'")
	}
	args := []string{"import", "media"}
	if opts.Slug != "" {
		args = append(args, "--slug="+opts.Slug)
	}
	out, err := exec.Command("vip", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("media import failed: %w\n%s", err, string(out))
	}
	return nil
}

// CompareFiles diffs the downloaded WPEngine wp-content (localPath) against
// the VIP repo checkout (d.repoPath) across plugins, themes, and
// client-mu-plugins. Returns top-level directory names that are present on one
// side but not the other.
func (d *VIPDestination) CompareFiles(localPath string) (*migration.CompareResult, error) {
	result := &migration.CompareResult{
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	for _, dir := range []string{"plugins", "themes", "client-mu-plugins"} {
		wpePath := filepath.Join(localPath, dir)
		vipPath := filepath.Join(d.repoPath, dir)

		wpeItems, _ := listTopLevel(wpePath)
		vipItems, _ := listTopLevel(vipPath)

		wpeSet := toSet(wpeItems)
		vipSet := toSet(vipItems)

		for item := range wpeSet {
			if !vipSet[item] {
				result.MissingFromVIP = append(result.MissingFromVIP, dir+"/"+item)
			}
		}
		for item := range vipSet {
			if !wpeSet[item] {
				result.MissingFromWPE = append(result.MissingFromWPE, dir+"/"+item)
			}
		}
	}

	return result, nil
}

// ---- phpcs JSON parsing ----

type phpcsOutput struct {
	Totals phpcsTotals          `json:"totals"`
	Files  map[string]phpcsFile `json:"files"`
}

type phpcsTotals struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
}

type phpcsFile struct {
	Errors   int            `json:"errors"`
	Warnings int            `json:"warnings"`
	Messages []phpcsMessage `json:"messages"`
}

type phpcsMessage struct {
	Message  string `json:"message"`
	Source   string `json:"source"`
	Severity int    `json:"severity"`
	Type     string `json:"type"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

func parsePHPCSReport(data []byte) ([]migration.FileAuditResult, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var out phpcsOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	results := make([]migration.FileAuditResult, 0, len(out.Files))
	for filePath, file := range out.Files {
		result := migration.FileAuditResult{
			FilePath: filePath,
			Errors:   file.Errors,
			Warnings: file.Warnings,
		}
		for _, msg := range file.Messages {
			result.Messages = append(result.Messages, migration.AuditMessage{
				Line:     msg.Line,
				Column:   msg.Column,
				Severity: msg.Severity,
				Type:     msg.Type,
				Message:  msg.Message,
				Source:   msg.Source,
			})
		}
		results = append(results, result)
	}
	return results, nil
}

// ---- file comparison helpers ----

func listTopLevel(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
```

- [ ] **Step 4: Run tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/providers/vip/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/migration/providers/vip/destination.go pkg/migration/providers/vip/destination_test.go
git commit -m "feat(migration): implement VIPDestination for audit, validate, import, and file compare"
```

---

## Task 6: Create migration service

**Files:**
- Create: `pkg/migration/service.go`
- Test: `pkg/migration/service_test.go`

- [ ] **Step 1: Write failing tests**

Create `pkg/migration/service_test.go`:

```go
package migration_test

import (
	"errors"
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/migration"
	"github.com/firecrown-media/stax/pkg/provider"
)

// Use the mockSource and mockDest types from registry_test.go (same package).
// If test files need to share helpers, move them to a testutil file within
// the package. For this plan, assume they are each defined in their own test file.

type mockSrc2 struct{ pullErr, exportErr error }

func (m *mockSrc2) PullFiles(_ migration.PullOptions) error        { return m.pullErr }
func (m *mockSrc2) ExportDatabase(_ migration.ExportOptions) error { return m.exportErr }

type mockDst2 struct{ auditErr, validateErr, importErr error }

func (m *mockDst2) Audit(_ string, _ migration.AuditOptions) (*migration.AuditReport, error) {
	if m.auditErr != nil {
		return nil, m.auditErr
	}
	return &migration.AuditReport{}, nil
}
func (m *mockDst2) ValidateDatabase(_ string) error                          { return m.validateErr }
func (m *mockDst2) ImportDatabase(_ string, _ migration.ImportOptions) error { return m.importErr }
func (m *mockDst2) ImportMedia(_ migration.ImportOptions) error              { return nil }
func (m *mockDst2) CompareFiles(_ string) (*migration.CompareResult, error) {
	return &migration.CompareResult{}, nil
}

func cfgWithDest(dest string) *config.Config {
	return &config.Config{
		Provider: "test-svc",
		ProviderConfig: map[string]any{
			"install": "mysite",
		},
		Migration: config.MigrationConfig{Destination: dest},
	}
}

func init() {
	migration.RegisterSource("test-svc", func(_ provider.Provider, _ *config.Config) migration.Source {
		return &mockSrc2{}
	})
	migration.RegisterDestination("test-dst", func(_ string) migration.Destination {
		return &mockDst2{}
	})
}

func TestRequireDestination_MissingError(t *testing.T) {
	cfg := cfgWithDest("")
	err := migration.Pull(nil, cfg, migration.PullOptions{})
	if err == nil {
		t.Fatal("expected error for missing migration.destination")
	}
	if !containsStr2(err.Error(), "migration.destination") {
		t.Errorf("error should mention migration.destination, got: %v", err)
	}
}

func TestAudit_MissingDestination(t *testing.T) {
	cfg := cfgWithDest("")
	_, err := migration.Audit(cfg, "/some/path", migration.AuditOptions{})
	if err == nil {
		t.Fatal("expected error for missing migration.destination")
	}
}

func containsStr2(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -run TestRequireDestination -v
```

Expected: FAIL — `migration.Pull`, `migration.Audit` not defined.

- [ ] **Step 3: Implement service.go**

Create `pkg/migration/service.go`:

```go
package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/ui"
)

// requireDestination returns an error if migration.destination is not set.
func requireDestination(cfg *config.Config) error {
	if cfg.Migration.Destination == "" {
		return fmt.Errorf("migration.destination is not set in .stax.yml\n\nAdd it:\n\n  migration:\n    destination: vip\n")
	}
	return nil
}

// Pull downloads wp-content from the source provider (uploads excluded).
func Pull(p provider.Provider, cfg *config.Config, opts PullOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}
	src, err := NewSource(cfg.Provider, p, cfg)
	if err != nil {
		return err
	}
	ui.Info("Pulling files from %s (uploads excluded)...", cfg.Provider)
	return src.PullFiles(opts)
}

// Export dumps the source database with VIP-compatible flags to a local SQL file.
// If opts.OutputPath is empty, it defaults to .stax/<install>-export.sql.
func Export(p provider.Provider, cfg *config.Config, opts ExportOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}
	if opts.OutputPath == "" {
		install := providerConfigString(cfg.ProviderConfig, "install")
		opts.OutputPath = filepath.Join(".stax", install+"-export.sql")
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	src, err := NewSource(cfg.Provider, p, cfg)
	if err != nil {
		return err
	}
	ui.Info("Exporting database with VIP-compatible flags...")
	if err := src.ExportDatabase(opts); err != nil {
		return err
	}
	ui.Success("Database exported to %s", opts.OutputPath)
	return nil
}

// Audit runs a phpcs compatibility audit against the WordPress-VIP-Go ruleset.
// localPath should be the downloaded wp-content directory.
func Audit(cfg *config.Config, localPath string, opts AuditOptions) (*AuditReport, error) {
	if err := requireDestination(cfg); err != nil {
		return nil, err
	}
	dest, err := NewDestination(cfg.Migration.Destination, "")
	if err != nil {
		return nil, err
	}
	ui.Info("Running phpcs audit (WordPress-VIP-Go ruleset)...")
	return dest.Audit(localPath, opts)
}

// Compare diffs the downloaded WPEngine wp-content against the local VIP repo.
// vipRepoPath must be the root of the VIP repo checkout.
func Compare(cfg *config.Config, localPath, vipRepoPath string) (*CompareResult, error) {
	if err := requireDestination(cfg); err != nil {
		return nil, err
	}
	dest, err := NewDestination(cfg.Migration.Destination, vipRepoPath)
	if err != nil {
		return nil, err
	}
	return dest.CompareFiles(localPath)
}

// Import validates then imports a SQL dump into the VIP destination.
func Import(cfg *config.Config, sqlPath string, opts ImportOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}
	dest, err := NewDestination(cfg.Migration.Destination, "")
	if err != nil {
		return err
	}
	ui.Info("Validating database...")
	if err := dest.ValidateDatabase(sqlPath); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	ui.Info("Importing database...")
	return dest.ImportDatabase(sqlPath, opts)
}

// ReportOptions configures stax migrate report.
type ReportOptions struct {
	LocalPath   string // path to downloaded wp-content
	VIPRepoPath string // path to VIP repo checkout
	SQLPath     string // path to exported SQL file (optional)
	OutputPath  string // output markdown path; defaults to .stax/migration-report.md
}

// Report runs audit and compare, then writes a combined markdown report.
func Report(p provider.Provider, cfg *config.Config, opts ReportOptions) error {
	if err := requireDestination(cfg); err != nil {
		return err
	}

	install := providerConfigString(cfg.ProviderConfig, "install")

	if opts.OutputPath == "" {
		opts.OutputPath = ".stax/migration-report.md"
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	audit, err := Audit(cfg, opts.LocalPath, AuditOptions{})
	if err != nil {
		return fmt.Errorf("audit failed: %w", err)
	}

	compare, err := Compare(cfg, opts.LocalPath, opts.VIPRepoPath)
	if err != nil {
		return fmt.Errorf("compare failed: %w", err)
	}

	var sqlSize int64
	if opts.SQLPath != "" {
		if info, err := os.Stat(opts.SQLPath); err == nil {
			sqlSize = info.Size()
		}
	}

	type reportData struct {
		Install         string
		Destination     string
		GeneratedAt     string
		AuditReport     *AuditReport
		CompareResult   *CompareResult
		SQLPath         string
		SQLSizeHuman    string
	}

	data := reportData{
		Install:       install,
		Destination:   cfg.Migration.Destination,
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05 MST"),
		AuditReport:   audit,
		CompareResult: compare,
		SQLPath:       opts.SQLPath,
		SQLSizeHuman:  humanizeBytes(sqlSize),
	}

	tmpl := template.Must(template.New("report").Funcs(template.FuncMap{
		"join": strings.Join,
		"sort": func(s []string) []string { sort.Strings(s); return s },
	}).Parse(reportTemplate))

	f, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	ui.Success("Report written to %s", opts.OutputPath)
	return nil
}

// Status prints the current migration configuration and the presence of key artifacts.
func Status(cfg *config.Config) error {
	install := providerConfigString(cfg.ProviderConfig, "install")

	ui.Section("Migration Status")
	ui.Info("Install:     %s", install)
	ui.Info("Source:      %s", cfg.Provider)
	ui.Info("Destination: %s", cfg.Migration.Destination)

	checks := []struct {
		label string
		path  string
	}{
		{"Files pulled", "wp-content"},
		{"Database exported", filepath.Join(".stax", install+"-export.sql")},
		{"Report generated", ".stax/migration-report.md"},
	}

	ui.Section("Artifacts")
	for _, c := range checks {
		if _, err := os.Stat(c.path); err == nil {
			ui.Success("  [x] %s (%s)", c.label, c.path)
		} else {
			ui.Info("  [ ] %s", c.label)
		}
	}
	return nil
}

func providerConfigString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func humanizeBytes(b int64) string {
	if b == 0 {
		return "unknown"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

const reportTemplate = `# Migration Report: {{ .Install }}

**Generated:** {{ .GeneratedAt }}
**Source:** wpengine/{{ .Install }}
**Destination:** {{ .Destination }}

---

## phpcs Audit (WordPress-VIP-Go)

**Total errors:** {{ .AuditReport.TotalErrors }}
**Total warnings:** {{ .AuditReport.TotalWarnings }}

{{ range .AuditReport.Files -}}
### {{ .FilePath }}

Errors: {{ .Errors }} | Warnings: {{ .Warnings }}

{{ range .Messages -}}
- Line {{ .Line }}, Col {{ .Column }} [{{ .Type }}] {{ .Message }} ({{ .Source }})
{{ end }}
{{ end }}

---

## File Comparison

### Present in WPEngine, missing from VIP repo

{{ if .CompareResult.MissingFromVIP -}}
{{ range (sort .CompareResult.MissingFromVIP) -}}
- {{ . }}
{{ end }}
{{- else }}
None.
{{ end }}

### Present in VIP repo, missing from WPEngine

{{ if .CompareResult.MissingFromWPE -}}
{{ range (sort .CompareResult.MissingFromWPE) -}}
- {{ . }}
{{ end }}
{{- else }}
None.
{{ end }}

---

## Database Export

{{ if .SQLPath -}}
File: {{ .SQLPath }}
Size: {{ .SQLSizeHuman }}
{{- else }}
No SQL export found. Run ` + "`" + `stax migrate export` + "`" + ` first.
{{ end }}
`
```

- [ ] **Step 4: Run tests**

```bash
PATH="/opt/homebrew/bin:$PATH" go test ./pkg/migration/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/migration/service.go pkg/migration/service_test.go
git commit -m "feat(migration): add migration service (pull, export, audit, compare, import, report, status)"
```

---

## Task 7: Create cmd/migrate.go and wire up imports

**Files:**
- Create: `cmd/migrate.go`

- [ ] **Step 1: Create cmd/migrate.go**

```go
package cmd

import (
	"github.com/firecrown-media/stax/pkg/migration"
	_ "github.com/firecrown-media/stax/pkg/migration/providers/vip"
	_ "github.com/firecrown-media/stax/pkg/migration/providers/wpengine"
	"github.com/spf13/cobra"
)

var (
	migrateDestination  string
	migrateLocalPath    string
	migrateVIPRepoPath  string
	migrateSQLPath      string
	migrateOutputPath   string
	migrateDryRun       bool
	migrateSlug         string
	migrateThemesOnly   bool
	migratePluginsOnly  bool
	migrateMuPluginsOnly bool
	migrateSeverity     int
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate WordPress sites between hosting providers",
	Long: `Orchestrate migration from one hosting provider to another.

Source is determined by the provider: field in .stax.yml.
Destination is set via migration.destination in .stax.yml or the --destination flag.

Requires migration.destination to be set:

  migration:
    destination: vip`,
}

var migratePullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull files from source provider to local wp-content/",
	Long:  `Download wp-content from the source provider (uploads excluded by default).`,
	Example: `  stax migrate pull
  stax migrate pull --themes-only
  stax migrate pull --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migrateDestination != "" {
			cfg.Migration.Destination = migrateDestination
		}
		p, err := getProvider(cfg)
		if err != nil {
			return err
		}
		return migration.Pull(p, cfg, migration.PullOptions{
			ThemesOnly:    migrateThemesOnly,
			PluginsOnly:   migratePluginsOnly,
			MuPluginsOnly: migrateMuPluginsOnly,
			DryRun:        migrateDryRun,
			ProjectDir:    getProjectDir(),
		})
	},
}

var migrateExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export database from source with VIP-compatible flags",
	Long:  `Export the source database with mysqldump flags required by WordPress VIP.`,
	Example: `  stax migrate export
  stax migrate export --output=mysite-export.sql`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migrateDestination != "" {
			cfg.Migration.Destination = migrateDestination
		}
		p, err := getProvider(cfg)
		if err != nil {
			return err
		}
		return migration.Export(p, cfg, migration.ExportOptions{
			OutputPath: migrateSQLPath,
		})
	},
}

var migrateAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run phpcs compatibility audit against the VIP ruleset",
	Long: `Scan plugins, themes, and client-mu-plugins in the local wp-content
directory against the WordPress-VIP-Go phpcs ruleset.`,
	Example: `  stax migrate audit
  stax migrate audit --path=../astronomy/wp-content
  stax migrate audit --severity=5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migrateDestination != "" {
			cfg.Migration.Destination = migrateDestination
		}
		localPath := migrateLocalPath
		if localPath == "" {
			localPath = getProjectDir() + "/wp-content"
		}
		report, err := migration.Audit(cfg, localPath, migration.AuditOptions{
			Severity: migrateSeverity,
		})
		if err != nil {
			return err
		}
		printAuditSummary(report)
		return nil
	},
}

var migrateCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Diff local files against the destination VIP repo",
	Long: `Compare plugins, themes, and client-mu-plugins between the downloaded
WPEngine wp-content and the local VIP repo checkout.`,
	Example: `  stax migrate compare --vip-repo=../vip-repo
  stax migrate compare --path=../wpe/wp-content --vip-repo=../vip-repo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migrateDestination != "" {
			cfg.Migration.Destination = migrateDestination
		}
		localPath := migrateLocalPath
		if localPath == "" {
			localPath = getProjectDir() + "/wp-content"
		}
		result, err := migration.Compare(cfg, localPath, migrateVIPRepoPath)
		if err != nil {
			return err
		}
		printCompareResult(result)
		return nil
	},
}

var migrateImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Push DB/media into destination (wraps VIP CLI)",
	Long:  `Validate and import the SQL dump into the VIP destination using the VIP CLI.`,
	Example: `  stax migrate import --sql=.stax/mysite-export.sql
  stax migrate import --sql=export.sql --slug=my-vip-env`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migrateDestination != "" {
			cfg.Migration.Destination = migrateDestination
		}
		return migration.Import(cfg, migrateSQLPath, migration.ImportOptions{
			DryRun: migrateDryRun,
			Slug:   migrateSlug,
		})
	},
}

var migrateReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate combined audit + compare report as markdown",
	Long:  `Run audit and compare, then write a combined migration report to .stax/migration-report.md.`,
	Example: `  stax migrate report --vip-repo=../vip-repo
  stax migrate report --path=../wpe/wp-content --vip-repo=../vip-repo --sql=.stax/mysite-export.sql`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migrateDestination != "" {
			cfg.Migration.Destination = migrateDestination
		}
		p, err := getProvider(cfg)
		if err != nil {
			return err
		}
		localPath := migrateLocalPath
		if localPath == "" {
			localPath = getProjectDir() + "/wp-content"
		}
		return migration.Report(p, cfg, migration.ReportOptions{
			LocalPath:   localPath,
			VIPRepoPath: migrateVIPRepoPath,
			SQLPath:     migrateSQLPath,
			OutputPath:  migrateOutputPath,
		})
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration state for this project",
	Long:  `Print migration configuration and the presence of key artifacts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		return migration.Status(cfg)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migratePullCmd)
	migrateCmd.AddCommand(migrateExportCmd)
	migrateCmd.AddCommand(migrateAuditCmd)
	migrateCmd.AddCommand(migrateCompareCmd)
	migrateCmd.AddCommand(migrateImportCmd)
	migrateCmd.AddCommand(migrateReportCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	// Shared flags on the parent command (inherited by all subcommands)
	migrateCmd.PersistentFlags().StringVar(&migrateDestination, "destination", "", "override migration.destination from config")

	// pull flags
	migratePullCmd.Flags().BoolVar(&migrateThemesOnly, "themes-only", false, "pull only themes directory")
	migratePullCmd.Flags().BoolVar(&migratePluginsOnly, "plugins-only", false, "pull only plugins directory")
	migratePullCmd.Flags().BoolVar(&migrateMuPluginsOnly, "mu-plugins-only", false, "pull only mu-plugins directory")
	migratePullCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "show what would be transferred without pulling")

	// export flags
	migrateExportCmd.Flags().StringVar(&migrateSQLPath, "output", "", "path for the SQL dump file (default: .stax/<install>-export.sql)")

	// audit flags
	migrateAuditCmd.Flags().StringVar(&migrateLocalPath, "path", "", "path to wp-content directory (default: <project>/wp-content)")
	migrateAuditCmd.Flags().IntVar(&migrateSeverity, "severity", 1, "minimum phpcs severity level (1–5)")

	// compare flags
	migrateCompareCmd.Flags().StringVar(&migrateLocalPath, "path", "", "path to downloaded wp-content (default: <project>/wp-content)")
	migrateCompareCmd.Flags().StringVar(&migrateVIPRepoPath, "vip-repo", "", "path to local VIP repo checkout")
	_ = migrateCompareCmd.MarkFlagRequired("vip-repo")

	// import flags
	migrateImportCmd.Flags().StringVar(&migrateSQLPath, "sql", "", "path to the SQL dump file")
	migrateImportCmd.Flags().StringVar(&migrateSlug, "slug", "", "VIP environment slug (passed to --slug)")
	migrateImportCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "dry run (no changes)")
	_ = migrateImportCmd.MarkFlagRequired("sql")

	// report flags
	migrateReportCmd.Flags().StringVar(&migrateLocalPath, "path", "", "path to wp-content (default: <project>/wp-content)")
	migrateReportCmd.Flags().StringVar(&migrateVIPRepoPath, "vip-repo", "", "path to local VIP repo checkout")
	migrateReportCmd.Flags().StringVar(&migrateSQLPath, "sql", "", "path to SQL dump file (optional)")
	migrateReportCmd.Flags().StringVar(&migrateOutputPath, "output", "", "output path for report (default: .stax/migration-report.md)")
	_ = migrateReportCmd.MarkFlagRequired("vip-repo")
}

// getProvider creates an authenticated provider from config.
// cmd/migrate.go calls this instead of provider.NewAuthenticatedProvider
// because it needs access outside of RunE.
func getProvider(cfg *config.Config) (provider.Provider, error) {
	return provider.NewAuthenticatedProvider(cfg)
}

func printAuditSummary(report *migration.AuditReport) {
	ui.Section("phpcs Audit Summary")
	ui.Info("Files scanned: %d", len(report.AuditReport.Files))   // fix: report not .AuditReport
	ui.Info("Total errors:  %d", report.TotalErrors)
	ui.Info("Total warnings:%d", report.TotalWarnings)
	for _, f := range report.Files {
		if f.Errors > 0 || f.Warnings > 0 {
			ui.Warning("  %s — %d errors, %d warnings", f.FilePath, f.Errors, f.Warnings)
		}
	}
}

func printCompareResult(result *migration.CompareResult) {
	ui.Section("File Comparison")
	if len(result.MissingFromVIP) == 0 && len(result.MissingFromWPE) == 0 {
		ui.Success("No differences found")
		return
	}
	if len(result.MissingFromVIP) > 0 {
		ui.Warning("Present in WPEngine, missing from VIP repo:")
		for _, p := range result.MissingFromVIP {
			ui.Info("  - %s", p)
		}
	}
	if len(result.MissingFromWPE) > 0 {
		ui.Info("Present in VIP repo, missing from WPEngine:")
		for _, p := range result.MissingFromWPE {
			ui.Info("  - %s", p)
		}
	}
}
```

**Important:** `cmd/migrate.go` references `provider.NewAuthenticatedProvider`. Check `cmd/root.go` or `cmd/provider.go` for the correct function name used in other commands (e.g., `files.go` calls `provider.NewAuthenticatedProvider(cfg)` — use the same). Also fix the `printAuditSummary` bug noted in the comment above — the field is `report.Files` not `report.AuditReport.Files`.

Also note: `migrateCmd` should NOT be in `skipConfigCommands` in `cmd/root.go` — it needs config. No changes needed to root.go.

- [ ] **Step 2: Build and verify**

```bash
PATH="/opt/homebrew/bin:$PATH" make build
```

Expected: binary builds without errors.

```bash
./bin/stax migrate --help
./bin/stax migrate pull --help
./bin/stax migrate export --help
```

Expected: help text displays for each subcommand.

- [ ] **Step 3: Run all tests**

```bash
PATH="/opt/homebrew/bin:$PATH" make test
```

Expected: all tests PASS, no race conditions.

- [ ] **Step 4: Fix the printAuditSummary bug**

The `printAuditSummary` function in the draft above has a typo. Fix it:

```go
func printAuditSummary(report *migration.AuditReport) {
	ui.Section("phpcs Audit Summary")
	ui.Info("Files scanned:  %d", len(report.Files))
	ui.Info("Total errors:   %d", report.TotalErrors)
	ui.Info("Total warnings: %d", report.TotalWarnings)
	for _, f := range report.Files {
		if f.Errors > 0 || f.Warnings > 0 {
			ui.Warning("  %s — %d errors, %d warnings", f.FilePath, f.Errors, f.Warnings)
		}
	}
}
```

- [ ] **Step 5: Commit**

```bash
git add cmd/migrate.go
git commit -m "feat(cmd): add stax migrate command group (pull, export, audit, compare, import, report, status)"
```

---

## Self-Review Checklist

- [x] MigrationConfig in config — Task 1
- [x] ExtraFlags for VIP-compatible export — Task 2
- [x] Source/Destination interfaces — Task 3
- [x] Registry with RegisterSource/RegisterDestination/NewSource/NewDestination — Task 3
- [x] WPEngineSource.PullFiles delegates to pkg/files — Task 4
- [x] WPEngineSource.ExportDatabase writes SQL to file with VIP flags — Task 4
- [x] VIPDestination.Audit runs phpcs with WordPress-VIP-Go — Task 5
- [x] VIPDestination.ValidateDatabase / ImportDatabase / ImportMedia wrap VIP CLI — Task 5
- [x] VIPDestination.CompareFiles diffs top-level dirs across plugins/themes/client-mu-plugins — Task 5
- [x] All 7 stax migrate subcommands — Task 7
- [x] --destination flag available on all migrate subcommands — Task 7
- [x] Clear error when migration.destination missing — Task 6 (requireDestination)
- [x] Error messages for missing phpcs and VIP CLI — Task 5
- [x] Report written to .stax/migration-report.md — Task 6
- [x] init() blank imports in cmd/migrate.go register providers — Task 7
