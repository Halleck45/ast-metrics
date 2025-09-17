package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	requirement "github.com/halleck45/ast-metrics/internal/analyzer/requirement"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
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
	Name           string `json:"name"`
	InformationURI string `json:"informationUri,omitempty"`
	Version        string `json:"version,omitempty"`
}

type sarifResult struct {
	RuleID    string              `json:"ruleId,omitempty"`
	Level     string              `json:"level,omitempty"`
	Message   sarifMessage        `json:"message"`
	Locations []sarifLocation     `json:"locations,omitempty"`
	Properties map[string]string  `json:"properties,omitempty"`
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
		if out.File != "" {
			res.Locations = []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: out.File},
					},
				},
			}
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
