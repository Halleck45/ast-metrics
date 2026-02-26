package command

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/halleck45/ast-metrics/internal/analyzer/ruleset"
	"github.com/halleck45/ast-metrics/internal/cli"
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

		fmt.Print(cli.ScreenHeader("Ruleset: " + ruleset.Category()))
		fmt.Println()

		italic := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true)
		fmt.Println("  " + italic.Render(ruleset.Description()))
		fmt.Println()

		nbRules := len(ruleset.All())
		fmt.Printf("  Found %d rules in this ruleset\n\n", nbRules)

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
