package report

import (
	"embed"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"

	"github.com/flosch/pongo2/v5"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/analyzer/classifier"
	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/halleck45/ast-metrics/internal/ui"
	pb "github.com/halleck45/ast-metrics/pb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
		"explorer.html",
		"linters.html",
		"classification.html",
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
		"busfactor.html",
		"architecture-roles.html",
		"layer-violations.html",
		"ambiguity-zones.html",
		"role-flow-graph.html",
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
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("risks.html", language, currentView, files, projectAggregated)
	}

	// Explorer
	v.GenerateLanguagePage("explorer.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("explorer.html", language, currentView, files, projectAggregated)
	}

	// Comparaison with another branch
	v.GenerateLanguagePage("compare.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("compare.html", language, currentView, files, projectAggregated)
	}

	// Communities page
	v.GenerateLanguagePage("communities.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("communities.html", language, currentView, files, projectAggregated)
	}

	// Linters page
	v.GenerateLanguagePage("linters.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("linters.html", language, currentView, files, projectAggregated)
	}

	// Bus Factor page
	v.GenerateLanguagePage("busfactor.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("busfactor.html", language, currentView, files, projectAggregated)
	}

	// Classification page
	v.GenerateLanguagePage("classification.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("classification.html", language, currentView, files, projectAggregated)
	}

	// Architecture Roles page
	v.GenerateLanguagePage("architecture-roles.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("architecture-roles.html", language, currentView, files, projectAggregated)
	}

	// Layer Violations page
	v.GenerateLanguagePage("layer-violations.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("layer-violations.html", language, currentView, files, projectAggregated)
	}

	// Ambiguity Zones page
	v.GenerateLanguagePage("ambiguity-zones.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("ambiguity-zones.html", language, currentView, files, projectAggregated)
	}

	// Role Flow Graph page
	v.GenerateLanguagePage("role-flow-graph.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("role-flow-graph.html", language, currentView, files, projectAggregated)
	}

	// copy images
	err = v.EnsureFolder(fmt.Sprintf("%s/images", v.ReportPath))
	if err != nil {
		return nil, err
	}

	// copy each image
	for _, file := range []string{
		"help-community.png",
		"logo-ast-metrics-small.png",
		"icon-ai.webp",
		"icon-classifier.webp",
		"icon-fingerprint.webp",
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
			Type:        "html",
			Description: "The HTML reports allow you to visualize the metrics of your project in a web browser.",
			Icon:        "📊",
		},
	}

	return reports, nil
}

type riskItemForTpl struct {
	ID       string
	Title    string
	Severity float64
	Details  string
}

// New Pruned JSON using protojson
func buildFilesJSONPruned(files []*pb.File, language string) string {
	pruned := make([]*pb.File, 0, len(files))
	for _, f := range files {
		if language != "All" && f.GetProgrammingLanguage() != language {
			continue
		}
		cf := proto.Clone(f).(*pb.File)
		pruneFile(cf)
		pruned = append(pruned, cf)
	}
	mo := protojson.MarshalOptions{EmitUnpopulated: false, UseEnumNumbers: false, Indent: ""}
	var b strings.Builder
	b.WriteString("[")
	for i, f := range pruned {
		if i > 0 {
			b.WriteString(",")
		}
		if data, err := mo.Marshal(f); err == nil {
			b.Write(data)
		} else {
			b.WriteString("{}")
		}
	}
	b.WriteString("]")
	return b.String()
}

func pruneFile(f *pb.File) {
	if f.Stmts == nil {
		return
	}
	s := f.Stmts

	s.StmtFunction = nil
	s.StmtInterface = nil
	s.StmtTrait = nil
	s.StmtUse = nil
	s.StmtNamespace = nil
	s.StmtDecisionIf = nil
	s.StmtDecisionElseIf = nil
	s.StmtDecisionElse = nil
	s.StmtDecisionCase = nil
	s.StmtLoop = nil
	s.StmtDecisionSwitch = nil
	s.StmtExternalDependencies = nil
	for _, c := range s.StmtClass {
		pruneClass(c)
	}
}

func pruneClass(c *pb.StmtClass) {
	c.Location = nil
	c.Comments = nil
	c.Operators = nil
	c.Operands = nil
	c.Extends = nil
	c.Implements = nil
	c.Uses = nil
	c.LinesOfCode = nil
	if c.Stmts != nil {
		for _, m := range c.Stmts.StmtFunction {
			pruneFunction(m)
		}
		c.Stmts.StmtClass = nil
		c.Stmts.StmtInterface = nil
		c.Stmts.StmtTrait = nil
		c.Stmts.StmtUse = nil
		c.Stmts.StmtNamespace = nil
		c.Stmts.StmtDecisionIf = nil
		c.Stmts.StmtDecisionElseIf = nil
		c.Stmts.StmtDecisionElse = nil
		c.Stmts.StmtDecisionCase = nil
		c.Stmts.StmtLoop = nil
		c.Stmts.StmtDecisionSwitch = nil
		c.Stmts.StmtExternalDependencies = nil
	}
}

func pruneFunction(m *pb.StmtFunction) {
	m.Location = nil
	m.Comments = nil
	m.Operators = nil
	m.Operands = nil
	m.MethodCalls = nil
	m.Parameters = nil
	m.Externals = nil
	m.LinesOfCode = nil
	if m.Stmts != nil {
		if m.Stmts.Analyze != nil {
			// keep Complexity only
			m.Stmts.Analyze.Volume = nil
			m.Stmts.Analyze.Maintainability = nil
			m.Stmts.Analyze.Risk = nil
			m.Stmts.Analyze.Coupling = nil
			m.Stmts.Analyze.ClassCohesion = nil
		}
		m.Stmts.StmtClass = nil
		m.Stmts.StmtFunction = nil
		m.Stmts.StmtInterface = nil
		m.Stmts.StmtTrait = nil
		m.Stmts.StmtUse = nil
		m.Stmts.StmtNamespace = nil
		m.Stmts.StmtDecisionIf = nil
		m.Stmts.StmtDecisionElseIf = nil
		m.Stmts.StmtDecisionElse = nil
		m.Stmts.StmtDecisionCase = nil
		m.Stmts.StmtLoop = nil
		m.Stmts.StmtDecisionSwitch = nil
		m.Stmts.StmtExternalDependencies = nil
	}
}

func buildRisksJSON(risksByPath map[string][]riskItemForTpl) string {
	b := strings.Builder{}
	b.WriteString("{")
	first := true
	for p, items := range risksByPath {
		if !first {
			b.WriteString(",")
		} else {
			first = false
		}
		pp := strings.ReplaceAll(strings.ReplaceAll(p, "\\", "\\\\"), "\"", "\\\"")
		b.WriteString("\"")
		b.WriteString(pp)
		b.WriteString("\":[")
		for i, r := range items {
			if i > 0 {
				b.WriteString(",")
			}
			d := strings.ReplaceAll(strings.ReplaceAll(r.Details, "\\", "\\\\"), "\"", "\\\"")
			t := strings.ReplaceAll(strings.ReplaceAll(r.Title, "\\", "\\\\"), "\"", "\\\"")
			b.WriteString("{\"id\":\"")
			b.WriteString(r.ID)
			b.WriteString("\",\"title\":\"")
			b.WriteString(t)
			b.WriteString("\",\"severity\":")
			b.WriteString(fmt.Sprintf("%g", r.Severity))
			b.WriteString(",\"details\":\"")
			b.WriteString(d)
			b.WriteString("\"}")
		}
		b.WriteString("]")
	}
	b.WriteString("}")
	return b.String()
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
	// build risks map for explorer and other pages that may use it
	risksByPath := map[string][]riskItemForTpl{}
	ra := analyzer.NewRiskAnalyzer()
	for _, f := range files {
		if language != "All" && f.ProgrammingLanguage != language {
			continue
		}
		items := ra.DetectFileRisks(f)
		if len(items) > 0 {
			converted := make([]riskItemForTpl, 0, len(items))
			for _, it := range items {
				converted = append(converted, riskItemForTpl{ID: it.ID, Title: it.Title, Severity: it.Severity, Details: it.Details})
			}
			risksByPath[f.Path] = converted
		}
	}

	filesJSON := buildFilesJSONPruned(files, language)
	risksJSON := buildRisksJSON(risksByPath)
	out, err := tpl.Execute(pongo2.Context{"datetime": datetime, "page": template, "currentLanguage": language, "currentView": currentView, "projectAggregated": projectAggregated, "files": files, "risksByPath": risksByPath, "filesJSON": filesJSON, "risksJSON": risksJSON, "classificationFamilies": classifier.ClassificationFamilies})
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

	// include_all_files_as_json: returns pre-rendered JSON for files of current page (by language)
	pongo2.RegisterFilter("include_all_files_as_json", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// in is []*pb.File; we will emit minimal JSON for Explorer page needs.
		files := in.Interface().([]*pb.File)
		// Build JSON manually to avoid bringing an extra dependency and to keep control on fields.
		// We purposefully escape quotes and backslashes for safety in HTML context when used inside <script type="application/json">.
		b := strings.Builder{}
		b.WriteString("[")
		first := true
		for _, f := range files {
			if !first {
				b.WriteString(",")
			} else {
				first = false
			}
			// basic fields with escaping
			path := strings.ReplaceAll(strings.ReplaceAll(f.Path, "\\", "\\\\"), "\"", "\\\"")
			shortPath := strings.ReplaceAll(strings.ReplaceAll(f.ShortPath, "\\", "\\\\"), "\"", "\\\"")
			lang := strings.ReplaceAll(strings.ReplaceAll(f.ProgrammingLanguage, "\\", "\\\\"), "\"", "\\\"")
			b.WriteString("{\"path\":\"")
			b.WriteString(path)
			b.WriteString("\",\"shortPath\":\"")
			b.WriteString(shortPath)
			b.WriteString("\",\"programmingLanguage\":\"")
			b.WriteString(lang)
			b.WriteString("\"")
			// risk score
			if f.Stmts != nil && f.Stmts.Analyze != nil && f.Stmts.Analyze.Risk != nil {
				b.WriteString(",\"risk\":{")
				b.WriteString("\"score\":")
				b.WriteString(fmt.Sprintf("%g", f.Stmts.Analyze.Risk.Score))
				b.WriteString("}")
			}
			// classes minimal metrics
			classes := engine.GetClassesInFile(f)
			b.WriteString(",\"classes\":[")
			cFirst := true
			for _, c := range classes {
				if !cFirst {
					b.WriteString(",")
				} else {
					cFirst = false
				}
				name := ""
				if c.Name != nil {
					name = c.Name.GetQualified()
				}
				name = strings.ReplaceAll(strings.ReplaceAll(name, "\\", "\\\\"), "\"", "\\\"")
				b.WriteString("{\"name\":\"")
				b.WriteString(name)
				b.WriteString("\"")
				// MI
				if c.Stmts != nil && c.Stmts.Analyze != nil && c.Stmts.Analyze.Maintainability != nil && c.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil {
					b.WriteString(",\"mi\":")
					b.WriteString(fmt.Sprintf("%g", *c.Stmts.Analyze.Maintainability.MaintainabilityIndex))
				}
				// efferent
				if c.Stmts != nil && c.Stmts.Analyze != nil && c.Stmts.Analyze.Coupling != nil {
					b.WriteString(",\"efferent\":")
					b.WriteString(fmt.Sprintf("%d", c.Stmts.Analyze.Coupling.Efferent))
				}
				// lcom4
				if c.Stmts != nil && c.Stmts.Analyze != nil && c.Stmts.Analyze.ClassCohesion != nil && c.Stmts.Analyze.ClassCohesion.Lcom4 != nil {
					b.WriteString(",\"lcom4\":")
					b.WriteString(fmt.Sprintf("%d", *c.Stmts.Analyze.ClassCohesion.Lcom4))
				}
				b.WriteString("}")
			}
			b.WriteString("]}")
		}
		b.WriteString("]")
		return pongo2.AsSafeValue(b.String()), nil
	})

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
			if files[i].Stmts == nil && files[j].Stmts == nil || files[i].Stmts == nil || files[j].Stmts == nil || files[i].Stmts.Analyze == nil || files[j].Stmts.Analyze == nil {
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

	// selectTopRiskEntries flattens files into class/file rows and caps the total number of rows
	pongo2.RegisterFilter("selectTopRiskEntries", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		rowsToKeep := 10
		if param != nil && param.Integer() > 0 {
			rowsToKeep = param.Integer()
		}

		// defensive: empty input
		if in == nil || in.IsNil() {
			return pongo2.AsValue([]interface{}{}), nil
		}

		// Sort by risk of file first (reuse logic)
		files := in.Interface().([]*pb.File)
		sort.Slice(files, func(i, j int) bool {
			if files[i] == nil || files[j] == nil || files[i].Stmts == nil || files[j].Stmts == nil || files[i].Stmts.Analyze == nil || files[j].Stmts.Analyze == nil {
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

		type RiskEntry struct {
			File  *pb.File
			Class *pb.StmtClass
			Name  string
		}

		entries := make([]*RiskEntry, 0, rowsToKeep)

		add := func(file *pb.File, class *pb.StmtClass, name string) bool {
			entries = append(entries, &RiskEntry{File: file, Class: class, Name: name})
			return len(entries) >= rowsToKeep
		}

		for _, file := range files {
			if file == nil || file.Stmts == nil {
				continue
			}
			// if no classes, treat file as a single row
			if len(file.Stmts.StmtClass) == 0 {
				name := file.Path
				if name == "" {
					name = "(unknown)"
				}
				// Create a dummy class holder so template fields (class.Stmts...) remain available
				dummy := &pb.StmtClass{Stmts: file.Stmts}
				if add(file, dummy, name) {
					break
				}
				continue
			}
			// else, iterate classes
			for _, class := range file.Stmts.StmtClass {
				if class == nil {
					continue
				}
				name := ""
				if class.Name != nil {
					name = class.Name.Qualified
				}
				if name == "" {
					name = file.Path
				}
				if add(file, class, name) {
					break
				}
			}
			if len(entries) >= rowsToKeep {
				break
			}
		}

		return pongo2.AsValue(entries), nil
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

	// filter barchartLcomRepartition
	pongo2.RegisterFilter("barchartLcomRepartition", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		// get the aggregated and files
		aggregated := in.Interface().(analyzer.Aggregated)
		files := aggregated.ConcernedFiles

		// create the component
		comp := ui.ComponentBarchartLcomRepartition{
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

	// filter groupByLabel
	pongo2.RegisterFilter("groupByLabel", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		predictions, ok := in.Interface().([]classifier.ClassPrediction)
		if !ok {
			return pongo2.AsValue(map[string][]classifier.ClassPrediction{}), nil
		}

		grouped := make(map[string][]classifier.ClassPrediction)
		for _, p := range predictions {
			if len(p.Predictions) > 0 {
				label := p.Predictions[0].Label
				grouped[label] = append(grouped[label], p)
			} else {
				grouped["Unknown"] = append(grouped["Unknown"], p)
			}
		}
		return pongo2.AsValue(grouped), nil
	})

	// filter getLabelDescription: returns the description for a classification label
	pongo2.RegisterFilter("getLabelDescription", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		label := in.String()
		description := classifier.GetDescription(label)
		return pongo2.AsValue(description), nil
	})

	// filter groupByFamilyAndLabel: groups predictions by family first, then by label
	pongo2.RegisterFilter("groupByFamilyAndLabel", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		predictions, ok := in.Interface().([]classifier.ClassPrediction)
		if !ok {
			return pongo2.AsValue(classifier.FamilyGroupedPredictions{}), nil
		}
		grouped := classifier.GroupByFamilyAndLabel(predictions)
		return pongo2.AsValue(grouped), nil
	})

	// filter capitalizeFirst: capitalizes the first letter of a string
	pongo2.RegisterFilter("capitalizeFirst", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		s := in.String()
		if len(s) == 0 {
			return pongo2.AsValue(""), nil
		}
		return pongo2.AsValue(strings.ToUpper(s[:1]) + s[1:]), nil
	})

	// filter getMapValue: gets a value from a map using a key
	pongo2.RegisterFilter("getMapValue", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		key := param.String()
		// Try different map types
		switch m := in.Interface().(type) {
		case map[string]interface{}:
			if val, exists := m[key]; exists {
				return pongo2.AsValue(val), nil
			}
		case classifier.FamilyGroupedPredictions:
			if val, exists := m[key]; exists {
				return pongo2.AsValue(val), nil
			}
		}
		return pongo2.AsValue(nil), nil
	})

	// filter countFamilyItems: counts total items in a family data map
	pongo2.RegisterFilter("countFamilyItems", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		familyData, ok := in.Interface().(map[string][]classifier.ClassPrediction)
		if !ok {
			return pongo2.AsValue(0), nil
		}
		count := 0
		for _, items := range familyData {
			count += len(items)
		}
		return pongo2.AsValue(count), nil
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

	// filter contributorInitials: extracts initials from a name (e.g., "John Doe" -> "JD")
	pongo2.RegisterFilter("contributorInitials", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		name := in.String()
		if name == "" {
			return pongo2.AsValue("?"), nil
		}

		// Split by common separators and get first letter of each word
		parts := strings.Fields(name)
		initials := strings.Builder{}
		for _, part := range parts {
			if len(part) > 0 {
				// Get first letter (handling unicode)
				for _, r := range part {
					if unicode.IsLetter(r) {
						initials.WriteRune(unicode.ToUpper(r))
						break
					}
				}
			}
		}

		result := initials.String()
		if result == "" {
			// Fallback: use first character
			for _, r := range name {
				if unicode.IsPrint(r) {
					result = strings.ToUpper(string(r))
					break
				}
			}
			if result == "" {
				result = "?"
			}
		}

		// Limit to 2-3 characters max
		if len([]rune(result)) > 3 {
			result = string([]rune(result)[:3])
		}

		return pongo2.AsValue(result), nil
	})

	// filter contributorColor: generates a consistent color based on name hash
	pongo2.RegisterFilter("contributorColor", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		name := in.String()
		if name == "" {
			return pongo2.AsValue("#9ca3af"), nil // gray fallback
		}

		// Generate hash from name
		h := fnv.New32a()
		h.Write([]byte(name))
		hash := h.Sum32()

		// Use a palette of pleasant colors
		colors := []string{
			"#3b82f6", // blue
			"#8b5cf6", // purple
			"#ec4899", // pink
			"#f59e0b", // amber
			"#10b981", // emerald
			"#06b6d4", // cyan
			"#ef4444", // red
			"#14b8a6", // teal
			"#f97316", // orange
			"#6366f1", // indigo
			"#84cc16", // lime
			"#a855f7", // violet
		}

		colorIndex := int(hash) % len(colors)
		return pongo2.AsValue(colors[colorIndex]), nil
	})

	// filter getRoleCategory: extracts category from a role label
	pongo2.RegisterFilter("getRoleCategory", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		label := in.String()
		parts := strings.Split(label, ":")
		if len(parts) >= 2 {
			return pongo2.AsValue(parts[1]), nil
		}
		return pongo2.AsValue("unknown"), nil
	})

	// filter getRoleShortName: extracts short name from a role label
	pongo2.RegisterFilter("getRoleShortName", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		label := in.String()
		parts := strings.Split(label, ":")
		if len(parts) >= 3 {
			return pongo2.AsValue(parts[2]), nil
		}
		if len(parts) >= 2 {
			return pongo2.AsValue(parts[1]), nil
		}
		return pongo2.AsValue(label), nil
	})

	// filter getUniqueRoles: extracts unique roles from role flows
	pongo2.RegisterFilter("getUniqueRoles", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		flows, ok := in.Interface().([]analyzer.RoleFlow)
		if !ok {
			return pongo2.AsValue([]string{}), nil
		}
		roleSet := make(map[string]bool)
		for _, flow := range flows {
			roleSet[flow.FromRole] = true
			roleSet[flow.ToRole] = true
		}
		roles := make([]string, 0, len(roleSet))
		for role := range roleSet {
			roles = append(roles, role)
		}
		sort.Strings(roles)
		return pongo2.AsValue(roles), nil
	})

	// filter escapejs: escapes a string for safe use in JavaScript
	pongo2.RegisterFilter("escapejs", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		str := in.String()
		// Escape backslashes first (important!)
		str = strings.ReplaceAll(str, "\\", "\\\\")
		// Escape quotes
		str = strings.ReplaceAll(str, "\"", "\\\"")
		str = strings.ReplaceAll(str, "'", "\\'")
		// Escape newlines
		str = strings.ReplaceAll(str, "\n", "\\n")
		str = strings.ReplaceAll(str, "\r", "\\r")
		// Escape tabs
		str = strings.ReplaceAll(str, "\t", "\\t")
		return pongo2.AsValue(str), nil
	})
}
