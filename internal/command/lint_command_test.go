package command

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/halleck45/ast-metrics/internal/storage"
	"github.com/pterm/pterm"
	"google.golang.org/protobuf/proto"
)

// fakeRunner dumps a single pb.File into storage so analyzer can read it
 type fakeRunner struct { cfg *configuration.Configuration }

func (f *fakeRunner) IsRequired() bool { return true }
func (f *fakeRunner) Ensure() error { return nil }
func (f *fakeRunner) DumpAST() {
	// create one file with zero LOC
	file := &pb.File{ Path: filepath.Join(os.TempDir(), "sample.php"), Stmts: &pb.Stmts{ Analyze: &pb.Analyze{ Volume: &pb.Volume{ Loc: int32Ptr(0) } } } }
	b, _ := proto.Marshal(file)
	_ = os.MkdirAll(f.cfg.Storage.Path(), 0o755)
	_ = os.WriteFile(filepath.Join(f.cfg.Storage.Path(), "one.bin"), b, 0o644)
}
func (f *fakeRunner) Finish() error { return nil }
func (f *fakeRunner) SetProgressbar(_ *pterm.SpinnerPrinter) {}
func (f *fakeRunner) SetConfiguration(c *configuration.Configuration) { f.cfg = c }
func (f *fakeRunner) Parse(filepath string) (*pb.File, error) { return &pb.File{Path: filepath}, nil }

func int32Ptr(v int32) *int32 { return &v }

func TestLintCommand_Execute_ReturnsErrorOnViolations(t *testing.T) {
	// Setup
	work := storage.Default()
	work.Purge()
	work.Ensure()

	cfg := configuration.NewConfiguration()
	cfg.Storage = work
	cfg.Requirements = configuration.NewConfigurationRequirements()
	cfg.Requirements.FailOnError = false
	cfg.Requirements.Rules.Volume.Loc = &configuration.ConfigurationDefaultRule{ Min: 1 }

	outWriter := bufio.NewWriter(os.Stdout)

	var runners []engine.Engine = []engine.Engine{ &fakeRunner{} }
	cmd := NewLintCommand(cfg, outWriter, runners)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected an error when violations exist, got nil")
	}
}

func TestExtractPathAndStrip(t *testing.T) {
	f := &pb.File{ Path: "/tmp/foo.php" }
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
