package analyzer

import (
	"io/ioutil"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"

	Complexity "github.com/halleck45/ast-metrics/internal/analyzer/complexity"
	Component "github.com/halleck45/ast-metrics/internal/analyzer/component"
	Volume "github.com/halleck45/ast-metrics/internal/analyzer/volume"
	engine "github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	storage "github.com/halleck45/ast-metrics/internal/storage"
	"github.com/pterm/pterm"
	"github.com/yargevad/filepathx"
	"google.golang.org/protobuf/proto"
)

func Start(workdir *storage.Workdir, progressbar *pterm.SpinnerPrinter) []*pb.File {

	// List all ASTs files (*.bin) in the workdir
	astFiles, err := filepathx.Glob(workdir.Path() + "/**/*.bin")
	if err != nil {
		panic(err)
	}

	// Wait for end of all goroutines
	var wg sync.WaitGroup

	// store results
	// channel should have value
	// https://stackoverflow.com/questions/58743038/why-does-this-goroutine-not-call-wg-done
	channelResult := make(chan *pb.File, len(astFiles))

	var nbParsingFiles atomic.Uint64

	// analyze each AST file running the runAnalysis function
	numWorkers := runtime.NumCPU()
	filesChan := make(chan string, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			for file := range filesChan {
				go func(file string) {
					defer wg.Done()
					nbParsingFiles.Add(1)

					executeFileAnalysis(file, channelResult)

					details := strconv.Itoa(int(nbParsingFiles.Load())) + "/" + strconv.Itoa(len(astFiles))

					if progressbar != nil {
						progressbar.UpdateText("Analyzing (" + details + ")")
					}
				}(file)
			}
		}()
	}

	for _, file := range astFiles {
		wg.Add(1)
		filesChan <- file
	}

	wg.Wait()
	close(filesChan)

	if progressbar != nil {
		progressbar.Info("AST Analysis finished")
	}

	// Convert it to slice of pb.File
	allResults := make([]*pb.File, 0, len(astFiles))
	for i := 0; i < len(astFiles); i++ {
		allResults = append(allResults, <-channelResult)
	}
	defer close(channelResult)
	return allResults
}

func executeFileAnalysis(file string, channelResult chan<- *pb.File) error {

	pbFile := &pb.File{}

	// load AST via ProtoBuf (using NodeType package)
	in, err := ioutil.ReadFile(file)
	if err != nil {
		if pbFile.Errors == nil {
			pbFile.Errors = make([]string, 0)
		}
		pbFile.Errors = append(pbFile.Errors, "Error reading file: "+err.Error())
		channelResult <- pbFile
		return err
	}

	// if file is empty, return
	if len(in) == 0 {
		if pbFile.Errors == nil {
			pbFile.Errors = make([]string, 0)
		}
		pbFile.Errors = append(pbFile.Errors, "File is empty: "+file)
		channelResult <- pbFile
		return err
	}

	if err := proto.Unmarshal(in, pbFile); err != nil {
		if pbFile.Errors == nil {
			pbFile.Errors = make([]string, 0)
		}
		pbFile.Errors = append(pbFile.Errors, "Failed to parse address pbFile ("+file+"): "+err.Error())
		channelResult <- pbFile
		return err
	}

	// Start analyse
	AnalyzeFile(pbFile)

	channelResult <- pbFile
	return nil
}

func AnalyzeFile(file *pb.File) {
	root := &ASTNode{children: file.Stmts}

	// register visitors
	cyclomaticVisitor := &Complexity.CyclomaticComplexityVisitor{}
	root.Accept(cyclomaticVisitor)

	locVisitor := &Volume.LocVisitor{}
	root.Accept(locVisitor)

	halsteadVisitor := &Volume.HalsteadMetricsVisitor{}
	root.Accept(halsteadVisitor)

	maintainabilityIndexVisitor := &Component.MaintainabilityIndexVisitor{}
	root.Accept(maintainabilityIndexVisitor)

	// visit AST
	root.Visit()

	// // After visitors, ensure file-level Volume metrics exist and are coherent
	// if file != nil {
	// 	if file.Stmts == nil {
	// 		file.Stmts = &pb.Stmts{}
	// 	}
	// 	if file.Stmts.Analyze == nil {
	// 		file.Stmts.Analyze = &pb.Analyze{}
	// 	}
	// 	if file.Stmts.Analyze.Volume == nil {
	// 		file.Stmts.Analyze.Volume = &pb.Volume{}
	// 	}
	// 	// Prefer full file LOC for file-level Loc
	// 	if file.LinesOfCode != nil {
	// 		v := file.LinesOfCode.LinesOfCode
	// 		file.Stmts.Analyze.Volume.Loc = &v
	// 	}
	// 	// Aggregate LLOC/CLOC from functions if missing or zero
	// 	var sumLloc int32
	// 	var sumCloc int32
	// 	funcs := engine.GetFunctionsInFile(file)
	// 	for _, fn := range funcs {
	// 		if fn == nil || fn.LinesOfCode == nil {
	// 			continue
	// 		}
	// 		ll := fn.LinesOfCode.LogicalLinesOfCode
	// 		if ll == 0 {
	// 			// Consider at least one logical line per function when compacted on one line
	// 			ll = 1
	// 		}
	// 		sumLloc += ll
	// 		sumCloc += fn.LinesOfCode.CommentLinesOfCode
	// 	}
	// 	if file.Stmts.Analyze.Volume.Lloc == nil || *file.Stmts.Analyze.Volume.Lloc == 0 {
	// 		file.Stmts.Analyze.Volume.Lloc = &sumLloc
	// 	}
	// 	if file.Stmts.Analyze.Volume.Cloc == nil || *file.Stmts.Analyze.Volume.Cloc == 0 {
	// 		file.Stmts.Analyze.Volume.Cloc = &sumCloc
	// 	}
	// }

	// Recompute Maintainability Index at file level after adjustments
	mi2 := &Component.MaintainabilityIndexVisitor{}
	mi2.Calculate(file.Stmts)

	// Ensure structure is complete
	engine.EnsureNodeTypeIsComplete(file)
}
