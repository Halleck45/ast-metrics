package Configuration

import (
	"os"
	"testing"
)

func TestConfigurationAcceptsSourcesToAnalyzePath(t *testing.T) {

	// create temporary folders
	defer os.RemoveAll("/tmp/folder1")
	defer os.RemoveAll("/tmp/folder2")
	os.Mkdir("/tmp/folder1", 0777)
	os.Mkdir("/tmp/folder2", 0777)
	configuration := NewConfiguration()
	err := configuration.SetSourcesToAnalyzePath([]string{"/tmp/folder1", "/tmp/folder2"})
	if err != nil {
		t.Errorf("Error setting sources to analyze path")
		return
	}

	// len should be 1
	if len(configuration.SourcesToAnalyzePath) != 2 {
		t.Errorf("SourcesToAnalyzePath should have 2 elements")
		return
	}

	if configuration.SourcesToAnalyzePath[0] != "/tmp/folder1" {
		t.Errorf("SourcesToAnalyzePath = %s; want %s", configuration.SourcesToAnalyzePath[1], "/tmp/folder1")
	}

	if configuration.SourcesToAnalyzePath[1] != "/tmp/folder2" {
		t.Errorf("SourcesToAnalyzePath = %s; want %s", configuration.SourcesToAnalyzePath[2], "/tmp/folder2")
	}
}

func TestConfigurationAcceptsExcludePatterns(t *testing.T) {

	configuration := NewConfiguration()
	configuration.SetExcludePatterns([]string{"/foo"})

	if configuration.ExcludePatterns[0] != "/foo" {
		t.Errorf("ExcludePatterns = %s; want %s", configuration.ExcludePatterns[0], "/foo")
	}
}
