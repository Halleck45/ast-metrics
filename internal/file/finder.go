package file

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/pterm/pterm"
)

type FileList struct {
	Files            []string
	FilesByDirectory map[string][]string
}

type Finder struct {
	Configuration configuration.Configuration
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
		var walkErr error
		// if is a PHP file, add it
		if strings.HasSuffix(path, fileExtension) {
			matches = append(matches, path)
		} else {
			// Use filepath.Walk for recursive search instead of filepathx.Glob
			// which seems to have issues with absolute paths
			walkErr = filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
				if err != nil {
					// Continue on errors (permission denied, etc.)
					return nil
				}
				if !info.IsDir() && strings.HasSuffix(walkPath, fileExtension) {
					matches = append(matches, walkPath)
				}
				return nil
			})
			if walkErr != nil {
				pterm.Warning.Printf("Walk error for %s: %v\n", path, walkErr)
			}
		}

		// deal with excluded files
		excludedCount := 0
		for _, file := range matches {
			var excluded bool = false

			for _, excludedFile := range r.Configuration.ExcludePatterns {
				excluded, _ = regexp.MatchString(excludedFile, file)
				if excluded {
					excludedCount++
					break
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
