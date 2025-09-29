package command

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/halleck45/ast-metrics/internal/analyzer/ruleset"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/pterm/pterm"
)

type RulesetShowCommand struct{ Name string }

func NewRulesetShowCommand(name string) *RulesetShowCommand { return &RulesetShowCommand{Name: name} }

func (c *RulesetShowCommand) Execute() error {
	if c.Name == "" {
		return errors.New("ruleset name is required")
	}

	cfg := configuration.NewConfigurationRequirements()

	sets := ruleset.Registry(cfg).AllRulesets()
	for _, ruleset := range sets {
		if ruleset.Category() != c.Name {
			continue
		}

		title := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Ruleset: %s", ruleset.Category()))
		italic := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true)
		pterm.Printfln("%s\n", title)
		pterm.Printfln("%s\n", italic.Render(ruleset.Description()))

		len := len(ruleset.All())
		pterm.Printfln("Found %d rules in this ruleset\n", len)

		data := pterm.TableData{}
		data = append(data, []string{"Rule Name", "Description"})
		for _, r := range ruleset.All() {
			data = append(data, []string{r.Name(), r.Description()})
		}

		pterm.DefaultTable.WithHasHeader().WithData(data).Render()
		return nil
	}

	return fmt.Errorf("ruleset '%s' not found", c.Name)
}
