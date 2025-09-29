package report

import (
	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
)

type Reporter interface {
	// generates a report based on the files and the project aggregated data
	Generate(files []*pb.File, projectAggregated analyzer.ProjectAggregated) ([]GeneratedReport, error)
}
