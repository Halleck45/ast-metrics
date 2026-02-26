package command

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/cli"
	"github.com/halleck45/ast-metrics/internal/storage"
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
	fmt.Print(cli.ScreenHeader("Clean"))
	fmt.Println()
	v.Storage.Purge()
	cli.PrintSuccess("Workdir cleaned")
	return nil
}
