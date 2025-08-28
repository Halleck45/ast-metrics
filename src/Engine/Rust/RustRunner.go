package Rust

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

type RustRunner struct {
	progressbar   *pterm.SpinnerPrinter
	configuration *Configuration.Configuration
	foundFiles    File.FileList
}

func (r RustRunner) Name() string                                     { return "Rust" }
func (r RustRunner) IsRequired() bool                                 { return len(r.getFileList().Files) > 0 }
func (r *RustRunner) Ensure() error                                   { return nil }
func (r *RustRunner) SetProgressbar(p *pterm.SpinnerPrinter)          { r.progressbar = p }
func (r *RustRunner) SetConfiguration(c *Configuration.Configuration) { r.configuration = c }

func (r RustRunner) DumpAST() {
	Engine.DumpFiles(
		r.getFileList().Files,
		r.configuration,
		r.progressbar,
		func(path string) (*pb.File, error) { return r.Parse(path) },
		Engine.DumpOptions{Label: r.Name()},
	)
}

func (r RustRunner) Finish() error {
	if r.progressbar != nil {
		r.progressbar.Stop()
	}
	return nil
}

func (r RustRunner) Parse(path string) (*pb.File, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return &pb.File{Path: path, ProgrammingLanguage: "Rust"}, err
	}

	parser := sitter.NewParser()
	adapter := NewTreeSitterAdapter(src)
	parser.SetLanguage(adapter.Language())

	tree := parser.Parse(nil, src)
	root := tree.RootNode()

	v := Treesitter.NewVisitor(adapter, path, src)
	v.Visit(root)

	file := v.Result()
	file.ProgrammingLanguage = "Rust"
	return file, nil
}

func (r *RustRunner) getFileList() File.FileList {
	if r.foundFiles.Files != nil {
		return r.foundFiles
	}
	finder := File.Finder{Configuration: *r.configuration}
	r.foundFiles = finder.Search(".rs")
	return r.foundFiles
}
