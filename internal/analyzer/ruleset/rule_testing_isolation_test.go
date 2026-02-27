package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
)

func TestIsolationRule_NilThreshold(t *testing.T) {
	rule := NewIsolationRule(nil)
	var errors []issue.RequirementError

	rule.CheckProject(ProjectContext{GlobalIsolationScore: 30}, func(e issue.RequirementError) {
		errors = append(errors, e)
	}, func(s string) {})

	if len(errors) != 0 {
		t.Errorf("expected no errors with nil threshold, got %d", len(errors))
	}
}

func TestIsolationRule_Violation(t *testing.T) {
	min := 50
	rule := NewIsolationRule(&min)
	var errors []issue.RequirementError

	rule.CheckProject(ProjectContext{GlobalIsolationScore: 35.0}, func(e issue.RequirementError) {
		errors = append(errors, e)
	}, func(s string) {})

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if errors[0].Severity != issue.SeverityMedium {
		t.Errorf("expected medium severity, got %s", errors[0].Severity)
	}
}

func TestIsolationRule_Success(t *testing.T) {
	min := 50
	rule := NewIsolationRule(&min)
	var successes []string

	rule.CheckProject(ProjectContext{GlobalIsolationScore: 80.0}, func(e issue.RequirementError) {
		t.Errorf("unexpected error: %s", e.Message)
	}, func(s string) {
		successes = append(successes, s)
	})

	if len(successes) != 1 {
		t.Fatalf("expected 1 success, got %d", len(successes))
	}
}
