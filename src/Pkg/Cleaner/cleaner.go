package Cleaner

import (
	"errors"
	"math"
	"reflect"
)

var (
	defaultfloat64 float64 = 0
	defaultFloat64 float64 = 0
)

// The CleanVal removes all NaN values from any value
// and sets them to the default float64 value, which is 0.
// For float64 values, it also sets them to 0.
//
// This function accepts a pointer because it needs
// to modify the provided value.
func CleanVal(val interface{}) error {
	v := reflect.ValueOf(val)

	if v.Kind() != reflect.Pointer {
		return errors.New("value must be a pointer")
	}

	clean(v)

	return nil
}

func clean(v reflect.Value) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		cleanStruct(v)
	case reflect.Slice:
		cleanSlice(v)
	default:
		cleanField(v)
	}
}

func cleanStruct(v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		clean(field)
	}
}

func cleanSlice(v reflect.Value) {
	for i := 0; i < v.Len(); i++ {
		v := v.Index(i)
		clean(v)
	}
}

func cleanField(field reflect.Value) {
	switch field.Kind() {
	case reflect.Float64:
		f := field.Float()
		isInvalidAndCanSet := field.CanSet() && (math.IsNaN(f) || math.IsInf(f, 0))
		if !isInvalidAndCanSet {
			return
		}

		field.Set(reflect.ValueOf(defaultFloat64))
	}
}
