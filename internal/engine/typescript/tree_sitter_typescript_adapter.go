package typescript

import (
	"path/filepath"
	"strings"

	Treesitter "github.com/halleck45/ast-metrics/internal/engine/treesitter"
	sitter "github.com/smacker/go-tree-sitter"
	tsTsx "github.com/smacker/go-tree-sitter/typescript/tsx"
)

type TreeSitterAdapter struct {
	src []byte
}

func NewTreeSitterAdapter(src []byte) *TreeSitterAdapter { return &TreeSitterAdapter{src: src} }
func (a *TreeSitterAdapter) SetSource(src []byte)        { a.src = src }
func (a *TreeSitterAdapter) Language() *sitter.Language   { return tsTsx.GetLanguage() }

func (a *TreeSitterAdapter) IsModule(n *sitter.Node) bool { return n.Type() == "program" }

func (a *TreeSitterAdapter) IsClass(n *sitter.Node) bool {
	switch n.Type() {
	case "class_declaration", "abstract_class_declaration", "enum_declaration":
		return true
	}
	return false
}

func (a *TreeSitterAdapter) IsInterface(n *sitter.Node) bool {
	return n.Type() == "interface_declaration"
}

func (a *TreeSitterAdapter) IsFunction(n *sitter.Node) bool {
	switch n.Type() {
	case "function_declaration", "method_definition", "generator_function_declaration":
		return true
	case "arrow_function":
		return true
	}
	return false
}

func (a *TreeSitterAdapter) NodeName(n *sitter.Node) string {
	if a.src == nil || n == nil {
		return ""
	}

	// Arrow functions get their name from the parent variable_declarator
	if n.Type() == "arrow_function" || n.Type() == "function" {
		p := n.Parent()
		if p != nil && p.Type() == "variable_declarator" {
			if nm := p.ChildByFieldName("name"); nm != nil {
				return text(a.src, nm)
			}
		}
		// Arrow as class property: parent is public_field_definition or property_definition
		if p != nil && (p.Type() == "public_field_definition" || p.Type() == "property_definition") {
			if nm := p.ChildByFieldName("name"); nm != nil {
				return text(a.src, nm)
			}
			if id := firstChildOfType(p, "property_identifier"); id != nil {
				return text(a.src, id)
			}
		}
		return ""
	}

	// method_definition: name field
	if n.Type() == "method_definition" {
		if nm := n.ChildByFieldName("name"); nm != nil {
			return text(a.src, nm)
		}
		if id := firstChildOfType(n, "property_identifier"); id != nil {
			return text(a.src, id)
		}
		return ""
	}

	// class_declaration, abstract_class_declaration, enum_declaration,
	// function_declaration, generator_function_declaration, interface_declaration
	if nm := n.ChildByFieldName("name"); nm != nil {
		return text(a.src, nm)
	}
	if id := firstChildOfType(n, "identifier"); id != nil {
		return text(a.src, id)
	}
	if id := firstChildOfType(n, "type_identifier"); id != nil {
		return text(a.src, id)
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
	if b := firstChildOfType(n, "statement_block"); b != nil {
		return b
	}
	if b := firstChildOfType(n, "class_body"); b != nil {
		return b
	}
	if b := firstChildOfType(n, "enum_body"); b != nil {
		return b
	}
	if b := firstChildOfType(n, "object_type"); b != nil {
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
	if p := firstChildOfType(n, "formal_parameters"); p != nil {
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
		// Skip type annotations to avoid counting type names as parameters
		if n.Type() == "type_annotation" || n.Type() == "type_identifier" {
			return
		}
		if n.Type() == "identifier" || n.Type() == "shorthand_property_identifier_pattern" {
			yield(text(a.src, n))
			return
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(params)
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
	case "switch_body":
		for i := 0; i < int(body.ChildCount()); i++ {
			ch := body.Child(i)
			if ch.Type() == "switch_case" || ch.Type() == "switch_default" {
				yield(ch)
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
	case "if_statement":
		if b := n.ChildByFieldName("consequence"); b != nil {
			return Treesitter.DecIf, b
		}
		return Treesitter.DecIf, firstChildOfType(n, "statement_block")

	case "else_clause":
		// else if: the else clause contains an if_statement
		if ifNode := firstChildOfType(n, "if_statement"); ifNode != nil {
			return Treesitter.DecElif, nil
		}
		return Treesitter.DecElse, firstChildOfType(n, "statement_block")

	case "switch_statement":
		return Treesitter.DecSwitch, firstChildOfType(n, "switch_body")

	case "switch_case":
		return Treesitter.DecCase, n

	case "switch_default":
		return Treesitter.DecCase, n

	case "for_statement", "for_in_statement", "while_statement", "do_statement":
		if b := n.ChildByFieldName("body"); b != nil {
			return Treesitter.DecLoop, b
		}
		return Treesitter.DecLoop, firstChildOfType(n, "statement_block")
	}
	return Treesitter.DecNone, nil
}

func (a *TreeSitterAdapter) Imports(n *sitter.Node) []Treesitter.ImportItem {
	if n == nil {
		return nil
	}
	if n.Type() != "import_statement" {
		return nil
	}
	items := []Treesitter.ImportItem{}

	// Find the source module (the string literal at the end)
	var module string
	if src := n.ChildByFieldName("source"); src != nil {
		module = stripQuotes(text(a.src, src))
	} else {
		// fallback: find a string child
		for i := 0; i < int(n.ChildCount()); i++ {
			ch := n.Child(i)
			if ch.Type() == "string" {
				module = stripQuotes(text(a.src, ch))
				break
			}
		}
	}
	if module == "" {
		return nil
	}

	// Walk import clause children
	var walkClause func(*sitter.Node)
	walkClause = func(cl *sitter.Node) {
		if cl == nil {
			return
		}
		switch cl.Type() {
		case "import_clause":
			for i := 0; i < int(cl.ChildCount()); i++ {
				walkClause(cl.Child(i))
			}
		case "identifier":
			// default import: import X from 'module'
			items = append(items, Treesitter.ImportItem{Module: module, Name: text(a.src, cl)})
		case "named_imports":
			for i := 0; i < int(cl.ChildCount()); i++ {
				spec := cl.Child(i)
				if spec.Type() == "import_specifier" {
					if nm := spec.ChildByFieldName("name"); nm != nil {
						items = append(items, Treesitter.ImportItem{Module: module, Name: text(a.src, nm)})
					} else if id := firstChildOfType(spec, "identifier"); id != nil {
						items = append(items, Treesitter.ImportItem{Module: module, Name: text(a.src, id)})
					}
				}
			}
		case "namespace_import":
			// import * as X from 'module'
			if id := firstChildOfType(cl, "identifier"); id != nil {
				items = append(items, Treesitter.ImportItem{Module: module, Name: text(a.src, id)})
			} else {
				items = append(items, Treesitter.ImportItem{Module: module})
			}
		}
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		walkClause(n.Child(i))
	}

	// If no symbols found, record as plain module import
	if len(items) == 0 {
		items = append(items, Treesitter.ImportItem{Module: module})
	}
	return items
}

// CountElseIfAsIf: treat else-if as if for complexity aggregation (consistent with Go/PHP)
func (a *TreeSitterAdapter) CountElseIfAsIf() bool { return true }

// FileLlocOffset returns the offset to subtract when computing file-level LLOC.
func (a *TreeSitterAdapter) FileLlocOffset() int { return 2 }

// CountComments counts TypeScript comment lines (// and /* */ and /** */) in the given range.
func (a *TreeSitterAdapter) CountComments(lines []string, start, end int) int {
	cnt := 0
	inBlock := false
	for i := start - 1; i < end && i < len(lines); i++ {
		ln := strings.TrimSpace(lines[i])
		if ln == "" {
			continue
		}
		clean := stripTSStrings(ln)
		if inBlock {
			cnt++
			if strings.Contains(clean, "*/") {
				inBlock = false
			}
			continue
		}
		if strings.HasPrefix(clean, "//") {
			cnt++
			continue
		}
		if strings.HasPrefix(clean, "/*") || strings.HasPrefix(clean, "/**") {
			cnt++
			if !strings.Contains(clean, "*/") {
				inBlock = true
			}
			continue
		}
	}
	return cnt
}

// ExtractOperatorsOperands extracts Halstead operators and operands from TypeScript source.
func (a *TreeSitterAdapter) ExtractOperatorsOperands(src []byte, startLine, endLine int) ([]string, []string) {
	if src == nil || startLine <= 0 || endLine <= 0 || endLine < startLine {
		return nil, nil
	}
	tokens := []string{
		">>>=", "===", "!==", ">>=", "<<=", "**=", "??=",
		"+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=",
		"==", "!=", "<=", ">=", "&&", "||", "??", "?.",
		"++", "--", "=>", "...", "**",
		">>>", "<<", ">>",
		"+", "-", "*", "/", "%", "&", "|", "^", "!", "<", ">", "=", "~",
		".",
	}

	lines := strings.Split(string(src), "\n")
	ops := []string{}
	opr := []string{}

	for i := startLine - 1; i < endLine && i < len(lines); i++ {
		raw := strings.TrimSpace(lines[i])
		if raw == "" {
			continue
		}
		line := raw
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		line = stripTSStrings(line)
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Scan operators
		rest := line
		for {
			found := false
			minPos := len(rest)
			minTok := ""
			for _, tok := range tokens {
				if p := strings.Index(rest, tok); p >= 0 && p < minPos {
					minPos = p
					minTok = tok
					found = true
				}
			}
			if !found {
				break
			}
			ops = append(ops, minTok)
			rest = rest[minPos+len(minTok):]
		}

		// Operands: identifiers
		cleaned := line
		replacers := []string{",", ";", "(", ")", "[", "]", "{", "}", "*", "&", "|", "^", "/", "+", "-", "%", ":", "<", ">", "=", "!", "~", "?", "."}
		for _, r := range replacers {
			cleaned = strings.ReplaceAll(cleaned, r, " ")
		}
		fields := strings.Fields(cleaned)
		for _, f := range fields {
			if f == "" || isTSKeyword(f) {
				continue
			}
			if f[0] >= '0' && f[0] <= '9' {
				continue
			}
			opr = append(opr, f)
		}
	}
	return ops, opr
}

// ExtractMethodCalls extracts method calls like this.foo, super.bar from TypeScript source.
func (a *TreeSitterAdapter) ExtractMethodCalls(src []byte, startLine, endLine int) []string {
	if src == nil || startLine <= 0 || endLine <= 0 || endLine < startLine {
		return nil
	}
	lines := strings.Split(string(src), "\n")
	var calls []string
	for i := startLine - 1; i < endLine && i < len(lines); i++ {
		ln := strings.TrimSpace(lines[i])
		if ln == "" {
			continue
		}
		// Strip comments
		if idx := strings.Index(ln, "//"); idx >= 0 {
			ln = ln[:idx]
		}
		ln = stripTSStrings(ln)
		// Find this.xxx( or super.xxx( patterns
		for _, prefix := range []string{"this.", "super."} {
			rest := ln
			for {
				idx := strings.Index(rest, prefix)
				if idx < 0 {
					break
				}
				after := rest[idx+len(prefix):]
				// Extract identifier
				end := 0
				for end < len(after) && (after[end] == '_' || after[end] == '$' || (after[end] >= 'a' && after[end] <= 'z') || (after[end] >= 'A' && after[end] <= 'Z') || (after[end] >= '0' && after[end] <= '9')) {
					end++
				}
				if end > 0 {
					name := prefix[:len(prefix)-1] + "." + after[:end]
					calls = append(calls, name)
				}
				rest = after[end:]
			}
		}
	}
	return calls
}

// ClassDirectOperands scans class body for property declarations and returns property names.
func (a *TreeSitterAdapter) ClassDirectOperands(n *sitter.Node) []string {
	if n == nil || a.src == nil {
		return nil
	}
	body := a.NodeBody(n)
	if body == nil {
		return nil
	}
	var props []string
	for i := 0; i < int(body.ChildCount()); i++ {
		ch := body.Child(i)
		switch ch.Type() {
		case "public_field_definition", "property_definition":
			if nm := ch.ChildByFieldName("name"); nm != nil {
				props = append(props, text(a.src, nm))
			} else if id := firstChildOfType(ch, "property_identifier"); id != nil {
				props = append(props, text(a.src, id))
			}
		}
	}
	return props
}

// --- helpers ---

func text(src []byte, n *sitter.Node) string {
	if n == nil || src == nil {
		return ""
	}
	return string(src[n.StartByte():n.EndByte()])
}

func firstChildOfType(n *sitter.Node, t string) *sitter.Node {
	if n == nil {
		return nil
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		if c := n.Child(i); c.Type() == t {
			return c
		}
	}
	return nil
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '`' && s[len(s)-1] == '`') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// stripTSStrings removes content inside quotes to avoid false positives in comment/operator scanning.
func stripTSStrings(s string) string {
	out := make([]rune, 0, len(s))
	inBack := false
	inDq := false
	inSq := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' {
			if i+1 < len(s) {
				i++
			}
			continue
		}
		if !inDq && !inSq && c == '`' {
			inBack = !inBack
			continue
		}
		if !inBack && !inSq && c == '"' {
			inDq = !inDq
			continue
		}
		if !inBack && !inDq && c == '\'' {
			inSq = !inSq
			continue
		}
		if inBack || inDq || inSq {
			continue
		}
		out = append(out, rune(c))
	}
	return string(out)
}

func isTSKeyword(s string) bool {
	switch s {
	case "import", "export", "from", "as", "default",
		"function", "class", "extends", "implements", "interface",
		"type", "enum", "namespace", "module", "declare",
		"const", "let", "var", "return", "yield",
		"if", "else", "for", "while", "do", "switch", "case", "break", "continue",
		"try", "catch", "finally", "throw",
		"new", "delete", "typeof", "instanceof", "void", "in", "of",
		"async", "await", "static", "get", "set",
		"public", "private", "protected", "readonly", "abstract", "override",
		"true", "false", "null", "undefined",
		"this", "super", "constructor":
		return true
	}
	return false
}
