package python

import (
	"path/filepath"
	"strings"

 Treesitter "github.com/halleck45/ast-metrics/internal/engine/treesitter"
	sitter "github.com/smacker/go-tree-sitter"
	tsPython "github.com/smacker/go-tree-sitter/python"
)

type TreeSitterAdapter struct {
	src []byte
}

func NewTreeSitterAdapter(src []byte) *TreeSitterAdapter { return &TreeSitterAdapter{src: src} }

func (a *TreeSitterAdapter) SetSource(src []byte) { a.src = src }

func (a *TreeSitterAdapter) NodeName(n *sitter.Node) string {
	if a.src == nil || n == nil {
		return ""
	}
	// champ nommé "name" si dispo
	if name := n.ChildByFieldName("name"); name != nil {
		return string(a.src[name.StartByte():name.EndByte()])
	}
	// fallback: premier identifier
	if id := firstChildOfType(n, "identifier"); id != nil {
		return string(a.src[id.StartByte():id.EndByte()])
	}
	return ""
}

func (a *TreeSitterAdapter) NodeBody(n *sitter.Node) *sitter.Node {
	if n == nil {
		return nil
	}
	if body := n.ChildByFieldName("body"); body != nil {
		return body
	}
	// fallback: last child of type "block"
	if b := firstChildOfType(n, "block"); b != nil {
		return b
	}
	return nil
}

func (a *TreeSitterAdapter) NodeParams(n *sitter.Node) *sitter.Node {
	if n == nil {
		return nil
	}
	if p := n.ChildByFieldName("parameters"); p != nil {
		return p
	}
	// some grammars name this "parameter_list"
	if p := firstChildOfType(n, "parameter_list"); p != nil {
		return p
	}
	return nil
}

func (a *TreeSitterAdapter) EachParamIdent(params *sitter.Node, yield func(string)) {
	if params == nil || a.src == nil {
		return
	}
	var walk func(*sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil {
			return
		}
		if n.Type() == "identifier" {
			yield(string(a.src[n.StartByte():n.EndByte()]))
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(params)
}

func (a *TreeSitterAdapter) Language() *sitter.Language { return tsPython.GetLanguage() }

func (a *TreeSitterAdapter) IsModule(n *sitter.Node) bool { return n.Type() == "module" }
func (a *TreeSitterAdapter) IsClass(n *sitter.Node) bool  { return n.Type() == "class_definition" }
func (a *TreeSitterAdapter) IsFunction(n *sitter.Node) bool {
	return n.Type() == "function_definition" || n.Type() == "async_function_definition"
}

func (a *TreeSitterAdapter) ModuleNameFromPath(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return base
}
func (a *TreeSitterAdapter) AttachQualified(parentClass string, fn string) string {
	if parentClass == "" {
		return fn
	}
	return parentClass + "." + fn
}
func (a *TreeSitterAdapter) EachChildBody(body *sitter.Node, yield func(*sitter.Node)) {
	if body == nil {
		return
	}

	switch body.Type() {
	case "block":
		for i := 0; i < int(body.ChildCount()); i++ {
			yield(body.Child(i))
		}

	case "match_statement":
		// Yield all case nodes, regardless of depth or wrapper.
		var walk func(*sitter.Node)
		walk = func(n *sitter.Node) {
			if n == nil {
				return
			}
			tt := n.Type()
			if tt == "case_clause" || tt == "case_block" {
				yield(n)
				// do not return; there can be nested blocks to visit later if needed
			}
			for i := 0; i < int(n.ChildCount()); i++ {
				walk(n.Child(i))
			}
		}
		walk(body)

	default:
		for i := 0; i < int(body.ChildCount()); i++ {
			yield(body.Child(i))
		}
	}
}
func (a *TreeSitterAdapter) text(n *sitter.Node) string {
	if n == nil || a.src == nil {
		return ""
	}
	return string(a.src[n.StartByte():n.EndByte()])
}
func findDescendantOfType(n *sitter.Node, t string) *sitter.Node {
	if n == nil {
		return nil
	}
	if n.Type() == t {
		return n
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		if got := findDescendantOfType(n.Child(i), t); got != nil {
			return got
		}
	}
	return nil
}

func (a *TreeSitterAdapter) Decision(n *sitter.Node) (Treesitter.DecisionKind, *sitter.Node) {
	switch n.Type() {
	case "if_statement":
		if b := n.ChildByFieldName("consequence"); b != nil {
			return Treesitter.DecIf, b
		}
		return Treesitter.DecIf, firstChildOfType(n, "block")

	case "elif_clause":
		if b := n.ChildByFieldName("consequence"); b != nil {
			return Treesitter.DecElif, b
		}
		return Treesitter.DecElif, firstChildOfType(n, "block")

	case "else_clause":
		if b := n.ChildByFieldName("body"); b != nil {
			return Treesitter.DecElse, b
		}
		return Treesitter.DecElse, firstChildOfType(n, "block")

	case "for_statement", "while_statement":
		if b := n.ChildByFieldName("body"); b != nil {
			return Treesitter.DecLoop, b
		}
		return Treesitter.DecLoop, firstChildOfType(n, "block")

	// Python 3.10+ structural pattern matching
	case "match_statement":
		// Return the node itself. EachChildBody will yield case nodes.
		return Treesitter.DecSwitch, n

	// Case nodes can be named "case_clause" (current grammar) or "case_block" (older).
	case "case_clause":
		if b := n.ChildByFieldName("body"); b != nil {
			return Treesitter.DecCase, b
		}
		// fallback: any descendant block
		if b := findDescendantOfType(n, "block"); b != nil {
			return Treesitter.DecCase, b
		}
		return Treesitter.DecCase, nil

	case "case_block":
		// older grammar fallback
		if b := findDescendantOfType(n, "block"); b != nil {
			return Treesitter.DecCase, b
		}
		return Treesitter.DecCase, nil
	}
	return Treesitter.DecNone, nil
}

func (a *TreeSitterAdapter) Imports(n *sitter.Node) []Treesitter.ImportItem {
	if n == nil {
		return nil
	}
	switch n.Type() {
	case "import_statement":
		return importsFromImportStatement(a, n)
	case "import_from_statement":
		return importsFromImportFromStatement(a, n)
	default:
		return nil
	}
}

func importsFromImportStatement(a *TreeSitterAdapter, n *sitter.Node) []Treesitter.ImportItem {
	items := []Treesitter.ImportItem{}
	// Robust: walk descendants and pick modules
	var walk func(*sitter.Node)
	walk = func(x *sitter.Node) {
		if x == nil {
			return
		}
		switch x.Type() {
		case "aliased_import":
			// Original symbol is in field "name"
			if nm := x.ChildByFieldName("name"); nm != nil {
				if txt := a.text(nm); txt != "" {
					items = append(items, Treesitter.ImportItem{Module: txt})
				}
				return // do not walk into alias
			}
			// Fallbacks
			if dn := firstChildOfType(x, "dotted_name"); dn != nil {
				items = append(items, Treesitter.ImportItem{Module: a.text(dn)})
				return
			}
			if id := firstChildOfType(x, "identifier"); id != nil {
				items = append(items, Treesitter.ImportItem{Module: a.text(id)})
				return
			}
		case "dotted_name":
			items = append(items, Treesitter.ImportItem{Module: a.text(x)})
			return
		case "identifier":
			txt := a.text(x)
			if txt != "" && txt != "import" && txt != "as" {
				items = append(items, Treesitter.ImportItem{Module: txt})
				return
			}
		}
		for i := 0; i < int(x.ChildCount()); i++ {
			walk(x.Child(i))
		}
	}
	walk(n)
	return dedup(items)
}

// helper: find first child of given type
func firstChildOfType(n *sitter.Node, t string) *sitter.Node {
	for i := 0; i < int(n.ChildCount()); i++ {
		ch := n.Child(i)
		if ch.Type() == t {
			return ch
		}
	}
	return nil
}

// helper: find the byte offset of the `import` keyword
func importKeywordStart(n *sitter.Node) (uint32, bool) {
	for i := 0; i < int(n.ChildCount()); i++ {
		ch := n.Child(i)
		if ch.Type() == "import" { // token node
			return ch.StartByte(), true
		}
	}
	return 0, false
}

func importsFromImportFromStatement(a *TreeSitterAdapter, n *sitter.Node) []Treesitter.ImportItem {
	items := []Treesitter.ImportItem{}
	if n == nil {
		return items
	}

	// 1) cut at `import`
	cut, ok := importKeywordStart(n)
	if !ok {
		// fallback to previous version if grammar differs
		host := findDescendantOfType(n, "import_list")
		if host == nil {
			host = n
		}
		// reuse your non-cut logic here if needed
	}

	// 2) resolve module: rightmost module-like node BEFORE cut
	var moduleNode *sitter.Node
	for i := 0; i < int(n.ChildCount()); i++ {
		ch := n.Child(i)
		if ch.EndByte() <= cut && (ch.Type() == "dotted_name" || ch.Type() == "relative_import") {
			moduleNode = ch // keep the last one before `import`
		}
	}
	module := strings.TrimLeft(strings.TrimSpace(a.text(moduleNode)), ".") // ".pkg" -> "pkg"

	// 3) collect names: nodes AFTER cut
	host := findDescendantOfType(n, "import_list")
	if host == nil {
		host = n
	}

	for i := 0; i < int(host.ChildCount()); i++ {
		ch := host.Child(i)
		if ch.StartByte() <= cut {
			continue
		}

		switch ch.Type() {
		case "aliased_import":
			// original symbol in field "name"
			if nm := ch.ChildByFieldName("name"); nm != nil {
				if name := a.text(nm); name != "" {
					items = append(items, Treesitter.ImportItem{Module: module, Name: name})
					continue
				}
			}
			// fallbacks
			if dn := firstChildOfType(ch, "dotted_name"); dn != nil {
				items = append(items, Treesitter.ImportItem{Module: module, Name: a.text(dn)})
				continue
			}
			if id := firstChildOfType(ch, "identifier"); id != nil {
				items = append(items, Treesitter.ImportItem{Module: module, Name: a.text(id)})
				continue
			}

		case "dotted_name":
			items = append(items, Treesitter.ImportItem{Module: module, Name: a.text(ch)})

		case "identifier":
			txt := a.text(ch)
			if txt != "" && txt != "import" && txt != "as" && txt != "*" {
				items = append(items, Treesitter.ImportItem{Module: module, Name: txt})
			}
		}
	}

	return dedup(items)
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
