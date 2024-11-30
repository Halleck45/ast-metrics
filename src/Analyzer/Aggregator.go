package Analyzer

import (
	"math"
	"regexp"
	"runtime"
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

type AggregateResult struct {
	Sum     float64
	Min     float64
	Max     float64
	Avg     float64
	Counter int
}

func NewAggregateResult() AggregateResult {
	return AggregateResult{
		Sum:     0,
		Min:     0,
		Max:     0,
		Avg:     0,
		Counter: 0,
	}
}

type Aggregated struct {
	ProgrammingLanguages map[string]int
	ConcernedFiles       []*pb.File
	ErroredFiles         []*pb.File
	Comparaison          *Comparaison
	// hashmap of classes, just with the qualified name, used for afferent coupling calculation
	ClassesAfferentCoupling                 map[string]int
	NbFiles                                 int
	NbFunctions                             int
	NbClasses                               int
	NbClassesWithCode                       int
	NbMethods                               int
	Loc                                     AggregateResult
	Cloc                                    AggregateResult
	Lloc                                    AggregateResult
	MethodsPerClass                         AggregateResult
	LocPerClass                             AggregateResult
	LocPerMethod                            AggregateResult
	LlocPerMethod                           AggregateResult
	ClocPerMethod                           AggregateResult
	CyclomaticComplexityPerMethod           AggregateResult
	CyclomaticComplexityPerClass            AggregateResult
	HalsteadDifficulty                      AggregateResult
	HalsteadEffort                          AggregateResult
	HalsteadVolume                          AggregateResult
	HalsteadTime                            AggregateResult
	HalsteadBugs                            AggregateResult
	MaintainabilityIndex                    AggregateResult
	MaintainabilityIndexWithoutComments     AggregateResult
	MaintainabilityCommentWeight            AggregateResult
	Instability                             AggregateResult
	EfferentCoupling                        AggregateResult
	AfferentCoupling                        AggregateResult
	MaintainabilityPerMethod                AggregateResult
	MaintainabilityPerMethodWithoutComments AggregateResult
	MaintainabilityCommentWeightPerMethod   AggregateResult
	CommitCountForPeriod                    int
	CommittedFilesCountForPeriod            int
	BusFactor                               int
	TopCommitters                           []TopCommitter
	ResultOfGitAnalysis                     []ResultOfGitAnalysis
	PackageRelations                        map[string]map[string]int // counter of dependencies. Ex: A -> B -> 2
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
		ProgrammingLanguages:                    make(map[string]int),
		ConcernedFiles:                          make([]*pb.File, 0),
		ClassesAfferentCoupling:                 make(map[string]int),
		ErroredFiles:                            make([]*pb.File, 0),
		NbClasses:                               0,
		NbClassesWithCode:                       0,
		NbMethods:                               0,
		NbFunctions:                             0,
		Loc:                                     NewAggregateResult(),
		MethodsPerClass:                         NewAggregateResult(),
		LocPerClass:                             NewAggregateResult(),
		LocPerMethod:                            NewAggregateResult(),
		ClocPerMethod:                           NewAggregateResult(),
		CyclomaticComplexityPerMethod:           NewAggregateResult(),
		CyclomaticComplexityPerClass:            NewAggregateResult(),
		HalsteadEffort:                          NewAggregateResult(),
		HalsteadVolume:                          NewAggregateResult(),
		HalsteadTime:                            NewAggregateResult(),
		HalsteadBugs:                            NewAggregateResult(),
		MaintainabilityIndex:                    NewAggregateResult(),
		MaintainabilityIndexWithoutComments:     NewAggregateResult(),
		MaintainabilityCommentWeight:            NewAggregateResult(),
		Instability:                             NewAggregateResult(),
		EfferentCoupling:                        NewAggregateResult(),
		AfferentCoupling:                        NewAggregateResult(),
		MaintainabilityPerMethod:                NewAggregateResult(),
		MaintainabilityPerMethodWithoutComments: NewAggregateResult(),
		MaintainabilityCommentWeightPerMethod:   NewAggregateResult(),
		CommitCountForPeriod:                    0,
		CommittedFilesCountForPeriod:            0,
		BusFactor:                               0,
		TopCommitters:                           make([]TopCommitter, 0),
		ResultOfGitAnalysis:                     nil,
		PackageRelations:                        make(map[string]map[string]int),
	}
}

// This method is the main entry point to get the aggregated data
// It will:
// - chunk the files by number of processors, to speed up the process
// - map the files to the aggregated object with sums
// - reduce the sums to get the averages
// - map the coupling
// - run the risk analysis
//
// it also computes the comparaison if the compared files are set
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

	projectAggregated := ProjectAggregated{
		ByFile:                newAggregated(),
		ByClass:               newAggregated(),
		Combined:              newAggregated(),
		ByProgrammingLanguage: make(map[string]Aggregated),
		ErroredFiles:          make([]*pb.File, 0),
		Evaluation:            nil,
		Comparaison:           nil,
	}

	// do the sums. Group files by number of processors
	var wg sync.WaitGroup
	numberOfProcessors := runtime.NumCPU()

	// Split the files into chunks
	chunkSize := len(files) / numberOfProcessors
	chunks := make([][]*pb.File, numberOfProcessors)
	for i := 0; i < numberOfProcessors; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numberOfProcessors-1 {
			end = len(files)
		}
		chunks[i] = files[start:end]
	}

	// for each programming language, we create a separeted result
	aggregateByLanguageChunk := make(map[string]Aggregated)
	for _, file := range files {
		if file.ProgrammingLanguage == "" {
			continue
		}
		if _, ok := aggregateByLanguageChunk[file.ProgrammingLanguage]; !ok {
			aggregateByLanguageChunk[file.ProgrammingLanguage] = newAggregated()
		}
	}

	// Create channels for the results
	resultsByClass := make(chan *Aggregated, numberOfProcessors)
	resultsByFile := make(chan *Aggregated, numberOfProcessors)
	resultsByProgrammingLanguage := make(chan *map[string]Aggregated, numberOfProcessors)

	// Deadlock prevention
	mu := sync.Mutex{}

	// Process each chunk of files
	// Please ensure that there is no data race here. If needed, use the mutex
	chunkIndex := 0
	for i := 0; i < numberOfProcessors; i++ {

		wg.Add(1)

		// Reduce results : we want to get sums, and to count calculated values into a AggregateResult
		go func(files []*pb.File) {
			defer wg.Done()

			if len(files) == 0 {
				return
			}

			// Prepare results
			aggregateByFileChunk := newAggregated()
			aggregateByClassChunk := newAggregated()

			// the process deal with its own chunk
			for _, file := range files {
				localFile := file

				// by file
				result := r.mapSums(localFile, aggregateByFileChunk)
				result.ConcernedFiles = append(result.ConcernedFiles, localFile)
				aggregateByFileChunk = result

				// by class
				result = r.mapSums(localFile, aggregateByClassChunk)
				result.ConcernedFiles = append(result.ConcernedFiles, localFile)
				aggregateByClassChunk = result

				// by language
				mu.Lock()
				byLanguage := r.mapSums(localFile, aggregateByLanguageChunk[localFile.ProgrammingLanguage])
				byLanguage.ConcernedFiles = append(byLanguage.ConcernedFiles, localFile)
				aggregateByLanguageChunk[localFile.ProgrammingLanguage] = byLanguage
				mu.Unlock()
			}

			// Send the result to the channels
			resultsByClass <- &aggregateByClassChunk
			resultsByFile <- &aggregateByFileChunk
			resultsByProgrammingLanguage <- &aggregateByLanguageChunk

		}(chunks[chunkIndex])
		chunkIndex++
	}

	wg.Wait()
	close(resultsByClass)
	close(resultsByFile)
	close(resultsByProgrammingLanguage)

	// Now we have chunk of sums. We want to reduce its into a single object
	wg.Add(1)
	go func() {
		defer wg.Done()
		for chunk := range resultsByClass {
			r := r.mergeChunks(projectAggregated.ByClass, chunk)
			projectAggregated.ByClass = r
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for chunk := range resultsByFile {
			r := r.mergeChunks(projectAggregated.ByFile, chunk)
			projectAggregated.ByFile = r
		}
	}()

	wg.Add(1)
	go func() {
		mu.Lock()
		defer wg.Done()
		defer mu.Unlock()

		for chunk := range resultsByProgrammingLanguage {
			for k, v := range *chunk {
				projectAggregated.ByProgrammingLanguage[k] = v
			}
		}
	}()

	wg.Wait()

	// Now  we have sums. We want to reduce metrics and get the averages
	projectAggregated.ByClass = r.reduceMetrics(projectAggregated.ByClass)
	projectAggregated.ByFile = r.reduceMetrics(projectAggregated.ByFile)
	for k, v := range projectAggregated.ByProgrammingLanguage {
		v = r.reduceMetrics(v)
		projectAggregated.ByProgrammingLanguage[k] = v
	}

	// Coupling (should be done separately, to avoid race condition)
	projectAggregated.ByClass = r.mapCoupling(&projectAggregated.ByClass)
	projectAggregated.ByFile = r.mapCoupling(&projectAggregated.ByFile)

	// Risks
	riskAnalyzer := NewRiskAnalyzer()
	riskAnalyzer.Analyze(projectAggregated)

	// For all languages
	projectAggregated.Combined = projectAggregated.ByFile
	projectAggregated.ErroredFiles = projectAggregated.ByFile.ErroredFiles

	return projectAggregated
}

// Add an analyzer to the aggregator
// You can add multiple analyzers. See the example of RiskAnalyzer
func (r *Aggregator) WithAggregateAnalyzer(analyzer AggregateAnalyzer) {
	r.analyzers = append(r.analyzers, analyzer)
}

// Set the files and branch to compare with
func (r *Aggregator) WithComparaison(allResultsCloned []*pb.File, comparedBranch string) {
	r.ComparedFiles = allResultsCloned
	r.ComparedBranch = comparedBranch
}

// Map the sums of a file to the aggregated object
func (r *Aggregator) mapSums(file *pb.File, specificAggregation Aggregated) Aggregated {
	// copy the specific aggregation to new object to avoid side effects
	result := specificAggregation
	result.NbFiles++

	// deal with errors
	if len(file.Errors) > 0 {
		result.ErroredFiles = append(result.ErroredFiles, file)
		return result
	}

	if file.Stmts == nil {
		return result
	}

	classes := Engine.GetClassesInFile(file)
	functions := Engine.GetFunctionsInFile(file)

	// Number of classes
	result.NbClasses += len(classes)

	// Ensure LOC is set
	if file.LinesOfCode == nil {
		if file.Stmts != nil && file.Stmts.Analyze != nil && file.Stmts.Analyze.Volume != nil {
			file.LinesOfCode = &pb.LinesOfCode{
				LinesOfCode:        *file.Stmts.Analyze.Volume.Loc,
				CommentLinesOfCode: *file.Stmts.Analyze.Volume.Cloc,
				LogicalLinesOfCode: *file.Stmts.Analyze.Volume.Lloc,
			}
		} else {
			file.LinesOfCode = &pb.LinesOfCode{
				LinesOfCode:        0,
				CommentLinesOfCode: 0,
				LogicalLinesOfCode: 0,
			}
		}
	}

	result.Loc.Sum += float64(file.LinesOfCode.LinesOfCode)
	result.Loc.Counter++
	result.Cloc.Sum += float64(file.LinesOfCode.CommentLinesOfCode)
	result.Cloc.Counter++
	result.Lloc.Sum += float64(file.LinesOfCode.LogicalLinesOfCode)
	result.Lloc.Counter++

	// Functions
	for _, function := range functions {

		if function == nil || function.Stmts == nil {
			continue
		}

		result.NbMethods++

		// Average cyclomatic complexity per method
		if function.Stmts.Analyze != nil && function.Stmts.Analyze.Complexity != nil {
			if function.Stmts.Analyze.Complexity.Cyclomatic != nil {

				// @todo: only for functions and methods of classes (not interfaces)
				// otherwise, average may be lower than 1
				ccn := float64(*function.Stmts.Analyze.Complexity.Cyclomatic)
				result.CyclomaticComplexityPerMethod.Sum += ccn
				result.CyclomaticComplexityPerMethod.Counter++
				if specificAggregation.CyclomaticComplexityPerMethod.Min == 0 || ccn < specificAggregation.CyclomaticComplexityPerMethod.Min {
					result.CyclomaticComplexityPerMethod.Min = ccn
				}
				if specificAggregation.CyclomaticComplexityPerMethod.Max == 0 || ccn > specificAggregation.CyclomaticComplexityPerMethod.Max {
					result.CyclomaticComplexityPerMethod.Max = ccn
				}
			}
		}

		// Average maintainability index per method
		if function.Stmts.Analyze != nil && function.Stmts.Analyze.Maintainability != nil {
			if function.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil && !math.IsNaN(float64(*function.Stmts.Analyze.Maintainability.MaintainabilityIndex)) {
				result.MaintainabilityIndex.Sum += *function.Stmts.Analyze.Maintainability.MaintainabilityIndex
				result.MaintainabilityIndex.Counter++
				if specificAggregation.MaintainabilityIndex.Min == 0 || *function.Stmts.Analyze.Maintainability.MaintainabilityIndex < specificAggregation.MaintainabilityIndex.Min {
					result.MaintainabilityIndex.Min = *function.Stmts.Analyze.Maintainability.MaintainabilityIndex
				}
				if specificAggregation.MaintainabilityIndex.Max == 0 || *function.Stmts.Analyze.Maintainability.MaintainabilityIndex > specificAggregation.MaintainabilityIndex.Max {
					result.MaintainabilityIndex.Max = *function.Stmts.Analyze.Maintainability.MaintainabilityIndex
				}
			}

			// Maintainability index without comments
			if function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments != nil && !math.IsNaN(float64(*function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments)) {
				result.MaintainabilityIndexWithoutComments.Sum += *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				result.MaintainabilityIndexWithoutComments.Counter++
				if specificAggregation.MaintainabilityIndexWithoutComments.Min == 0 || *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments < specificAggregation.MaintainabilityIndexWithoutComments.Min {
					result.MaintainabilityIndexWithoutComments.Min = *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				}
				if specificAggregation.MaintainabilityIndexWithoutComments.Max == 0 || *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments > specificAggregation.MaintainabilityIndexWithoutComments.Max {
					result.MaintainabilityIndexWithoutComments.Max = *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				}
			}

			// Comment weight
			if function.Stmts.Analyze.Maintainability.CommentWeight != nil && !math.IsNaN(float64(*function.Stmts.Analyze.Maintainability.CommentWeight)) {
				result.MaintainabilityCommentWeight.Sum += *function.Stmts.Analyze.Maintainability.CommentWeight
				result.MaintainabilityCommentWeight.Counter++
				if specificAggregation.MaintainabilityCommentWeight.Min == 0 || *function.Stmts.Analyze.Maintainability.CommentWeight < specificAggregation.MaintainabilityCommentWeight.Min {
					result.MaintainabilityCommentWeight.Min = *function.Stmts.Analyze.Maintainability.CommentWeight
				}
				if specificAggregation.MaintainabilityCommentWeight.Max == 0 || *function.Stmts.Analyze.Maintainability.CommentWeight > specificAggregation.MaintainabilityCommentWeight.Max {
					result.MaintainabilityCommentWeight.Max = *function.Stmts.Analyze.Maintainability.CommentWeight
				}
			}

			// Maintainability index per method
			if function.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil && !math.IsNaN(float64(*function.Stmts.Analyze.Maintainability.MaintainabilityIndex)) {
				result.MaintainabilityPerMethod.Sum += *function.Stmts.Analyze.Maintainability.MaintainabilityIndex
				result.MaintainabilityPerMethod.Counter++
				if specificAggregation.MaintainabilityPerMethod.Min == 0 || *function.Stmts.Analyze.Maintainability.MaintainabilityIndex < specificAggregation.MaintainabilityPerMethod.Min {
					result.MaintainabilityPerMethod.Min = *function.Stmts.Analyze.Maintainability.MaintainabilityIndex
				}
				if specificAggregation.MaintainabilityPerMethod.Max == 0 || *function.Stmts.Analyze.Maintainability.MaintainabilityIndex > specificAggregation.MaintainabilityPerMethod.Max {
					result.MaintainabilityPerMethod.Max = *function.Stmts.Analyze.Maintainability.MaintainabilityIndex
				}
			}

			// Maintainability index per method without comments
			if function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments != nil && !math.IsNaN(float64(*function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments)) {
				result.MaintainabilityPerMethodWithoutComments.Sum += *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				result.MaintainabilityPerMethodWithoutComments.Counter++
				if specificAggregation.MaintainabilityPerMethodWithoutComments.Min == 0 || *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments < specificAggregation.MaintainabilityPerMethodWithoutComments.Min {
					result.MaintainabilityPerMethodWithoutComments.Min = *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				}
				if specificAggregation.MaintainabilityPerMethodWithoutComments.Max == 0 || *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments > specificAggregation.MaintainabilityPerMethodWithoutComments.Max {
					result.MaintainabilityPerMethodWithoutComments.Max = *function.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				}
			}

			// Comment weight per method
			if function.Stmts.Analyze.Maintainability.CommentWeight != nil && !math.IsNaN(float64(*function.Stmts.Analyze.Maintainability.CommentWeight)) {
				result.MaintainabilityCommentWeightPerMethod.Sum += *function.Stmts.Analyze.Maintainability.CommentWeight
				result.MaintainabilityCommentWeightPerMethod.Counter++
				if specificAggregation.MaintainabilityCommentWeightPerMethod.Min == 0 || *function.Stmts.Analyze.Maintainability.CommentWeight < specificAggregation.MaintainabilityCommentWeightPerMethod.Min {
					result.MaintainabilityCommentWeightPerMethod.Min = *function.Stmts.Analyze.Maintainability.CommentWeight
				}
				if specificAggregation.MaintainabilityCommentWeightPerMethod.Max == 0 || *function.Stmts.Analyze.Maintainability.CommentWeight > specificAggregation.MaintainabilityCommentWeightPerMethod.Max {
					result.MaintainabilityCommentWeightPerMethod.Max = *function.Stmts.Analyze.Maintainability.CommentWeight
				}
			}
		}
		// average lines of code per method
		if function.Stmts.Analyze != nil && function.Stmts.Analyze.Volume != nil {
			if function.Stmts.Analyze.Volume.Loc != nil {
				result.LocPerMethod.Sum += float64(*function.Stmts.Analyze.Volume.Loc)
				result.LocPerMethod.Counter++
			}
			if function.Stmts.Analyze.Volume.Cloc != nil {
				result.ClocPerMethod.Sum += float64(*function.Stmts.Analyze.Volume.Cloc)
				result.ClocPerMethod.Counter++
			}
			if function.Stmts.Analyze.Volume.Lloc != nil {
				result.LlocPerMethod.Sum += float64(*function.Stmts.Analyze.Volume.Lloc)
				result.LlocPerMethod.Counter++
			}
		}
	}

	for _, class := range classes {

		if class == nil || class.Stmts == nil {
			continue
		}

		// Number of classes with code
		//if class.LinesOfCode != nil && class.LinesOfCode.LinesOfCode > 0 {
		result.NbClassesWithCode++
		//}

		// Maintainability Index
		if class.Stmts.Analyze.Maintainability != nil {
			if class.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)) {
				result.MaintainabilityIndex.Sum += *class.Stmts.Analyze.Maintainability.MaintainabilityIndex
				result.MaintainabilityIndex.Counter++
				if specificAggregation.MaintainabilityIndex.Min == 0 || *class.Stmts.Analyze.Maintainability.MaintainabilityIndex < specificAggregation.MaintainabilityIndex.Min {
					result.MaintainabilityIndex.Min = *class.Stmts.Analyze.Maintainability.MaintainabilityIndex
				}
				if specificAggregation.MaintainabilityIndex.Max == 0 || *class.Stmts.Analyze.Maintainability.MaintainabilityIndex > specificAggregation.MaintainabilityIndex.Max {
					result.MaintainabilityIndex.Max = *class.Stmts.Analyze.Maintainability.MaintainabilityIndex
				}
			}
			if class.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments != nil && !math.IsNaN(float64(*class.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments)) {
				result.MaintainabilityIndexWithoutComments.Sum += *class.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				result.MaintainabilityIndexWithoutComments.Counter++
				if specificAggregation.MaintainabilityIndexWithoutComments.Min == 0 || *class.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments < specificAggregation.MaintainabilityIndexWithoutComments.Min {
					result.MaintainabilityIndexWithoutComments.Min = *class.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				}
				if specificAggregation.MaintainabilityIndexWithoutComments.Max == 0 || *class.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments > specificAggregation.MaintainabilityIndexWithoutComments.Max {
					result.MaintainabilityIndexWithoutComments.Max = *class.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				}
			}
		}

		// Coupling
		if class.Stmts.Analyze.Coupling != nil {
			result.EfferentCoupling.Sum += float64(class.Stmts.Analyze.Coupling.Efferent)
			result.EfferentCoupling.Counter++
			result.AfferentCoupling.Sum += float64(class.Stmts.Analyze.Coupling.Afferent)
			result.AfferentCoupling.Counter++

			// Instability for class
			if class.Stmts.Analyze.Coupling.Efferent > 0 {
				class.Stmts.Analyze.Coupling.Instability = float64(class.Stmts.Analyze.Coupling.Efferent) / float64(class.Stmts.Analyze.Coupling.Efferent+class.Stmts.Analyze.Coupling.Afferent)
			}
		}

		// cyclomatic complexity per class
		if class.Stmts.Analyze.Complexity != nil && class.Stmts.Analyze.Complexity.Cyclomatic != nil {

			result.CyclomaticComplexityPerClass.Sum += float64(*class.Stmts.Analyze.Complexity.Cyclomatic)
			result.CyclomaticComplexityPerClass.Counter++
			if specificAggregation.CyclomaticComplexityPerClass.Min == 0 || float64(*class.Stmts.Analyze.Complexity.Cyclomatic) < specificAggregation.CyclomaticComplexityPerClass.Min {
				result.CyclomaticComplexityPerClass.Min = float64(*class.Stmts.Analyze.Complexity.Cyclomatic)
			}
			if specificAggregation.CyclomaticComplexityPerClass.Max == 0 || float64(*class.Stmts.Analyze.Complexity.Cyclomatic) > specificAggregation.CyclomaticComplexityPerClass.Max {
				result.CyclomaticComplexityPerClass.Max = float64(*class.Stmts.Analyze.Complexity.Cyclomatic)
			}
		}

		// Halstead
		if class.Stmts.Analyze.Volume != nil {
			if class.Stmts.Analyze.Volume.HalsteadDifficulty != nil && !math.IsNaN(*class.Stmts.Analyze.Volume.HalsteadDifficulty) {
				result.HalsteadDifficulty.Sum += *class.Stmts.Analyze.Volume.HalsteadDifficulty
				result.HalsteadDifficulty.Counter++
			}
			if class.Stmts.Analyze.Volume.HalsteadEffort != nil && !math.IsNaN(*class.Stmts.Analyze.Volume.HalsteadEffort) {
				result.HalsteadEffort.Sum += *class.Stmts.Analyze.Volume.HalsteadEffort
				result.HalsteadEffort.Counter++
			}
			if class.Stmts.Analyze.Volume.HalsteadVolume != nil && !math.IsNaN(*class.Stmts.Analyze.Volume.HalsteadVolume) {
				result.HalsteadVolume.Sum += *class.Stmts.Analyze.Volume.HalsteadVolume
				result.HalsteadVolume.Counter++
			}
			if class.Stmts.Analyze.Volume.HalsteadTime != nil && !math.IsNaN(*class.Stmts.Analyze.Volume.HalsteadTime) {
				result.HalsteadTime.Sum += *class.Stmts.Analyze.Volume.HalsteadTime
				result.HalsteadTime.Counter++
			}
		}

		// Coupling
		if class.Stmts.Analyze.Coupling == nil {
			class.Stmts.Analyze.Coupling = &pb.Coupling{
				Efferent: 0,
				Afferent: 0,
			}
		}

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
	if len(classes) > 0 && file.Stmts.Analyze.Coupling != nil {
		file.Stmts.Analyze.Coupling.Efferent = file.Stmts.Analyze.Coupling.Efferent / int32(len(classes))
		file.Stmts.Analyze.Coupling.Afferent = file.Stmts.Analyze.Coupling.Afferent / int32(len(classes))
	}

	return result
}

// Merge the chunks of files to get the aggregated data (sums)
func (r *Aggregator) mergeChunks(aggregated Aggregated, chunk *Aggregated) Aggregated {

	result := aggregated
	result.ConcernedFiles = append(result.ConcernedFiles, chunk.ConcernedFiles...)
	result.NbFiles += chunk.NbFiles
	result.NbClasses += chunk.NbClasses
	result.NbClassesWithCode += chunk.NbClassesWithCode
	result.NbMethods += chunk.NbMethods

	result.Loc.Sum += chunk.Loc.Sum
	result.Loc.Counter += chunk.Loc.Counter
	result.Cloc.Sum += chunk.Cloc.Sum
	result.Cloc.Counter += chunk.Cloc.Counter
	result.Lloc.Sum += chunk.Lloc.Sum
	result.Lloc.Counter += chunk.Lloc.Counter

	result.MethodsPerClass.Sum += chunk.MethodsPerClass.Sum
	result.MethodsPerClass.Counter += chunk.MethodsPerClass.Counter
	result.LocPerClass.Sum += chunk.LocPerClass.Sum
	result.LocPerClass.Counter += chunk.LocPerClass.Counter
	result.LocPerMethod.Sum += chunk.LocPerMethod.Sum
	result.LocPerMethod.Counter += chunk.LocPerMethod.Counter
	result.CyclomaticComplexityPerMethod.Sum += chunk.CyclomaticComplexityPerMethod.Sum
	result.CyclomaticComplexityPerMethod.Counter += chunk.CyclomaticComplexityPerMethod.Counter

	result.CyclomaticComplexityPerClass.Sum += chunk.CyclomaticComplexityPerClass.Sum
	result.CyclomaticComplexityPerClass.Counter += chunk.CyclomaticComplexityPerClass.Counter

	result.HalsteadDifficulty.Sum += chunk.HalsteadDifficulty.Sum
	result.HalsteadDifficulty.Counter += chunk.HalsteadDifficulty.Counter
	result.HalsteadEffort.Sum += chunk.HalsteadEffort.Sum
	result.HalsteadEffort.Counter += chunk.HalsteadEffort.Counter
	result.HalsteadVolume.Sum += chunk.HalsteadVolume.Sum
	result.HalsteadVolume.Counter += chunk.HalsteadVolume.Counter
	result.HalsteadTime.Sum += chunk.HalsteadTime.Sum
	result.HalsteadTime.Counter += chunk.HalsteadTime.Counter
	result.HalsteadBugs.Sum += chunk.HalsteadBugs.Sum
	result.HalsteadBugs.Counter += chunk.HalsteadBugs.Counter

	result.MaintainabilityIndex.Sum += chunk.MaintainabilityIndex.Sum
	result.MaintainabilityIndex.Counter += chunk.MaintainabilityIndex.Counter
	result.MaintainabilityIndexWithoutComments.Sum += chunk.MaintainabilityIndexWithoutComments.Sum
	result.MaintainabilityIndexWithoutComments.Counter += chunk.MaintainabilityIndexWithoutComments.Counter
	result.MaintainabilityCommentWeight.Sum += chunk.MaintainabilityCommentWeight.Sum
	result.MaintainabilityCommentWeight.Counter += chunk.MaintainabilityCommentWeight.Counter

	result.EfferentCoupling.Sum += chunk.EfferentCoupling.Sum
	result.EfferentCoupling.Counter += chunk.EfferentCoupling.Counter
	result.AfferentCoupling.Sum += chunk.AfferentCoupling.Sum
	result.AfferentCoupling.Counter += chunk.AfferentCoupling.Counter

	result.MaintainabilityPerMethod.Sum += chunk.MaintainabilityPerMethod.Sum
	result.MaintainabilityPerMethod.Counter += chunk.MaintainabilityPerMethod.Counter
	result.MaintainabilityPerMethodWithoutComments.Sum += chunk.MaintainabilityPerMethodWithoutComments.Sum
	result.MaintainabilityPerMethodWithoutComments.Counter += chunk.MaintainabilityPerMethodWithoutComments.Counter
	result.MaintainabilityCommentWeightPerMethod.Sum += chunk.MaintainabilityCommentWeightPerMethod.Sum
	result.MaintainabilityCommentWeightPerMethod.Counter += chunk.MaintainabilityCommentWeightPerMethod.Counter

	result.CommitCountForPeriod += chunk.CommitCountForPeriod
	result.CommittedFilesCountForPeriod += chunk.CommittedFilesCountForPeriod

	result.PackageRelations = make(map[string]map[string]int)
	for k, v := range chunk.PackageRelations {
		result.PackageRelations[k] = v
	}

	result.ErroredFiles = append(result.ErroredFiles, chunk.ErroredFiles...)

	return result
}

// Reduce the sums to get the averages
func (r *Aggregator) reduceMetrics(aggregated Aggregated) Aggregated {
	// here we reduce metrics by averaging them
	result := aggregated
	if result.Loc.Counter > 0 {
		result.Loc.Avg = result.Loc.Sum / float64(result.Loc.Counter)
	}
	if result.Cloc.Counter > 0 {
		result.Cloc.Avg = result.Cloc.Sum / float64(result.Cloc.Counter)
	}
	if result.Lloc.Counter > 0 {
		result.Lloc.Avg = result.Lloc.Sum / float64(result.Lloc.Counter)
	}
	if result.MethodsPerClass.Counter > 0 {
		result.MethodsPerClass.Avg = result.MethodsPerClass.Sum / float64(result.MethodsPerClass.Counter)
	}
	if result.LocPerClass.Counter > 0 {
		result.LocPerClass.Avg = result.LocPerClass.Sum / float64(result.LocPerClass.Counter)
	}
	if result.ClocPerMethod.Counter > 0 {
		result.ClocPerMethod.Avg = result.ClocPerMethod.Sum / float64(result.ClocPerMethod.Counter)
	}
	if result.LlocPerMethod.Counter > 0 {
		result.LlocPerMethod.Avg = result.LlocPerMethod.Sum / float64(result.LlocPerMethod.Counter)
	}
	if result.LocPerMethod.Counter > 0 {
		result.LocPerMethod.Avg = result.LocPerMethod.Sum / float64(result.LocPerMethod.Counter)
	}
	if result.CyclomaticComplexityPerMethod.Counter > 0 {
		result.CyclomaticComplexityPerMethod.Avg = result.CyclomaticComplexityPerMethod.Sum / float64(result.CyclomaticComplexityPerMethod.Counter)
	}
	if result.CyclomaticComplexityPerClass.Counter > 0 {
		result.CyclomaticComplexityPerClass.Avg = result.CyclomaticComplexityPerClass.Sum / float64(result.CyclomaticComplexityPerClass.Counter)
	}
	if result.HalsteadDifficulty.Counter > 0 {
		result.HalsteadDifficulty.Avg = result.HalsteadDifficulty.Sum / float64(result.HalsteadDifficulty.Counter)
	}
	if result.HalsteadEffort.Counter > 0 {
		result.HalsteadEffort.Avg = result.HalsteadEffort.Sum / float64(result.HalsteadEffort.Counter)
	}
	if result.HalsteadVolume.Counter > 0 {
		result.HalsteadVolume.Avg = result.HalsteadVolume.Sum / float64(result.HalsteadVolume.Counter)
	}
	if result.HalsteadTime.Counter > 0 {
		result.HalsteadTime.Avg = result.HalsteadTime.Sum / float64(result.HalsteadTime.Counter)
	}
	if result.MaintainabilityIndex.Counter > 0 {
		result.MaintainabilityIndex.Avg = result.MaintainabilityIndex.Sum / float64(result.MaintainabilityIndex.Counter)
	}
	if result.MaintainabilityIndexWithoutComments.Counter > 0 {
		result.MaintainabilityIndexWithoutComments.Avg = result.MaintainabilityIndexWithoutComments.Sum / float64(result.MaintainabilityIndexWithoutComments.Counter)
	}
	if result.MaintainabilityCommentWeight.Counter > 0 {
		result.MaintainabilityCommentWeight.Avg = result.MaintainabilityCommentWeight.Sum / float64(result.MaintainabilityCommentWeight.Counter)
	}
	if result.MaintainabilityPerMethod.Counter > 0 {
		result.MaintainabilityPerMethod.Avg = result.MaintainabilityPerMethod.Sum / float64(result.MaintainabilityPerMethod.Counter)
	}
	if result.MaintainabilityPerMethodWithoutComments.Counter > 0 {
		result.MaintainabilityPerMethodWithoutComments.Avg = result.MaintainabilityPerMethodWithoutComments.Sum / float64(result.MaintainabilityPerMethodWithoutComments.Counter)
	}
	if result.MaintainabilityCommentWeightPerMethod.Counter > 0 {
		result.MaintainabilityCommentWeightPerMethod.Avg = result.MaintainabilityCommentWeightPerMethod.Sum / float64(result.MaintainabilityCommentWeightPerMethod.Counter)
	}

	if result.EfferentCoupling.Counter > 0 {
		result.EfferentCoupling.Avg = result.EfferentCoupling.Sum / float64(result.EfferentCoupling.Counter)
	}
	if result.AfferentCoupling.Counter > 0 {
		result.AfferentCoupling.Avg = result.AfferentCoupling.Sum / float64(result.AfferentCoupling.Counter)
	}

	// afferent coupling
	// Ce / (Ce + Ca)
	if result.AfferentCoupling.Counter > 0 {
		result.Instability.Avg = result.EfferentCoupling.Sum / result.AfferentCoupling.Sum
	}

	// Count commits for the period based on `ResultOfGitAnalysis` data
	result.ResultOfGitAnalysis = r.gitSummaries
	if result.ResultOfGitAnalysis != nil {
		for _, gitAnalysis := range result.ResultOfGitAnalysis {
			result.CommitCountForPeriod += gitAnalysis.CountCommitsForLanguage
		}
	}

	// Bus factor and other metrics based on aggregated data
	for _, analyzer := range r.analyzers {
		analyzer.Calculate(&result)
	}

	return result
}

// Map the coupling to get the package relations and the afferent coupling
func (r *Aggregator) mapCoupling(aggregated *Aggregated) Aggregated {
	result := *aggregated
	reg := regexp.MustCompile("[^A-Za-z0-9.]+")

	for _, file := range aggregated.ConcernedFiles {
		classes := Engine.GetClassesInFile(file)

		for _, class := range classes {

			if class == nil {
				continue
			}

			// dependencies
			dependencies := file.Stmts.StmtExternalDependencies

			for _, dependency := range dependencies {
				if dependency == nil {
					continue
				}

				namespaceTo := dependency.Namespace
				namespaceFrom := dependency.From

				if namespaceFrom == "" || namespaceTo == "" {
					continue
				}

				// Keep only 2 levels in namespace
				separator := reg.FindString(namespaceFrom)
				parts := reg.Split(namespaceTo, -1)
				if len(parts) > 2 {
					namespaceTo = parts[0] + separator + parts[1]
				}

				if namespaceFrom == "" || namespaceTo == "" {
					continue
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

				// create the map if not exists
				if _, ok := result.PackageRelations[namespaceFrom]; !ok {
					result.PackageRelations[namespaceFrom] = make(map[string]int)
				}

				if _, ok := result.PackageRelations[namespaceFrom][namespaceTo]; !ok {
					result.PackageRelations[namespaceFrom][namespaceTo] = 0
				}

				// increment the counter
				result.PackageRelations[namespaceFrom][namespaceTo]++
			}

			uniqueDependencies := make(map[string]bool)
			for _, dependency := range class.Stmts.StmtExternalDependencies {
				dependencyName := dependency.ClassName

				// check if dependency is already in hashmap
				if _, ok := result.ClassesAfferentCoupling[dependencyName]; !ok {
					result.ClassesAfferentCoupling[dependencyName] = 0
				}
				result.ClassesAfferentCoupling[dependencyName]++

				// check if dependency is unique
				if _, ok := uniqueDependencies[dependencyName]; !ok {
					uniqueDependencies[dependencyName] = true
				}
			}

			if class.Stmts.Analyze.Coupling == nil {
				class.Stmts.Analyze.Coupling = &pb.Coupling{
					Efferent: 0,
					Afferent: 0,
				}
			}
			class.Stmts.Analyze.Coupling.Efferent = int32(len(uniqueDependencies))

			// Afferent coupling
			class.Stmts.Analyze.Coupling.Afferent = int32(len(class.Stmts.StmtExternalDependencies))

			// Increment result
			result.EfferentCoupling.Sum += float64(class.Stmts.Analyze.Coupling.Efferent)
			result.EfferentCoupling.Counter++
			result.AfferentCoupling.Sum += float64(class.Stmts.Analyze.Coupling.Afferent)
			result.AfferentCoupling.Counter++
		}
	}

	// Afferent coupling
	// Ce / (Ce + Ca)
	if result.AfferentCoupling.Counter > 0 {
		result.Instability.Avg = result.EfferentCoupling.Sum / result.AfferentCoupling.Sum
	}
	result.EfferentCoupling.Avg = result.EfferentCoupling.Sum / float64(result.EfferentCoupling.Counter)
	result.AfferentCoupling.Avg = result.AfferentCoupling.Sum / float64(result.AfferentCoupling.Counter)

	return result
}
