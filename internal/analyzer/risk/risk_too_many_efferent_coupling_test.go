package risk

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestTooManyEfferentCouplingDetector_Name(t *testing.T) {
	detector := &TooManyEfferentCouplingDetector{}
	if detector.Name() != "risk_too_many_efferent_coupling" {
		t.Errorf("expected 'risk_too_many_efferent_coupling', got %s", detector.Name())
	}
}

func TestTooManyEfferentCouplingDetector_Detect_NilFile(t *testing.T) {
	detector := &TooManyEfferentCouplingDetector{}
	risks := detector.Detect(nil)
	if len(risks) != 0 {
		t.Errorf("expected 0 risks for nil file, got %d", len(risks))
	}
}

func TestTooManyEfferentCouplingDetector_Detect_HighCoupling(t *testing.T) {
	detector := &TooManyEfferentCouplingDetector{}
	
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Qualified: "TestClass"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Coupling: &pb.Coupling{Efferent: 25},
						},
					},
				},
			},
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 1 {
		t.Fatalf("expected 1 risk, got %d", len(risks))
	}

	risk := risks[0]
	if risk.ID != "risk_too_many_efferent_coupling" {
		t.Errorf("expected ID 'risk_too_many_efferent_coupling', got %s", risk.ID)
	}
	if risk.Severity != 0.6 {
		t.Errorf("expected severity 0.6, got %f", risk.Severity)
	}
	if risk.Title != "Excessive efferent coupling" {
		t.Errorf("expected title 'Excessive efferent coupling', got %s", risk.Title)
	}
}

func TestTooManyEfferentCouplingDetector_Detect_VeryHighCoupling(t *testing.T) {
	detector := &TooManyEfferentCouplingDetector{}
	
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Qualified: "TestClass"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Coupling: &pb.Coupling{Efferent: 45},
						},
					},
				},
			},
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 1 {
		t.Fatalf("expected 1 risk, got %d", len(risks))
	}

	if risks[0].Severity != 0.85 {
		t.Errorf("expected severity 0.85 for very high coupling, got %f", risks[0].Severity)
	}
}

func TestTooManyEfferentCouplingDetector_Detect_LowCoupling(t *testing.T) {
	detector := &TooManyEfferentCouplingDetector{}
	
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Qualified: "TestClass"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Coupling: &pb.Coupling{Efferent: 10},
						},
					},
				},
			},
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 0 {
		t.Errorf("expected 0 risks for low coupling, got %d", len(risks))
	}
}
