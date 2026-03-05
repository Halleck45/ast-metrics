package report

import (
	"embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path"
	"sort"
	"strings"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"

	"github.com/flosch/pongo2/v5"
	"github.com/halleck45/ast-metrics/internal/analyzer"
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

// cachedLangData holds pre-computed JSON strings for a given language view.
type cachedLangData struct {
	filesJSON           string
	risksJSON           string
	risksByPath         map[string][]riskItemForTpl
	nodeToCommunityJSON string
	testQualityJSON     string
	fileDepsJSON        string
	folderDepsJSON      string
	depFileCount        int
	dictionaryJSON      string
}

type HtmlReportGenerator struct {
	// The path where the report will be generated
	ReportPath string
	// langCache holds pre-computed JSON per language key (built once in Generate)
	langCache map[string]*cachedLangData
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
		"componentBubbleChart.html",
		"componentComparaisonBadge.html",
		"componentComparaisonOperator.html",
		"communities.html",
		"dependencies.html",
		"busfactor.html",
		"testquality.html",
		"partials/suggestions.html",
		"partials/file_explorer_sidebar.html",
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

	// Pre-compute JSON data once per language to avoid redundant work across pages
	v.langCache = make(map[string]*cachedLangData)
	langKeys := []string{"All"}
	for lang := range projectAggregated.ByProgrammingLanguage {
		langKeys = append(langKeys, lang)
	}
	for _, lang := range langKeys {
		cd := &cachedLangData{}
		dict := NewStringDictionary()

		cd.filesJSON = buildFilesJSONPruned(files, lang)

		// Build risks
		cd.risksByPath = map[string][]riskItemForTpl{}
		ra := analyzer.NewRiskAnalyzer()
		for _, f := range files {
			if lang != "All" && f.ProgrammingLanguage != lang {
				continue
			}
			items := ra.DetectFileRisks(f)
			if len(items) > 0 {
				converted := make([]riskItemForTpl, 0, len(items))
				for _, it := range items {
					converted = append(converted, riskItemForTpl{ID: it.ID, Title: it.Title, Severity: it.Severity, Details: it.Details})
				}
				cd.risksByPath[f.Path] = converted
			}
		}
		cd.risksJSON = buildRisksJSON(cd.risksByPath, dict)

		// Community
		var currentView analyzer.Aggregated
		if lang == "All" {
			currentView = projectAggregated.Combined
		} else {
			currentView = projectAggregated.ByProgrammingLanguage[lang]
		}
		cd.nodeToCommunityJSON = "{}"
		if currentView.Community != nil && len(currentView.Community.NodeToCommunity) > 0 {
			cd.nodeToCommunityJSON = buildNodeToCommunityJSON(currentView.Community.NodeToCommunity)
		}

		cd.testQualityJSON = "{}"
		if currentView.TestQuality != nil {
			cd.testQualityJSON = analyzer.BuildTestQualityJSON(currentView.TestQuality)
		}

		cd.fileDepsJSON = buildFileDepsJSON(files, lang, dict)

		// Count files for this language
		fileCount := 0
		for _, f := range files {
			if lang != "All" && f.GetProgrammingLanguage() != lang {
				continue
			}
			fileCount++
		}
		cd.depFileCount = fileCount

		// Build folder-level deps for dependency graph folder view
		cd.folderDepsJSON = buildFolderDepsJSON(files, lang, dict)

		cd.dictionaryJSON = dict.ToJSON()
		v.langCache[lang] = cd
	}

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

	// Dependencies page
	v.GenerateLanguagePage("dependencies.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("dependencies.html", language, currentView, files, projectAggregated)
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

	// Test Quality page
	v.GenerateLanguagePage("testquality.html", "All", projectAggregated.Combined, files, projectAggregated)
	for language, currentView := range projectAggregated.ByProgrammingLanguage {
		v.GenerateLanguagePage("testquality.html", language, currentView, files, projectAggregated)
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
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Severity float64 `json:"severity"`
	Details  string  `json:"details"`
}

// buildFilesJSONPruned builds a pruned JSON array of files with pathHash injected.
func buildFilesJSONPruned(files []*pb.File, language string) string {
	mo := protojson.MarshalOptions{EmitUnpopulated: false, UseEnumNumbers: false, Indent: ""}
	var b strings.Builder
	b.WriteString("[")
	first := true
	for _, f := range files {
		if language != "All" && f.GetProgrammingLanguage() != language {
			continue
		}
		cf := proto.Clone(f).(*pb.File)
		pruneFile(cf)

		data, err := mo.Marshal(cf)
		if err != nil {
			data = []byte("{}")
		}

		// Round-trip: unmarshal into map, add pathHash, re-marshal
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			m = map[string]any{}
		}
		m["pathHash"] = hashPathForExplorer(cf.GetPath())
		reData, err := json.Marshal(m)
		if err != nil {
			reData = []byte("{}")
		}

		if !first {
			b.WriteString(",")
		}
		b.Write(reData)
		first = false
	}
	b.WriteString("]")
	return b.String()
}

func hashPathForExplorer(path string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(path))
	return fmt.Sprintf("%016x", h.Sum64())
}

func pruneFile(f *pb.File) {
	if f.Stmts == nil {
		return
	}
	s := f.Stmts

	classes := engine.GetClassesInFile(f)
	for _, c := range classes {
		pruneClass(c)
	}
	s.StmtClass = classes

	outsideFunctions := engine.GetFunctionsOutsideClassesInFile(f)
	for _, fn := range outsideFunctions {
		pruneFunction(fn)
	}
	s.StmtFunction = outsideFunctions

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

func buildNodeToCommunityJSON(n2c map[string]string) string {
	if len(n2c) == 0 {
		return "{}"
	}
	data, err := json.Marshal(n2c)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func buildRisksJSON(risksByPath map[string][]riskItemForTpl, dict *StringDictionary) string {
	hashed := make(map[string][]riskItemForTpl, len(risksByPath))
	for p, items := range risksByPath {
		hashed[dict.Add(p)] = items
	}
	data, err := json.Marshal(hashed)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// buildFileDepsJSON builds a JSON map of file dependency relationships keyed by path hash.
func buildFileDepsJSON(files []*pb.File, language string, dict *StringDictionary) string {
	// Step 1: Build class qualified name -> file path lookup
	classToFile := map[string]string{}
	for _, f := range files {
		if language != "All" && f.GetProgrammingLanguage() != language {
			continue
		}
		if f.Stmts == nil {
			continue
		}
		classes := engine.GetClassesInFile(f)
		for _, c := range classes {
			if c.Name == nil {
				continue
			}
			if q := c.Name.GetQualified(); q != "" {
				classToFile[q] = f.Path
			}
			if s := c.Name.GetShort(); s != "" {
				if _, exists := classToFile[s]; !exists {
					classToFile[s] = f.Path
				}
			}
		}
	}

	// Step 2: Build efferent map from StmtExternalDependencies
	type depInfo struct {
		path  string
		short string
	}
	efferent := map[string]map[string]depInfo{}

	for _, f := range files {
		if language != "All" && f.GetProgrammingLanguage() != language {
			continue
		}
		if f.Stmts == nil {
			continue
		}

		deps := f.Stmts.GetStmtExternalDependencies()
		for _, ns := range f.Stmts.GetStmtNamespace() {
			if ns != nil && ns.Stmts != nil {
				deps = append(deps, ns.Stmts.GetStmtExternalDependencies()...)
			}
		}

		for _, dep := range deps {
			if dep == nil {
				continue
			}
			targetFile := ""
			if ns := dep.GetNamespace(); ns != "" {
				if fp, ok := classToFile[ns]; ok {
					targetFile = fp
				}
			}
			if targetFile == "" {
				if cn := dep.GetClassName(); cn != "" {
					if fp, ok := classToFile[cn]; ok {
						targetFile = fp
					}
				}
			}
			if targetFile == "" || targetFile == f.Path {
				continue
			}
			if efferent[f.Path] == nil {
				efferent[f.Path] = map[string]depInfo{}
			}
			short := targetFile
			if idx := strings.LastIndex(targetFile, "/"); idx >= 0 {
				short = targetFile[idx+1:]
			}
			efferent[f.Path][targetFile] = depInfo{path: targetFile, short: short}
		}
	}

	// Step 3: Invert to get afferent
	afferent := map[string]map[string]depInfo{}
	for srcFile, targets := range efferent {
		srcShort := srcFile
		if idx := strings.LastIndex(srcFile, "/"); idx >= 0 {
			srcShort = srcFile[idx+1:]
		}
		for tgtPath := range targets {
			if afferent[tgtPath] == nil {
				afferent[tgtPath] = map[string]depInfo{}
			}
			afferent[tgtPath][srcFile] = depInfo{path: srcFile, short: srcShort}
		}
	}

	// Step 4: Collect all files that have any dependency
	allFiles := map[string]struct{}{}
	for k := range efferent {
		allFiles[k] = struct{}{}
	}
	for k := range afferent {
		allFiles[k] = struct{}{}
	}

	if len(allFiles) == 0 {
		return "{}"
	}

	// Step 5: Build struct map keyed by hash
	result := make(map[string]fileDepsEntry, len(allFiles))
	for fp := range allFiles {
		entry := fileDepsEntry{
			Efferent: make([]depRef, 0),
			Afferent: make([]depRef, 0),
		}
		if eff, ok := efferent[fp]; ok {
			for _, d := range eff {
				entry.Efferent = append(entry.Efferent, depRef{
					Path:  dict.Add(d.path),
					Short: d.short,
				})
			}
		}
		if aff, ok := afferent[fp]; ok {
			for _, d := range aff {
				entry.Afferent = append(entry.Afferent, depRef{
					Path:  dict.Add(d.path),
					Short: d.short,
				})
			}
		}
		result[dict.Add(fp)] = entry
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// buildFolderDepsJSON aggregates file-level dependencies to folder-level.
// Keys are hashed via the dictionary.
func buildFolderDepsJSON(files []*pb.File, language string, dict *StringDictionary) string {
	classToFile := map[string]string{}
	for _, f := range files {
		if language != "All" && f.GetProgrammingLanguage() != language {
			continue
		}
		if f.Stmts == nil {
			continue
		}
		classes := engine.GetClassesInFile(f)
		for _, c := range classes {
			if c.Name == nil {
				continue
			}
			if q := c.Name.GetQualified(); q != "" {
				classToFile[q] = f.Path
			}
			if s := c.Name.GetShort(); s != "" {
				if _, exists := classToFile[s]; !exists {
					classToFile[s] = f.Path
				}
			}
		}
	}

	type edge struct {
		src string
		dst string
	}
	var edges []edge
	filesByFolder := map[string]map[string]struct{}{}

	for _, f := range files {
		if language != "All" && f.GetProgrammingLanguage() != language {
			continue
		}
		if f.Stmts == nil {
			continue
		}

		srcDir := path.Dir(f.Path)
		if filesByFolder[srcDir] == nil {
			filesByFolder[srcDir] = map[string]struct{}{}
		}
		filesByFolder[srcDir][f.Path] = struct{}{}

		deps := f.Stmts.GetStmtExternalDependencies()
		for _, ns := range f.Stmts.GetStmtNamespace() {
			if ns != nil && ns.Stmts != nil {
				deps = append(deps, ns.Stmts.GetStmtExternalDependencies()...)
			}
		}

		for _, dep := range deps {
			if dep == nil {
				continue
			}
			targetFile := ""
			if ns := dep.GetNamespace(); ns != "" {
				if fp, ok := classToFile[ns]; ok {
					targetFile = fp
				}
			}
			if targetFile == "" {
				if cn := dep.GetClassName(); cn != "" {
					if fp, ok := classToFile[cn]; ok {
						targetFile = fp
					}
				}
			}
			if targetFile == "" || targetFile == f.Path {
				continue
			}
			edges = append(edges, edge{src: f.Path, dst: targetFile})
		}
	}

	// Aggregate to folder level
	type folderEdgeCount struct {
		count int
	}
	folderEfferent := map[string]map[string]*folderEdgeCount{}
	folderAfferent := map[string]map[string]*folderEdgeCount{}
	folderFileCount := map[string]int{}

	for dir, fset := range filesByFolder {
		folderFileCount[dir] = len(fset)
	}

	for _, e := range edges {
		srcDir := path.Dir(e.src)
		dstDir := path.Dir(e.dst)
		if srcDir == dstDir {
			continue
		}
		if folderEfferent[srcDir] == nil {
			folderEfferent[srcDir] = map[string]*folderEdgeCount{}
		}
		if folderEfferent[srcDir][dstDir] == nil {
			folderEfferent[srcDir][dstDir] = &folderEdgeCount{}
		}
		folderEfferent[srcDir][dstDir].count++

		if folderAfferent[dstDir] == nil {
			folderAfferent[dstDir] = map[string]*folderEdgeCount{}
		}
		if folderAfferent[dstDir][srcDir] == nil {
			folderAfferent[dstDir][srcDir] = &folderEdgeCount{}
		}
		folderAfferent[dstDir][srcDir].count++
	}

	allFolders := map[string]struct{}{}
	for k := range folderEfferent {
		allFolders[k] = struct{}{}
	}
	for k := range folderAfferent {
		allFolders[k] = struct{}{}
	}

	if len(allFolders) == 0 {
		return ""
	}

	// Build payload using structs
	payload := folderDepsPayload{
		Folders:       make(map[string]folderDepsEntry, len(allFolders)),
		FilesByFolder: make(map[string][]string),
	}

	for dir := range allFolders {
		entry := folderDepsEntry{
			Efferent: make([]folderDepRef, 0),
			Afferent: make([]folderDepRef, 0),
		}
		if eff, ok := folderEfferent[dir]; ok {
			for target, fe := range eff {
				entry.Efferent = append(entry.Efferent, folderDepRef{
					Path:  dict.Add(target),
					Count: fe.count,
				})
			}
		}
		if aff, ok := folderAfferent[dir]; ok {
			for source, fe := range aff {
				entry.Afferent = append(entry.Afferent, folderDepRef{
					Path:  dict.Add(source),
					Count: fe.count,
				})
			}
		}
		fc := folderFileCount[dir]
		if fc == 0 {
			fc = 1
		}
		entry.FileCount = fc
		payload.Folders[dict.Add(dir)] = entry

		// filesByFolder
		if fset, ok := filesByFolder[dir]; ok && len(fset) > 0 {
			flist := make([]string, 0, len(fset))
			for fp := range fset {
				flist = append(flist, dict.Add(fp))
			}
			payload.FilesByFolder[dict.Add(dir)] = flist
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
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

	// Use pre-computed cached data for this language
	cd := v.langCache[language]
	out, err := tpl.Execute(pongo2.Context{"datetime": datetime, "page": template, "currentLanguage": language, "currentView": currentView, "projectAggregated": projectAggregated, "files": files, "risksByPath": cd.risksByPath, "filesJSON": cd.filesJSON, "risksJSON": cd.risksJSON, "nodeToCommunityJSON": cd.nodeToCommunityJSON, "testQualityJSON": cd.testQualityJSON, "fileDepsJSON": cd.fileDepsJSON, "folderDepsJSON": cd.folderDepsJSON, "depFileCount": cd.depFileCount, "dictionaryJSON": cd.dictionaryJSON})
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

	// filter convertOneFileToCollection
	pongo2.RegisterFilter("convertOneFileToCollection", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		file := in.Interface().(*pb.File)
		return pongo2.AsValue([]*pb.File{file}), nil
	})

	// filter getClassesInFile: returns classes via GetClassesInFile (namespace-aware).
	// After protobuf serialization/deserialization, file.Stmts.StmtClass and
	// namespace.Stmts.StmtClass are different objects. Coupling is computed on
	// GetClassesInFile results, so templates must use this filter.
	pongo2.RegisterFilter("getClassesInFile", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		file := in.Interface().(*pb.File)
		return pongo2.AsValue(engine.GetClassesInFile(file)), nil
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
}
