package rust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestRustRunner_Name(t *testing.T) {
	runner := RustRunner{}
	if runner.Name() != "Rust" {
		t.Errorf("expected 'Rust', got %s", runner.Name())
	}
}

func TestRustRunner_IsRequired_NoFiles(t *testing.T) {
	runner := RustRunner{
		Configuration: &configuration.Configuration{},
	}
	
	if runner.IsRequired() {
		t.Error("expected IsRequired to be false when no Rust files found")
	}
}

func TestRustRunner_Parse_ValidRustFile(t *testing.T) {
	// Create temporary Rust file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.rs")
	
	rustCode := `fn main() {
    println!("Hello, world!");
}`
	
	err := os.WriteFile(tmpFile, []byte(rustCode), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := RustRunner{}
	file, err := runner.Parse(tmpFile)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if file.Path != tmpFile {
		t.Errorf("expected path %s, got %s", tmpFile, file.Path)
	}

	if file.ProgrammingLanguage != "Rust" {
		t.Errorf("expected language 'Rust', got %s", file.ProgrammingLanguage)
	}
}

func TestRustRunner_Parse_NonExistentFile(t *testing.T) {
	runner := RustRunner{}
	file, err := runner.Parse("/nonexistent/file.rs")

	if err == nil {
		t.Error("expected error for non-existent file")
	}

	if file.Path != "/nonexistent/file.rs" {
		t.Errorf("expected path to be preserved even on error")
	}

	if file.ProgrammingLanguage != "Rust" {
		t.Errorf("expected language 'Rust' even on error, got %s", file.ProgrammingLanguage)
	}
}

func TestRustRunner_Ensure(t *testing.T) {
	runner := RustRunner{}
	err := runner.Ensure()
	if err != nil {
		t.Errorf("expected no error from Ensure, got %v", err)
	}
}

func TestRustRunner_Finish(t *testing.T) {
	runner := RustRunner{}
	err := runner.Finish()
	if err != nil {
		t.Errorf("expected no error from Finish, got %v", err)
	}
}
