package Golang

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGolangRunner(t *testing.T) {

	goFile := `
	
	package mymath
	
	func addition(a int, b int) int {
		a = a + 1
		c := a + b
		return c
	}
	
	func division(a int, b int) int {
		// Avoid division by zero
		/* axesome comment */
		if b == 0 {
			return 0
		}

		return a / b
	}`

	// Create a temporary file
	tmpFile := t.TempDir() + "/test.go"
	if _, err := os.Create(tmpFile); err != nil {
		t.Error(err)
	}
	if err := os.WriteFile(tmpFile, []byte(goFile), 0644); err != nil {
		t.Error(err)
	}

	result := parseGoFile(tmpFile)

	// Ensure path
	assert.Equal(t, tmpFile, result.Path, "Expected path to be %s, got %s", tmpFile, result.Path)

	// Ensure functions
	assert.Equal(t, 2, len(result.Stmts.StmtFunction), "Expected 2 functions, got %d", len(result.Stmts.StmtFunction))
	assert.Equal(t, "addition", result.Stmts.StmtFunction[0].Name.Short, "Expected function name to be 'addition', got %s", result.Stmts.StmtFunction[0].Name)
	assert.Equal(t, "division", result.Stmts.StmtFunction[1].Name.Short, "Expected function name to be 'division', got %s", result.Stmts.StmtFunction[1].Name)

	// Ensure operands
	// [name:"a" name:"b" name:"a" name:"a" name:"c" name:"a" name:"b" name:"c"]
	// Convert to string (for easier comparison)
	operandsAsString := fmt.Sprintf("%v", result.Stmts.StmtFunction[0].Operands)
	operandsExpectedAsString := "[name:\"a\" name:\"b\" name:\"a\" name:\"a\" name:\"c\" name:\"a\" name:\"b\" name:\"c\"]"
	assert.Equal(t, operandsExpectedAsString, operandsAsString, "Expected operands to be %s, got %s", operandsExpectedAsString, operandsAsString)

	// Ensure operators
	//   addition
	operatorsAsString := fmt.Sprintf("%v", result.Stmts.StmtFunction[0].Operators)
	operatorsExpectedAsString := "[name:\"+\" name:\"+\"]"
	assert.Equal(t, operatorsExpectedAsString, operatorsAsString, "Expected operators to be %s, got %s", operatorsExpectedAsString, operatorsAsString)
	//   division
	operatorsAsString = fmt.Sprintf("%v", result.Stmts.StmtFunction[1].Operators)
	operatorsExpectedAsString = "[name:\"==\" name:\"/\"]"
	assert.Equal(t, operatorsExpectedAsString, operatorsAsString, "Expected operators to be %s, got %s", operatorsExpectedAsString, operatorsAsString)

	// Ensure lines of code
	assert.Equal(t, int32(5), result.Stmts.StmtFunction[0].LinesOfCode.LinesOfCode, "Expected lines of code")
	assert.Equal(t, int32(0), result.Stmts.StmtFunction[0].LinesOfCode.CommentLinesOfCode, "Expected comment lines of code")
	assert.Equal(t, int32(3), result.Stmts.StmtFunction[0].LinesOfCode.LogicalLinesOfCode, "Expected logical lines of code")

	assert.Equal(t, int32(9), result.Stmts.StmtFunction[1].LinesOfCode.LinesOfCode, "Expected lines of code to be 10, got %d", result.Stmts.StmtFunction[1].LinesOfCode.LinesOfCode)
	assert.Equal(t, int32(2), result.Stmts.StmtFunction[1].LinesOfCode.CommentLinesOfCode, "Expected comment lines")
	assert.Equal(t, int32(4), result.Stmts.StmtFunction[1].LinesOfCode.LogicalLinesOfCode, "Expected logical lines")
}

func TestGoLangStructureExtraction(t *testing.T) {

	goFile := `
	
	package mymath

	import "fmt"
	
	func foo(a int, b int) int {
		if a == 0 {
			for i := 0; i < 10; i++ {
				a = a + 1

				if a == 5 {
					a = a + 1
				} else if a == 6 {
					a = a + 2
				} else {
					a = a + 3
				}
			}
		}

		myrange := make([]int, 5)
		for i, v := range myrange {
			fmt.Println(i, v)
		}
	}`

	// Create a temporary file
	tmpFile := t.TempDir() + "/test.go"
	if _, err := os.Create(tmpFile); err != nil {
		t.Error(err)
	}
	if err := os.WriteFile(tmpFile, []byte(goFile), 0644); err != nil {
		t.Error(err)
	}

	result := parseGoFile(tmpFile)
	testedFunction := result.Stmts.StmtFunction[0]

	// Ifs
	ifs := testedFunction.Stmts.StmtDecisionIf
	assert.Equal(t, 3, len(ifs), "Expected ifs")

	// Loops
	loops := testedFunction.Stmts.StmtLoop
	assert.Equal(t, 2, len(loops), "Expected loops")
}
