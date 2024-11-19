package Analyzer

import (
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type CyclomaticComplexityVisitor struct {
	complexity int
}

func (v *CyclomaticComplexityVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {
	if stmts == nil {
		return
	}

	var ccn int32 = v.Calculate(stmts)
	if stmts.Analyze == nil {
		stmts.Analyze = &pb.Analyze{}
	}
	if stmts.Analyze.Complexity == nil {
		stmts.Analyze.Complexity = &pb.Complexity{}
	}

	stmts.Analyze.Complexity.Cyclomatic = &ccn
}

func (v *CyclomaticComplexityVisitor) LeaveNode(stmts *pb.Stmts) {
}

/**
 * Calculates cyclomatic complexity
 */
func (v *CyclomaticComplexityVisitor) Calculate(stmts *pb.Stmts) int32 {

	if stmts == nil {
		return 0
	}

	var ccn int32 = 0

	// count decision points
	ccn = int32(len(stmts.StmtLoop) +
		len(stmts.StmtDecisionIf) +
		len(stmts.StmtDecisionElseIf) +
		len(stmts.StmtDecisionCase) +
		len(stmts.StmtDecisionSwitch) +
		len(stmts.StmtFunction)) // +1 for the function itself
	// else is not a decision point for ccn
	// class is not a decision point for ccn

	// iterate over children
	for _, stmt := range stmts.StmtFunction {
		ccn += v.Calculate(stmt.Stmts)
	}

	for _, stmt := range stmts.StmtLoop {
		ccn += v.Calculate(stmt.Stmts)
	}

	for _, stmt := range stmts.StmtDecisionIf {
		ccn += v.Calculate(stmt.Stmts)
	}

	for _, stmt := range stmts.StmtDecisionElseIf {
		ccn += v.Calculate(stmt.Stmts)
	}

	for _, stmt := range stmts.StmtDecisionElse {
		ccn += v.Calculate(stmt.Stmts)
	}

	for _, stmt := range stmts.StmtDecisionCase {
		ccn += v.Calculate(stmt.Stmts)
	}

	for _, stmt := range stmts.StmtClass {
		ccn += v.Calculate(stmt.Stmts)
	}

	for _, stmt := range stmts.StmtDecisionSwitch {
		ccn += v.Calculate(stmt.Stmts)
	}

	return ccn
}
