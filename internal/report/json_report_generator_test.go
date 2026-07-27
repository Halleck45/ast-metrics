package report

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/analyzer"
	pb "github.com/ast-metrics/ast-metrics/pb"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJson(t *testing.T) {

	reportPath := filepath.Join(t.TempDir(), "report.json")
	generator := &JsonReportGenerator{ReportPath: reportPath}

	files := []*pb.File{
		{
			Path: "file1.php",
		},
	}
	projectAggregated := analyzer.ProjectAggregated{
		Combined: analyzer.Aggregated{
			ConcernedFiles: files,
		},
	}

	reports, err := generator.Generate(files, projectAggregated)

	// Check if the error is nil
	assert.Nil(t, err)
	assert.Equal(t, 1, len(reports))

	// Check if the file was created
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		t.Errorf("Report file was not created")
		return
	}

	// Cleanup
	defer os.Remove(reportPath)

	// Check if the file contains valid JSON, then load it
	// and check if it contains the expected keys

	// Load the file
	jsonFile, err := os.Open(reportPath)
	if err != nil {
		t.Errorf("Could not open the file")
		return
	}

	// Close the file
	defer jsonFile.Close()

	// Read the file
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		t.Errorf("Could not read the file")
		return
	}

	// Check if the file contains valid JSON
	var report map[string]interface{}
	err = json.Unmarshal(bytes, &report)
	if err != nil {
		t.Errorf("The file does not contain valid JSON")
		return
	}

	// Check if the file contains the list of concerned files (key concernedFiles)
	_, ok := report["concernedFiles"]
	if !ok {
		t.Errorf("The file does not contain the concernedFiles key")
		return
	}

	// Check if the file contains the list of concerned files (key concernedFiles)
	concernedFiles, ok := report["concernedFiles"].([]interface{})
	if !ok {
		t.Errorf("The concernedFiles key is not a list")
		return
	}

	// Check if the file contains the list of concerned files (key concernedFiles)
	if len(concernedFiles) != 1 {
		t.Errorf("The concernedFiles key does not contain the expected number of files")
		return
	}

	// Check if the file contains the list of concerned files (key concernedFiles)
	concernedFile, ok := concernedFiles[0].(map[string]interface{})
	if !ok {
		t.Errorf("The concernedFiles key does not contain a map")
		return
	}

	// Check if the file contains the list of concerned files (key concernedFiles)
	_, ok = concernedFile["path"]
	if !ok {
		t.Errorf("The concernedFiles key does not contain the path key")
		return
	}
}

// The three per-method maintainability averages come from three distinct
// aggregates; a copy-paste once made MIwoc mirror the comment weight.
func TestBuildReportMapsPerMethodMaintainability(t *testing.T) {
	generator := &JsonReportGenerator{}
	aggregated := analyzer.ProjectAggregated{
		Combined: analyzer.Aggregated{
			MaintainabilityPerMethod:                analyzer.AggregateResult{Avg: 80},
			MaintainabilityPerMethodWithoutComments: analyzer.AggregateResult{Avg: 60},
			MaintainabilityCommentWeightPerMethod:   analyzer.AggregateResult{Avg: 20},
		},
	}

	r := generator.buildReport(aggregated)

	if r.AverageMIPerMethod != 80 {
		t.Errorf("expected AverageMIPerMethod 80, got %v", r.AverageMIPerMethod)
	}
	if r.AverageMIwocPerMethod != 60 {
		t.Errorf("expected AverageMIwocPerMethod 60, got %v", r.AverageMIwocPerMethod)
	}
	if r.AverageMIcwPerMethod != 20 {
		t.Errorf("expected AverageMIcwPerMethod 20, got %v", r.AverageMIcwPerMethod)
	}
}
