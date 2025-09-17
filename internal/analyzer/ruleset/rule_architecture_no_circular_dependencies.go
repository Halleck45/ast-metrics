package ruleset

import (
	"fmt"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

type noCircularDependenciesRule struct {
	enabled bool
}

func NewNoCircularDependenciesRule(enabled *bool) Rule {
	if enabled == nil {
		return &noCircularDependenciesRule{enabled: false}
	}
	return &noCircularDependenciesRule{enabled: *enabled}
}

func (r *noCircularDependenciesRule) Name() string {
	return "no_circular_dependencies"
}

func (r *noCircularDependenciesRule) Description() string {
	return "Detect circular dependencies between classes"
}

func (r *noCircularDependenciesRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if !r.enabled || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Coupling == nil {
		return
	}

	// Simple heuristic: high instability might indicate circular dependencies
	if file.Stmts.Analyze.Coupling.Instability > 0.8 {
		if file.Stmts.Analyze.Coupling.Afferent > 0 && file.Stmts.Analyze.Coupling.Efferent > 0 {
			addError(issue.RequirementError{
				Severity: issue.SeverityLow,
				Message:  fmt.Sprintf("File may have circular dependencies (instability: %.2f)", file.Stmts.Analyze.Coupling.Instability),
				Code:     r.Name(),
			})
			return
		}
	}

	addSuccess(fmt.Sprintf("No circular dependencies detected in file %s", file.Path))
}
