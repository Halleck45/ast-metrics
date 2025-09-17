package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestVolumeRuleset_Category(t *testing.T) {
	ruleset := &volumeRuleset{}
	if ruleset.Category() != "volume" {
		t.Errorf("expected 'volume', got %s", ruleset.Category())
	}
}

func TestVolumeRuleset_Description(t *testing.T) {
	ruleset := &volumeRuleset{}
	expected := "Volume metrics (e.g., lines of code)"
	if ruleset.Description() != expected {
		t.Errorf("expected '%s', got %s", expected, ruleset.Description())
	}
}

func TestVolumeRuleset_IsEnabled_EmptyConfig(t *testing.T) {
	ruleset := &volumeRuleset{cfg: &configuration.ConfigurationRequirements{}}
	if ruleset.IsEnabled() {
		t.Error("expected ruleset to be disabled with empty config")
	}
}

func TestVolumeRuleset_IsEnabled_WithRules(t *testing.T) {
	maxLoc := 1000
	cfg := &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Volume: &configuration.ConfigurationVolumeRules{
				Loc: &maxLoc,
			},
		},
	}
	ruleset := &volumeRuleset{cfg: cfg}
	
	if !ruleset.IsEnabled() {
		t.Error("expected ruleset to be enabled with configured rules")
	}
}

func TestVolumeRuleset_Enabled_ReturnsConfiguredRules(t *testing.T) {
	maxLoc := 1000
	maxLloc := 600
	maxLocByMethod := 30
	cfg := &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Volume: &configuration.ConfigurationVolumeRules{
				Loc:           &maxLoc,
				Lloc:          &maxLloc,
				LocByMethod:   &maxLocByMethod,
			},
		},
	}
	ruleset := &volumeRuleset{cfg: cfg}
	
	enabled := ruleset.Enabled()
	if len(enabled) != 3 {
		t.Fatalf("expected 3 enabled rules, got %d", len(enabled))
	}

	ruleNames := make(map[string]bool)
	for _, rule := range enabled {
		ruleNames[rule.Name()] = true
	}

	expectedRules := []string{"max_loc", "max_lloc", "max_loc_by_method"}
	for _, name := range expectedRules {
		if !ruleNames[name] {
			t.Errorf("missing expected rule: %s", name)
		}
	}
}
