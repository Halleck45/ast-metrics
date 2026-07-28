package golang

import (
	"slices"
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/analyzer"
	"github.com/ast-metrics/ast-metrics/internal/engine"
	pb "github.com/ast-metrics/ast-metrics/pb"
)

const sampleGoStruct = `package main

type Example struct {
	counter int
	label   string
}

func (e *Example) Increment() {
	e.counter = e.counter + 1
}

func (e *Example) Counter() int {
	return e.counter
}

func (e *Example) Rename(label string) {
	e.label = label
}

func (e *Example) Reset() {
	e.Increment()
}
`

func Test_Go_Struct_Methods_Are_Attached_To_The_Struct(t *testing.T) {
	file, err := engine.CreateTestFileWithCode(&GolangRunner{}, sampleGoStruct)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	analyzer.AnalyzeFile(file)

	if len(file.Stmts.StmtClass) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(file.Stmts.StmtClass))
	}
	class := file.Stmts.StmtClass[0]
	if len(class.Stmts.StmtFunction) != 4 {
		t.Fatalf("expected the 4 methods of Example to be attached to it, got %d", len(class.Stmts.StmtFunction))
	}
	// methods with a receiver are not functions of the file
	if len(file.Stmts.StmtFunction) != 0 {
		t.Fatalf("expected no top-level function, got %d", len(file.Stmts.StmtFunction))
	}

	var rename *pb.StmtFunction
	for _, method := range class.Stmts.StmtFunction {
		// the receiver qualifies the method: two structs of a file may declare
		// the same method name
		if method.Name.Qualified != "main\\Example."+method.Name.Short {
			t.Fatalf("expected the method to be qualified with its struct, got %q", method.Name.Qualified)
		}
		if method.Name.Short == "Rename" {
			rename = method
		}
	}
	if rename == nil {
		t.Fatalf("method Rename not found")
	}
	// the receiver is not a parameter
	if len(rename.Parameters) != 1 || rename.Parameters[0].Name != "label" {
		t.Fatalf("expected Rename to declare the single parameter label, got %+v", rename.Parameters)
	}
}

func Test_Go_Method_Declared_Before_Its_Struct_Is_Attached(t *testing.T) {
	src := `package main

func (e *Later) Value() int {
	return e.value
}

type Later struct {
	value int
}
`
	file, err := engine.CreateTestFileWithCode(&GolangRunner{}, src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(file.Stmts.StmtClass) != 1 || len(file.Stmts.StmtClass[0].Stmts.StmtFunction) != 1 {
		t.Fatalf("expected Value to be attached to Later, even though it is declared first")
	}
}

func Test_Go_Struct_Has_Lack_Of_Cohesion(t *testing.T) {
	file, err := engine.CreateTestFileWithCode(&GolangRunner{}, sampleGoStruct)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	analyzer.AnalyzeFile(file)

	class := file.Stmts.StmtClass[0]
	if class.Stmts.Analyze.ClassCohesion == nil || class.Stmts.Analyze.ClassCohesion.Lcom4 == nil {
		t.Fatalf("expected LCOM4 to be computed on the struct Example")
	}
	// Increment and Counter share the attribute counter, Reset calls Increment:
	// they form one component. Rename only touches label.
	expected := int32(2)
	if *class.Stmts.Analyze.ClassCohesion.Lcom4 != expected {
		t.Fatalf("expected LCOM4=%d, got %d", expected, *class.Stmts.Analyze.ClassCohesion.Lcom4)
	}
}

func Test_Go_Multiline_Signature_Does_Not_Leak_Types(t *testing.T) {
	src := []byte(`package main

type Wide struct {
	total int
}

func (w *Wide) Add(
	amount int,
	factor int,
) int {
	return w.total + amount*factor
}
`)
	adapter := NewTreeSitterAdapter(src)
	_, operands := adapter.ExtractOperatorsOperands(src, 7, 12)

	for _, expected := range []string{"Add", "amount", "factor", "this.total"} {
		if !slices.Contains(operands, expected) {
			t.Errorf("expected the operand %q, got %v", expected, operands)
		}
	}
	for _, leaked := range []string{"int", "Wide", "w", "w.total"} {
		if slices.Contains(operands, leaked) {
			t.Errorf("%q leaked into the operands: %v", leaked, operands)
		}
	}
}

func Test_Go_Receiver_Accesses_Are_Reported_As_Attributes(t *testing.T) {
	adapter := NewTreeSitterAdapter([]byte(sampleGoStruct))
	src := []byte(sampleGoStruct)

	// `func (e *Example) Increment() {` ... on lines 8 to 10
	_, operands := adapter.ExtractOperatorsOperands(src, 8, 10)
	if !slices.Contains(operands, "this.counter") {
		t.Fatalf("expected the receiver access to be normalized as this.counter, got %v", operands)
	}
	if slices.Contains(operands, "e.counter") || slices.Contains(operands, "e") || slices.Contains(operands, "Example") {
		t.Fatalf("expected the receiver and its type to be normalized, got %v", operands)
	}

	// `func (e *Example) Reset() { e.Increment() }` on lines 20 to 22
	calls := adapter.ExtractMethodCalls(src, 20, 22)
	if len(calls) != 1 || calls[0] != "this.Increment" {
		t.Fatalf("expected the call on the receiver to be reported as this.Increment, got %v", calls)
	}
	_, operands = adapter.ExtractOperatorsOperands(src, 20, 22)
	if slices.Contains(operands, "this.Increment") {
		t.Fatalf("a call on the receiver is not an operand, got %v", operands)
	}
}
