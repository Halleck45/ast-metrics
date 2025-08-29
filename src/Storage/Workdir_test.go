package Storage

import (
	"os"
	"strings"
	"testing"
)

func TestItReturnsPath(t *testing.T) {
	storage := Default()
	providedPath := storage.WorkDir()

	// providedPath should contain ".ast-metrics-cache" folder
	expectedPath := "ast-metrics-cache"
	if strings.Contains(providedPath, expectedPath) == false {
		t.Errorf("WorkDir() = %s; want %s", providedPath, expectedPath)
	}
}

func TestItCreatesPath(t *testing.T) {
	storage := Default()
	providedPath := storage.WorkDir()

	// Create folder
	storage.Ensure()

	// providedPath should exist
	if _, err := os.Stat(providedPath); os.IsNotExist(err) {
		t.Errorf("WorkDir() = %s; want it to exist", providedPath)
	}

	// Remove folder
	storage.Purge()

	// providedPath should not exist
	if _, err := os.Stat(providedPath); os.IsNotExist(err) == false {
		t.Errorf("WorkDir() = %s; want it to not exist", providedPath)
	}
}
