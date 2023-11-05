package Analyzer

import (
    pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type LocVisitor struct {
    complexity int
}

func (v *LocVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {

    if stmts == nil {
        return
    }

    // calculate the number of lines of code in each method
    for _, stmt := range parents.StmtFunction {
        if stmt.Stmts == nil {
            continue
        }

        if stmt.LinesOfCode == nil {
            continue
        }

        if stmt.Stmts != nil && stmt.Stmts.Analyze != nil {
            stmt.Stmts.Analyze.Volume.Loc = &stmt.LinesOfCode.LinesOfCode
            stmt.Stmts.Analyze.Volume.Lloc = &stmt.LinesOfCode.LogicalLinesOfCode
            stmt.Stmts.Analyze.Volume.Cloc = &stmt.LinesOfCode.CommentLinesOfCode
        }
    }


    // Consolidate foreach class
    for _, class := range stmts.StmtClass {

        if class.Stmts == nil {
            continue
        }

        class.Stmts.Analyze.Volume = &pb.Volume{}
        class.Stmts.Analyze.Volume.Loc = &class.LinesOfCode.LinesOfCode

        var lloc int32
        var cloc int32
        for _, function := range class.Stmts.StmtFunction {
            lloc += function.LinesOfCode.LogicalLinesOfCode
            cloc += function.LinesOfCode.CommentLinesOfCode
        }

        class.Stmts.Analyze.Volume.Lloc = &lloc
        class.Stmts.Analyze.Volume.Cloc = &cloc
    }
}


func (v *LocVisitor) LeaveNode(stmts *pb.Stmts) {

}

