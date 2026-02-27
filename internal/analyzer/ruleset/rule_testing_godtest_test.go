package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
)

func TestGodTestRule_NilThreshold(t *testing.T) {
	rule := NewGodTestRule(nil)
	var errors []issue.RequirementError

	ctx := ProjectContext{
		GodTests: []GodTestInfo{{FilePath: "test.go", FanOut: 10}},
	}
	rule.CheckProject(ctx, func(e issue.RequirementError) {
		errors = append(errors, e)
	}, func(s string) {})

	if len(errors) != 0 {
		t.Errorf("expected no errors with nil threshold, got %d", len(errors))
	}
}

func TestGodTestRule_Violation(t *testing.T) {
	max := 5
	rule := NewGodTestRule(&max)
	var errors []issue.RequirementError

	ctx := ProjectContext{
		GodTests: []GodTestInfo{{FilePath: "big_test.go", FanOut: 8}},
	}
	rule.CheckProject(ctx, func(e issue.RequirementError) {
		errors = append(errors, e)
	}, func(s string) {})

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if errors[0].Severity != issue.SeverityMedium {
		t.Errorf("expected medium severity, got %s", errors[0].Severity)
	}
}

func TestGodTestRule_Success(t *testing.T) {
	max := 5
	rule := NewGodTestRule(&max)
	var successes []string

	ctx := ProjectContext{
		GodTests: []GodTestInfo{{FilePath: "small_test.go", FanOut: 3}},
	}
	rule.CheckProject(ctx, func(e issue.RequirementError) {
		t.Errorf("unexpected error: %s", e.Message)
	}, func(s string) {
		successes = append(successes, s)
	})

	if len(successes) != 1 {
		t.Fatalf("expected 1 success, got %d", len(successes))
	}
}

func TestGodTestRule_MultipleViolations(t *testing.T) {
	max := 5
	rule := NewGodTestRule(&max)
	var errors []issue.RequirementError

	ctx := ProjectContext{
		GodTests: []GodTestInfo{
			{FilePath: "a_test.go", FanOut: 10},
			{FilePath: "b_test.go", FanOut: 7},
			{FilePath: "c_test.go", FanOut: 3}, // under threshold
		},
	}
	rule.CheckProject(ctx, func(e issue.RequirementError) {
		errors = append(errors, e)
	}, func(s string) {})

	if len(errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errors))
	}
}
