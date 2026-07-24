package review

import (
	"encoding/json"
	"fmt"
	"strings"

	requirement "github.com/halleck45/ast-metrics/internal/analyzer/requirement"
)

// GateLabel returns "passed" or "failed" depending on the fail-on level.
// level is one of "high", "medium", "any" (alias of "low"), "never".
func (r *Result) EvaluateGate(level string) string {
	failed := false
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "high":
		failed = r.HasRegressionAtLeast(SeverityHigh)
	case "medium":
		failed = r.HasRegressionAtLeast(SeverityMedium)
	case "any", "low":
		failed = r.HasRegressionAtLeast(SeverityLow)
	case "", "never":
		failed = false
	}
	if failed {
		return "failed"
	}
	return "passed"
}

func (r *Result) headline() string {
	if r.Gate == "failed" {
		return "AST Metrics: quality gate failed"
	}
	return "AST Metrics: quality gate passed"
}

func (r *Result) statsLine() string {
	parts := []string{}
	changed := r.Summary.FilesChanged + r.Summary.FilesAdded + r.Summary.FilesDeleted
	parts = append(parts, fmt.Sprintf("%d file(s) changed", changed))
	parts = append(parts, fmt.Sprintf("%d new critical issue(s)", r.Summary.High))
	parts = append(parts, fmt.Sprintf("%d other regression(s)", r.Summary.Medium+r.Summary.Low))
	if r.Summary.Improvements > 0 {
		parts = append(parts, fmt.Sprintf("%d improvement(s)", r.Summary.Improvements))
	}
	return strings.Join(parts, ", ")
}

func (f *Finding) title() string {
	if f.Subject != "" {
		return f.Subject
	}
	return f.File
}

func (f *Finding) location() string {
	if f.Line > 0 {
		return fmt.Sprintf("%s:%d", f.File, f.Line)
	}
	return f.File
}

// Text renders a compact terminal report, capped to maxFindings regressions.
func (r *Result) Text(maxFindings int) string {
	var b strings.Builder
	b.WriteString(r.headline() + "\n\n")
	b.WriteString(r.statsLine() + "\n")

	if len(r.Regressions) > 0 {
		b.WriteString("\nRegressions:\n")
		for i, f := range r.Regressions {
			if maxFindings > 0 && i >= maxFindings {
				b.WriteString(fmt.Sprintf("  ... and %d more (see JSON or SARIF report)\n", len(r.Regressions)-maxFindings))
				break
			}
			b.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", strings.ToUpper(string(f.Severity)), f.title(), f.location()))
			b.WriteString("      " + f.Message + "\n")
			if f.Suggestion != "" {
				b.WriteString("      Suggested action: " + f.Suggestion + "\n")
			}
		}
	}

	if len(r.Improvements) > 0 {
		b.WriteString("\nImprovements:\n")
		for i, f := range r.Improvements {
			if i >= 3 {
				b.WriteString(fmt.Sprintf("  ... and %d more\n", len(r.Improvements)-3))
				break
			}
			b.WriteString(fmt.Sprintf("- %s (%s): %s\n", f.title(), f.location(), f.Message))
		}
	}

	b.WriteString("\nExisting debt is not reported. Methodology v" + MethodologyVersion + "\n")
	return b.String()
}

// Markdown renders a report suitable for a PR comment or a job summary.
func (r *Result) Markdown(maxFindings int) string {
	var b strings.Builder

	icon := "✅"
	if r.Gate == "failed" {
		icon = "❌"
	} else if len(r.Regressions) > 0 {
		icon = "⚠️"
	}
	b.WriteString("## " + icon + " " + r.headline() + "\n\n")
	b.WriteString(r.statsLine() + "\n")

	if len(r.Regressions) > 0 {
		b.WriteString("\n### Regressions\n\n")
		for i, f := range r.Regressions {
			if maxFindings > 0 && i >= maxFindings {
				b.WriteString(fmt.Sprintf("\n_... and %d more. Download the full report for details._\n", len(r.Regressions)-maxFindings))
				break
			}
			b.WriteString(fmt.Sprintf("- **%s** (`%s`, %s)\n", escapeMarkdown(f.title()), f.location(), f.Severity))
			b.WriteString("  " + escapeMarkdown(f.Message) + "\n")
			if f.Suggestion != "" {
				b.WriteString("  Suggested action: " + escapeMarkdown(f.Suggestion) + "\n")
			}
		}
	}

	if len(r.Improvements) > 0 {
		b.WriteString("\n### Improvements\n\n")
		for i, f := range r.Improvements {
			if i >= 3 {
				b.WriteString(fmt.Sprintf("\n_... and %d more._\n", len(r.Improvements)-3))
				break
			}
			b.WriteString(fmt.Sprintf("- **%s** (`%s`): %s\n", escapeMarkdown(f.title()), f.location(), escapeMarkdown(f.Message)))
		}
	}

	if len(r.Regressions) == 0 && len(r.Improvements) == 0 {
		b.WriteString("\nNo architectural change detected on modified code.\n")
	}

	b.WriteString("\n---\n")
	b.WriteString(fmt.Sprintf("_Compared with `%s`. Existing debt is not reported. Methodology v%s._\n", r.BaseRef, MethodologyVersion))
	return b.String()
}

// JSON renders the full, uncapped machine-readable report.
func (r *Result) JSON() (string, error) {
	out, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}

// ToRuleOutcomes converts regressions so they can feed the existing SARIF
// generator. Improvements are not exported to SARIF.
func (r *Result) ToRuleOutcomes() []requirement.RuleOutcome {
	outcomes := make([]requirement.RuleOutcome, 0, len(r.Regressions))
	for _, f := range r.Regressions {
		message := f.Message
		if f.Subject != "" {
			message = f.Subject + ": " + message
		}
		outcomes = append(outcomes, requirement.RuleOutcome{
			Severity: sarifSeverity(f.Severity),
			Rule:     f.Rule,
			Message:  message,
			File:     f.File,
			Line:     f.Line,
		})
	}
	return outcomes
}

func sarifSeverity(s Severity) requirement.Severity {
	switch s {
	case SeverityHigh:
		return requirement.SeverityHigh
	case SeverityMedium:
		return requirement.SeverityMedium
	default:
		return requirement.SeverityLow
	}
}

// escapeMarkdown neutralizes characters that could open an HTML tag or break
// a table cell. A lone ">" (e.g. in "8 -> 15") is harmless and kept as-is.
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer("<", "\\<", "|", "\\|")
	return replacer.Replace(s)
}
