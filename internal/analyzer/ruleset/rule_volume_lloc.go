package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

type llocRule struct {
	max int
}

func NewLlocRule(max *int) Rule {
	if max == nil {
		return &llocRule{max: 0}
	}
	return &llocRule{max: *max}
}

func (l *llocRule) Name() string {
	return "max_lloc"
}

func (l *llocRule) Description() string {
	return "Checks the logical lines of code in a file"
}

func (l *llocRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if l.max == 0 || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Volume == nil || file.Stmts.Analyze.Volume.Lloc == nil {
		return
	}

	value := int(*file.Stmts.Analyze.Volume.Lloc)
	if l.max > 0 && value > l.max {
		addError(issue.RequirementError{
			Severity: issue.SeverityMedium,
			Message:  fmt.Sprintf("Logical lines of code too high : got %d (max: %d)", value, l.max),
			Code:     l.Name(),
		})
		return
	}

	addSuccess("Logical lines of code OK")
}
