package Analyzer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindGitRoot(t *testing.T) {
	// Test case where .git directory is found
	t.Run("finds .git directory", func(t *testing.T) {
		// Setup a temporary directory with a .git directory inside
		tmpDir, err := ioutil.TempDir("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		gitDir := filepath.Join(tmpDir, ".git")
		err = os.Mkdir(gitDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		got, err := findGitRoot(tmpDir)
		if err != nil {
			t.Fatalf("findGitRoot() error = %v", err)
		}
		if got != tmpDir {
			t.Errorf("findGitRoot() = %v, want %v", got, tmpDir)
		}
	})

	// Test case where no .git directory is found
	t.Run("does not find .git directory", func(t *testing.T) {
		// Setup a temporary directory without a .git directory
		tmpDir, err := ioutil.TempDir("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		_, err = findGitRoot(tmpDir)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no git repository found") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}
