package Python

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/halleck45/ast-metrics/src/Engine/Treesitter"
	"github.com/halleck45/ast-metrics/src/File"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Storage"

	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"

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
	files := r.getFileList().Files
	total := len(files)
	if total == 0 {
		if r.progressbar != nil {
			r.progressbar.Info("No Python files detected")
		}
		return
	}

	for i, filePath := range files {
		if r.progressbar != nil {
			base := filepath.Base(filePath)
			r.progressbar.UpdateText(fmt.Sprintf("Dumping AST of Python files (%s) [%d/%d]", base, i+1, total))
		}

		hash, err := Storage.GetFileHash(filePath)
		if err != nil {
			log.WithError(err).Warn("failed to get file hash")
			continue
		}
		binPath := r.configuration.Storage.AstDirectory() + string(os.PathSeparator) + hash + ".bin"
		// if file exists, skip it
		if _, err := os.Stat(binPath); err == nil {
			continue
		}

		protoFile, err := r.Parse(filePath)
		if err != nil {
			log.WithError(err).Warn("python parse failed")
			continue
		}

		// Dump protobuf object to destination
		err = Engine.DumpProtobuf(protoFile, binPath)
		if err != nil {
			log.Error(err)
		}
	}

	if r.progressbar != nil {
		r.progressbar.Info("Python code dumped (tree-sitter)")
	}
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
