package Analyzer

import (
	"math"

	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type ProjectAggregated struct {
	ByFile                Aggregated
	ByClass               Aggregated
	Combined              Aggregated
	ByProgrammingLanguage map[string]Aggregated
}

type Aggregated struct {
	ConcernedFiles                       []*pb.File
	NbFiles                              int
	NbFunctions                          int
	NbClasses                            int
	NbClassesWithCode                    int
	NbMethods                            int
	Loc                                  int
	Cloc                                 int
	Lloc                                 int
	AverageMethodsPerClass               float64
	AverageLocPerMethod                  float64
	AverageLlocPerMethod                 float64
	AverageClocPerMethod                 float64
	AverageCyclomaticComplexityPerMethod float64
	AverageCyclomaticComplexityPerClass  float64
	MinCyclomaticComplexity              int
	MaxCyclomaticComplexity              int
	AverageHalsteadDifficulty            float64
	AverageHalsteadEffort                float64
	AverageHalsteadVolume                float64
	AverageHalsteadTime                  float64
	AverageHalsteadBugs                  float64
	SumHalsteadDifficulty                float64
	SumHalsteadEffort                    float64
	SumHalsteadVolume                    float64
	SumHalsteadTime                      float64
	SumHalsteadBugs                      float64
	AverageMI                            float64
	AverageMIwoc                         float64
	AverageMIcw                          float64
	AverageMIPerMethod                   float64
	AverageMIwocPerMethod                float64
	AverageMIcwPerMethod                 float64
	CommitCountForPeriod                 int
	CommittedFilesCountForPeriod         int // for example if one commit concerns 10 files, it will be 10
	BusFactor                            int
	TopCommitters                        []TopCommitter
	ResultOfGitAnalysis                  []ResultOfGitAnalysis
}

type Aggregator struct {
	files             []*pb.File
	projectAggregated ProjectAggregated
	analyzers         []AggregateAnalyzer
	gitSummaries      []ResultOfGitAnalysis
}

type TopCommitter struct {
	Name  string
	Count int
}

type ResultOfGitAnalysis struct {
	ProgrammingLanguage     string
	ReportRootDir           string
	CountCommits            int
	CountCommiters          int
	CountCommitsForLanguage int
	CountCommitsIgnored     int
}

func NewAggregator(files []*pb.File, gitSummaries []ResultOfGitAnalysis) *Aggregator {
	return &Aggregator{
		files:             files,
		projectAggregated: ProjectAggregated{},
		gitSummaries:      gitSummaries,
	}
}

type AggregateAnalyzer interface {
	Calculate(aggregate *Aggregated)
}

func newAggregated() Aggregated {
	return Aggregated{
		ConcernedFiles:                       make([]*pb.File, 0),
		NbClasses:                            0,
		NbClassesWithCode:                    0,
		NbMethods:                            0,
		NbFunctions:                          0,
		Loc:                                  0,
		Cloc:                                 0,
		Lloc:                                 0,
		AverageLocPerMethod:                  0,
		AverageLlocPerMethod:                 0,
		AverageClocPerMethod:                 0,
		AverageCyclomaticComplexityPerMethod: 0,
		AverageCyclomaticComplexityPerClass:  0,
		MinCyclomaticComplexity:              0,
		MaxCyclomaticComplexity:              0,
		AverageHalsteadDifficulty:            0,
		AverageHalsteadEffort:                0,
		AverageHalsteadVolume:                0,
		AverageHalsteadTime:                  0,
		AverageHalsteadBugs:                  0,
		SumHalsteadDifficulty:                0,
		SumHalsteadEffort:                    0,
		SumHalsteadVolume:                    0,
		SumHalsteadTime:                      0,
		SumHalsteadBugs:                      0,
		AverageMI:                            0,
		AverageMIwoc:                         0,
		AverageMIcw:                          0,
		AverageMIPerMethod:                   0,
		AverageMIwocPerMethod:                0,
		AverageMIcwPerMethod:                 0,
		CommitCountForPeriod:                 0,
		ResultOfGitAnalysis:                  nil,
	}
}

func (r *Aggregator) Aggregates() ProjectAggregated {
	files := r.files

	// We create a new aggregated object for each type of aggregation
	// ByFile, ByClass, Combined
	r.projectAggregated.ByFile = newAggregated()
	r.projectAggregated.ByClass = newAggregated()
	r.projectAggregated.Combined = newAggregated()

	// Count files
	r.projectAggregated.ByClass.NbFiles = len(files)
	r.projectAggregated.ByFile.NbFiles = len(files)
	r.projectAggregated.Combined.NbFiles = len(files)

	for _, file := range files {

		if file.Stmts == nil {
			continue
		}

		// By language
		if r.projectAggregated.ByProgrammingLanguage == nil {
			r.projectAggregated.ByProgrammingLanguage = make(map[string]Aggregated)
		}
		if _, ok := r.projectAggregated.ByProgrammingLanguage[file.ProgrammingLanguage]; !ok {
			r.projectAggregated.ByProgrammingLanguage[file.ProgrammingLanguage] = newAggregated()

		}
		byLanguage := r.projectAggregated.ByProgrammingLanguage[file.ProgrammingLanguage]
		byLanguage.NbFiles++

		// Make calculations: sums of metrics, etc.
		r.calculateSums(file, &r.projectAggregated.ByFile)
		r.calculateSums(file, &r.projectAggregated.ByClass)
		r.calculateSums(file, &r.projectAggregated.Combined)
		r.calculateSums(file, &byLanguage)
		r.projectAggregated.ByProgrammingLanguage[file.ProgrammingLanguage] = byLanguage
	}

	// Consolidate averages
	r.consolidate(&r.projectAggregated.ByFile)
	r.consolidate(&r.projectAggregated.ByClass)
	r.consolidate(&r.projectAggregated.Combined)

	// by language
	for lng, byLanguage := range r.projectAggregated.ByProgrammingLanguage {
		r.consolidate(&byLanguage)
		r.projectAggregated.ByProgrammingLanguage[lng] = byLanguage
	}

	// Risks
	riskAnalyzer := NewRiskAnalyzer()
	riskAnalyzer.Analyze(r.projectAggregated)

	return r.projectAggregated
}

// Consolidate the aggregated data
func (r *Aggregator) consolidate(aggregated *Aggregated) {

	if aggregated.NbClasses > 0 {
		aggregated.AverageMethodsPerClass = float64(aggregated.NbMethods) / float64(aggregated.NbClasses)
		aggregated.AverageCyclomaticComplexityPerClass = aggregated.AverageCyclomaticComplexityPerClass / float64(aggregated.NbClasses)
	} else {
		aggregated.AverageMethodsPerClass = 0
		aggregated.AverageCyclomaticComplexityPerClass = 0
	}

	if aggregated.AverageMI > 0 {
		aggregated.AverageMI = aggregated.AverageMI / float64(aggregated.NbClasses)
		aggregated.AverageMIwoc = aggregated.AverageMIwoc / float64(aggregated.NbClasses)
		aggregated.AverageMIcw = aggregated.AverageMIcw / float64(aggregated.NbClasses)
	}

	if aggregated.NbMethods > 0 {
		aggregated.AverageLocPerMethod = aggregated.AverageLocPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageClocPerMethod = aggregated.AverageClocPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageLlocPerMethod = aggregated.AverageLlocPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageCyclomaticComplexityPerMethod = aggregated.AverageCyclomaticComplexityPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageMIPerMethod = aggregated.AverageMIPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageMIwocPerMethod = aggregated.AverageMIwocPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageMIcwPerMethod = aggregated.AverageMIcwPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageHalsteadDifficulty = aggregated.AverageHalsteadDifficulty / float64(aggregated.NbClasses)
		aggregated.AverageHalsteadEffort = aggregated.AverageHalsteadEffort / float64(aggregated.NbClasses)
		aggregated.AverageHalsteadVolume = aggregated.AverageHalsteadVolume / float64(aggregated.NbClasses)
		aggregated.AverageHalsteadTime = aggregated.AverageHalsteadTime / float64(aggregated.NbClasses)
		aggregated.AverageHalsteadBugs = aggregated.AverageHalsteadBugs / float64(aggregated.NbClasses)
	}

	// if langage without classes
	if aggregated.NbClasses == 0 {
		aggregated.AverageMI = aggregated.AverageMIPerMethod
		aggregated.AverageMIwoc = aggregated.AverageMIwocPerMethod
		aggregated.AverageMIcw = aggregated.AverageMIcwPerMethod
	}

	// Total locs: increment loc of each file
	aggregated.Loc = 0
	aggregated.Cloc = 0
	aggregated.Lloc = 0

	for _, file := range aggregated.ConcernedFiles {

		if file.LinesOfCode == nil {
			continue
		}

		aggregated.Loc += int(file.LinesOfCode.LinesOfCode)
		aggregated.Cloc += int(file.LinesOfCode.CommentLinesOfCode)
		aggregated.Lloc += int(file.LinesOfCode.LogicalLinesOfCode)

		// Calculate alternate MI using average MI per method when file has no class
		if file.Stmts.StmtClass == nil || len(file.Stmts.StmtClass) == 0 {
			if file.Stmts.Analyze.Maintainability == nil {
				file.Stmts.Analyze.Maintainability = &pb.Maintainability{}
			}

			methods := file.Stmts.StmtFunction
			if methods == nil || len(methods) == 0 {
				continue
			}
			averageForFile := float32(0)
			for _, method := range methods {
				if method.Stmts.Analyze == nil || method.Stmts.Analyze.Maintainability == nil {
					continue
				}
				averageForFile += float32(*method.Stmts.Analyze.Maintainability.MaintainabilityIndex)
			}
			averageForFile = averageForFile / float32(len(methods))
			file.Stmts.Analyze.Maintainability.MaintainabilityIndex = &averageForFile
		}
	}

	// Count commits
	aggregated.ResultOfGitAnalysis = r.gitSummaries
	if aggregated.ResultOfGitAnalysis != nil {
		for _, result := range aggregated.ResultOfGitAnalysis {
			aggregated.CommitCountForPeriod += result.CountCommitsForLanguage
		}
	}

	// Bus factor and other metrics based on aggregated data
	for _, analyzer := range r.analyzers {
		analyzer.Calculate(aggregated)
	}
}

// Add an analyzer to the aggregator
// You can add multiple analyzers. See the example of RiskAnalyzer
func (r *Aggregator) WithAggregateAnalyzer(analyzer AggregateAnalyzer) {
	r.analyzers = append(r.analyzers, analyzer)
}

// Calculate the aggregated data
func (r *Aggregator) calculateSums(file *pb.File, specificAggregation *Aggregated) {
	classes := Engine.GetClassesInFile(file)
	functions := Engine.GetFunctionsInFile(file)

	if specificAggregation.ConcernedFiles == nil {
		specificAggregation.ConcernedFiles = make([]*pb.File, 0)
	}

	specificAggregation.ConcernedFiles = append(specificAggregation.ConcernedFiles, file)

	// Number of classes
	specificAggregation.NbClasses += len(classes)

	// Prepare the file for analysis
	if file.Stmts == nil {
		return
	}

	if file.Stmts.Analyze == nil {
		file.Stmts.Analyze = &pb.Analyze{}
	}

	// lines of code (it should be done in the analayzer. This case occurs only in test, or when the analyzer has issue)
	if file.LinesOfCode == nil && file.Stmts.Analyze.Volume != nil {
		file.LinesOfCode = &pb.LinesOfCode{
			LinesOfCode:        *file.Stmts.Analyze.Volume.Loc,
			CommentLinesOfCode: *file.Stmts.Analyze.Volume.Cloc,
			LogicalLinesOfCode: *file.Stmts.Analyze.Volume.Lloc,
		}
	}

	// Prepare the file for analysis
	if file.Stmts.Analyze == nil {
		file.Stmts.Analyze = &pb.Analyze{}
	}
	if file.Stmts.Analyze.Complexity == nil {
		zero := int32(0)
		file.Stmts.Analyze.Complexity = &pb.Complexity{
			Cyclomatic: &zero,
		}
	}

	// Functions
	for _, function := range functions {

		if function == nil || function.Stmts == nil {
			continue
		}

		specificAggregation.NbMethods++

		// Average cyclomatic complexity per method
		if function.Stmts.Analyze.Complexity != nil {
			if function.Stmts.Analyze.Complexity.Cyclomatic != nil {
				specificAggregation.AverageCyclomaticComplexityPerMethod += float64(*function.Stmts.Analyze.Complexity.Cyclomatic)
			}
		}

		// Average maintainability index per method
		if function.Stmts.Analyze.Maintainability != nil {
			if function.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil && !math.IsNaN(float64(*function.Stmts.Analyze.Maintainability.MaintainabilityIndex)) {
				specificAggregation.AverageMIPerMethod += float64(*function.Stmts.Analyze.Maintainability.MaintainabilityIndex)
				specificAggregation.AverageMIwocPerMethod += float64(*function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments)
				specificAggregation.AverageMIcwPerMethod += float64(*function.Stmts.Analyze.Maintainability.CommentWeight)
			}
		}

		// average lines of code per method
		if function.Stmts.Analyze.Volume != nil {
			if function.Stmts.Analyze.Volume.Loc != nil {
				specificAggregation.AverageLocPerMethod += float64(*function.Stmts.Analyze.Volume.Loc)
			}
			if function.Stmts.Analyze.Volume.Cloc != nil {
				specificAggregation.AverageClocPerMethod += float64(*function.Stmts.Analyze.Volume.Cloc)
			}
			if function.Stmts.Analyze.Volume.Lloc != nil {
				specificAggregation.AverageLlocPerMethod += float64(*function.Stmts.Analyze.Volume.Lloc)
			}
		}
	}

	for _, class := range classes {

		if class == nil || class.Stmts == nil {
			continue
		}

		// Number of classes with code
		//if class.LinesOfCode != nil && class.LinesOfCode.LinesOfCode > 0 {
		specificAggregation.NbClassesWithCode++
		//}

		// Maintainability Index
		if class.Stmts.Analyze.Maintainability != nil {
			if class.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)) {
				specificAggregation.AverageMI += float64(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)
				specificAggregation.AverageMIwoc += float64(*class.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments)
				specificAggregation.AverageMIcw += float64(*class.Stmts.Analyze.Maintainability.CommentWeight)
			}
		}

		// cyclomatic complexity per class
		if class.Stmts.Analyze.Complexity != nil && class.Stmts.Analyze.Complexity.Cyclomatic != nil {
			specificAggregation.AverageCyclomaticComplexityPerClass += float64(*class.Stmts.Analyze.Complexity.Cyclomatic)
			if specificAggregation.MinCyclomaticComplexity == 0 || int(*class.Stmts.Analyze.Complexity.Cyclomatic) < specificAggregation.MinCyclomaticComplexity {
				specificAggregation.MinCyclomaticComplexity = int(*class.Stmts.Analyze.Complexity.Cyclomatic)
			}
			if specificAggregation.MaxCyclomaticComplexity == 0 || int(*class.Stmts.Analyze.Complexity.Cyclomatic) > specificAggregation.MaxCyclomaticComplexity {
				specificAggregation.MaxCyclomaticComplexity = int(*class.Stmts.Analyze.Complexity.Cyclomatic)
			}
		}

		// Halstead
		if class.Stmts.Analyze.Volume != nil {
			if class.Stmts.Analyze.Volume.HalsteadDifficulty != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadDifficulty)) {
				specificAggregation.AverageHalsteadDifficulty += float64(*class.Stmts.Analyze.Volume.HalsteadDifficulty)
				specificAggregation.SumHalsteadDifficulty += float64(*class.Stmts.Analyze.Volume.HalsteadDifficulty)
			}
			if class.Stmts.Analyze.Volume.HalsteadEffort != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadEffort)) {
				specificAggregation.AverageHalsteadEffort += float64(*class.Stmts.Analyze.Volume.HalsteadEffort)
				specificAggregation.SumHalsteadEffort += float64(*class.Stmts.Analyze.Volume.HalsteadEffort)
			}
			if class.Stmts.Analyze.Volume.HalsteadVolume != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadVolume)) {
				specificAggregation.AverageHalsteadVolume += float64(*class.Stmts.Analyze.Volume.HalsteadVolume)
				specificAggregation.SumHalsteadVolume += float64(*class.Stmts.Analyze.Volume.HalsteadVolume)
			}
			if class.Stmts.Analyze.Volume.HalsteadTime != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadTime)) {
				specificAggregation.AverageHalsteadTime += float64(*class.Stmts.Analyze.Volume.HalsteadTime)
				specificAggregation.SumHalsteadTime += float64(*class.Stmts.Analyze.Volume.HalsteadTime)
			}
		}
	}
}
