package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/stretchr/testify/assert"
)

func parseGoSource(t *testing.T, filename, source string) map[string]bool {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	r := &GolangRunner{}
	file, err := r.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	deps := map[string]bool{}
	for _, d := range engine.GetDependenciesInFile(file) {
		deps[d.GetClassName()] = true
	}
	return deps
}

func TestTestSymbolRefs_SamePackage(t *testing.T) {
	deps := parseGoSource(t, "counter_test.go", `package counter

import "testing"

type mockClock struct{}

func makeCounter() *Counter {
	return NewCounter()
}

func TestCounter_Add(t *testing.T) {
	c := makeCounter()
	c.Add(2)
	var other Counter
	other.Add(len("x"))
}
`)

	// struct references are qualified with the package name, like the visitor does
	assert.True(t, deps[`counter\Counter`], "expected struct reference counter\\Counter, got %v", deps)
	// top-level function calls keep their short name, which is how they are indexed
	assert.True(t, deps["NewCounter"], "expected function reference NewCounter, got %v", deps)

	// symbols declared by the test file itself are not production references
	assert.False(t, deps[`counter\mockClock`])
	assert.False(t, deps["makeCounter"])
	assert.False(t, deps["TestCounter_Add"])
	// builtins are ignored
	assert.False(t, deps["len"])
	assert.False(t, deps[`counter\int`])
	// method calls on variables say nothing about which type is tested
	assert.False(t, deps["Add"])
}

func TestTestSymbolRefs_ExternalTestPackage(t *testing.T) {
	deps := parseGoSource(t, "counter_test.go", `package counter_test

import (
	"testing"

	"example.com/mod/counter"
)

func TestCounter(t *testing.T) {
	c := counter.Counter{}
	_ = c
	counter.NewCounter()
}
`)

	assert.True(t, deps[`counter\Counter`], "expected qualified type reference, got %v", deps)
	assert.True(t, deps[`counter\NewCounter`], "expected qualified call reference, got %v", deps)
	// selectors on variables are not package references
	assert.False(t, deps[`t\Run`])
}

func TestTestSymbolRefs_NotAttachedToProdFiles(t *testing.T) {
	deps := parseGoSource(t, "counter.go", `package counter

type Counter struct{ value int }

func NewCounter() *Counter {
	return &Counter{}
}
`)

	assert.False(t, deps[`counter\Counter`])
	assert.False(t, deps["NewCounter"])
}
