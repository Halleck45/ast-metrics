package ruleset

import (
	"github.com/halleck45/ast-metrics/internal/configuration"
)

// Object-oriented programming ruleset

type oopRuleset struct {
	cfg *configuration.ConfigurationRequirements
}

func (o *oopRuleset) Category() string {
	return "object-oriented-programming"
}
func (o *oopRuleset) Description() string {
	return "Object-oriented metrics (e.g., maintainability index)"
}
func (o *oopRuleset) Enabled() []Rule {
	rules := []Rule{}
	if o.cfg.Rules.ObjectOrientedProgramming != nil && o.cfg.Rules.ObjectOrientedProgramming.Maintainability != nil {
		rules = append(rules, NewMaintainabilityRule(o.cfg.Rules.ObjectOrientedProgramming.Maintainability))
	}
	return rules
}
func (o *oopRuleset) All() []Rule {
	return []Rule{
		NewMaintainabilityRule(o.cfg.Rules.ObjectOrientedProgramming.Maintainability),
	}
}

func (o *oopRuleset) IsEnabled() bool {
	enabled := o.Enabled()
	return len(enabled) > 0
}
