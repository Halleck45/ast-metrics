package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
)

type isolationRule struct {
	min *int
}

func NewIsolationRule(min *int) ProjectRule {
	return &isolationRule{min: min}
}

func (r *isolationRule) Name() string {
	return "min_isolation_score"
}

func (r *isolationRule) Description() string {
	return "Checks that the global test isolation score meets a minimum threshold"
}

func (r *isolationRule) CheckProject(ctx ProjectContext, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.min == nil {
		return
	}

	value := ctx.GlobalIsolationScore
	threshold := float64(*r.min)
	if value < threshold {
		addError(issue.RequirementError{
			Severity: issue.SeverityMedium,
			Message:  fmt.Sprintf("Global test isolation score too low: got %.1f (min: %d)", value, *r.min),
			Code:     r.Name(),
		})
		return
	}

	addSuccess("Global test isolation score OK")
}
