package configuration

import (
	"testing"
)

func TestConfigurationAcceptsSourcesToAnalyzePath(t *testing.T) {
	// create temporary folders (portable)
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	configuration := NewConfiguration()
	err := configuration.SetSourcesToAnalyzePath([]string{dir1, dir2})
	if err != nil {
		t.Errorf("Error setting sources to analyze path")
		return
	}

	// expect two entries
	if len(configuration.SourcesToAnalyzePath) != 2 {
		t.Errorf("SourcesToAnalyzePath should have 2 elements")
		return
	}

	if configuration.SourcesToAnalyzePath[0] != dir1 {
		t.Errorf("SourcesToAnalyzePath = %s; want %s", configuration.SourcesToAnalyzePath[0], dir1)
	}

	if configuration.SourcesToAnalyzePath[1] != dir2 {
		t.Errorf("SourcesToAnalyzePath = %s; want %s", configuration.SourcesToAnalyzePath[1], dir2)
	}
}

func TestConfigurationAcceptsExcludePatterns(t *testing.T) {

	configuration := NewConfiguration()
	configuration.SetExcludePatterns([]string{"/foo"})

	if configuration.ExcludePatterns[0] != "/foo" {
		t.Errorf("ExcludePatterns = %s; want %s", configuration.ExcludePatterns[0], "/foo")
	}
}
