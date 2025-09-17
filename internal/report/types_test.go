package report

import "testing"

func TestReport_Structure(t *testing.T) {
	r := report{
		NbFiles:     10,
		NbFunctions: 50,
		NbClasses:   5,
		Loc:         1000,
	}

	if r.NbFiles != 10 {
		t.Errorf("expected NbFiles 10, got %d", r.NbFiles)
	}
	if r.NbFunctions != 50 {
		t.Errorf("expected NbFunctions 50, got %d", r.NbFunctions)
	}
}

func TestContributor_Structure(t *testing.T) {
	c := contributor{
		Name:  "John Doe",
		Count: 25,
	}

	if c.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %s", c.Name)
	}
	if c.Count != 25 {
		t.Errorf("expected count 25, got %d", c.Count)
	}
}

func TestFile_Structure(t *testing.T) {
	f := file{
		Path: "/test/file.go",
		Complexity: complexity{Cyclomatic: 5},
		Volume:     volume{Loc: 100},
	}

	if f.Path != "/test/file.go" {
		t.Errorf("expected path '/test/file.go', got %s", f.Path)
	}
	if f.Complexity.Cyclomatic != 5 {
		t.Errorf("expected cyclomatic 5, got %d", f.Complexity.Cyclomatic)
	}
}
