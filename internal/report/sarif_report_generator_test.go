package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	requirement "github.com/halleck45/ast-metrics/internal/analyzer/requirement"
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
				{Rule: "max_cyclomatic", Severity: requirement.SeverityHigh, Message: "Cyclomatic complexity too high (20 > 10)", File: "/tmp/file.go"},
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
}

func TestGenerateSarifFromOutcomes_Helper(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lint.sarif.json")
	out := []requirement.RuleOutcome{{Rule: "foo", Severity: requirement.SeverityLow, Message: "Something", File: "a.go"}}
	report, err := GenerateSarifFromOutcomes(path, out)
	assert.NoError(t, err)
	assert.Equal(t, path, report.Path)
	_, statErr := os.Stat(path)
	assert.NoError(t, statErr)
}
