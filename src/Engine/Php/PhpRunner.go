package Php

import (
	"os"
	"strings"

	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"

	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/traverser"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/File"
	"github.com/pterm/pterm"
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
	Engine.DumpFiles(
		r.getFileList().Files, r.configuration, r.progressbar,
		func(path string) (*pb.File, error) { return parsePhpFile(path) },
		Engine.DumpOptions{Label: r.Name()},
	)
}

func (r PhpRunner) Name() string {
	return "PHP"
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
