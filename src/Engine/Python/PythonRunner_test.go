package Python

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPythonRunner(t *testing.T) {

	// randomly selected, and found at https://github.com/jaraco/path/blob/main/path/classes.py
	pythonSource := `
import functools

class calculatrice:
	"""
	A multiline line comment is here
	A multiline line comment is here
	"""

	def add(self, a, b):
		"""
		A multiline line comment is here
		A multiline line comment is here
		"""
		return a + b


	def divide(self, a, b):
		if b == 0:
			raise ValueError("Cannot divide by zero")

		d = a / b
		d += 1
		e = self.add(a, d)
		return e
`

	// Create a temporary file
	tmpFile := t.TempDir() + "/test.py"
	if _, err := os.Create(tmpFile); err != nil {
		t.Error(err)
	}
	if err := os.WriteFile(tmpFile, []byte(pythonSource), 0644); err != nil {
		t.Error(err)
	}

	result, err := parsePythonFile(tmpFile)

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
	assert.Equal(t, "calculatrice.add", func1.Name.Qualified, "Expected function name to be 'calculatrice.add', got %s", func1.Name.Qualified)
	func2 := class1.Stmts.StmtFunction[1]
	assert.Equal(t, "divide", func2.Name.Short, "Expected function name to be 'divide', got %s", func2.Name)
	assert.Equal(t, "calculatrice.divide", func2.Name.Qualified, "Expected function name to be 'calculatrice.divide', got %s", func2.Name.Qualified)

	// Ensure operands
	// [name:"aself" name:"a" name:"b" name:"a" name:"b"]
	// Convert to string (for easier comparison)
	operandsAsString := fmt.Sprintf("%v", func1.Operands)
	operandsExpectedAsString := "[name:\"self\" name:\"a\" name:\"b\" name:\"a\" name:\"b\"]"
	assert.Equal(t, operandsExpectedAsString, operandsAsString, "Expected operands to be %s, got %s", operandsExpectedAsString, operandsAsString)

	// Ensure operands of function 2
	// [self, a, b, b, d, a, b, d, e, a, d, e]
	// Convert to string (for easier comparison)
	operandsAsString = fmt.Sprintf("%v", func2.Operands)
	operandsExpectedAsString = "[name:\"self\" name:\"a\" name:\"b\" name:\"b\" name:\"d\" name:\"a\" name:\"b\" name:\"d\" name:\"e\" name:\"a\" name:\"d\" name:\"e\"]"
	assert.Equal(t, operandsExpectedAsString, operandsAsString, "Expected operands to be %s, got %s", operandsExpectedAsString, operandsAsString)

	// Ensure operators

	// Ensure LOC
	assert.Equal(t, int32(6), func1.LinesOfCode.LinesOfCode, "Expected LOC")
	assert.Equal(t, int32(2), func1.LinesOfCode.LogicalLinesOfCode, "Expected LLOC")
	assert.Equal(t, int32(2), func1.LinesOfCode.CommentLinesOfCode, "Expected LLOC")
	// Ensure LOC
	assert.Equal(t, int32(8), func2.LinesOfCode.LinesOfCode, "Expected LOC")
	assert.Equal(t, int32(5), func2.LinesOfCode.LogicalLinesOfCode, "Expected LLOC")
	assert.Equal(t, int32(0), func2.LinesOfCode.CommentLinesOfCode, "Expected LLOC")

}
