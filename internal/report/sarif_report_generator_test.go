package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ast-metrics/ast-metrics/internal/analyzer"
	requirement "github.com/ast-metrics/ast-metrics/internal/analyzer/requirement"
)

func TestSarifGenerator_EmptyPath(t *testing.T) {
	gen := &SarifReportGenerator{ReportPath: ""}
	reports, err := gen.Generate(nil, analyzer.ProjectAggregated{})
	assert.NoError(t, err)
	assert.Nil(t, reports)
}

func TestSarifGenerator_GenerateWithOneViolation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.sarif.json")
	gen := &SarifReportGenerator{ReportPath: path}

	pa := analyzer.ProjectAggregated{
		Evaluation: &requirement.EvaluationResult{
			Errors: []requirement.RuleOutcome{
				{Rule: "max_cyclomatic", Severity: requirement.SeverityHigh, Message: "Cyclomatic complexity too high (20 > 10)", File: "/tmp/file.go", Line: 42},
			},
		},
	}

	reports, err := gen.Generate(nil, pa)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(reports))
	// file exists
	_, statErr := os.Stat(path)
	assert.NoError(t, statErr)
	// content sanity
	b, readErr := os.ReadFile(path)
	assert.NoError(t, readErr)
	content := string(b)
	assert.Contains(t, content, "\"version\": \"2.1.0\"")
	assert.Contains(t, content, "\"ruleId\": \"max_cyclomatic\"")
	assert.Contains(t, content, "\"level\": \"error\"")
	// Region carries the violation line for GitHub annotations
	assert.Contains(t, content, "\"startLine\": 42")
	// Rule metadata is declared and referenced
	assert.Contains(t, content, "\"rules\": [")
	assert.Contains(t, content, "\"ruleIndex\": 0")
	// Stable fingerprint for deduplication
	assert.Contains(t, content, "\"partialFingerprints\"")
	assert.Contains(t, content, "\"astMetrics/v1\"")
	// Quality tags so GitHub code scanning does not treat findings as security alerts
	assert.Contains(t, content, "\"maintainability\"")
	assert.Contains(t, content, "\"quality\"")
	assert.NotContains(t, content, "security-severity")
	assert.Contains(t, content, "\"defaultConfiguration\"")
}

func TestSarifGenerator_FileLevelFindingDefaultsToLineOne(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.sarif.json")
	gen := &SarifReportGenerator{ReportPath: path}

	pa := analyzer.ProjectAggregated{
		Evaluation: &requirement.EvaluationResult{
			Errors: []requirement.RuleOutcome{
				// No Line set: a file-level finding must still anchor to line 1
				{Rule: "max_loc", Severity: requirement.SeverityMedium, Message: "Too many Lines of code", File: "/tmp/file.go"},
			},
		},
	}

	_, err := gen.Generate(nil, pa)
	assert.NoError(t, err)
	b, readErr := os.ReadFile(path)
	assert.NoError(t, readErr)
	assert.Contains(t, string(b), "\"startLine\": 1")
}

func TestGenerateSarifFromOutcomes_Helper(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lint.sarif.json")
	out := []requirement.RuleOutcome{{Rule: "foo", Severity: requirement.SeverityLow, Message: "Something", File: "a.go"}}
	report, err := GenerateSarifFromOutcomes(path, out, "")
	assert.NoError(t, err)
	assert.Equal(t, path, report.Path)
	_, statErr := os.Stat(path)
	assert.NoError(t, statErr)
}

func TestSarifGenerator_MaxLevelCapsTheLevel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.sarif.json")
	gen := &SarifReportGenerator{ReportPath: path, MaxLevel: "warning"}

	pa := analyzer.ProjectAggregated{
		Evaluation: &requirement.EvaluationResult{
			Errors: []requirement.RuleOutcome{
				{Rule: "max_cyclomatic", Severity: requirement.SeverityHigh, Message: "Cyclomatic complexity too high", File: "/tmp/file.go", Line: 42},
				{Rule: "max_loc", Severity: requirement.SeverityLow, Message: "Too many Lines of code", File: "/tmp/file.go", Line: 1},
			},
		},
	}

	_, err := gen.Generate(nil, pa)
	assert.NoError(t, err)
	b, readErr := os.ReadFile(path)
	assert.NoError(t, readErr)
	content := string(b)
	// The high severity finding is lowered to the ceiling, so GitHub code
	// scanning stops failing its own pull request check.
	assert.NotContains(t, content, "\"level\": \"error\"")
	assert.Contains(t, content, "\"level\": \"warning\"")
	// A level already below the ceiling is left untouched.
	assert.Contains(t, content, "\"level\": \"note\"")
}

func TestSarifGenerator_MaxLevelRejectsUnknownValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.sarif.json")
	out := []requirement.RuleOutcome{{Rule: "foo", Severity: requirement.SeverityHigh, Message: "Something", File: "a.go"}}

	_, err := GenerateSarifFromOutcomes(path, out, "critical")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SARIF level")
}

func TestCapSarifLevel(t *testing.T) {
	assert.Equal(t, "error", capSarifLevel("error", ""))
	assert.Equal(t, "warning", capSarifLevel("error", "warning"))
	assert.Equal(t, "note", capSarifLevel("error", "note"))
	assert.Equal(t, "note", capSarifLevel("note", "warning"))
	assert.Equal(t, "error", capSarifLevel("error", "unknown"))
}
