package ruleset

import (
	"testing"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestMaxParametersPerMethodRule(t *testing.T) {
	threshold := 3
	rule := NewMaxParametersPerMethodRule(&threshold)

	// Test function with too many parameters
	file := &pb.File{
		Path: "test.go",
		Stmts: &pb.Stmts{
			StmtFunction: []*pb.StmtFunction{
				{
					Name: &pb.Name{Short: "testFunc"},
					Parameters: []*pb.StmtParameter{
						{Name: "param1"},
						{Name: "param2"},
						{Name: "param3"},
						{Name: "param4"}, // Exceeds threshold
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
		t.Error("Expected error to be called for function with too many parameters")
	}
	if successCalled {
		t.Error("Expected success not to be called for function with too many parameters")
	}

	// Test function within threshold
	file.Stmts.StmtFunction[0].Parameters = file.Stmts.StmtFunction[0].Parameters[:3]
	errorCalled = false
	successCalled = false
	
	rule.CheckFile(file, func(err issue.RequirementError) {
		errorCalled = true
	}, func(msg string) {
		successCalled = true
	})

	if errorCalled {
		t.Error("Expected error not to be called for function within threshold")
	}
	if !successCalled {
		t.Error("Expected success to be called for function within threshold")
	}
}
