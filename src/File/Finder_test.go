package File

import (
	"os"
	"strings"
	"testing"

	"github.com/halleck45/ast-metrics/src/Configuration"
)

func TestFinder_Search(t *testing.T) {
	t.Run("should return a list of files under multiple directories", func(t *testing.T) {

		// First we create two directories with files
		os.Mkdir("/tmp/test1", 0777)
		os.Mkdir("/tmp/test2", 0777)
		os.Create("/tmp/test1/file1.js")
		os.Create("/tmp/test1/file2.js")
		os.Create("/tmp/test2/file3.js")

		// Then we create a Finder
		finder := Finder{Configuration: Configuration.Configuration{SourcesToAnalyzePath: []string{"/tmp/test1", "/tmp/test2"}}}

		// Then we search for files
		result := finder.Search(".js")

		// Then we check the result
		if len(result.Files) != 3 {
			t.Errorf("Expected 3 files, got %d", len(result.Files))
		}

		if len(result.FilesByDirectory) != 2 {
			t.Errorf("Expected 2 directories, got %d", len(result.FilesByDirectory))
		}

		if len(result.FilesByDirectory["/tmp/test1"]) != 2 {
			t.Errorf("Expected 2 files in /tmp/test1, got %d", len(result.FilesByDirectory["/tmp/test1"]))
		}

		if len(result.FilesByDirectory["/tmp/test2"]) != 1 {
			t.Errorf("Expected 1 file in /tmp/test2, got %d", len(result.FilesByDirectory["/tmp/test2"]))
		}

		if !strings.Contains(result.FilesByDirectory["/tmp/test1"][0], "/tmp/test1/file1.js") {
			t.Errorf("Expected /tmp/test1/file1.js, got %s", result.FilesByDirectory["/tmp/test1"][0])
		}

		if !strings.Contains(result.FilesByDirectory["/tmp/test1"][1], "/tmp/test1/file2.js") {
			t.Errorf("Expected /tmp/test1/file2.js, got %s", result.FilesByDirectory["/tmp/test1"][1])
		}

		// Finally we remove the directories and files
		os.Remove("/tmp/test1/file1.js")
		os.Remove("/tmp/test1/file2.js")
		os.Remove("/tmp/test2/file3.js")
		os.Remove("/tmp/test1")
		os.Remove("/tmp/test2")
	})
}
