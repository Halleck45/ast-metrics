package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestRulesetAddCommand_Execute_AddsVolumeToConfig(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Ensure no config file exists; command should create default then modify it
	cmd := NewRulesetAddCommand("volume")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Load the created file
	loader := configuration.NewConfigurationLoader()
	cfg := configuration.NewConfiguration()
	cfg, err := loader.Loads(cfg)
	if err != nil {
		t.Fatalf("load cfg: %v", err)
	}
	if cfg.Requirements == nil || cfg.Requirements.Rules == nil || cfg.Requirements.Rules.Volume == nil {
		t.Fatalf("expected volume rules to be created in config")
	}
	// sanity check defaults written
	if cfg.Requirements.Rules.Volume.Loc == nil || *cfg.Requirements.Rules.Volume.Loc <= 0 {
		t.Fatalf("expected default volume.loc to be set")
	}

	// File should exist in temp dir
	found := false
	for _, fn := range loader.FilenameToChecks {
		if _, statErr := os.Stat(filepath.Join(tmp, fn)); statErr == nil {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a config file to be created in temp dir")
	}
}
