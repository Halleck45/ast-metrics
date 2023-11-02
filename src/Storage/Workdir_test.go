package Storage

import (
    "testing"
    "strings"
    "os"
)

func TestItReturnsPath(t *testing.T) {
    providedPath := Path()

    // providedPath should contain ".ast-metrics-cache" folder
    expectedPath := ".ast-metrics-cache"
    if strings.Contains(providedPath, expectedPath) == false {
        t.Errorf("Path() = %s; want %s", providedPath, expectedPath)
    }
}

func TestItCreatesPath(t *testing.T) {
    providedPath := Path()

    // Create folder
    Ensure()

    // providedPath should exist
    if _, err := os.Stat(providedPath); os.IsNotExist(err) {
        t.Errorf("Path() = %s; want it to exist", providedPath)
    }

    // Remove folder
    Purge()

    // providedPath should not exist
    if _, err := os.Stat(providedPath); os.IsNotExist(err) == false {
        t.Errorf("Path() = %s; want it to not exist", providedPath)
    }
}