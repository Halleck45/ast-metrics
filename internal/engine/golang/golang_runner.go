package golang

import (
	"fmt"
	"os"
	"strings"

	"github.com/halleck45/ast-metrics/internal/configuration"
	engine "github.com/halleck45/ast-metrics/internal/engine"
	Treesitter "github.com/halleck45/ast-metrics/internal/engine/treesitter"
	File "github.com/halleck45/ast-metrics/internal/file"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/pterm/pterm"
	"golang.org/x/mod/modfile"

	sitter "github.com/smacker/go-tree-sitter"
)

type GolangRunner struct {
	progressbar      *pterm.SpinnerPrinter
	Configuration    *configuration.Configuration
	foundFiles       File.FileList
	currentGoModFile *modfile.File
	currentGoModPath string
}

// IsRequired returns true if at least one Go file is found
func (r GolangRunner) IsRequired() bool {
	// If at least one Go file is found, we need to run PHP engine
	return len(r.getFileList().Files) > 0
}

// SetProgressbar sets the progressbar
func (r *GolangRunner) SetProgressbar(progressbar *pterm.SpinnerPrinter) {
	(*r).progressbar = progressbar
}

// SetConfiguration sets the configuration
func (r *GolangRunner) SetConfiguration(configuration *configuration.Configuration) {
	(*r).Configuration = configuration
}

// Ensure ensures Go is ready to run.
func (r *GolangRunner) Ensure() error {
	return nil
}

// Finish cleans up the workspace
func (r GolangRunner) Finish() error {
	if r.progressbar != nil {
		r.progressbar.Stop()
	}
	return nil
}

// DumpAST dumps the AST of Go files using tree-sitter (aligned with PHP/Python)
func (r GolangRunner) DumpAST() {
	engine.DumpFiles(
		r.getFileList().Files, r.Configuration, r.progressbar,
		func(path string) (*pb.File, error) { return r.Parse(path) },
		engine.DumpOptions{Label: r.Name()},
	)
}

func (r GolangRunner) Name() string {
	return "Golang"
}

func (r *GolangRunner) SearchModfile(path string) (*modfile.File, error) {

	// Avoid duplicate search
	if r.currentGoModFile != nil {
		// if directory is a subdirectory of the current mod file, return it
		if strings.Contains(path, r.currentGoModPath) {
			return r.currentGoModFile, nil
		}
	}

	goModFile := path + string(os.PathSeparator) + "go.mod"

	if _, err := os.Stat(goModFile); err == nil {

		fileBytes, err := os.ReadFile(goModFile)
		if err != nil {
			return nil, err
		}
		f, err := modfile.Parse("go.mod", fileBytes, nil)
		if err != nil {
			return nil, err
		}

		r.currentGoModFile = f
		r.currentGoModPath = path

		return f, nil
	}

	// Search in parent directory
	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) <= 2 {
		return nil, fmt.Errorf("go.mod file not found")
	}
	parts = parts[:len(parts)-1]
	parentDirectory := strings.Join(parts, string(os.PathSeparator))
	return r.SearchModfile(parentDirectory)
}

func (r *GolangRunner) Parse(path string) (*pb.File, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return &pb.File{Path: path, ProgrammingLanguage: "Golang"}, err
	}

	parser := sitter.NewParser()
	adapter := NewTreeSitterAdapter(src)
	parser.SetLanguage(adapter.Language())

	tree := parser.Parse(nil, src)
	root := tree.RootNode()

	v := Treesitter.NewVisitor(adapter, path, src)
	v.Visit(root)
	file := v.Result()
	file.ProgrammingLanguage = "Golang"
	if root.HasError() {
		file.Errors = append(file.Errors, "Parse error")
	}
	return file, nil
}

// getFileList returns the list of PHP files to analyze, and caches it in memory
func (r *GolangRunner) getFileList() File.FileList {

	if r.foundFiles.Files != nil {
		return r.foundFiles
	}

	finder := File.Finder{Configuration: *r.Configuration}
	r.foundFiles = finder.Search(".go")

	return r.foundFiles
}
