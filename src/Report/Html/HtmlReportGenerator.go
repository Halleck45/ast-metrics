package Report

import (
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ReportGenerator struct {
	// The path where the report will be generated
	ReportPath string
}

func NewReportGenerator(reportPath string) *ReportGenerator {
	return &ReportGenerator{
		ReportPath: reportPath,
	}
}

func (v *ReportGenerator) Generate(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) error {
	return nil
}
