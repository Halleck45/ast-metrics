package analyzer

import (
	"math"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type HalsteadMetricsVisitor struct {
	operatorCount int
	operandCount  int
}

func (v *HalsteadMetricsVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {

	if stmts == nil {
		return
	}

	// calculate the number of lines of code in each method
	var operatorSet map[string]bool
	var operandSet map[string]bool
	operatorSet = make(map[string]bool)
	operandSet = make(map[string]bool)
	var name string

	var n int32  // program vocabulary (η)
	var n1 int32 // number of unique operators
	var n2 int32 // number of unique operands
	var N int32  // program length (N)
	var N1 int32
	var N2 int32
	var hatN float64 = 0 // estimated program length (𝑁̂)
	var V float64 = 0    // volume (V)
	var D float64 = 0    // difficulty (D)
	var E float64 = 0    // effort (E)
	var T float64 = 0    // time required to program (T)

	for _, stmt := range parents.StmtFunction {
		if stmt.Stmts == nil {
			continue
		}

		// get unique operators and operands

		if stmts == nil {
			return
		}

		for _, operator := range stmt.Operators {
			name = operator.Name
			if _, ok := operatorSet[name]; !ok {
				operatorSet[name] = true
			}
		}

		for _, operand := range stmt.Operands {
			name = operand.Name
			if _, ok := operandSet[name]; !ok {
				operandSet[name] = true
			}
		}

		// Calculate Halstead metrics
		n1 = int32(len(operatorSet))
		n2 = int32(len(operandSet))
		N1 = int32(len(stmt.Operators))
		N2 = int32(len(stmt.Operands))

		// Calculate program vocabulary (η)
		n = int32(n1 + n2)

		// Calculate program length (N)
		N = int32(N1 + N2)

		// Calculate estimated program length (𝑁̂)
		hatN = float64(n1)*float64(math.Log2(float64(n1))) + float64(n2)*float64(math.Log2(float64(n2)))
		if math.IsNaN(float64(hatN)) {
			hatN = 0
		}

		// Calculate volume (V)
		V = float64(N) * float64(math.Log2(float64(n)))
		if math.IsNaN(float64(V)) {
			V = 0
		}

		// Calculate difficulty (D)
		D = float64(n1) / 2 * float64(N2) / float64(n2)
		if math.IsNaN(float64(D)) {
			D = 0
		}

		// Calculate effort (E)
		E = D * V

		// Calculate time required to program (T)
		T = E / 18

		// Assign to result
		if stmt.Stmts.Analyze == nil {
			stmt.Stmts.Analyze = &pb.Analyze{}
			stmt.Stmts.Analyze.Volume = &pb.Volume{}
		}

		stmt.Stmts.Analyze.Volume.HalsteadVocabulary = &n
		stmt.Stmts.Analyze.Volume.HalsteadLength = &N
		stmt.Stmts.Analyze.Volume.HalsteadEstimatedLength = &hatN
		stmt.Stmts.Analyze.Volume.HalsteadVolume = &V
		stmt.Stmts.Analyze.Volume.HalsteadDifficulty = &D
		stmt.Stmts.Analyze.Volume.HalsteadEffort = &E
		stmt.Stmts.Analyze.Volume.HalsteadTime = &T
	}

	// When there are no classes, aggregate Halstead to the parent (file) level, similar to LOC consolidation
	if len(stmts.StmtClass) == 0 && parents != nil {
		var nn int32 = 0
		var NN int32 = 0
		var hhatN float64
		var VV float64
		var DD float64
		var EE float64
		var TT float64
		count := 0

		for _, fn := range parents.StmtFunction {
			if fn.Stmts != nil && fn.Stmts.Analyze != nil && fn.Stmts.Analyze.Volume != nil && fn.Stmts.Analyze.Volume.HalsteadVocabulary != nil {
				nn += int32(*fn.Stmts.Analyze.Volume.HalsteadVocabulary)
				NN += int32(*fn.Stmts.Analyze.Volume.HalsteadLength)
				hhatN += *fn.Stmts.Analyze.Volume.HalsteadEstimatedLength
				VV += *fn.Stmts.Analyze.Volume.HalsteadVolume
				DD += *fn.Stmts.Analyze.Volume.HalsteadDifficulty
				EE += *fn.Stmts.Analyze.Volume.HalsteadEffort
				TT += *fn.Stmts.Analyze.Volume.HalsteadTime
				count++
			}
		}

		if count > 0 {
			nn = nn / int32(count)
			NN = NN / int32(count)
			hhatN = hhatN / float64(count)
			VV = VV / float64(count)
			DD = DD / float64(count)
			EE = EE / float64(count)
			TT = TT / float64(count)
			if parents.Analyze == nil {
				parents.Analyze = &pb.Analyze{}
			}
			if parents.Analyze.Volume == nil {
				parents.Analyze.Volume = &pb.Volume{}
			}
			parents.Analyze.Volume.HalsteadVocabulary = &nn
			parents.Analyze.Volume.HalsteadLength = &NN
			parents.Analyze.Volume.HalsteadEstimatedLength = &hhatN
			parents.Analyze.Volume.HalsteadVolume = &VV
			parents.Analyze.Volume.HalsteadDifficulty = &DD
			parents.Analyze.Volume.HalsteadEffort = &EE
			parents.Analyze.Volume.HalsteadTime = &TT
		}
	}
}

func (v *HalsteadMetricsVisitor) LeaveNode(stmts *pb.Stmts) {
	if stmts == nil {
		return
	}

	// aggregates for classes: we use the average of the methods
	if len(stmts.StmtClass) > 0 {
		for _, stmt := range stmts.StmtClass {

			if stmt.Stmts == nil {
				continue
			}

			var n int32 = 0
			var N int32 = 0
			var hatN float64
			var V float64
			var D float64
			var E float64
			var T float64

			// initialize default values
			hatN = 0
			V = 0
			D = 0
			E = 0
			T = 0

			for _, method := range stmt.Stmts.StmtFunction {
				if method.Stmts != nil {
					if method.Stmts.Analyze.Volume == nil || method.Stmts.Analyze.Volume.HalsteadVocabulary == nil {
						continue
					}
					n += int32(*method.Stmts.Analyze.Volume.HalsteadVocabulary)
					N += int32(*method.Stmts.Analyze.Volume.HalsteadLength)
					hatN += *method.Stmts.Analyze.Volume.HalsteadEstimatedLength
					V += *method.Stmts.Analyze.Volume.HalsteadVolume
					D += *method.Stmts.Analyze.Volume.HalsteadDifficulty
					E += *method.Stmts.Analyze.Volume.HalsteadEffort
					T += *method.Stmts.Analyze.Volume.HalsteadTime
				}
			}

			// calculate the average
			if len(stmt.Stmts.StmtFunction) > 0 {
				n = n / int32(len(stmt.Stmts.StmtFunction))
				N = N / int32(len(stmt.Stmts.StmtFunction))
				hatN = hatN / float64(len(stmt.Stmts.StmtFunction))
				V = V / float64(len(stmt.Stmts.StmtFunction))
				D = D / float64(len(stmt.Stmts.StmtFunction))
				E = E / float64(len(stmt.Stmts.StmtFunction))
				T = T / float64(len(stmt.Stmts.StmtFunction))
			}

			// Assign to result
			if stmt.Stmts.Analyze == nil {
				stmt.Stmts.Analyze = &pb.Analyze{}
			}
			if stmt.Stmts.Analyze.Volume == nil {
				stmt.Stmts.Analyze.Volume = &pb.Volume{}
			}

			stmt.Stmts.Analyze.Volume.HalsteadVocabulary = &n
			stmt.Stmts.Analyze.Volume.HalsteadLength = &N
			stmt.Stmts.Analyze.Volume.HalsteadEstimatedLength = &hatN
			stmt.Stmts.Analyze.Volume.HalsteadVolume = &V
			stmt.Stmts.Analyze.Volume.HalsteadDifficulty = &D
			stmt.Stmts.Analyze.Volume.HalsteadEffort = &E
			stmt.Stmts.Analyze.Volume.HalsteadTime = &T
		}
	} else {
		// No classes: aggregate Halstead at the current (file/namespace) level using its functions
		var n int32 = 0
		var N int32 = 0
		var hatN float64
		var V float64
		var D float64
		var E float64
		var T float64
		// initialize
		hatN = 0
		V = 0
		D = 0
		E = 0
		T = 0

		cnt := 0
		for _, fn := range stmts.StmtFunction {
			if fn.Stmts != nil && fn.Stmts.Analyze != nil && fn.Stmts.Analyze.Volume != nil && fn.Stmts.Analyze.Volume.HalsteadVocabulary != nil {
				n += int32(*fn.Stmts.Analyze.Volume.HalsteadVocabulary)
				N += int32(*fn.Stmts.Analyze.Volume.HalsteadLength)
				hatN += *fn.Stmts.Analyze.Volume.HalsteadEstimatedLength
				V += *fn.Stmts.Analyze.Volume.HalsteadVolume
				D += *fn.Stmts.Analyze.Volume.HalsteadDifficulty
				E += *fn.Stmts.Analyze.Volume.HalsteadEffort
				T += *fn.Stmts.Analyze.Volume.HalsteadTime
				cnt++
			}
		}

		if cnt > 0 {
			// average
			n = n / int32(cnt)
			N = N / int32(cnt)
			hatN = hatN / float64(cnt)
			V = V / float64(cnt)
			D = D / float64(cnt)
			E = E / float64(cnt)
			T = T / float64(cnt)

			if stmts.Analyze == nil {
				stmts.Analyze = &pb.Analyze{}
			}
			if stmts.Analyze.Volume == nil {
				stmts.Analyze.Volume = &pb.Volume{}
			}
			stmts.Analyze.Volume.HalsteadVocabulary = &n
			stmts.Analyze.Volume.HalsteadLength = &N
			stmts.Analyze.Volume.HalsteadEstimatedLength = &hatN
			stmts.Analyze.Volume.HalsteadVolume = &V
			stmts.Analyze.Volume.HalsteadDifficulty = &D
			stmts.Analyze.Volume.HalsteadEffort = &E
			stmts.Analyze.Volume.HalsteadTime = &T
		}
	}
}
