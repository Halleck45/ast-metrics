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
	}
	return rules
}

func (v *volumeRuleset) Enabled() []Rule {
	rules := []Rule{}
	if v.cfg.Rules.Volume != nil && v.cfg.Rules.Volume.Loc != nil {
		rules = append(rules, NewLocRule(v.cfg.Rules.Volume.Loc))
	}
	return rules
}

func (v *volumeRuleset) IsEnabled() bool {
	enabled := v.Enabled()
	return len(enabled) > 0
}
