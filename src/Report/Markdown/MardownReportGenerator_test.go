package Report

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		reportPath  string
		expectError bool
	}{
		{
			name:        "Test with valid report path",
			reportPath:  "/tmp/report.md",
			expectError: false,
		},
		{
			name:        "Test with empty report path",
			reportPath:  "",
			expectError: false,
		},
		{
			name:        "Test with non-writable report path",
			reportPath:  "/nonexistent/report.md",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := &MarkdownReportGenerator{ReportPath: tt.reportPath}
			files := []*pb.File{}
			projectAggregated := Analyzer.ProjectAggregated{}

			_, err := generator.Generate(files, projectAggregated)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				} else {
					if tt.reportPath != "" {
						if _, err := os.Stat(tt.reportPath); os.IsNotExist(err) {
							t.Errorf("Report file was not created")
						} else {
							// cleanup
							os.Remove(tt.reportPath)
						}
					}
				}
			}
		})
	}
}

func TestGenerateWithTemplateFiles(t *testing.T) {
	// This test assumes that a valid template file "index.md" exists in the templates directory
	generator := &MarkdownReportGenerator{ReportPath: "/tmp/report.md"}
	files := []*pb.File{}
	projectAggregated := Analyzer.ProjectAggregated{}

	// Create a temporary template file
	ioutil.WriteFile("/tmp/templates/index.md", []byte("Test template"), 0644)

	_, err := generator.Generate(files, projectAggregated)

	if err != nil {
		t.Errorf("Did not expect an error but got: %v", err)
	} else {
		if _, err := os.Stat("/tmp/report.md"); os.IsNotExist(err) {
			t.Errorf("Report file was not created")
		} else {
			// cleanup
			os.Remove("/tmp/report.md")
		}
	}

	// cleanup
	os.RemoveAll("/tmp/templates")
}
