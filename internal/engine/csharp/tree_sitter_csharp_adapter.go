package csharp

import (
	"strings"

	Treesitter "github.com/halleck45/ast-metrics/internal/engine/treesitter"
	sitter "github.com/smacker/go-tree-sitter"
	tsCSharp "github.com/smacker/go-tree-sitter/csharp"
)

type TreeSitterAdapter struct {
	src []byte
	// ns caches the declared namespace (parsed lazily from src)
	ns       string
	nsParsed bool
}

func NewTreeSitterAdapter(src []byte) *TreeSitterAdapter { return &TreeSitterAdapter{src: src} }
func (a *TreeSitterAdapter) SetSource(src []byte)        { a.src = src }
func (a *TreeSitterAdapter) Language() *sitter.Language  { return tsCSharp.GetLanguage() }

func (a *TreeSitterAdapter) IsModule(n *sitter.Node) bool {
	// namespace_declaration bodies are reached through the fallback recursion;
	// the namespace name itself comes from ModuleNameFromPath
	return n.Type() == "compilation_unit"
}

func (a *TreeSitterAdapter) IsClass(n *sitter.Node) bool {
	switch n.Type() {
	// structs, records (incl. record struct) and enums are class-like
	// containers in C#: they hold fields and methods, so treat them as
	// classes for metrics.
	case "class_declaration", "struct_declaration", "record_declaration", "enum_declaration":
		return true
	}
	return false
}

func (a *TreeSitterAdapter) IsInterface(n *sitter.Node) bool {
	return n.Type() == "interface_declaration"
}

func (a *TreeSitterAdapter) IsFunction(n *sitter.Node) bool {
	// Property accessors (get/set/init) and lambdas are intentionally not
	// treated as named functions; their bodies still contribute decisions via
	// the fallback recursion.
	switch n.Type() {
	case "method_declaration", "constructor_declaration", "local_function_statement", "destructor_declaration":
		return true
	}
	return false
}

func (a *TreeSitterAdapter) NodeName(n *sitter.Node) string {
	if a.src == nil || n == nil {
		return ""
	}
	if nm := n.ChildByFieldName("name"); nm != nil {
		return text(a.src, nm)
	}
	if id := firstChildOfType(n, "identifier"); id != nil {
		return text(a.src, id)
	}
	return ""
}

func (a *TreeSitterAdapter) NodeBody(n *sitter.Node) *sitter.Node {
	if n == nil {
		return nil
	}
	// body field covers: declaration_list (types), block (methods),
	// arrow_expression_clause (expression-bodied members),
	// enum_member_declaration_list (enums)
	if body := n.ChildByFieldName("body"); body != nil {
		return body
	}
	for _, t := range []string{"declaration_list", "enum_member_declaration_list", "block", "arrow_expression_clause"} {
		if b := firstChildOfType(n, t); b != nil {
			return b
		}
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
	return firstChildOfType(n, "parameter_list")
}

func (a *TreeSitterAdapter) EachParamIdent(params *sitter.Node, yield func(string)) {
	if params == nil || a.src == nil {
		return
	}
	for i := 0; i < int(params.ChildCount()); i++ {
		p := params.Child(i)
		if p.Type() != "parameter" {
			continue
		}
		// skip type and ref/out/in modifiers; only the name field is a parameter name
		if nm := p.ChildByFieldName("name"); nm != nil {
			yield(text(a.src, nm))
		}
	}
}

// ModuleNameFromPath ignores the file path and returns the declared namespace
// (e.g. "App.Services"), parsed once from the source. Handles both block
// namespaces and file-scoped namespaces. When several block namespaces are
// declared, the first one is used (the shared Visitor stores a single
// namespace per file). Empty string when no namespace is declared.
func (a *TreeSitterAdapter) ModuleNameFromPath(path string) string {
	if a.nsParsed {
		return a.ns
	}
	a.nsParsed = true
	if a.src == nil {
		return ""
	}
	parser := sitter.NewParser()
	parser.SetLanguage(a.Language())
	tree := parser.Parse(nil, a.src)
	if tree == nil {
		return ""
	}
	var find func(n *sitter.Node) string
	find = func(n *sitter.Node) string {
		if n.Type() == "namespace_declaration" || n.Type() == "file_scoped_namespace_declaration" {
			if nm := n.ChildByFieldName("name"); nm != nil {
				return text(a.src, nm)
			}
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			if found := find(n.Child(i)); found != "" {
				return found
			}
		}
		return ""
	}
	a.ns = find(tree.RootNode())
	return a.ns
}

func (a *TreeSitterAdapter) AttachQualified(parentClass string, fn string) string {
	if parentClass == "" {
		return fn
	}
	return parentClass + "." + fn
}

// NamespaceSeparator joins the namespace and the class name with "." (C# style).
func (a *TreeSitterAdapter) NamespaceSeparator() string { return "." }

func (a *TreeSitterAdapter) EachChildBody(body *sitter.Node, yield func(*sitter.Node)) {
	if body == nil {
		return
	}
	switch body.Type() {
	case "switch_body":
		// yield only the case sections
		for i := 0; i < int(body.ChildCount()); i++ {
			ch := body.Child(i)
			if ch.Type() == "switch_section" {
				yield(ch)
			}
		}
	case "switch_expression":
		// switch expressions have no body node: arms are direct children
		for i := 0; i < int(body.ChildCount()); i++ {
			ch := body.Child(i)
			if ch.Type() == "switch_expression_arm" {
				yield(ch)
			}
		}
	default:
		for i := 0; i < int(body.ChildCount()); i++ {
			yield(body.Child(i))
		}
	}
}

// isAlternativeOfIf reports whether n is the "alternative" field of a parent
// if_statement (i.e. the else branch). Like tree-sitter-java, the C# grammar
// has no else_clause wrapper node: `else if` is an if_statement directly in
// the alternative field, and a bare `else` body is a block/statement there.
func isAlternativeOfIf(n *sitter.Node) bool {
	p := n.Parent()
	if p == nil || p.Type() != "if_statement" {
		return false
	}
	alt := p.ChildByFieldName("alternative")
	return alt != nil && alt.StartByte() == n.StartByte() && alt.EndByte() == n.EndByte() && alt.Type() == n.Type()
}

func (a *TreeSitterAdapter) Decision(n *sitter.Node) (Treesitter.DecisionKind, *sitter.Node) {
	if isAlternativeOfIf(n) {
		if n.Type() == "if_statement" {
			// else-if: return the node itself as body so the whole chain
			// (condition, consequence, nested alternative) is re-visited and
			// deeper else-if/else branches keep being counted.
			return Treesitter.DecElif, n
		}
		// bare else branch (block or single statement)
		return Treesitter.DecElse, n
	}

	switch n.Type() {
	case "if_statement":
		return Treesitter.DecIf, n.ChildByFieldName("consequence")

	case "switch_statement":
		return Treesitter.DecSwitch, n.ChildByFieldName("body")

	case "switch_expression":
		// arms are direct children; EachChildBody filters them
		return Treesitter.DecSwitch, n

	case "switch_section", "switch_expression_arm":
		return Treesitter.DecCase, n

	case "for_statement", "foreach_statement", "while_statement", "do_statement":
		return Treesitter.DecLoop, n.ChildByFieldName("body")
	}
	// conditional_expression (ternary) and catch_clause intentionally left
	// out, consistent with the other engines.
	return Treesitter.DecNone, nil
}

func (a *TreeSitterAdapter) Imports(n *sitter.Node) []Treesitter.ImportItem {
	if n == nil || n.Type() != "using_directive" {
		return nil
	}
	// using System;                       -> {Module: "System"}
	// using static System.Math;          -> {Module: "System.Math"}
	// global using System.Linq;          -> {Module: "System.Linq"}
	// using Foo = System.Text.Builder;   -> {Module: "System.Text.Builder", Name: "Foo"}
	alias := ""
	module := ""
	if nm := n.ChildByFieldName("name"); nm != nil {
		// alias form: the name field is the alias identifier
		alias = text(a.src, nm)
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		ch := n.Child(i)
		switch ch.Type() {
		case "qualified_name":
			module = text(a.src, ch)
		case "identifier":
			// plain single-segment using; for the alias form the alias is the
			// name field, the target is the qualified_name/identifier after "="
			if alias != "" && text(a.src, ch) == alias && module == "" {
				// this child IS the alias; skip, the target comes after "="
				continue
			}
			module = text(a.src, ch)
		}
	}
	if module == "" {
		return nil
	}
	return []Treesitter.ImportItem{{Module: module, Name: alias}}
}

// CountElseIfAsIf: treat else-if as if for complexity aggregation (consistent with Go/PHP/TS)
func (a *TreeSitterAdapter) CountElseIfAsIf() bool { return true }

// FileLlocOffset returns the offset to subtract when computing file-level LLOC.
func (a *TreeSitterAdapter) FileLlocOffset() int { return 2 }

// CountComments counts C# comment lines (//, /// and /* */) in the given range.
func (a *TreeSitterAdapter) CountComments(lines []string, start, end int) int {
	cnt := 0
	inBlock := false
	for i := start - 1; i < end && i < len(lines); i++ {
		ln := strings.TrimSpace(lines[i])
		if ln == "" {
			continue
		}
		clean := stripCSharpStrings(ln)
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
		if strings.HasPrefix(clean, "/*") {
			cnt++
			if !strings.Contains(clean, "*/") {
				inBlock = true
			}
			continue
		}
	}
	return cnt
}

// ExtractOperatorsOperands extracts Halstead operators and operands from C# source.
func (a *TreeSitterAdapter) ExtractOperatorsOperands(src []byte, startLine, endLine int) ([]string, []string) {
	if src == nil || startLine <= 0 || endLine <= 0 || endLine < startLine {
		return nil, nil
	}
	tokens := []string{
		">>>=", ">>=", "<<=", "??=", ">>>",
		"+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=",
		"==", "!=", "<=", ">=", "&&", "||", "??", "?.",
		"++", "--", "=>",
		"<<", ">>",
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
		line := stripCSharpStrings(raw)
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
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
		replacers := []string{",", ";", "(", ")", "[", "]", "{", "}", "*", "&", "|", "^", "/", "+", "-", "%", ":", "<", ">", "=", "!", "~", "?", ".", "@"}
		for _, r := range replacers {
			cleaned = strings.ReplaceAll(cleaned, r, " ")
		}
		fields := strings.Fields(cleaned)
		for _, f := range fields {
			if f == "" || isCSharpKeyword(f) {
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

// ExtractMethodCalls extracts method calls like this.Foo, base.Bar from C# source.
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
		ln = stripCSharpStrings(ln)
		if idx := strings.Index(ln, "//"); idx >= 0 {
			ln = ln[:idx]
		}
		for _, prefix := range []string{"this.", "base."} {
			rest := ln
			for {
				idx := strings.Index(rest, prefix)
				if idx < 0 {
					break
				}
				after := rest[idx+len(prefix):]
				end := 0
				for end < len(after) && (after[end] == '_' || (after[end] >= 'a' && after[end] <= 'z') || (after[end] >= 'A' && after[end] <= 'Z') || (after[end] >= '0' && after[end] <= '9')) {
					end++
				}
				if end > 0 {
					calls = append(calls, prefix[:len(prefix)-1]+"."+after[:end])
				}
				rest = after[end:]
			}
		}
	}
	return calls
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

// stripCSharpStrings removes content inside string and char literals to avoid
// false positives in comment/operator scanning.
func stripCSharpStrings(s string) string {
	out := make([]rune, 0, len(s))
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
		if !inSq && c == '"' {
			inDq = !inDq
			continue
		}
		if !inDq && c == '\'' {
			inSq = !inSq
			continue
		}
		if inDq || inSq {
			continue
		}
		out = append(out, rune(c))
	}
	return string(out)
}

func isCSharpKeyword(s string) bool {
	switch s {
	case "abstract", "as", "base", "bool", "break", "byte", "case", "catch",
		"char", "checked", "class", "const", "continue", "decimal", "default",
		"delegate", "do", "double", "else", "enum", "event", "explicit",
		"extern", "finally", "fixed", "float", "for", "foreach", "goto",
		"if", "implicit", "in", "int", "interface", "internal", "is", "lock",
		"long", "namespace", "new", "object", "operator", "out", "override",
		"params", "private", "protected", "public", "readonly", "record",
		"ref", "return", "sbyte", "sealed", "short", "sizeof", "stackalloc",
		"static", "string", "struct", "switch", "this", "throw", "try",
		"typeof", "uint", "ulong", "unchecked", "unsafe", "ushort", "using",
		"var", "virtual", "void", "volatile", "while", "yield",
		"async", "await", "get", "set", "init", "value", "global", "partial",
		"when", "where", "with", "or", "and", "not",
		"true", "false", "null":
		return true
	}
	return false
}
