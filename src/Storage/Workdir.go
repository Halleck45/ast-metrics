package Storage

import (
	"os"
	"path/filepath"

	"github.com/halleck45/ast-metrics/src/Engine"
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

func DeleteCache(filePath string) {
	// realpath filePath
	//filePath, err := filepath.Abs(filePath)
	//if err != nil {
	//	return
	//}

	// If hash has not changed, we can delete the file directly
	hash, err := Engine.GetFileHash(filePath)
	if err == nil {
		binPath := OutputPath() + string(os.PathSeparator) + hash + ".bin"
		if _, err := os.Stat(binPath); err == nil {
			os.Remove(binPath)
			return
		}
	}

	// If hash has  changed, we iterate over all files in order to retrieve it via the Path attribute
	files, err := os.ReadDir(OutputPath())
	if err != nil {
		return
	}

	for _, file := range files {
		// load the file via protobuf
		// if the path is the same, we remove it
		binPath := OutputPath() + string(os.PathSeparator) + file.Name()
		pbFile, err := Engine.UnmarshalProtobuf(binPath)

		if err != nil {
			continue
		}

		if pbFile.Path == filePath {
			os.Remove(binPath)
			return
		}
	}

}
