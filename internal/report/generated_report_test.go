package report

import "testing"

func TestGeneratedReport_Structure(t *testing.T) {
	report := GeneratedReport{
		Path:        "/test/report.html",
		Type:        "html",
		Description: "HTML Report",
		Icon:        "📊",
	}

	if report.Path != "/test/report.html" {
		t.Errorf("expected path '/test/report.html', got %s", report.Path)
	}
	if report.Type != "html" {
		t.Errorf("expected type 'html', got %s", report.Type)
	}
	if report.Description != "HTML Report" {
		t.Errorf("expected description 'HTML Report', got %s", report.Description)
	}
	if report.Icon != "📊" {
		t.Errorf("expected icon '📊', got %s", report.Icon)
	}
}

func TestGeneratedReport_ZeroValue(t *testing.T) {
	var report GeneratedReport

	if report.Path != "" {
		t.Errorf("expected empty path, got %s", report.Path)
	}
	if report.Type != "" {
		t.Errorf("expected empty type, got %s", report.Type)
	}
}
