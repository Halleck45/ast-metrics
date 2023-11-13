package CommandExecutor

import (
	"embed"
	"strings"
	"testing"
)

func TestEmbeddedWorkspace_Ensure(t *testing.T) {

	t.Run("should extract files from embed.FS to local directory", func(t *testing.T) {

		embed := embed.FS{}

		workspace := EmbeddedWorkspace{
			Name:               "test",
			PathToLocalSources: embed,
		}

		defer workspace.Cleanup()
		err := workspace.Ensure()

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("should be specific", func(t *testing.T) {

		embed := embed.FS{}

		workspace := EmbeddedWorkspace{
			Name:               "test",
			PathToLocalSources: embed,
		}

		defer workspace.Cleanup()
		path := workspace.GetPath()

		// should contain ".ast-metrics-cache/test/.temp"
		if strings.HasSuffix(path, ".ast-metrics-cache/test/.temp") == false {
			t.Errorf("Expected path to end with .ast-metrics-cache/test/.temp, got %v", path)
		}
	})
}
