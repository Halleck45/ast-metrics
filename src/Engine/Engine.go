package Engine


import (
    "embed"
    "github.com/pterm/pterm"
    "github.com/halleck45/ast-metrics/src/Driver"
)

type Engine interface {
    IsRequired() (bool)
    Ensure() (error)
    DumpAST()
    Finish() (error)
    SetProgressbar(progressbar *pterm.SpinnerPrinter)
    SetSourcesToAnalyzePath(path string)
    SetEmbeddedSources(sources embed.FS)
    SetDriver(driver Driver.Driver)
}
