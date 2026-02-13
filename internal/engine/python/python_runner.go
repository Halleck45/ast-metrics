package python

import (
	"os"
	"path/filepath"
	"strings"

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

	// Detect if file is a test file
	file.IsTest = r.isTestFile(path, file)

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

// isTestFile determines if a Python file is a test file based on:
// 1. Filename pattern (starts with test_ or ends with _test.py)
// 2. Class inheritance (extends unittest.TestCase or similar test base classes)
func (r PythonRunner) isTestFile(path string, file *pb.File) bool {
	baseName := strings.ToLower(path)
	fileName := strings.ToLower(filepath.Base(path))

	// Check filename pattern
	if strings.HasPrefix(fileName, "test_") || strings.HasSuffix(baseName, "_test.py") {
		return true
	}

	// Check if any class extends a test base class
	classes := engine.GetClassesInFile(file)
	for _, class := range classes {
		if class == nil {
			continue
		}
		// Check extends (Python uses extends for inheritance)
		for _, ext := range class.Extends {
			if ext == nil {
				continue
			}
			qualified := strings.ToLower(ext.Qualified)
			short := strings.ToLower(ext.Short)
			// Common Python test base classes
			if strings.Contains(qualified, "testcase") ||
				strings.Contains(short, "testcase") ||
				strings.Contains(qualified, "unittest") ||
				strings.Contains(qualified, "pytest") {
				return true
			}
		}
	}

	return false
}
