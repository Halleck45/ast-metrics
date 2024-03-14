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
	BusFactor                            int
	TopCommitters                        []TopCommitter
}

type Aggregator struct {
	files             []*pb.File
	projectAggregated ProjectAggregated
	analyzers         []AggregateAnalyzer
}

type TopCommitter struct {
	Name  string
	Count int
}

func NewAggregator(files []*pb.File) *Aggregator {
	return &Aggregator{
		files:             files,
		projectAggregated: ProjectAggregated{},
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
	}
}

func (r *Aggregator) Aggregates() ProjectAggregated {
	files := r.files

	aggregated := newAggregated()

	r.projectAggregated.ByFile = aggregated
	r.projectAggregated.ByClass = aggregated
	r.projectAggregated.Combined = aggregated

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

		// files
		r.projectAggregated.Combined.ConcernedFiles = append(r.projectAggregated.Combined.ConcernedFiles, file)
		// by programming language
		if byLanguage.ConcernedFiles == nil {
			byLanguage.ConcernedFiles = make([]*pb.File, 0)
		}
		byLanguage.ConcernedFiles = append(byLanguage.ConcernedFiles, file)

		// function included directly in file
		r.calculate(file.Stmts, &r.projectAggregated.ByFile)
		r.calculate(file.Stmts, &r.projectAggregated.Combined)
		r.calculate(file.Stmts, &byLanguage)

		// classes
		for _, stmt := range file.Stmts.StmtClass {
			r.projectAggregated.ByClass.NbClasses++
			r.projectAggregated.ByFile.NbClasses++
			r.projectAggregated.Combined.NbClasses++
			byLanguage.NbClasses++

			if stmt.LinesOfCode != nil && stmt.LinesOfCode.LinesOfCode > 0 {
				r.projectAggregated.ByClass.NbClassesWithCode++
				r.projectAggregated.ByFile.NbClassesWithCode++
				r.projectAggregated.Combined.NbClassesWithCode++
				byLanguage.NbClassesWithCode++
			}

			r.calculate(stmt.Stmts, &r.projectAggregated.ByClass)
			r.calculate(stmt.Stmts, &r.projectAggregated.Combined)
			r.calculate(stmt.Stmts, &byLanguage)
		}

		// classes in namespace
		for _, stmt := range file.Stmts.StmtNamespace {
			for _, s := range stmt.Stmts.StmtClass {
				r.projectAggregated.ByClass.NbClasses++
				r.projectAggregated.ByFile.NbClasses++
				r.projectAggregated.Combined.NbClasses++
				byLanguage.NbClasses++

				if s.LinesOfCode != nil && s.LinesOfCode.LinesOfCode > 0 {
					r.projectAggregated.ByClass.NbClassesWithCode++
					r.projectAggregated.ByFile.NbClassesWithCode++
					r.projectAggregated.Combined.NbClassesWithCode++
					byLanguage.NbClassesWithCode++
				}

				r.calculate(s.Stmts, &r.projectAggregated.ByClass)
				r.calculate(s.Stmts, &r.projectAggregated.Combined)
				r.calculate(s.Stmts, &byLanguage)
			}
		}

		// functions in namespace
		for _, stmt := range file.Stmts.StmtNamespace {
			for _, s := range stmt.Stmts.StmtFunction {
				r.projectAggregated.ByFile.NbMethods++
				r.projectAggregated.ByClass.NbMethods++
				r.projectAggregated.Combined.NbMethods++
				byLanguage.NbMethods++
				r.calculate(s.Stmts, &r.projectAggregated.ByFile)
				r.calculate(s.Stmts, &r.projectAggregated.Combined)
				r.calculate(s.Stmts, &byLanguage)
			}
		}

		// update by language
		r.projectAggregated.ByProgrammingLanguage[file.ProgrammingLanguage] = byLanguage

		// @todo: make a visitor method to call file itself
		functions := Engine.GetFunctionsInFile(file)
		for _, function := range functions {
			if file.Stmts == nil {
				continue
			}

			if file.Stmts.Analyze == nil {
				zero := int32(0)
				file.Stmts.Analyze = &pb.Analyze{
					Complexity: &pb.Complexity{
						Cyclomatic: &zero,
					},
				}
			}

			// calculate sum of cyclomatic complexity
			if function.Stmts.Analyze.Complexity != nil && function.Stmts.Analyze.Complexity.Cyclomatic != nil {
				*file.Stmts.Analyze.Complexity.Cyclomatic = *file.Stmts.Analyze.Complexity.Cyclomatic + *function.Stmts.Analyze.Complexity.Cyclomatic
			}
		}

	}

	// averages
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

func (r *Aggregator) calculate(stmts *pb.Stmts, specificAggregation *Aggregated) {
	// methods per class
	if stmts == nil {
		return
	}

	if stmts.Analyze == nil {
		stmts.Analyze = &pb.Analyze{}
	}

	if stmts.StmtFunction != nil {
		specificAggregation.NbMethods += len(stmts.StmtFunction)
	}
	// class per file
	if stmts.StmtClass != nil {
		specificAggregation.NbClasses += len(stmts.StmtClass)
	}

	// Average cyclomatic complexity per method
	if stmts.StmtFunction != nil {
		for _, method := range stmts.StmtFunction {
			if method.Stmts == nil {
				continue
			}
			if method.Stmts.Analyze.Complexity != nil {
				if method.Stmts.Analyze.Complexity.Cyclomatic != nil {
					specificAggregation.AverageCyclomaticComplexityPerMethod += float64(*method.Stmts.Analyze.Complexity.Cyclomatic)
				}
			}
		}
	}

	// Average maintainability index per method
	if stmts.StmtFunction != nil {
		for _, method := range stmts.StmtFunction {
			if method.Stmts == nil {
				continue
			}

			if (method.Stmts.Analyze.Maintainability == nil) ||
				(method.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil) || math.IsNaN(float64(*method.Stmts.Analyze.Maintainability.MaintainabilityIndex)) {
				continue
			}

			specificAggregation.AverageMIPerMethod += float64(*method.Stmts.Analyze.Maintainability.MaintainabilityIndex)
			specificAggregation.AverageMIwocPerMethod += float64(*method.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments)
			specificAggregation.AverageMIcwPerMethod += float64(*method.Stmts.Analyze.Maintainability.CommentWeight)
		}
	}

	// lines of code
	if stmts.Analyze.Volume == nil {
		stmts.Analyze.Volume = &pb.Volume{
			Loc:  new(int32),
			Cloc: new(int32),
			Lloc: new(int32),
		}
	}
	specificAggregation.Loc += int(*stmts.Analyze.Volume.Loc)
	specificAggregation.Cloc += int(*stmts.Analyze.Volume.Cloc)
	specificAggregation.Lloc += int(*stmts.Analyze.Volume.Lloc)

	// average lines of code per method
	if stmts.StmtFunction != nil {
		for _, method := range stmts.StmtFunction {

			if method.Stmts == nil {
				continue
			}

			if method.Stmts.Analyze.Volume != nil {
				if method.Stmts.Analyze.Volume.Loc != nil {
					specificAggregation.AverageLocPerMethod += float64(*method.Stmts.Analyze.Volume.Loc)
				}
				if method.Stmts.Analyze.Volume.Cloc != nil {
					specificAggregation.AverageClocPerMethod += float64(*method.Stmts.Analyze.Volume.Cloc)
				}
				if method.Stmts.Analyze.Volume.Lloc != nil {
					specificAggregation.AverageLlocPerMethod += float64(*method.Stmts.Analyze.Volume.Lloc)
				}
			}
		}
	}

	// Maintainability Index
	if stmts.Analyze.Maintainability != nil {
		if stmts.Analyze.Maintainability.MaintainabilityIndex != nil && !math.IsNaN(float64(*stmts.Analyze.Maintainability.MaintainabilityIndex)) {
			specificAggregation.AverageMI += float64(*stmts.Analyze.Maintainability.MaintainabilityIndex)
			specificAggregation.AverageMIwoc += float64(*stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments)
			specificAggregation.AverageMIcw += float64(*stmts.Analyze.Maintainability.CommentWeight)
		}
	}

	// cyclomatic complexity per classq
	if stmts.Analyze.Complexity != nil && stmts.Analyze.Complexity.Cyclomatic != nil {
		specificAggregation.AverageCyclomaticComplexityPerClass += float64(*stmts.Analyze.Complexity.Cyclomatic)
		if specificAggregation.MinCyclomaticComplexity == 0 || int(*stmts.Analyze.Complexity.Cyclomatic) < specificAggregation.MinCyclomaticComplexity {
			specificAggregation.MinCyclomaticComplexity = int(*stmts.Analyze.Complexity.Cyclomatic)
		}
		if specificAggregation.MaxCyclomaticComplexity == 0 || int(*stmts.Analyze.Complexity.Cyclomatic) > specificAggregation.MaxCyclomaticComplexity {
			specificAggregation.MaxCyclomaticComplexity = int(*stmts.Analyze.Complexity.Cyclomatic)
		}
	}

	// Halstead
	if stmts.Analyze.Volume != nil {
		if stmts.Analyze.Volume.HalsteadDifficulty != nil && !math.IsNaN(float64(*stmts.Analyze.Volume.HalsteadDifficulty)) {
			specificAggregation.AverageHalsteadDifficulty += float64(*stmts.Analyze.Volume.HalsteadDifficulty)
			specificAggregation.SumHalsteadDifficulty += float64(*stmts.Analyze.Volume.HalsteadDifficulty)
		}
		if stmts.Analyze.Volume.HalsteadEffort != nil && !math.IsNaN(float64(*stmts.Analyze.Volume.HalsteadEffort)) {
			specificAggregation.AverageHalsteadEffort += float64(*stmts.Analyze.Volume.HalsteadEffort)
			specificAggregation.SumHalsteadEffort += float64(*stmts.Analyze.Volume.HalsteadEffort)
		}
		if stmts.Analyze.Volume.HalsteadVolume != nil && !math.IsNaN(float64(*stmts.Analyze.Volume.HalsteadVolume)) {
			specificAggregation.AverageHalsteadVolume += float64(*stmts.Analyze.Volume.HalsteadVolume)
			specificAggregation.SumHalsteadVolume += float64(*stmts.Analyze.Volume.HalsteadVolume)
		}
		if stmts.Analyze.Volume.HalsteadTime != nil && !math.IsNaN(float64(*stmts.Analyze.Volume.HalsteadTime)) {
			specificAggregation.AverageHalsteadTime += float64(*stmts.Analyze.Volume.HalsteadTime)
			specificAggregation.SumHalsteadTime += float64(*stmts.Analyze.Volume.HalsteadTime)
		}
	}
}

func (r *Aggregator) consolidate(aggregated *Aggregated) {

	if aggregated.NbClasses > 0 {
		aggregated.AverageMethodsPerClass = float64(aggregated.NbMethods) / float64(aggregated.NbClasses)
		aggregated.AverageCyclomaticComplexityPerClass = aggregated.AverageCyclomaticComplexityPerClass / float64(aggregated.NbClasses)
	} else {
		aggregated.AverageMethodsPerClass = 0
		aggregated.AverageCyclomaticComplexityPerClass = 0
	}

	if aggregated.NbMethods > 0 {
		aggregated.AverageHalsteadDifficulty = aggregated.AverageHalsteadDifficulty / float64(aggregated.NbClasses)
		aggregated.AverageHalsteadEffort = aggregated.AverageHalsteadEffort / float64(aggregated.NbClasses)
		aggregated.AverageHalsteadVolume = aggregated.AverageHalsteadVolume / float64(aggregated.NbClasses)
		aggregated.AverageHalsteadTime = aggregated.AverageHalsteadTime / float64(aggregated.NbClasses)
	}

	if aggregated.NbMethods > 0 {
		aggregated.AverageLocPerMethod = aggregated.AverageLocPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageClocPerMethod = aggregated.AverageClocPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageLlocPerMethod = aggregated.AverageLlocPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageCyclomaticComplexityPerMethod = aggregated.AverageCyclomaticComplexityPerMethod / float64(aggregated.NbMethods)
	}

	if aggregated.AverageMI > 0 {
		aggregated.AverageMI = aggregated.AverageMI / float64(aggregated.NbClassesWithCode)
		aggregated.AverageMIwoc = aggregated.AverageMIwoc / float64(aggregated.NbClassesWithCode)
		aggregated.AverageMIcw = aggregated.AverageMIcw / float64(aggregated.NbClassesWithCode)
	}

	if aggregated.AverageMIPerMethod > 0 {
		aggregated.AverageMIPerMethod = aggregated.AverageMIPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageMIwocPerMethod = aggregated.AverageMIwocPerMethod / float64(aggregated.NbMethods)
		aggregated.AverageMIcwPerMethod = aggregated.AverageMIcwPerMethod / float64(aggregated.NbMethods)
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
		aggregated.Loc += int(*&file.LinesOfCode.LinesOfCode)
		aggregated.Cloc += int(*&file.LinesOfCode.CommentLinesOfCode)
		aggregated.Lloc += int(*&file.LinesOfCode.LogicalLinesOfCode)

		// count of commits
		if file.Commits != nil {
			aggregated.CommitCountForPeriod += int(file.Commits.Count)
		}

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

	// Bus factor and other metrics based on aggregated data
	for _, analyzer := range r.analyzers {
		analyzer.Calculate(aggregated)
	}
}

func (r *Aggregator) WithAggregateAnalyzer(analyzer AggregateAnalyzer) {
	r.analyzers = append(r.analyzers, analyzer)
}
