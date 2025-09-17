package ruleset

import (
	"testing"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestMaxMethodsPerClassRule(t *testing.T) {
	threshold := 3
	rule := NewMaxMethodsPerClassRule(&threshold)

	// Test class with too many methods
	file := &pb.File{
		Path: "test.go",
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Short: "TestClass"},
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{
							{Name: &pb.Name{Short: "method1"}},
							{Name: &pb.Name{Short: "method2"}},
							{Name: &pb.Name{Short: "method3"}},
							{Name: &pb.Name{Short: "method4"}}, // Exceeds threshold
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
		t.Error("Expected error to be called for class with too many methods")
	}
	if successCalled {
		t.Error("Expected success not to be called for class with too many methods")
	}

	// Test class within threshold
	file.Stmts.StmtClass[0].Stmts.StmtFunction = file.Stmts.StmtClass[0].Stmts.StmtFunction[:3]
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
