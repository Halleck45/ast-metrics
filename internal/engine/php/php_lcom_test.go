package php

import (
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/analyzer"
	"github.com/ast-metrics/ast-metrics/internal/engine"
)

func assertLcom4(t *testing.T, src string, expected int32) {
	t.Helper()

	r := &PhpRunner{}
	file, _ := engine.CreateTestFileWithCode(r, src)
	analyzer.AnalyzeFile(file)

	class1 := file.Stmts.StmtClass[0]
	if class1.Stmts.Analyze.ClassCohesion == nil || class1.Stmts.Analyze.ClassCohesion.Lcom4 == nil {
		t.Fatalf("Expected LCOM4=%d, got no measure at all", expected)
	}
	if *class1.Stmts.Analyze.ClassCohesion.Lcom4 != expected {
		t.Errorf("Expected LCOM4=%d, got %d", expected, *class1.Stmts.Analyze.ClassCohesion.Lcom4)
	}
}

func Test_Real_Code_Has_Lack_Of_Cohesion(t *testing.T) {

	src := `
<?php
class Example {
    private $a;

    public function m1() {
        $this->m2();
    }

    public function m2() {
        $this->a = 1;
    }

    public function m3() {
        $this->a = 1;
    }

    public function m4() {
        $this->m5();
    }

    public function m5() {
        echo 'ok';
    }
}`
	assertLcom4(t, src, 2)
}

func TestAClassWithoutMethodsHasLCOM4OfZero(t *testing.T) {

	src := `
<?php
class Example {
    private $a;
    private $b;
}`
	assertLcom4(t, src, 0)
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

    public function useA() {
        return $this->a;
    }

    public function useB() {
        return $this->b;
    }
}`
	assertLcom4(t, src, 2)
}

func TestDestructorDoesNotArtificiallyCreateCohesion(t *testing.T) {

	src := `
<?php
class Example {
    private $a;
    private $b;

    public function __destruct() {
        unset($this->a);
        unset($this->b);
    }

    public function useA() {
        return $this->a;
    }

    public function useB() {
        return $this->b;
    }
}`
	assertLcom4(t, src, 2)
}

func TestEmptyMethodsDoNotIncreaseLCOM4Components(t *testing.T) {

	src := `
<?php
class Example {
    private $a;

    public function useA() {
        return $this->a;
    }

    public function alsoUseA() {
        $this->a = 2;
    }

    public function notImplementedYet() {}
}`
	assertLcom4(t, src, 1)
}

func TestAClassWithOnlyLifecycleAndEmptyMethodsHasLCOM4OfZero(t *testing.T) {

	src := `
<?php
class Example {
    private $a;

    public function __construct() {
        $this->a = 1;
    }

    public function __destruct() {
        unset($this->a);
    }

    public function notImplementedYet() {}
}`
	assertLcom4(t, src, 0)
}

func TestCallingAnEmptyMethodDoesNotConnectComponents(t *testing.T) {

	src := `
<?php
class Example {
    private $a;
    private $b;

    public function useA() {
        $this->a = 1;
        $this->hook();
    }

    public function useB() {
        $this->b = 2;
        $this->hook();
    }

    public function hook() {}
}`
	assertLcom4(t, src, 2)
}

func TestPhp4StyleConstructorDoesNotArtificiallyCreateCohesion(t *testing.T) {

	src := `
<?php
class Example {
    private $a;
    private $b;

    public function Example() {
        $this->a = 1;
        $this->b = 2;
    }

    public function useA() {
        return $this->a;
    }

    public function useB() {
        return $this->b;
    }
}`
	assertLcom4(t, src, 2)
}
