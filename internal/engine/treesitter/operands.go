package treesitter

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// OperandSpec drives the AST-based extraction of Halstead operators and
// operands for languages whose objects carry attributes. On top of the plain
// token walk of ExtractOperatorsOperandsFromAST, it prunes type positions (a
// type annotation is not an operand), keeps member access chains whole
// ("a.b.c" is one operand, not three) and normalizes accesses made through the
// current object as "this.attr", the form the cohesion metrics (LCOM4) expect.
type OperandSpec struct {
	// OperatorTokens lists the anonymous leaf token types counted as operators.
	OperatorTokens map[string]bool
	// OperandTypes lists the named node types counted as operands.
	OperandTypes map[string]bool
	// PruneTypes lists the node types whose whole subtree yields nothing:
	// type annotations, type arguments, type declarations.
	PruneTypes map[string]bool
	// PruneFields lists the field names whose child is skipped, for children
	// that hold no operand but have no dedicated node type (the "receiver" and
	// "result" of a Go method).
	PruneFields map[string]bool
	// ChainTypes lists the member access node types ("selector_expression" in
	// Go, "member_expression" in TypeScript).
	ChainTypes map[string]bool
	// Receiver names the variable standing for the current object (the
	// receiver of a Go method). Languages where the object is a keyword
	// ("this", "super") leave it empty.
	Receiver string
}

// Extract walks the AST between startLine and endLine (1-based, inclusive) and
// returns the Halstead operators and operands found there. Strings and
// comments never produce tokens, so they are excluded by construction.
func (s OperandSpec) Extract(root *sitter.Node, src []byte, startLine, endLine int) ([]string, []string) {
	if root == nil || src == nil || startLine <= 0 || endLine < startLine {
		return nil, nil
	}
	ops := []string{}
	operands := []string{}

	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if outsideLines(n, startLine, endLine) {
			return
		}
		t := n.Type()
		if s.PruneTypes[t] {
			return
		}
		if s.ChainTypes[t] {
			if operand, ok := s.chainOperand(n, src); ok {
				if operand != "" {
					operands = append(operands, operand)
				}
				return
			}
			// not a plain chain ("foo().bar"): its property names nothing on
			// its own, only the object is worth walking
			if object := n.NamedChild(0); object != nil {
				walk(object)
			}
			return
		}
		if n.IsNamed() && s.OperandTypes[t] {
			// the object alone ("return e") names no attribute
			if txt := nodeText(src, n); s.Receiver == "" || txt != s.Receiver {
				operands = append(operands, txt)
			}
			return
		}
		if !n.IsNamed() && n.ChildCount() == 0 && s.OperatorTokens[t] {
			ops = append(ops, t)
			return
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			if s.PruneFields[n.FieldNameForChild(i)] {
				continue
			}
			walk(n.Child(i))
		}
	}
	walk(root)
	return ops, operands
}

// MethodCalls returns the methods called on the current object between
// startLine and endLine, normalized as "this.name" ("super.name" for a call on
// the parent class). Calls made on another variable or on a package say
// nothing about the cohesion of the class and are not reported.
func (s OperandSpec) MethodCalls(root *sitter.Node, src []byte, startLine, endLine int) []string {
	if root == nil || src == nil || startLine <= 0 || endLine < startLine {
		return nil
	}
	var calls []string
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if outsideLines(n, startLine, endLine) {
			return
		}
		if n.Type() == "call_expression" {
			if fn := n.ChildByFieldName("function"); fn != nil && s.ChainTypes[fn.Type()] {
				if segments, ok := s.flatten(fn, src); ok && len(segments) == 2 {
					if object := s.normalizeObject(segments[0]); object != "" {
						calls = append(calls, object+"."+segments[1])
					}
				}
			}
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(root)
	return calls
}

// chainOperand renders a member access chain as a single operand. An access
// through the current object reads an attribute, whatever is done with it
// afterwards: "this.items", "this.items.length" and "this.items.push()" all
// read "this.items" (Go receivers are normalized the same way: "e.counter"
// gives "this.counter"). A direct call on the object ("this.load()") reads no
// attribute at all: it is a method call, reported by MethodCalls. ok is false
// when the chain is not a plain chain of identifiers ("foo().bar").
func (s OperandSpec) chainOperand(n *sitter.Node, src []byte) (operand string, ok bool) {
	segments, ok := s.flatten(n, src)
	if !ok {
		return "", false
	}
	if object := s.normalizeObject(segments[0]); object != "" {
		if len(segments) == 2 && isCallee(n) {
			return "", true
		}
		return object + "." + segments[1], true
	}
	return strings.Join(segments, "."), true
}

// normalizeObject returns "this" for the current object ("this", the Go
// receiver), "super" for the parent class, and "" for anything else.
func (s OperandSpec) normalizeObject(root string) string {
	if s.Receiver != "" && root == s.Receiver {
		return "this"
	}
	if root == "this" || root == "super" {
		return root
	}
	return ""
}

// flatten splits a member access chain into its segments: `this.items.length`
// gives ["this", "items", "length"]. ok is false when a link is not a plain
// identifier (a call, an index, a literal).
func (s OperandSpec) flatten(n *sitter.Node, src []byte) ([]string, bool) {
	if s.ChainTypes[n.Type()] {
		if n.NamedChildCount() != 2 {
			return nil, false
		}
		left, ok := s.flatten(n.NamedChild(0), src)
		if !ok {
			return nil, false
		}
		return append(left, nodeText(src, n.NamedChild(1))), true
	}
	if n.NamedChildCount() == 0 && !strings.HasSuffix(n.Type(), "literal") {
		return []string{nodeText(src, n)}, true
	}
	return nil, false
}

// isCallee reports whether the node is the function of a call expression.
func isCallee(n *sitter.Node) bool {
	parent := n.Parent()
	if parent == nil || parent.Type() != "call_expression" {
		return false
	}
	fn := parent.ChildByFieldName("function")
	return fn != nil && fn.StartByte() == n.StartByte() && fn.EndByte() == n.EndByte()
}

func outsideLines(n *sitter.Node, startLine, endLine int) bool {
	return int(n.EndPoint().Row)+1 < startLine || int(n.StartPoint().Row)+1 > endLine
}

func nodeText(src []byte, n *sitter.Node) string {
	if int(n.EndByte()) > len(src) {
		return ""
	}
	return string(src[n.StartByte():n.EndByte()])
}
