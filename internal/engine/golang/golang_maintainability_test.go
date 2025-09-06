package golang

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	enginePkg "github.com/halleck45/ast-metrics/internal/engine"
	"github.com/stretchr/testify/assert"
)

const proceduralGo = `package main

func A(x int) int { if x > 0 { x = x + 1 } else { x = x - 1 }; return x }
func B(y int) int { for i:=0;i<y;i++ { y += i }; return y }
`

func Test_FileLevel_Maintainability_And_Halstead_IsComputed_ForProceduralGo(t *testing.T) {
	r := &GolangRunner{}
	file, _ := enginePkg.CreateTestFileWithCode(r, proceduralGo)

	analyzer.AnalyzeFile(file)

	if file.Stmts.Analyze == nil {
		t.Fatalf("missing Analyze on file")
	}
	if file.Stmts.Analyze.Volume == nil {
		t.Fatalf("missing Volume on file")
	}
	if file.Stmts.Analyze.Volume.Loc == nil || file.Stmts.Analyze.Volume.Lloc == nil || file.Stmts.Analyze.Volume.Cloc == nil {
		t.Fatalf("missing LOC/LLOC/CLOC on file: %+v", file.Stmts.Analyze.Volume)
	}
	if file.Stmts.Analyze.Complexity == nil || file.Stmts.Analyze.Complexity.Cyclomatic == nil {
		t.Fatalf("missing Complexity on file: %+v", file.Stmts.Analyze.Complexity)
	}
	if file.Stmts.Analyze.Volume.HalsteadVolume == nil {
		t.Fatalf("missing HalsteadVolume on file")
	}
	if file.Stmts.Analyze.Maintainability == nil || file.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil {
		t.Fatalf("expected file-level maintainability index to be computed")
	}

	mi := *file.Stmts.Analyze.Maintainability.MaintainabilityIndex
	if mi <= 0 {
		t.Fatalf("expected file-level MI > 0, got %v", mi)
	}
	if mi == 171 {
		t.Fatalf("file-level MI should not default to 171; got %v", mi)
	}
}

func Test_FileLevel_Maintainability_SampleGo(t *testing.T) {
	r := &GolangRunner{}
	file, _ := enginePkg.CreateTestFileWithCode(r, sampleGo)
	analyzer.AnalyzeFile(file)

	if file.Stmts.Analyze == nil {
		t.Fatalf("missing Analyze on sampleGo file")
	}
	if file.Stmts.Analyze.Volume == nil {
		t.Fatalf("missing Volume on sampleGo file")
	}
	if *file.Stmts.Analyze.Volume.HalsteadVocabulary != int32(15) {
		t.Fatalf("incorrect halstead volume on file")
	}
	if file.Stmts.Analyze.Volume.Loc == nil || file.Stmts.Analyze.Volume.Lloc == nil || file.Stmts.Analyze.Volume.Cloc == nil {
		t.Fatalf("missing LOC/LLOC/CLOC on sampleGo file")
	}
	if file.Stmts.Analyze.Complexity == nil || file.Stmts.Analyze.Complexity.Cyclomatic == nil {
		t.Fatalf("missing Complexity on sampleGo file")
	}
	if file.Stmts.Analyze.Volume.HalsteadVolume == nil {
		t.Fatalf("missing HalsteadVolume on sampleGo file")
	}
	if file.Stmts.Analyze.Maintainability == nil || file.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil {
		t.Fatalf("expected file-level maintainability index to be computed on sampleGo")
	}

	mi := *file.Stmts.Analyze.Maintainability.MaintainabilityIndex
	if mi <= 0 {
		t.Fatalf("expected MI > 0, got %v", mi)
	}
	if mi == 171 {
		t.Fatalf("MI should not be the constant 171 for non-empty file; got %v", mi)
	}
}

func Test_FileLevel_Loc_SampleGo(t *testing.T) {
	r := &GolangRunner{}
	file, _ := enginePkg.CreateTestFileWithCode(r, sampleGo)
	analyzer.AnalyzeFile(file)

	if file.Stmts.Analyze == nil {
		t.Fatalf("missing Analyze on sampleGo file")
	}
	if file.Stmts.Analyze.Volume == nil {
		t.Fatalf("missing Volume on sampleGo file")
	}
	if *file.Stmts.Analyze.Volume.Loc != int32(28) {
		t.Fatalf("incorrect Loc on sampleGo file, got %d", *file.Stmts.Analyze.Volume.Loc)
	}
	if *file.Stmts.Analyze.Volume.Lloc != int32(20) {
		t.Fatalf("incorrect logical Loc on sampleGo file, got %d", *file.Stmts.Analyze.Volume.Lloc)
	}
	if *file.Stmts.Analyze.Volume.Cloc != int32(3) {
		t.Fatalf("incorrect comment Loc on sampleGo file, got %d", *file.Stmts.Analyze.Volume.Cloc)
	}

	if len(file.Stmts.StmtFunction) != 2 {
		t.Fatal("functions not found in gofile")
	}
	function1 := file.Stmts.StmtFunction[0]
	assert.Equal(t, "M", function1.Name.Short)
	expected := int32(1)
	if *function1.Stmts.Analyze.Volume.Cloc != expected {
		t.Fatalf("incorrect comment lines of code for function, got %d", *function1.Stmts.Analyze.Volume.Cloc)
	}

}
