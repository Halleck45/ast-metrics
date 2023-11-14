package Analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

func TestLocVisitorVisit(t *testing.T) {
	visitor := LocVisitor{}

	stmts := pb.Stmts{
		StmtFunction: []*pb.StmtFunction{
			{
				LinesOfCode: &pb.LinesOfCode{
					LinesOfCode:        10,
					LogicalLinesOfCode: 20,
					CommentLinesOfCode: 30,
				},
				Stmts: &pb.Stmts{
					Analyze: &pb.Analyze{
						Volume: &pb.Volume{},
					},
				},
			},
		},
		StmtClass: []*pb.StmtClass{
			{
				LinesOfCode: &pb.LinesOfCode{
					LinesOfCode:        40,
					LogicalLinesOfCode: 50,
					CommentLinesOfCode: 60,
				},
				Stmts: &pb.Stmts{},
			},
		},
	}

	visitor.Visit(&stmts, &stmts)

	if stmts.StmtFunction[0].Stmts.Analyze.Volume.GetLoc() != 10 {
		t.Errorf("Expected 10, got %d", stmts.StmtFunction[0].Stmts.Analyze.Volume.GetLoc())
	}

	if stmts.StmtFunction[0].Stmts.Analyze.Volume.GetLloc() != 20 {
		t.Errorf("Expected 20, got %d", stmts.StmtFunction[0].Stmts.Analyze.Volume.GetLloc())
	}

	if stmts.StmtFunction[0].Stmts.Analyze.Volume.GetCloc() != 30 {
		t.Errorf("Expected 30, got %d", stmts.StmtFunction[0].Stmts.Analyze.Volume.GetCloc())
	}

	// Assertions on class
	if stmts.StmtClass[0].LinesOfCode.GetLinesOfCode() != 40 {
		t.Errorf("Expected 40, got %d", stmts.StmtClass[0].LinesOfCode.GetLinesOfCode())
	}

	if stmts.StmtClass[0].LinesOfCode.GetLogicalLinesOfCode() != 50 {
		t.Errorf("Expected 50, got %d", stmts.StmtClass[0].LinesOfCode.GetLogicalLinesOfCode())
	}

	if stmts.StmtClass[0].LinesOfCode.GetCommentLinesOfCode() != 60 {
		t.Errorf("Expected 60, got %d", stmts.StmtClass[0].LinesOfCode.GetCommentLinesOfCode())
	}

}
