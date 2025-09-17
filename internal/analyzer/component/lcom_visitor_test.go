package analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestLackOfCohesionOfMethodsVisitor_Visit_NilStmts(t *testing.T) {
	visitor := &LackOfCohesionOfMethodsVisitor{}
	visitor.Visit(nil, nil)
	// Should not panic
}

func TestLackOfCohesionOfMethodsVisitor_LeaveNode_NilStmts(t *testing.T) {
	visitor := &LackOfCohesionOfMethodsVisitor{}
	visitor.LeaveNode(nil)
	// Should not panic
}

func TestLackOfCohesionOfMethodsVisitor_Calculate_EmptyStmts(t *testing.T) {
	visitor := &LackOfCohesionOfMethodsVisitor{}
	stmts := &pb.Stmts{}
	
	// Should not panic with empty stmts
	visitor.Calculate(stmts)
	
	// No classes to analyze, so no cohesion data should be created
}

func TestLackOfCohesionOfMethodsVisitor_Calculate_NilStmts(t *testing.T) {
	visitor := &LackOfCohesionOfMethodsVisitor{}
	
	// Should not panic with nil stmts
	visitor.Calculate(nil)
}
