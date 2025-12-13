package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

// Helper to capture stdout/stderr
func captureOutput(fn func()) (string, string) {
	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	// Run the function
	fn()

	// Close writers and restore
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String()
}

// resetState resets all package-level state for test isolation
func resetState() {
	verbose = false
	debug = false
	quiet = false
	noColor = false
	color.NoColor = false
}

func TestSetVerbose(t *testing.T) {
	resetState()
	defer resetState()

	if verbose {
		t.Error("verbose should be false initially")
	}

	SetVerbose(true)
	if !verbose {
		t.Error("verbose should be true after SetVerbose(true)")
	}

	SetVerbose(false)
	if verbose {
		t.Error("verbose should be false after SetVerbose(false)")
	}
}

func TestSetDebug(t *testing.T) {
	resetState()
	defer resetState()

	if debug {
		t.Error("debug should be false initially")
	}

	SetDebug(true)
	if !debug {
		t.Error("debug should be true after SetDebug(true)")
	}
	if !verbose {
		t.Error("verbose should be true when debug is enabled")
	}
}

func TestSetQuiet(t *testing.T) {
	resetState()
	defer resetState()

	if quiet {
		t.Error("quiet should be false initially")
	}

	SetQuiet(true)
	if !quiet {
		t.Error("quiet should be true after SetQuiet(true)")
	}
}

func TestSetNoColor(t *testing.T) {
	resetState()
	defer resetState()

	SetNoColor(true)
	if !noColor {
		t.Error("noColor should be true after SetNoColor(true)")
	}
	if !color.NoColor {
		t.Error("color.NoColor should be true after SetNoColor(true)")
	}
}

func TestSuccess(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true) // Disable colors for easier testing

	stdout, _ := captureOutput(func() {
		Success("Operation completed")
	})

	if !strings.Contains(stdout, "Operation completed") {
		t.Errorf("Success output should contain message, got: %s", stdout)
	}
}

func TestSuccess_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	stdout, _ := captureOutput(func() {
		Success("This should not appear")
	})

	if stdout != "" {
		t.Errorf("Success should not output in quiet mode, got: %s", stdout)
	}
}

func TestSuccess_WithArgs(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	stdout, _ := captureOutput(func() {
		Success("Created %d files in %s", 5, "/tmp")
	})

	if !strings.Contains(stdout, "Created 5 files in /tmp") {
		t.Errorf("Success should format args, got: %s", stdout)
	}
}

func TestError(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	_, stderr := captureOutput(func() {
		Error("Something went wrong")
	})

	if !strings.Contains(stderr, "Something went wrong") {
		t.Errorf("Error output should contain message, got: %s", stderr)
	}
}

func TestError_NotSuppressedInQuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)
	SetNoColor(true)

	_, stderr := captureOutput(func() {
		Error("Error should still appear")
	})

	if !strings.Contains(stderr, "Error should still appear") {
		t.Errorf("Error should output even in quiet mode, got: %s", stderr)
	}
}

func TestWarning(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	stdout, _ := captureOutput(func() {
		Warning("Be careful")
	})

	if !strings.Contains(stdout, "Be careful") {
		t.Errorf("Warning output should contain message, got: %s", stdout)
	}
}

func TestWarning_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	stdout, _ := captureOutput(func() {
		Warning("This should not appear")
	})

	if stdout != "" {
		t.Errorf("Warning should not output in quiet mode, got: %s", stdout)
	}
}

func TestInfo(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	stdout, _ := captureOutput(func() {
		Info("Some information")
	})

	if !strings.Contains(stdout, "Some information") {
		t.Errorf("Info output should contain message, got: %s", stdout)
	}
}

func TestInfo_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	stdout, _ := captureOutput(func() {
		Info("This should not appear")
	})

	if stdout != "" {
		t.Errorf("Info should not output in quiet mode, got: %s", stdout)
	}
}

func TestDebug(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)
	SetDebug(true)

	stdout, _ := captureOutput(func() {
		Debug("Debug message")
	})

	if !strings.Contains(stdout, "Debug message") {
		t.Errorf("Debug output should contain message when debug enabled, got: %s", stdout)
	}
	if !strings.Contains(stdout, "[DEBUG]") {
		t.Errorf("Debug output should contain [DEBUG] prefix, got: %s", stdout)
	}
}

func TestDebug_DisabledByDefault(t *testing.T) {
	resetState()
	defer resetState()

	stdout, _ := captureOutput(func() {
		Debug("This should not appear")
	})

	if stdout != "" {
		t.Errorf("Debug should not output when debug disabled, got: %s", stdout)
	}
}

func TestVerbose(t *testing.T) {
	resetState()
	defer resetState()
	SetVerbose(true)

	stdout, _ := captureOutput(func() {
		Verbose("Verbose message")
	})

	if !strings.Contains(stdout, "Verbose message") {
		t.Errorf("Verbose output should contain message when verbose enabled, got: %s", stdout)
	}
}

func TestVerbose_DisabledByDefault(t *testing.T) {
	resetState()
	defer resetState()

	stdout, _ := captureOutput(func() {
		Verbose("This should not appear")
	})

	if stdout != "" {
		t.Errorf("Verbose should not output when verbose disabled, got: %s", stdout)
	}
}

func TestPrintHeader(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	stdout, _ := captureOutput(func() {
		PrintHeader("Section Header")
	})

	if !strings.Contains(stdout, "Section Header") {
		t.Errorf("PrintHeader output should contain message, got: %s", stdout)
	}
	if !strings.Contains(stdout, "==>") {
		t.Errorf("PrintHeader should have ==> prefix, got: %s", stdout)
	}
}

func TestPrintHeader_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	stdout, _ := captureOutput(func() {
		PrintHeader("This should not appear")
	})

	if stdout != "" {
		t.Errorf("PrintHeader should not output in quiet mode, got: %s", stdout)
	}
}

func TestSection(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	stdout, _ := captureOutput(func() {
		Section("Section Title")
	})

	if !strings.Contains(stdout, "Section Title") {
		t.Errorf("Section output should contain message, got: %s", stdout)
	}
}

func TestSection_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	stdout, _ := captureOutput(func() {
		Section("This should not appear")
	})

	if stdout != "" {
		t.Errorf("Section should not output in quiet mode, got: %s", stdout)
	}
}

func TestPrint(t *testing.T) {
	resetState()
	defer resetState()

	stdout, _ := captureOutput(func() {
		Print("Plain message")
	})

	if stdout != "Plain message" {
		t.Errorf("Print should output exact message, got: %s", stdout)
	}
}

func TestPrint_WithArgs(t *testing.T) {
	resetState()
	defer resetState()

	stdout, _ := captureOutput(func() {
		Print("Count: %d", 42)
	})

	if stdout != "Count: 42" {
		t.Errorf("Print should format args, got: %s", stdout)
	}
}

func TestPrint_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	stdout, _ := captureOutput(func() {
		Print("This should not appear")
	})

	if stdout != "" {
		t.Errorf("Print should not output in quiet mode, got: %s", stdout)
	}
}

func TestPrintln(t *testing.T) {
	resetState()
	defer resetState()

	stdout, _ := captureOutput(func() {
		Println("Line message")
	})

	if stdout != "Line message\n" {
		t.Errorf("Println should add newline, got: %q", stdout)
	}
}

func TestPrintln_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	stdout, _ := captureOutput(func() {
		Println("This should not appear")
	})

	if stdout != "" {
		t.Errorf("Println should not output in quiet mode, got: %s", stdout)
	}
}

func TestProgressMsg(t *testing.T) {
	resetState()
	defer resetState()

	stdout, _ := captureOutput(func() {
		ProgressMsg("Step 1 completed")
	})

	if !strings.Contains(stdout, "Step 1 completed") {
		t.Errorf("ProgressMsg should contain message, got: %s", stdout)
	}
	if !strings.Contains(stdout, "- ") {
		t.Errorf("ProgressMsg should have dash prefix, got: %s", stdout)
	}
}

func TestProgressMsg_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	stdout, _ := captureOutput(func() {
		ProgressMsg("This should not appear")
	})

	if stdout != "" {
		t.Errorf("ProgressMsg should not output in quiet mode, got: %s", stdout)
	}
}

func TestProgressMsg_WithArgs(t *testing.T) {
	resetState()
	defer resetState()

	stdout, _ := captureOutput(func() {
		ProgressMsg("Processing file %d of %d", 3, 10)
	})

	if !strings.Contains(stdout, "Processing file 3 of 10") {
		t.Errorf("ProgressMsg should format args, got: %s", stdout)
	}
}

func TestConfirm_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	result := Confirm("Do you want to proceed?")
	if result {
		t.Error("Confirm should return false in quiet mode")
	}
}

func TestPromptString_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	result := PromptString("Enter name", "default-value")
	if result != "default-value" {
		t.Errorf("PromptString should return default in quiet mode, got: %s", result)
	}
}

// Test color output behavior (basic verification)
func TestColorOutput(t *testing.T) {
	resetState()
	defer resetState()

	// Verify colors are defined
	if colorSuccess == nil {
		t.Error("colorSuccess should not be nil")
	}
	if colorError == nil {
		t.Error("colorError should not be nil")
	}
	if colorWarning == nil {
		t.Error("colorWarning should not be nil")
	}
	if colorInfo == nil {
		t.Error("colorInfo should not be nil")
	}
	if colorDebug == nil {
		t.Error("colorDebug should not be nil")
	}
	if colorHeader == nil {
		t.Error("colorHeader should not be nil")
	}
	if colorSection == nil {
		t.Error("colorSection should not be nil")
	}
}

func TestMultipleSetters(t *testing.T) {
	resetState()
	defer resetState()

	// Test setting multiple modes
	SetVerbose(true)
	SetQuiet(true)
	SetNoColor(true)

	if !verbose {
		t.Error("verbose should be true")
	}
	if !quiet {
		t.Error("quiet should be true")
	}
	if !noColor {
		t.Error("noColor should be true")
	}
}
