package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type afferentCouplingRule struct {
	max *int
}

func NewAfferentCouplingRule(max *int) Rule {
	return &afferentCouplingRule{max: max}
}

func (r *afferentCouplingRule) Name() string { return "afferent_coupling" }

func (r *afferentCouplingRule) Description() string {
	return "Checks the afferent coupling of files/classes"
}

func (r *afferentCouplingRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.max == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Coupling == nil {
		return
	}

	value := int(file.Stmts.Analyze.Coupling.Afferent)
	if r.max != nil && *r.max > 0 && value > *r.max {
		addError(issue.RequirementError{
			Severity: issue.SeverityUnknown,
			Code:     r.Name(),
			Message:  fmt.Sprintf("Afferent coupling too high in file %s: got %d (max: %d)", file.Path, value, *r.max),
		})
		return
	}
	addSuccess(fmt.Sprintf("Afferent coupling OK in file %s", file.Path))
}
