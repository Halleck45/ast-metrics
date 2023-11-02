package Analyzer

import (
    pb "github.com/halleck45/ast-metrics/src/NodeType"
    "math"
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

    var n int32 // program vocabulary (η)
    var n1 int32 // number of unique operators
    var n2 int32 // number of unique operands
    var N int32 // program length (N)
    var N1 int32
    var N2 int32
    var hatN float64 // estimated program length (𝑁̂)
    var V float64 // volume (V)
    var D float64 // difficulty (D)
    var E float64 // effort (E)
    var T float64 // time required to program (T)

    // initialize default values
    hatN = 0
    V = 0
    D = 0
    E = 0
    T = 0


    for _, stmt := range parents.StmtFunction {
        if stmt.Stmts == nil {
            continue
        }

        // get unique operators and operands

        if stmts == nil {
            return
        }

        for _,operator := range stmt.Operators {
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
        hatN = float64(n1)*math.Log2(float64(n1)) + float64(n2)*math.Log2(float64(n2))

        // Calculate volume (V)
        V = float64(N) * math.Log2(float64(n))

        // Calculate difficulty (D)
        D = float64(n1) / 2 * float64(N2) / float64(n2)

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

            var n int32
            var N int32
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
                    n += int32(*method.Stmts.Analyze.Volume.HalsteadVocabulary)
                    N += int32(*method.Stmts.Analyze.Volume.HalsteadLength)
                    hatN += float64(*method.Stmts.Analyze.Volume.HalsteadEstimatedLength)
                    V += float64(*method.Stmts.Analyze.Volume.HalsteadVolume)
                    D += float64(*method.Stmts.Analyze.Volume.HalsteadDifficulty)
                    E += float64(*method.Stmts.Analyze.Volume.HalsteadEffort)
                    T += float64(*method.Stmts.Analyze.Volume.HalsteadTime)
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

            // convert float to float32
            V32 := float32(V)
            hatN32 := float32(hatN)
            D32 := float32(D)
            E32 := float32(E)
            T32 := float32(T)

            // Assign to result
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