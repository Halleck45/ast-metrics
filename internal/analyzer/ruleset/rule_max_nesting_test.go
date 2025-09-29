package ruleset

import (
	"testing"

	enginePkg "github.com/halleck45/ast-metrics/internal/engine"
	golangrunner "github.com/halleck45/ast-metrics/internal/engine/golang"
	pb "github.com/halleck45/ast-metrics/pb"
)

const nestedGo = `package main

func f(x int) int {
    if x > 0 {
        for i := 0; i < 3; i++ {
            if i > 1 {
                return i
            }
        }
    }
    return x
}
`

// Ensure the max depth calculator detects at least 3 nested levels on a simple Go snippet.
func TestRuleMaxNesting_DepthAtLeastThree(t *testing.T) {
	r := &golangrunner.GolangRunner{}
	file, err := enginePkg.CreateTestFileWithCode(r, nestedGo)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if file == nil || file.Stmts == nil {
		t.Fatalf("nil file or stmts")
	}
	// Try to locate nested structures in the parsed tree to understand structure
	var fn *pb.StmtFunction
	for _, ns := range file.Stmts.StmtNamespace {
		for _, f := range ns.Stmts.StmtFunction {
			if f.Name != nil && f.Name.Short == "f" {
				fn = f
				break
			}
		}
	}
	if fn == nil || fn.Stmts == nil {
		t.Fatalf("function f not found or has no stmts")
	}
	// Log some structure info to diagnose
	ifCount := len(fn.Stmts.StmtDecisionIf)
	loopCount := len(fn.Stmts.StmtLoop)
	switchCount := len(fn.Stmts.StmtDecisionSwitch)
	t.Logf("fn-level counts: if=%d, loop=%d, switch=%d", ifCount, loopCount, switchCount)
	if ifCount > 0 {
		inner := fn.Stmts.StmtDecisionIf[0].GetStmts()
		if inner != nil {
			t.Logf("inner if stmts: loops=%d ifs=%d", len(inner.StmtLoop), len(inner.StmtDecisionIf))
		}
	}
	depth := maxDepthStmts(file.Stmts, 0)
	if depth != 3 {
		t.Fatalf("expected nesting depth >= 3, got %d", depth)
	}
}
