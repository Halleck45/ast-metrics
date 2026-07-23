package java

import (
	"strings"

	"github.com/halleck45/ast-metrics/internal/engine"
	Treesitter "github.com/halleck45/ast-metrics/internal/engine/treesitter"
	sitter "github.com/smacker/go-tree-sitter"
	tsJava "github.com/smacker/go-tree-sitter/java"
)

type TreeSitterAdapter struct {
	src []byte
	// pkg caches the declared package name (parsed lazily from src)
	pkg       string
	pkgParsed bool
}

func NewTreeSitterAdapter(src []byte) *TreeSitterAdapter { return &TreeSitterAdapter{src: src} }
func (a *TreeSitterAdapter) SetSource(src []byte)        { a.src = src }
func (a *TreeSitterAdapter) Language() *sitter.Language  { return tsJava.GetLanguage() }

func (a *TreeSitterAdapter) IsModule(n *sitter.Node) bool { return n.Type() == "program" }

func (a *TreeSitterAdapter) IsClass(n *sitter.Node) bool {
	switch n.Type() {
	// enums, records and annotation types are class-like containers in Java:
	// they hold fields and methods, so treat them as classes for metrics.
	case "class_declaration", "enum_declaration", "record_declaration", "annotation_type_declaration":
		return true
	}
	return false
}

func (a *TreeSitterAdapter) IsInterface(n *sitter.Node) bool {
	return n.Type() == "interface_declaration"
}

func (a *TreeSitterAdapter) IsFunction(n *sitter.Node) bool {
	// Lambdas are intentionally not treated as named functions: they have no
	// name and would pollute method counts. Their bodies still contribute
	// decisions to the enclosing method via the fallback recursion.
	switch n.Type() {
	case "method_declaration", "constructor_declaration":
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
	if body := n.ChildByFieldName("body"); body != nil {
		return body
	}
	for _, t := range []string{"class_body", "interface_body", "enum_body", "annotation_type_body", "constructor_body", "block"} {
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
	return firstChildOfType(n, "formal_parameters")
}

func (a *TreeSitterAdapter) EachParamIdent(params *sitter.Node, yield func(string)) {
	if params == nil || a.src == nil {
		return
	}
	for i := 0; i < int(params.ChildCount()); i++ {
		p := params.Child(i)
		switch p.Type() {
		case "formal_parameter", "receiver_parameter":
			if nm := p.ChildByFieldName("name"); nm != nil {
				yield(text(a.src, nm))
			}
		case "spread_parameter":
			// int... rest -> (spread_parameter (variable_declarator name: identifier))
			if vd := firstChildOfType(p, "variable_declarator"); vd != nil {
				if nm := vd.ChildByFieldName("name"); nm != nil {
					yield(text(a.src, nm))
				}
			}
		}
	}
}

// ModuleNameFromPath ignores the file path and returns the declared package
// name (e.g. "com.example.app"), parsed once from the source. Empty string
// for the default package.
func (a *TreeSitterAdapter) ModuleNameFromPath(path string) string {
	if a.pkgParsed {
		return a.pkg
	}
	a.pkgParsed = true
	if a.src == nil {
		return ""
	}
	parser := sitter.NewParser()
	parser.SetLanguage(a.Language())
	tree := parser.Parse(nil, a.src)
	if tree == nil {
		return ""
	}
	root := tree.RootNode()
	for i := 0; i < int(root.ChildCount()); i++ {
		ch := root.Child(i)
		if ch.Type() != "package_declaration" {
			continue
		}
		for j := 0; j < int(ch.ChildCount()); j++ {
			c := ch.Child(j)
			if c.Type() == "scoped_identifier" || c.Type() == "identifier" {
				a.pkg = text(a.src, c)
				return a.pkg
			}
		}
	}
	return ""
}

func (a *TreeSitterAdapter) AttachQualified(parentClass string, fn string) string {
	if parentClass == "" {
		return fn
	}
	return parentClass + "." + fn
}

// NamespaceSeparator joins the package and the class name with "." (Java style).
func (a *TreeSitterAdapter) NamespaceSeparator() string { return "." }

func (a *TreeSitterAdapter) EachChildBody(body *sitter.Node, yield func(*sitter.Node)) {
	if body == nil {
		return
	}
	switch body.Type() {
	case "switch_block":
		// yield only the case groups: colon style groups and arrow rules
		for i := 0; i < int(body.ChildCount()); i++ {
			ch := body.Child(i)
			if ch.Type() == "switch_block_statement_group" || ch.Type() == "switch_rule" {
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
// if_statement (i.e. the else branch). tree-sitter-java has no else_clause
// wrapper node: `else if` is an if_statement directly in the alternative
// field, and a bare `else` body is a block/statement in that field.
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

	case "switch_expression":
		// the Java grammar uses switch_expression for both switch statements
		// and switch expressions
		return Treesitter.DecSwitch, n.ChildByFieldName("body")

	case "switch_block_statement_group", "switch_rule":
		return Treesitter.DecCase, n

	case "for_statement", "enhanced_for_statement", "while_statement", "do_statement":
		return Treesitter.DecLoop, n.ChildByFieldName("body")
	}
	// ternary_expression and catch_clause intentionally left out, consistent
	// with the other engines.
	return Treesitter.DecNone, nil
}

func (a *TreeSitterAdapter) Imports(n *sitter.Node) []Treesitter.ImportItem {
	if n == nil || n.Type() != "import_declaration" {
		return nil
	}
	var path string
	isWildcard := false
	for i := 0; i < int(n.ChildCount()); i++ {
		ch := n.Child(i)
		switch ch.Type() {
		case "scoped_identifier", "identifier":
			path = text(a.src, ch)
		case "asterisk":
			isWildcard = true
		}
	}
	if path == "" {
		return nil
	}
	if isWildcard {
		// import java.util.*;
		return []Treesitter.ImportItem{{Module: path, Name: ""}}
	}
	// import java.util.List; / import static org.junit.Assert.assertEquals;
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		return []Treesitter.ImportItem{{Module: path[:idx], Name: path[idx+1:]}}
	}
	return []Treesitter.ImportItem{{Module: path, Name: ""}}
}

// CountElseIfAsIf: treat else-if as if for complexity aggregation (consistent with Go/PHP/TS)
func (a *TreeSitterAdapter) CountElseIfAsIf() bool { return true }

// IsLogicalNode reports whether a node begins a logical line. In Java, local
// variable declarations are statements but their node type does not carry the
// "_statement" suffix.
func (a *TreeSitterAdapter) IsLogicalNode(n *sitter.Node) bool {
	if n.Type() == "local_variable_declaration" {
		return true
	}
	return Treesitter.IsDefaultLogicalNode(n.Type())
}

// CountComments counts Java comment lines (//, /* */ and /** */) in the given range.
// CommentMarkers declares Java comment tokens: "//" and "/* */" only.
// "#" has no meaning in Java source.
func (a *TreeSitterAdapter) CommentMarkers() engine.CommentMarkers {
	return engine.CommentMarkers{SlashSlash: true, SlashStar: true}
}

func (a *TreeSitterAdapter) CountComments(lines []string, start, end int) int {
	cnt := 0
	inBlock := false
	for i := start - 1; i < end && i < len(lines); i++ {
		ln := strings.TrimSpace(lines[i])
		if ln == "" {
			continue
		}
		clean := stripJavaStrings(ln)
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

// ExtractOperatorsOperands extracts Halstead operators and operands from Java source.
func (a *TreeSitterAdapter) ExtractOperatorsOperands(src []byte, startLine, endLine int) ([]string, []string) {
	if src == nil || startLine <= 0 || endLine <= 0 || endLine < startLine {
		return nil, nil
	}
	tokens := []string{
		">>>=", ">>=", "<<=", ">>>",
		"+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=",
		"==", "!=", "<=", ">=", "&&", "||",
		"++", "--", "->", "::",
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
		line := stripJavaStrings(raw)
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
			if f == "" || isJavaKeyword(f) {
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

// ExtractMethodCalls extracts method calls like this.foo, super.bar from Java source.
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
		ln = stripJavaStrings(ln)
		if idx := strings.Index(ln, "//"); idx >= 0 {
			ln = ln[:idx]
		}
		for _, prefix := range []string{"this.", "super."} {
			rest := ln
			for {
				idx := strings.Index(rest, prefix)
				if idx < 0 {
					break
				}
				after := rest[idx+len(prefix):]
				end := 0
				for end < len(after) && (after[end] == '_' || after[end] == '$' || (after[end] >= 'a' && after[end] <= 'z') || (after[end] >= 'A' && after[end] <= 'Z') || (after[end] >= '0' && after[end] <= '9')) {
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

// stripJavaStrings removes content inside string and char literals to avoid
// false positives in comment/operator scanning.
func stripJavaStrings(s string) string {
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

func isJavaKeyword(s string) bool {
	switch s {
	case "abstract", "assert", "boolean", "break", "byte", "case", "catch",
		"char", "class", "const", "continue", "default", "do", "double",
		"else", "enum", "extends", "final", "finally", "float", "for",
		"goto", "if", "implements", "import", "instanceof", "int",
		"interface", "long", "native", "new", "package", "private",
		"protected", "public", "record", "return", "sealed", "short",
		"static", "strictfp", "super", "switch", "synchronized", "this",
		"throw", "throws", "transient", "try", "var", "void", "volatile",
		"while", "yield", "permits", "non-sealed",
		"true", "false", "null":
		return true
	}
	return false
}
