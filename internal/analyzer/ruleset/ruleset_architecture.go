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
	if a == nil || a.cfg == nil || a.cfg.Rules == nil || a.cfg.Rules.Architecture == nil {
		return rules
	}
	arch := a.cfg.Rules.Architecture
	if arch.Coupling != nil {
		rules = append(rules, NewCouplingRule(arch.Coupling))
	}
	if arch.AfferentCoupling != nil {
		rules = append(rules, NewAfferentCouplingRule(arch.AfferentCoupling))
	}
	if arch.EfferentCoupling != nil {
		rules = append(rules, NewEfferentCouplingRule(arch.EfferentCoupling))
	}
	if arch.Maintainability != nil {
		rules = append(rules, NewMaintainabilityRule(arch.Maintainability))
	}
	return rules
}

func (a *architectureRuleset) All() []Rule {
	var coupling *configuration.ConfigurationCouplingRule
	var afferent *int
	var efferent *int
	var maintainability *int
	if a != nil && a.cfg != nil && a.cfg.Rules != nil && a.cfg.Rules.Architecture != nil {
		arch := a.cfg.Rules.Architecture
		coupling = arch.Coupling
		afferent = arch.AfferentCoupling
		efferent = arch.EfferentCoupling
		maintainability = arch.Maintainability
	}
	return []Rule{
		NewCouplingRule(coupling),
		NewAfferentCouplingRule(afferent),
		NewEfferentCouplingRule(efferent),
		NewMaintainabilityRule(maintainability),
	}
}

func (a *architectureRuleset) IsEnabled() bool {
	enabled := a.Enabled()
	return len(enabled) > 0
}
