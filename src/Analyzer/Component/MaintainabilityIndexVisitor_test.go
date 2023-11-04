package Analyzer

import (
    "testing"
    pb "github.com/halleck45/ast-metrics/src/NodeType"
)

func TestItCalculateMaintainabilityIndex(t *testing.T) {

   visitor := MaintainabilityIndexVisitor{}

   stmts := pb.Stmts{}
   class1 := pb.StmtClass{}
   class1.Stmts = &pb.Stmts{}
   stmts.StmtClass = append(stmts.StmtClass, &class1)

   stmts.Analyze = &pb.Analyze{}
   stmts.Analyze.Volume = &pb.Volume{}

   loc := int32(10)
   lloc := int32(8)
   cloc := int32(2)
   cyclomatic := int32(3)
   halsteadVolume := float32(10)

   stmts.Analyze.Volume.Loc = &loc
   stmts.Analyze.Volume.Lloc = &lloc
   stmts.Analyze.Volume.Cloc = &cloc
   stmts.Analyze.Complexity = &pb.Complexity{}
   stmts.Analyze.Complexity.Cyclomatic = &cyclomatic
   stmts.Analyze.Volume.HalsteadVolume = &halsteadVolume

   visitor.Calculate(&stmts)

    MI := int(*stmts.Analyze.Maintainability.MaintainabilityIndex)
    MIwoc := int(*stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments)
    commentWeight := int(*stmts.Analyze.Maintainability.CommentWeight)

    if MI != 104 {
        t.Error("Expected 104, got ", MI)
    }

    if MIwoc != 72 {
        t.Error("Expected 72, got ", MIwoc)
    }

    if commentWeight != 31 {
        t.Error("Expected 31, got ", commentWeight)
    }
}