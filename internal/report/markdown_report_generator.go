package report

import (
	"embed"
	"fmt"
	"os"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/flosch/pongo2/v5"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
)

var (
	//go:embed templates/*
	mdContent embed.FS
)

type MarkdownReportGenerator struct {
	// The path where the report will be generated
	ReportPath string
}

func NewMarkdownReportGenerator(reportPath string) Reporter {
	return &MarkdownReportGenerator{
		ReportPath: reportPath,
	}
}

func (v *MarkdownReportGenerator) Generate(files []*pb.File, projectAggregated analyzer.ProjectAggregated) ([]GeneratedReport, error) {

	// Ensure report is required
	if v.ReportPath == "" {
		return nil, nil
	}

	// copy the templates from embed, to temporary folder
	templateDir := fmt.Sprintf("%s/templates", os.TempDir())
	err := os.MkdirAll(templateDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	for _, file := range []string{"index.md"} {
		// read the file
		mdContent, err := mdContent.ReadFile(fmt.Sprintf("templates/markdown/%s", file))
		if err != nil {
			return nil, err
		}

		// write the file to temporary folder (/tmp)
		err = os.WriteFile(fmt.Sprintf("%s/%s", templateDir, file), mdContent, 0644)
		if err != nil {
			return nil, err
		}
	}

	// Define loader in order to retrieve templates in the Report/Html/templates folder
	loader := pongo2.MustNewLocalFileSystemLoader(templateDir)
	pongo2.DefaultSet = pongo2.NewSet(templateDir, loader)

	// Custom filters
	v.RegisterFilters()

	// Compile the index.md template
	tpl, err := pongo2.DefaultSet.FromFile("index.md")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// Render it, passing projectAggregated and files as context
	out, err := tpl.Execute(pongo2.Context{"projectAggregated": projectAggregated, "files": files})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Write the result to the file
	file, err := os.Create(v.ReportPath)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer file.Close()
	file.WriteString(out)

	// cleanup temporary folder
	err = os.RemoveAll(templateDir)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	reports := []GeneratedReport{
		{
			Path:        v.ReportPath,
			Type:        "file",
			Description: "The markdown report is useful for CI/CD pipelines, displaying the results in a human-readable format.",
			Icon:        "📄",
		},
	}
	return reports, nil

}

func (v *MarkdownReportGenerator) RegisterFilters() {

	pongo2.RegisterFilter("sortRisk", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the list to sort
		// create new empty list
		list := make([]*pb.File, 0)

		// append to the list when file contains at lease one class
		for _, file := range in.Interface().([]*pb.File) {
			if file.Stmts.StmtClass == nil {
				continue
			}

			list = append(list, file)
		}

		// sort the list
		sort.Slice(list, func(i, j int) bool {

			if list[i].Stmts.Analyze.Risk == nil {
				return false
			}

			if list[i].Stmts.StmtClass == nil {
				return true
			}

			if list[j].Stmts.StmtClass == nil {
				return true
			}

			class1 := list[i].Stmts.StmtClass[0]
			class2 := list[j].Stmts.StmtClass[0]

			if class1.Stmts.Analyze.Risk == nil {
				return false
			}
			if class2.Stmts.Analyze.Risk == nil {
				return true
			}

			return class1.Stmts.Analyze.Risk.Score > class2.Stmts.Analyze.Risk.Score
		})

		// keep only the first 10
		if len(list) > 10 {
			list = list[:10]
		}

		return pongo2.AsValue(list), nil
	})

	// filter to format number. Ex: 1234 -> 1 K
	if !pongo2.FilterExists("stringifyNumber") {
		pongo2.RegisterFilter("stringifyNumber", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
			// get the number to format
			number := in.Integer()

			// format it
			if number > 1000000 {
				return pongo2.AsValue(fmt.Sprintf("%.1f M", float64(number)/1000000)), nil
			} else if number > 1000 {
				return pongo2.AsValue(fmt.Sprintf("%.1f K", float64(number)/1000)), nil
			}

			return pongo2.AsValue(number), nil
		})
	}
}
