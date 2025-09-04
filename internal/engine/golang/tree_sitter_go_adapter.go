package golang

import (
    "path/filepath"
    "strings"

    Treesitter "github.com/halleck45/ast-metrics/internal/engine/treesitter"
    tsGo "github.com/smacker/go-tree-sitter/golang"
    sitter "github.com/smacker/go-tree-sitter"
)

type TreeSitterAdapter struct {
    src []byte
}

func NewTreeSitterAdapter(src []byte) *TreeSitterAdapter { return &TreeSitterAdapter{src: src} }
func (a *TreeSitterAdapter) SetSource(src []byte)        { a.src = src }
func (a *TreeSitterAdapter) Language() *sitter.Language  { return tsGo.GetLanguage() }

func (a *TreeSitterAdapter) IsModule(n *sitter.Node) bool { return n.Type() == "source_file" }
func (a *TreeSitterAdapter) IsClass(n *sitter.Node) bool  { return n.Type() == "type_declaration" && firstChildOfType(n, "type_spec") != nil && firstDescendantOfType(n, "type_identifier") != nil && firstDescendantOfType(n, "type_parameter_list") == nil && firstDescendantOfType(n, "struct_type") != nil }
func (a *TreeSitterAdapter) IsFunction(n *sitter.Node) bool { return n.Type() == "function_declaration" || n.Type() == "method_declaration" }

func (a *TreeSitterAdapter) NodeName(n *sitter.Node) string {
    switch n.Type() {
    case "function_declaration":
        if id := firstChildOfType(n, "identifier"); id != nil { return text(a.src, id) }
    case "method_declaration":
        if id := firstChildOfType(n, "field_identifier"); id != nil { return text(a.src, id) }
    case "type_declaration":
        if id := firstDescendantOfType(n, "type_identifier"); id != nil { return text(a.src, id) }
    }
    return ""
}

func (a *TreeSitterAdapter) NodeBody(n *sitter.Node) *sitter.Node {
    switch n.Type() {
    case "function_declaration":
        return firstChildOfType(n, "block")
    case "method_declaration":
        return firstChildOfType(n, "block")
    case "type_declaration":
        return firstDescendantOfType(n, "field_declaration_list")
    }
    return nil
}

func (a *TreeSitterAdapter) NodeParams(n *sitter.Node) *sitter.Node {
    switch n.Type() {
    case "function_declaration", "method_declaration":
        return firstChildOfType(n, "parameter_list")
    }
    return nil
}

func (a *TreeSitterAdapter) EachParamIdent(params *sitter.Node, yield func(string)) {
    if params == nil { return }
    for i := 0; i < int(params.ChildCount()); i++ {
        p := params.Child(i)
        if p.Type() == "parameter_declaration" {
            // collect identifiers under this param decl
            eachDescendantOfType(p, "identifier", func(id *sitter.Node) { yield(text(a.src, id)) })
        }
    }
}

func (a *TreeSitterAdapter) ModuleNameFromPath(path string) string {
    base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
    return base
}

func (a *TreeSitterAdapter) AttachQualified(parent string, fn string) string {
    if parent == "" { return fn }
    return parent + "." + fn
}

func (a *TreeSitterAdapter) EachChildBody(body *sitter.Node, yield func(*sitter.Node)) {
    if body == nil { return }
    switch body.Type() {
    case "switch_statement", "expression_switch_statement", "type_switch_statement":
        // Yield case_clause nodes wherever they are in the subtree
        var walk func(*sitter.Node)
        walk = func(n *sitter.Node) {
            if n == nil { return }
            if n.Type() == "case_clause" {
                yield(n)
                // do not return; keep walking
            }
            for i := 0; i < int(n.ChildCount()); i++ { walk(n.Child(i)) }
        }
        walk(body)
    default:
        for i := 0; i < int(body.ChildCount()); i++ { yield(body.Child(i)) }
    }
}

func (a *TreeSitterAdapter) Decision(n *sitter.Node) (Treesitter.DecisionKind, *sitter.Node) {
    switch n.Type() {
    case "if_statement":
        return Treesitter.DecIf, firstChildOfType(n, "block")
    case "else_clause", "else":
        // could be else if or else
        // detect else if by presence of an if-node inside the else structure
        var foundIf bool
        var walk func(*sitter.Node)
        walk = func(x *sitter.Node) {
            if x == nil || foundIf { return }
            if strings.Contains(x.Type(), "if") {
                foundIf = true
                return
            }
            for i := 0; i < int(x.ChildCount()); i++ { walk(x.Child(i)) }
        }
        walk(n)
        if foundIf {
            return Treesitter.DecElif, firstDescendantOfType(n, "block")
        }
        return Treesitter.DecElse, firstDescendantOfType(n, "block")
    case "switch_statement", "expression_switch_statement", "type_switch_statement":
        // return the node itself; EachChildBody will yield case_clause nodes
        return Treesitter.DecSwitch, n
    case "case_clause":
        if b := firstDescendantOfType(n, "statement_block"); b != nil {
            return Treesitter.DecCase, b
        }
        return Treesitter.DecCase, firstDescendantOfType(n, "block")
    default:
        // Some grammars may name case nodes differently (e.g., "case_statement", "case")
        if strings.Contains(n.Type(), "case") && n.Type() != "case_identifier" {
            return Treesitter.DecCase, firstDescendantOfType(n, "block")
        }
    case "for_statement", "range_clause":
        return Treesitter.DecLoop, firstDescendantOfType(n, "block")
    }
    return Treesitter.DecNone, nil
}

func (a *TreeSitterAdapter) Imports(n *sitter.Node) []Treesitter.ImportItem {
    if n == nil { return nil }
    items := []Treesitter.ImportItem{}
    switch n.Type() {
    case "import_declaration":
        // walk import specs
        var walk func(*sitter.Node)
        walk = func(x *sitter.Node) {
            if x == nil { return }
            if x.Type() == "import_spec" {
                var module string
                var alias string
                // path is string_literal
                if p := firstDescendantOfType(x, "interpreted_string_literal"); p != nil {
                    module = strings.Trim(text(a.src, p), "`\"")
                } else if q := firstDescendantOfType(x, "raw_string_literal"); q != nil {
                    module = strings.Trim(text(a.src, q), "`\"")
                }
                // alias is optional identifier as first child
                if id := firstChildOfType(x, "identifier"); id != nil { alias = text(a.src, id) }
                name := alias
                if name == "" {
                    // default to last segment
                    if idx := strings.LastIndex(module, "/"); idx >= 0 { name = module[idx+1:] } else { name = module }
                }
                if module != "" {
                    items = append(items, Treesitter.ImportItem{Module: module, Name: name})
                }
            }
            for i := 0; i < int(x.ChildCount()); i++ { walk(x.Child(i)) }
        }
        walk(n)
    }
    return items
}

// helpers
func text(src []byte, n *sitter.Node) string { return string(src[n.StartByte():n.EndByte()]) }
func firstChildOfType(n *sitter.Node, t string) *sitter.Node {
    for i := 0; i < int(n.ChildCount()); i++ { if c := n.Child(i); c.Type() == t { return c } }
    return nil
}
func firstDescendantOfType(n *sitter.Node, t string) *sitter.Node {
    var res *sitter.Node
    eachDescendantOfType(n, t, func(n *sitter.Node){ if res == nil { res = n } })
    return res
}
func eachDescendantOfType(n *sitter.Node, t string, yield func(*sitter.Node)) {
    for i := 0; i < int(n.ChildCount()); i++ {
        c := n.Child(i)
        if c.Type() == t { yield(c) }
        eachDescendantOfType(c, t, yield)
    }
}

// Align with PHP counting: treat else-if as if for complexity aggregation
func (a *TreeSitterAdapter) CountElseIfAsIf() bool { return true }
