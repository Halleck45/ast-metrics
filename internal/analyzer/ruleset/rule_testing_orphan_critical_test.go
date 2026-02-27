package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
)

func TestOrphanCriticalRule_NilThreshold(t *testing.T) {
	rule := NewOrphanCriticalRule(nil)
	var errors []issue.RequirementError

	ctx := ProjectContext{
		OrphanClasses: []OrphanClassInfo{{ClassName: "Foo", Weight: 100}},
	}
	rule.CheckProject(ctx, func(e issue.RequirementError) {
		errors = append(errors, e)
	}, func(s string) {})

	if len(errors) != 0 {
		t.Errorf("expected no errors with nil threshold, got %d", len(errors))
	}
}

func TestOrphanCriticalRule_Violation(t *testing.T) {
	max := 20.0
	rule := NewOrphanCriticalRule(&max)
	var errors []issue.RequirementError

	ctx := ProjectContext{
		OrphanClasses: []OrphanClassInfo{{ClassName: "BigService", FilePath: "service.go", Weight: 50.0}},
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

func TestOrphanCriticalRule_Success(t *testing.T) {
	max := 20.0
	rule := NewOrphanCriticalRule(&max)
	var successes []string

	ctx := ProjectContext{
		OrphanClasses: []OrphanClassInfo{{ClassName: "SmallHelper", Weight: 5.0}},
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
