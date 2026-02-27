package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
)

func TestTestingRuleset_IsEnabled_NoConfig(t *testing.T) {
	rs := &testingRuleset{cfg: &configuration.ConfigurationRequirements{}}
	if rs.IsEnabled() {
		t.Error("expected testing ruleset to be disabled with no config")
	}
}

func TestTestingRuleset_IsEnabled_WithConfig(t *testing.T) {
	min := 60
	rs := &testingRuleset{cfg: &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Testing: &configuration.ConfigurationTestingRules{
				MinTraceability: &min,
			},
		},
	}}
	if !rs.IsEnabled() {
		t.Error("expected testing ruleset to be enabled with min_traceability configured")
	}
}

func TestTestingRuleset_EnabledProjectRules_Count(t *testing.T) {
	min := 60
	minIso := 50
	maxFan := 5
	maxWeight := 20.0
	rs := &testingRuleset{cfg: &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Testing: &configuration.ConfigurationTestingRules{
				MinTraceability:   &min,
				MinIsolationScore: &minIso,
				MaxGodTestFanOut:  &maxFan,
				MaxOrphanWeight:   &maxWeight,
			},
		},
	}}

	rules := rs.EnabledProjectRules()
	if len(rules) != 4 {
		t.Fatalf("expected 4 project rules, got %d", len(rules))
	}
}

func TestTestingRuleset_Enabled_ReturnsEmpty(t *testing.T) {
	min := 60
	rs := &testingRuleset{cfg: &configuration.ConfigurationRequirements{
		Rules: &configuration.ConfigurationRequirementsRules{
			Testing: &configuration.ConfigurationTestingRules{
				MinTraceability: &min,
			},
		},
	}}

	// File-level rules should always be empty for testing ruleset
	if len(rs.Enabled()) != 0 {
		t.Error("expected Enabled() to return empty slice for testing ruleset")
	}
	if len(rs.All()) != 0 {
		t.Error("expected All() to return empty slice for testing ruleset")
	}
}
