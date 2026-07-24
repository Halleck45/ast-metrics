package review

import (
	"regexp"
	"strings"

	requirement "github.com/halleck45/ast-metrics/internal/analyzer/requirement"
)

var digitsPattern = regexp.MustCompile(`\d+(\.\d+)?`)

// DiffLint keeps only the requirement violations introduced by head: a
// violation already present in base is existing debt and is not reported.
// Matching ignores numeric values so that a metric drifting from 60 to 62
// on an already-failing rule is not flagged as new.
func DiffLint(headOutcomes []requirement.RuleOutcome, baseOutcomes []requirement.RuleOutcome, headRoot string, baseRoot string) []Finding {
	existing := map[string]bool{}
	for _, out := range baseOutcomes {
		existing[lintKey(out, baseRoot)] = true
	}

	findings := []Finding{}
	for _, out := range headOutcomes {
		if existing[lintKey(out, headRoot)] {
			continue
		}
		findings = append(findings, Finding{
			Kind:     KindRegression,
			Severity: lintSeverity(out.Severity),
			Rule:     "new-violation:" + out.Rule,
			File:     relativize(out.File, headRoot),
			Line:     out.Line,
			Message:  strings.ReplaceAll(out.Message, relativizeToken(headRoot), ""),
		})
	}
	return findings
}

func lintKey(out requirement.RuleOutcome, root string) string {
	normalized := digitsPattern.ReplaceAllString(out.Message, "#")
	normalized = strings.ReplaceAll(normalized, relativizeToken(root), "")
	return out.Rule + "\x00" + relativize(out.File, root) + "\x00" + normalized
}

// relativizeToken returns the root with a trailing separator, so absolute
// paths embedded in messages are neutralized before comparison.
func relativizeToken(root string) string {
	if root == "" {
		return "\x01never-matches\x01"
	}
	if !strings.HasSuffix(root, "/") {
		return root + "/"
	}
	return root
}

func lintSeverity(s requirement.Severity) Severity {
	switch s {
	case requirement.SeverityHigh:
		return SeverityHigh
	case requirement.SeverityMedium:
		return SeverityMedium
	default:
		return SeverityLow
	}
}
