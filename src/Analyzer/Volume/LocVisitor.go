package Analyzer

import (
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type LocVisitor struct {
	complexity int
}

func (v *LocVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {

	// calculate the number of lines of code in each method
	for _, stmt := range parents.StmtFunction {

		if stmt.LinesOfCode == nil {
			continue
		}

		if stmt.Stmts == nil {
			stmt.Stmts = &pb.Stmts{}
		}

		if stmt.Stmts.Analyze == nil {
			stmt.Stmts.Analyze = &pb.Analyze{}
		}

		if stmt.Stmts.Analyze.Volume == nil {
			stmt.Stmts.Analyze.Volume = &pb.Volume{}
		}

		stmt.Stmts.Analyze.Volume.Loc = &stmt.LinesOfCode.LinesOfCode
		stmt.Stmts.Analyze.Volume.Lloc = &stmt.LinesOfCode.LogicalLinesOfCode
		stmt.Stmts.Analyze.Volume.Cloc = &stmt.LinesOfCode.CommentLinesOfCode
	}

	if stmts == nil {
		return
	}

	// Consolidate foreach class
	for _, class := range stmts.StmtClass {

		if class.Stmts == nil {
			continue
		}

		v.consolidate(class.Stmts, class.LinesOfCode)
	}
}

func (v *LocVisitor) LeaveNode(stmts *pb.Stmts) {

}

func (v *LocVisitor) consolidate(stmts *pb.Stmts, loc *pb.LinesOfCode) {

	if stmts == nil {
		return
	}

	if loc == nil {
		return
	}

	if stmts.Analyze == nil {
		stmts.Analyze = &pb.Analyze{}
	}

	stmts.Analyze.Volume = &pb.Volume{}
	stmts.Analyze.Volume.Loc = &loc.LinesOfCode

	var lloc int32
	var cloc int32
	for _, function := range stmts.StmtFunction {
		lloc += function.LinesOfCode.LogicalLinesOfCode
		cloc += function.LinesOfCode.CommentLinesOfCode
	}

	stmts.Analyze.Volume.Lloc = &lloc
	stmts.Analyze.Volume.Cloc = &cloc
}
