package watcher

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestNewCommandWatcher(t *testing.T) {
	config := &configuration.Configuration{
		SourcesToAnalyzePath: []string{"/test/path1", "/test/path2"},
		Watching:             true,
	}

	watcher := NewCommandWatcher(config)

	if watcher.Configuration != config {
		t.Error("expected configuration to be set")
	}

	if len(watcher.SourcesToAnalyzePath) != 2 {
		t.Errorf("expected 2 source paths, got %d", len(watcher.SourcesToAnalyzePath))
	}

	if watcher.SourcesToAnalyzePath[0] != "/test/path1" {
		t.Errorf("expected first path '/test/path1', got %s", watcher.SourcesToAnalyzePath[0])
	}
}

func TestCommandWatcher_Start_WatchingDisabled(t *testing.T) {
	config := &configuration.Configuration{
		Watching: false,
	}

	watcher := NewCommandWatcher(config)
	err := watcher.Start(nil)

	if err != nil {
		t.Errorf("expected no error when watching disabled, got %v", err)
	}
}
