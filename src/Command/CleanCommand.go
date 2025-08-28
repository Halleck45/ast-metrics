package Command

import (
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
)

type CleanCommand struct {
	storage *Storage.Workdir
}

func NewCleanCommand(storage *Storage.Workdir) *CleanCommand {
	return &CleanCommand{
		storage: storage,
	}
}

func (v *CleanCommand) Execute() error {
	v.storage.Purge()
	pterm.Success.Println("Workdir cleaned")
	return nil
}
