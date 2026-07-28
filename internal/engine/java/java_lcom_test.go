package java

import (
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/analyzer"
	"github.com/ast-metrics/ast-metrics/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestJavaConstructorDoesNotArtificiallyCreateCohesion(t *testing.T) {
	src := `
public class Example {
    private int a;
    private int b;

    public Example() {
        this.a = 1;
        this.b = 2;
    }

    public int useA() {
        return this.a;
    }

    public int useB() {
        return this.b;
    }
}`
	r := &JavaRunner{}
	file, _ := engine.CreateTestFileWithCode(r, src)
	analyzer.AnalyzeFile(file)

	class := file.Stmts.StmtClass[0]
	assert.NotNil(t, class.Stmts.Analyze.ClassCohesion.Lcom4)
	assert.Equal(t, int32(2), *class.Stmts.Analyze.ClassCohesion.Lcom4)
}

func TestJavaClassWithoutMethodHasLCOM4OfZero(t *testing.T) {
	src := `
public class Example {
    private int a;
}`
	r := &JavaRunner{}
	file, _ := engine.CreateTestFileWithCode(r, src)
	analyzer.AnalyzeFile(file)

	class := file.Stmts.StmtClass[0]
	assert.NotNil(t, class.Stmts.Analyze.ClassCohesion.Lcom4)
	assert.Equal(t, int32(0), *class.Stmts.Analyze.ClassCohesion.Lcom4)
}
