package report

import (
	"embed"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/flosch/pongo2/v5"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/halleck45/ast-metrics/internal/ui"
)

var (
	//go:embed templates/*
	htmlContent embed.FS
)

type HtmlReportGenerator struct {
	// The path where the report will be generated
	ReportPath string
}

func NewHtmlReportGenerator(reportPath string) Reporter {
	return &HtmlReportGenerator{
		ReportPath: reportPath,
	}
}

func (v *HtmlReportGenerator) Generate(files []*pb.File, projectAggregated analyzer.ProjectAggregated) ([]GeneratedReport, error) {

	// Ensure report is required
	if v.ReportPath == "" {
		return nil, nil
	}

	// Ensure destination folder exists
	err := v.EnsureFolder(v.ReportPath)
	if err != nil {
		return nil, err
	}

	// copy the templates from embed, to temporary folder
	baseTemplateDir := fmt.Sprintf("%s/templates", os.TempDir())
	err = os.MkdirAll(baseTemplateDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	// ensure partials subfolder exists under base
	partialsDir := fmt.Sprintf("%s/partials", baseTemplateDir)
	if err := os.MkdirAll(partialsDir, os.ModePerm); err != nil {
		return nil, err
	}

	for _, file := range []string{
		"index.html",
		"layout.html",
		"risks.html",
		"compare.html",
		"componentChartRadiusBar.html",
		"componentTableRisks.html",
		"componentTableCompareBranch.html",
		"componentChartRadiusBarMaintainability.html",
		"componentChartRadiusBarLoc.html",
		"componentChartRadiusBarComplexity.html",
		"componentChartRadiusBarInstability.html",
		"componentChartRadiusBarEfferent.html",
		"componentChartRadiusBarAfferent.html",
		"componentDependencyDiagram.html",
		"componentComparaisonBadge.html",
		"componentComparaisonOperator.html",
		"communities.html",
		"partials/suggestions.html",
	} {
		// read the file
		bytes, err := htmlContent.ReadFile(fmt.Sprintf("templates/html/%s", file))
		if err != nil {
			return nil, err
		}

		// write the file to temporary folder (/tmp) preserving subpaths under baseTemplateDir
		outPath := fmt.Sprintf("%s/%s", baseTemplateDir, file)
		// ensure parent directory exists (e.g., for partials)
		if dir := outPath[:len(outPath)-len(file)]; dir != "" {
			if err := os.MkdirAll(strings.TrimRight(dir, "/"), os.ModePerm); err != nil {
				return nil, err
			}
		}
		err = os.WriteFile(outPath, bytes, 0644)
		if err != nil {
			return nil, err
		}
	}

	// Define loader rooted at the base template directory
	loader := pongo2.MustNewLocalFileSystemLoader(baseTemplateDir)
	pongo2.DefaultSet = pongo2.NewSet(baseTemplateDir, loader)

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

	// Comparaison with another branch
	v.GenerateLanguagePage("compare.html", "All", projectAggregated.Combined, files, projectAggregated)
	// by language overview
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("compare.html", language, currentView, files, projectAggregated)
	}

	// Communities page
	v.GenerateLanguagePage("communities.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("communities.html", language, currentView, files, projectAggregated)
	}

	// copy images
	err = v.EnsureFolder(fmt.Sprintf("%s/images", v.ReportPath))
	if err != nil {
		return nil, err
	}

	// copy each image
	for _, file := range []string{
		"help-community.png",
	} {
		// read the file
		htmlContent, err := htmlContent.ReadFile(fmt.Sprintf("templates/html/images/%s", file))
		if err != nil {
			return nil, err
		}

		// write the file to temporary folder
		err = os.WriteFile(fmt.Sprintf("%s/images/%s", v.ReportPath, file), htmlContent, 0644)
		if err != nil {
			return nil, err
		}
	}

	// cleanup temporary folder
	err = os.RemoveAll(baseTemplateDir)
	if err != nil {
		return nil, err
	}

	reports := []GeneratedReport{
		{
			Path:        v.ReportPath,
			Type:        "directory",
			Description: "The HTML reports allow you to visualize the metrics of your project in a web browser.",
			Icon:        "📊",
		},
	}

	return reports, nil
}

func (v *HtmlReportGenerator) GenerateLanguagePage(template string, language string, currentView analyzer.Aggregated, files []*pb.File, projectAggregated analyzer.ProjectAggregated) error {

	// Compile the index.html template
	tpl, err := pongo2.DefaultSet.FromFile(template)
	if err != nil {
		log.Error(err)
		return err
	}
	// Render it, passing projectAggregated and files as context
	datetime := time.Now().Format("2006-01-02 15:04")
	out, err := tpl.Execute(pongo2.Context{"datetime": datetime, "page": template, "currentLanguage": language, "currentView": currentView, "projectAggregated": projectAggregated, "files": files})
	if err != nil {
		log.Error(err)
		return err
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

			classes := engine.GetClassesInFile(file)

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

	pongo2.RegisterFilter("jsonForChartDependency", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// create json for chart dependency, like:
		// [ { "source": "A", "target": "B", "value": 1 }, { "source": "A", "target": "C", "value": 1 } ]

		// receive map[string]map[string]int in input
		relations := in.Interface().(map[string]map[string]int)
		json := "["
		for source, targets := range relations {
			for target, value := range targets {
				json += fmt.Sprintf(
					"{ \"source\": \"%s\", \"target\": \"%s\", \"value\": %d },",
					strings.ReplaceAll(source, "\\", "\\\\"),
					strings.ReplaceAll(target, "\\", "\\\\"),
					value,
				)
			}
		}
		json = json[:len(json)-1] + "]"

		if json == "]" {
			// occurs when no relations are found
			json = "[]"
		}

		return pongo2.AsSafeValue(json), nil
	})

	pongo2.RegisterFilter("sortRisk", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {

		rowsToKeep := 10
		if param.Integer() > 0 {
			rowsToKeep = param.Integer()
		}

		// Sort by risk of file
		files := in.Interface().([]*pb.File)
		sort.Slice(files, func(i, j int) bool {
			if files[i].Stmts == nil && files[j].Stmts == nil || files[i].Stmts.Analyze == nil || files[j].Stmts.Analyze == nil {
				return false
			}

			if files[i].Stmts.Analyze.Risk == nil && files[j].Stmts.Analyze.Risk == nil {
				return false
			}

			if files[i].Stmts.Analyze.Risk == nil {
				return false
			}

			if files[j].Stmts.Analyze.Risk == nil {
				return true
			}

			return files[i].Stmts.Analyze.Risk.Score > files[j].Stmts.Analyze.Risk.Score
		})

		// keep only the first n
		if len(files) > rowsToKeep {
			files = files[:rowsToKeep]
		}

		return pongo2.AsValue(files), nil
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
		aggregated := in.Interface().(analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := ui.ComponentBarchartCyclomaticByMethodRepartition{
			Aggregated: aggregated,
			Files:      files,
		}
		return pongo2.AsSafeValue(comp.AsHtml()), nil
	})

	// filter barchartCyclomaticByMethodRepartition
	pongo2.RegisterFilter("barchartCyclomaticByMethodRepartition", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := ui.ComponentBarchartCyclomaticByMethodRepartition{
			Aggregated: aggregated,
			Files:      files,
		}
		return pongo2.AsSafeValue(comp.AsHtml()), nil
	})

	// filter barchartMaintainabilityIndexRepartition
	pongo2.RegisterFilter("barchartMaintainabilityIndexRepartition", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := ui.ComponentBarchartMaintainabilityIndexRepartition{
			Aggregated: aggregated,
			Files:      files,
		}

		return pongo2.AsSafeValue(comp.AsHtml()), nil
	})

	// filter barchartLocPerMethodRepartition
	pongo2.RegisterFilter("barchartLocPerMethodRepartition", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := ui.ComponentBarchartLocByMethodRepartition{
			Aggregated: aggregated,
			Files:      files,
		}
		return pongo2.AsSafeValue(comp.AsHtml()), nil
	})

	// filter lineChartGitActivity
	pongo2.RegisterFilter("lineChartGitActivity", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := ui.ComponentLineChartGitActivity{
			Aggregated: aggregated,
			Files:      files,
		}
		return pongo2.AsSafeValue(comp.AsHtml()), nil
	})

	// filter convertOneFileToCollection
	pongo2.RegisterFilter("convertOneFileToCollection", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		file := in.Interface().(*pb.File)
		return pongo2.AsValue([]*pb.File{file}), nil
	})

	// filter : has class or uis procedural script
	pongo2.RegisterFilter("fileHasClasses", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		file := in.Interface().(*pb.File)
		return pongo2.AsValue(len(engine.GetClassesInFile(file)) > 0), nil
	})

	// filter : has class or uis procedural script
	pongo2.RegisterFilter("toCollectionOfParsableComponents", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		file := in.Interface().(*pb.File)

		if len(engine.GetClassesInFile(file)) > 0 {
			return pongo2.AsValue(engine.GetClassesInFile(file)), nil
		}

		collection := make([]*pb.StmtFunction, 0)
		collection = append(collection, file.Stmts.StmtFunction...)

		return pongo2.AsValue(collection), nil
	})
}
