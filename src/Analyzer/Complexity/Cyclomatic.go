package Analyzer

import (
    pb "github.com/halleck45/ast-metrics/src/NodeType"
    "fmt"
)

type ComplexityVisitor struct {
    complexity int
}

func (v *ComplexityVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {

    if stmts == nil {
        return
    }

    // count decision points
    var ccn int32
    ccn  = int32(len(stmts.StmtLoop) + len(stmts.StmtDecisionIf) + len(stmts.StmtDecisionElseIf) + len(stmts.StmtDecisionElse) + len(stmts.StmtDecisionCase))
    currentCyclomatic := *stmts.Analyze.Complexity.Cyclomatic
    currentCyclomatic += ccn
    *stmts.Analyze.Complexity.Cyclomatic = currentCyclomatic

    // increase parents complexity
    if parents != nil {
        currentCyclomatic = *parents.Analyze.Complexity.Cyclomatic
        currentCyclomatic += ccn
        *parents.Analyze.Complexity.Cyclomatic = currentCyclomatic
    }
}

func (v *ComplexityVisitor) LeaveNode(stmts *pb.Stmts) {

    if stmts == nil {
        return
    }

    // aggregates complexity for classes
    if len(stmts.StmtClass) > 0 {
        for _, stmt := range stmts.StmtClass {
            fmt.Println("LeaveNode:" + stmt.Name.Qualified)

            var ccn int32 = 0
            for _, method := range stmt.Stmts.StmtFunction {
                if method.Stmts != nil {
                    ccn += *method.Stmts.Analyze.Complexity.Cyclomatic
                }
            }
            stmt.Stmts.Analyze.Complexity.Cyclomatic = &ccn

            fmt.Println("ccn:" + fmt.Sprint(ccn))
        }
    }
}


func (v *ComplexityVisitor) GetComplexity() int {
	return v.complexity
}