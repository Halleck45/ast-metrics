package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestComplexityRuleset_Category(t *testing.T) {
	ruleset := &complexityRuleset{}
	if ruleset.Category() != "complexity" {
		t.Errorf("expected 'complexity', got %s", ruleset.Category())
	}
}

func TestComplexityRuleset_Description(t *testing.T) {
	ruleset := &complexityRuleset{}
	expected := "Complexity metrics (e.g., cyclomatic complexity)"
	if ruleset.Description() != expected {
		t.Errorf("expected '%s', got %s", expected, ruleset.Description())
	}
}

func TestComplexityRuleset_IsEnabled_EmptyConfig(t *testing.T) {
	ruleset := &complexityRuleset{cfg: &configuration.ConfigurationRequirements{}}
	if ruleset.IsEnabled() {
		t.Error("expected ruleset to be disabled with empty config")
	}
}

func TestComplexityRuleset_IsEnabled_WithRules(t *testing.T) {
	maxCyclomatic := 10
	cfg := &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Complexity: &configuration.ConfigurationComplexityRules{
				Cyclomatic: &maxCyclomatic,
			},
		},
	}
	ruleset := &complexityRuleset{cfg: cfg}
	
	if !ruleset.IsEnabled() {
		t.Error("expected ruleset to be enabled with configured rules")
	}
}

func TestComplexityRuleset_Enabled_ReturnsConfiguredRules(t *testing.T) {
	maxCyclomatic := 10
	cfg := &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Complexity: &configuration.ConfigurationComplexityRules{
				Cyclomatic: &maxCyclomatic,
			},
		},
	}
	ruleset := &complexityRuleset{cfg: cfg}
	
	enabled := ruleset.Enabled()
	if len(enabled) != 1 {
		t.Fatalf("expected 1 enabled rule, got %d", len(enabled))
	}

	ruleNames := make(map[string]bool)
	for _, rule := range enabled {
		ruleNames[rule.Name()] = true
	}

	expectedRules := []string{"cyclomatic_complexity"}
	for _, name := range expectedRules {
		if !ruleNames[name] {
			t.Errorf("missing expected rule: %s", name)
		}
	}
}

func TestComplexityRuleset_All_ReturnsAllPossibleRules(t *testing.T) {
	ruleset := &complexityRuleset{}
	all := ruleset.All()
	
	if len(all) != 1 {
		t.Fatalf("expected 1 total rule, got %d", len(all))
	}

	ruleNames := make(map[string]bool)
	for _, rule := range all {
		ruleNames[rule.Name()] = true
	}

	expectedRules := []string{"cyclomatic_complexity"}
	for _, name := range expectedRules {
		if !ruleNames[name] {
			t.Errorf("missing expected rule: %s", name)
		}
	}
}
