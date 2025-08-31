package Php

import (
	"strings"
	"testing"

	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/stretchr/testify/assert"
)

// Ensure that PHP pseudo-class keywords self and static are not reported as external dependencies
func TestPhpSelfAndStaticNotExternalDependencies(t *testing.T) {
	phpSource := `<?php

namespace N1;

class A {
	public static function foo() {}
	public static function bar() {}
	public function x() {
		self::foo();
		static::bar();
	}
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)
	assert.Empty(t, result.Errors)

	// one class A
	assert.Equal(t, 1, len(result.Stmts.StmtClass))
	class1 := result.Stmts.StmtClass[0]
	deps := class1.Stmts.StmtExternalDependencies
	for _, d := range deps {
		low := d.ClassName
		if len(low) > 0 {
			// Case-insensitive check of last segment
			last := low
			if idx := lastLastIndex(last, "\\"); idx >= 0 {
				last = last[idx+1:]
			}
			if eqIgnoreCase(last, "self") || eqIgnoreCase(last, "static") {
				assert.Failf(t, "self/static should not be external dependencies", "found dependency: %s", d.ClassName)
			}
		}
	}
}

// helpers for the test file (avoid importing extra packages)
func eqIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		// quick check first; convert if necessary
	}
	return strings.EqualFold(a, b)
}

func lastLastIndex(s, sep string) int {
	return strings.LastIndex(s, sep)
}
