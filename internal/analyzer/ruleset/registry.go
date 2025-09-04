package ruleset

import (
	"github.com/halleck45/ast-metrics/internal/configuration"
)

type RegistryImpl struct {
	cfg *configuration.ConfigurationRequirements
}

func Registry(cfg *configuration.ConfigurationRequirements) *RegistryImpl {
	return &RegistryImpl{cfg: cfg}
}

func (r *RegistryImpl) AllRulesets() []Ruleset {
	return []Ruleset{
		&architectureRuleset{cfg: r.cfg},
		&volumeRuleset{cfg: r.cfg},
		&complexityRuleset{cfg: r.cfg},
		&oopRuleset{cfg: r.cfg},
	}
}

func (r *RegistryImpl) EnabledRulesets() []Ruleset {
	var enabled []Ruleset
	for _, ruleset := range r.AllRulesets() {
		if !ruleset.IsEnabled() {
			continue
		}

		enabled = append(enabled, ruleset)
	}

	return enabled
}
