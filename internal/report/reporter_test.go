package report

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// Mock reporter for testing
type mockReporter struct {
	reports []GeneratedReport
	err     error
}

func (m *mockReporter) Generate(files []*pb.File, projectAggregated analyzer.ProjectAggregated) ([]GeneratedReport, error) {
	return m.reports, m.err
}

func TestReporter_Interface(t *testing.T) {
	reports := []GeneratedReport{
		{Path: "/test/report.html", Type: "html"},
	}
	
	reporter := &mockReporter{reports: reports}
	
	result, err := reporter.Generate(nil, analyzer.ProjectAggregated{})
	
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	if len(result) != 1 {
		t.Errorf("expected 1 report, got %d", len(result))
	}
	
	if result[0].Path != "/test/report.html" {
		t.Errorf("expected path '/test/report.html', got %s", result[0].Path)
	}
}
