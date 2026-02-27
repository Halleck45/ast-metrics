package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
)

func TestTraceabilityRule_NilThreshold(t *testing.T) {
	rule := NewTraceabilityRule(nil)
	var errors []issue.RequirementError
	var successes []string

	rule.CheckProject(ProjectContext{TraceabilityPct: 30}, func(e issue.RequirementError) {
		errors = append(errors, e)
	}, func(s string) {
		successes = append(successes, s)
	})

	if len(errors) != 0 {
		t.Errorf("expected no errors with nil threshold, got %d", len(errors))
	}
	if len(successes) != 0 {
		t.Errorf("expected no successes with nil threshold, got %d", len(successes))
	}
}

func TestTraceabilityRule_Violation(t *testing.T) {
	min := 60
	rule := NewTraceabilityRule(&min)
	var errors []issue.RequirementError

	rule.CheckProject(ProjectContext{TraceabilityPct: 40.5}, func(e issue.RequirementError) {
		errors = append(errors, e)
	}, func(s string) {})

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if errors[0].Severity != issue.SeverityHigh {
		t.Errorf("expected high severity, got %s", errors[0].Severity)
	}
}

func TestTraceabilityRule_Success(t *testing.T) {
	min := 60
	rule := NewTraceabilityRule(&min)
	var successes []string

	rule.CheckProject(ProjectContext{TraceabilityPct: 75.0}, func(e issue.RequirementError) {
		t.Errorf("unexpected error: %s", e.Message)
	}, func(s string) {
		successes = append(successes, s)
	})

	if len(successes) != 1 {
		t.Fatalf("expected 1 success, got %d", len(successes))
	}
}
