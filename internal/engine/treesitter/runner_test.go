package treesitter

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
	sitter "github.com/smacker/go-tree-sitter"
)

// Mock adapter for testing
type mockAdapter struct{}

func (m *mockAdapter) Language() *sitter.Language                                { return nil }
func (m *mockAdapter) IsModule(*sitter.Node) bool                               { return false }
func (m *mockAdapter) IsClass(*sitter.Node) bool                                { return false }
func (m *mockAdapter) IsFunction(*sitter.Node) bool                             { return false }
func (m *mockAdapter) NodeName(*sitter.Node) string                             { return "" }
func (m *mockAdapter) NodeBody(*sitter.Node) *sitter.Node                       { return nil }
func (m *mockAdapter) NodeParams(*sitter.Node) *sitter.Node                     { return nil }
func (m *mockAdapter) ModuleNameFromPath(path string) string                    { return "" }
func (m *mockAdapter) AttachQualified(parentClass, fn string) string           { return "" }
func (m *mockAdapter) EachChildBody(n *sitter.Node, yield func(*sitter.Node))  {}
func (m *mockAdapter) EachParamIdent(params *sitter.Node, yield func(string))  {}
func (m *mockAdapter) Decision(n *sitter.Node) (DecisionKind, *sitter.Node)    { return DecNone, nil }
func (m *mockAdapter) Imports(n *sitter.Node) []ImportItem                     { return nil }

func TestRunner_ParseAndStore_NonExistentFile(t *testing.T) {
	runner := Runner{
		Adapter:       &mockAdapter{},
		Configuration: &configuration.Configuration{},
	}

	err := runner.ParseAndStore("/nonexistent/file.go")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestRunner_WalkAndProcess_EmptyList(t *testing.T) {
	runner := Runner{
		Adapter:       &mockAdapter{},
		Configuration: &configuration.Configuration{},
	}

	var messages []string
	runner.UpdateText = func(msg string) {
		messages = append(messages, msg)
	}

	runner.WalkAndProcess([]string{})

	if len(messages) != 0 {
		t.Errorf("expected 0 messages for empty list, got %d", len(messages))
	}
}

func TestRunner_WalkAndProcess_WithFiles(t *testing.T) {
	runner := Runner{
		Adapter:       &mockAdapter{},
		Configuration: &configuration.Configuration{},
	}

	var messages []string
	runner.UpdateText = func(msg string) {
		messages = append(messages, msg)
	}

	files := []string{"/test/file1.go", "/test/file2.go"}
	runner.WalkAndProcess(files)

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}
