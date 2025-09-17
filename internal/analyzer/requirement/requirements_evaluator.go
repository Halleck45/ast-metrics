package requirement

import (
	"regexp"
	"strings"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	"github.com/halleck45/ast-metrics/internal/analyzer/ruleset"
	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// Expose Severity and RequirementError in this package via alias to avoid import cycles
type Severity = issue.Severity
type RequirementError = issue.RequirementError

const (
	SeverityUnknown Severity = issue.SeverityUnknown
	SeverityLow     Severity = issue.SeverityLow
	SeverityMedium  Severity = issue.SeverityMedium
	SeverityHigh    Severity = issue.SeverityHigh
)

// RuleOutcome is a structured message produced by rules
// Message should not include severity prefix; Rule is the rule name; File is the concerned file path when applicable.
type RuleOutcome struct {
	Severity Severity
	Rule     string
	Message  string
	File     string
}

type RequirementsEvaluator struct {
	Requirements configuration.ConfigurationRequirements
}

type EvaluationResult struct {
	Files             []*pb.File
	ProjectAggregated ProjectAggregated
	Errors            []RuleOutcome
	Successes         []RuleOutcome
	Succeeded         bool
}

// method to get number of errors by severity
func (er *EvaluationResult) CountErrorsBySeverity(sev Severity) int {
	count := 0
	for _, e := range er.Errors {
		if e.Severity == sev {
			count++
		}
	}
	return count
}

// methd to get the number of errors by severity (as string)
func (er *EvaluationResult) CountErrorsBySeverityString(sev string) int {
	sev = strings.ToLower(strings.TrimSpace(sev))
	var severity Severity
	switch sev {
	case "high":
		severity = SeverityHigh
	case "medium":
		severity = SeverityMedium
	case "low":
		severity = SeverityLow
	default:
		severity = SeverityUnknown
	}

	return er.CountErrorsBySeverity(severity)
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

// parseSeverityFromMessage extracts a leading [Label] severity and returns (severity, cleanedMessage)
func parseSeverityFromMessage(msg string) (Severity, string) {
	trim := strings.TrimSpace(msg)
	sev := SeverityUnknown
	// handle two optional bracketed prefixes (e.g., [Medium][rule])
	re := regexp.MustCompile(`^(\[[^\]]+\])+`)
	if loc := re.FindStringIndex(trim); loc != nil && loc[0] == 0 {
		prefix := trim[loc[0]:loc[1]]
		// Find a severity token inside the brackets
		lower := strings.ToLower(prefix)
		switch {
		case strings.Contains(lower, "[high]") || strings.Contains(lower, "[medium→high]") || strings.Contains(lower, "[medium->high]"):
			sev = SeverityHigh
		case strings.Contains(lower, "[medium]"):
			sev = SeverityMedium
		case strings.Contains(lower, "[low]") || strings.Contains(lower, "[low→medium]") || strings.Contains(lower, "[low->medium]"):
			sev = SeverityLow
		}
		trim = strings.TrimSpace(trim[loc[1]:])
	}
	return sev, trim
}

func (r *RequirementsEvaluator) Evaluate(files []*pb.File, projectAggregated ProjectAggregated) EvaluationResult {
	evaluation := EvaluationResult{
		Files:             files,
		ProjectAggregated: projectAggregated,
		Succeeded:         true,
		Successes:         []RuleOutcome{},
		Errors:            []RuleOutcome{},
	}

	// Delegate to registry-based rulesets
	for _, rlset := range ruleset.Registry(&r.Requirements).EnabledRulesets() {
		for _, rule := range rlset.Enabled() {
			for _, file := range files {

				// Exclusions
				if r.Requirements.Exclude != nil {
					for _, pattern := range r.Requirements.Exclude {
						if regexp.MustCompile(pattern).MatchString(file.Path) {
							continue
						}
					}
				}

				rule := rule // capture
				file := file // capture
				rule.CheckFile(
					file,
					func(err RequirementError) {
						// Severity provided by rule; message should be clean already
						evaluation.Errors = append(evaluation.Errors, RuleOutcome{Severity: err.Severity, Rule: rule.Name(), Message: err.Message, File: file.Path})
					},
					func(ok string) {
						sev, msg := parseSeverityFromMessage(ok)
						evaluation.Successes = append(evaluation.Successes, RuleOutcome{Severity: sev, Rule: rule.Name(), Message: msg, File: file.Path})
					},
				)
			}
		}
	}

	if len(evaluation.Errors) > 0 {
		evaluation.Succeeded = false
	}

	return evaluation
}
