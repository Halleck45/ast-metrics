package Report

import (
	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type Reporter interface {
	// generates a report based on the files and the project aggregated data
	Generate(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) ([]GeneratedReport, error)
}
