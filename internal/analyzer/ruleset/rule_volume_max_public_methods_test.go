package ruleset

import (
	"testing"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

func TestMaxPublicMethodsRule(t *testing.T) {
	threshold := 3
	rule := NewMaxPublicMethodsRule(&threshold)

	// Test class with too many public methods
	file := &pb.File{
		Path: "test.go",
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Short: "TestClass"},
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{
							{Name: &pb.Name{Short: "PublicMethod1"}}, // Public (uppercase)
							{Name: &pb.Name{Short: "PublicMethod2"}}, // Public
							{Name: &pb.Name{Short: "PublicMethod3"}}, // Public
							{Name: &pb.Name{Short: "PublicMethod4"}}, // Public - exceeds threshold
							{Name: &pb.Name{Short: "_privateMethod"}}, // Private (starts with _)
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
		t.Error("Expected error to be called for class with too many public methods")
	}
	if successCalled {
		t.Error("Expected success not to be called for class with too many public methods")
	}

	// Test class within threshold (remove one public method)
	file.Stmts.StmtClass[0].Stmts.StmtFunction = []*pb.StmtFunction{
		{Name: &pb.Name{Short: "PublicMethod1"}}, // Public
		{Name: &pb.Name{Short: "PublicMethod2"}}, // Public
		{Name: &pb.Name{Short: "PublicMethod3"}}, // Public - exactly at threshold
		{Name: &pb.Name{Short: "_privateMethod"}}, // Private (doesn't count)
	}
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
