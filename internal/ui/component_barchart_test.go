package ui

import (
	"strings"
	"testing"
)

func TestComponentBarchart_AsTerminalElement(t *testing.T) {
	chart := &ComponentBarchart{
		data: map[string]float64{
			"A": 10.0,
			"B": 20.0,
			"C": 15.0,
		},
		height:   3,
		barWidth: 8,
	}

	result := chart.AsTerminalElement()

	if result == "" {
		t.Error("expected non-empty terminal element")
	}

	// Should contain some chart representation
	if !strings.Contains(result, "A") || !strings.Contains(result, "B") || !strings.Contains(result, "C") {
		t.Error("expected chart to contain data labels")
	}
}

func TestComponentBarchart_DefaultValues(t *testing.T) {
	chart := &ComponentBarchart{
		data: map[string]float64{"Test": 5.0},
	}

	result := chart.AsTerminalElement()

	if result == "" {
		t.Error("expected chart with default values to work")
	}

	// Check that defaults were applied
	if chart.height != 3 {
		t.Errorf("expected default height 3, got %d", chart.height)
	}
	if chart.barWidth != 8 {
		t.Errorf("expected default barWidth 8, got %d", chart.barWidth)
	}
}

func TestComponentBarchart_EmptyData(t *testing.T) {
	chart := &ComponentBarchart{
		data: map[string]float64{},
	}

	result := chart.AsTerminalElement()

	// Should not panic with empty data
	if result == "" {
		t.Error("expected some output even with empty data")
	}
}
