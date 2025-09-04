package command

import (
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/pterm/pterm"
)

type InitConfigurationCommand struct {
}

func NewInitConfigurationCommand() *InitConfigurationCommand {
	return &InitConfigurationCommand{}
}
func (v *InitConfigurationCommand) Execute() error {

	loader := configuration.NewConfigurationLoader()

	err := loader.CreateDefaultFile()
	if err != nil {
		pterm.Error.Println(err.Error())
		return err
	}

	pterm.Success.Println("Configuration file created: .ast-metrics.yaml")

	return nil
}
