package treesitter

import "testing"

func TestDecisionKind_Constants(t *testing.T) {
	if DecNone != 0 {
		t.Errorf("expected DecNone to be 0, got %d", DecNone)
	}
	if DecIf != 1 {
		t.Errorf("expected DecIf to be 1, got %d", DecIf)
	}
	if DecElif != 2 {
		t.Errorf("expected DecElif to be 2, got %d", DecElif)
	}
	if DecElse != 3 {
		t.Errorf("expected DecElse to be 3, got %d", DecElse)
	}
	if DecLoop != 4 {
		t.Errorf("expected DecLoop to be 4, got %d", DecLoop)
	}
	if DecSwitch != 5 {
		t.Errorf("expected DecSwitch to be 5, got %d", DecSwitch)
	}
	if DecCase != 6 {
		t.Errorf("expected DecCase to be 6, got %d", DecCase)
	}
}

func TestImportItem_Structure(t *testing.T) {
	item := ImportItem{
		Module: "pkg.sub",
		Name:   "Function",
	}

	if item.Module != "pkg.sub" {
		t.Errorf("expected module 'pkg.sub', got %s", item.Module)
	}
	if item.Name != "Function" {
		t.Errorf("expected name 'Function', got %s", item.Name)
	}
}

func TestImportItem_ZeroValue(t *testing.T) {
	var item ImportItem

	if item.Module != "" {
		t.Errorf("expected empty module, got %s", item.Module)
	}
	if item.Name != "" {
		t.Errorf("expected empty name, got %s", item.Name)
	}
}
