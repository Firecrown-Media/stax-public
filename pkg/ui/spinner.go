package ui

import (
	"time"

	"github.com/briandowns/spinner"
)

// Spinner wraps the briandowns/spinner library
type Spinner struct {
	s *spinner.Spinner
}

// NewSpinner creates a new spinner with a message
func NewSpinner(message string) *Spinner {
	if quiet {
		return &Spinner{s: nil}
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	// Ignore color errors - it's cosmetic UI
	_ = s.Color("cyan")

	return &Spinner{s: s}
}

// Start starts the spinner
func (sp *Spinner) Start() {
	if sp.s != nil && !quiet {
		sp.s.Start()
	}
}

// Stop stops the spinner
func (sp *Spinner) Stop() {
	if sp.s != nil {
		sp.s.Stop()
	}
}

// Success stops the spinner and shows a success message
func (sp *Spinner) Success(message string) {
	if sp.s != nil {
		sp.s.Stop()
	}
	Success(message)
}

// Error stops the spinner and shows an error message
func (sp *Spinner) Error(message string) {
	if sp.s != nil {
		sp.s.Stop()
	}
	Error(message)
}

// UpdateMessage updates the spinner message
func (sp *Spinner) UpdateMessage(message string) {
	if sp.s != nil {
		sp.s.Suffix = " " + message
	}
}

// WithSpinner runs a function with a spinner
func WithSpinner(message string, fn func() error) error {
	s := NewSpinner(message)
	s.Start()
	defer s.Stop()

	err := fn()
	if err != nil {
		s.Error(err.Error())
		return err
	}

	s.Success(message + " - Done!")
	return nil
}
