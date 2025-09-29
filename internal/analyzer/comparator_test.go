package analyzer

import (
	"testing"
)

func TestComparator_Compare(t *testing.T) {
	first := Aggregated{
		NbFiles:     10,
		NbFunctions: 20,
	}

	second := Aggregated{
		NbFiles:     5,
		NbFunctions: 7,
	}

	comparator := Comparator{}
	result := comparator.Compare(first, second)

	if result.NbFiles != 5 {
		t.Errorf("Expected NbFiles to be 5, got %d", result.NbFiles)
	}

	if result.NbFunctions != 13 {
		t.Errorf("Expected NbFunctions to be 13, got %d", result.NbFunctions)
	}
}
