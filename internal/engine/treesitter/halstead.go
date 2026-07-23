package treesitter

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// ExtractOperatorsOperandsFromAST collects Halstead operators and operands by
// walking the AST between startLine and endLine (1-based, inclusive).
// Operators are anonymous leaf tokens whose type belongs to operatorTokens
// (e.g. "+", "=="); operands are named nodes whose type belongs to
// operandTypes (identifiers and literals), reported by their source text.
// Strings and comments never produce tokens, so they are excluded by
// construction.
func ExtractOperatorsOperandsFromAST(root *sitter.Node, src []byte, startLine, endLine int, operatorTokens, operandTypes map[string]bool) ([]string, []string) {
	if root == nil || src == nil || startLine <= 0 || endLine < startLine {
		return nil, nil
	}

	ops := []string{}
	operands := []string{}

	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		// prune subtrees fully outside the line range
		if int(n.EndPoint().Row)+1 < startLine || int(n.StartPoint().Row)+1 > endLine {
			return
		}
		t := n.Type()
		if n.IsNamed() && operandTypes[t] {
			if start, end := n.StartByte(), n.EndByte(); int(end) <= len(src) {
				operands = append(operands, string(src[start:end]))
			}
			return
		}
		if !n.IsNamed() && n.ChildCount() == 0 && operatorTokens[t] {
			ops = append(ops, t)
			return
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(root)

	return ops, operands
}
