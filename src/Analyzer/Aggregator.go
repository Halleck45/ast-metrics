package Analyzer

import (
	"math"
	"regexp"
	"sync"

	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Scm"
)

type ProjectAggregated struct {
	ByFile                Aggregated
	ByClass               Aggregated
	Combined              Aggregated
	ByProgrammingLanguage map[string]Aggregated
	ErroredFiles          []*pb.File
	Evaluation            *EvaluationResult
	Comparaison           *ProjectComparaison
}

type Aggregated struct {
	ConcernedFiles []*pb.File
	Comparaison    *Comparaison
	// hashmap of classes, just with the qualified name, used for afferent coupling calculation
	ClassesAfferentCoupling              map[string]int
	NbFiles                              int
	NbFunctions                          int
	NbClasses                            int
	NbClassesWithCode                    int
	NbMethods                            int
	Loc                                  int
	Cloc                                 int
	Lloc                                 int
	AverageMethodsPerClass               float32
	AverageLocPerMethod                  float32
	AverageLlocPerMethod                 float32
	AverageClocPerMethod                 float32
	AverageCyclomaticComplexityPerMethod float32
	AverageCyclomaticComplexityPerClass  float32
	MinCyclomaticComplexity              int
	MaxCyclomaticComplexity              int
	AverageHalsteadDifficulty            float32
	AverageHalsteadEffort                float32
	AverageHalsteadVolume                float32
	AverageHalsteadTime                  float32
	AverageHalsteadBugs                  float32
	SumHalsteadDifficulty                float32
	SumHalsteadEffort                    float32
	SumHalsteadVolume                    float32
	SumHalsteadTime                      float32
	SumHalsteadBugs                      float32
	AverageMI                            float32
	AverageMIwoc                         float32
	AverageMIcw                          float32
	AverageMIPerMethod                   float32
	AverageMIwocPerMethod                float32
	AverageMIcwPerMethod                 float32
	AverageAfferentCoupling              float32
	AverageEfferentCoupling              float32
	AverageInstability                   float32
	CommitCountForPeriod                 int
	CommittedFilesCountForPeriod         int // for example if one commit concerns 10 files, it will be 10
	BusFactor                            int
	TopCommitters                        []TopCommitter
	ResultOfGitAnalysis                  []ResultOfGitAnalysis
	PackageRelations                     map[string]map[string]int // counter of dependencies. Ex: A -> B -> 2
}

type ProjectComparaison struct {
	ByFile                Comparaison
	ByClass               Comparaison
	Combined              Comparaison
	ByProgrammingLanguage map[string]Comparaison
}

type Aggregator struct {
	files             []*pb.File
	projectAggregated ProjectAggregated
	analyzers         []AggregateAnalyzer
	gitSummaries      []ResultOfGitAnalysis
	ComparedFiles     []*pb.File
	ComparedBranch    string
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
	GitRepository           Scm.GitRepository
}

func NewAggregator(files []*pb.File, gitSummaries []ResultOfGitAnalysis) *Aggregator {
	return &Aggregator{
		files:        files,
		gitSummaries: gitSummaries,
	}
}

type AggregateAnalyzer interface {
	Calculate(aggregate *Aggregated)
}

func newAggregated() Aggregated {
	return Aggregated{
		ConcernedFiles:                       make([]*pb.File, 0),
		ClassesAfferentCoupling:              make(map[string]int),
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
		AverageAfferentCoupling:              0,
		AverageEfferentCoupling:              0,
		AverageInstability:                   0,
		AverageMIcwPerMethod:                 0,
		CommitCountForPeriod:                 0,
		ResultOfGitAnalysis:                  nil,
		PackageRelations:                     make(map[string]map[string]int),
	}
}

func (r *Aggregator) Aggregates() ProjectAggregated {

	// We create a new aggregated object for each type of aggregation
	r.projectAggregated = r.executeAggregationOnFiles(r.files)

	// Do the same for the comparaison files (if needed)
	if r.ComparedFiles != nil {
		comparaidAggregated := r.executeAggregationOnFiles(r.ComparedFiles)

		// Compare
		comparaison := ProjectComparaison{}
		comparator := NewComparator(r.ComparedBranch)
		comparaison.Combined = comparator.Compare(r.projectAggregated.Combined, comparaidAggregated.Combined)
		r.projectAggregated.Combined.Comparaison = &comparaison.Combined

		comparaison.ByClass = comparator.Compare(r.projectAggregated.ByClass, comparaidAggregated.ByClass)
		r.projectAggregated.ByClass.Comparaison = &comparaison.ByClass

		comparaison.ByFile = comparator.Compare(r.projectAggregated.ByFile, comparaidAggregated.ByFile)
		r.projectAggregated.ByFile.Comparaison = &comparaison.ByFile

		// By language
		comparaison.ByProgrammingLanguage = make(map[string]Comparaison)
		for lng, byLanguage := range r.projectAggregated.ByProgrammingLanguage {
			if _, ok := comparaidAggregated.ByProgrammingLanguage[lng]; !ok {
				continue
			}
			c := comparator.Compare(byLanguage, comparaidAggregated.ByProgrammingLanguage[lng])
			comparaison.ByProgrammingLanguage[lng] = c

			// assign to the original object (slow, but otherwise we need to change the whole structure ByProgrammingLanguage map)
			// @see https://stackoverflow.com/questions/42605337/cannot-assign-to-struct-field-in-a-map
			// Feel free to change this
			entry := r.projectAggregated.ByProgrammingLanguage[lng]
			entry.Comparaison = &c
			r.projectAggregated.ByProgrammingLanguage[lng] = entry
		}
		r.projectAggregated.Comparaison = &comparaison
	}

	return r.projectAggregated
}

func (r *Aggregator) executeAggregationOnFiles(files []*pb.File) ProjectAggregated {

	// We create a new aggregated object for each type of aggregation
	// ByFile, ByClass, Combined
	projectAggregated := ProjectAggregated{}
	projectAggregated.ByFile = newAggregated()
	projectAggregated.ByClass = newAggregated()
	projectAggregated.Combined = newAggregated()

	// Count files
	projectAggregated.ByClass.NbFiles = len(files)
	projectAggregated.ByFile.NbFiles = len(files)
	projectAggregated.Combined.NbFiles = len(files)

	// Prepare errors
	projectAggregated.ErroredFiles = make([]*pb.File, 0)

	for _, file := range files {

		// Files with errors
		if file.Errors != nil && len(file.Errors) > 0 {
			projectAggregated.ErroredFiles = append(projectAggregated.ErroredFiles, file)
		}

		if file.Stmts == nil {
			continue
		}

		// By language
		if projectAggregated.ByProgrammingLanguage == nil {
			projectAggregated.ByProgrammingLanguage = make(map[string]Aggregated)
		}
		if _, ok := projectAggregated.ByProgrammingLanguage[file.ProgrammingLanguage]; !ok {
			projectAggregated.ByProgrammingLanguage[file.ProgrammingLanguage] = newAggregated()

		}
		byLanguage := projectAggregated.ByProgrammingLanguage[file.ProgrammingLanguage]
		byLanguage.NbFiles++

		// Make calculations: sums of metrics, etc.
		var wg sync.WaitGroup
		wg.Add(4)

		go func() {
			defer wg.Done()
			r.calculateSums(file, &projectAggregated.ByFile)
		}()

		go func() {
			defer wg.Done()
			r.calculateSums(file, &projectAggregated.ByClass)
		}()

		go func() {
			defer wg.Done()
			r.calculateSums(file, &projectAggregated.Combined)
		}()

		go func() {
			defer wg.Done()
			r.calculateSums(file, &byLanguage)
		}()

		wg.Wait()
		projectAggregated.ByProgrammingLanguage[file.ProgrammingLanguage] = byLanguage
	}

	// Consolidate averages
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		r.consolidate(&projectAggregated.ByFile)
	}()

	go func() {
		defer wg.Done()
		r.consolidate(&projectAggregated.ByClass)
	}()

	go func() {
		defer wg.Done()
		r.consolidate(&projectAggregated.Combined)
	}()

	wg.Wait()

	// by language
	wg.Add(len(projectAggregated.ByProgrammingLanguage))
	for lng, byLanguage := range projectAggregated.ByProgrammingLanguage {
		go func(language string, langAggregated Aggregated) {
			defer wg.Done()
			r.consolidate(&langAggregated)
			projectAggregated.ByProgrammingLanguage[language] = langAggregated
		}(lng, byLanguage)
	}
	wg.Wait()

	// Risks
	riskAnalyzer := NewRiskAnalyzer()
	riskAnalyzer.Analyze(projectAggregated)

	return projectAggregated
}

// Consolidate the aggregated data
func (r *Aggregator) consolidate(aggregated *Aggregated) {

	if aggregated.NbClasses > 0 {
		aggregated.AverageMethodsPerClass = float32(aggregated.NbMethods) / float32(aggregated.NbClasses)
		aggregated.AverageCyclomaticComplexityPerClass = aggregated.AverageCyclomaticComplexityPerClass / float32(aggregated.NbClasses)
	} else {
		aggregated.AverageMethodsPerClass = 0
		aggregated.AverageCyclomaticComplexityPerClass = 0
	}

	if aggregated.AverageMI > 0 {
		aggregated.AverageMI = aggregated.AverageMI / float32(aggregated.NbClasses)
		aggregated.AverageMIwoc = aggregated.AverageMIwoc / float32(aggregated.NbClasses)
		aggregated.AverageMIcw = aggregated.AverageMIcw / float32(aggregated.NbClasses)
	}

	if aggregated.AverageInstability > 0 {
		aggregated.AverageEfferentCoupling = aggregated.AverageEfferentCoupling / float32(aggregated.NbClasses)
		aggregated.AverageAfferentCoupling = aggregated.AverageAfferentCoupling / float32(aggregated.NbClasses)
	}

	if aggregated.NbMethods > 0 {
		aggregated.AverageLocPerMethod = aggregated.AverageLocPerMethod / float32(aggregated.NbMethods)
		aggregated.AverageClocPerMethod = aggregated.AverageClocPerMethod / float32(aggregated.NbMethods)
		aggregated.AverageLlocPerMethod = aggregated.AverageLlocPerMethod / float32(aggregated.NbMethods)
		aggregated.AverageCyclomaticComplexityPerMethod = aggregated.AverageCyclomaticComplexityPerMethod / float32(aggregated.NbMethods)
		aggregated.AverageMIPerMethod = aggregated.AverageMIPerMethod / float32(aggregated.NbMethods)
		aggregated.AverageMIwocPerMethod = aggregated.AverageMIwocPerMethod / float32(aggregated.NbMethods)
		aggregated.AverageMIcwPerMethod = aggregated.AverageMIcwPerMethod / float32(aggregated.NbMethods)
		aggregated.AverageHalsteadDifficulty = aggregated.AverageHalsteadDifficulty / float32(aggregated.NbClasses)
		aggregated.AverageHalsteadEffort = aggregated.AverageHalsteadEffort / float32(aggregated.NbClasses)
		aggregated.AverageHalsteadVolume = aggregated.AverageHalsteadVolume / float32(aggregated.NbClasses)
		aggregated.AverageHalsteadTime = aggregated.AverageHalsteadTime / float32(aggregated.NbClasses)
		aggregated.AverageHalsteadBugs = aggregated.AverageHalsteadBugs / float32(aggregated.NbClasses)
	}

	// if langage without classes
	if aggregated.NbClasses == 0 {
		aggregated.AverageMI = aggregated.AverageMIPerMethod
		aggregated.AverageMIwoc = aggregated.AverageMIwocPerMethod
		aggregated.AverageMIcw = aggregated.AverageMIcwPerMethod
		aggregated.AverageInstability = 0
		aggregated.AverageEfferentCoupling = 0
		aggregated.AverageAfferentCoupling = 0
	}

	// Total locs: increment loc of each file
	aggregated.Loc = 0
	aggregated.Cloc = 0
	aggregated.Lloc = 0

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, file := range aggregated.ConcernedFiles {
		wg.Add(1)
		go func(file *pb.File) {
			defer wg.Done()

			if file.LinesOfCode == nil {
				return
			}

			mu.Lock()
			aggregated.Loc += int(file.LinesOfCode.LinesOfCode)
			aggregated.Cloc += int(file.LinesOfCode.CommentLinesOfCode)
			aggregated.Lloc += int(file.LinesOfCode.LogicalLinesOfCode)
			mu.Unlock()

			// Create local variables for file processing
            localFile := &pb.File{
                Stmts: file.Stmts,
                LinesOfCode: file.LinesOfCode,
            }

			// Calculate alternate MI using average MI per method when file has no class
			if len(localFile.Stmts.StmtClass) == 0 {
				if localFile.Stmts.Analyze.Maintainability == nil {
					localFile.Stmts.Analyze.Maintainability = &pb.Maintainability{}
				}

				methods := file.Stmts.StmtFunction
				if len(methods) == 0 {
					return
				}
				averageForFile := float32(0)
				for _, method := range methods {
					if method.Stmts.Analyze == nil || method.Stmts.Analyze.Maintainability == nil {
						continue
					}
					averageForFile += float32(*method.Stmts.Analyze.Maintainability.MaintainabilityIndex)
				}
				averageForFile = averageForFile / float32(len(methods))
				localFile.Stmts.Analyze.Maintainability.MaintainabilityIndex = &averageForFile
			}

			// Update the original file with processed data
			mu.Lock()
			file.Stmts = localFile.Stmts
			mu.Unlock()

			// LOC of file is the sum of all classes and methods
			// That's useful when we navigate over the files instead of the classes
			zero := int32(0)
			loc := int32(0)
			lloc := int32(0)
			cloc := int32(0)

			if file.Stmts.Analyze.Volume == nil {
				file.Stmts.Analyze.Volume = &pb.Volume{
					Lloc: &zero,
					Cloc: &zero,
					Loc:  &zero,
				}
			}

			classes := Engine.GetClassesInFile(file)
			functions := file.Stmts.StmtFunction

			// Initialize file complexity if needed
			if file.Stmts.Analyze.Complexity.Cyclomatic == nil {
				file.Stmts.Analyze.Complexity.Cyclomatic = &zero
			}

			// Process functions
			for _, function := range functions {
				// Handle LOC
				if function.LinesOfCode != nil {
					loc += function.LinesOfCode.LinesOfCode
					lloc += function.LinesOfCode.LogicalLinesOfCode
					cloc += function.LinesOfCode.CommentLinesOfCode
				}

				// Handle complexity
				if function.Stmts.Analyze != nil && function.Stmts.Analyze.Complexity != nil {
					*file.Stmts.Analyze.Complexity.Cyclomatic += *function.Stmts.Analyze.Complexity.Cyclomatic
				}
			}

			// Process classes
			for _, class := range classes {
				// Handle LOC
				if class.LinesOfCode != nil {
					loc += class.LinesOfCode.LinesOfCode
					lloc += class.LinesOfCode.LogicalLinesOfCode
					cloc += class.LinesOfCode.CommentLinesOfCode
				}

				// Handle coupling
				if class.Stmts != nil && class.Stmts.Analyze != nil {
					if class.Stmts.Analyze.Coupling == nil {
						class.Stmts.Analyze.Coupling = &pb.Coupling{
							Efferent: 0,
							Afferent: 0,
						}
					}
					class.Stmts.Analyze.Coupling.Afferent = 0

					if class.Name != nil {
						mu.Lock()
						// if in hashmap
						if _, ok := aggregated.ClassesAfferentCoupling[class.Name.Qualified]; ok {
							class.Stmts.Analyze.Coupling.Afferent = int32(aggregated.ClassesAfferentCoupling[class.Name.Qualified])
							file.Stmts.Analyze.Coupling.Afferent += class.Stmts.Analyze.Coupling.Afferent
						}

						// instability
						if class.Stmts.Analyze.Coupling.Afferent > 0 || class.Stmts.Analyze.Coupling.Efferent > 0 {
							instability := float32(class.Stmts.Analyze.Coupling.Efferent) / float32(class.Stmts.Analyze.Coupling.Efferent+class.Stmts.Analyze.Coupling.Afferent)
							class.Stmts.Analyze.Coupling.Instability = instability
							aggregated.AverageInstability += instability
						}
						mu.Unlock()
					}
				}
			}

			file.Stmts.Analyze.Volume.Loc = &loc
			file.Stmts.Analyze.Volume.Lloc = &lloc
			file.Stmts.Analyze.Volume.Cloc = &cloc

			dependencies := file.Stmts.StmtExternalDependencies

			for _, dependency := range dependencies {
				if dependency == nil {
					continue
				}

				namespaceTo := dependency.Namespace
				namespaceFrom := dependency.From

				// Keep only 2 levels in namespace
				reg := regexp.MustCompile("[^A-Za-z0-9.]+")

				if namespaceFrom == "" || namespaceTo == "" {
					continue
				}

				separator := reg.FindString(namespaceFrom)
				parts := reg.Split(namespaceTo, -1)
				if len(parts) > 2 {
					namespaceTo = parts[0] + separator + parts[1]
				}

				parts = reg.Split(namespaceFrom, -1)
				if len(parts) > 2 {
					namespaceFrom = parts[0] + separator + parts[1]
				}

				// if same, continue
				if namespaceFrom == namespaceTo {
					continue
				}

				// if root namespace, continue
				if namespaceFrom == "" || namespaceTo == "" {
					continue
				}

				mu.Lock()
				// create the map if not exists
				if _, ok := aggregated.PackageRelations[namespaceFrom]; !ok {
					aggregated.PackageRelations[namespaceFrom] = make(map[string]int)
				}

				if _, ok := aggregated.PackageRelations[namespaceFrom][namespaceTo]; !ok {
					aggregated.PackageRelations[namespaceFrom][namespaceTo] = 0
				}

				// increment the counter
				aggregated.PackageRelations[namespaceFrom][namespaceTo]++
				mu.Unlock()
			}
		}(file)
	}
	wg.Wait()

	// Consolidate
	aggregated.AverageInstability = aggregated.AverageInstability / float32(aggregated.NbClasses)

	// Count commits for the period based on `ResultOfGitAnalysis` data
	aggregated.ResultOfGitAnalysis = r.gitSummaries
	if aggregated.ResultOfGitAnalysis != nil {
		var wg sync.WaitGroup
		var mu sync.Mutex

		for _, result := range aggregated.ResultOfGitAnalysis {
			wg.Add(1)
			go func(res ResultOfGitAnalysis) {
				defer wg.Done()
				mu.Lock()
				aggregated.CommitCountForPeriod += res.CountCommitsForLanguage
				mu.Unlock()
			}(result)
		}
		wg.Wait()
	}

	// Bus factor and other metrics based on aggregated data
	wg.Add(len(r.analyzers))
	for _, analyzer := range r.analyzers {
		go func(a AggregateAnalyzer) {
			defer wg.Done()
			a.Calculate(aggregated)
		}(analyzer)
	}
	wg.Wait()
}

// Add an analyzer to the aggregator
// You can add multiple analyzers. See the example of RiskAnalyzer
func (r *Aggregator) WithAggregateAnalyzer(analyzer AggregateAnalyzer) {
	r.analyzers = append(r.analyzers, analyzer)
}

func (r *Aggregator) WithComparaison(allResultsCloned []*pb.File, comparedBranch string) {
	r.ComparedFiles = allResultsCloned
	r.ComparedBranch = comparedBranch
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
		if function.Stmts.Analyze != nil && function.Stmts.Analyze.Complexity != nil {
			if function.Stmts.Analyze.Complexity.Cyclomatic != nil {
				specificAggregation.AverageCyclomaticComplexityPerMethod += float32(*function.Stmts.Analyze.Complexity.Cyclomatic)
			}
		}

		// Average maintainability index per method
		if function.Stmts.Analyze != nil && function.Stmts.Analyze.Maintainability != nil {
			if function.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil && !math.IsNaN(float64(*function.Stmts.Analyze.Maintainability.MaintainabilityIndex)) {
				specificAggregation.AverageMIPerMethod += *function.Stmts.Analyze.Maintainability.MaintainabilityIndex
				specificAggregation.AverageMIwocPerMethod += *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				specificAggregation.AverageMIcwPerMethod += *function.Stmts.Analyze.Maintainability.CommentWeight
			}
		}
		// average lines of code per method
		if function.Stmts.Analyze != nil && function.Stmts.Analyze.Volume != nil {
			if function.Stmts.Analyze.Volume.Loc != nil {
				specificAggregation.AverageLocPerMethod += float32(*function.Stmts.Analyze.Volume.Loc)
			}
			if function.Stmts.Analyze.Volume.Cloc != nil {
				specificAggregation.AverageClocPerMethod += float32(*function.Stmts.Analyze.Volume.Cloc)
			}
			if function.Stmts.Analyze.Volume.Lloc != nil {
				specificAggregation.AverageLlocPerMethod += float32(*function.Stmts.Analyze.Volume.Lloc)
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
				specificAggregation.AverageMI += *class.Stmts.Analyze.Maintainability.MaintainabilityIndex
				specificAggregation.AverageMIwoc += *class.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				specificAggregation.AverageMIcw += *class.Stmts.Analyze.Maintainability.CommentWeight
			}
		}

		// Coupling
		if class.Stmts.Analyze.Coupling != nil {
			specificAggregation.AverageInstability += class.Stmts.Analyze.Coupling.Instability
			specificAggregation.AverageEfferentCoupling += float32(class.Stmts.Analyze.Coupling.Efferent)
			specificAggregation.AverageAfferentCoupling += float32(class.Stmts.Analyze.Coupling.Afferent)
		}

		// cyclomatic complexity per class
		if class.Stmts.Analyze.Complexity != nil && class.Stmts.Analyze.Complexity.Cyclomatic != nil {
			specificAggregation.AverageCyclomaticComplexityPerClass += float32(*class.Stmts.Analyze.Complexity.Cyclomatic)
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
				specificAggregation.AverageHalsteadDifficulty += *class.Stmts.Analyze.Volume.HalsteadDifficulty
				specificAggregation.SumHalsteadDifficulty += *class.Stmts.Analyze.Volume.HalsteadDifficulty
			}
			if class.Stmts.Analyze.Volume.HalsteadEffort != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadEffort)) {
				specificAggregation.AverageHalsteadEffort += *class.Stmts.Analyze.Volume.HalsteadEffort
				specificAggregation.SumHalsteadEffort += *class.Stmts.Analyze.Volume.HalsteadEffort
			}
			if class.Stmts.Analyze.Volume.HalsteadVolume != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadVolume)) {
				specificAggregation.AverageHalsteadVolume += *class.Stmts.Analyze.Volume.HalsteadVolume
				specificAggregation.SumHalsteadVolume += *class.Stmts.Analyze.Volume.HalsteadVolume
			}
			if class.Stmts.Analyze.Volume.HalsteadTime != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Volume.HalsteadTime)) {
				specificAggregation.AverageHalsteadTime += *class.Stmts.Analyze.Volume.HalsteadTime
				specificAggregation.SumHalsteadTime += *class.Stmts.Analyze.Volume.HalsteadTime
			}
		}

		// Coupling
		if class.Stmts.Analyze.Coupling == nil {
			class.Stmts.Analyze.Coupling = &pb.Coupling{
				Efferent: 0,
				Afferent: 0,
			}
		}
		class.Stmts.Analyze.Coupling.Efferent = 0
		uniqueDependencies := make(map[string]bool)
		for _, dependency := range class.Stmts.StmtExternalDependencies {
			dependencyName := dependency.ClassName

			// check if dependency is already in hashmap
			if _, ok := specificAggregation.ClassesAfferentCoupling[dependencyName]; !ok {
				specificAggregation.ClassesAfferentCoupling[dependencyName] = 0
			}
			specificAggregation.ClassesAfferentCoupling[dependencyName]++

			// check if dependency is unique
			if _, ok := uniqueDependencies[dependencyName]; !ok {
				uniqueDependencies[dependencyName] = true
			}
		}

		class.Stmts.Analyze.Coupling.Efferent = int32(len(uniqueDependencies))

		// Add dependencies to file
		if file.Stmts.Analyze.Coupling == nil {
			file.Stmts.Analyze.Coupling = &pb.Coupling{
				Efferent: 0,
				Afferent: 0,
			}
		}
		if file.Stmts.StmtExternalDependencies == nil {
			file.Stmts.StmtExternalDependencies = make([]*pb.StmtExternalDependency, 0)
		}

		file.Stmts.Analyze.Coupling.Efferent += class.Stmts.Analyze.Coupling.Efferent
		file.Stmts.Analyze.Coupling.Afferent += class.Stmts.Analyze.Coupling.Afferent
		file.Stmts.StmtExternalDependencies = append(file.Stmts.StmtExternalDependencies, class.Stmts.StmtExternalDependencies...)
	}

	// consolidate coupling for file
	if file.Stmts.Analyze.Coupling != nil && len(classes) > 0 {
		file.Stmts.Analyze.Coupling.Efferent = file.Stmts.Analyze.Coupling.Efferent / int32(len(classes))
		file.Stmts.Analyze.Coupling.Afferent = file.Stmts.Analyze.Coupling.Afferent / int32(len(classes))
	}

}
