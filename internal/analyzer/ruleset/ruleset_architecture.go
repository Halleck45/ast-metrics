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
	if a.cfg.Rules.Architecture != nil {
		if a.cfg.Rules.Architecture.Coupling != nil {
			rules = append(rules, NewCouplingRule(a.cfg.Rules.Architecture.Coupling))
		}
		if a.cfg.Rules.Architecture.AfferentCoupling != nil {
			rules = append(rules, NewAfferentCouplingRule(a.cfg.Rules.Architecture.AfferentCoupling))
		}
		if a.cfg.Rules.Architecture.EfferentCoupling != nil {
			rules = append(rules, NewEfferentCouplingRule(a.cfg.Rules.Architecture.EfferentCoupling))
		}
		if a.cfg.Rules.Architecture.Maintainability != nil {
			rules = append(rules, NewMaintainabilityRule(a.cfg.Rules.Architecture.Maintainability))
		}
	}
	return rules
}

func (a *architectureRuleset) All() []Rule {
	return []Rule{
		NewCouplingRule(a.cfg.Rules.Architecture.Coupling),
		NewAfferentCouplingRule(a.cfg.Rules.Architecture.AfferentCoupling),
		NewEfferentCouplingRule(a.cfg.Rules.Architecture.EfferentCoupling),
		NewMaintainabilityRule(a.cfg.Rules.Architecture.Maintainability),
	}
}

func (a *architectureRuleset) IsEnabled() bool {
	enabled := a.Enabled()
	return len(enabled) > 0
}
