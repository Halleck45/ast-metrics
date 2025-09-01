package analyzer

import (
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// Represents the AST as it is visited
type ASTNode struct {
	children *pb.Stmts
	Visitors []Visitor
}

func (n *ASTNode) Accept(visitor Visitor) {
	n.Visitors = append(n.Visitors, visitor)
}

func (n *ASTNode) Visit() {
	for _, v := range n.Visitors {
		recurser := &HelperRecursionVisitor{}
		recurser.Recurse(n.children, v)
	}
}
