package analyzer

import (
	"math"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type MaintainabilityIndexVisitor struct {
	complexity int
}

func (v *MaintainabilityIndexVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {
    // Compute MI for the current node when analyzable data is available
    if stmts == nil {
        return
    }
    v.Calculate(stmts)
}

func (v *MaintainabilityIndexVisitor) LeaveNode(stmts *pb.Stmts) {
    // Ensure MI is computed for the current node as well (root/file level included)
    if stmts == nil { return }
    v.Calculate(stmts)
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
	var halsteadVolume float64 = *stmts.Analyze.Volume.HalsteadVolume
	var MIwoC float64 = 0
	var MI float64 = 0
	var commentWeight float64 = 0

	// // maintainability index without comment
	MIwoC = float64(math.Max((171-
		(5.2*math.Log(float64(halsteadVolume)))-
		(0.23*float64(cyclomatic))-
		(16.2*math.Log(float64(lloc))))*100/171, 0))

	if math.IsInf(float64(MIwoC), 0) {
		// Avoid defaulting to 171 which makes tests fail; treat as 0 when undefined
		MIwoC = 0
	}

	if loc > 0 {
		CM := float64(cloc) / float64(loc)
		commentWeight = float64(50 * math.Sin(math.Sqrt(2.4*CM)))
	}

	MI = MIwoC + commentWeight

	// Case where no code is found
	if loc+lloc+cloc == 0 {
		MI = 0
		MIwoC = 0
		commentWeight = 0
	}

	MI32 := float64(MI)
	MIwoC32 := float64(MIwoC)
	commentWeight32 := float64(commentWeight)

	if stmts.Analyze.Maintainability == nil {
		stmts.Analyze.Maintainability = &pb.Maintainability{}
	}

	// Do not force a default of 171; keep computed values or zeros if missing metrics
	stmts.Analyze.Maintainability.MaintainabilityIndex = &MI32
	stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments = &MIwoC32
	stmts.Analyze.Maintainability.CommentWeight = &commentWeight32
}
