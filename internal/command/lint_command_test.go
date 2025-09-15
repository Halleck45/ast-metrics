package command

import (
	"bufio"
	"os"
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/halleck45/ast-metrics/internal/engine/php"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/halleck45/ast-metrics/internal/storage"
)

func TestLintCommand_Execute_ReturnsErrorOnViolations(t *testing.T) {
	// Setup
	work := storage.Default()
	work.Purge()
	work.Ensure()

	cfg := configuration.NewConfiguration()
	cfg.Storage = work
	cfg.Requirements = configuration.NewConfigurationRequirements()
	intVal := func(i int) *int { return &i }
	cfg.Requirements.Rules.Volume.Loc = intVal(1)

	outWriter := bufio.NewWriter(os.Stdout)

	phpSource := `<?php
function foo() {
	echo "Hello, World!";
}
function bar() {
	echo "Hello, World!";
}
`
	runners := []engine.Engine{&php.PhpRunner{}}

	// create temporary file
	file, err := os.CreateTemp("", "lint_test_*.php")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name()) // clean up
	if _, err := file.WriteString(phpSource); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	cfg.SourcesToAnalyzePath = []string{file.Name()}
	cmd := NewLintCommand(cfg, outWriter, runners)

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected an error when violations exist, got nil")
	}
}

func TestExtractPathAndStrip(t *testing.T) {
	f := &pb.File{Path: "/tmp/foo.php"}
	files := []*pb.File{f}
	msg := "Lines of code too low in file /tmp/foo.php: got 0 (min: 1)"
	p := extractPath(msg, files)
	if p != f.Path {
		t.Fatalf("extractPath failed, got %q", p)
	}
	stripped := stripPathPrefix(msg, f.Path)
	if stripped == msg {
		t.Fatalf("stripPathPrefix did not strip anything")
	}
}
