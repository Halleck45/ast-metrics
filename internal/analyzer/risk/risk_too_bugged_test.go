package risk

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestTooBuggedDetector_Name(t *testing.T) {
	detector := &TooBuggedDetector{}
	if detector.Name() != "risk_too_bugged" {
		t.Errorf("expected 'risk_too_bugged', got %s", detector.Name())
	}
}

func TestTooBuggedDetector_Detect_HighBugs(t *testing.T) {
	detector := &TooBuggedDetector{}
	bugs := 1.2
	
	file := &pb.File{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Volume: &pb.Volume{HalsteadBugs: &bugs},
			},
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 1 {
		t.Fatalf("expected 1 risk, got %d", len(risks))
	}

	risk := risks[0]
	if risk.Title != "High estimated bugs (Halstead)" {
		t.Errorf("expected title 'High estimated bugs (Halstead)', got %s", risk.Title)
	}
}

func TestTooBuggedDetector_Detect_LowBugs(t *testing.T) {
	detector := &TooBuggedDetector{}
	bugs := 0.2
	
	file := &pb.File{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Volume: &pb.Volume{HalsteadBugs: &bugs},
			},
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 0 {
		t.Errorf("expected 0 risks for low bugs, got %d", len(risks))
	}
}

func TestClamp01Float(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{-0.5, 0.0},
		{0.5, 0.5},
		{1.5, 1.0},
		{0.0, 0.0},
		{1.0, 1.0},
	}

	for _, test := range tests {
		result := clamp01Float(test.input)
		if result != test.expected {
			t.Errorf("clamp01Float(%f) = %f, expected %f", test.input, result, test.expected)
		}
	}
}
