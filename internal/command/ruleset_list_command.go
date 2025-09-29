package command

import (
	"github.com/halleck45/ast-metrics/internal/analyzer/ruleset"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/pterm/pterm"
)

type RulesetListCommand struct {
}

func NewRulesetListCommand() *RulesetListCommand {
	return &RulesetListCommand{}
}

func (c *RulesetListCommand) Execute() error {
	cfg := configuration.NewConfigurationRequirements()
	sets := ruleset.Registry(cfg).AllRulesets()

	if len(sets) == 0 {
		pterm.Info.Println("No ruleset available")
		return nil
	}

	data := pterm.TableData{}
	data = append(data, []string{"How to install ruleset", "Description"})
	for _, s := range sets {
		command := "ast-metrics ruleset add " + s.Category()
		data = append(data, []string{command, s.Description()})
	}
	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	return nil
}
