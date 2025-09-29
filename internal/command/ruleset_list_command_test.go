package command

import (
	"testing"
)

func TestRulesetListCommand_Execute_NoError(t *testing.T) {
	cmd := NewRulesetListCommand()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
