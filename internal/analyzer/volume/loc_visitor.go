package analyzer

import (
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type LocVisitor struct {
	complexity int
}

func (v *LocVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {

	// calculate the number of lines of code in each method
	if parents != nil {
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
	}

	// Consolidate foreach class
	if parents != nil {
		for _, class := range parents.StmtClass {

		if class.Stmts == nil {
			continue
		}

		v.consolidate(class.Stmts, class.LinesOfCode)
		}
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

	// Consolidate foreach method (if file is not a class)
	if len(stmts.StmtClass) == 0 {
		for _, function := range stmts.StmtFunction {
		
			if function.Stmts == nil {
				continue
			}
		
			target := parents
			if target == nil {
				target = stmts
			}
			v.consolidate(target, function.LinesOfCode)
		}
	}

	// Aggregate at file-level across namespaces (useful for languages like Go)
	if parents != nil && len(parents.StmtNamespace) > 0 {
		var sumLoc int32
		var sumLloc int32
		var sumCloc int32
		for _, ns := range parents.StmtNamespace {
			if ns == nil || ns.Stmts == nil {
				continue
			}
			for _, fn := range ns.Stmts.StmtFunction {
				if fn == nil || fn.LinesOfCode == nil { continue }
				sumLoc += fn.LinesOfCode.LinesOfCode
				sumLloc += fn.LinesOfCode.LogicalLinesOfCode
				sumCloc += fn.LinesOfCode.CommentLinesOfCode
			}
			for _, cls := range ns.Stmts.StmtClass {
				if cls == nil || cls.Stmts == nil {
					continue
				}
				for _, fn := range cls.Stmts.StmtFunction {
					if fn == nil || fn.LinesOfCode == nil { continue }
					sumLoc += fn.LinesOfCode.LinesOfCode
					sumLloc += fn.LinesOfCode.LogicalLinesOfCode
					sumCloc += fn.LinesOfCode.CommentLinesOfCode
				}
			}
		}
		if parents.Analyze == nil { parents.Analyze = &pb.Analyze{} }
		if parents.Analyze.Volume == nil { parents.Analyze.Volume = &pb.Volume{} }
		parents.Analyze.Volume.Loc = &sumLoc
		parents.Analyze.Volume.Lloc = &sumLloc
		parents.Analyze.Volume.Cloc = &sumCloc
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
	// Sum LOC across functions within this scope
	var sumLoc int32
	var lloc int32
	var cloc int32
	for _, function := range stmts.StmtFunction {
		sumLoc += function.LinesOfCode.LinesOfCode
		lloc += function.LinesOfCode.LogicalLinesOfCode
		cloc += function.LinesOfCode.CommentLinesOfCode
	}
	stmts.Analyze.Volume.Loc = &sumLoc
	stmts.Analyze.Volume.Lloc = &lloc
	stmts.Analyze.Volume.Cloc = &cloc
}
