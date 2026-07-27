package cli

import (
	"testing"

	pb "github.com/ast-metrics/ast-metrics/pb"
)

func TestDecorateMaintainabilityIndex_HighValue(t *testing.T) {
	analyze := &pb.Analyze{
		Volume: &pb.Volume{
			Lloc: func() *int32 { v := int32(100); return &v }(),
		},
	}
	
	result := DecorateMaintainabilityIndex(90, analyze)
	if result != "🟢 90" {
		t.Errorf("expected '🟢 90', got %s", result)
	}
}

func TestDecorateMaintainabilityIndex_MediumValue(t *testing.T) {
	analyze := &pb.Analyze{
		Volume: &pb.Volume{
			Lloc: func() *int32 { v := int32(100); return &v }(),
		},
	}
	
	result := DecorateMaintainabilityIndex(75, analyze)
	if result != "🟡 75" {
		t.Errorf("expected '🟡 75', got %s", result)
	}
}

func TestDecorateMaintainabilityIndex_LowValue(t *testing.T) {
	analyze := &pb.Analyze{
		Volume: &pb.Volume{
			Lloc: func() *int32 { v := int32(100); return &v }(),
		},
	}
	
	result := DecorateMaintainabilityIndex(50, analyze)
	if result != "🔴 50" {
		t.Errorf("expected '🔴 50', got %s", result)
	}
}

func TestDecorateMaintainabilityIndex_LowLloc(t *testing.T) {
	analyze := &pb.Analyze{
		Volume: &pb.Volume{
			Lloc: func() *int32 { v := int32(0); return &v }(),
		},
	}
	
	result := DecorateMaintainabilityIndex(90, analyze)
	if result != "-" {
		t.Errorf("expected '-', got %s", result)
	}
}

func TestRound(t *testing.T) {
	if Round(3.7) != 4 {
		t.Errorf("expected 4, got %d", Round(3.7))
	}
	if Round(3.2) != 3 {
		t.Errorf("expected 3, got %d", Round(3.2))
	}
}

func TestToFixed(t *testing.T) {
	result := ToFixed(3.14159, 2)
	if result != 3.14 {
		t.Errorf("expected 3.14, got %f", result)
	}
}
