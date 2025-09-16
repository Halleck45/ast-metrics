package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type noGodClassRule struct {
	enabled bool
}

func NewNoGodClassRule(enabled *bool) Rule {
	if enabled == nil {
		return &noGodClassRule{enabled: false}
	}
	return &noGodClassRule{enabled: *enabled}
}

func (r *noGodClassRule) Name() string {
	return "no_god_class"
}

func (r *noGodClassRule) Description() string {
	return "Detect god classes (classes with too many responsibilities)"
}

func (r *noGodClassRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if !r.enabled || file.Stmts == nil || file.Stmts.StmtClass == nil {
		return
	}

	classes := engine.GetClassesInFile(file)
	for _, class := range classes {
		if class.Stmts != nil && class.Stmts.Analyze != nil && class.Stmts.Analyze.Volume != nil && class.Stmts.Analyze.Volume.Loc != nil {
			loc := int(*class.Stmts.Analyze.Volume.Loc)
			methodCount := 0
			if class.Stmts != nil && class.Stmts.StmtFunction != nil {
				methodCount = len(class.Stmts.StmtFunction)
			}
			lcom := 0
			if class.Stmts.Analyze.ClassCohesion != nil && class.Stmts.Analyze.ClassCohesion.Lcom4 != nil {
				lcom = int(*class.Stmts.Analyze.ClassCohesion.Lcom4)
			}

			// God class heuristic: high LOC, many methods, high LCOM
			if loc > 500 && methodCount > 20 && lcom > 10 {
				addError(issue.RequirementError{
					Severity: issue.SeverityLow,
					Message:  fmt.Sprintf("God class detected: %d LOC, %d methods, LCOM4 %d", loc, methodCount, lcom),
					Code:     r.Name(),
				})
				return
			}
		}
	}

	addSuccess("No god classes detected")
}
