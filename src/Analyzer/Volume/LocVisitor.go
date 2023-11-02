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

        var loc int32 = 0 // lines of code
        var lloc int32 = 0 // logical lines of code
        var cloc int32 = 0 // comment lines of code

        // count the number of lines of code in the function
        loc = stmt.Location.EndLine - stmt.Location.StartLine + 1
        lloc = loc - stmt.Location.BlankLines

        // count the number of lines of code in the comments
        if stmt.Comments != nil {
            for _, comment := range stmt.Comments {
                cloc += comment.Location.EndLine - comment.Location.StartLine + 1
            }
        }

        if stmt.Stmts != nil && stmt.Stmts.Analyze == nil {
            stmt.Stmts.Analyze.Volume.Loc = &loc
            stmt.Stmts.Analyze.Volume.Lloc = &lloc
            stmt.Stmts.Analyze.Volume.Cloc = &cloc
        }
    }
}

func (v *LocVisitor) LeaveNode(stmts *pb.Stmts) {

    if stmts == nil {
        return
    }

    // aggregates loc for classes
    if len(stmts.StmtClass) > 0 {
        for _, stmt := range stmts.StmtClass {

            if stmt.Stmts == nil {
                continue
            }

            var loc int32 = stmt.Location.EndLine - stmt.Location.StartLine + 1 // lines of code
            var lloc int32 = 0 // logical lines of code
            var cloc int32 = 0 // comment lines of code

            for _, method := range stmt.Stmts.StmtFunction {
                if method.Stmts != nil {
                    lloc += *method.Stmts.Analyze.Volume.Lloc
                    cloc += *method.Stmts.Analyze.Volume.Cloc
                }
            }

            stmt.Stmts.Analyze.Volume.Loc = &loc
            stmt.Stmts.Analyze.Volume.Lloc = &lloc
            stmt.Stmts.Analyze.Volume.Cloc = &cloc
        }
    }

}

