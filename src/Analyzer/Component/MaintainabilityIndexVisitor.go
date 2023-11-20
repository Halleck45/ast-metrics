package Analyzer

import (
	"math"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type MaintainabilityIndexVisitor struct {
	complexity int
}

func (v *MaintainabilityIndexVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {

	if stmts == nil {
		return
	}

	for _, stmt := range parents.StmtClass {
		v.Calculate(stmt.Stmts)
	}

	for _, stmt := range parents.StmtFunction {
		v.Calculate(stmt.Stmts)
	}
}

func (v *MaintainabilityIndexVisitor) LeaveNode(stmts *pb.Stmts) {

}

/**
 * Calculates Maintainability Index
 *
 *      According to Wikipedia, "Maintainability Index is a software metric which measures how maintainable (easy to
 *      support and change) the source code is. The maintainability index is calculated as a factored formula consisting
 *      of Lines Of Code, Cyclomatic Complexity and Halstead volume."
 *
 *      MIwoc: Maintainability Index without comments
 *      MIcw: Maintainability Index comment weight
 *      MI: Maintainability Index = MIwoc + MIcw
 *
 *      MIwoc = 171 - 5.2 * ln(aveV) -0.23 * aveG -16.2 * ln(aveLOC)
 *      MIcw = 50 * sin(sqrt(2.4 * perCM))
 *      MI = MIwoc + MIcw
 *
 * @author Jean-François Lépine <https://twitter.com/Halleck45>
 */
func (v *MaintainabilityIndexVisitor) Calculate(stmts *pb.Stmts) {
	if stmts == nil {
		return
	}

	if stmts.Analyze == nil ||
		stmts.Analyze.Volume == nil ||
		stmts.Analyze.Volume.Loc == nil ||
		stmts.Analyze.Volume.Lloc == nil ||
		stmts.Analyze.Volume.Cloc == nil ||
		stmts.Analyze.Complexity.Cyclomatic == nil ||
		stmts.Analyze.Volume.HalsteadVolume == nil {
		return
	}

	var loc int32 = *stmts.Analyze.Volume.Loc
	var lloc int32 = *stmts.Analyze.Volume.Lloc
	var cloc int32 = *stmts.Analyze.Volume.Cloc
	var cyclomatic int32 = *stmts.Analyze.Complexity.Cyclomatic
	var halsteadVolume float32 = *stmts.Analyze.Volume.HalsteadVolume
	var MIwoC float64 = 0
	var MI float64 = 0
	var commentWeight float64 = 0

	// // maintainability index without comment
	MIwoC = max((171-
		(5.2*math.Log(float64(halsteadVolume)))-
		(0.23*float64(cyclomatic))-
		(16.2*math.Log(float64(lloc))))*100/171, 0)

	if math.IsInf(MIwoC, 0) {
		MIwoC = 171
	}

	if loc > 0 {
		CM := float64(cloc) / float64(loc)
		commentWeight = 50 * math.Sin(math.Sqrt(2.4*CM))
	}

	MI = MIwoC + commentWeight

	// Case where no code is found
	if loc+lloc+cloc == 0 {
		MI = 0
		MIwoC = 0
		commentWeight = 0
	}

	MI32 := float32(MI)
	MIwoC32 := float32(MIwoC)
	commentWeight32 := float32(commentWeight)

	if stmts.Analyze.Maintainability == nil {
		stmts.Analyze.Maintainability = &pb.Maintainability{}
	}

	stmts.Analyze.Maintainability.MaintainabilityIndex = &MI32
	stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments = &MIwoC32
	stmts.Analyze.Maintainability.CommentWeight = &commentWeight32
}
