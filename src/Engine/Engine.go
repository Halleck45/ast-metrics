package Engine

import (
	"github.com/halleck45/ast-metrics/src/Configuration"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/pterm/pterm"
)

type Engine interface {
	IsRequired() bool
	Ensure() error
	DumpAST()
	Finish() error
	SetProgressbar(progressbar *pterm.SpinnerPrinter)
	SetConfiguration(configuration *Configuration.Configuration)
	Parse(filepath string) (*pb.File, error)
}
