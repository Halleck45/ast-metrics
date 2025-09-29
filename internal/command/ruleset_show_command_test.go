package command

import (
	"errors"
	"os"
	"testing"
)

func TestRulesetShowCommand_Execute_Success(t *testing.T) {
	cmd := NewRulesetShowCommand("volume")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRulesetShowCommand_Execute_NotFound(t *testing.T) {
	cmd := NewRulesetShowCommand("unknown-ruleset")
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for unknown ruleset")
	}
	if !contains(err.Error(), "not found") && !errors.Is(err, os.ErrNotExist) {
		// basic message check
		t.Fatalf("unexpected error: %v", err)
	}
}
