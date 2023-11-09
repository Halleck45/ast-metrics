package Engine

import (
    "github.com/pterm/pterm"
    "github.com/halleck45/ast-metrics/src/Configuration"
)

type Engine interface {
    IsRequired() (bool)
    Ensure() (error)
    DumpAST()
    Finish() (error)
    SetProgressbar(progressbar *pterm.SpinnerPrinter)
    SetConfiguration(configuration *Configuration.Configuration)
}
