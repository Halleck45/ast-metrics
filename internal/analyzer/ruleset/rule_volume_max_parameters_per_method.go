package ruleset

import (
	"fmt"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type maxParametersPerMethodRule struct {
	threshold int
}

func NewMaxParametersPerMethodRule(threshold *int) Rule {
	if threshold == nil {
		return &maxParametersPerMethodRule{threshold: 0}
	}
	return &maxParametersPerMethodRule{threshold: *threshold}
}

func (r *maxParametersPerMethodRule) Name() string {
	return "max_parameters_per_method"
}

func (r *maxParametersPerMethodRule) Description() string {
	return "Maximum number of parameters per method"
}

func (r *maxParametersPerMethodRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.threshold == 0 || file.Stmts == nil || file.Stmts.StmtFunction == nil {
		return
	}

	for _, function := range file.Stmts.StmtFunction {
		if function.Parameters != nil && len(function.Parameters) > r.threshold {
			addError(issue.RequirementError{
				Severity: issue.SeverityMedium,
				Message:  fmt.Sprintf("Method has %d parameters, maximum allowed is %d", len(function.Parameters), r.threshold),
				Code:     r.Name(),
			})
			return
		}
	}

	// Check methods in classes
	if file.Stmts.StmtClass != nil {
		for _, class := range file.Stmts.StmtClass {
			if class.Stmts != nil && class.Stmts.StmtFunction != nil {
				for _, method := range class.Stmts.StmtFunction {
					if method.Parameters != nil && len(method.Parameters) > r.threshold {
						addError(issue.RequirementError{
							Severity: issue.SeverityMedium,
							Message:  fmt.Sprintf("Method has %d parameters, maximum allowed is %d", len(method.Parameters), r.threshold),
							Code:     r.Name(),
						})
						return
					}
				}
			}
		}
	}

	addSuccess(fmt.Sprintf("Max parameters per method OK in file %s", file.Path))
}
