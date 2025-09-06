package requirement

import (
	"github.com/halleck45/ast-metrics/internal/analyzer/ruleset"
	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type RequirementsEvaluator struct {
	Requirements configuration.ConfigurationRequirements
}

type EvaluationResult struct {
	Files             []*pb.File
	ProjectAggregated ProjectAggregated
	Errors            []string
	Successes         []string
	Succeeded         bool
}

// minimal view of ProjectAggregated to avoid import cycle; use original type via alias in caller
// We reuse the original analyzer.ProjectAggregated at call site; here we just keep it opaque
// Define a tiny interface type to hold reference without methods
type ProjectAggregated struct{}

func NewRequirementsEvaluator(requirements configuration.ConfigurationRequirements) *RequirementsEvaluator {
	return &RequirementsEvaluator{Requirements: requirements}
}

// Rules with boundaries, to evaluate
// Deprecated path-based rules retained for backward compatibility

type BoundariesRule struct {
	Name  string
	Rule  *configuration.ConfigurationDefaultRule
	Label string
}

func (r *RequirementsEvaluator) Evaluate(files []*pb.File, projectAggregated ProjectAggregated) EvaluationResult {
	evaluation := EvaluationResult{
		Files:             files,
		ProjectAggregated: projectAggregated,
		Succeeded:         true,
		Successes:         []string{},
		Errors:            []string{},
	}

	// Delegate to registry-based rulesets
	for _, rlset := range ruleset.Registry(&r.Requirements).EnabledRulesets() {
		for _, rule := range rlset.Enabled() {
			for _, file := range files {
				rule.CheckFile(
					file,
					func(err string) { evaluation.Errors = append(evaluation.Errors, err) },
					func(ok string) { evaluation.Successes = append(evaluation.Successes, ok) },
				)
			}
		}
	}

	if len(evaluation.Errors) > 0 {
		evaluation.Succeeded = false
	}

	return evaluation
}
