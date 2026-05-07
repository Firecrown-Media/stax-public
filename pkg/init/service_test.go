package staxinit_test

import (
	"testing"

	staxinit "github.com/firecrown-media/stax/pkg/init"
)

func TestValidateOptions_MissingName(t *testing.T) {
	opts := staxinit.Options{Name: ""}
	err := staxinit.ValidateOptions(opts)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

func TestValidateOptions_ValidName(t *testing.T) {
	opts := staxinit.Options{Name: "my-project"}
	err := staxinit.ValidateOptions(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
