package python

import (
	"testing"
)

func TestNewTreeSitterAdapter(t *testing.T) {
	src := []byte("def test(): pass")
	adapter := NewTreeSitterAdapter(src)
	
	if adapter == nil {
		t.Error("expected non-nil adapter")
	}
	if string(adapter.src) != "def test(): pass" {
		t.Errorf("expected source to be set, got %s", string(adapter.src))
	}
}

func TestTreeSitterAdapter_SetSource(t *testing.T) {
	adapter := &TreeSitterAdapter{}
	src := []byte("def main(): return")
	
	adapter.SetSource(src)
	
	if string(adapter.src) != "def main(): return" {
		t.Errorf("expected source 'def main(): return', got %s", string(adapter.src))
	}
}

func TestTreeSitterAdapter_Language(t *testing.T) {
	adapter := &TreeSitterAdapter{}
	lang := adapter.Language()
	
	if lang == nil {
		t.Error("expected non-nil language")
	}
}

func TestTreeSitterAdapter_NodeName_NilNode(t *testing.T) {
	adapter := &TreeSitterAdapter{src: []byte("test")}
	name := adapter.NodeName(nil)
	
	if name != "" {
		t.Errorf("expected empty name for nil node, got %s", name)
	}
}

func TestTreeSitterAdapter_NodeName_NilSource(t *testing.T) {
	adapter := &TreeSitterAdapter{}
	name := adapter.NodeName(nil)
	
	if name != "" {
		t.Errorf("expected empty name for nil source, got %s", name)
	}
}

func TestTreeSitterAdapter_NodeBody_NilNode(t *testing.T) {
	adapter := &TreeSitterAdapter{}
	body := adapter.NodeBody(nil)
	
	if body != nil {
		t.Error("expected nil body for nil node")
	}
}

func TestTreeSitterAdapter_CountComments(t *testing.T) {
	adapter := NewTreeSitterAdapter(nil)

	lines := []string{
		"# module comment",
		"#!shebang-like comment",
		"def divide(a, b):",
		"    # inner comment",
		"    return a // b  # floor division, line has code: not counted",
		"",
	}

	cnt := adapter.CountComments(lines, 1, len(lines))
	if cnt != 3 {
		t.Errorf("expected 3 comment lines, got %d", cnt)
	}

	// "//" is floor division in Python, never a comment
	floorDiv := []string{"x = a // b"}
	if got := adapter.CountComments(floorDiv, 1, 1); got != 0 {
		t.Errorf("expected 0 comment lines for floor division, got %d", got)
	}
}

func TestTreeSitterAdapter_ExtractOperatorsOperands(t *testing.T) {
	src := []byte(`def divide(a, b):
    q = a // b
    return q + 1
`)
	adapter := NewTreeSitterAdapter(src)
	ops, operands := adapter.ExtractOperatorsOperands(src, 1, 3)

	// the "," of the parameter list, =, // and + ("//" is an operator here, not
	// a comment), and the "return"
	if len(ops) != 5 {
		t.Fatalf("expected 5 operators, got %d: %v", len(ops), ops)
	}
	// divide, a, b (signature), q, a, b, q, 1
	if len(operands) != 8 {
		t.Fatalf("expected 8 operands, got %d: %v", len(operands), operands)
	}

	// a "//" inside a string or comment is not an operator
	src2 := []byte(`def f():
    s = "no // here"  # nor // here
    return s
`)
	adapter2 := NewTreeSitterAdapter(src2)
	ops2, _ := adapter2.ExtractOperatorsOperands(src2, 1, 3)
	if len(ops2) != 2 { // only the assignment and the "return"
		t.Fatalf("expected 2 operators, got %d: %v", len(ops2), ops2)
	}
}
