package ruleset

import (
	"github.com/halleck45/ast-metrics/internal/configuration"
)

// Volume ruleset

type volumeRuleset struct {
	cfg *configuration.ConfigurationRequirements
}

func (v *volumeRuleset) Category() string {
	return "volume"
}
func (v *volumeRuleset) Description() string {
	return "Volume metrics (e.g., lines of code)"
}
func (v *volumeRuleset) All() []Rule {
	var loc, lloc, locByMethod, llocByMethod, maxMethodsPerClass, maxSwitchCases, maxParametersPerMethod, maxNestedBlocks, maxPublicMethods *int
	if v != nil && v.cfg != nil && v.cfg.Rules != nil && v.cfg.Rules.Volume != nil {
		loc = v.cfg.Rules.Volume.Loc
		lloc = v.cfg.Rules.Volume.Lloc
		locByMethod = v.cfg.Rules.Volume.LocByMethod
		llocByMethod = v.cfg.Rules.Volume.LlocByMethod
		maxMethodsPerClass = v.cfg.Rules.Volume.MaxMethodsPerClass
		maxSwitchCases = v.cfg.Rules.Volume.MaxSwitchCases
		maxParametersPerMethod = v.cfg.Rules.Volume.MaxParametersPerMethod
		maxNestedBlocks = v.cfg.Rules.Volume.MaxNestedBlocks
		maxPublicMethods = v.cfg.Rules.Volume.MaxPublicMethods
	}
	rules := []Rule{
		NewLocRule(loc),
		NewLlocRule(lloc),
		NewLocByMethodRule(locByMethod),
		NewLlocByMethodRule(llocByMethod),
		NewMaxMethodsPerClassRule(maxMethodsPerClass),
		NewMaxSwitchCasesRule(maxSwitchCases),
		NewMaxParametersPerMethodRule(maxParametersPerMethod),
		NewMaxNestedBlocksRule(maxNestedBlocks),
		NewMaxPublicMethodsRule(maxPublicMethods),
	}
	return rules
}

func (v *volumeRuleset) Enabled() []Rule {
	rules := []Rule{}
	if v == nil || v.cfg == nil || v.cfg.Rules == nil || v.cfg.Rules.Volume == nil {
		return rules
	}
	if v.cfg.Rules.Volume.Loc != nil {
		rules = append(rules, NewLocRule(v.cfg.Rules.Volume.Loc))
	}
	if v.cfg.Rules.Volume.Lloc != nil {
		rules = append(rules, NewLlocRule(v.cfg.Rules.Volume.Lloc))
	}
	if v.cfg.Rules.Volume.LocByMethod != nil {
		rules = append(rules, NewLocByMethodRule(v.cfg.Rules.Volume.LocByMethod))
	}
	if v.cfg.Rules.Volume.LlocByMethod != nil {
		rules = append(rules, NewLlocByMethodRule(v.cfg.Rules.Volume.LlocByMethod))
	}
	if v.cfg.Rules.Volume.MaxMethodsPerClass != nil {
		rules = append(rules, NewMaxMethodsPerClassRule(v.cfg.Rules.Volume.MaxMethodsPerClass))
	}
	if v.cfg.Rules.Volume.MaxSwitchCases != nil {
		rules = append(rules, NewMaxSwitchCasesRule(v.cfg.Rules.Volume.MaxSwitchCases))
	}
	if v.cfg.Rules.Volume.MaxParametersPerMethod != nil {
		rules = append(rules, NewMaxParametersPerMethodRule(v.cfg.Rules.Volume.MaxParametersPerMethod))
	}
	if v.cfg.Rules.Volume.MaxNestedBlocks != nil {
		rules = append(rules, NewMaxNestedBlocksRule(v.cfg.Rules.Volume.MaxNestedBlocks))
	}
	if v.cfg.Rules.Volume.MaxPublicMethods != nil {
		rules = append(rules, NewMaxPublicMethodsRule(v.cfg.Rules.Volume.MaxPublicMethods))
	}
	return rules
}

func (v *volumeRuleset) IsEnabled() bool {
	enabled := v.Enabled()
	return len(enabled) > 0
}
