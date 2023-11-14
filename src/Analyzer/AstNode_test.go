package Analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type MockVisitor struct {
	visited bool
}

func (v *MockVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {
	v.visited = true
}

func (v *MockVisitor) LeaveNode(stmts *pb.Stmts) {
}

func TestASTNode_Accept(t *testing.T) {
	node := ASTNode{
		children: &pb.Stmts{},
		Visitors: []Visitor{},
	}

	visitor := &MockVisitor{}

	node.Accept(visitor)

	if len(node.Visitors) != 1 {
		t.Errorf("Expected 1, got %d", len(node.Visitors))
	}
}

func TestASTNode_Visit(t *testing.T) {
	node := ASTNode{
		children: &pb.Stmts{},
		Visitors: []Visitor{},
	}

	// append child node
	node.children.StmtFunction = append(node.children.StmtFunction, &pb.StmtFunction{})

	visitor := &MockVisitor{}

	node.Accept(visitor)
	node.Visit()

	if visitor.visited != true {
		t.Errorf("Expected true, got %v", visitor.visited)
	}
}
