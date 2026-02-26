package treesitter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/pb"

	sitter "github.com/smacker/go-tree-sitter"
)

type Runner struct {
	Adapter       LangAdapter
	Configuration *configuration.Configuration
	UpdateText    func(msg string) // optionnel
}

func (r Runner) ParseFile(path string) (*pb.File, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parser := sitter.NewParser()
	parser.SetLanguage(r.Adapter.Language())

	tree := parser.Parse(nil, src)
	root := tree.RootNode()

	v := NewVisitor(r.Adapter, path, src)
	v.Visit(root)
	return v.Result(), nil
}

func (r Runner) WalkAndProcess(files []string) []*pb.File {
	total := len(files)
	results := make([]*pb.File, 0, total)
	for i, f := range files {
		if r.UpdateText != nil {
			r.UpdateText(fmt.Sprintf("Tree-sitter: %s [%d/%d]", filepath.Base(f), i+1, total))
		}
		if file, err := r.ParseFile(f); err == nil && file != nil {
			results = append(results, file)
		}
	}
	return results
}
