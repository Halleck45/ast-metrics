package Php

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

func TestPhpVisitorNameObject(t *testing.T) {
	// Initialize PhpVisitor
	v := &PhpVisitor{
		currentNamespace: &pb.StmtNamespace{
			Name: &pb.Name{
				Qualified: "TestNamespace",
			},
		},
	}

	// Test case
	result := v.nameObject("TestObject")
	expected := "TestNamespaceTestObject"

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestPhpVisitorNameMethod(t *testing.T) {
	// Initialize PhpVisitor
	v := &PhpVisitor{
		currentClass: &pb.StmtClass{
			Name: &pb.Name{
				Qualified: "TestClass",
			},
		},
	}

	// Test case
	result := v.nameMethod("TestMethod")
	expected := "TestClass::TestMethod"

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestFindPhpDocBlock(t *testing.T) {
	v := &PhpVisitor{
		linesOfFile: []string{
			"// This is a single line comment",
			"/*",
			"This is a multiline comment",
			"*/",
			"function test() {}",
		},
	}
	linesOfCode := &pb.LinesOfCode{}

	v.findPhpDocBlock(4, 4, linesOfCode)

	if linesOfCode.CommentLinesOfCode != 3 {
		t.Errorf("Expected 3, but got %d", linesOfCode.CommentLinesOfCode)
	}
}

func TestFindPhpDocBlock_NoDocBlock(t *testing.T) {
	v := &PhpVisitor{
		linesOfFile: []string{
			"// This is a single line comment",
			"function test() {}",
		},
	}
	linesOfCode := &pb.LinesOfCode{}

	v.findPhpDocBlock(1, 1, linesOfCode)

	if linesOfCode.CommentLinesOfCode != 0 {
		t.Errorf("Expected 0, but got %d", linesOfCode.CommentLinesOfCode)
	}
}

func TestFindPhpDocBlock_ImproperDocBlock(t *testing.T) {
	v := &PhpVisitor{
		linesOfFile: []string{
			"/*",
			"This is a multiline comment",
			"function test() {}",
		},
	}
	linesOfCode := &pb.LinesOfCode{}

	v.findPhpDocBlock(2, 2, linesOfCode)

	if linesOfCode.CommentLinesOfCode != 0 {
		t.Errorf("Expected 0, but got %d", linesOfCode.CommentLinesOfCode)
	}
}
