package ruleset

import (
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/analyzer/issue"
	pb "github.com/ast-metrics/ast-metrics/pb"
)

func TestLocByMethodRule_ReportsFunctionLine(t *testing.T) {
	max := 10
	rule := NewLocByMethodRule(&max)

	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtFunction: []*pb.StmtFunction{
				{
					Name:        &pb.Name{Short: "TooLong"},
					LinesOfCode: &pb.LinesOfCode{LinesOfCode: 50},
					Location:    &pb.StmtLocationInFile{StartLine: 123},
				},
			},
		},
	}

	var errors []issue.RequirementError
	rule.CheckFile(file,
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(string) {})

	if len(errors) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(errors))
	}
	if errors[0].Line != 123 {
		t.Errorf("expected violation anchored to line 123, got %d", errors[0].Line)
	}
}

func TestMaxResponsibilitiesRule_ReportsClassLine(t *testing.T) {
	threshold := 5
	rule := NewMaxResponsibilitiesRule(&threshold)

	lcom := int32(20)
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Location: &pb.StmtLocationInFile{StartLine: 7},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							ClassCohesion: &pb.ClassCohesion{Lcom4: &lcom},
						},
					},
				},
			},
		},
	}

	var errors []issue.RequirementError
	rule.CheckFile(file,
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(string) {})

	if len(errors) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(errors))
	}
	if errors[0].Line != 7 {
		t.Errorf("expected violation anchored to line 7, got %d", errors[0].Line)
	}
}
