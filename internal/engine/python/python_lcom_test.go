package python

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestPythonInitDoesNotArtificiallyCreateCohesion(t *testing.T) {
	src := `
class Example:
    def __init__(self):
        self.a = 1
        self.b = 2

    def use_a(self):
        return self.a

    def use_b(self):
        return self.b
`
	r := &PythonRunner{}
	file, _ := engine.CreateTestFileWithCode(r, src)
	analyzer.AnalyzeFile(file)

	class := file.Stmts.StmtClass[0]
	assert.NotNil(t, class.Stmts.Analyze.ClassCohesion.Lcom4)
	assert.Equal(t, int32(2), *class.Stmts.Analyze.ClassCohesion.Lcom4)
}
