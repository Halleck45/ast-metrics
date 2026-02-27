package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
)

type orphanCriticalRule struct {
	max *float64
}

func NewOrphanCriticalRule(max *float64) ProjectRule {
	return &orphanCriticalRule{max: max}
}

func (r *orphanCriticalRule) Name() string {
	return "max_orphan_weight"
}

func (r *orphanCriticalRule) Description() string {
	return "Checks that no untested production class exceeds a maximum importance weight"
}

func (r *orphanCriticalRule) CheckProject(ctx ProjectContext, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.max == nil {
		return
	}

	found := false
	for _, oc := range ctx.OrphanClasses {
		if oc.Weight > *r.max {
			found = true
			addError(issue.RequirementError{
				Severity: issue.SeverityMedium,
				Message:  fmt.Sprintf("Critical orphan class: %s (weight: %.1f, max: %.1f)", oc.ClassName, oc.Weight, *r.max),
				Code:     r.Name(),
			})
		}
	}

	if !found {
		addSuccess("No critical orphan classes detected")
	}
}
