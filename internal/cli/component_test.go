package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Mock component for testing
type mockComponent struct {
	content string
	updated bool
}

func (m *mockComponent) Render() string {
	return m.content
}

func (m *mockComponent) Update(msg tea.Msg) {
	m.updated = true
}

func TestComponent_Interface(t *testing.T) {
	component := &mockComponent{content: "test content"}

	if component.Render() != "test content" {
		t.Errorf("expected 'test content', got %s", component.Render())
	}

	component.Update(nil)
	if !component.updated {
		t.Error("expected component to be updated")
	}
}
