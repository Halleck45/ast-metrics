package rust

import (
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/analyzer"
	"github.com/ast-metrics/ast-metrics/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestRustNewDoesNotArtificiallyCreateCohesion(t *testing.T) {
	src := `
struct Example { a: i32, b: i32 }

impl Example {
    fn new() -> Example {
        Example { a: 1, b: 2 }
    }

    fn use_a(&self) -> i32 {
        self.a
    }

    fn use_b(&self) -> i32 {
        self.b
    }
}
`
	r := &RustRunner{}
	file, _ := engine.CreateTestFileWithCode(r, src)
	analyzer.AnalyzeFile(file)

	// the impl block carries the methods
	var measured *int32
	for _, class := range file.Stmts.StmtClass {
		if len(class.Stmts.StmtFunction) == 0 {
			continue
		}
		measured = class.Stmts.Analyze.ClassCohesion.Lcom4
	}

	assert.NotNil(t, measured)
	assert.Equal(t, int32(2), *measured)
}
