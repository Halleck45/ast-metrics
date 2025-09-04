package rust

import (
	"path/filepath"
	"strings"

 Treesitter "github.com/halleck45/ast-metrics/internal/engine/treesitter"
	sitter "github.com/smacker/go-tree-sitter"
	tsRust "github.com/smacker/go-tree-sitter/rust"
)

type TreeSitterAdapter struct {
	src []byte
}

func NewTreeSitterAdapter(src []byte) *TreeSitterAdapter { return &TreeSitterAdapter{src: src} }
func (a *TreeSitterAdapter) SetSource(src []byte)        { a.src = src }

func (a *TreeSitterAdapter) Language() *sitter.Language { return tsRust.GetLanguage() }

func (a *TreeSitterAdapter) NodeName(n *sitter.Node) string {
	if a.src == nil || n == nil {
		return ""
	}
	// functions and methods: field "name"
	if name := n.ChildByFieldName("name"); name != nil {
		return a.text(name)
	}
	// impl blocks and type items: take identifier
	if id := firstChildOfType(n, "type_identifier"); id != nil {
		return a.text(id)
	}
	if id := firstChildOfType(n, "identifier"); id != nil {
		return a.text(id)
	}
	// qualified path fallback: scoped_identifier segments
	if q := firstChildOfType(n, "scoped_identifier"); q != nil {
		txt := a.text(q)
		if i := strings.LastIndex(txt, "::"); i >= 0 && i+2 < len(txt) {
			return txt[i+2:]
		}
		return txt
	}
	return ""
}

func (a *TreeSitterAdapter) NodeBody(n *sitter.Node) *sitter.Node {
	if n == nil {
		return nil
	}
	// function_item / method_item → body (block)
	if b := n.ChildByFieldName("body"); b != nil {
		return b
	}
	// impl_item block body
	if b := firstChildOfType(n, "declaration_list"); b != nil {
		return b
	}
	// trait_item body
	if b := firstChildOfType(n, "trait_item"); b != nil {
		return b
	}
	// generic block fallback
	if b := firstChildOfType(n, "block"); b != nil {
		return b
	}
	return nil
}

func (a *TreeSitterAdapter) NodeParams(n *sitter.Node) *sitter.Node {
	if n == nil {
		return nil
	}
	// Rust grammar uses "parameters" under function_item and method_item
	if p := n.ChildByFieldName("parameters"); p != nil {
		return p
	}
	return firstChildOfType(n, "parameters")
}

func (a *TreeSitterAdapter) EachParamIdent(params *sitter.Node, yield func(string)) {
	if params == nil || a.src == nil {
		return
	}
	var walk func(*sitter.Node)
	walk = func(x *sitter.Node) {
		if x == nil {
			return
		}
		typ := x.Type()
		// identifiers inside parameter patterns
		if typ == "identifier" || typ == "type_identifier" || typ == "shorthand_field_identifier" {
			yield(a.text(x))
		}
		// self parameter
		if typ == "self" || typ == "self_parameter" {
			yield("self")
			return
		}
		// pattern_identifier covers simple `x: T`
		if typ == "pattern_identifier" {
			yield(a.text(x))
		}
		for i := 0; i < int(x.ChildCount()); i++ {
			walk(x.Child(i))
		}
	}
	walk(params)
}

func (a *TreeSitterAdapter) IsModule(n *sitter.Node) bool {
	return n.Type() == "source_file"
}

func (a *TreeSitterAdapter) IsClass(n *sitter.Node) bool {
	// Rust has no classes; treat struct, enum, trait, impl as class-like containers
	switch n.Type() {
	case "struct_item", "enum_item", "union_item", "trait_item", "impl_item":
		return true
	}
	return false
}

func (a *TreeSitterAdapter) IsFunction(n *sitter.Node) bool {
	switch n.Type() {
	case "function_item", "function_signature_item", "method_item":
		return true
	}
	return false
}

func (a *TreeSitterAdapter) ModuleNameFromPath(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func (a *TreeSitterAdapter) AttachQualified(parent string, fn string) string {
	if parent == "" {
		return fn
	}
	return parent + "::" + fn
}

func (a *TreeSitterAdapter) EachChildBody(body *sitter.Node, yield func(*sitter.Node)) {
	if body == nil {
		return
	}
	switch body.Type() {
	case "block", "declaration_list":
		for i := 0; i < int(body.ChildCount()); i++ {
			yield(body.Child(i))
		}
	case "match_expression":
		// enumerate match arms
		for i := 0; i < int(body.ChildCount()); i++ {
			n := body.Child(i)
			if n.Type() == "match_block" || n.Type() == "match_body" {
				for j := 0; j < int(n.ChildCount()); j++ {
					arm := n.Child(j)
					if arm.Type() == "match_arm" {
						yield(arm)
					}
				}
			}
		}
	default:
		for i := 0; i < int(body.ChildCount()); i++ {
			yield(body.Child(i))
		}
	}
}

func (a *TreeSitterAdapter) Decision(n *sitter.Node) (Treesitter.DecisionKind, *sitter.Node) {
	switch n.Type() {
	case "if_expression":
		if b := firstChildOfType(n, "block"); b != nil {
			return Treesitter.DecIf, b
		}
		return Treesitter.DecIf, nil
	case "else_clause":
		if b := firstChildOfType(n, "block"); b != nil {
			return Treesitter.DecElse, b
		}
		return Treesitter.DecElse, nil
	case "if_let_expression": // treat as if
		if b := firstChildOfType(n, "block"); b != nil {
			return Treesitter.DecIf, b
		}
		return Treesitter.DecIf, nil
	case "match_expression":
		// let EachChildBody enumerate arms
		return Treesitter.DecSwitch, n
	case "match_arm":
		// body may be an expression or block; prefer block
		if b := firstChildOfType(n, "block"); b != nil {
			return Treesitter.DecCase, b
		}
		// fallback: last child
		if n.ChildCount() > 0 {
			return Treesitter.DecCase, n.Child(int(n.ChildCount() - 1))
		}
		return Treesitter.DecCase, nil
	case "for_expression", "while_expression", "loop_expression":
		if b := firstChildOfType(n, "block"); b != nil {
			return Treesitter.DecLoop, b
		}
		return Treesitter.DecLoop, nil
	}
	return Treesitter.DecNone, nil
}

func (a *TreeSitterAdapter) Imports(n *sitter.Node) []Treesitter.ImportItem {
	if n == nil {
		return nil
	}
	if n.Type() != "use_declaration" {
		return nil
	}
	return a.parseUse(n)
}

func (a *TreeSitterAdapter) parseUse(n *sitter.Node) []Treesitter.ImportItem {
	items := []Treesitter.ImportItem{}
	add := func(full string) {
		full = strings.TrimSpace(full)
		if full == "" {
			return
		}
		// cut alias if present: "foo::bar as Baz"
		full = strings.Split(full, " as ")[0]
		mod, name := splitModuleLeaf(full, "::")
		items = append(items, Treesitter.ImportItem{Module: mod, Name: name})
	}
	var walk func(*sitter.Node, string)
	walk = func(x *sitter.Node, prefix string) {
		if x == nil {
			return
		}
		switch x.Type() {
		case "use_tree":
			// may contain scoped_identifier, use_list, use_as_clause
		case "scoped_identifier", "identifier", "crate", "super", "self":
			path := a.text(x)
			if prefix != "" {
				path = strings.TrimSuffix(prefix, "::") + "::" + strings.TrimPrefix(path, "::")
			}
			add(path)
			return
		case "use_list":
			// grouped: foo::{bar, baz as Qux}
			// prefix is set by preceding scoped_identifier
		case "use_as_clause":
			// "path as Alias" → child 0 is path, we ignore alias for identity
			if x.ChildCount() > 0 {
				path := a.text(x.Child(0))
				if prefix != "" {
					path = strings.TrimSuffix(prefix, "::") + "::" + strings.TrimPrefix(path, "::")
				}
				add(path)
				return
			}
		}
		// derive group prefix if parent has a scoped_identifier
		if x.Type() == "use_list" || x.Type() == "use_tree" {
			p := prefix
			if q := firstChildOfType(x, "scoped_identifier"); q != nil {
				p = a.text(q)
			}
			for i := 0; i < int(x.ChildCount()); i++ {
				walk(x.Child(i), p)
			}
			return
		}
		for i := 0; i < int(x.ChildCount()); i++ {
			walk(x.Child(i), prefix)
		}
	}
	walk(n, "")
	return dedup(items)
}

// helpers

func (a *TreeSitterAdapter) text(n *sitter.Node) string {
	if n == nil || a.src == nil {
		return ""
	}
	return string(a.src[n.StartByte():n.EndByte()])
}

func firstChildOfType(n *sitter.Node, t string) *sitter.Node {
	if n == nil {
		return nil
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		ch := n.Child(i)
		if ch.Type() == t {
			return ch
		}
	}
	return nil
}

func splitModuleLeaf(full string, sep string) (string, string) {
	full = strings.Trim(full, sep)
	if i := strings.LastIndex(full, sep); i >= 0 {
		return full[:i], full[i+len(sep):]
	}
	return full, ""
}

func dedup(in []Treesitter.ImportItem) []Treesitter.ImportItem {
	if len(in) <= 1 {
		return in
	}
	seen := map[string]struct{}{}
	out := in[:0]
	for _, it := range in {
		key := it.Module + " " + it.Name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, it)
	}
	return out
}
