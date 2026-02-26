package command

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/cli"
	"github.com/halleck45/ast-metrics/internal/configuration"
)

type InitConfigurationCommand struct {
}

func NewInitConfigurationCommand() *InitConfigurationCommand {
	return &InitConfigurationCommand{}
}
func (v *InitConfigurationCommand) Execute() error {
	fmt.Print(cli.ScreenHeader("Init configuration"))
	fmt.Println()

	loader := configuration.NewConfigurationLoader()

	err := loader.CreateDefaultFile()
	if err != nil {
		cli.PrintError(err.Error())
		return err
	}

	cli.PrintSuccess("Configuration file created: .ast-metrics.yaml")

	return nil
}
