package Golang

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/halleck45/ast-metrics/src/Configuration"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Storage"
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

	parser := GolangRunner{}
	result := parser.ParseGoFile(tmpFile)

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

	parser := GolangRunner{}
	result := parser.ParseGoFile(tmpFile)
	testedFunction := result.Stmts.StmtFunction[0]

	// Ifs
	ifs := testedFunction.Stmts.StmtDecisionIf
	assert.Equal(t, 3, len(ifs), "Expected ifs")

	// Loops
	loops := testedFunction.Stmts.StmtLoop
	assert.Equal(t, 2, len(loops), "Expected loops")
}

func TestParsingGoFiles(t *testing.T) {
	goFile := `
	
	package mymath

	import "fmt"
	
	func foo(a int, b int) int {
		if a == 0 {
			
		}

		myrange := make([]int, 5)
		for i, v := range myrange {
			fmt.Println(i, v)
		}
	}`

	// Create a temporary file
	sourceDirectory := t.TempDir()
	tmpFile := sourceDirectory + string(os.PathSeparator) + "test.go"
	defer os.Remove(tmpFile)
	if _, err := os.Create(tmpFile); err != nil {
		t.Error(err)
	}
	if err := os.WriteFile(tmpFile, []byte(goFile), 0644); err != nil {
		t.Error(err)
	}

	// Configure destination
	storage := Storage.Default()
	storage.Ensure()
	workdir := storage.AstDirectory()

	// Configure the runner
	configuration := Configuration.NewConfiguration()
	configuration.Storage = storage
	configuration.SetSourcesToAnalyzePath([]string{sourceDirectory})
	runner := GolangRunner{}
	runner.SetConfiguration(configuration)
	runner.DumpAST()

	// list files
	files, err := os.ReadDir(workdir)
	if err != nil {
		t.Error(err)
	}
	// check if bin file exists
	assert.Equal(t, 1, len(files), "Expected 1 file in %s, got %d", workdir, len(files))

	// read the file, and deserialize it to check if it's a protobuf
	file := files[0]
	binPath := workdir + string(os.PathSeparator) + file.Name()
	in, err := os.ReadFile(binPath)
	if err != nil {
		t.Error(err)
	}
	pbFile := &pb.File{}
	if err := proto.Unmarshal(in, pbFile); err != nil {
		t.Error(err)
	}

	// Ensure path
	assert.Contains(t, pbFile.Path, "test.go", "Expected path to contain 'test.go', got %s", pbFile.Path)
}

func TestSearchModfile(t *testing.T) {
	runner := GolangRunner{}

	// Test when go.mod file exists in the given path
	t.Run("go.mod exists in path", func(t *testing.T) {
		tmpDir := t.TempDir()
		goModPath := tmpDir + "/go.mod"
		_, err := os.Create(goModPath)
		if err != nil {
			t.Error(err)
		}

		_, err = runner.SearchModfile(tmpDir)
		if err != nil {
			t.Error(err)
		}

		if runner.currentGoModPath != tmpDir {
			t.Errorf("Expected currentGoModPath to be %s, got %s", tmpDir, runner.currentGoModPath)
		}
	})

	// Test when go.mod file does not exist in the given path, but exists in the parent directory
	t.Run("go.mod exists in parent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		goModPath := tmpDir + "/go.mod"
		_, err := os.Create(goModPath)
		if err != nil {
			t.Error(err)
		}

		subDir := tmpDir + "/subdir"
		os.Mkdir(subDir, 0755)

		_, err = runner.SearchModfile(subDir)
		if err != nil {
			t.Error(err)
		}

		if runner.currentGoModPath != tmpDir {
			t.Errorf("Expected currentGoModPath to be %s, got %s", tmpDir, runner.currentGoModPath)
		}
	})

	// Test when go.mod file does not exist in the given path or any of its parent directories
	t.Run("go.mod does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := runner.SearchModfile(tmpDir)
		if err == nil {
			t.Error("Expected error, got nil")
		}

		expectedError := "go.mod file not found"
		if err.Error() != expectedError {
			t.Errorf("Expected error to be '%s', got '%s'", expectedError, err.Error())
		}
	})
}
