package analyzer

import (
	pb "github.com/halleck45/ast-metrics/pb"
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
	if stmts == nil {
		return
	}
	ccn := v.Calculate(stmts)
	if stmts.Analyze == nil {
		stmts.Analyze = &pb.Analyze{}
	}
	if stmts.Analyze.Complexity == nil {
		stmts.Analyze.Complexity = &pb.Complexity{}
	}
	stmts.Analyze.Complexity.Cyclomatic = &ccn
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
	// However, in some languages (e.g., Go via tree-sitter), an "else if" can be represented
	// as an else branch that contains an if-statement inside its body. In that case, count
	// one extra decision to align with expected CCN.
	for _, el := range stmts.StmtDecisionElse {
		if el != nil && el.Stmts != nil && len(el.Stmts.StmtDecisionIf) > 0 {
			ccn++
		}
	}
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
