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
	if c == nil || c.cfg == nil || c.cfg.Rules == nil || c.cfg.Rules.Complexity == nil {
		return rules
	}
	if c.cfg.Rules.Complexity.Cyclomatic != nil {
		rules = append(rules, NewCyclomaticRule(c.cfg.Rules.Complexity.Cyclomatic))
	}
	return rules
}
func (c *complexityRuleset) All() []Rule {
	var cyclo *configuration.ConfigurationDefaultRule
	if c != nil && c.cfg != nil && c.cfg.Rules != nil && c.cfg.Rules.Complexity != nil {
		cyclo = c.cfg.Rules.Complexity.Cyclomatic
	}
	return []Rule{
		NewCyclomaticRule(cyclo),
	}
}

func (c *complexityRuleset) IsEnabled() bool {
	enabled := c.Enabled()
	return len(enabled) > 0
}
