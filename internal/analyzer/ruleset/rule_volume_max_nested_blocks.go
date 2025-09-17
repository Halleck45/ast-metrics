package ruleset

import (
	"fmt"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

type maxNestedBlocksRule struct {
	threshold int
}

func NewMaxNestedBlocksRule(threshold *int) Rule {
	if threshold == nil {
		return &maxNestedBlocksRule{threshold: 0}
	}
	return &maxNestedBlocksRule{threshold: *threshold}
}

func (r *maxNestedBlocksRule) Name() string {
	return "max_nested_blocks"
}

func (r *maxNestedBlocksRule) Description() string {
	return "Maximum nesting depth of blocks"
}

func (r *maxNestedBlocksRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.threshold == 0 || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil {
		return
	}

	// Use cyclomatic complexity as approximation for nesting depth
	if file.Stmts.Analyze.Complexity.Cyclomatic != nil && int(*file.Stmts.Analyze.Complexity.Cyclomatic) > r.threshold {
		addError(issue.RequirementError{
			Severity: issue.SeverityMedium,
			Message:  fmt.Sprintf("File has complexity of %d, maximum nesting allowed is %d", int(*file.Stmts.Analyze.Complexity.Cyclomatic), r.threshold),
			Code:     r.Name(),
		})
		return
	}

	addSuccess(fmt.Sprintf("Max nested blocks OK in file %s", file.Path))
}
