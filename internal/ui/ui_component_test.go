package ui

import "testing"

// Mock implementation for testing
type mockUiComponent struct{}

func (m *mockUiComponent) AsTerminalElement() string {
	return "terminal output"
}

func (m *mockUiComponent) AsHtml() string {
	return "<div>html output</div>"
}

func TestUiComponent_Interface(t *testing.T) {
	var component UiComponent = &mockUiComponent{}
	
	terminal := component.AsTerminalElement()
	if terminal != "terminal output" {
		t.Errorf("expected 'terminal output', got %s", terminal)
	}
	
	html := component.AsHtml()
	if html != "<div>html output</div>" {
		t.Errorf("expected '<div>html output</div>', got %s", html)
	}
}
