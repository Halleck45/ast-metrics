package Command

import (
    "testing"
    "os"
    "github.com/halleck45/ast-metrics/src/Storage"
)

func TestCommandCleanupWorkspace(t *testing.T) {

    // Suite
    t.Run("TestCommandCleanupWorkspace", func(t *testing.T) {

        storage := Storage.Default()
        providedPath := storage.Path()

        // Create folder
        storage.Ensure()

        // providedPath should exist
        if _, err := os.Stat(providedPath); os.IsNotExist(err) {
            t.Errorf("Path() = %s; want it to exist", providedPath)
        }

        // Run command
        cmd := NewCleanCommand()
        err := cmd.Execute()

        if err != nil {
            t.Errorf("CleanCommand.Execute() = %s; want it to be nil", err.Error())
        }

        // providedPath should not exist
        if _, err := os.Stat(providedPath); os.IsNotExist(err) == false {
            t.Errorf("Path() = %s; want it to not exist", providedPath)
        }
    })
}