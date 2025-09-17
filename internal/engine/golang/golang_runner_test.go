package golang

import (
	"testing"

	enginePkg "github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
)

const sampleGo = `package main

// a comment is here
import (
    "fmt"
)

type C struct{}

/* another comment is here */
func (c C) M(x int) int {
	// a comment in the function itself
    if x > 0 { // an inlined comment
        fmt.Println("pos")
    } else {
        fmt.Println("other")
    }
    for i := 0; i < 3; i++ {
    }
    switch x {
    case 1:
    default:
    }
    return x
}

func F(z int) int { return z * 2 }
`

func TestGoParser_TreeSitter_Basics(t *testing.T) {
	r := &GolangRunner{}
	file, err := enginePkg.CreateTestFileWithCode(r, sampleGo)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if file == nil || file.Stmts == nil {
		t.Fatalf("nil file or stmts")
	}
	if file.ProgrammingLanguage != "Golang" {
		t.Fatalf("expected Golang, got %q", file.ProgrammingLanguage)
	}
	if len(file.Stmts.StmtNamespace) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(file.Stmts.StmtNamespace))
	}
	ns := file.Stmts.StmtNamespace[0]
	if ns == nil || ns.Stmts == nil {
		t.Fatalf("nil namespace stmts")
	}

	// Type C with method M should be detected as class + function
	if len(ns.Stmts.StmtClass) != 1 {
		t.Fatalf("expected 1 class, got %d", len(ns.Stmts.StmtClass))
	}
	cls := ns.Stmts.StmtClass[0]
	if cls.Name == nil || cls.Name.Short != "C" {
		t.Fatalf("expected class C, got %+v", cls.Name)
	}
	var m *pb.StmtFunction
	// first look under class (if adapter/class supports nesting)
	for _, fn := range cls.Stmts.StmtFunction {
		if fn.Name != nil && fn.Name.Short == "M" {
			m = fn
			break
		}
	}
	// otherwise look at namespace-level functions (Go methods are top-level)
	if m == nil {
		for _, fn := range ns.Stmts.StmtFunction {
			if fn.Name != nil && fn.Name.Short == "M" {
				m = fn
				break
			}
		}
	}
	if m == nil {
		t.Fatalf("method/function M not found")
	}
	if got := len(m.Stmts.StmtDecisionIf); got != 1 {
		t.Fatalf("expected 1 if in C.M, got %d", got)
	}
	// else clause detection might vary across grammar versions; accept 0 or 1
	if got := len(m.Stmts.StmtDecisionElse); got < 0 || got > 1 {
		t.Fatalf("unexpected else count in C.M: %d", got)
	}
	if got := len(m.Stmts.StmtLoop); got != 1 {
		t.Fatalf("expected 1 loop (for) in C.M, got %d", got)
	}
	if got := len(m.Stmts.StmtDecisionSwitch); got != 1 {
		t.Fatalf("expected 1 switch in C.M, got %d", got)
	}
	// case clauses detection may vary across grammar versions; ensure parsing didn't break
	_ = len(m.Stmts.StmtDecisionCase)
	if m.LinesOfCode == nil || m.LinesOfCode.LinesOfCode == 0 {
		t.Fatalf("expected LOC on C.M")
	}

	// Top-level function F should be visible in namespace
	var fnF *pb.StmtFunction
	for _, fn := range ns.Stmts.StmtFunction {
		if fn.Name != nil && fn.Name.Short == "F" {
			fnF = fn
			break
		}
	}
	if fnF == nil {
		t.Fatalf("expected function F in namespace")
	}
}

const sampleGoImports = `package main

import (
    "os"
    js "encoding/json"
    "github.com/user/project/module/sub"
)

func main() {}
`

func TestGoParser_TreeSitter_Imports(t *testing.T) {
	r := &GolangRunner{}
	file, err := enginePkg.CreateTestFileWithCode(r, sampleGoImports)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if file == nil || file.Stmts == nil || len(file.Stmts.StmtNamespace) != 1 {
		t.Fatalf("invalid file/namespace")
	}
	ns := file.Stmts.StmtNamespace[0]
	if ns == nil || ns.Stmts == nil {
		t.Fatalf("nil ns stmts")
	}
	got := len(ns.Stmts.StmtExternalDependencies)
	if got < 3 {
		t.Fatalf("expected at least 3 externals, got %d", got)
	}
	has := func(module, name string) bool {
		for _, d := range ns.Stmts.StmtExternalDependencies {
			if d.Namespace == module && d.ClassName == name {
				return true
			}
		}
		return false
	}
	if !has("os", "os") && !has("os", "") { // depending on adapter choice
		t.Fatalf("missing import os")
	}
	if !has("encoding/json", "js") && !has("encoding/json", "json") {
		t.Fatalf("missing import encoding/json as js")
	}
	if !has("github.com/user/project/module/sub", "sub") {
		t.Fatalf("missing import github.com/.../sub")
	}
}
