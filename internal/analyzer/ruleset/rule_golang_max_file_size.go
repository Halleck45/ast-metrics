package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// Rule: Max file LOC

type ruleMaxFileLoc struct{ max int }

func (r *ruleMaxFileLoc) Name() string {
	return "max_file_size"
}
func (r *ruleMaxFileLoc) Description() string {
	return "Limit file size (LOC)"
}
func (r *ruleMaxFileLoc) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if file == nil || file.LinesOfCode == nil {
		return
	}
	loc := int(file.LinesOfCode.LinesOfCode)
	if r.max > 0 && loc > r.max {
		addError(issue.RequirementError{
			Severity: issue.SeverityMedium,
			Code:     r.Name(),
			Message:  fmt.Sprintf("File too large: %d LOC > %d in %s", loc, r.max, file.Path),
		})
		return
	}
	addSuccess(fmt.Sprintf("[%s] LOC %d ≤ %d in %s", r.Name(), loc, r.max, file.Path))
}
