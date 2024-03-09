package Report

import (
	"fmt"
	"log"
	"os"

	"github.com/flosch/pongo2/v5"
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

	// Ensure destination folder exists
	err := v.EnsureFolder(v.ReportPath)
	if err != nil {
		return err
	}

	// Define loader in order to retrieve templates in the Report/Html/templates folder
	loader := pongo2.MustNewLocalFileSystemLoader("src/Report/Html/templates")
	pongo2.DefaultSet = pongo2.NewSet("src/Report/Html/templates", loader)

	// Compile the index.html template
	tpl, err := pongo2.DefaultSet.FromFile("index.html")
	if err != nil {
		log.Fatal(err)
	}
	// Render it, passing projectAggregated and files as context
	out, err := tpl.Execute(pongo2.Context{"projectAggregated": projectAggregated, "files": files})
	if err != nil {
		log.Fatal(err)
	}

	// Write the result to the file
	file, err := os.Create(fmt.Sprintf("%s/index.html", v.ReportPath))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString(out)

	return nil
}

func (v *ReportGenerator) EnsureFolder(path string) error {
	// check if the folder exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// create it
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
