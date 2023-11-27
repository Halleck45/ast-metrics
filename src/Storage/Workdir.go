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

func OutputPath() string {
	workDir := Path()
	return filepath.Join(workDir, "output")
}

func Ensure() {
	workDir := Path()
	// create workdir if not exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		os.Mkdir(workDir, 0755)
	}

	// Ensure outdir exists
	outputDir := OutputPath()
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}
}

func Purge() {
	workDir := Path()
	os.RemoveAll(workDir)
}
