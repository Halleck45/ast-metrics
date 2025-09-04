package ruleset

import (
	"github.com/halleck45/ast-metrics/internal/configuration"
)

// Complexity ruleset

type complexityRuleset struct {
	cfg *configuration.ConfigurationRequirements
}

func (c *complexityRuleset) Category() string {
	return "complexity"
}
func (c *complexityRuleset) Description() string {
	return "Complexity metrics (e.g., cyclomatic complexity)"
}
func (c *complexityRuleset) Enabled() []Rule {
	rules := []Rule{}
	if c.cfg.Rules.Complexity != nil && c.cfg.Rules.Complexity.Cyclomatic != nil {
		rules = append(rules, NewCyclomaticRule(c.cfg.Rules.Complexity.Cyclomatic))
	}
	return rules
}
func (c *complexityRuleset) All() []Rule {
	return []Rule{
		NewCyclomaticRule(c.cfg.Rules.Complexity.Cyclomatic),
	}
}

func (c *complexityRuleset) IsEnabled() bool {
	enabled := c.Enabled()
	return len(enabled) > 0
}
