package Python

import (
	"os"

	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Engine"
	Treesitter "github.com/halleck45/ast-metrics/src/Engine/TreeSitter"
	"github.com/halleck45/ast-metrics/src/File"
	pb "github.com/halleck45/ast-metrics/src/NodeType"

	"github.com/pterm/pterm"

	sitter "github.com/smacker/go-tree-sitter"
)

type PythonRunner struct {
	progressbar   *pterm.SpinnerPrinter
	configuration *Configuration.Configuration
	foundFiles    File.FileList
}

// IsRequired returns true when analyzed files are concerned by the programming language
func (r PythonRunner) IsRequired() bool {
	return len(r.getFileList().Files) > 0
}

// Prepare the engine
func (r *PythonRunner) Ensure() error { return nil }

// First step of analysis. Parse all files, and generate protobuf-compatible AST files
func (r PythonRunner) DumpAST() {
	Engine.DumpFiles(
		r.getFileList().Files, r.configuration, r.progressbar,
		func(path string) (*pb.File, error) { return r.Parse(path) },
		Engine.DumpOptions{Label: r.Name()},
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
func (r *PythonRunner) SetConfiguration(configuration *Configuration.Configuration) {
	r.configuration = configuration
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

func (r *PythonRunner) getFileList() File.FileList {
	if r.foundFiles.Files != nil {
		return r.foundFiles
	}

	finder := File.Finder{Configuration: *r.configuration}
	r.foundFiles = finder.Search(".py")
	return r.foundFiles
}
