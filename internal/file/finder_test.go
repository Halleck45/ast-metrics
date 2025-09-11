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
