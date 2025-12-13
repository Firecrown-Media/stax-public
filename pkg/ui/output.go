package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

var (
	verbose bool
	debug   bool
	quiet   bool
	noColor bool
)

// Color functions
var (
	colorSuccess = color.New(color.FgGreen, color.Bold)
	colorError   = color.New(color.FgRed, color.Bold)
	colorWarning = color.New(color.FgYellow, color.Bold)
	colorInfo    = color.New(color.FgCyan)
	colorDebug   = color.New(color.FgMagenta)
	colorHeader  = color.New(color.FgBlue, color.Bold)
	colorSection = color.New(color.FgWhite, color.Bold)
)

// SetVerbose enables verbose output
func SetVerbose(v bool) {
	verbose = v
}

// SetDebug enables debug output
func SetDebug(d bool) {
	debug = d
	if d {
		verbose = true // Debug implies verbose
	}
}

// SetQuiet suppresses non-error output
func SetQuiet(q bool) {
	quiet = q
}

// SetNoColor disables colored output
func SetNoColor(nc bool) {
	noColor = nc
	if nc {
		color.NoColor = true
	}
}

// Success prints a success message
func Success(message string, args ...interface{}) {
	if quiet {
		return
	}
	msg := fmt.Sprintf(message, args...)
	colorSuccess.Fprintf(os.Stdout, "✓ %s\n", msg)
}

// Error prints an error message
func Error(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	colorError.Fprintf(os.Stderr, "✗ %s\n", msg)
}

// Warning prints a warning message
func Warning(message string, args ...interface{}) {
	if quiet {
		return
	}
	msg := fmt.Sprintf(message, args...)
	colorWarning.Fprintf(os.Stdout, "⚠ %s\n", msg)
}

// Info prints an info message
func Info(message string, args ...interface{}) {
	if quiet {
		return
	}
	msg := fmt.Sprintf(message, args...)
	colorInfo.Fprintf(os.Stdout, "  %s\n", msg)
}

// Debug prints a debug message (only if debug is enabled)
func Debug(message string, args ...interface{}) {
	if !debug {
		return
	}
	msg := fmt.Sprintf(message, args...)
	colorDebug.Fprintf(os.Stdout, "[DEBUG] %s\n", msg)
}

// Verbose prints a message only if verbose mode is enabled
func Verbose(message string, args ...interface{}) {
	if !verbose {
		return
	}
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(os.Stdout, "  %s\n", msg)
}

// PrintHeader prints a section header
func PrintHeader(message string) {
	if quiet {
		return
	}
	fmt.Println()
	colorHeader.Fprintf(os.Stdout, "==> %s\n", message)
	fmt.Println()
}

// Section prints a section separator
func Section(message string) {
	if quiet {
		return
	}
	fmt.Println()
	colorSection.Fprintf(os.Stdout, "%s\n", message)
}

// Print prints a plain message (respects quiet mode)
func Print(message string, args ...interface{}) {
	if quiet {
		return
	}
	fmt.Fprintf(os.Stdout, message, args...)
}

// Println prints a plain message with newline (respects quiet mode)
func Println(message string, args ...interface{}) {
	if quiet {
		return
	}
	fmt.Fprintf(os.Stdout, message+"\n", args...)
}

// ErrorAndExit prints an error message and exits with code 1
func ErrorAndExit(message string, args ...interface{}) {
	Error(message, args...)
	os.Exit(1)
}

// Confirm prompts the user for yes/no confirmation
func Confirm(message string) bool {
	if quiet {
		return false
	}

	fmt.Fprintf(os.Stdout, "%s (yes/no): ", message)

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}

	return response == "yes" || response == "y"
}

// PromptString prompts the user for a string input
func PromptString(message string, defaultValue string) string {
	if quiet {
		return defaultValue
	}

	if defaultValue != "" {
		fmt.Fprintf(os.Stdout, "%s [%s]: ", message, defaultValue)
	} else {
		fmt.Fprintf(os.Stdout, "%s: ", message)
	}

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil || response == "" {
		return defaultValue
	}

	return response
}

// ProgressMsg prints a progress message (step in a process)
func ProgressMsg(message string, args ...interface{}) {
	if quiet {
		return
	}
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(os.Stdout, "  - %s\n", msg)
}

// PromptPassword prompts the user for a password input (hidden/obfuscated)
func PromptPassword(message string) string {
	if quiet {
		return ""
	}

	fmt.Fprintf(os.Stdout, "%s: ", message)

	// Use term.ReadPassword to hide input as it's typed
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // Print newline after password entry

	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(passwordBytes))
}
