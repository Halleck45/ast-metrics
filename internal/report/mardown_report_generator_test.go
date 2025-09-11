package report

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		reportPath  string
		expectError bool
	}{
		{
			name:        "Test with valid report path",
			reportPath:  "", // will be set to a temp file path in test body
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
			// Use a temp directory for portable paths
			reportPath := tt.reportPath
			if tt.name == "Test with valid report path" {
				dir, _ := ioutil.TempDir("", "report")
				defer os.RemoveAll(dir)
				reportPath = filepath.Join(dir, "report.md")
			}
			generator := &MarkdownReportGenerator{ReportPath: reportPath}
			files := []*pb.File{}
			projectAggregated := analyzer.ProjectAggregated{}

			_, err := generator.Generate(files, projectAggregated)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				} else {
					if reportPath != "" {
						if _, err := os.Stat(reportPath); os.IsNotExist(err) {
							t.Errorf("Report file was not created")
						} else {
							// cleanup
							os.Remove(reportPath)
						}
					}
				}
			}
		})
	}
}

func TestGenerateWithTemplateFiles(t *testing.T) {
	// Create a temporary directory for templates and output
	tmpDir, err := ioutil.TempDir("", "md-report")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	templatesDir := filepath.Join(tmpDir, "templates")
	_ = os.MkdirAll(templatesDir, 0o755)
	generator := &MarkdownReportGenerator{ReportPath: filepath.Join(tmpDir, "report.md")}
	files := []*pb.File{}
	projectAggregated := analyzer.ProjectAggregated{}

	// Create a temporary template file
	_ = ioutil.WriteFile(filepath.Join(templatesDir, "index.md"), []byte("Test template"), 0o644)

	reports, err := generator.Generate(files, projectAggregated)

	if err != nil {
		t.Errorf("Did not expect an error but got: %v", err)
	} else {
		if _, err := os.Stat(generator.ReportPath); os.IsNotExist(err) {
			t.Errorf("Report file was not created")
		} else {
			// cleanup
			os.Remove(generator.ReportPath)
		}
	}

	assert.Equal(t, 1, len(reports))
}
