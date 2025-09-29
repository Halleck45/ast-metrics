package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

type maxMethodsPerClassRule struct {
	threshold int
}

func NewMaxMethodsPerClassRule(threshold *int) Rule {
	if threshold == nil {
		return &maxMethodsPerClassRule{threshold: 0}
	}
	return &maxMethodsPerClassRule{threshold: *threshold}
}

func (r *maxMethodsPerClassRule) Name() string {
	return "max_methods_per_class"
}

func (r *maxMethodsPerClassRule) Description() string {
	return "Maximum number of methods per class"
}

func (r *maxMethodsPerClassRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.threshold == 0 || file.Stmts == nil || file.Stmts.StmtClass == nil {
		return
	}

	for _, class := range file.Stmts.StmtClass {
		if class.Stmts != nil && class.Stmts.StmtFunction != nil {
			methodCount := len(class.Stmts.StmtFunction)
			if methodCount > r.threshold {
				addError(issue.RequirementError{
					Severity: issue.SeverityMedium,
					Message:  fmt.Sprintf("Class has %d methods, maximum allowed is %d", methodCount, r.threshold),
					Code:     r.Name(),
				})
				return
			}
		}
	}

	addSuccess("Max methods per class OK")
}
