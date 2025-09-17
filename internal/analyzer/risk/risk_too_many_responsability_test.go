package risk

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestTooManyResponsibilityDetector_Name(t *testing.T) {
	detector := &TooManyResponsibilityDetector{}
	if detector.Name() != "risk_too_many_responsability" {
		t.Errorf("expected 'risk_too_many_responsability', got %s", detector.Name())
	}
}

func TestTooManyResponsibilityDetector_Detect_TooManyMethods(t *testing.T) {
	detector := &TooManyResponsibilityDetector{}
	
	// Create 25 methods
	methods := make([]*pb.StmtFunction, 25)
	for i := range methods {
		methods[i] = &pb.StmtFunction{}
	}
	
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name:  &pb.Name{Qualified: "TestClass"},
					Stmts: &pb.Stmts{StmtFunction: methods},
				},
			},
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 1 {
		t.Fatalf("expected 1 risk, got %d", len(risks))
	}

	if risks[0].Severity != 0.5 {
		t.Errorf("expected severity 0.5, got %f", risks[0].Severity)
	}
}

func TestTooManyResponsibilityDetector_Detect_HighLCOM4(t *testing.T) {
	detector := &TooManyResponsibilityDetector{}
	lcom4 := int32(3)
	
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Qualified: "TestClass"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							ClassCohesion: &pb.ClassCohesion{Lcom4: &lcom4},
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

	if risks[0].Severity != 0.7 {
		t.Errorf("expected severity 0.7 for high LCOM4, got %f", risks[0].Severity)
	}
}

func TestTooManyResponsibilityDetector_Detect_NoRisk(t *testing.T) {
	detector := &TooManyResponsibilityDetector{}
	
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name:  &pb.Name{Qualified: "TestClass"},
					Stmts: &pb.Stmts{StmtFunction: make([]*pb.StmtFunction, 5)},
				},
			},
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 0 {
		t.Errorf("expected 0 risks, got %d", len(risks))
	}
}
