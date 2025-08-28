package Storage

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Workdir struct {
	directory string
}

func Default() *Workdir {
	return NewWithName("ast-metrics-cache")
}

func NewWithName(path string) *Workdir {
	uniqRand := uuid.New().String()
	return &Workdir{
		directory: os.TempDir() + string(os.PathSeparator) + path + "-" + uniqRand + string(os.PathSeparator),
	}
}

func (s *Workdir) WorkDir() string {
	return s.directory
}

func (s *Workdir) Path() string {
	return s.directory
}

func (s *Workdir) AstDirectory() string {
	workDir := s.WorkDir()
	return filepath.Join(workDir, "ast")
}

func (s *Workdir) Ensure() {
	workDir := s.WorkDir()
	// create workdir if not exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		os.Mkdir(workDir, 0755)
	}

	// Ensure outdir exists
	outputDir := s.AstDirectory()
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}
}

func (s *Workdir) Purge() {
	workDir := s.WorkDir()
	os.RemoveAll(workDir)
}

func (s *Workdir) DeleteCache(filePath string) {
	// realpath filePath
	//filePath, err := filepath.Abs(filePath)
	//if err != nil {
	//	return
	//}

	// If hash has not changed, we can delete the file directly
	hash, err := GetFileHash(filePath)
	if err == nil {
		binPath := s.AstDirectory() + string(os.PathSeparator) + hash + ".bin"
		if _, err := os.Stat(binPath); err == nil {
			os.Remove(binPath)
			return
		}
	}

	// If hash has  changed, we iterate over all files in order to retrieve it via the Path attribute
	files, err := os.ReadDir(s.AstDirectory())
	if err != nil {
		return
	}

	for _, file := range files {
		// load the file via protobuf
		// if the path is the same, we remove it
		binPath := s.AstDirectory() + string(os.PathSeparator) + file.Name()
		pbFile, err := UnmarshalProtobuf(binPath)

		if err != nil {
			continue
		}

		if pbFile.Path == filePath {
			os.Remove(binPath)
			return
		}
	}

}
