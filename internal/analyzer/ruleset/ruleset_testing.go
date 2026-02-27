package ruleset

import (
	"github.com/halleck45/ast-metrics/internal/configuration"
)

type testingRuleset struct {
	cfg *configuration.ConfigurationRequirements
}

func (t *testingRuleset) Category() string {
	return "testing"
}

func (t *testingRuleset) Description() string {
	return "Test quality metrics (traceability, isolation, god tests, orphans)"
}

// All returns an empty slice — testing rules are project-level, not file-level.
func (t *testingRuleset) All() []Rule {
	return []Rule{}
}

// Enabled returns an empty slice — testing rules are project-level, not file-level.
func (t *testingRuleset) Enabled() []Rule {
	return []Rule{}
}

func (t *testingRuleset) IsEnabled() bool {
	return len(t.EnabledProjectRules()) > 0
}

// AllProjectRules returns all project-level rules regardless of configuration.
func (t *testingRuleset) AllProjectRules() []ProjectRule {
	return []ProjectRule{
		NewTraceabilityRule(nil),
		NewIsolationRule(nil),
		NewGodTestRule(nil),
		NewOrphanCriticalRule(nil),
	}
}

// EnabledProjectRules returns project-level rules that are configured.
func (t *testingRuleset) EnabledProjectRules() []ProjectRule {
	var rules []ProjectRule
	if t == nil || t.cfg == nil || t.cfg.Rules == nil || t.cfg.Rules.Testing == nil {
		return rules
	}

	tc := t.cfg.Rules.Testing
	if tc.MinTraceability != nil {
		rules = append(rules, NewTraceabilityRule(tc.MinTraceability))
	}
	if tc.MinIsolationScore != nil {
		rules = append(rules, NewIsolationRule(tc.MinIsolationScore))
	}
	if tc.MaxGodTestFanOut != nil {
		rules = append(rules, NewGodTestRule(tc.MaxGodTestFanOut))
	}
	if tc.MaxOrphanWeight != nil {
		rules = append(rules, NewOrphanCriticalRule(tc.MaxOrphanWeight))
	}

	return rules
}
