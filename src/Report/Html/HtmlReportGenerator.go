package Report

import (
	"embed"
	"fmt"
	"os"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/flosch/pongo2/v5"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	"github.com/halleck45/ast-metrics/src/Cli"
	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

var (
	//go:embed templates/*
	content embed.FS
)

type HtmlReportGenerator struct {
	// The path where the report will be generated
	ReportPath string
}

func NewHtmlReportGenerator(reportPath string) *HtmlReportGenerator {
	return &HtmlReportGenerator{
		ReportPath: reportPath,
	}
}

func (v *HtmlReportGenerator) Generate(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) error {

	// Ensure report is required
	if v.ReportPath == "" {
		return nil
	}

	// Ensure destination folder exists
	err := v.EnsureFolder(v.ReportPath)
	if err != nil {
		return err
	}

	// copy the templates from embed, to temporary folder
	templateDir := fmt.Sprintf("%s/templates", os.TempDir())
	err = os.MkdirAll(templateDir, os.ModePerm)
	if err != nil {
		return err
	}

	for _, file := range []string{"index.html", "layout.html", "risks.html", "componentTableRisks.html"} {
		// read the file
		content, err := content.ReadFile(fmt.Sprintf("templates/%s", file))
		if err != nil {
			return err
		}

		// write the file to temporary folder (/tmp)
		err = os.WriteFile(fmt.Sprintf("%s/%s", templateDir, file), content, 0644)
		if err != nil {
			return err
		}
	}

	// Define loader in order to retrieve templates in the Report/Html/templates folder
	loader := pongo2.MustNewLocalFileSystemLoader(templateDir)
	pongo2.DefaultSet = pongo2.NewSet(templateDir, loader)

	// Custom filters
	v.RegisterFilters()

	// Overview
	v.GenerateLanguagePage("index.html", "All", projectAggregated.Combined, files, projectAggregated)
	// by language overview
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("index.html", language, currentView, files, projectAggregated)
	}

	// Risks
	v.GenerateLanguagePage("risks.html", "All", projectAggregated.Combined, files, projectAggregated)
	// by language overview
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("risks.html", language, currentView, files, projectAggregated)
	}

	// cleanup temporary folder
	err = os.RemoveAll(templateDir)
	if err != nil {
		return err
	}

	return nil
}

func (v *HtmlReportGenerator) GenerateLanguagePage(template string, language string, currentView Analyzer.Aggregated, files []*pb.File, projectAggregated Analyzer.ProjectAggregated) error {

	// Compile the index.html template
	tpl, err := pongo2.DefaultSet.FromFile(template)
	if err != nil {
		log.Error(err)
	}
	// Render it, passing projectAggregated and files as context
	out, err := tpl.Execute(pongo2.Context{"page": template, "currentLanguage": language, "currentView": currentView, "projectAggregated": projectAggregated, "files": files})
	if err != nil {
		log.Error(err)
	}

	// Write the result to the file
	pageSuffix := ""
	if language != "All" {
		pageSuffix = fmt.Sprintf("_%s", language)
	}
	// prefix is template without the .html part
	pagePrefix := template[:len(template)-5]
	file, err := os.Create(fmt.Sprintf("%s/%s%s.html", v.ReportPath, pagePrefix, pageSuffix))
	if err != nil {
		log.Error(err)
	}
	defer file.Close()
	file.WriteString(out)

	return nil
}

func (v *HtmlReportGenerator) EnsureFolder(path string) error {
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

func (v *HtmlReportGenerator) RegisterFilters() {

	pongo2.RegisterFilter("sortMaintainabilityIndex", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the list to sort
		// create new empty list
		list := make([]*pb.StmtClass, 0)

		// append to the list when file contians at lease one class
		for _, file := range in.Interface().([]*pb.File) {
			if len(file.Stmts.StmtClass) == 0 {
				continue
			}

			classes := Engine.GetClassesInFile(file)

			for _, class := range classes {
				if class.Stmts.Analyze.Maintainability == nil {
					continue
				}

				if *class.Stmts.Analyze.Maintainability.MaintainabilityIndex < 1 {
					continue
				}

				if *class.Stmts.Analyze.Maintainability.MaintainabilityIndex > 65 {
					continue
				}

				list = append(list, class)
			}
		}

		// sort the list, manually
		sort.Slice(list, func(i, j int) bool {
			if list[i].Stmts.Analyze.Maintainability == nil {
				return false
			}
			if list[j].Stmts.Analyze.Maintainability == nil {
				return true
			}

			// get first class in file
			class1 := list[i]
			class2 := list[j]

			return *class1.Stmts.Analyze.Maintainability.MaintainabilityIndex < *class2.Stmts.Analyze.Maintainability.MaintainabilityIndex
		})

		// keep only the first 10
		if len(list) > 10 {
			list = list[:10]
		}

		return pongo2.AsValue(list), nil
	})

	pongo2.RegisterFilter("sortRisk", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the list to sort
		// create new empty list
		list := make([]*pb.File, 0)

		rowsToKeep := 10
		if param.Integer() > 0 {
			rowsToKeep = param.Integer()
		}

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
		if len(list) > rowsToKeep {
			list = list[:rowsToKeep]
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

	// filter that Return new Cli.NewComponentBarchartCyclomaticByMethodRepartition(aggregated, files)
	pongo2.RegisterFilter("barchartCyclomaticByMethodRepartition", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(Analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := Cli.NewComponentBarchartCyclomaticByMethodRepartition(aggregated, files)
		return pongo2.AsSafeValue(comp.RenderHTML()), nil
	})

	// filter barchartCyclomaticByMethodRepartition
	pongo2.RegisterFilter("barchartCyclomaticByMethodRepartition", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(Analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := Cli.NewComponentBarchartCyclomaticByMethodRepartition(aggregated, files)
		return pongo2.AsSafeValue(comp.RenderHTML()), nil
	})

	// filter barchartMaintainabilityIndexRepartition
	pongo2.RegisterFilter("barchartMaintainabilityIndexRepartition", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(Analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := Cli.NewComponentBarchartMaintainabilityIndexRepartition(aggregated, files)
		return pongo2.AsSafeValue(comp.RenderHTML()), nil
	})

	// filter barchartLocPerMethodRepartition
	pongo2.RegisterFilter("barchartLocPerMethodRepartition", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(Analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := Cli.NewComponentBarchartLocByMethodRepartition(aggregated, files)
		return pongo2.AsSafeValue(comp.RenderHTML()), nil
	})

	// filter lineChartGitActivity
	pongo2.RegisterFilter("lineChartGitActivity", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(Analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := Cli.NewComponentLineChartGitActivity(aggregated, files)
		return pongo2.AsSafeValue(comp.RenderHTML()), nil
	})
}
