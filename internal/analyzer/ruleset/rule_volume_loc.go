package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type locRule struct {
	max int
}

func NewLocRule(max *int) Rule {
	if max == nil {
		return &locRule{max: 0}
	}
	return &locRule{max: *max}
}

func (l *locRule) Name() string {
	return "max_loc"
}

func (l *locRule) Description() string {
	return "Checks the lines of code in a file"
}

func (l *locRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {

	if l.max == 0 || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Volume == nil || file.Stmts.Analyze.Volume.Loc == nil {
		return
	}

	value := int(*file.Stmts.Analyze.Volume.Loc)

	if l.max > 0 && value > l.max {
		addError(issue.RequirementError{
			Severity: issue.SeverityMedium,
			Message:  fmt.Sprintf("Too many Lines of code (%d > %d)", value, l.max),
			Code:     l.Name(),
		})
		return
	}

	addSuccess("Max Lines of code OK")
}
