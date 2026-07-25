package golang

import (
	"testing"

	enginePkg "github.com/halleck45/ast-metrics/internal/engine"
)

func TestGoOperatorsAndOperandsExtraction(t *testing.T) {
	code := `package main

func calc(a, b int) int {
	c := a + b
	d := a - b
	e := a * b
	f := a / b
	g := a % b
	h := a << 2
	i := a >> 2
	j := a & b
	k := a | b
	l := a ^ b
	m := a &^ b
	n := a == b
	o := a != b
	p := a <= b
	q := a >= b
	r := a && (b > 0)
	s := a || b > 0
	a++
	b--
	z := foo.bar
	return c + d + e + f + g + h + i + j + k + l + m
}
`
	file, err := enginePkg.CreateTestFileWithCode(&GolangRunner{}, code)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(file.Stmts.StmtNamespace) != 1 {
		t.Fatalf("ns")
	}
	ns := file.Stmts.StmtNamespace[0]
	if len(ns.Stmts.StmtFunction) != 1 {
		t.Fatalf("expected 1 function")
	}
	fn := ns.Stmts.StmtFunction[0]
	if len(fn.Operators) == 0 {
		t.Fatalf("expected some operators, got 0")
	}
	if len(fn.Operands) == 0 {
		t.Fatalf("expected some operands, got 0")
	}

	// "foo.bar" is a single operand: its "." is not an operator
	if len(fn.Operators) != 49 {
		t.Fatalf("expected 49 operators, got %d", len(fn.Operators))
	}
	// the declaration line contributes "calc", "a" and "b": the types of the
	// parameters and of the result are not operands
	if len(fn.Operands) != 67 {
		t.Fatalf("expected 67 operands, got %d", len(fn.Operands))
	}
}
