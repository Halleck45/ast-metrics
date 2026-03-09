package php

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
)

func TestAClassWithoutMethodsHasLCOM4OfZero(t *testing.T) {
	src := `
<?php
class Example {
}`

	lcom4 := analyzeLCOM4ForFirstPHPClass(t, src)

	if lcom4 != 0 {
		t.Fatalf("expected LCOM4=0, got %d", lcom4)
	}
}

func TestConstructorDoesNotArtificiallyCreateCohesion(t *testing.T) {
	src := `
<?php
class Example {
    private $a;
    private $b;

    public function __construct() {
        $this->a = 1;
        $this->b = 2;
    }

    public function m1() {
        return $this->a;
    }

    public function m2() {
        return $this->b;
    }
}`

	lcom4 := analyzeLCOM4ForFirstPHPClass(t, src)

	if lcom4 < 2 {
		t.Fatalf("expected LCOM4>=2 when constructor is ignored, got %d", lcom4)
	}
}

func TestDestructorDoesNotArtificiallyCreateCohesion(t *testing.T) {
	src := `
<?php
class Example {
    private $a;
    private $b;

    public function __destruct() {
        $this->a = 1;
        $this->b = 2;
    }

    public function m1() {
        return $this->a;
    }

    public function m2() {
        return $this->b;
    }
}`

	lcom4 := analyzeLCOM4ForFirstPHPClass(t, src)

	if lcom4 < 2 {
		t.Fatalf("expected LCOM4>=2 when destructor is ignored, got %d", lcom4)
	}
}

func TestEmptyMethodsDoNotIncreaseLCOM4Components(t *testing.T) {
	src := `
<?php
class Example {
    private $a;

    public function m1() {
        $this->a = 1;
    }

    public function m2() {
        return $this->a;
    }

    public function emptyMethod() {
    }
}`

	lcom4 := analyzeLCOM4ForFirstPHPClass(t, src)

	if lcom4 != 1 {
		t.Fatalf("expected LCOM4=1 when empty methods are ignored, got %d", lcom4)
	}
}

func TestAClassWithOneCohesiveResponsibilityHasLCOM4OfOne(t *testing.T) {
	src := `
<?php
class Example {
    private $a;

    public function m1() {
        $this->a = 1;
    }

    public function m2() {
        return $this->a;
    }
}`

	lcom4 := analyzeLCOM4ForFirstPHPClass(t, src)

	if lcom4 != 1 {
		t.Fatalf("expected LCOM4=1, got %d", lcom4)
	}
}

func TestAClassWithDisconnectedResponsibilitiesHasLCOM4OfAtLeastTwo(t *testing.T) {
	src := `
<?php
class Example {
    private $a;
    private $b;

    public function m1() {
        $this->a = 1;
    }

    public function m2() {
        return $this->a;
    }

    public function m3() {
        $this->b = 2;
    }

    public function m4() {
        return $this->b;
    }
}`

	lcom4 := analyzeLCOM4ForFirstPHPClass(t, src)

	if lcom4 < 2 {
		t.Fatalf("expected LCOM4>=2, got %d", lcom4)
	}
}

func analyzeLCOM4ForFirstPHPClass(t *testing.T, src string) int32 {
	t.Helper()

	file, err := engine.CreateTestFileWithCode(&PhpRunner{}, src)
	if err != nil {
		t.Fatalf("unexpected error while parsing PHP source: %v", err)
	}

	analyzer.AnalyzeFile(file)

	if file == nil || file.Stmts == nil || len(file.Stmts.StmtClass) == 0 {
		t.Fatalf("expected one parsed class")
	}

	classNode := file.Stmts.StmtClass[0]
	if classNode == nil || classNode.Stmts == nil || classNode.Stmts.Analyze == nil || classNode.Stmts.Analyze.ClassCohesion == nil || classNode.Stmts.Analyze.ClassCohesion.Lcom4 == nil {
		t.Fatalf("expected LCOM4 metric to be set")
	}

	return *classNode.Stmts.Analyze.ClassCohesion.Lcom4
}
