package ruleset

import (
	"testing"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

func TestNoGodClassRule(t *testing.T) {
	enabled := true
	rule := NewNoGodClassRule(&enabled)

	// Test god class
	loc := int32(600)
	lcom4 := int32(15)
	file := &pb.File{
		Path: "test.go",
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Short: "GodClass"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Volume: &pb.Volume{
								Loc: &loc,
							},
							ClassCohesion: &pb.ClassCohesion{
								Lcom4: &lcom4,
							},
						},
						StmtFunction: make([]*pb.StmtFunction, 25), // > 20 methods
					},
				},
			},
		},
	}

	errorCalled := false
	successCalled := false
	
	rule.CheckFile(file, func(err issue.RequirementError) {
		errorCalled = true
	}, func(msg string) {
		successCalled = true
	})

	if !errorCalled {
		t.Error("Expected error to be called for god class")
	}
	if successCalled {
		t.Error("Expected success not to be called for god class")
	}

	// Test normal class
	loc = 200
	lcom4 = 5
	file.Stmts.StmtClass[0].Stmts.StmtFunction = make([]*pb.StmtFunction, 10)
	errorCalled = false
	successCalled = false
	
	rule.CheckFile(file, func(err issue.RequirementError) {
		errorCalled = true
	}, func(msg string) {
		successCalled = true
	})

	if errorCalled {
		t.Error("Expected error not to be called for normal class")
	}
	if !successCalled {
		t.Error("Expected success to be called for normal class")
	}
}
