package Storage

import (
    "os"
    "path/filepath"
)

func Path() string {
    // workdir: folder ".ast-metrics" in the current directory
    workDir, err := os.Getwd()
    if err != nil {
        panic(err)
    }
    workDir = filepath.Join(workDir, ".ast-metrics-cache")

    return workDir
}

func Ensure() {
    workDir := Path()
    // create workdir if not exists
    if _, err := os.Stat(workDir); os.IsNotExist(err) {
        os.Mkdir(workDir, 0755)
    }
}

func Purge() {
    workDir := Path()
    os.RemoveAll(workDir)
}
