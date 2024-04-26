package Command

import (
    "github.com/pterm/pterm"
    "github.com/halleck45/ast-metrics/src/Storage"
)

type CleanCommand struct {}

func NewCleanCommand() *CleanCommand {
    return &CleanCommand{}
}

func (v *CleanCommand) Execute() error {
    Storage.Default().Purge()
    pterm.Success.Println("Workdir cleaned")
    return nil
}