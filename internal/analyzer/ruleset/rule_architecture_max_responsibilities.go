package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

type maxResponsibilitiesRule struct {
	threshold int
}

func NewMaxResponsibilitiesRule(threshold *int) Rule {
	if threshold == nil {
		return &maxResponsibilitiesRule{threshold: 0}
	}
	return &maxResponsibilitiesRule{threshold: *threshold}
}

func (r *maxResponsibilitiesRule) Name() string {
	return "max_responsibilities"
}

func (r *maxResponsibilitiesRule) Description() string {
	return "Maximum number of responsibilities (LCOM) per class"
}

func (r *maxResponsibilitiesRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.threshold == 0 || file.Stmts == nil || file.Stmts.StmtClass == nil {
		return
	}

	for _, class := range file.Stmts.StmtClass {
		if class.Stmts != nil && class.Stmts.Analyze != nil && class.Stmts.Analyze.ClassCohesion != nil && class.Stmts.Analyze.ClassCohesion.Lcom4 != nil && int(*class.Stmts.Analyze.ClassCohesion.Lcom4) > r.threshold {
			addError(issue.RequirementError{
				Severity: issue.SeverityMedium,
				Message:  fmt.Sprintf("Class has LCOM4 of %d, maximum allowed is %d", int(*class.Stmts.Analyze.ClassCohesion.Lcom4), r.threshold),
				Code:     r.Name(),
			})
			return
		}
	}

	addSuccess("Max responsibilities (LCOM4) OK")
}
