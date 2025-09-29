package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

func TestCyclomaticRule_Name(t *testing.T) {
	rule := NewCyclomaticRule(nil)
	if rule.Name() != "cyclomatic_complexity" {
		t.Errorf("expected 'cyclomatic_complexity', got %s", rule.Name())
	}
}

func TestCyclomaticRule_Description(t *testing.T) {
	rule := NewCyclomaticRule(nil)
	expected := "Checks the cyclomatic complexity of functions"
	if rule.Description() != expected {
		t.Errorf("expected '%s', got %s", expected, rule.Description())
	}
}

func TestCyclomaticRule_CheckFile_NilMax(t *testing.T) {
	rule := NewCyclomaticRule(nil)
	file := &pb.File{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{
					Cyclomatic: func() *int32 { v := int32(15); return &v }(),
				},
			},
		},
	}

	errors := []issue.RequirementError{}
	successes := []string{}

	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(s string) { successes = append(successes, s) })

	if len(errors) != 0 {
		t.Errorf("expected no errors with nil max, got %d", len(errors))
	}
	if len(successes) != 0 {
		t.Errorf("expected no successes with nil max, got %d", len(successes))
	}
}

func TestCyclomaticRule_CheckFile_Violation(t *testing.T) {
	max := 10
	rule := NewCyclomaticRule(&max)
	file := &pb.File{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{
					Cyclomatic: func() *int32 { v := int32(15); return &v }(),
				},
			},
		},
	}

	errors := []issue.RequirementError{}
	successes := []string{}

	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(s string) { successes = append(successes, s) })

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	err := errors[0]
	if err.Severity != issue.SeverityMedium {
		t.Errorf("expected medium severity, got %s", err.Severity)
	}
	if err.Code != "cyclomatic_complexity" {
		t.Errorf("expected 'cyclomatic_complexity' code, got %s", err.Code)
	}
	if err.Message != "Cyclomatic complexity too high: got 15 (max: 10)" {
		t.Errorf("unexpected error message: %s", err.Message)
	}
}

func TestCyclomaticRule_CheckFile_Success(t *testing.T) {
	max := 10
	rule := NewCyclomaticRule(&max)
	file := &pb.File{
		Stmts: &pb.Stmts{
			Analyze: &pb.Analyze{
				Complexity: &pb.Complexity{
					Cyclomatic: func() *int32 { v := int32(5); return &v }(),
				},
			},
		},
	}

	errors := []issue.RequirementError{}
	successes := []string{}

	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(s string) { successes = append(successes, s) })

	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}
	if len(successes) != 1 {
		t.Fatalf("expected 1 success, got %d", len(successes))
	}
	if successes[0] != "Cyclomatic complexity OK" {
		t.Errorf("unexpected success message: %s", successes[0])
	}
}

func TestCyclomaticRule_CheckFile_EmptyFile(t *testing.T) {
	max := 10
	rule := NewCyclomaticRule(&max)
	file := &pb.File{} // Empty file with no Stmts

	errors := []issue.RequirementError{}
	successes := []string{}

	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(s string) { successes = append(successes, s) })

	if len(errors) != 0 || len(successes) != 0 {
		t.Error("expected no errors or successes with empty file")
	}
}
