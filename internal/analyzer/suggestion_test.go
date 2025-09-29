package analyzer

import "testing"

func TestSuggestion_Structure(t *testing.T) {
	suggestion := Suggestion{
		Summary:             "Reduce method complexity",
		Location:            "MyClass.complexMethod",
		Why:                 "Cyclomatic complexity is 15, exceeds threshold of 10",
		DetailedExplanation: "Consider breaking this method into smaller functions.",
	}

	if suggestion.Summary != "Reduce method complexity" {
		t.Errorf("expected summary 'Reduce method complexity', got %s", suggestion.Summary)
	}
	if suggestion.Location != "MyClass.complexMethod" {
		t.Errorf("expected location 'MyClass.complexMethod', got %s", suggestion.Location)
	}
	if suggestion.Why != "Cyclomatic complexity is 15, exceeds threshold of 10" {
		t.Errorf("expected why 'Cyclomatic complexity is 15, exceeds threshold of 10', got %s", suggestion.Why)
	}
}

func TestSuggestion_ZeroValue(t *testing.T) {
	var suggestion Suggestion

	if suggestion.Summary != "" {
		t.Errorf("expected empty summary, got %s", suggestion.Summary)
	}
	if suggestion.Location != "" {
		t.Errorf("expected empty location, got %s", suggestion.Location)
	}
}
