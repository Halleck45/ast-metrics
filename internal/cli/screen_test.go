package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
)

// Mock screen for testing
type mockScreen struct {
	name  string
	model tea.Model
	reset bool
}

func (m *mockScreen) GetModel() tea.Model {
	return m.model
}

func (m *mockScreen) GetScreenName() string {
	return m.name
}

func (m *mockScreen) Reset(files []*pb.File, projectAggregated analyzer.ProjectAggregated) {
	m.reset = true
}

func TestScreen_Interface(t *testing.T) {
	screen := &mockScreen{name: "test-screen"}

	if screen.GetScreenName() != "test-screen" {
		t.Errorf("expected 'test-screen', got %s", screen.GetScreenName())
	}

	screen.Reset(nil, analyzer.ProjectAggregated{})
	if !screen.reset {
		t.Error("expected screen to be reset")
	}
}
