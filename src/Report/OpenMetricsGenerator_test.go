package Report

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"google.golang.org/protobuf/proto"

	"github.com/halleck45/ast-metrics/src/Analyzer"
)

func TestGenerate_ReportPathEmpty(t *testing.T) {
	v := &OpenMetricsReportGenerator{ReportPath: ""}
	files := []*pb.File{}
	projectAggregated := Analyzer.ProjectAggregated{}

	reports, err := v.Generate(files, projectAggregated)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if reports != nil {
		t.Fatalf("expected nil reports, got %v", reports)
	}
}

func TestGenerate_EmptyFiles(t *testing.T) {
	v := &OpenMetricsReportGenerator{ReportPath: "test_report"}
	files := []*pb.File{}
	projectAggregated := Analyzer.ProjectAggregated{}

	reports, err := v.Generate(files, projectAggregated)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].Path != "test_report" {
		t.Fatalf("expected report path 'test_report', got %s", reports[0].Path)
	}
}

func TestGenerate_ValidFiles(t *testing.T) {
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
						MaintainabilityIndex:                proto.Float32(75.5),
						MaintainabilityIndexWithoutComments: proto.Float32(70.0),
					},
					Coupling: &pb.Coupling{
						Afferent: *proto.Int32(5),
						Efferent: *proto.Int32(3),
					},
				},
			},
		},
	}
	projectAggregated := Analyzer.ProjectAggregated{}

	reports, err := v.Generate(files, projectAggregated)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].Path != "test_report" {
		t.Fatalf("expected report path 'test_report', got %s", reports[0].Path)
	}
}

func TestGenerate_CreateFileError(t *testing.T) {
	v := &OpenMetricsReportGenerator{ReportPath: "/invalid_path/test_report"}
	files := []*pb.File{}
	projectAggregated := Analyzer.ProjectAggregated{}

	_, err := v.Generate(files, projectAggregated)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
