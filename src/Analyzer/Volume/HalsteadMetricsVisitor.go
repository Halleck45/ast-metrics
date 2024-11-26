package Analyzer

import (
	"math"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
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
	var hatN float32 = 0 // estimated program length (𝑁̂)
	var V float32 = 0    // volume (V)
	var D float32 = 0    // difficulty (D)
	var E float32 = 0    // effort (E)
	var T float32 = 0    // time required to program (T)

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
		hatN = float32(n1)*float32(math.Log2(float64(n1))) + float32(n2)*float32(math.Log2(float64(n2)))
		if math.IsNaN(float64(hatN)) {
			hatN = 0
		}

		// Calculate volume (V)
		V = float32(N) * float32(math.Log2(float64(n)))
		if math.IsNaN(float64(V)) {
			V = 0
		}

		// Calculate difficulty (D)
		D = float32(n1)/2*float32(N2)/float32(n2)
		if math.IsNaN(float64(D)) {
			D = 0
		}

		// Calculate effort (E)
		E = D * V

		// Calculate time required to program (T)
		T = E / 18

		// convert float to float32
		V32 := float32(V)
		hatN32 := float32(hatN)
		D32 := float32(D)
		E32 := float32(E)
		T32 := float32(T)

		// Assign to result
		if stmt.Stmts.Analyze == nil {
			stmt.Stmts.Analyze = &pb.Analyze{}
			stmt.Stmts.Analyze.Volume = &pb.Volume{}
		}

		stmt.Stmts.Analyze.Volume.HalsteadVocabulary = &n
		stmt.Stmts.Analyze.Volume.HalsteadLength = &N
		stmt.Stmts.Analyze.Volume.HalsteadEstimatedLength = &hatN32
		stmt.Stmts.Analyze.Volume.HalsteadVolume = &V32
		stmt.Stmts.Analyze.Volume.HalsteadDifficulty = &D32
		stmt.Stmts.Analyze.Volume.HalsteadEffort = &E32
		stmt.Stmts.Analyze.Volume.HalsteadTime = &T32
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
			var hatN float32
			var V float32
			var D float32
			var E float32
			var T float32

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
				hatN = hatN / float32(len(stmt.Stmts.StmtFunction))
				V = V / float32(len(stmt.Stmts.StmtFunction))
				D = D / float32(len(stmt.Stmts.StmtFunction))
				E = E / float32(len(stmt.Stmts.StmtFunction))
				T = T / float32(len(stmt.Stmts.StmtFunction))
			}

			// convert float to float32
			V32 := float32(V)
			hatN32 := float32(hatN)
			D32 := float32(D)
			E32 := float32(E)
			T32 := float32(T)

			// Assign to result
			if stmt.Stmts.Analyze == nil {
				stmt.Stmts.Analyze = &pb.Analyze{}
			}
			if stmt.Stmts.Analyze.Volume == nil {
				stmt.Stmts.Analyze.Volume = &pb.Volume{}
			}

			stmt.Stmts.Analyze.Volume.HalsteadVocabulary = &n
			stmt.Stmts.Analyze.Volume.HalsteadLength = &N
			stmt.Stmts.Analyze.Volume.HalsteadEstimatedLength = &hatN32
			stmt.Stmts.Analyze.Volume.HalsteadVolume = &V32
			stmt.Stmts.Analyze.Volume.HalsteadDifficulty = &D32
			stmt.Stmts.Analyze.Volume.HalsteadEffort = &E32
			stmt.Stmts.Analyze.Volume.HalsteadTime = &T32
		}
	}
}
