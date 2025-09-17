package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

type maxSwitchCasesRule struct {
	threshold int
}

func NewMaxSwitchCasesRule(threshold *int) Rule {
	if threshold == nil {
		return &maxSwitchCasesRule{threshold: 0}
	}
	return &maxSwitchCasesRule{threshold: *threshold}
}

func (r *maxSwitchCasesRule) Name() string {
	return "max_switch_cases"
}

func (r *maxSwitchCasesRule) Description() string {
	return "Maximum number of cases in switch statements"
}

func (r *maxSwitchCasesRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.threshold == 0 || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil {
		return
	}

	// Use cyclomatic complexity as a proxy for switch complexity
	if file.Stmts.Analyze.Complexity.Cyclomatic != nil && int(*file.Stmts.Analyze.Complexity.Cyclomatic) > r.threshold {
		addError(issue.RequirementError{
			Severity: issue.SeverityMedium,
			Message:  fmt.Sprintf("File has cyclomatic complexity of %d, maximum allowed is %d", int(*file.Stmts.Analyze.Complexity.Cyclomatic), r.threshold),
			Code:     r.Name(),
		})
		return
	}

	addSuccess("Max switch cases OK")
}
