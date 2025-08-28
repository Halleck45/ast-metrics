package Treesitter

import (
	sitter "github.com/smacker/go-tree-sitter"
)

type DecisionKind int

const (
	DecNone DecisionKind = iota
	DecIf
	DecElif
	DecElse
	DecLoop
	DecSwitch
	DecCase
)

type ImportItem struct {
	Module string // e.g., "pkg.sub" (for `from pkg.sub import X`) or full module for `import pkg.sub`
	Name   string // imported symbol (empty for plain `import pkg.sub`)
}

type LangAdapter interface {
	Language() *sitter.Language

	// structure
	IsModule(*sitter.Node) bool
	IsClass(*sitter.Node) bool
	IsFunction(*sitter.Node) bool

	// attributes
	NodeName(*sitter.Node) string
	NodeBody(*sitter.Node) *sitter.Node
	NodeParams(*sitter.Node) *sitter.Node
	ModuleNameFromPath(path string) string
	AttachQualified(parentClass, fn string) string
	EachChildBody(n *sitter.Node, yield func(*sitter.Node))
	EachParamIdent(params *sitter.Node, yield func(string))

	// decisions
	Decision(n *sitter.Node) (DecisionKind, *sitter.Node)

	// imports
	Imports(n *sitter.Node) []ImportItem
}
