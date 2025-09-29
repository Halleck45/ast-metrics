package ruleset

import (
	"testing"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

func TestMaxSwitchCasesRule(t *testing.T) {
	threshold := 5
	rule := NewMaxSwitchCasesRule(&threshold)

	// Test file with high cyclomatic complexity
	cyclomatic := int32(6)
	file := &pb.File{
		Path: "test.go",
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{
					Cyclomatic: &cyclomatic,
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
		t.Error("Expected error to be called for file with high cyclomatic complexity")
	}
	if successCalled {
		t.Error("Expected success not to be called for file with high cyclomatic complexity")
	}

	// Test file within threshold
	cyclomatic = 5
	errorCalled = false
	successCalled = false
	
	rule.CheckFile(file, func(err issue.RequirementError) {
		errorCalled = true
	}, func(msg string) {
		successCalled = true
	})

	if errorCalled {
		t.Error("Expected error not to be called for file within threshold")
	}
	if !successCalled {
		t.Error("Expected success to be called for file within threshold")
	}
}
