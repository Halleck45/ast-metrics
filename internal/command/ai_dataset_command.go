package command

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/halleck45/ast-metrics/internal/engine/golang"
	"github.com/halleck45/ast-metrics/internal/engine/php"
	"github.com/halleck45/ast-metrics/internal/engine/python"
	"github.com/halleck45/ast-metrics/internal/engine/rust"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/pterm/pterm"
)

type AIDatasetCommand struct {
	InputPath  string
	OutputPath string
	runners    []engine.Engine
}

func NewAIDatasetCommand(inputPath, outputPath string) *AIDatasetCommand {
	runnerPhp := php.PhpRunner{}
	runnerGolang := golang.GolangRunner{}
	runnerPython := python.PythonRunner{}
	runnerRust := rust.RustRunner{}
	runners := []engine.Engine{&runnerPhp, &runnerGolang, &runnerPython, &runnerRust}

	return &AIDatasetCommand{
		InputPath:  inputPath,
		OutputPath: outputPath,
		runners:    runners,
	}
}

func (c *AIDatasetCommand) Execute() error {
	// Validate input path
	if c.InputPath == "" {
		return fmt.Errorf("input path is required")
	}
	if c.OutputPath == "" {
		return fmt.Errorf("output path is required (use --output flag)")
	}

	// Check if input path exists
	info, err := os.Stat(c.InputPath)
	if err != nil {
		return fmt.Errorf("input path does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("input path must be a directory")
	}

	pterm.Info.Printf("Analyzing directory: %s\n", c.InputPath)
	pterm.Info.Printf("Output CSV file: %s\n", c.OutputPath)

	// Create configuration
	config := configuration.NewConfiguration()
	err = config.SetSourcesToAnalyzePath([]string{c.InputPath})
	if err != nil {
		return fmt.Errorf("failed to set source path: %w", err)
	}

	// Prepare workdir
	config.Storage.Purge()
	config.Storage.Ensure()

	// Set configuration for each runner
	for _, runner := range c.runners {
		runner.SetConfiguration(config)
		if !runner.IsRequired() {
			continue
		}

		err := runner.Ensure()
		if err != nil {
			pterm.Warning.Printf("Failed to ensure runner %T: %v\n", runner, err)
			continue
		}

		// Dump ASTs (in a goroutine like in analyze_command.go to avoid blocking)
		done := make(chan struct{})
		go func() {
			runner.DumpAST()
			close(done)
		}()
		<-done

		// Cleanup
		err = runner.Finish()
		if err != nil {
			pterm.Warning.Printf("Failed to finish runner %T: %v\n", runner, err)
		}
	}

	// Analyze all AST files
	allResults := analyzer.Start(config.Storage, nil)

	// Extract metrics and generate CSV
	err = c.generateCSV(allResults)
	if err != nil {
		return fmt.Errorf("failed to generate CSV: %w", err)
	}

	pterm.Success.Printf("Dataset generated successfully: %s\n", c.OutputPath)
	return nil
}

func (c *AIDatasetCommand) generateCSV(files []*pb.File) error {
	// Create CSV file
	file, err := os.Create(c.OutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"stmt_name",
		"stmt_type",
		"file_path",
		"method_calls_raw",
		"uses_raw",
		"namespace_raw",
		"path_raw",
		"class_loc",
		"logical_loc",
		"comment_loc",
		"nb_comments",
		"nb_methods",
		"nb_extends",
		"nb_implements",
		"nb_traits",
		"count_if",
		"count_elseif",
		"count_else",
		"count_case",
		"count_switch",
		"count_loop",
		"nb_external_dependencies",
		"depth_estimate",
		"nb_method_calls",
		"nb_getters",
		"nb_setters",
		"nb_attributes",
		"nb_unique_operators",
		"programming_language",
		"cyclomatic_complexity",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Process each file
	for _, file := range files {
		if file == nil || len(file.Errors) > 0 {
			continue
		}

		// Analyze file first to compute metrics
		analyzer.AnalyzeFile(file)

		// Get namespace
		namespace := c.getNamespace(file)

		// Process classes
		classes := engine.GetClassesInFile(file)
		if len(classes) > 0 {
			// If file has classes, process each class
			for _, class := range classes {
				row := c.extractClassMetrics(class, file, namespace)
				if err := writer.Write(row); err != nil {
					return err
				}
			}
		} else {
			// If file has no classes, use filename as stmt_name
			row := c.extractFileMetrics(file, namespace)
			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *AIDatasetCommand) getNamespace(file *pb.File) string {
	if file == nil || file.Stmts == nil {
		return ""
	}
	if len(file.Stmts.StmtNamespace) > 0 {
		ns := file.Stmts.StmtNamespace[0]
		if ns != nil && ns.Name != nil {
			if ns.Name.Qualified != "" {
				return ns.Name.Qualified
			}
			return ns.Name.Short
		}
	}
	return ""
}

func (c *AIDatasetCommand) getInterfacesInFile(file *pb.File) []*pb.StmtInterface {
	var interfaces []*pb.StmtInterface
	if file == nil || file.Stmts == nil {
		return interfaces
	}
	// Get interfaces from namespaces
	if file.Stmts.StmtNamespace != nil {
		for _, namespace := range file.Stmts.StmtNamespace {
			if namespace != nil && namespace.Stmts != nil {
				interfaces = append(interfaces, namespace.Stmts.StmtInterface...)
			}
		}
	}
	// Get top-level interfaces
	interfaces = append(interfaces, file.Stmts.StmtInterface...)
	return interfaces
}

func (c *AIDatasetCommand) isMethod(function *pb.StmtFunction, classes []*pb.StmtClass) bool {
	for _, class := range classes {
		if class != nil && class.Stmts != nil {
			for _, method := range class.Stmts.StmtFunction {
				if method == function {
					return true
				}
			}
		}
	}
	return false
}

func (c *AIDatasetCommand) extractClassMetrics(class *pb.StmtClass, file *pb.File, namespace string) []string {
	if class == nil {
		return c.emptyRow()
	}

	// Metrics are already computed by analyzer.AnalyzeFile

	stmtName := ""
	if class.Name != nil {
		stmtName = class.Name.Qualified
		if stmtName == "" {
			stmtName = class.Name.Short
		}
	}

	// Get externals (just count, not the list)
	externals := c.getExternalsForClass(class, file)

	// Get method calls from all methods
	methodCalls := c.getMethodCallsForClass(class)
	methodCallsRaw := strings.Join(methodCalls, ";")

	// Get uses (imports)
	uses := c.getUsesForClass(class, file)
	usesRaw := strings.Join(uses, ";")

	// LOC metrics
	classLoc := int32(0)
	logicalLoc := int32(0)
	commentLoc := int32(0)
	if class.LinesOfCode != nil {
		classLoc = class.LinesOfCode.LinesOfCode
		logicalLoc = class.LinesOfCode.LogicalLinesOfCode
		commentLoc = class.LinesOfCode.CommentLinesOfCode
	} else if class.Stmts != nil && class.Stmts.Analyze != nil && class.Stmts.Analyze.Volume != nil {
		if class.Stmts.Analyze.Volume.Loc != nil {
			classLoc = *class.Stmts.Analyze.Volume.Loc
		}
		if class.Stmts.Analyze.Volume.Lloc != nil {
			logicalLoc = *class.Stmts.Analyze.Volume.Lloc
		}
		if class.Stmts.Analyze.Volume.Cloc != nil {
			commentLoc = *class.Stmts.Analyze.Volume.Cloc
		}
	}

	// Count comments
	nbComments := int32(0)
	if class.Comments != nil {
		nbComments = int32(len(class.Comments))
	}

	// Count methods
	nbMethods := int32(0)
	if class.Stmts != nil {
		nbMethods = int32(len(class.Stmts.StmtFunction))
	}

	// Count extends, implements, traits
	nbExtends := int32(len(class.Extends))
	nbImplements := int32(len(class.Implements))
	nbTraits := int32(len(class.Uses))

	// Count control structures
	countIf := int32(0)
	countElseif := int32(0)
	countElse := int32(0)
	countCase := int32(0)
	countSwitch := int32(0)
	countLoop := int32(0)
	if class.Stmts != nil {
		countIf = int32(len(class.Stmts.StmtDecisionIf))
		countElseif = int32(len(class.Stmts.StmtDecisionElseIf))
		countElse = int32(len(class.Stmts.StmtDecisionElse))
		countCase = int32(len(class.Stmts.StmtDecisionCase))
		countSwitch = int32(len(class.Stmts.StmtDecisionSwitch))
		countLoop = int32(len(class.Stmts.StmtLoop))
	}

	// Count external dependencies
	nbExternalDependencies := int32(len(externals))

	// Depth estimate
	depthEstimate := c.calculateDepthEstimate(class.Stmts)

	// Count method calls
	nbMethodCalls := int32(len(methodCalls))

	// Count getters and setters
	nbGetters := c.countGetters(class)
	nbSetters := c.countSetters(class)

	// Count attributes (operands/properties)
	nbAttributes := int32(0)
	if class.Operands != nil {
		nbAttributes = int32(len(class.Operands))
	}

	// Count unique operators
	nbUniqueOperators := c.countUniqueOperators(class)

	// Programming language
	programmingLanguage := file.ProgrammingLanguage
	if programmingLanguage == "" {
		programmingLanguage = "unknown"
	}

	// Cyclomatic complexity
	cyclomaticComplexity := int32(0)
	if class.Stmts != nil && class.Stmts.Analyze != nil && class.Stmts.Analyze.Complexity != nil {
		if class.Stmts.Analyze.Complexity.Cyclomatic != nil {
			cyclomaticComplexity = *class.Stmts.Analyze.Complexity.Cyclomatic
		}
	}

	// Path
	pathRaw := file.Path
	if pathRaw == "" {
		pathRaw = file.ShortPath
	}

	// Get file path (relative)
	filePath := c.getRelativeFilePath(file)

	return []string{
		stmtName,
		"class",
		filePath,
		methodCallsRaw,
		usesRaw,
		namespace,
		pathRaw,
		fmt.Sprintf("%d", classLoc),
		fmt.Sprintf("%d", logicalLoc),
		fmt.Sprintf("%d", commentLoc),
		fmt.Sprintf("%d", nbComments),
		fmt.Sprintf("%d", nbMethods),
		fmt.Sprintf("%d", nbExtends),
		fmt.Sprintf("%d", nbImplements),
		fmt.Sprintf("%d", nbTraits),
		fmt.Sprintf("%d", countIf),
		fmt.Sprintf("%d", countElseif),
		fmt.Sprintf("%d", countElse),
		fmt.Sprintf("%d", countCase),
		fmt.Sprintf("%d", countSwitch),
		fmt.Sprintf("%d", countLoop),
		fmt.Sprintf("%d", nbExternalDependencies),
		fmt.Sprintf("%d", depthEstimate),
		fmt.Sprintf("%d", nbMethodCalls),
		fmt.Sprintf("%d", nbGetters),
		fmt.Sprintf("%d", nbSetters),
		fmt.Sprintf("%d", nbAttributes),
		fmt.Sprintf("%d", nbUniqueOperators),
		programmingLanguage,
		fmt.Sprintf("%d", cyclomaticComplexity),
	}
}

func (c *AIDatasetCommand) extractInterfaceMetrics(iface *pb.StmtInterface, file *pb.File, namespace string) []string {
	if iface == nil {
		return c.emptyRow()
	}

	stmtName := ""
	if iface.Name != nil {
		stmtName = iface.Name.Qualified
		if stmtName == "" {
			stmtName = iface.Name.Short
		}
	}

	// Interfaces typically don't have much code, but we'll extract what we can
	externalsRaw := ""
	methodCallsRaw := ""
	usesRaw := ""

	// LOC metrics (usually 0 for interfaces)
	classLoc := int32(0)
	logicalLoc := int32(0)
	commentLoc := int32(0)
	if iface.Stmts != nil && iface.Stmts.Analyze != nil && iface.Stmts.Analyze.Volume != nil {
		if iface.Stmts.Analyze.Volume.Loc != nil {
			classLoc = *iface.Stmts.Analyze.Volume.Loc
		}
		if iface.Stmts.Analyze.Volume.Lloc != nil {
			logicalLoc = *iface.Stmts.Analyze.Volume.Lloc
		}
		if iface.Stmts.Analyze.Volume.Cloc != nil {
			commentLoc = *iface.Stmts.Analyze.Volume.Cloc
		}
	}

	nbComments := int32(0)
	nbMethods := int32(0)
	if iface.Stmts != nil {
		nbMethods = int32(len(iface.Stmts.StmtFunction))
	}

	nbExtends := int32(len(iface.Extends))
	nbImplements := int32(0) // Interfaces don't implement, they extend
	nbTraits := int32(0)

	// Count control structures
	countIf := int32(0)
	countElseif := int32(0)
	countElse := int32(0)
	countCase := int32(0)
	countSwitch := int32(0)
	countLoop := int32(0)
	if iface.Stmts != nil {
		countIf = int32(len(iface.Stmts.StmtDecisionIf))
		countElseif = int32(len(iface.Stmts.StmtDecisionElseIf))
		countElse = int32(len(iface.Stmts.StmtDecisionElse))
		countCase = int32(len(iface.Stmts.StmtDecisionCase))
		countSwitch = int32(len(iface.Stmts.StmtDecisionSwitch))
		countLoop = int32(len(iface.Stmts.StmtLoop))
	}

	nbExternalDependencies := int32(0)
	depthEstimate := c.calculateDepthEstimate(iface.Stmts)
	nbMethodCalls := int32(0)

	pathRaw := file.Path
	if pathRaw == "" {
		pathRaw = file.ShortPath
	}

	return []string{
		stmtName,
		"interface",
		externalsRaw,
		methodCallsRaw,
		usesRaw,
		namespace,
		pathRaw,
		fmt.Sprintf("%d", classLoc),
		fmt.Sprintf("%d", logicalLoc),
		fmt.Sprintf("%d", commentLoc),
		fmt.Sprintf("%d", nbComments),
		fmt.Sprintf("%d", nbMethods),
		fmt.Sprintf("%d", nbExtends),
		fmt.Sprintf("%d", nbImplements),
		fmt.Sprintf("%d", nbTraits),
		fmt.Sprintf("%d", countIf),
		fmt.Sprintf("%d", countElseif),
		fmt.Sprintf("%d", countElse),
		fmt.Sprintf("%d", countCase),
		fmt.Sprintf("%d", countSwitch),
		fmt.Sprintf("%d", countLoop),
		fmt.Sprintf("%d", nbExternalDependencies),
		fmt.Sprintf("%d", depthEstimate),
		fmt.Sprintf("%d", nbMethodCalls),
	}
}

func (c *AIDatasetCommand) extractFunctionMetrics(function *pb.StmtFunction, file *pb.File, namespace string) []string {
	if function == nil {
		return c.emptyRow()
	}

	// Metrics are already computed by analyzer.AnalyzeFile

	stmtName := ""
	if function.Name != nil {
		stmtName = function.Name.Qualified
		if stmtName == "" {
			stmtName = function.Name.Short
		}
	}

	// Get externals
	externals := c.getExternalsForFunction(function)
	externalsRaw := strings.Join(externals, ";")

	// Get method calls
	methodCalls := c.getMethodCallsForFunction(function)
	methodCallsRaw := strings.Join(methodCalls, ";")

	// Get uses (from file level)
	uses := c.getUsesForFile(file)
	usesRaw := strings.Join(uses, ";")

	// LOC metrics
	classLoc := int32(0)
	logicalLoc := int32(0)
	commentLoc := int32(0)
	if function.LinesOfCode != nil {
		classLoc = function.LinesOfCode.LinesOfCode
		logicalLoc = function.LinesOfCode.LogicalLinesOfCode
		commentLoc = function.LinesOfCode.CommentLinesOfCode
	} else if function.Stmts != nil && function.Stmts.Analyze != nil && function.Stmts.Analyze.Volume != nil {
		if function.Stmts.Analyze.Volume.Loc != nil {
			classLoc = *function.Stmts.Analyze.Volume.Loc
		}
		if function.Stmts.Analyze.Volume.Lloc != nil {
			logicalLoc = *function.Stmts.Analyze.Volume.Lloc
		}
		if function.Stmts.Analyze.Volume.Cloc != nil {
			commentLoc = *function.Stmts.Analyze.Volume.Cloc
		}
	}

	// Count comments
	nbComments := int32(0)
	if function.Comments != nil {
		nbComments = int32(len(function.Comments))
	}

	nbMethods := int32(0) // Functions don't have methods
	nbExtends := int32(0)
	nbImplements := int32(0)
	nbTraits := int32(0)

	// Count control structures
	countIf := int32(0)
	countElseif := int32(0)
	countElse := int32(0)
	countCase := int32(0)
	countSwitch := int32(0)
	countLoop := int32(0)
	if function.Stmts != nil {
		countIf = int32(len(function.Stmts.StmtDecisionIf))
		countElseif = int32(len(function.Stmts.StmtDecisionElseIf))
		countElse = int32(len(function.Stmts.StmtDecisionElse))
		countCase = int32(len(function.Stmts.StmtDecisionCase))
		countSwitch = int32(len(function.Stmts.StmtDecisionSwitch))
		countLoop = int32(len(function.Stmts.StmtLoop))
	}

	nbExternalDependencies := int32(len(externals))
	depthEstimate := c.calculateDepthEstimate(function.Stmts)
	nbMethodCalls := int32(len(methodCalls))

	pathRaw := file.Path
	if pathRaw == "" {
		pathRaw = file.ShortPath
	}

	return []string{
		stmtName,
		"function",
		externalsRaw,
		methodCallsRaw,
		usesRaw,
		namespace,
		pathRaw,
		fmt.Sprintf("%d", classLoc),
		fmt.Sprintf("%d", logicalLoc),
		fmt.Sprintf("%d", commentLoc),
		fmt.Sprintf("%d", nbComments),
		fmt.Sprintf("%d", nbMethods),
		fmt.Sprintf("%d", nbExtends),
		fmt.Sprintf("%d", nbImplements),
		fmt.Sprintf("%d", nbTraits),
		fmt.Sprintf("%d", countIf),
		fmt.Sprintf("%d", countElseif),
		fmt.Sprintf("%d", countElse),
		fmt.Sprintf("%d", countCase),
		fmt.Sprintf("%d", countSwitch),
		fmt.Sprintf("%d", countLoop),
		fmt.Sprintf("%d", nbExternalDependencies),
		fmt.Sprintf("%d", depthEstimate),
		fmt.Sprintf("%d", nbMethodCalls),
	}
}

// analyzeStatements is not needed as analyzer.AnalyzeFile already computes
// all metrics at the file level. Statement-level metrics are computed
// during the file analysis and are available in the Stmts.Analyze structure.

func (c *AIDatasetCommand) getExternalsForClass(class *pb.StmtClass, file *pb.File) []string {
	var externals []string
	seen := make(map[string]bool)

	// Get externals from class stmts
	if class.Stmts != nil {
		for _, dep := range class.Stmts.StmtExternalDependencies {
			if dep != nil {
				key := fmt.Sprintf("%s::%s", dep.Namespace, dep.ClassName)
				if dep.FunctionName != "" {
					key = fmt.Sprintf("%s::%s", key, dep.FunctionName)
				}
				if !seen[key] {
					externals = append(externals, key)
					seen[key] = true
				}
			}
		}
	}

	// Get externals from extends/implements/uses
	for _, ext := range class.Extends {
		if ext != nil {
			key := ext.Qualified
			if key == "" {
				key = ext.Short
			}
			if !seen[key] {
				externals = append(externals, key)
				seen[key] = true
			}
		}
	}
	for _, impl := range class.Implements {
		if impl != nil {
			key := impl.Qualified
			if key == "" {
				key = impl.Short
			}
			if !seen[key] {
				externals = append(externals, key)
				seen[key] = true
			}
		}
	}
	for _, use := range class.Uses {
		if use != nil {
			key := use.Qualified
			if key == "" {
				key = use.Short
			}
			if !seen[key] {
				externals = append(externals, key)
				seen[key] = true
			}
		}
	}

	// Get externals from methods
	if class.Stmts != nil {
		for _, method := range class.Stmts.StmtFunction {
			for _, ext := range method.Externals {
				if ext != nil {
					key := ext.Qualified
					if key == "" {
						key = ext.Short
					}
					if !seen[key] {
						externals = append(externals, key)
						seen[key] = true
					}
				}
			}
		}
	}

	return externals
}

func (c *AIDatasetCommand) getExternalsForFunction(function *pb.StmtFunction) []string {
	var externals []string
	seen := make(map[string]bool)

	// Get externals from function
	for _, ext := range function.Externals {
		if ext != nil {
			key := ext.Qualified
			if key == "" {
				key = ext.Short
			}
			if !seen[key] {
				externals = append(externals, key)
				seen[key] = true
			}
		}
	}

	// Get externals from stmts
	if function.Stmts != nil {
		for _, dep := range function.Stmts.StmtExternalDependencies {
			if dep != nil {
				key := fmt.Sprintf("%s::%s", dep.Namespace, dep.ClassName)
				if dep.FunctionName != "" {
					key = fmt.Sprintf("%s::%s", key, dep.FunctionName)
				}
				if !seen[key] {
					externals = append(externals, key)
					seen[key] = true
				}
			}
		}
	}

	return externals
}

func (c *AIDatasetCommand) getMethodCallsForClass(class *pb.StmtClass) []string {
	var methodCalls []string
	seen := make(map[string]bool)

	if class.Stmts != nil {
		for _, method := range class.Stmts.StmtFunction {
			for _, call := range method.MethodCalls {
				if call != nil && call.Name != "" {
					if !seen[call.Name] {
						methodCalls = append(methodCalls, call.Name)
						seen[call.Name] = true
					}
				}
			}
		}
	}

	return methodCalls
}

func (c *AIDatasetCommand) getMethodCallsForFunction(function *pb.StmtFunction) []string {
	var methodCalls []string
	seen := make(map[string]bool)

	for _, call := range function.MethodCalls {
		if call != nil && call.Name != "" {
			if !seen[call.Name] {
				methodCalls = append(methodCalls, call.Name)
				seen[call.Name] = true
			}
		}
	}

	return methodCalls
}

func (c *AIDatasetCommand) getUsesForClass(class *pb.StmtClass, file *pb.File) []string {
	// Uses are typically at file/namespace level, not class level
	return c.getUsesForFile(file)
}

func (c *AIDatasetCommand) getUsesForFile(file *pb.File) []string {
	var uses []string
	seen := make(map[string]bool)

	if file == nil || file.Stmts == nil {
		return uses
	}

	// Get uses from file level
	for _, use := range file.Stmts.StmtUse {
		if use != nil && use.Name != nil {
			key := use.Name.Qualified
			if key == "" {
				key = use.Name.Short
			}
			if key != "" && !seen[key] {
				uses = append(uses, key)
				seen[key] = true
			}
		}
	}

	// Get uses from namespaces
	for _, ns := range file.Stmts.StmtNamespace {
		if ns != nil && ns.Stmts != nil {
			for _, use := range ns.Stmts.StmtUse {
				if use != nil && use.Name != nil {
					key := use.Name.Qualified
					if key == "" {
						key = use.Name.Short
					}
					if key != "" && !seen[key] {
						uses = append(uses, key)
						seen[key] = true
					}
				}
			}
		}
	}

	// Get external dependencies as uses
	deps := engine.GetDependenciesInFile(file)
	for _, dep := range deps {
		if dep != nil && dep.Namespace != "" {
			if !seen[dep.Namespace] {
				uses = append(uses, dep.Namespace)
				seen[dep.Namespace] = true
			}
		}
	}

	return uses
}

func (c *AIDatasetCommand) calculateDepthEstimate(stmts *pb.Stmts) int32 {
	if stmts == nil {
		return 0
	}

	// Use a simplified depth calculation based on nested control structures
	var countDepth func(*pb.Stmts, int32) int32
	countDepth = func(s *pb.Stmts, currentDepth int32) int32 {
		if s == nil {
			return currentDepth
		}

		maxDepth := currentDepth

		// Count nested structures
		for _, stmt := range s.StmtDecisionIf {
			d := countDepth(stmt.Stmts, currentDepth+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
		for _, stmt := range s.StmtDecisionElseIf {
			d := countDepth(stmt.Stmts, currentDepth+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
		for _, stmt := range s.StmtDecisionElse {
			d := countDepth(stmt.Stmts, currentDepth+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
		for _, stmt := range s.StmtDecisionSwitch {
			d := countDepth(stmt.Stmts, currentDepth+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
		for _, stmt := range s.StmtDecisionCase {
			d := countDepth(stmt.Stmts, currentDepth)
			if d > maxDepth {
				maxDepth = d
			}
		}
		for _, stmt := range s.StmtLoop {
			d := countDepth(stmt.Stmts, currentDepth+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
		for _, stmt := range s.StmtFunction {
			d := countDepth(stmt.Stmts, currentDepth)
			if d > maxDepth {
				maxDepth = d
			}
		}
		for _, stmt := range s.StmtClass {
			d := countDepth(stmt.Stmts, currentDepth)
			if d > maxDepth {
				maxDepth = d
			}
		}

		return maxDepth
	}

	return countDepth(stmts, 0)
}

func (c *AIDatasetCommand) extractFileMetrics(file *pb.File, namespace string) []string {
	if file == nil {
		return c.emptyRow()
	}

	// Get filename without extension as stmt_name
	stmtName := ""
	pathRaw := file.Path
	if pathRaw == "" {
		pathRaw = file.ShortPath
	}
	if pathRaw != "" {
		base := filepath.Base(pathRaw)
		ext := filepath.Ext(base)
		stmtName = strings.TrimSuffix(base, ext)
	}

	// Get uses (imports)
	uses := c.getUsesForFile(file)
	usesRaw := strings.Join(uses, ";")

	// Get externals from file level (just count)
	deps := engine.GetDependenciesInFile(file)

	// Get method calls from file-level functions
	methodCalls := make([]string, 0)
	seenCalls := make(map[string]bool)
	functions := engine.GetFunctionsInFile(file)
	for _, function := range functions {
		for _, call := range function.MethodCalls {
			if call != nil && call.Name != "" {
				if !seenCalls[call.Name] {
					methodCalls = append(methodCalls, call.Name)
					seenCalls[call.Name] = true
				}
			}
		}
	}
	methodCallsRaw := strings.Join(methodCalls, ";")

	// LOC metrics from file
	classLoc := int32(0)
	logicalLoc := int32(0)
	commentLoc := int32(0)
	if file.LinesOfCode != nil {
		classLoc = file.LinesOfCode.LinesOfCode
		logicalLoc = file.LinesOfCode.LogicalLinesOfCode
		commentLoc = file.LinesOfCode.CommentLinesOfCode
	} else if file.Stmts != nil && file.Stmts.Analyze != nil && file.Stmts.Analyze.Volume != nil {
		if file.Stmts.Analyze.Volume.Loc != nil {
			classLoc = *file.Stmts.Analyze.Volume.Loc
		}
		if file.Stmts.Analyze.Volume.Lloc != nil {
			logicalLoc = *file.Stmts.Analyze.Volume.Lloc
		}
		if file.Stmts.Analyze.Volume.Cloc != nil {
			commentLoc = *file.Stmts.Analyze.Volume.Cloc
		}
	}

	// Count comments (from file-level)
	nbComments := int32(0)
	if file.Stmts != nil {
		for _, fn := range file.Stmts.StmtFunction {
			if fn != nil && fn.Comments != nil {
				nbComments += int32(len(fn.Comments))
			}
		}
	}

	nbMethods := int32(0)
	if file.Stmts != nil {
		nbMethods = int32(len(file.Stmts.StmtFunction))
	}

	nbExtends := int32(0)
	nbImplements := int32(0)
	nbTraits := int32(0)

	// Count control structures from file-level
	countIf := int32(0)
	countElseif := int32(0)
	countElse := int32(0)
	countCase := int32(0)
	countSwitch := int32(0)
	countLoop := int32(0)
	if file.Stmts != nil {
		countIf = int32(len(file.Stmts.StmtDecisionIf))
		countElseif = int32(len(file.Stmts.StmtDecisionElseIf))
		countElse = int32(len(file.Stmts.StmtDecisionElse))
		countCase = int32(len(file.Stmts.StmtDecisionCase))
		countSwitch = int32(len(file.Stmts.StmtDecisionSwitch))
		countLoop = int32(len(file.Stmts.StmtLoop))
	}

	nbExternalDependencies := int32(len(deps))
	depthEstimate := c.calculateDepthEstimate(file.Stmts)
	nbMethodCalls := int32(len(methodCalls))

	// For files, getters/setters/attributes don't apply (they're 0)
	nbGetters := int32(0)
	nbSetters := int32(0)
	nbAttributes := int32(0)

	// Count unique operators from file-level
	nbUniqueOperators := c.countUniqueOperatorsForFile(file)

	// Programming language
	programmingLanguage := file.ProgrammingLanguage
	if programmingLanguage == "" {
		programmingLanguage = "unknown"
	}

	// Cyclomatic complexity
	cyclomaticComplexity := int32(0)
	if file.Stmts != nil && file.Stmts.Analyze != nil && file.Stmts.Analyze.Complexity != nil {
		if file.Stmts.Analyze.Complexity.Cyclomatic != nil {
			cyclomaticComplexity = *file.Stmts.Analyze.Complexity.Cyclomatic
		}
	}

	// Get file path (relative)
	filePath := c.getRelativeFilePath(file)

	return []string{
		stmtName,
		"file",
		filePath,
		methodCallsRaw,
		usesRaw,
		namespace,
		pathRaw,
		fmt.Sprintf("%d", classLoc),
		fmt.Sprintf("%d", logicalLoc),
		fmt.Sprintf("%d", commentLoc),
		fmt.Sprintf("%d", nbComments),
		fmt.Sprintf("%d", nbMethods),
		fmt.Sprintf("%d", nbExtends),
		fmt.Sprintf("%d", nbImplements),
		fmt.Sprintf("%d", nbTraits),
		fmt.Sprintf("%d", countIf),
		fmt.Sprintf("%d", countElseif),
		fmt.Sprintf("%d", countElse),
		fmt.Sprintf("%d", countCase),
		fmt.Sprintf("%d", countSwitch),
		fmt.Sprintf("%d", countLoop),
		fmt.Sprintf("%d", nbExternalDependencies),
		fmt.Sprintf("%d", depthEstimate),
		fmt.Sprintf("%d", nbMethodCalls),
		fmt.Sprintf("%d", nbGetters),
		fmt.Sprintf("%d", nbSetters),
		fmt.Sprintf("%d", nbAttributes),
		fmt.Sprintf("%d", nbUniqueOperators),
		programmingLanguage,
		fmt.Sprintf("%d", cyclomaticComplexity),
	}
}

func (c *AIDatasetCommand) getRelativeFilePath(file *pb.File) string {
	if file == nil {
		return ""
	}

	// Use ShortPath if available (it's usually relative)
	if file.ShortPath != "" {
		return file.ShortPath
	}

	// Otherwise use Path and try to make it relative to InputPath
	if file.Path != "" {
		// If InputPath is set, try to make it relative
		if c.InputPath != "" {
			relPath, err := filepath.Rel(c.InputPath, file.Path)
			if err == nil {
				return relPath
			}
		}
		return file.Path
	}

	return ""
}

func (c *AIDatasetCommand) countGetters(class *pb.StmtClass) int32 {
	if class == nil || class.Stmts == nil {
		return 0
	}
	count := int32(0)
	for _, method := range class.Stmts.StmtFunction {
		if method == nil || method.Name == nil {
			continue
		}
		methodName := method.Name.Short
		if methodName == "" {
			methodName = method.Name.Qualified
		}
		// Check if method name starts with get, is, or has (common getter patterns)
		lowerName := strings.ToLower(methodName)
		if strings.HasPrefix(lowerName, "get") || strings.HasPrefix(lowerName, "is") || strings.HasPrefix(lowerName, "has") {
			count++
		}
	}
	return count
}

func (c *AIDatasetCommand) countSetters(class *pb.StmtClass) int32 {
	if class == nil || class.Stmts == nil {
		return 0
	}
	count := int32(0)
	for _, method := range class.Stmts.StmtFunction {
		if method == nil || method.Name == nil {
			continue
		}
		methodName := method.Name.Short
		if methodName == "" {
			methodName = method.Name.Qualified
		}
		// Check if method name starts with set (common setter pattern)
		lowerName := strings.ToLower(methodName)
		if strings.HasPrefix(lowerName, "set") {
			count++
		}
	}
	return count
}

func (c *AIDatasetCommand) countUniqueOperators(class *pb.StmtClass) int32 {
	if class == nil || class.Operators == nil {
		return 0
	}
	seen := make(map[string]bool)
	for _, op := range class.Operators {
		if op != nil && op.Name != "" {
			seen[op.Name] = true
		}
	}
	return int32(len(seen))
}

func (c *AIDatasetCommand) countUniqueOperatorsForFile(file *pb.File) int32 {
	if file == nil || file.Stmts == nil {
		return 0
	}
	seen := make(map[string]bool)
	// Collect operators from all functions in the file
	functions := engine.GetFunctionsInFile(file)
	for _, fn := range functions {
		if fn != nil && fn.Operators != nil {
			for _, op := range fn.Operators {
				if op != nil && op.Name != "" {
					seen[op.Name] = true
				}
			}
		}
	}
	return int32(len(seen))
}

func (c *AIDatasetCommand) emptyRow() []string {
	return []string{
		"", "", "", "", "", "", "", "0", "0", "0", "0", "0", "0", "0", "0",
		"0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "", "0",
	}
}
