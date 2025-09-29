package rust

import (
	"testing"
)

func TestNewTreeSitterAdapter(t *testing.T) {
	src := []byte("fn main() {}")
	adapter := NewTreeSitterAdapter(src)
	
	if adapter == nil {
		t.Error("expected non-nil adapter")
	}
	if string(adapter.src) != "fn main() {}" {
		t.Errorf("expected source to be set, got %s", string(adapter.src))
	}
}

func TestTreeSitterAdapter_SetSource(t *testing.T) {
	adapter := &TreeSitterAdapter{}
	src := []byte("fn test() {}")
	
	adapter.SetSource(src)
	
	if string(adapter.src) != "fn test() {}" {
		t.Errorf("expected source 'fn test() {}', got %s", string(adapter.src))
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
	// Can't create a real node without parsing, so test with nil
	name := adapter.NodeName(nil)
	
	if name != "" {
		t.Errorf("expected empty name for nil source, got %s", name)
	}
}
