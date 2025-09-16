package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type llocByMethodRule struct {
	max *int
}

func NewLlocByMethodRule(max *int) Rule {
	return &llocByMethodRule{max: max}
}

func (r *llocByMethodRule) Name() string { return "lloc_by_method" }

func (r *llocByMethodRule) Description() string {
	return "Checks the logical lines of code by method/function"
}

func (r *llocByMethodRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.max == nil || file.Stmts == nil || file.Stmts.StmtFunction == nil {
		return
	}

	ok := true
	for _, f := range file.Stmts.StmtFunction {
		if f == nil || f.LinesOfCode == nil || f.LinesOfCode.LogicalLinesOfCode == 0 {
			continue
		}
		value := int(f.LinesOfCode.LogicalLinesOfCode)
		if r.max != nil && *r.max > 0 && value > *r.max {
			addError(issue.RequirementError{
				Severity: issue.SeverityMedium,
				Message:  fmt.Sprintf("LLOC too high in method %s(): got %d (max: %d)", f.Name.Short, value, *r.max),
				Code:     r.Name(),
			})
			ok = false
			continue
		}
	}
	if ok {
		addSuccess("LLOC by method OK")
	}
}
