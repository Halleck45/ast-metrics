package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestArchitectureRuleset_Category(t *testing.T) {
	ruleset := &architectureRuleset{}
	if ruleset.Category() != "architecture" {
		t.Errorf("expected 'architecture', got %s", ruleset.Category())
	}
}

func TestArchitectureRuleset_Description(t *testing.T) {
	ruleset := &architectureRuleset{}
	expected := "Architecture-related constraints (e.g., coupling)"
	if ruleset.Description() != expected {
		t.Errorf("expected '%s', got %s", expected, ruleset.Description())
	}
}

func TestArchitectureRuleset_IsEnabled_EmptyConfig(t *testing.T) {
	ruleset := &architectureRuleset{cfg: &configuration.ConfigurationRequirements{}}
	if ruleset.IsEnabled() {
		t.Error("expected ruleset to be disabled with empty config")
	}
}

func TestArchitectureRuleset_IsEnabled_WithRules(t *testing.T) {
	maxCoupling := 10
	cfg := &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Architecture: &configuration.ConfigurationArchitectureRules{
				AfferentCoupling: &maxCoupling,
			},
		},
	}
	ruleset := &architectureRuleset{cfg: cfg}
	
	if !ruleset.IsEnabled() {
		t.Error("expected ruleset to be enabled with configured rules")
	}
}

func TestArchitectureRuleset_Enabled_ReturnsConfiguredRules(t *testing.T) {
	maxAfferent := 10
	maxEfferent := 15
	minMaintainability := 70
	cfg := &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Architecture: &configuration.ConfigurationArchitectureRules{
				AfferentCoupling:  &maxAfferent,
				EfferentCoupling:  &maxEfferent,
				Maintainability:   &minMaintainability,
			},
		},
	}
	ruleset := &architectureRuleset{cfg: cfg}
	
	enabled := ruleset.Enabled()
	if len(enabled) != 3 {
		t.Fatalf("expected 3 enabled rules, got %d", len(enabled))
	}

	ruleNames := make(map[string]bool)
	for _, rule := range enabled {
		ruleNames[rule.Name()] = true
	}

	expectedRules := []string{"afferent_coupling", "efferent_coupling", "maintainability"}
	for _, name := range expectedRules {
		if !ruleNames[name] {
			t.Errorf("missing expected rule: %s", name)
		}
	}
}

func TestArchitectureRuleset_All_ReturnsAllPossibleRules(t *testing.T) {
	ruleset := &architectureRuleset{}
	all := ruleset.All()
	
	if len(all) != 7 {
		t.Fatalf("expected 7 total rules, got %d", len(all))
	}

	ruleNames := make(map[string]bool)
	for _, rule := range all {
		ruleNames[rule.Name()] = true
	}

	expectedRules := []string{
		"coupling", "afferent_coupling", "efferent_coupling", 
		"maintainability", "no_circular_dependencies", 
		"max_responsibilities", "no_god_class",
	}
	for _, name := range expectedRules {
		if !ruleNames[name] {
			t.Errorf("missing expected rule: %s", name)
		}
	}
}
