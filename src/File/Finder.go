package File

import (
	"strings"

	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/yargevad/filepathx"
)

type FileList struct {
	Files            []string
	FilesByDirectory map[string][]string
}

type Finder struct {
	Configuration Configuration.Configuration
}

func (r Finder) Search(fileExtension string) FileList {

	// Ensur extension starts with a dot
	if !strings.HasPrefix(fileExtension, ".") {
		fileExtension = "." + fileExtension
	}

	var result FileList
	result.FilesByDirectory = make(map[string][]string)
	result.Files = []string{}

	// Search for PHP files in each directory
	for _, path := range r.Configuration.SourcesToAnalyzePath {

		path := strings.TrimRight(path, "/")
		var matches []string
		// if is a PHP file, add it
		if strings.HasSuffix(path, fileExtension) {
			matches = append(matches, path)
		} else {
			matches, _ = filepathx.Glob(path + "/**/*" + fileExtension)
		}

		// deal with excluded files
		for _, file := range matches {
			var excluded bool = false

			for _, excludedFile := range r.Configuration.ExcludePatterns {
				if strings.Contains(file, excludedFile) {
					excluded = true
				}
			}

			if !excluded {
				result.Files = append(result.Files, file)

				// add file to filesByDirectory
				directory := path
				if _, ok := result.FilesByDirectory[directory]; !ok {
					result.FilesByDirectory[directory] = []string{}
				}

				result.FilesByDirectory[directory] = append(result.FilesByDirectory[directory], file)

			}
		}
	}

	return result
}
