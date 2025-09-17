package issue

import "testing"

func TestSeverity_Constants(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityUnknown, "unknown"},
		{SeverityLow, "low"},
		{SeverityMedium, "medium"},
		{SeverityHigh, "high"},
	}

	for _, test := range tests {
		if string(test.severity) != test.expected {
			t.Errorf("expected %s, got %s", test.expected, string(test.severity))
		}
	}
}

func TestRequirementError_Structure(t *testing.T) {
	err := RequirementError{
		Message:  "Test error message",
		Code:     "test_code",
		Severity: SeverityHigh,
	}

	if err.Message != "Test error message" {
		t.Errorf("expected message 'Test error message', got %s", err.Message)
	}
	if err.Code != "test_code" {
		t.Errorf("expected code 'test_code', got %s", err.Code)
	}
	if err.Severity != SeverityHigh {
		t.Errorf("expected severity high, got %s", err.Severity)
	}
}

func TestRequirementError_ZeroValue(t *testing.T) {
	var err RequirementError

	if err.Message != "" {
		t.Errorf("expected empty message, got %s", err.Message)
	}
	if err.Code != "" {
		t.Errorf("expected empty code, got %s", err.Code)
	}
	if err.Severity != "" {
		t.Errorf("expected empty severity, got %s", err.Severity)
	}
}
