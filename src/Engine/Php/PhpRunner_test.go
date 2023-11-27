package Php

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhpRunner(t *testing.T) {
	pythonSource := `
<?php
namespace Foo\Bar;

class calculatrice {
	// A single line comment is here
	// A single line comment is here

	public function add($a, $b) {
		// A single line comment is here
		// A single line comment is here
		// A single line comment is here
		// A single line comment is here
		return $a + $b;
	}


	public function divide(int $a, int $b) {
		if ($b == 0) {
			throw new \InvalidArgumentException('Division by zero.');
		}



		$d = $a / $b;
		$d += 1;
		$e = $this->add($this->a1, $d);
		return $e;
	}
}
`

	// Create a temporary file
	tmpFile := t.TempDir() + "/test.php"
	if _, err := os.Create(tmpFile); err != nil {
		t.Error(err)
	}
	if err := os.WriteFile(tmpFile, []byte(pythonSource), 0644); err != nil {
		t.Error(err)
	}

	result, err := parsePhpFile(tmpFile)

	// Ensure no error
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure path
	assert.Equal(t, tmpFile, result.Path, "Expected path to be %s, got %s", tmpFile, result.Path)

	// Ensure functions
	assert.Equal(t, 0, len(result.Stmts.StmtFunction), "Incorrect number of functions")

	// Ensure classes
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "calculatrice", class1.Name.Short, "Expected class name to be 'calculatrice', got %s", class1.Name)

	// Ensure functions
	assert.Equal(t, 2, len(class1.Stmts.StmtFunction), "Incorrect number of functions in class")

	func1 := class1.Stmts.StmtFunction[0]
	assert.Equal(t, "add", func1.Name.Short, "Expected function name to be 'add', got %s", func1.Name)
	assert.Equal(t, "Foo\\Bar\\calculatrice::add", func1.Name.Qualified, "Expected function name")
	func2 := class1.Stmts.StmtFunction[1]
	assert.Equal(t, "divide", func2.Name.Short, "Expected function name to be 'divide', got %s", func2.Name)
	assert.Equal(t, "Foo\\Bar\\calculatrice::divide", func2.Name.Qualified, "Expected function name")

	// Ensure operands
	// [name:"a" name:"b" name:"a" name:"b"]
	// Convert to string (for easier comparison)
	operandsAsString := fmt.Sprintf("%v", func1.Operands)
	operandsExpectedAsString := "[name:\"$a\" name:\"$b\" name:\"$a\" name:\"$b\"]"
	assert.Equal(t, operandsExpectedAsString, operandsAsString, "Expected operands to be %s, got %s", operandsExpectedAsString, operandsAsString)

	// Ensure operands of function 2
	// [a, b, b, d, a, b, d, e, a, d, e]
	// Convert to string (for easier comparison)
	operandsAsString = fmt.Sprintf("%v", func2.Operands)
	operandsExpectedAsString = "[name:\"$a\" name:\"$b\" name:\"$b\" name:\"$d\" name:\"$a\" name:\"$b\" name:\"$d\" name:\"$e\" name:\"$this->a1\" name:\"$d\" name:\"$e\"]"
	assert.Equal(t, operandsExpectedAsString, operandsAsString, "Expected operands to be %s, got %s", operandsExpectedAsString, operandsAsString)

	// Ensure operators
	// [+]
	// Convert to string (for easier comparison)
	operatorsAsString := fmt.Sprintf("%v", func1.Operators)
	operatorsExpectedAsString := "[name:\"+\"]"
	assert.Equal(t, operatorsExpectedAsString, operatorsAsString, "Expected operators to be %s, got %s", operatorsExpectedAsString, operatorsAsString)

	// Ensure operators of function 2
	// [==, / ]
	// Convert to string (for easier comparison)
	operatorsAsString = fmt.Sprintf("%v", func2.Operators)
	operatorsExpectedAsString = "[name:\"==\" name:\"/\" name:\"+=\"]"
	assert.Equal(t, operatorsExpectedAsString, operatorsAsString, "Expected operators to be %s, got %s", operatorsExpectedAsString, operatorsAsString)

	// Ensure LOC
	assert.Equal(t, int32(7), func1.LinesOfCode.LinesOfCode, "Expected LOC")
	assert.Equal(t, int32(1), func1.LinesOfCode.LogicalLinesOfCode, "Expected LLOC")
	assert.Equal(t, int32(4), func1.LinesOfCode.CommentLinesOfCode, "Expected LLOC")
	// Ensure LOC
	assert.Equal(t, int32(12), func2.LinesOfCode.LinesOfCode, "Expected LOC")
	assert.Equal(t, int32(7), func2.LinesOfCode.LogicalLinesOfCode, "Expected LLOC")
	assert.Equal(t, int32(0), func2.LinesOfCode.CommentLinesOfCode, "Expected LLOC")

}
