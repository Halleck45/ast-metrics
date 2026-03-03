package file

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/yargevad/filepathx"
)

type FileList struct {
	Files            []string
	FilesByDirectory map[string][]string
}

// FileDiscovery caches multi-extension search results so that
// multiple calls to Search() with different extensions reuse a single walk.
type FileDiscovery struct {
	results map[string]FileList
}

// Precompute walks all source paths once for the given extensions
// and caches the results.
func (fd *FileDiscovery) Precompute(finder Finder, extensions []string) {
	fd.results = finder.SearchMultiple(extensions)
}

// Get returns the cached FileList for the given extension, or nil if not cached.
func (fd *FileDiscovery) Get(ext string) *FileList {
	if fd.results == nil {
		return nil
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	if fl, ok := fd.results[ext]; ok {
		return &fl
	}
	return nil
}

type Finder struct {
	Configuration configuration.Configuration
	// Discovery is an optional shared cache for multi-extension search.
	// When set, Search() will check it before doing a full glob.
	Discovery *FileDiscovery
}

func (r Finder) Search(fileExtension string) FileList {

	// Ensur extension starts with a dot
	if !strings.HasPrefix(fileExtension, ".") {
		fileExtension = "." + fileExtension
	}

	// Check shared discovery cache first
	if r.Discovery != nil {
		if cached := r.Discovery.Get(fileExtension); cached != nil {
			return *cached
		}
	}

	var result FileList
	result.FilesByDirectory = make(map[string][]string)
	result.Files = []string{}

	// Pre-compile exclude patterns once
	compiledExcludes := make([]*regexp.Regexp, 0, len(r.Configuration.ExcludePatterns))
	for _, pattern := range r.Configuration.ExcludePatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			compiledExcludes = append(compiledExcludes, re)
		}
	}

	// Search for files in each directory
	for _, path := range r.Configuration.SourcesToAnalyzePath {

		path := strings.TrimRight(path, "/")
		var matches []string
		if strings.HasSuffix(path, fileExtension) {
			matches = append(matches, path)
		} else {
			matches, _ = filepathx.Glob(path + "/**/*" + fileExtension)
		}

		// deal with excluded files
		for _, file := range matches {
			excluded := false

			for _, re := range compiledExcludes {
				if re.MatchString(file) {
					excluded = true
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

// SearchMultiple performs a single directory walk and dispatches files by extension.
// Extensions should include the leading dot (e.g. ".go", ".php").
// Returns a map from extension to FileList.
func (r Finder) SearchMultiple(extensions []string) map[string]FileList {
	results := make(map[string]FileList, len(extensions))
	extSet := make(map[string]bool, len(extensions))
	for _, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extSet[ext] = true
		results[ext] = FileList{
			Files:            []string{},
			FilesByDirectory: make(map[string][]string),
		}
	}

	// Pre-compile exclude patterns once
	compiledExcludes := make([]*regexp.Regexp, 0, len(r.Configuration.ExcludePatterns))
	for _, pattern := range r.Configuration.ExcludePatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			compiledExcludes = append(compiledExcludes, re)
		}
	}

	for _, srcPath := range r.Configuration.SourcesToAnalyzePath {
		srcPath = strings.TrimRight(srcPath, "/")

		// If the source path itself is a file, check its extension
		info, err := os.Stat(srcPath)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			ext := filepath.Ext(srcPath)
			if extSet[ext] {
				if !isExcluded(srcPath, compiledExcludes) {
					fl := results[ext]
					fl.Files = append(fl.Files, srcPath)
					fl.FilesByDirectory[srcPath] = append(fl.FilesByDirectory[srcPath], srcPath)
					results[ext] = fl
				}
			}
			continue
		}

		// Single walk for all extensions
		filepath.WalkDir(srcPath, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if !extSet[ext] {
				return nil
			}
			if isExcluded(path, compiledExcludes) {
				return nil
			}
			fl := results[ext]
			fl.Files = append(fl.Files, path)
			if _, ok := fl.FilesByDirectory[srcPath]; !ok {
				fl.FilesByDirectory[srcPath] = []string{}
			}
			fl.FilesByDirectory[srcPath] = append(fl.FilesByDirectory[srcPath], path)
			results[ext] = fl
			return nil
		})
	}

	return results
}

func MergeFileLists(lists ...FileList) FileList {
	result := FileList{Files: []string{}, FilesByDirectory: map[string][]string{}}
	for _, fl := range lists {
		result.Files = append(result.Files, fl.Files...)
		for dir, files := range fl.FilesByDirectory {
			result.FilesByDirectory[dir] = append(result.FilesByDirectory[dir], files...)
		}
	}
	return result
}

func isExcluded(path string, compiledExcludes []*regexp.Regexp) bool {
	for _, re := range compiledExcludes {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}
