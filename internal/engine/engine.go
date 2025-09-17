package engine

import (
	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/pterm/pterm"
)

type Engine interface {
	// Returns true when analyzed files are concerned by the programming language
	IsRequired() bool

	// Prepare the engine for the analysis. For example, in order to prepare caches
	Ensure() error

	// First step of analysis. Parse all files, and generate protobuff compatible AST files
	DumpAST()

	// Cleanups the engine. For example, to remove caches
	Finish() error

	// Give a UI progress bar to the engine
	SetProgressbar(progressbar *pterm.SpinnerPrinter)

	// Give the configuration to the engine
	SetConfiguration(configuration *configuration.Configuration)

	// Parse a file and return a protobuff compatible AST object
	Parse(filepath string) (*pb.File, error)
}

type NamedEngine interface {
	Name() string
}
