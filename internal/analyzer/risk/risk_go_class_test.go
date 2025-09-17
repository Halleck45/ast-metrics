package risk

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestTooManyGoClassesDetector_Name(t *testing.T) {
	detector := &TooManyGoClassesDetector{}
	if detector.Name() != "risk_go_class" {
		t.Errorf("expected 'risk_go_class', got %s", detector.Name())
	}
}

func TestTooManyGoClassesDetector_Detect_NonGoFile(t *testing.T) {
	detector := &TooManyGoClassesDetector{}

	file := &pb.File{
		ProgrammingLanguage: "Python",
		Stmts: &pb.Stmts{
			StmtClass: make([]*pb.StmtClass, 5),
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 0 {
		t.Errorf("expected 0 risks for non-Go file, got %d", len(risks))
	}
}

func TestTooManyGoClassesDetector_Detect_FewClasses(t *testing.T) {
	detector := &TooManyGoClassesDetector{}

	file := &pb.File{
		ProgrammingLanguage: "Golang",
		Stmts: &pb.Stmts{
			StmtClass: make([]*pb.StmtClass, 2),
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 0 {
		t.Errorf("expected 0 risks for few classes, got %d", len(risks))
	}
}

func TestTooManyGoClassesDetector_Detect_ManyClasses(t *testing.T) {
	detector := &TooManyGoClassesDetector{}

	file := &pb.File{
		ProgrammingLanguage: "Golang",
		Stmts: &pb.Stmts{
			StmtClass: make([]*pb.StmtClass, 5),
		},
	}

	risks := detector.Detect(file)
	if len(risks) != 1 {
		t.Fatalf("expected 1 risk, got %d", len(risks))
	}

	risk := risks[0]
	if risk.Title != "Too many types in a single Go file" {
		t.Errorf("expected title 'Too many types in a single Go file', got %s", risk.Title)
	}

	expectedSeverity := 0.6 + float64(5-3)*0.05
	if risk.Severity != expectedSeverity {
		t.Errorf("expected severity %f, got %f", expectedSeverity, risk.Severity)
	}
}
