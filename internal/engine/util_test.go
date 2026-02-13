package engine

import (
	"os"
	"strings"
	"testing"

	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/stretchr/testify/assert"
)

func TestGetLocPositionFromSource_BasicCase(t *testing.T) {
	source := []string{
		"package main",
		"",
		"func main() {",
		"    // comment",
		"    fmt.Println(\"hello\")",
		"}",
	}

	loc := GetLocPositionFromSource(source, 1, 6)
	if loc.LinesOfCode != 6 {
		t.Errorf("expected 6 lines of code, got %d", loc.LinesOfCode)
	}
	if loc.CommentLinesOfCode != 1 {
		t.Errorf("expected 1 comment line, got %d", loc.CommentLinesOfCode)
	}
}

func TestGetLocPositionFromSource_InvalidBounds(t *testing.T) {
	source := []string{"line1", "line2"}

	loc := GetLocPositionFromSource(source, 0, 10)
	if loc.LinesOfCode != 2 {
		t.Errorf("expected 2 lines of code, got %d", loc.LinesOfCode)
	}
}

func TestStripQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello world"`, "             "},
		{`'test'`, "      "},
		{`no quotes`, "no quotes"},
		{`"mixed 'quotes'"`, "                "},
	}

	for _, test := range tests {
		result := stripQuotes(test.input)
		if result != test.expected {
			t.Errorf("stripQuotes(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestFactoryStmts(t *testing.T) {
	stmts := FactoryStmts()

	if stmts == nil {
		t.Error("expected non-nil stmts")
	}
	if stmts.StmtDecisionIf == nil {
		t.Error("expected non-nil StmtDecisionIf")
	}
	if stmts.StmtFunction == nil {
		t.Error("expected non-nil StmtFunction")
	}
}

func TestGetClassesInFile_EmptyFile(t *testing.T) {
	file := &pb.File{}
	classes := GetClassesInFile(file)

	if len(classes) != 0 {
		t.Errorf("expected 0 classes, got %d", len(classes))
	}
}

func TestGetClassesInFile_WithClasses(t *testing.T) {
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{Name: &pb.Name{Short: "TestClass"}},
			},
		},
	}

	classes := GetClassesInFile(file)
	if len(classes) != 1 {
		t.Errorf("expected 1 class, got %d", len(classes))
	}
}

func TestGetClassesInFile_DeduplicatesNamespaceAndFileEntries(t *testing.T) {
	class := &pb.StmtClass{Name: &pb.Name{Short: "Foo"}}
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtClass: []*pb.StmtClass{class},
					},
				},
			},
			StmtClass: []*pb.StmtClass{class},
		},
	}

	classes := GetClassesInFile(file)
	if len(classes) != 1 {
		t.Errorf("expected 1 deduplicated class, got %d", len(classes))
	}
}

func TestGetClassesInFile_DeduplicatesEquivalentClassesByQualifiedName(t *testing.T) {
	classInNamespace := &pb.StmtClass{Name: &pb.Name{Short: "Foo", Qualified: "Acme\\Foo"}}
	classInFile := &pb.StmtClass{Name: &pb.Name{Short: "Foo", Qualified: "Acme\\Foo"}}
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtClass: []*pb.StmtClass{classInNamespace},
					},
				},
			},
			StmtClass: []*pb.StmtClass{classInFile},
		},
	}

	classes := GetClassesInFile(file)
	if len(classes) != 1 {
		t.Errorf("expected 1 deduplicated class by qualified name, got %d", len(classes))
	}
}

func TestGetFunctionsInFile_EmptyFile(t *testing.T) {
	file := &pb.File{}
	functions := GetFunctionsInFile(file)

	if len(functions) != 0 {
		t.Errorf("expected 0 functions, got %d", len(functions))
	}
}

func TestGetFunctionsInFile_DeduplicatesNamespaceClassAndFileEntries(t *testing.T) {
	method := &pb.StmtFunction{Name: &pb.Name{Short: "method"}}
	topLevel := &pb.StmtFunction{Name: &pb.Name{Short: "top"}}
	class := &pb.StmtClass{
		Stmts: &pb.Stmts{
			StmtFunction: []*pb.StmtFunction{method},
		},
	}
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{method, topLevel},
						StmtClass:    []*pb.StmtClass{class},
					},
				},
			},
			StmtFunction: []*pb.StmtFunction{topLevel},
		},
	}

	functions := GetFunctionsInFile(file)
	if len(functions) != 2 {
		t.Errorf("expected 2 deduplicated functions, got %d", len(functions))
	}
}

func TestGetFunctionsInFile_DeduplicatesEquivalentFunctionsByQualifiedName(t *testing.T) {
	fnInNamespace := &pb.StmtFunction{Name: &pb.Name{Short: "f", Qualified: "Acme\\f"}}
	fnInFile := &pb.StmtFunction{Name: &pb.Name{Short: "f", Qualified: "Acme\\f"}}
	file := &pb.File{
		Stmts: &pb.Stmts{
			StmtNamespace: []*pb.StmtNamespace{
				{
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{fnInNamespace},
					},
				},
			},
			StmtFunction: []*pb.StmtFunction{fnInFile},
		},
	}

	functions := GetFunctionsInFile(file)
	if len(functions) != 1 {
		t.Errorf("expected 1 deduplicated function by qualified name, got %d", len(functions))
	}
}

func TestReduceDepthOfNamespace(t *testing.T) {
	tests := []struct {
		namespace string
		depth     int
		expected  string
	}{
		{"com.example.package", 2, "com.example"},
		{"simple", 1, "simple"},
		{"one.two.three.four", 3, "one.two.three"},
	}

	for _, test := range tests {
		result := ReduceDepthOfNamespace(test.namespace, test.depth)
		if result != test.expected {
			t.Errorf("ReduceDepthOfNamespace(%q, %d) = %q, expected %q",
				test.namespace, test.depth, result, test.expected)
		}
	}

	// Test github.com special case separately
	result := ReduceDepthOfNamespace("github.com/user/repo", 2)
	// github.com case adds 1 to depth, so depth=2 becomes depth=3
	// This should return the full namespace since depth >= parts length
	if !strings.Contains(result, "github.com") {
		t.Errorf("ReduceDepthOfNamespace with github.com should preserve github.com, got %q", result)
	}
}

func TestSearchFilesByExtension(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "search_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFile := tmpDir + "/test.go"
	err = os.WriteFile(testFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	files, err := SearchFilesByExtension([]string{tmpDir}, ".go")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestGetDependenciesInFile_NilFile(t *testing.T) {
	deps := GetDependenciesInFile(nil)
	if len(deps) != 0 {
		t.Errorf("expected 0 dependencies for nil file, got %d", len(deps))
	}
}

func TestNamespacesParsing(t *testing.T) {
	t.Run("Should correctly reduce depth of namespace", func(t *testing.T) {
		assert.Equal(t, "abcd/def", ReduceDepthOfNamespace("abcd/def/ghi/lkm", 2))
		assert.Equal(t, "abcd", ReduceDepthOfNamespace("abcd/def/ghi/lkm", 1))
		assert.Equal(t, "abcd.def.ghi", ReduceDepthOfNamespace("abcd.def.ghi.lkm", 3))
		assert.Equal(t, "abcd.def", ReduceDepthOfNamespace("abcd.def", 3))
		assert.Equal(t, "abcd", ReduceDepthOfNamespace("abcd/def.ghi.lkm", 1))
		assert.Equal(t, "abcd", ReduceDepthOfNamespace("abcd", 2))
	})

	t.Run("Should avoid github.com namespace", func(t *testing.T) {
		assert.Equal(t, "github.com/test/test", ReduceDepthOfNamespace("github.com/test/test/test/test", 2))
		assert.Equal(t, "github.com/test", ReduceDepthOfNamespace("github.com/test/test/test/test", 1))
		assert.Equal(t, "github.com/test/test/test", ReduceDepthOfNamespace("github.com/test.test.test.test", 3))
		assert.Equal(t, "github.com/test.test", ReduceDepthOfNamespace("github.com/test.test", 3))
		assert.Equal(t, "github.com/test", ReduceDepthOfNamespace("github.com/test/test.test.test", 1))
		assert.Equal(t, "github.com/test", ReduceDepthOfNamespace("github.com/test", 2))
	})

}
