package command

import (
	"bufio"

	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/pterm/pterm"
)

// CICommand runs lint and then generates all reports like `analyze --ci`.
// It returns the lint error (if any) as final status so CI pipelines can fail,
// but will still generate the reports before exiting.
type CICommand struct {
	Configuration *configuration.Configuration
	outWriter     *bufio.Writer
	runners       []engine.Engine
}

func NewCICommand(configuration *configuration.Configuration, outWriter *bufio.Writer, runners []engine.Engine) *CICommand {
	return &CICommand{
		Configuration: configuration,
		outWriter:     outWriter,
		runners:       runners,
	}
}

func (c *CICommand) Execute() error {
	// Force non interactive for CI
	pterm.DisableColor()

	// 1) Run lint first. Do not stop on error; keep it to return later.
	lintCmd := NewLintCommand(c.Configuration, c.outWriter, c.runners)
	lintErr := lintCmd.Execute()

	// 2) Run full analysis with reports (non-interactive)
	analyzeCmd := NewAnalyzeCommand(c.Configuration, c.outWriter, c.runners, false)
	_ = analyzeCmd.Execute()

	// 3) Return lint error (if any) so CI can fail
	return lintErr
}
