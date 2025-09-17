package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestGolangRuleset_Category(t *testing.T) {
	ruleset := &golangRuleset{}
	if ruleset.Category() != "golang" {
		t.Errorf("expected 'golang', got %s", ruleset.Category())
	}
}

func TestGolangRuleset_Description(t *testing.T) {
	ruleset := &golangRuleset{}
	expected := "Golang-specific best practices and API hygiene"
	if ruleset.Description() != expected {
		t.Errorf("expected '%s', got %s", expected, ruleset.Description())
	}
}

func TestGolangRuleset_IsEnabled_EmptyConfig(t *testing.T) {
	ruleset := &golangRuleset{cfg: &configuration.ConfigurationRequirements{}}
	if ruleset.IsEnabled() {
		t.Error("expected ruleset to be disabled with empty config")
	}
}

func TestGolangRuleset_IsEnabled_WithRules(t *testing.T) {
	maxNesting := 4
	cfg := &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Golang: &configuration.ConfigurationGolangRuleset{
				MaxNesting: &maxNesting,
			},
		},
	}
	ruleset := &golangRuleset{cfg: cfg}
	
	if !ruleset.IsEnabled() {
		t.Error("expected ruleset to be enabled with configured rules")
	}
}

func TestGolangRuleset_Enabled_ReturnsConfiguredRules(t *testing.T) {
	maxNesting := 4
	maxFileSize := 1000
	slicePrealloc := true
	cfg := &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Golang: &configuration.ConfigurationGolangRuleset{
				MaxNesting:    &maxNesting,
				MaxFileSize:   &maxFileSize,
				SlicePrealloc: &slicePrealloc,
			},
		},
	}
	ruleset := &golangRuleset{cfg: cfg}
	
	enabled := ruleset.Enabled()
	if len(enabled) != 3 {
		t.Fatalf("expected 3 enabled rules, got %d", len(enabled))
	}

	ruleNames := make(map[string]bool)
	for _, rule := range enabled {
		ruleNames[rule.Name()] = true
	}

	expectedRules := []string{"max_nesting_depth", "max_file_size", "slice_prealloc"}
	for _, name := range expectedRules {
		if !ruleNames[name] {
			t.Errorf("missing expected rule: %s", name)
		}
	}
}

func TestGolangRuleset_All_ReturnsAllPossibleRules(t *testing.T) {
	ruleset := &golangRuleset{}
	all := ruleset.All()
	
	// Check that we have all expected golang rules
	if len(all) < 7 { // At least the main golang rules
		t.Fatalf("expected at least 7 total rules, got %d", len(all))
	}

	ruleNames := make(map[string]bool)
	for _, rule := range all {
		ruleNames[rule.Name()] = true
	}

	expectedRules := []string{
		"no_package_name_in_method", "max_nesting_depth", "max_file_size",
		"max_files_per_package", "slice_prealloc", 
		"no_context_missing", "no_context_ignored",
	}
	for _, name := range expectedRules {
		if !ruleNames[name] {
			t.Errorf("missing expected rule: %s", name)
		}
	}
}
