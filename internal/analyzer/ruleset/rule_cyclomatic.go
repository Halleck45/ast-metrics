package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type cyclomaticRule struct {
	max *int
}

func NewCyclomaticRule(max *int) Rule {
	return &cyclomaticRule{max: max}
}

func (r *cyclomaticRule) Name() string {
	return "cyclomatic_complexity"
}

func (r *cyclomaticRule) Description() string {
	return "Checks the cyclomatic complexity of functions"
}

func (r *cyclomaticRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.max == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil || file.Stmts.Analyze.Complexity.Cyclomatic == nil {
		return
	}

	value := int(*file.Stmts.Analyze.Complexity.Cyclomatic)
	if r.max != nil && value > *r.max {
		addError(issue.RequirementError{
			Severity: issue.SeverityMedium,
			Code:     r.Name(),
			Message:  fmt.Sprintf("Cyclomatic complexity too high: got %d (max: %d)", value, *r.max),
		})
		return
	}
	addSuccess("Cyclomatic complexity OK")
}
