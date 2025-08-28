package Treesitter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Storage"
	"google.golang.org/protobuf/proto"

	sitter "github.com/smacker/go-tree-sitter"
)

type Runner struct {
	Adapter       LangAdapter
	Configuration *Configuration.Configuration
	UpdateText    func(msg string) // optionnel
}

func (r Runner) ParseAndStore(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	parser := sitter.NewParser()
	parser.SetLanguage(r.Adapter.Language())

	tree := parser.Parse(nil, src)
	root := tree.RootNode()

	v := NewVisitor(r.Adapter, path, src)
	v.Visit(root)
	file := v.Result()

	hash, err := Storage.GetFileHash(path)
	if err != nil {
		return err
	}
	bin := r.Configuration.Storage.AstDirectory() + string(os.PathSeparator) + hash + ".bin"
	data, err := proto.Marshal(file)
	if err != nil {
		return err
	}
	return os.WriteFile(bin, data, 0o644)
}

func (r Runner) WalkAndProcess(files []string) {
	total := len(files)
	for i, f := range files {
		if r.UpdateText != nil {
			r.UpdateText(fmt.Sprintf("Tree-sitter: %s [%d/%d]", filepath.Base(f), i+1, total))
		}
		_ = r.ParseAndStore(f) // log à l’appelant si besoin
	}
}
