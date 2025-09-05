package golang

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	enginePkg "github.com/halleck45/ast-metrics/internal/engine"
	"github.com/stretchr/testify/assert"
)

func Test_Cyclomatic_Complexity_Is_Correct(t *testing.T) {
	src := `
package foo
	func M() {
	// a comment
}
	`
	r := &GolangRunner{}
	file, _ := enginePkg.CreateTestFileWithCode(r, src)
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
