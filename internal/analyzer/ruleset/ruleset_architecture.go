package ruleset

import (
	"github.com/halleck45/ast-metrics/internal/configuration"
)

// Architecture ruleset
type architectureRuleset struct {
	cfg *configuration.ConfigurationRequirements
}

func (a *architectureRuleset) Category() string {
	return "architecture"
}
func (a *architectureRuleset) Description() string {
	return "Architecture-related constraints (e.g., coupling)"
}
func (a *architectureRuleset) Enabled() []Rule {
	rules := []Rule{}
	if a.cfg.Rules.Architecture != nil && a.cfg.Rules.Architecture.Coupling != nil {
		rules = append(rules, NewCouplingRule(a.cfg.Rules.Architecture.Coupling))
	}
	return rules
}

func (a *architectureRuleset) All() []Rule {
	return []Rule{
		NewCouplingRule(a.cfg.Rules.Architecture.Coupling),
	}
}

func (a *architectureRuleset) IsEnabled() bool {
	enabled := a.Enabled()
	return len(enabled) > 0
}
