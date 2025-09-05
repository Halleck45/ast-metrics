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
	rules := []Rule{
		NewLocRule(v.cfg.Rules.Volume.Loc),
		NewLlocRule(v.cfg.Rules.Volume.Lloc),
		NewLocByMethodRule(v.cfg.Rules.Volume.LocByMethod),
		NewLlocByMethodRule(v.cfg.Rules.Volume.LlocByMethod),
	}
	return rules
}

func (v *volumeRuleset) Enabled() []Rule {
	rules := []Rule{}
	if v.cfg.Rules.Volume != nil {
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
	}
	return rules
}

func (v *volumeRuleset) IsEnabled() bool {
	enabled := v.Enabled()
	return len(enabled) > 0
}
