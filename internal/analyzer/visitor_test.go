package analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// Mock visitor for testing
type mockVisitor struct {
	visitCalls     int
	leaveNodeCalls int
}

func (m *mockVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {
	m.visitCalls++
}

func (m *mockVisitor) LeaveNode(stmts *pb.Stmts) {
	m.leaveNodeCalls++
}

func TestHelperRecursionVisitor_Recurse_NilStmts(t *testing.T) {
	visitor := &mockVisitor{}
	recurser := &HelperRecursionVisitor{}

	recurser.Recurse(nil, visitor)

	if visitor.visitCalls != 0 {
		t.Errorf("expected 0 visit calls, got %d", visitor.visitCalls)
	}
}

func TestHelperRecursionVisitor_Recurse_EmptyStmts(t *testing.T) {
	visitor := &mockVisitor{}
	recurser := &HelperRecursionVisitor{}
	stmts := &pb.Stmts{}

	recurser.Recurse(stmts, visitor)

	if visitor.leaveNodeCalls != 1 {
		t.Errorf("expected 1 leave node call, got %d", visitor.leaveNodeCalls)
	}
	if stmts.Analyze == nil {
		t.Error("expected Analyze to be initialized")
	}
}

func TestHelperRecursionVisitor_Recurse_WithClass(t *testing.T) {
	visitor := &mockVisitor{}
	recurser := &HelperRecursionVisitor{}
	stmts := &pb.Stmts{
		StmtClass: []*pb.StmtClass{
			{Stmts: &pb.Stmts{}},
		},
	}

	recurser.Recurse(stmts, visitor)

	if visitor.visitCalls != 1 {
		t.Errorf("expected 1 visit call for class, got %d", visitor.visitCalls)
	}
	if visitor.leaveNodeCalls != 2 { // class + root
		t.Errorf("expected 2 leave node calls, got %d", visitor.leaveNodeCalls)
	}
}
