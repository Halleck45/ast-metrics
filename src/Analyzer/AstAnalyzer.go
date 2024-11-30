package Analyzer

import (
	"io/ioutil"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"

	Complexity "github.com/halleck45/ast-metrics/src/Analyzer/Complexity"
	Component "github.com/halleck45/ast-metrics/src/Analyzer/Component"
	Volume "github.com/halleck45/ast-metrics/src/Analyzer/Volume"
	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
	"github.com/yargevad/filepathx"
	"google.golang.org/protobuf/proto"
)

func Start(workdir *Storage.Workdir, progressbar *pterm.SpinnerPrinter) []*pb.File {

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

	root := &ASTNode{children: pbFile.Stmts}

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

	// Ensure structure is complete
	Engine.EnsureNodeTypeIsComplete(pbFile)

	channelResult <- pbFile
	return nil
}
