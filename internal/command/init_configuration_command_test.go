package command

import (
	"os"
	"testing"
)

func TestNewInitConfigurationCommand(t *testing.T) {
	cmd := NewInitConfigurationCommand()
	if cmd == nil {
		t.Error("expected non-nil command")
	}
}

func TestInitConfigurationCommand_Execute(t *testing.T) {
	// Change to temp directory to avoid creating config file in project
	originalDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	cmd := NewInitConfigurationCommand()
	err := cmd.Execute()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Check if config file was created
	if _, err := os.Stat(".ast-metrics.yaml"); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}
