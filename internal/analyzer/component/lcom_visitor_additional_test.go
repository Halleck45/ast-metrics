package analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestLackOfCohesionOfMethodsVisitor_Calculate_WithClasses(t *testing.T) {
	visitor := &LackOfCohesionOfMethodsVisitor{}
	
	stmts := &pb.Stmts{
		StmtClass: []*pb.StmtClass{
			{
				Name: &pb.Name{Short: "TestClass"},
				Stmts: &pb.Stmts{
					StmtFunction: []*pb.StmtFunction{
						{Name: &pb.Name{Short: "method1"}},
						{Name: &pb.Name{Short: "method2"}},
					},
					Analyze: &pb.Analyze{},
				},
			},
		},
		Analyze: &pb.Analyze{},
	}
	
	// Test should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Calculate panicked: %v", r)
		}
	}()
	
	visitor.Calculate(stmts)
}

func TestLackOfCohesionOfMethodsVisitor_Visit_WithParents(t *testing.T) {
	visitor := &LackOfCohesionOfMethodsVisitor{}
	
	stmts := &pb.Stmts{}
	parents := &pb.Stmts{
		StmtClass: []*pb.StmtClass{
			{Stmts: stmts},
		},
	}
	
	visitor.Visit(stmts, parents)
	// Should not panic
}

func TestLackOfCohesionOfMethodsVisitor_LeaveNode_WithData(t *testing.T) {
	visitor := &LackOfCohesionOfMethodsVisitor{}
	
	stmts := &pb.Stmts{
		StmtFunction: []*pb.StmtFunction{
			{Name: &pb.Name{Short: "testFunction"}},
		},
	}
	
	visitor.LeaveNode(stmts)
	// Should not panic
}
