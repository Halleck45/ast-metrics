package report

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ast-metrics/ast-metrics/internal/analyzer"
	requirement "github.com/ast-metrics/ast-metrics/internal/analyzer/requirement"
	pb "github.com/ast-metrics/ast-metrics/pb"
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
	// MaxLevel caps the SARIF level of every result and rule. Empty means no cap.
	MaxLevel string
}

func NewSarifReportGenerator(reportPath string, maxLevel string) Reporter {
	return &SarifReportGenerator{ReportPath: reportPath, MaxLevel: maxLevel}
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

	if err := writeSarifFile(g.ReportPath, outcomes, g.MaxLevel); err != nil {
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
func GenerateSarifFromOutcomes(reportPath string, outcomes []requirement.RuleOutcome, maxLevel string) (GeneratedReport, error) {
	if reportPath == "" {
		return GeneratedReport{}, fmt.Errorf("report path is empty")
	}
	if err := writeSarifFile(reportPath, outcomes, maxLevel); err != nil {
		return GeneratedReport{}, err
	}
	return GeneratedReport{Path: reportPath, Type: "file", Description: "SARIF report of requirement violations", Icon: "📄"}, nil
}

func writeSarifFile(reportPath string, outcomes []requirement.RuleOutcome, maxLevel string) error {
	if err := validateSarifLevel(maxLevel); err != nil {
		return err
	}

	log := sarifLog{
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{Driver: sarifDriver{Name: "ast-metrics", InformationURI: "https://github.com/ast-metrics/ast-metrics"}},
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
			DefaultConfiguration: &sarifRuleConfig{Level: capSarifLevel(mapSeverity(out.Severity), maxLevel)},
			// Quality tags (and no security-severity): GitHub code scanning
			// then renders the alert with a plain error/warning/note severity
			// instead of a security severity.
			Properties: map[string]interface{}{
				"tags":             []string{"quality", "maintainability"},
				"problem.severity": capSarifLevel(mapSeverity(out.Severity), maxLevel),
			},
		})
	}

	for _, out := range outcomes {
		level := capSarifLevel(mapSeverity(out.Severity), maxLevel)
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

// sarifLevelRank orders the SARIF levels ast-metrics emits, from the quietest
// to the loudest.
var sarifLevelRank = map[string]int{"note": 1, "warning": 2, "error": 3}

// SarifLevels lists the accepted values for the SARIF level ceiling.
var SarifLevels = []string{"error", "warning", "note"}

func validateSarifLevel(maxLevel string) error {
	if maxLevel == "" {
		return nil
	}
	if _, ok := sarifLevelRank[maxLevel]; !ok {
		return fmt.Errorf("invalid SARIF level %q: expected one of %s", maxLevel, strings.Join(SarifLevels, ", "))
	}
	return nil
}

// capSarifLevel lowers a level to the given ceiling, leaving it untouched when
// no ceiling is set. GitHub code scanning fails its own pull request check as
// soon as a new error level alert appears in the diff, whatever the quality gate
// decided; capping the level keeps code scanning purely informative without
// downgrading the severity of the finding itself.
func capSarifLevel(level string, maxLevel string) string {
	if maxLevel == "" {
		return level
	}
	ceiling, ok := sarifLevelRank[maxLevel]
	if !ok {
		return level
	}
	if sarifLevelRank[level] > ceiling {
		return maxLevel
	}
	return level
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
