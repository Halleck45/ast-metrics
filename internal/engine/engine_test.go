package engine

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/pterm/pterm"
)

// Mock engine for testing
type mockEngine struct {
	required bool
	name     string
}

func (m *mockEngine) IsRequired() bool                                     { return m.required }
func (m *mockEngine) Ensure() error                                        { return nil }
func (m *mockEngine) DumpAST()                                             {}
func (m *mockEngine) Finish() error                                        { return nil }
func (m *mockEngine) SetProgressbar(progressbar *pterm.SpinnerPrinter)     {}
func (m *mockEngine) SetConfiguration(configuration *configuration.Configuration) {}
func (m *mockEngine) Parse(filepath string) (*pb.File, error) {
	return &pb.File{Path: filepath}, nil
}
func (m *mockEngine) Name() string { return m.name }

func TestEngine_Interface(t *testing.T) {
	engine := &mockEngine{required: true, name: "test-engine"}

	if !engine.IsRequired() {
		t.Error("expected engine to be required")
	}

	if err := engine.Ensure(); err != nil {
		t.Errorf("expected no error from Ensure, got %v", err)
	}

	if err := engine.Finish(); err != nil {
		t.Errorf("expected no error from Finish, got %v", err)
	}

	file, err := engine.Parse("/test/file.go")
	if err != nil {
		t.Errorf("expected no error from Parse, got %v", err)
	}
	if file.Path != "/test/file.go" {
		t.Errorf("expected path '/test/file.go', got %s", file.Path)
	}
}

func TestNamedEngine_Interface(t *testing.T) {
	engine := &mockEngine{name: "test-engine"}

	if engine.Name() != "test-engine" {
		t.Errorf("expected name 'test-engine', got %s", engine.Name())
	}
}
