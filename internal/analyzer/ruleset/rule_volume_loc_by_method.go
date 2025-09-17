package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

type locByMethodRule struct {
	max *int
}

func NewLocByMethodRule(max *int) Rule {
	return &locByMethodRule{max: max}
}

func (r *locByMethodRule) Name() string { return "max_loc_by_method" }

func (r *locByMethodRule) Description() string {
	return "Checks the lines of code by method/function"
}

func (r *locByMethodRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.max == nil || file.Stmts == nil || file.Stmts.StmtFunction == nil {
		return
	}

	ok := true
	for _, f := range file.Stmts.StmtFunction {
		if f == nil || f.LinesOfCode == nil || f.LinesOfCode.LinesOfCode == 0 {
			continue
		}
		value := int(f.LinesOfCode.LinesOfCode)
		if r.max != nil && *r.max > 0 && value > *r.max {
			addError(issue.RequirementError{
				Severity: issue.SeverityMedium,
				Message:  fmt.Sprintf("LOC too high in method %s(): got %d (max: %d)", f.Name.Short, value, *r.max),
				Code:     r.Name(),
			})
			ok = false
			continue
		}
	}

	if ok {
		addSuccess("LOC by method OK")
	}
}
