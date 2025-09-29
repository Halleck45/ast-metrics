package rust

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

type RustRunner struct {
	progressbar   *pterm.SpinnerPrinter
	Configuration *configuration.Configuration
	foundFiles    file.FileList
}

func (r RustRunner) Name() string                                     { return "Rust" }
func (r RustRunner) IsRequired() bool                                 { return len(r.getFileList().Files) > 0 }
func (r *RustRunner) Ensure() error                                   { return nil }
func (r *RustRunner) SetProgressbar(p *pterm.SpinnerPrinter)          { r.progressbar = p }
func (r *RustRunner) SetConfiguration(c *configuration.Configuration) { r.Configuration = c }

func (r RustRunner) DumpAST() {
	engine.DumpFiles(
		r.getFileList().Files,
		r.Configuration,
		r.progressbar,
		func(path string) (*pb.File, error) { return r.Parse(path) },
		engine.DumpOptions{Label: r.Name()},
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

func (r *RustRunner) getFileList() file.FileList {
	if r.foundFiles.Files != nil {
		return r.foundFiles
	}
	finder := file.Finder{Configuration: *r.Configuration}
	r.foundFiles = finder.Search(".rs")
	return r.foundFiles
}
