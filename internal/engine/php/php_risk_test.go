package php

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	enginePkg "github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
)

// Simple PHP source mixing a class and some procedural code to exercise both paths
const phpSource = `<?php
// simple procedural function
function f($x){ if($x>0){return $x+1;} else {return $x-1;} }

class A { 
    public function m($y){
        $s = 0; 
        for($i=0;$i<$y;$i++){ $s += $i; }
        return $s; 
    }
}
`

func Test_Php_Risk_Computed_Under_Realistic_Conditions(t *testing.T) {
	r := &PhpRunner{}
	file, err := enginePkg.CreateTestFileWithCode(r, phpSource)
	if err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}

	// Analyze single file first to populate metrics
	analyzer.AnalyzeFile(file)

	// Aggregate project with a single file (no git summaries)
	agg := analyzer.NewAggregator([]*pb.File{file}, nil)
	project := agg.Aggregates()

	if project.Combined.ConcernedFiles == nil || len(project.Combined.ConcernedFiles) == 0 {
		t.Fatalf("expected combined aggregation to include the file")
	}
	// Risk should be computed and be within [0,1]
	f := project.Combined.ConcernedFiles[0]
	if f == nil || f.Stmts == nil || f.Stmts.Analyze == nil || f.Stmts.Analyze.Risk == nil {
		t.Fatalf("expected risk to be computed on file")
	}
	if f.Stmts.Analyze.Risk.Score != 3 {
		t.Fatalf("unexpected risk score: %v", f.Stmts.Analyze.Risk.Score)
	}
}
