package python

import (
	"os"
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/pterm/pterm"
)

func TestPythonRunner_SetProgressbar(t *testing.T) {
	runner := &PythonRunner{}
	progressbar := &pterm.SpinnerPrinter{}
	
	runner.SetProgressbar(progressbar)
	
	if runner.progressbar != progressbar {
		t.Error("expected progressbar to be set")
	}
}

func TestPythonRunner_SetConfiguration(t *testing.T) {
	runner := &PythonRunner{}
	config := &configuration.Configuration{}
	
	runner.SetConfiguration(config)
	
	if runner.Configuration != config {
		t.Error("expected configuration to be set")
	}
}

func TestPythonRunner_Parse_NonExistentFile(t *testing.T) {
	runner := &PythonRunner{}
	
	_, err := runner.Parse("/nonexistent/file.py")
	
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestPythonRunner_Parse_ValidFile(t *testing.T) {
	runner := &PythonRunner{}
	
	// Create temporary Python file
	tmpFile, err := os.CreateTemp("", "test*.py")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	
	_, err = tmpFile.WriteString("def hello():\n    print('world')")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	
	file, err := runner.Parse(tmpFile.Name())
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if file == nil {
		t.Error("expected non-nil file")
	}
}
