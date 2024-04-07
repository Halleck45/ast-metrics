package Analyzer

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/src/Configuration"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type RequirementsEvaluator struct {
	Requirements Configuration.ConfigurationRequirements
}

type EvaluationResult struct {
	Files             []*pb.File
	ProjectAggregated ProjectAggregated
	Errors            []string
	Successes         []string
	Succeeded         bool
}

func NewRequirementsEvaluator(requirements Configuration.ConfigurationRequirements) *RequirementsEvaluator {
	return &RequirementsEvaluator{Requirements: requirements}
}

func (r *RequirementsEvaluator) Evaluate(files []*pb.File, projectAggregated ProjectAggregated) EvaluationResult {
	evaluation := EvaluationResult{
		Files:             files,
		ProjectAggregated: projectAggregated,
		Succeeded:         true,
		Successes:         []string{},
		Errors:            []string{},
	}

	if r.Requirements.Rules == nil {
		return evaluation
	}

	// Cyclomatic
	if r.Requirements.Rules.CyclomaticComplexity != nil {
		cyclomatic := r.Requirements.Rules.CyclomaticComplexity
		excludedFiles := cyclomatic.ExcludePatterns
		for _, file := range files {

			// if the file is excluded, we skip it (use regular expression)
			excluded := false
			if excludedFiles != nil {
				for _, pattern := range excludedFiles {
					if regexp.MustCompile(pattern).MatchString(file.Path) {
						excluded = true
						break
					}
				}
			}

			if excluded {
				continue
			}

			if file.Stmts.Analyze.Complexity == nil {
				continue
			}

			if int(*file.Stmts.Analyze.Complexity.Cyclomatic) > cyclomatic.Max {
				evaluation.Errors = append(evaluation.Errors, fmt.Sprintf("Cyclomatic complexity too high in file %s: got %d (max: %d)", file.Path, *file.Stmts.Analyze.Complexity.Cyclomatic, cyclomatic.Max))
			} else {
				evaluation.Successes = append(evaluation.Successes, "Cyclomatic complexity OK in file "+file.Path)
			}
		}
	}

	// Lines of code
	if r.Requirements.Rules.Loc != nil {
		loc := r.Requirements.Rules.Loc
		excludedFiles := loc.ExcludePatterns

		for _, file := range files {

			// if the file is excluded, we skip it (use regular expression)
			excluded := false
			if excludedFiles != nil {
				for _, pattern := range excludedFiles {
					if regexp.MustCompile(pattern).MatchString(file.Path) {
						excluded = true
						break
					}
				}
			}

			if excluded {
				continue
			}

			if file.Stmts.Analyze == nil || file.Stmts.Analyze.Volume.Loc == nil {
				continue
			}

			if int(*file.Stmts.Analyze.Volume.Loc) > loc.Max {
				evaluation.Errors = append(evaluation.Errors, fmt.Sprintf("Lines of code too high in file %s: got %d (max: %d)", file.Path, *file.Stmts.Analyze.Volume.Loc, loc.Max))
			} else {
				evaluation.Successes = append(evaluation.Successes, "Lines of code OK in file "+file.Path)
			}
		}
	}

	// Coupling and dependencies
	if r.Requirements.Rules.Coupling != nil && r.Requirements.Rules.Coupling.Forbidden != nil {

		for _, file := range files {

			if file.Stmts.StmtExternalDependencies == nil {
				continue
			}
			hasError := false

			for _, forbidden := range r.Requirements.Rules.Coupling.Forbidden {

				// Should match "from" expression
				if !regexp.MustCompile(forbidden.From).MatchString(file.Path) {
					continue
				}

				dependencies := file.Stmts.StmtExternalDependencies

				for _, dependency := range dependencies {

					// Should match "to" expression
					if regexp.MustCompile(forbidden.To).MatchString(dependency.ClassName) {
						evaluation.Errors = append(evaluation.Errors, fmt.Sprintf("Forbidden coupling between %s and %s", file.Path, dependency.ClassName))
						hasError = true
						break
					}
				}
			}

			if !hasError {
				evaluation.Successes = append(evaluation.Successes, "Coupling OK in file "+file.Path)
			}
		}
	}

	if len(evaluation.Errors) > 0 {
		evaluation.Succeeded = false
	}

	return evaluation
}
