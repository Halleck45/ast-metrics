package Cleaner

import (
	"math"
	"reflect"
)

var (
	defaultFloat64 float64 = 0
	defaultFloat32 float32 = 0
)

// The CleanVal removes all NaN values from any value
// and sets them to the default float64 value, which is 0.
// For float32 values, it also sets them to 0.
//
// This function accepts a pointer because it needs
// to modify the provided value.
func CleanVal(val interface{}) {
	v := reflect.ValueOf(val)

	if v.Kind() != reflect.Pointer {
		panic("Val must be a pointer")
	}

	clean(v)
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
	case reflect.Float32, reflect.Float64:
		f := field.Float()
		isInvalidAndCanSet := (math.IsNaN(f) || math.IsInf(f, 0)) && field.CanSet()
		if !isInvalidAndCanSet {
			return
		}

		switch field.Kind() {
		case reflect.Float64:
			field.Set(reflect.ValueOf(defaultFloat64))
		case reflect.Float32:
			field.Set(reflect.ValueOf(defaultFloat32))
		}
	}
}
