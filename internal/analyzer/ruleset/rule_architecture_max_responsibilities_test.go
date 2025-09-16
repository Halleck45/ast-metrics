package ruleset

import (
	"testing"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestMaxResponsibilitiesRule(t *testing.T) {
	threshold := 10
	rule := NewMaxResponsibilitiesRule(&threshold)

	// Test class with high LCOM4
	lcom4 := int32(15)
	file := &pb.File{
		Path: "test.go",
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Short: "TestClass"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							ClassCohesion: &pb.ClassCohesion{
								Lcom4: &lcom4,
							},
						},
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
		t.Error("Expected error to be called for class with high LCOM4")
	}
	if successCalled {
		t.Error("Expected success not to be called for class with high LCOM4")
	}

	// Test class within threshold
	lcom4 = 8
	errorCalled = false
	successCalled = false
	
	rule.CheckFile(file, func(err issue.RequirementError) {
		errorCalled = true
	}, func(msg string) {
		successCalled = true
	})

	if errorCalled {
		t.Error("Expected error not to be called for class within threshold")
	}
	if !successCalled {
		t.Error("Expected success to be called for class within threshold")
	}
}
