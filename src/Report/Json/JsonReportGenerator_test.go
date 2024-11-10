package Report

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJson(t *testing.T) {

	reportPath := "/tmp/report.json"
	generator := &JsonReportGenerator{ReportPath: reportPath}

	files := []*pb.File{
		{
			Path: "file1.php",
		},
	}
	projectAggregated := Analyzer.ProjectAggregated{
		Combined: Analyzer.Aggregated{
			ConcernedFiles: files,
		},
	}

	_, err := generator.Generate(files, projectAggregated)

	// Check if the error is nil
	assert.Nil(t, err)

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
