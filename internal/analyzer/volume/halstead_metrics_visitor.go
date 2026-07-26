package analyzer

import (
	"math"

	pb "github.com/halleck45/ast-metrics/pb"
)

type HalsteadMetricsVisitor struct {
	operatorCount int
	operandCount  int
}

func (v *HalsteadMetricsVisitor) Visit(stmts *pb.Stmts, parents *pb.Stmts) {

	if stmts == nil {
		return
	}

	for _, stmt := range parents.StmtFunction {
		if stmt.Stmts == nil {
			continue
		}

		// Everything below is scoped to this method. Sharing the counters (or
		// the pointers taken on them) across iterations would give every method
		// of a class the figures of the last one visited.
		operatorSet := make(map[string]bool)
		operandSet := make(map[string]bool)

		for _, operator := range stmt.Operators {
			operatorSet[operator.Name] = true
		}
		for _, operand := range stmt.Operands {
			operandSet[operand.Name] = true
		}

		n1 := int32(len(operatorSet)) // unique operators
		n2 := int32(len(operandSet))  // unique operands
		N1 := int32(len(stmt.Operators))
		N2 := int32(len(stmt.Operands))

		n := n1 + n2 // program vocabulary (η)
		N := N1 + N2 // program length (N)

		// estimated program length (𝑁̂)
		hatN := float64(n1)*math.Log2(float64(n1)) + float64(n2)*math.Log2(float64(n2))
		if math.IsNaN(hatN) || math.IsInf(hatN, 0) {
			hatN = 0
		}

		// volume (V)
		V := float64(N) * math.Log2(float64(n))
		if math.IsNaN(V) || math.IsInf(V, 0) {
			V = 0
		}

		// difficulty (D)
		D := float64(n1) / 2 * float64(N2) / float64(n2)
		if math.IsNaN(D) || math.IsInf(D, 0) {
			D = 0
		}

		E := D * V  // effort (E)
		T := E / 18 // time required to program (T)

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

		// Halstead estimated bugs (B) ≈ V / 3000
		if V > 0 {
			b := V / 3000.0
			stmt.Stmts.Analyze.Volume.HalsteadBugs = &b
		}
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
			nn = int32(math.Round(float64(nn) / float64(count)))
			NN = int32(math.Round(float64(NN) / float64(count)))
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
			// Halstead estimated bugs (B) at parent level
			if VV > 0 {
				bb := VV / 3000.0
				parents.Analyze.Volume.HalsteadBugs = &bb
			}
		}
	}
}

// aggregateScope fills the scope's Halstead metrics with the average of its
// classes and top-level functions.
func (v *HalsteadMetricsVisitor) aggregateScope(stmts *pb.Stmts) {
	var n int32
	var N int32
	var hatN, V, D, E, T float64
	cnt := 0

	accumulate := func(vol *pb.Volume) {
		if vol == nil || vol.HalsteadVocabulary == nil {
			return
		}
		// an empty scope (e.g. a class without methods) has nothing to measure
		// and must not drag the average down
		if *vol.HalsteadVocabulary == 0 && *vol.HalsteadLength == 0 {
			return
		}
		n += *vol.HalsteadVocabulary
		N += *vol.HalsteadLength
		hatN += *vol.HalsteadEstimatedLength
		V += *vol.HalsteadVolume
		D += *vol.HalsteadDifficulty
		E += *vol.HalsteadEffort
		T += *vol.HalsteadTime
		cnt++
	}

	for _, cls := range stmts.StmtClass {
		if cls.Stmts != nil && cls.Stmts.Analyze != nil {
			accumulate(cls.Stmts.Analyze.Volume)
		}
	}
	for _, fn := range stmts.StmtFunction {
		if fn.Stmts != nil && fn.Stmts.Analyze != nil {
			accumulate(fn.Stmts.Analyze.Volume)
		}
	}

	if cnt == 0 {
		return
	}
	n = int32(math.Round(float64(n) / float64(cnt)))
	N = int32(math.Round(float64(N) / float64(cnt)))
	hatN /= float64(cnt)
	V /= float64(cnt)
	D /= float64(cnt)
	E /= float64(cnt)
	T /= float64(cnt)

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
	if V > 0 {
		b := V / 3000.0
		stmts.Analyze.Volume.HalsteadBugs = &b
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

			// calculate the average. Vocabulary and length are rounded rather
			// than truncated: an integer division would report zero for a class
			// whose methods each hold only a symbol or two.
			if len(stmt.Stmts.StmtFunction) > 0 {
				count := float64(len(stmt.Stmts.StmtFunction))
				n = int32(math.Round(float64(n) / count))
				N = int32(math.Round(float64(N) / count))
				hatN = hatN / count
				V = V / count
				D = D / count
				E = E / count
				T = T / count
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
			// Halstead estimated bugs for class average
			if V > 0 {
				b := V / 3000.0
				stmt.Stmts.Analyze.Volume.HalsteadBugs = &b
			}
		}

		// Aggregate at the current scope (file or namespace) as the average of
		// its classes and top-level functions, so files made only of classes
		// still expose Halstead metrics (and thus a maintainability index).
		v.aggregateScope(stmts)
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
			n = int32(math.Round(float64(n) / float64(cnt)))
			N = int32(math.Round(float64(N) / float64(cnt)))
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
			// Halstead estimated bugs at this level
			if V > 0 {
				b := V / 3000.0
				stmts.Analyze.Volume.HalsteadBugs = &b
			}
		}
	}
}
