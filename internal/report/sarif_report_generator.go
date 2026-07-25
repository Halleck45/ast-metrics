package report

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	requirement "github.com/halleck45/ast-metrics/internal/analyzer/requirement"
	pb "github.com/halleck45/ast-metrics/pb"
)

// Minimal SARIF 2.1.0 structures we need
// Spec: https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results,omitempty"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string          `json:"name"`
	InformationURI string          `json:"informationUri,omitempty"`
	Version        string          `json:"version,omitempty"`
	Rules          []sarifRuleMeta `json:"rules,omitempty"`
}

type sarifRuleMeta struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name,omitempty"`
	ShortDescription     *sarifMessage          `json:"shortDescription,omitempty"`
	DefaultConfiguration *sarifRuleConfig       `json:"defaultConfiguration,omitempty"`
	Properties           map[string]interface{} `json:"properties,omitempty"`
}

type sarifRuleConfig struct {
	Level string `json:"level,omitempty"`
}

type sarifResult struct {
	RuleID              string            `json:"ruleId,omitempty"`
	RuleIndex           *int              `json:"ruleIndex,omitempty"`
	Level               string            `json:"level,omitempty"`
	Message             sarifMessage      `json:"message"`
	Locations           []sarifLocation   `json:"locations,omitempty"`
	PartialFingerprints map[string]string `json:"partialFingerprints,omitempty"`
	Properties          map[string]string `json:"properties,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion          `json:"region,omitempty"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine,omitempty"`
}

// SarifReportGenerator implements Reporter and uses requirement outcomes when present
// in the provided projectAggregated to create a SARIF file.

type SarifReportGenerator struct {
	ReportPath string
}

func NewSarifReportGenerator(reportPath string) Reporter {
	return &SarifReportGenerator{ReportPath: reportPath}
}

func (g *SarifReportGenerator) Generate(files []*pb.File, projectAggregated analyzer.ProjectAggregated) ([]GeneratedReport, error) {
	if g.ReportPath == "" {
		return nil, nil
	}

	// Collect outcomes from evaluation if available
	var outcomes []requirement.RuleOutcome
	if projectAggregated.Evaluation != nil {
		outcomes = projectAggregated.Evaluation.Errors
	}

	if err := writeSarifFile(g.ReportPath, outcomes); err != nil {
		return nil, err
	}

	reports := []GeneratedReport{
		{
			Path:        g.ReportPath,
			Type:        "file",
			Description: "SARIF report of requirement violations",
			Icon:        "📄",
		},
	}
	return reports, nil
}

// Export function to build SARIF directly from outcomes (to be used by lint command)
func GenerateSarifFromOutcomes(reportPath string, outcomes []requirement.RuleOutcome) (GeneratedReport, error) {
	if reportPath == "" {
		return GeneratedReport{}, fmt.Errorf("report path is empty")
	}
	if err := writeSarifFile(reportPath, outcomes); err != nil {
		return GeneratedReport{}, err
	}
	return GeneratedReport{Path: reportPath, Type: "file", Description: "SARIF report of requirement violations", Icon: "📄"}, nil
}

func writeSarifFile(reportPath string, outcomes []requirement.RuleOutcome) error {
	log := sarifLog{
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{Driver: sarifDriver{Name: "ast-metrics", InformationURI: "https://github.com/halleck45/ast-metrics"}},
				Results: make([]sarifResult, 0, len(outcomes)),
			},
		},
	}

	// Build the rule metadata catalog. Each distinct rule id is declared once
	// in tool.driver.rules and referenced from results via ruleIndex, as
	// recommended by the SARIF spec and expected by GitHub code scanning.
	ruleIndexByID := make(map[string]int)
	for _, out := range outcomes {
		if out.Rule == "" {
			continue
		}
		if _, ok := ruleIndexByID[out.Rule]; ok {
			continue
		}
		idx := len(log.Runs[0].Tool.Driver.Rules)
		ruleIndexByID[out.Rule] = idx
		log.Runs[0].Tool.Driver.Rules = append(log.Runs[0].Tool.Driver.Rules, sarifRuleMeta{
			ID:                   out.Rule,
			Name:                 out.Rule,
			ShortDescription:     &sarifMessage{Text: out.Rule},
			DefaultConfiguration: &sarifRuleConfig{Level: mapSeverity(out.Severity)},
			// Quality tags (and no security-severity): GitHub code scanning
			// then renders the alert with a plain error/warning/note severity
			// instead of a security severity.
			Properties: map[string]interface{}{
				"tags":             []string{"quality", "maintainability"},
				"problem.severity": mapSeverity(out.Severity),
			},
		})
	}

	for _, out := range outcomes {
		level := mapSeverity(out.Severity)
		res := sarifResult{
			RuleID:  out.Rule,
			Level:   level,
			Message: sarifMessage{Text: out.Message},
			Properties: map[string]string{
				"rule": out.Rule,
			},
		}
		if idx, ok := ruleIndexByID[out.Rule]; ok {
			i := idx
			res.RuleIndex = &i
		}
		if out.File != "" {
			// GitHub places annotations using a physical region; startLine must
			// be >= 1. File-level findings (no specific line) anchor to line 1.
			startLine := out.Line
			if startLine < 1 {
				startLine = 1
			}
			res.Locations = []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: out.File},
						Region:           &sarifRegion{StartLine: startLine},
					},
				},
			}
		}
		// Stable fingerprint so GitHub can track an alert across commits and
		// avoid duplicates. Built from rule + file + line only (not the volatile
		// metric values in the message), so the alert persists as code evolves.
		res.PartialFingerprints = map[string]string{
			"astMetrics/v1": fingerprint(out.Rule, out.File, out.Line),
		}
		log.Runs[0].Results = append(log.Runs[0].Results, res)
	}

	f, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("cannot create SARIF report at %s: %w", reportPath, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(log); err != nil {
		return fmt.Errorf("cannot write SARIF report: %w", err)
	}
	return nil
}

// fingerprint returns a stable hex digest identifying a finding by rule,
// file and line, independent of the run.
func fingerprint(rule, file string, line int) string {
	h := sha256.Sum256([]byte(rule + "\x00" + file + "\x00" + strconv.Itoa(line)))
	return hex.EncodeToString(h[:])
}

func mapSeverity(sev requirement.Severity) string {
	switch sev {
	case requirement.SeverityHigh:
		return "error"
	case requirement.SeverityMedium:
		return "warning"
	case requirement.SeverityLow:
		return "note"
	default:
		return "warning"
	}
}
