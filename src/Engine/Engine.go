package Engine


import (
    "embed"
    "github.com/pterm/pterm"
)

type Engine interface {
    IsRequired() (bool)
    Ensure() (error)
    DumpAST()
    Finish() (error)
    SetProgressbar(progressbar *pterm.SpinnerPrinter)
    SetSourcesToAnalyzePath(path string)
    SetEmbeddedSources(sources embed.FS)
}
