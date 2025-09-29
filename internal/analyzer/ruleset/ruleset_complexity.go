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
	if c == nil || c.cfg == nil || c.cfg.Rules == nil {
		return rules
	}
	// New API
	if c.cfg.Rules.Complexity != nil && c.cfg.Rules.Complexity.Cyclomatic != nil {
		rules = append(rules, NewCyclomaticRule(c.cfg.Rules.Complexity.Cyclomatic))
	}
	// Legacy support: requirements.rules.cyclomatic_complexity: { max: X }
	if c.cfg.Rules.CyclomaticLegacy != nil && c.cfg.Rules.CyclomaticLegacy.Max > 0 {
		m := c.cfg.Rules.CyclomaticLegacy.Max
		rules = append(rules, NewCyclomaticRule(&m))
	}
	return rules
}
func (c *complexityRuleset) All() []Rule {
	var cyclo *int
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
