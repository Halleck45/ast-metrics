package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type maintainabilityRule struct {
	min *int
}

func NewMaintainabilityRule(min *int) Rule {
	return &maintainabilityRule{min: min}
}

func (r *maintainabilityRule) Name() string {
	return "maintainability"
}

func (r *maintainabilityRule) Description() string {
	return "Checks the maintainability of the code"
}

func (r *maintainabilityRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.min == nil || file.Stmts == nil {
		return
	}
	// apply to classes
	hasAny := false
	for _, class := range file.Stmts.StmtClass {
		if class.Stmts == nil || class.Stmts.Analyze == nil || class.Stmts.Analyze.Maintainability == nil || class.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil {
			continue
		}
		hasAny = true
		value := int(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)
		if r.min != nil && value < *r.min {
			addError(issue.RequirementError{
				Severity: issue.SeverityHigh,
				Message:  fmt.Sprintf("Maintainability too low in file %s: got %d (min: %d)", file.Path, value, *r.min),
				Code:     r.Name(),
			})
			return
		}
	}
	if hasAny {
		addSuccess(fmt.Sprintf("Maintainability OK in file %s", file.Path))
	}
}
