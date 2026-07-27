package rust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/configuration"
	pb "github.com/ast-metrics/ast-metrics/pb"
)

func TestRustRunner_Name(t *testing.T) {
	runner := RustRunner{}
	if runner.Name() != "Rust" {
		t.Errorf("expected 'Rust', got %s", runner.Name())
	}
}

func TestRustRunner_IsRequired_NoFiles(t *testing.T) {
	runner := RustRunner{
		Configuration: &configuration.Configuration{},
	}

	if runner.IsRequired() {
		t.Error("expected IsRequired to be false when no Rust files found")
	}
}

func TestRustRunner_Parse_ValidRustFile(t *testing.T) {
	// Create temporary Rust file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.rs")

	rustCode := `fn main() {
    println!("Hello, world!");
}`

	err := os.WriteFile(tmpFile, []byte(rustCode), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := RustRunner{}
	file, err := runner.Parse(tmpFile)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if file.Path != tmpFile {
		t.Errorf("expected path %s, got %s", tmpFile, file.Path)
	}

	if file.ProgrammingLanguage != "Rust" {
		t.Errorf("expected language 'Rust', got %s", file.ProgrammingLanguage)
	}
}

func TestRustRunner_Parse_NonExistentFile(t *testing.T) {
	runner := RustRunner{}
	file, err := runner.Parse("/nonexistent/file.rs")

	if err == nil {
		t.Error("expected error for non-existent file")
	}

	if file.Path != "/nonexistent/file.rs" {
		t.Errorf("expected path to be preserved even on error")
	}

	if file.ProgrammingLanguage != "Rust" {
		t.Errorf("expected language 'Rust' even on error, got %s", file.ProgrammingLanguage)
	}
}

func TestRustRunner_Ensure(t *testing.T) {
	runner := RustRunner{}
	err := runner.Ensure()
	if err != nil {
		t.Errorf("expected no error from Ensure, got %v", err)
	}
}

func TestRustRunner_Finish(t *testing.T) {
	runner := RustRunner{}
	err := runner.Finish()
	if err != nil {
		t.Errorf("expected no error from Finish, got %v", err)
	}
}

func TestRustRunner_IsTest_ByFilename(t *testing.T) {
	rustCode := `fn add(a: i32, b: i32) -> i32 {
	a + b
}
`
	// Create a temporary file with _test.rs suffix
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "calculator_test.rs")

	err := os.WriteFile(tmpFile, []byte(rustCode), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := RustRunner{}
	file, err := runner.Parse(tmpFile)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !file.IsTest {
		t.Fatalf("expected file to be detected as test (_test.rs suffix)")
	}
}

func TestRustRunner_IsTest_ByAttribute(t *testing.T) {
	rustCode := `#[cfg(test)]
mod tests {
	#[test]
	fn test_add() {
		assert_eq!(3, 1 + 2);
	}
}
`
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "calculator.rs")

	err := os.WriteFile(tmpFile, []byte(rustCode), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := RustRunner{}
	file, err := runner.Parse(tmpFile)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !file.IsTest {
		t.Fatalf("expected file to be detected as test (contains #[cfg(test)])")
	}
}

func TestRustRunner_IsTest_NormalFile(t *testing.T) {
	rustCode := `fn add(a: i32, b: i32) -> i32 {
	a + b
}
`
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "calculator.rs")

	err := os.WriteFile(tmpFile, []byte(rustCode), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := RustRunner{}
	file, err := runner.Parse(tmpFile)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if file.IsTest {
		t.Fatalf("expected file NOT to be detected as test")
	}
}

// "#[...]" introduces an attribute in Rust, not a comment: attribute lines
// are not counted as comment lines, and being metadata rather than
// statements they are not logical lines either.
func TestRustAttributesAreNotComments(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "attrs.rs")

	rustCode := `fn compute(a: i32) -> i32 {
    #[cfg(debug_assertions)]
    let _dbg = true;
    #[allow(unused)]
    let _unused = 0;
    a + 1
}`

	if err := os.WriteFile(tmpFile, []byte(rustCode), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := RustRunner{}
	file, err := runner.Parse(tmpFile)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var fn *pb.StmtFunction
	all := file.Stmts.StmtFunction
	for _, ns := range file.Stmts.StmtNamespace {
		if ns != nil && ns.Stmts != nil {
			all = append(all, ns.Stmts.StmtFunction...)
		}
	}
	for _, f := range all {
		if f != nil && f.Name != nil && f.Name.Short == "compute" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatalf("function compute not found")
	}
	if fn.LinesOfCode == nil {
		t.Fatalf("expected LinesOfCode on compute")
	}
	if fn.LinesOfCode.CommentLinesOfCode != 0 {
		t.Errorf("expected 0 comment lines, got %d (attributes counted as comments)", fn.LinesOfCode.CommentLinesOfCode)
	}
	// 3 statement lines: the two "let" bindings and the "a + 1" tail
	// expression; attribute lines are neither comments nor statements
	if fn.LinesOfCode.LogicalLinesOfCode != 3 {
		t.Errorf("expected 3 logical lines, got %d", fn.LinesOfCode.LogicalLinesOfCode)
	}
}
