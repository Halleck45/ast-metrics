package Configuration

import (
	"os"
	"path/filepath"

	"github.com/halleck45/ast-metrics/src/Driver"
)

type Configuration struct {
	// The path to the sources to analyze
	SourcesToAnalyzePath []string

	// Exclude patterns (list of regular expressions. When a file matches one of these patterns, it is not analyzed)
	ExcludePatterns []string

	// Drivers to use
	Driver Driver.Driver
}

func NewConfiguration() *Configuration {
	return &Configuration{
		SourcesToAnalyzePath: []string{},
		ExcludePatterns:      []string{"/vendor/", "/node_modules/", "/.git/", "/.idea/", "/tests/", "/Tests/", "/test/", "/Test/", "/spec/", "/Spec/"},
		Driver:               Driver.Docker,
	}
}

func (c *Configuration) SetSourcesToAnalyzePath(paths []string) error {

	(*c).SourcesToAnalyzePath = []string{}

	// foreach path, make it absolute
	for i := range paths {

		// ensure path exists
		if _, err := os.Stat(paths[i]); err != nil {
			return err
		}

		// make path absolute
		if !filepath.IsAbs(paths[i]) {
			var err error
			paths[i], err = filepath.Abs(paths[i])
			if err != nil {
				return err
			}
		}

		// ensure path exists
		if _, err := os.Stat(paths[i]); err != nil {
			return err
		}

		// Remove trailing slash
		paths[i] = filepath.Clean(paths[i])
	}

	(*c).SourcesToAnalyzePath = paths

	return nil
}

func (c *Configuration) SetExcludePatterns(patterns []string) {
	// Ensure patterns are valid regular expressions
	// @todo
	(*c).ExcludePatterns = patterns
}

func (c *Configuration) SetDriver(driver Driver.Driver) {
	(*c).Driver = driver
}
