package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
)

type traceabilityRule struct {
	min *int
}

func NewTraceabilityRule(min *int) ProjectRule {
	return &traceabilityRule{min: min}
}

func (r *traceabilityRule) Name() string {
	return "min_traceability"
}

func (r *traceabilityRule) Description() string {
	return "Checks that the percentage of production classes covered by tests meets a minimum threshold"
}

func (r *traceabilityRule) CheckProject(ctx ProjectContext, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.min == nil {
		return
	}

	value := ctx.TraceabilityPct
	threshold := float64(*r.min)
	if value < threshold {
		addError(issue.RequirementError{
			Severity: issue.SeverityHigh,
			Message:  fmt.Sprintf("Test traceability too low: got %.1f%% (min: %d%%)", value, *r.min),
			Code:     r.Name(),
		})
		return
	}

	addSuccess("Test traceability OK")
}
