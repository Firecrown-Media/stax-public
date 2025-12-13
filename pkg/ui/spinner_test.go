package ui

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewSpinner(t *testing.T) {
	resetState()
	defer resetState()

	sp := NewSpinner("Loading...")
	if sp == nil {
		t.Fatal("NewSpinner should not return nil")
	}

	if sp.s == nil {
		t.Error("spinner internal should not be nil in normal mode")
	}
}

func TestNewSpinner_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	sp := NewSpinner("Loading...")
	if sp == nil {
		t.Fatal("NewSpinner should not return nil even in quiet mode")
	}

	if sp.s != nil {
		t.Error("spinner internal should be nil in quiet mode")
	}
}

func TestSpinner_StartStop(t *testing.T) {
	resetState()
	defer resetState()

	sp := NewSpinner("Processing...")

	// Should not panic
	sp.Start()
	time.Sleep(10 * time.Millisecond)
	sp.Stop()
}

func TestSpinner_StartStop_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	sp := NewSpinner("Processing...")

	// Should not panic even with nil internal spinner
	sp.Start()
	sp.Stop()
}

func TestSpinner_Success(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	sp := NewSpinner("Loading...")
	sp.Start()

	stdout, _ := captureOutput(func() {
		sp.Success("Completed successfully")
	})

	if !strings.Contains(stdout, "Completed successfully") {
		t.Errorf("Spinner.Success should output message, got: %s", stdout)
	}
}

func TestSpinner_Success_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	sp := NewSpinner("Loading...")
	sp.Start()

	stdout, _ := captureOutput(func() {
		sp.Success("This should not appear")
	})

	// Success is suppressed in quiet mode
	if stdout != "" {
		t.Errorf("Spinner.Success should not output in quiet mode, got: %s", stdout)
	}
}

func TestSpinner_Error(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	sp := NewSpinner("Loading...")
	sp.Start()

	_, stderr := captureOutput(func() {
		sp.Error("Something failed")
	})

	if !strings.Contains(stderr, "Something failed") {
		t.Errorf("Spinner.Error should output message, got: %s", stderr)
	}
}

func TestSpinner_Error_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)
	SetNoColor(true)

	sp := NewSpinner("Loading...")
	sp.Start()

	_, stderr := captureOutput(func() {
		sp.Error("Error should still appear")
	})

	// Errors should still appear in quiet mode
	if !strings.Contains(stderr, "Error should still appear") {
		t.Errorf("Spinner.Error should output even in quiet mode, got: %s", stderr)
	}
}

func TestSpinner_UpdateMessage(t *testing.T) {
	resetState()
	defer resetState()

	sp := NewSpinner("Initial message")
	sp.Start()

	// Should not panic
	sp.UpdateMessage("Updated message")

	sp.Stop()
}

func TestSpinner_UpdateMessage_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	sp := NewSpinner("Initial message")
	sp.Start()

	// Should not panic with nil internal spinner
	sp.UpdateMessage("Updated message")

	sp.Stop()
}

func TestWithSpinner_Success(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	called := false
	stdout, _ := captureOutput(func() {
		err := WithSpinner("Test operation", func() error {
			called = true
			return nil
		})
		if err != nil {
			t.Errorf("WithSpinner should not return error on success, got: %v", err)
		}
	})

	if !called {
		t.Error("WithSpinner should call the function")
	}

	if !strings.Contains(stdout, "Done!") {
		t.Errorf("WithSpinner should output success message, got: %s", stdout)
	}
}

func TestWithSpinner_Error(t *testing.T) {
	resetState()
	defer resetState()
	SetNoColor(true)

	expectedErr := errors.New("operation failed")

	_, stderr := captureOutput(func() {
		err := WithSpinner("Test operation", func() error {
			return expectedErr
		})
		if err != expectedErr {
			t.Errorf("WithSpinner should return the error, got: %v", err)
		}
	})

	if !strings.Contains(stderr, "operation failed") {
		t.Errorf("WithSpinner should output error message, got: %s", stderr)
	}
}

func TestWithSpinner_QuietMode(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	called := false
	err := WithSpinner("Test operation", func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("WithSpinner should not return error, got: %v", err)
	}

	if !called {
		t.Error("WithSpinner should still call the function in quiet mode")
	}
}

func TestSpinner_NilSafeOperations(t *testing.T) {
	resetState()
	defer resetState()
	SetQuiet(true)

	sp := NewSpinner("Test")

	// All operations should be safe with nil internal spinner
	sp.Start()
	sp.UpdateMessage("New message")
	sp.Stop()

	// Success and Error should work but respect quiet mode
	_, stderr := captureOutput(func() {
		sp.Success("Success message") // Won't output in quiet mode
		sp.Error("Error message")     // Will output (errors not suppressed)
	})

	// Error should still appear
	if !strings.Contains(stderr, "Error message") {
		t.Errorf("Error should appear even in quiet mode, got: %s", stderr)
	}
}

func TestSpinner_MultipleStartStop(t *testing.T) {
	resetState()
	defer resetState()

	sp := NewSpinner("Test")

	// Multiple start/stop cycles should not panic
	for i := 0; i < 3; i++ {
		sp.Start()
		time.Sleep(5 * time.Millisecond)
		sp.Stop()
	}
}
