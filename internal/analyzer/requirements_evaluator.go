package analyzer

import (
	"fmt"
	"regexp"

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

func NewRequirementsEvaluator(requirements configuration.ConfigurationRequirements) *RequirementsEvaluator {
	return &RequirementsEvaluator{Requirements: requirements}
}

// Rules with boundaries, to evaluate
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

	if r.Requirements.Rules == nil {
		return evaluation
	}

	rulesToCheck := []BoundariesRule{
		{Name: "cyclomatic", Rule: r.Requirements.Rules.CyclomaticComplexity, Label: "Cyclomatic complexity"},
		{Name: "loc", Rule: r.Requirements.Rules.Loc, Label: "Lines of code"},
		{Name: "maintainability", Rule: r.Requirements.Rules.Maintainability, Label: "Maintainability"},
	}

	for _, rule := range rulesToCheck {

		if rule.Rule == nil {
			continue
		}

		excludedFiles := []string{}
		if rule.Rule.ExcludePatterns != nil {
			excludedFiles = rule.Rule.ExcludePatterns
		}

		for _, file := range files {

			// if the file is excluded, we skip it (use regular expression)
			excluded := false
			for _, pattern := range excludedFiles {
				if regexp.MustCompile(pattern).MatchString(file.Path) {
					excluded = true
					break
				}
			}

			if excluded {
				continue
			}

			valueOfMetric := 0
			switch rule.Name {
			case "cyclomatic":
				if file.Stmts.Analyze.Complexity == nil || file.Stmts.Analyze.Complexity.Cyclomatic == nil {
					continue
				}
				valueOfMetric = int(*file.Stmts.Analyze.Complexity.Cyclomatic)
				r.EvaluateRule(rule, valueOfMetric, file, &evaluation)
			case "loc":
				if file.Stmts.Analyze.Volume == nil || file.Stmts.Analyze.Volume.Loc == nil {
					continue
				}
				valueOfMetric = int(*file.Stmts.Analyze.Volume.Loc)
				r.EvaluateRule(rule, valueOfMetric, file, &evaluation)
			case "maintainability":
				for _, class := range file.Stmts.StmtClass {
					if class.Stmts.Analyze.Maintainability == nil || class.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil {
						continue
					}
					valueOfMetric = int(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)
					r.EvaluateRule(rule, valueOfMetric, file, &evaluation)
				}
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

func (r *RequirementsEvaluator) EvaluateRule(rule BoundariesRule, valueOfMetric int, file *pb.File, evaluation *EvaluationResult) {

	maxExpected := rule.Rule.Max
	minExpected := rule.Rule.Min

	if maxExpected > 0 && valueOfMetric > maxExpected {
		evaluation.Errors = append(evaluation.Errors, fmt.Sprintf("%s too high in file %s: got %d (max: %d)", rule.Label, file.Path, valueOfMetric, maxExpected))
		return
	}

	if minExpected > 0 && valueOfMetric < minExpected {
		evaluation.Errors = append(evaluation.Errors, fmt.Sprintf("%s too low in file %s: got %d (min: %d)", rule.Label, file.Path, valueOfMetric, minExpected))
		return
	}

	evaluation.Successes = append(evaluation.Successes, fmt.Sprintf("%s OK in file %s", rule.Label, file.Path))
}
