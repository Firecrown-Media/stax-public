package build

import (
	"testing"
	"time"
)

func TestBuildOptions_Defaults(t *testing.T) {
	opts := BuildOptions{}

	// Default values should be zero values
	if opts.Verbose {
		t.Error("expected Verbose to be false by default")
	}
	if opts.Force {
		t.Error("expected Force to be false by default")
	}
	if opts.Clean {
		t.Error("expected Clean to be false by default")
	}
	if opts.SkipComposer {
		t.Error("expected SkipComposer to be false by default")
	}
	if opts.SkipNPM {
		t.Error("expected SkipNPM to be false by default")
	}
}

func TestBuildOptions_AllFlags(t *testing.T) {
	opts := BuildOptions{
		ProjectPath:  "/path/to/project",
		Verbose:      true,
		Force:        true,
		Clean:        true,
		SkipComposer: true,
		SkipNPM:      true,
		Parallel:     true,
		Timeout:      300,
	}

	if opts.ProjectPath != "/path/to/project" {
		t.Errorf("unexpected ProjectPath: %s", opts.ProjectPath)
	}
	if !opts.Verbose {
		t.Error("expected Verbose to be true")
	}
	if !opts.Force {
		t.Error("expected Force to be true")
	}
	if !opts.Clean {
		t.Error("expected Clean to be true")
	}
	if !opts.SkipComposer {
		t.Error("expected SkipComposer to be true")
	}
	if !opts.SkipNPM {
		t.Error("expected SkipNPM to be true")
	}
	if !opts.Parallel {
		t.Error("expected Parallel to be true")
	}
	if opts.Timeout != 300 {
		t.Errorf("expected Timeout to be 300, got %d", opts.Timeout)
	}
}

func TestComposerOptions_AllFlags(t *testing.T) {
	opts := ComposerOptions{
		WorkingDir:         "/path/to/dir",
		NoDev:              true,
		NoScripts:          true,
		IgnorePlatformReqs: true,
		PreferDist:         true,
		PreferSource:       true,
		Optimize:           true,
		Verbose:            true,
		Timeout:            120,
	}

	if opts.WorkingDir != "/path/to/dir" {
		t.Errorf("unexpected WorkingDir: %s", opts.WorkingDir)
	}
	if !opts.NoDev {
		t.Error("expected NoDev to be true")
	}
	if !opts.NoScripts {
		t.Error("expected NoScripts to be true")
	}
	if !opts.IgnorePlatformReqs {
		t.Error("expected IgnorePlatformReqs to be true")
	}
	if !opts.PreferDist {
		t.Error("expected PreferDist to be true")
	}
	if !opts.PreferSource {
		t.Error("expected PreferSource to be true")
	}
	if !opts.Optimize {
		t.Error("expected Optimize to be true")
	}
	if !opts.Verbose {
		t.Error("expected Verbose to be true")
	}
	if opts.Timeout != 120 {
		t.Errorf("expected Timeout to be 120, got %d", opts.Timeout)
	}
}

func TestNPMOptions_AllFlags(t *testing.T) {
	opts := NPMOptions{
		WorkingDir:     "/path/to/dir",
		Production:     true,
		LegacyPeerDeps: true,
		Clean:          true,
		Verbose:        true,
		Timeout:        300,
	}

	if opts.WorkingDir != "/path/to/dir" {
		t.Errorf("unexpected WorkingDir: %s", opts.WorkingDir)
	}
	if !opts.Production {
		t.Error("expected Production to be true")
	}
	if !opts.LegacyPeerDeps {
		t.Error("expected LegacyPeerDeps to be true")
	}
	if !opts.Clean {
		t.Error("expected Clean to be true")
	}
	if !opts.Verbose {
		t.Error("expected Verbose to be true")
	}
	if opts.Timeout != 300 {
		t.Errorf("expected Timeout to be 300, got %d", opts.Timeout)
	}
}

func TestBuildStatus_Fields(t *testing.T) {
	status := BuildStatus{
		NeedsBuild:         true,
		BuildScriptExists:  true,
		CustomBuildScripts: []string{"10-mu-plugins.sh", "20-theme.sh"},
		ComposerStatus: DependencyStatus{
			Installed: true,
		},
		NPMStatus: DependencyStatus{
			Installed: true,
		},
		LastBuildTime: time.Now(),
		Reasons:       []string{"Source files modified"},
	}

	if !status.NeedsBuild {
		t.Error("expected NeedsBuild to be true")
	}
	if !status.BuildScriptExists {
		t.Error("expected BuildScriptExists to be true")
	}
	if len(status.CustomBuildScripts) != 2 {
		t.Errorf("expected 2 custom build scripts, got %d", len(status.CustomBuildScripts))
	}
	if !status.ComposerStatus.Installed {
		t.Error("expected ComposerStatus.Installed to be true")
	}
	if !status.NPMStatus.Installed {
		t.Error("expected NPMStatus.Installed to be true")
	}
	if len(status.Reasons) != 1 {
		t.Errorf("expected 1 reason, got %d", len(status.Reasons))
	}
}

func TestDependencyStatus_Fields(t *testing.T) {
	now := time.Now()
	status := DependencyStatus{
		ConfigFile:     "/path/to/composer.json",
		LockFile:       "/path/to/composer.lock",
		VendorDir:      "/path/to/vendor",
		ConfigExists:   true,
		LockExists:     true,
		VendorExists:   true,
		Installed:      true,
		NeedsUpdate:    false,
		ConfigModified: now,
		LockModified:   now,
		VendorModified: now,
	}

	if status.ConfigFile != "/path/to/composer.json" {
		t.Errorf("expected ConfigFile to be /path/to/composer.json, got %s", status.ConfigFile)
	}
	if !status.ConfigExists {
		t.Error("expected ConfigExists to be true")
	}
	if !status.LockExists {
		t.Error("expected LockExists to be true")
	}
	if !status.VendorExists {
		t.Error("expected VendorExists to be true")
	}
	if !status.Installed {
		t.Error("expected Installed to be true")
	}
	if status.NeedsUpdate {
		t.Error("expected NeedsUpdate to be false")
	}
}

func TestPHPCSOptions_Fields(t *testing.T) {
	opts := PHPCSOptions{
		WorkingDir:      "/path/to/dir",
		ConfigFile:      "/path/to/phpcs.xml",
		Standard:        "PSR12",
		Extensions:      []string{"php", "inc"},
		Ignore:          "vendor/*,node_modules/*",
		Report:          "json",
		ShowSniffs:      true,
		Severity:        5,
		ErrorSeverity:   3,
		WarningSeverity: 7,
		Files:           []string{"src/", "tests/"},
	}

	if opts.WorkingDir != "/path/to/dir" {
		t.Errorf("unexpected WorkingDir: %s", opts.WorkingDir)
	}
	if opts.ConfigFile != "/path/to/phpcs.xml" {
		t.Errorf("expected ConfigFile to be /path/to/phpcs.xml, got %s", opts.ConfigFile)
	}
	if opts.Standard != "PSR12" {
		t.Errorf("expected Standard to be PSR12, got %s", opts.Standard)
	}
	if len(opts.Extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(opts.Extensions))
	}
	if opts.Ignore != "vendor/*,node_modules/*" {
		t.Errorf("unexpected Ignore value: %s", opts.Ignore)
	}
	if opts.Report != "json" {
		t.Errorf("expected Report to be json, got %s", opts.Report)
	}
	if !opts.ShowSniffs {
		t.Error("expected ShowSniffs to be true")
	}
	if opts.Severity != 5 {
		t.Errorf("expected Severity to be 5, got %d", opts.Severity)
	}
}

func TestPHPCSResult_Fields(t *testing.T) {
	result := PHPCSResult{
		Success:  false,
		ExitCode: 1,
		Errors:   5,
		Warnings: 3,
		Fixable:  2,
		Output:   "test output",
		Files: []PHPCSFileResult{
			{
				File:     "/path/to/file.php",
				Errors:   3,
				Warnings: 2,
				Messages: []PHPCSMessage{
					{
						Line:     10,
						Column:   1,
						Type:     "ERROR",
						Message:  "Test message",
						Source:   "Test.Source",
						Severity: 5,
						Fixable:  true,
					},
				},
			},
		},
	}

	if result.Success {
		t.Error("expected Success to be false")
	}
	if result.ExitCode != 1 {
		t.Errorf("expected ExitCode to be 1, got %d", result.ExitCode)
	}
	if result.Errors != 5 {
		t.Errorf("expected 5 errors, got %d", result.Errors)
	}
	if result.Warnings != 3 {
		t.Errorf("expected 3 warnings, got %d", result.Warnings)
	}
	if result.Fixable != 2 {
		t.Errorf("expected 2 fixable, got %d", result.Fixable)
	}
	if len(result.Files) != 1 {
		t.Errorf("expected 1 file result, got %d", len(result.Files))
	}
	if len(result.Files[0].Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(result.Files[0].Messages))
	}
}

func TestScriptInfo_Fields(t *testing.T) {
	info := ScriptInfo{
		Name:        "10-mu-plugins.sh",
		Path:        "/path/to/scripts/build/10-mu-plugins.sh",
		Type:        "custom",
		Description: "Build MU plugins",
		Order:       10,
	}

	if info.Name != "10-mu-plugins.sh" {
		t.Errorf("expected Name to be 10-mu-plugins.sh, got %s", info.Name)
	}
	if info.Type != "custom" {
		t.Errorf("expected Type to be custom, got %s", info.Type)
	}
	if info.Order != 10 {
		t.Errorf("expected Order to be 10, got %d", info.Order)
	}
}

func TestWatchOptions_Fields(t *testing.T) {
	opts := WatchOptions{
		Paths:          []string{"src/", "assets/"},
		IgnorePatterns: []string{"*.tmp", "*.log"},
		Command:        "npm run build",
		Debounce:       100,
		Recursive:      true,
		Verbose:        true,
	}

	if len(opts.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(opts.Paths))
	}
	if len(opts.IgnorePatterns) != 2 {
		t.Errorf("expected 2 ignore patterns, got %d", len(opts.IgnorePatterns))
	}
	if opts.Command != "npm run build" {
		t.Errorf("unexpected Command: %s", opts.Command)
	}
	if opts.Debounce != 100 {
		t.Errorf("expected Debounce to be 100, got %d", opts.Debounce)
	}
	if !opts.Recursive {
		t.Error("expected Recursive to be true")
	}
	if !opts.Verbose {
		t.Error("expected Verbose to be true")
	}
}

func TestBuildResult_Fields(t *testing.T) {
	result := BuildResult{
		Success:  true,
		Duration: 5 * time.Second,
		Output:   "Build completed",
		Steps:    []BuildStep{},
	}

	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Duration != 5*time.Second {
		t.Errorf("expected Duration to be 5s, got %v", result.Duration)
	}
	if result.Output != "Build completed" {
		t.Errorf("unexpected Output: %s", result.Output)
	}
}

func TestBuildStep_Fields(t *testing.T) {
	step := BuildStep{
		Name:     "Install dependencies",
		Command:  "composer install",
		Success:  true,
		Duration: 30 * time.Second,
		Output:   "Dependencies installed",
	}

	if step.Name != "Install dependencies" {
		t.Errorf("unexpected Name: %s", step.Name)
	}
	if step.Command != "composer install" {
		t.Errorf("unexpected Command: %s", step.Command)
	}
	if !step.Success {
		t.Error("expected Success to be true")
	}
	if step.Duration != 30*time.Second {
		t.Errorf("expected Duration to be 30s, got %v", step.Duration)
	}
}

func TestComposerJSON_Fields(t *testing.T) {
	cj := ComposerJSON{
		Name:        "vendor/package",
		Type:        "library",
		Description: "A test package",
		Require: map[string]string{
			"php": ">=8.0",
		},
		RequireDev: map[string]string{
			"phpunit/phpunit": "^9.0",
		},
		Scripts: map[string]interface{}{
			"test": "phpunit",
		},
		Config: map[string]interface{}{
			"sort-packages": true,
		},
		Autoload: map[string]interface{}{
			"psr-4": map[string]string{
				"Vendor\\Package\\": "src/",
			},
		},
	}

	if cj.Name != "vendor/package" {
		t.Errorf("unexpected Name: %s", cj.Name)
	}
	if cj.Type != "library" {
		t.Errorf("unexpected Type: %s", cj.Type)
	}
	if cj.Require["php"] != ">=8.0" {
		t.Errorf("unexpected Require[php]: %s", cj.Require["php"])
	}
	if cj.RequireDev["phpunit/phpunit"] != "^9.0" {
		t.Errorf("unexpected RequireDev: %s", cj.RequireDev["phpunit/phpunit"])
	}
	if cj.Config == nil {
		t.Error("expected Config to be set")
	}
	if cj.Autoload == nil {
		t.Error("expected Autoload to be set")
	}
}

func TestPackageJSON_Fields(t *testing.T) {
	pj := PackageJSON{
		Name:        "my-package",
		Version:     "1.0.0",
		Description: "A test package",
		Scripts: map[string]string{
			"build": "webpack",
			"test":  "jest",
		},
		Dependencies: map[string]string{
			"react": "^18.0.0",
		},
		DevDependencies: map[string]string{
			"webpack": "^5.0.0",
		},
		Engines: map[string]string{
			"node": ">=18.0.0",
		},
	}

	if pj.Name != "my-package" {
		t.Errorf("unexpected Name: %s", pj.Name)
	}
	if pj.Version != "1.0.0" {
		t.Errorf("unexpected Version: %s", pj.Version)
	}
	if pj.Scripts["build"] != "webpack" {
		t.Errorf("unexpected Scripts[build]: %s", pj.Scripts["build"])
	}
	if pj.Dependencies["react"] != "^18.0.0" {
		t.Errorf("unexpected Dependencies[react]: %s", pj.Dependencies["react"])
	}
	if pj.Engines["node"] != ">=18.0.0" {
		t.Errorf("unexpected Engines[node]: %s", pj.Engines["node"])
	}
}

func TestHuskyConfig_Fields(t *testing.T) {
	config := HuskyConfig{
		Enabled:    true,
		ConfigFile: ".husky",
		PreCommit:  "npm run lint",
		PrePush:    "npm test",
		CommitMsg:  "commitlint --edit $1",
	}

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}
	if config.ConfigFile != ".husky" {
		t.Errorf("unexpected ConfigFile: %s", config.ConfigFile)
	}
	if config.PreCommit != "npm run lint" {
		t.Errorf("unexpected PreCommit: %s", config.PreCommit)
	}
	if config.PrePush != "npm test" {
		t.Errorf("unexpected PrePush: %s", config.PrePush)
	}
	if config.CommitMsg != "commitlint --edit $1" {
		t.Errorf("unexpected CommitMsg: %s", config.CommitMsg)
	}
}
