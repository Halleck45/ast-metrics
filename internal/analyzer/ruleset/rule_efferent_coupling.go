package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type efferentCouplingRule struct {
	max *int
}

func NewEfferentCouplingRule(max *int) Rule {
	return &efferentCouplingRule{max: max}
}

func (r *efferentCouplingRule) Name() string { return "efferent_coupling" }

func (r *efferentCouplingRule) Description() string {
	return "Checks the efferent coupling of files/classes"
}

func (r *efferentCouplingRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.max == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Coupling == nil {
		return
	}

	value := int(file.Stmts.Analyze.Coupling.Efferent)
	if r.max != nil && *r.max > 0 && value > *r.max {
		addError(issue.RequirementError{
			Severity: issue.SeverityUnknown,
			Code:     r.Name(),
			Message:  fmt.Sprintf("Efferent coupling too high: got %d (max: %d)", value, *r.max),
		})
		return
	}

	addSuccess("Efferent coupling OK")
}
