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
	"github.com/pterm/pterm"
	"github.com/yargevad/filepathx"
	"google.golang.org/protobuf/proto"
)

func Start(workdir string, progressbar *pterm.SpinnerPrinter) []*pb.File {

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

			if progressbar != nil {
				progressbar.UpdateText("Analyzing (" + details + ")")
			}

		}(file)
	}

	wg.Wait()
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
		log.Error("Error reading file: ", err)
		if pbFile.Errors == nil {
			pbFile.Errors = make([]string, 0)
		}
		pbFile.Errors = append(pbFile.Errors, "Error reading file: "+err.Error())
		channelResult <- pbFile
		return err
	}

	// if file is empty, return
	if len(in) == 0 {
		log.Error("File is empty: ", file)
		if pbFile.Errors == nil {
			pbFile.Errors = make([]string, 0)
		}
		pbFile.Errors = append(pbFile.Errors, "File is empty: "+file)
		channelResult <- pbFile
		return err
	}

	if err := proto.Unmarshal(in, pbFile); err != nil {
		log.Errorln("Failed to parse address pbFile ("+file+"):", err)
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
	channelResult <- pbFile
	return nil
}
