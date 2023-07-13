package csvlib

import (
	"reflect"
)

func isKindOrPtrOf(t reflect.Type, kinds ...reflect.Kind) bool {
	k := t.Kind()
	if k == reflect.Pointer {
		k = t.Elem().Kind()
	}
	for _, kk := range kinds {
		if k == kk {
			return true
		}
	}
	return false
}

func indirectType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		return t.Elem()
	}
	return t
}

func indirectValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		return v.Elem()
	}
	return v
}

func initAndIndirectValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			// NOTE: v.CanSet must return true in order to call v.Set
			v.Set(reflect.New(v.Type().Elem()))
		}
		return v.Elem()
	}
	return v
}
