package python

import (
	"testing"

	enginePkg "github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

const samplePy = `# -*- coding: utf-8 -*-
import sys

class C:
    def m(self, x):
        if x > 0:
            print(f"pos {x}")
        elif x == 0:
            print("zero")
        else:
            print("neg")
        for i in range(3):
            pass
        while x < 2:
            x += 1
        return x

async def a(y):
    match y:
        case 1:
            return "one"
        case _:
            return "other"

def f(z):
    return z * 2
`

func TestPythonParser_TreeSitter_Decisions(t *testing.T) {
	r := &PythonRunner{}
	file, err := enginePkg.CreateTestFileWithCode(r, samplePy)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if file == nil || file.Stmts == nil {
		t.Fatalf("nil file or stmts")
	}
	if file.ProgrammingLanguage != "Python" {
		t.Fatalf("expected Python, got %q", file.ProgrammingLanguage)
	}
	if len(file.Stmts.StmtNamespace) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(file.Stmts.StmtNamespace))
	}
	ns := file.Stmts.StmtNamespace[0]
	if ns == nil || ns.Stmts == nil {
		t.Fatalf("nil namespace stmts")
	}

	// Classe C et méthode m
	if len(ns.Stmts.StmtClass) != 1 {
		t.Fatalf("expected 1 class, got %d", len(ns.Stmts.StmtClass))
	}
	cls := ns.Stmts.StmtClass[0]
	if cls.Name == nil || cls.Name.Short != "C" {
		t.Fatalf("expected class C, got %+v", cls.Name)
	}
	var m *pb.StmtFunction
	for _, fn := range cls.Stmts.StmtFunction {
		if fn.Name != nil && fn.Name.Short == "m" {
			m = fn
			break
		}
	}
	if m == nil {
		t.Fatalf("method m not found")
	}
	if got := len(m.Stmts.StmtDecisionIf); got != 1 {
		t.Fatalf("expected 1 if in C.m, got %d", got)
	}
	if got := len(m.Stmts.StmtDecisionElseIf); got != 1 {
		t.Fatalf("expected 1 elif in C.m, got %d", got)
	}
	if got := len(m.Stmts.StmtDecisionElse); got != 1 {
		t.Fatalf("expected 1 else in C.m, got %d", got)
	}
	if got := len(m.Stmts.StmtLoop); got != 2 {
		t.Fatalf("expected 2 loops (for, while) in C.m, got %d", got)
	}
	if m.LinesOfCode == nil || m.LinesOfCode.LinesOfCode == 0 {
		t.Fatalf("expected LOC on C.m")
	}

	// Fonctions top-level visibles dans le namespace
	if len(ns.Stmts.StmtFunction) < 3 {
		t.Fatalf("expected >=3 functions in namespace, got %d", len(ns.Stmts.StmtFunction))
	}

	// async def a : match/case → switch/case
	var fnA, fnF *pb.StmtFunction
	for _, fn := range ns.Stmts.StmtFunction {
		if fn.Name != nil && fn.Name.Short == "a" {
			fnA = fn
		}
		if fn.Name != nil && fn.Name.Short == "f" {
			fnF = fn
		}
	}
	if fnA == nil || fnF == nil {
		t.Fatalf("expected functions a and f")
	}
	if got := len(fnA.Stmts.StmtDecisionSwitch); got != 1 {
		t.Fatalf("expected 1 switch (match) in a, got %d", got)
	}
	if got := len(fnA.Stmts.StmtDecisionCase); got != 2 {
		t.Fatalf("expected 2 case in a, got %d", got)
	}
	if fnA.LinesOfCode == nil || fnA.LinesOfCode.LinesOfCode == 0 {
		t.Fatalf("expected LOC on a")
	}
	if fnF.LinesOfCode == nil || fnF.LinesOfCode.LinesOfCode == 0 {
		t.Fatalf("expected LOC on f")
	}
}

const sampleImports = `
import os
import json as js, pkg.sub.module
from typing import List, Dict as D
from .pkg import util, helpers as h
`

func TestPythonParser_TreeSitter_Imports(t *testing.T) {
	r := &PythonRunner{}
	file, err := enginePkg.CreateTestFileWithCode(r, sampleImports)
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

	// Expect at least these dependencies:
	wantMin := 7

	// You attached deps both to current scope and to namespace view.
	got := len(ns.Stmts.StmtExternalDependencies)
	if got < wantMin {
		t.Fatalf("expected at least %d external deps, got %d", wantMin, got)
	}

	// Spot-check a few entries
	has := func(module, name string) bool {
		for _, d := range ns.Stmts.StmtExternalDependencies {
			if d.Namespace == module && d.ClassName == name {
				return true
			}
			// plain import: we store module in Namespace and empty name
			if name == "" && d.Namespace == module && d.ClassName == "" {
				return true
			}
		}

		return false
	}

	if !has("os", "") {
		t.Fatalf("missing dep: import os")
	}
	if !has("json", "") && !has("json", "json") {
		// depending on your adapter choice for plain import
		t.Fatalf("missing dep: import json")
	}
	if !has("pkg.sub.module", "") {
		t.Fatalf("missing dep: import pkg.sub.module")
	}
	if !has("typing", "List") {
		t.Fatalf("missing dep: from typing import List")
	}
	if !has("typing", "Dict") {
		t.Fatalf("missing dep: from typing import Dict")
	}
	if !has("pkg", "util") {
		t.Fatalf("missing dep: from .pkg import util")
	}
	if !has("pkg", "helpers") {
		t.Fatalf("missing dep: from .pkg import helpers")
	}
}
