package command

import (
	"github.com/halleck45/ast-metrics/internal/storage"
	"github.com/pterm/pterm"
)

type CleanCommand struct {
	Storage *storage.Workdir
}

func NewCleanCommand(storage *storage.Workdir) *CleanCommand {
	return &CleanCommand{
		Storage: storage,
	}
}

func (v *CleanCommand) Execute() error {
	v.Storage.Purge()
	pterm.Success.Println("Workdir cleaned")
	return nil
}
