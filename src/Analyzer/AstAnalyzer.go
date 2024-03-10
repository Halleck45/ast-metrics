package Analyzer

import (
	"io/ioutil"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	Complexity "github.com/halleck45/ast-metrics/src/Analyzer/Complexity"
	Component "github.com/halleck45/ast-metrics/src/Analyzer/Component"
	Volume "github.com/halleck45/ast-metrics/src/Analyzer/Volume"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
	"github.com/yargevad/filepathx"
	"google.golang.org/protobuf/proto"
)

func Start(progressbar *pterm.SpinnerPrinter) []*pb.File {

	workdir := Storage.Path()
	// List all ASTs files (*.bin) in the workdir
	astFiles, err := filepathx.Glob(workdir + "/**/*.bin")
	if err != nil {
		panic(err)
	}

	// Wait for end of all goroutines
	var wg sync.WaitGroup

	// store results
	// channel should have value
	// https://stackoverflow.com/questions/58743038/why-does-this-goroutine-not-call-wg-done
	channelResult := make(chan *pb.File, len(astFiles))

	nbParsingFiles := 0
	// in parallel, 8 process max, analyze each AST file running the runAnalysis function
	for _, file := range astFiles {
		wg.Add(1)
		nbParsingFiles++
		go func(file string) {
			defer wg.Done()
			executeFileAnalysis(file, channelResult)
			// details is the number of files processed / total number of files
			details := strconv.Itoa(nbParsingFiles) + "/" + strconv.Itoa(len(astFiles))
			progressbar.UpdateText("Analyzing (" + details + ")")
		}(file)
	}

	wg.Wait()
	progressbar.Info("AST Analysis finished")
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
	if pbFile.Errors == nil {
		pbFile.Errors = make([]string, 0)
	}

	// load AST via ProtoBuf (using NodeType package)
	in, err := ioutil.ReadFile(file)
	if err != nil {
		pbFile.Errors = append(pbFile.Errors, "Error reading file: "+err.Error())
		channelResult <- pbFile
		log.Error("Error reading file: ", err)
		return err
	}

	// if file is empty, return
	if len(in) == 0 {
		pbFile.Errors = append(pbFile.Errors, "File is empty: "+file)
		channelResult <- pbFile
		log.Error("File is empty: ", file)
		return err
	}

	if err := proto.Unmarshal(in, pbFile); err != nil {
		pbFile.Errors = append(pbFile.Errors, "Failed to parse address pbFile ("+file+"): "+err.Error())
		channelResult <- pbFile
		log.Error("Failed to parse address pbFile ("+file+"):", err)
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
	channelResult <- pbFile
	return nil
}
