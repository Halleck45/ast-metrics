package Php

import (
	"os"
	"runtime"
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
	if r.progressbar != nil {
		r.progressbar.Stop()
	}
	return nil
}

// DumpAST dumps the AST of python files in protobuf format
func (r PhpRunner) DumpAST() {

	cpuCount := runtime.NumCPU()
	var wg sync.WaitGroup
	cnt := 0
	filesChan := make(chan string, cpuCount)

	for i := 0; i < cpuCount; i++ {
		go func() {
			for filePath := range filesChan {
				cnt++
				if r.progressbar != nil {
					r.progressbar.UpdateText("Dumping AST of PHP files (" + fmt.Sprintf("%d", cnt) + "/" + fmt.Sprintf("%d", len(r.getFileList().Files)) + ")")
				}
				wg.Add(1)
				go r.dumpOneAst(&wg, filePath)
			}
		}()
	}

	// split files between workers
	for _, filePath := range r.getFileList().Files {
		filesChan <- filePath
	}

	// wait for all workers to finish
	close(filesChan)
	wg.Wait()

	if r.progressbar != nil {
		r.progressbar.Info("PHP code dumped")
	}
}

func (r PhpRunner) dumpOneAst(wg *sync.WaitGroup, filePath string) {
	defer wg.Done()
	hash, err := Storage.GetFileHash(filePath)
	if err != nil {
		log.Error("Error while hashing file " + filePath + ": " + err.Error())
	}
	binPath := r.configuration.Storage.AstDirectory() + string(os.PathSeparator) + hash + ".bin"
	// if file exists, skip it
	if _, err := os.Stat(binPath); err == nil {
		return
	}

	// Create protobuf object
	protoFile, _ := parsePhpFile(filePath)
	protoFile.Checksum = hash

	// Dump protobuf object to destination
	err = Engine.DumpProtobuf(protoFile, binPath)
	if err != nil {
		log.Error("Error while dumping file " + filePath + ": " + err.Error())
	}
}

func (r PhpRunner) Parse(filePath string) (*pb.File, error) {
	return parsePhpFile(filePath)
}

// @deprecated. Please use the Parse function
func parsePhpFile(filename string) (*pb.File, error) {

	stmts := Engine.FactoryStmts()

	file := &pb.File{
		Path:                filename,
		ProgrammingLanguage: "PHP",
		Stmts:               stmts,
		LinesOfCode:         &pb.LinesOfCode{},
	}

	sourceCode, err := os.ReadFile(filename)
	if err != nil {
		return file, err
	}
	linesOfFile := strings.Split(string(sourceCode), "\n")
	file.LinesOfCode.LinesOfCode = int32(len(linesOfFile))

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
		parserErrors = append(parserErrors, errors.NewError(err.Error(), nil))
	}

	if len(parserErrors) > 0 {
		for _, e := range parserErrors {
			file.Errors = append(file.Errors, e.Msg)
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
