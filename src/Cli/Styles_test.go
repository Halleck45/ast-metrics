package Cli

import (
	"testing"
)

func TestStyleTitle(t *testing.T) {
	style := StyleTitle("Hello")

	if style.GetWidth() != 100 {
		t.Errorf("Expected correct width', got %d", style.GetWidth())
	}

	// Add more assertions here for the other properties of the style...
}

func TestDecorateMaintainabilityIndex(t *testing.T) {
	if DecorateMaintainabilityIndex(63, nil) != "🔴 63" {
		t.Errorf("Expected '🔴 63', got '%s'", DecorateMaintainabilityIndex(63, nil))
	}

	if DecorateMaintainabilityIndex(84, nil) != "🟡 84" {
		t.Errorf("Expected '🟡 84', got '%s'", DecorateMaintainabilityIndex(84, nil))
	}

	if DecorateMaintainabilityIndex(85, nil) != "🟢 85" {
		t.Errorf("Expected '🟢 85', got '%s'", DecorateMaintainabilityIndex(85, nil))
	}
}

func TestRound(t *testing.T) {
	if Round(1.4) != 1 {
		t.Errorf("Expected 1, got %d", Round(1.4))
	}

	if Round(1.5) != 2 {
		t.Errorf("Expected 2, got %d", Round(1.5))
	}
}

func TestToFixed(t *testing.T) {
	if ToFixed(1.2345, 2) != 1.23 {
		t.Errorf("Expected 1.23, got %f", ToFixed(1.2345, 2))
	}

	if ToFixed(1.2345, 3) != 1.235 {
		t.Errorf("Expected 1.235, got %f", ToFixed(1.2345, 3))
	}
}
