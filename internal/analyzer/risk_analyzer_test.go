package analyzer

import (
	"math"
	"testing"
)

func TestGetRisk_BasicNonZero(t *testing.T) {
	ra := NewRiskAnalyzer()
	// maxCommits 10, maxComplexity 100, nbCommits 5, complexity 50 => centered, distance sqrt(0.5^2+0.5^2)=~0.7071, risk ~0.2929
	risk := ra.GetRisk(10, 100, 5, 50)
	if math.IsNaN(risk) || math.IsInf(risk, 0) {
		t.Fatalf("risk is invalid: %v", risk)
	}
	if risk <= 0 || risk >= 1 {
		t.Fatalf("expected risk in (0,1), got %v", risk)
	}
}

func TestGetRisk_ZeroMaxCommits(t *testing.T) {
	ra := NewRiskAnalyzer()
	// No commits in project but some complexity should still yield a valid score based on complexity axis alone
	risk := ra.GetRisk(0, 100, 0, 80)
	if math.IsNaN(risk) || math.IsInf(risk, 0) {
		t.Fatalf("risk is invalid: %v", risk)
	}
	if risk < 0 || risk > 1 {
		t.Fatalf("expected risk in [0,1], got %v", risk)
	}
}

func TestGetRisk_ZeroMaxComplexity(t *testing.T) {
	ra := NewRiskAnalyzer()
	// No complexity measured but churn exists; should rely on commits axis
	risk := ra.GetRisk(10, 0, 7, 0)
	if math.IsNaN(risk) || math.IsInf(risk, 0) {
		t.Fatalf("risk is invalid: %v", risk)
	}
	if risk < 0 || risk > 1 {
		t.Fatalf("expected risk in [0,1], got %v", risk)
	}
}

func TestGetRisk_TopRightCornerHighRisk(t *testing.T) {
	ra := NewRiskAnalyzer()
	// At top-right corner (nbCommits=maxCommits, complexity=maxComplexity) => distance 0 => risk 1
	risk := ra.GetRisk(10, 100, 10, 100)
	if math.Abs(risk-1) > 1e-9 {
		t.Fatalf("expected risk ~1, got %v", risk)
	}
}
