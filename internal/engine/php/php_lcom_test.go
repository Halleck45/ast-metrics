package php

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
)

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
	r := &PhpRunner{}
	file, _ := engine.CreateTestFileWithCode(r, src)
	analyzer.AnalyzeFile(file)

	class1 := file.Stmts.StmtClass[0]
	expected := int32(2)
	if *class1.Stmts.Analyze.ClassCohesion.Lcom4 != expected {
		t.Errorf("Expected LCOM4=%d, got %d", expected, *class1.Stmts.Analyze.ClassCohesion.Lcom4)
	}
}
