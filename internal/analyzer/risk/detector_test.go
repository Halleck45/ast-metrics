package risk

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/pb"
)

func TestRiskItem_Structure(t *testing.T) {
	risk := RiskItem{
		ID:       "test-risk",
		Title:    "Test Risk",
		Severity: 0.8,
		Details:  "Test details",
	}

	if risk.ID != "test-risk" {
		t.Errorf("expected ID 'test-risk', got %s", risk.ID)
	}
	if risk.Severity != 0.8 {
		t.Errorf("expected severity 0.8, got %f", risk.Severity)
	}
}

// Mock detector for testing
type mockDetector struct {
	name  string
	risks []RiskItem
}

func (m *mockDetector) Name() string {
	return m.name
}

func (m *mockDetector) Detect(file *pb.File) []RiskItem {
	return m.risks
}

func TestDetector_Interface(t *testing.T) {
	risks := []RiskItem{
		{ID: "risk1", Title: "Risk 1", Severity: 0.5},
		{ID: "risk2", Title: "Risk 2", Severity: 0.9},
	}

	detector := &mockDetector{
		name:  "test-detector",
		risks: risks,
	}

	if detector.Name() != "test-detector" {
		t.Errorf("expected name 'test-detector', got %s", detector.Name())
	}

	file := &pb.File{Path: "/test/file.go"}
	detected := detector.Detect(file)

	if len(detected) != 2 {
		t.Fatalf("expected 2 risks, got %d", len(detected))
	}

	if detected[0].ID != "risk1" {
		t.Errorf("expected first risk ID 'risk1', got %s", detected[0].ID)
	}
	if detected[1].Severity != 0.9 {
		t.Errorf("expected second risk severity 0.9, got %f", detected[1].Severity)
	}
}
