package analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestNewGitAnalyzer(t *testing.T) {
	analyzer := NewGitAnalyzer()
	if analyzer == nil {
		t.Error("expected non-nil GitAnalyzer")
	}
}

func TestGitAnalyzer_Start_EmptyFiles(t *testing.T) {
	analyzer := NewGitAnalyzer()
	
	results := analyzer.Start([]*pb.File{})
	
	if results == nil {
		t.Error("expected non-nil results")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty files, got %d", len(results))
	}
}

func TestGitAnalyzer_Start_WithFiles(t *testing.T) {
	analyzer := NewGitAnalyzer()
	
	files := []*pb.File{
		{Path: "/test/file1.go"},
		{Path: "/test/file2.go"},
	}
	
	results := analyzer.Start(files)
	
	if results == nil {
		t.Error("expected non-nil results")
	}
}
