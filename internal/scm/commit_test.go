package scm

import "testing"

func TestCommit_Structure(t *testing.T) {
	commit := Commit{
		Hash:      "abc123",
		Author:    "John Doe",
		Timestamp: 1234567890,
		Files:     []string{"file1.go", "file2.go"},
	}

	if commit.Hash != "abc123" {
		t.Errorf("expected hash 'abc123', got %s", commit.Hash)
	}
	if commit.Author != "John Doe" {
		t.Errorf("expected author 'John Doe', got %s", commit.Author)
	}
	if commit.Timestamp != 1234567890 {
		t.Errorf("expected timestamp 1234567890, got %d", commit.Timestamp)
	}
	if len(commit.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(commit.Files))
	}
}

func TestCommit_ZeroValue(t *testing.T) {
	var commit Commit

	if commit.Hash != "" {
		t.Errorf("expected empty hash, got %s", commit.Hash)
	}
	if commit.Author != "" {
		t.Errorf("expected empty author, got %s", commit.Author)
	}
	if commit.Timestamp != 0 {
		t.Errorf("expected timestamp 0, got %d", commit.Timestamp)
	}
	if commit.Files != nil {
		t.Errorf("expected nil files, got %v", commit.Files)
	}
}
