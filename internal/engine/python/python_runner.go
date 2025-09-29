package python

import (
	"os"

	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	Treesitter "github.com/halleck45/ast-metrics/internal/engine/treesitter"
	"github.com/halleck45/ast-metrics/internal/file"
	pb "github.com/halleck45/ast-metrics/pb"

	"github.com/pterm/pterm"

	sitter "github.com/smacker/go-tree-sitter"
)

type PythonRunner struct {
	progressbar   *pterm.SpinnerPrinter
	Configuration *configuration.Configuration
	foundFiles    file.FileList
}

// IsRequired returns true when analyzed files are concerned by the programming language
func (r PythonRunner) IsRequired() bool {
	return len(r.getFileList().Files) > 0
}

// Prepare the engine
func (r *PythonRunner) Ensure() error { return nil }

// First step of analysis. Parse all files, and generate protobuf-compatible AST files
func (r PythonRunner) DumpAST() {
	engine.DumpFiles(
		r.getFileList().Files, r.Configuration, r.progressbar,
		func(path string) (*pb.File, error) { return r.Parse(path) },
		engine.DumpOptions{Label: r.Name()},
	)
}

func (r PythonRunner) Name() string {
	return "Python"
}

// Cleanups the engine
func (r PythonRunner) Finish() error {
	if r.progressbar != nil {
		r.progressbar.Stop()
	}
	return nil
}

// Give a UI progress bar to the engine
func (r *PythonRunner) SetProgressbar(progressbar *pterm.SpinnerPrinter) {
	r.progressbar = progressbar
}

// Give the configuration to the engine
func (r *PythonRunner) SetConfiguration(configuration *configuration.Configuration) {
	r.Configuration = configuration
}

// Parse a file and return a protobuf-compatible AST object (no store)
func (r PythonRunner) Parse(path string) (*pb.File, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return &pb.File{Path: path, ProgrammingLanguage: "Python"}, err
	}

	parser := sitter.NewParser()
	adapter := NewTreeSitterAdapter(src)
	parser.SetLanguage(adapter.Language())

	tree := parser.Parse(nil, src)
	root := tree.RootNode()

	v := Treesitter.NewVisitor(adapter, path, src)
	v.Visit(root)
	file := v.Result()
	file.ProgrammingLanguage = "Python"

	return file, nil
}

func (r *PythonRunner) getFileList() file.FileList {
	if r.foundFiles.Files != nil {
		return r.foundFiles
	}

	finder := file.Finder{Configuration: *r.Configuration}
	r.foundFiles = finder.Search(".py")
	return r.foundFiles
}
