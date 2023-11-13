package CommandExecutor

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/halleck45/ast-metrics/src/Storage"
)

type EmbeddedWorkspace struct {
	Name               string
	PathToLocalSources embed.FS
}

func (r EmbeddedWorkspace) Ensure() error {
	// clean up
	r.Cleanup()

	// Install sources locally (vendors)
	tempDir := r.GetPath()
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return err
	}

	// Extract PHP sources for directories "vendor", etc
	if err := fs.WalkDir(r.PathToLocalSources, ".", func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if d.Type().IsRegular() {
			content, err := r.PathToLocalSources.ReadFile(path)
			if err != nil {
				return err
			}
			outputPath := tempDir + string(os.PathSeparator) + path
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(outputPath, content, 0644); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r EmbeddedWorkspace) Cleanup() error {

	// Remove temp directory
	tempDir := r.GetPath()

	// check if tempDir exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		return nil
	}
	if err := os.RemoveAll(tempDir); err != nil {
		return err
	}

	return nil
}

func (r EmbeddedWorkspace) GetPath() string {
	return Storage.Path() + string(os.PathSeparator) + r.Name + string(os.PathSeparator) + ".temp"
}
