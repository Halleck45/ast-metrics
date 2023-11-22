package Php

import (
	"os"
	"strings"
	"sync"

	"fmt"

	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Storage"

	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/traverser"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/File"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
)

type PhpRunner struct {
	progressbar   *pterm.SpinnerPrinter
	configuration *Configuration.Configuration
	foundFiles    File.FileList
}

// IsRequired returns true if at least one Go file is found
func (r PhpRunner) IsRequired() bool {
	// If at least one Go file is found, we need to run PHP engine
	return len(r.getFileList().Files) > 0
}

// SetProgressbar sets the progressbar
func (r *PhpRunner) SetProgressbar(progressbar *pterm.SpinnerPrinter) {
	(*r).progressbar = progressbar
}

// SetConfiguration sets the configuration
func (r *PhpRunner) SetConfiguration(configuration *Configuration.Configuration) {
	(*r).configuration = configuration
}

// Ensure ensures Go is ready to run.
func (r *PhpRunner) Ensure() error {
	return nil
}

// Finish cleans up the workspace
func (r PhpRunner) Finish() error {
	r.progressbar.Stop()
	return nil
}

// DumpAST dumps the AST of python files in protobuf format
func (r PhpRunner) DumpAST() {

	var wg sync.WaitGroup
	cnt := 0
	for _, filePath := range r.getFileList().Files {
		cnt++
		r.progressbar.UpdateText("Dumping AST of PHP files (" + fmt.Sprintf("%d", cnt) + "/" + fmt.Sprintf("%d", len(r.getFileList().Files)) + ")")
		wg.Add(1)
		go r.dumpOneAst(&wg, filePath)
	}

	wg.Wait()

	r.progressbar.Info("PHP code dumped")
}

func (r PhpRunner) dumpOneAst(wg *sync.WaitGroup, filePath string) {
	defer wg.Done()
	hash, err := Engine.GetFileHash(filePath)
	if err != nil {
		log.Error(err)
	}
	binPath := Storage.OutputPath() + string(os.PathSeparator) + hash + ".bin"
	// if file exists, skip it
	if _, err := os.Stat(binPath); err == nil {
		return
	}

	// Create protobuf object
	protoFile, _ := parsePhpFile(filePath)

	// Dump protobuf object to destination
	Engine.DumpProtobuf(protoFile, binPath)
}

func parsePhpFile(filename string) (*pb.File, error) {

	stmts := Engine.FactoryStmts()

	file := &pb.File{
		Path:                filename,
		ProgrammingLanguage: "PHP",
		Stmts:               stmts,
	}

	sourceCode, err := os.ReadFile(filename)
	if err != nil {
		return file, err
	}
	linesOfFile := strings.Split(string(sourceCode), "\n")

	// Error handler
	var parserErrors []*errors.Error
	errorHandler := func(e *errors.Error) {
		e.Msg += " for file " + filename
		parserErrors = append(parserErrors, e)
	}

	// Parse
	rootNode, err := parser.Parse(sourceCode, conf.Config{
		Version:          &version.Version{Major: 8, Minor: 0},
		ErrorHandlerFunc: errorHandler,
	})

	if err != nil {
		log.Error("Error:" + err.Error())
	}

	if len(parserErrors) > 0 {
		for _, e := range parserErrors {
			log.Println(e.String())
		}
	}

	// visit the AST
	visitor := PhpVisitor{file: file, linesOfFile: linesOfFile}
	traverser := traverser.NewTraverser(&visitor)
	traverser.Traverse(rootNode)

	return file, nil
}

// getFileList returns the list of PHP files to analyze, and caches it in memory
func (r *PhpRunner) getFileList() File.FileList {

	if r.foundFiles.Files != nil {
		return r.foundFiles
	}

	finder := File.Finder{Configuration: *r.configuration}
	r.foundFiles = finder.Search(".php")

	return r.foundFiles
}
