package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
)

type godTestRule struct {
	max *int
}

func NewGodTestRule(max *int) ProjectRule {
	return &godTestRule{max: max}
}

func (r *godTestRule) Name() string {
	return "max_god_test_fan_out"
}

func (r *godTestRule) Description() string {
	return "Checks that no test file has excessive fan-out (god test)"
}

func (r *godTestRule) CheckProject(ctx ProjectContext, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.max == nil {
		return
	}

	found := false
	for _, gt := range ctx.GodTests {
		if gt.FanOut > *r.max {
			found = true
			addError(issue.RequirementError{
				Severity: issue.SeverityMedium,
				Message:  fmt.Sprintf("God test detected: %s has fan-out %d (max: %d)", gt.FilePath, gt.FanOut, *r.max),
				Code:     r.Name(),
			})
		}
	}

	if !found {
		addSuccess("No god tests detected")
	}
}
