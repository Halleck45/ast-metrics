package report

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/halleck45/ast-metrics/internal/analyzer"
)

func TestGenerateOpenMetricsReports(t *testing.T) {
	t.Run("Should not generate report when Path is empty", func(t *testing.T) {
		v := NewOpenMetricsReportGenerator("")
		files := []*pb.File{}
		projectAggregated := analyzer.ProjectAggregated{}

		reports, err := v.Generate(files, projectAggregated)
		assert.Nil(t, reports)
		assert.Nil(t, err)
	})

	t.Run("Should generate report event when source code contains empty file", func(t *testing.T) {
		v := &OpenMetricsReportGenerator{ReportPath: "test_report"}
		files := []*pb.File{}
		projectAggregated := analyzer.ProjectAggregated{}

		reports, err := v.Generate(files, projectAggregated)
		assert.Nil(t, err)
		assert.Len(t, reports, 1)
		assert.Equal(t, "test_report", reports[0].Path)
	})

	t.Run("Should generate report are sources contain valid files", func(t *testing.T) {
		v := &OpenMetricsReportGenerator{ReportPath: "test_report"}
		files := []*pb.File{
			{
				Path: "file1",
				Stmts: &pb.Stmts{
					Analyze: &pb.Analyze{
						Complexity: &pb.Complexity{Cyclomatic: proto.Int32(10)},
						Volume: &pb.Volume{
							Loc:  proto.Int32(100),
							Lloc: proto.Int32(80),
							Cloc: proto.Int32(20),
						},
						Maintainability: &pb.Maintainability{
							MaintainabilityIndex:                proto.Float64(75.5),
							MaintainabilityIndexWithoutComments: proto.Float64(70.0),
						},
						Coupling: &pb.Coupling{
							Afferent: *proto.Int32(5),
							Efferent: *proto.Int32(3),
						},
					},
				},
			},
		}
		projectAggregated := analyzer.ProjectAggregated{}

		reports, err := v.Generate(files, projectAggregated)
		assert.Nil(t, err)
		assert.Len(t, reports, 1)
		assert.Equal(t, "test_report", reports[0].Path)
	})

	t.Run("Should not generate report when path is incorrect", func(t *testing.T) {
		v := &OpenMetricsReportGenerator{ReportPath: "/invalid_path/test_report"}
		files := []*pb.File{}
		projectAggregated := analyzer.ProjectAggregated{}

		_, err := v.Generate(files, projectAggregated)
		assert.NotNil(t, err)
	})
}
