package php

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestPhpClassOperandsFromProperties(t *testing.T) {
	phpSource := `<?php
class A {
   private int $a;
   public $c;
}`

	result, err := engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)
	assert.Empty(t, result.Errors)

	// Ensure classes
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	// Expect 2 direct operands from properties: $a and $c
	if assert.Equal(t, 2, len(class1.Operands), "Class should have 2 operands from direct attributes") {
  assert.Equal(t, "a", class1.Operands[0].Name)
		assert.Equal(t, "c", class1.Operands[1].Name)
	}
}
