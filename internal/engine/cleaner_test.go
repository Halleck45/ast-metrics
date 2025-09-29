package engine

import (
	"math"
	"testing"
)

func TestCleanVal_NonPointer(t *testing.T) {
	val := 42.0
	err := CleanVal(val)
	if err == nil {
		t.Error("expected error for non-pointer value")
	}
}

func TestCleanVal_NaNFloat(t *testing.T) {
	val := math.NaN()
	err := CleanVal(&val)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if val != 0.0 {
		t.Errorf("expected NaN to be cleaned to 0.0, got %f", val)
	}
}

func TestCleanVal_InfFloat(t *testing.T) {
	val := math.Inf(1)
	err := CleanVal(&val)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if val != 0.0 {
		t.Errorf("expected Inf to be cleaned to 0.0, got %f", val)
	}
}

func TestCleanVal_Struct(t *testing.T) {
	type TestStruct struct {
		Value float64
	}
	
	s := TestStruct{Value: math.NaN()}
	err := CleanVal(&s)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if s.Value != 0.0 {
		t.Errorf("expected struct field to be cleaned to 0.0, got %f", s.Value)
	}
}

func TestCleanVal_Slice(t *testing.T) {
	slice := []float64{1.0, math.NaN(), 3.0}
	err := CleanVal(&slice)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if slice[1] != 0.0 {
		t.Errorf("expected slice element to be cleaned to 0.0, got %f", slice[1])
	}
}
