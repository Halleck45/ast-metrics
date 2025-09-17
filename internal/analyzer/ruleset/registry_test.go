package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestRegistry_AllRulesets(t *testing.T) {
	cfg := configuration.NewConfigurationRequirements()
	registry := Registry(cfg)

	rulesets := registry.AllRulesets()

	if len(rulesets) != 4 {
		t.Fatalf("expected 4 rulesets, got %d", len(rulesets))
	}

	categories := make(map[string]bool)
	for _, ruleset := range rulesets {
		categories[ruleset.Category()] = true
	}

	expected := []string{"architecture", "volume", "complexity", "golang"}
	for _, category := range expected {
		if !categories[category] {
			t.Errorf("missing ruleset category: %s", category)
		}
	}
}

func TestRegistry_EnabledRulesets_EmptyConfig(t *testing.T) {
	cfg := &configuration.ConfigurationRequirements{}
	registry := Registry(cfg)

	enabled := registry.EnabledRulesets()

	if len(enabled) != 0 {
		t.Fatalf("expected 0 enabled rulesets with empty config, got %d", len(enabled))
	}
}

func TestRegistry_EnabledRulesets_WithRules(t *testing.T) {
	maxCoupling := 10
	cfg := &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Architecture: &configuration.ConfigurationArchitectureRules{
				AfferentCoupling: &maxCoupling,
			},
		},
	}
	registry := Registry(cfg)

	enabled := registry.EnabledRulesets()

	if len(enabled) != 1 {
		t.Fatalf("expected 1 enabled ruleset, got %d", len(enabled))
	}

	if enabled[0].Category() != "architecture" {
		t.Errorf("expected architecture ruleset, got %s", enabled[0].Category())
	}
}
