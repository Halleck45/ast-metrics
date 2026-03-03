package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestFinder_Search(t *testing.T) {
	t.Run("should return a list of files under multiple directories", func(t *testing.T) {

		// First we create two directories with files (portable)
		base := t.TempDir()
		dir1 := filepath.Join(base, "test1")
		dir2 := filepath.Join(base, "test2")
		_ = os.MkdirAll(dir1, 0o777)
		_ = os.MkdirAll(dir2, 0o777)
		_ = os.WriteFile(filepath.Join(dir1, "file1.js"), []byte("// test\n"), 0o644)
		_ = os.WriteFile(filepath.Join(dir1, "file2.js"), []byte("// test\n"), 0o644)
		_ = os.WriteFile(filepath.Join(dir2, "file3.js"), []byte("// test\n"), 0o644)

		// Then we create a Finder
		finder := Finder{Configuration: configuration.Configuration{SourcesToAnalyzePath: []string{dir1, dir2}}}

		// Then we search for files
		result := finder.Search(".js")

		// Then we check the result
		if len(result.Files) != 3 {
			t.Errorf("Expected 3 files, got %d", len(result.Files))
		}

		if len(result.FilesByDirectory) != 2 {
			t.Errorf("Expected 2 directories, got %d", len(result.FilesByDirectory))
		}

		if len(result.FilesByDirectory[dir1]) != 2 {
			t.Errorf("Expected 2 files in %s, got %d", dir1, len(result.FilesByDirectory[dir1]))
		}

		if len(result.FilesByDirectory[dir2]) != 1 {
			t.Errorf("Expected 1 file in %s, got %d", dir2, len(result.FilesByDirectory[dir2]))
		}

		if !strings.Contains(result.FilesByDirectory[dir1][0], filepath.Join(dir1, "file1.js")) && !strings.Contains(result.FilesByDirectory[dir1][1], filepath.Join(dir1, "file1.js")) {
			t.Errorf("Expected %s in directory listing", filepath.Join(dir1, "file1.js"))
		}

		if !strings.Contains(result.FilesByDirectory[dir1][0], filepath.Join(dir1, "file2.js")) && !strings.Contains(result.FilesByDirectory[dir1][1], filepath.Join(dir1, "file2.js")) {
			t.Errorf("Expected %s in directory listing", filepath.Join(dir1, "file2.js"))
		}
	})
}

func TestMergeFileLists(t *testing.T) {
	t.Run("merges multiple file lists", func(t *testing.T) {
		list1 := FileList{
			Files:            []string{"a.php", "b.php"},
			FilesByDirectory: map[string][]string{"/src": {"a.php", "b.php"}},
		}
		list2 := FileList{
			Files:            []string{"c.inc"},
			FilesByDirectory: map[string][]string{"/src": {"c.inc"}},
		}
		merged := MergeFileLists(list1, list2)
		if len(merged.Files) != 3 {
			t.Errorf("Expected 3 files, got %d", len(merged.Files))
		}
		if len(merged.FilesByDirectory["/src"]) != 3 {
			t.Errorf("Expected 3 files in /src, got %d", len(merged.FilesByDirectory["/src"]))
		}
	})

	t.Run("handles empty lists", func(t *testing.T) {
		merged := MergeFileLists()
		if len(merged.Files) != 0 {
			t.Errorf("Expected 0 files, got %d", len(merged.Files))
		}
	})
}

func TestFinder_SearchMultiple(t *testing.T) {
	t.Run("should find files of multiple extensions in a single walk", func(t *testing.T) {
		base := t.TempDir()
		_ = os.WriteFile(filepath.Join(base, "main.go"), []byte("package main\n"), 0o644)
		_ = os.WriteFile(filepath.Join(base, "index.php"), []byte("<?php\n"), 0o644)
		_ = os.WriteFile(filepath.Join(base, "app.py"), []byte("pass\n"), 0o644)
		_ = os.WriteFile(filepath.Join(base, "lib.rs"), []byte("fn main() {}\n"), 0o644)
		_ = os.WriteFile(filepath.Join(base, "readme.txt"), []byte("hello\n"), 0o644)

		finder := Finder{Configuration: configuration.Configuration{SourcesToAnalyzePath: []string{base}}}
		results := finder.SearchMultiple([]string{".go", ".php", ".py", ".rs"})

		if len(results[".go"].Files) != 1 {
			t.Errorf("Expected 1 .go file, got %d", len(results[".go"].Files))
		}
		if len(results[".php"].Files) != 1 {
			t.Errorf("Expected 1 .php file, got %d", len(results[".php"].Files))
		}
		if len(results[".py"].Files) != 1 {
			t.Errorf("Expected 1 .py file, got %d", len(results[".py"].Files))
		}
		if len(results[".rs"].Files) != 1 {
			t.Errorf("Expected 1 .rs file, got %d", len(results[".rs"].Files))
		}
	})

	t.Run("should use SearchMultiple cache in Search", func(t *testing.T) {
		base := t.TempDir()
		_ = os.WriteFile(filepath.Join(base, "main.go"), []byte("package main\n"), 0o644)
		_ = os.WriteFile(filepath.Join(base, "other.go"), []byte("package main\n"), 0o644)

		finder := Finder{Configuration: configuration.Configuration{SourcesToAnalyzePath: []string{base}}}
		discovery := &FileDiscovery{}
		discovery.Precompute(finder, []string{".go"})
		finder.Discovery = discovery

		result := finder.Search(".go")
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 .go files from cache, got %d", len(result.Files))
		}
	})
}
