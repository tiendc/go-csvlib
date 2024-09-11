package csvlib

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
)

var (
	textUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	csvUnmarshaler  = reflect.TypeOf((*CSVUnmarshaler)(nil)).Elem()
)

func getDecodeFunc(typ reflect.Type) (DecodeFunc, error) {
	if typ.Implements(csvUnmarshaler) {
		return decodeCSVUnmarshaler, nil
	}
	if reflect.PointerTo(typ).Implements(csvUnmarshaler) {
		return decodePtrCSVUnmarshaler, nil
	}
	if typ.Implements(textUnmarshaler) {
		return decodeTextUnmarshaler, nil
	}
	if reflect.PointerTo(typ).Implements(textUnmarshaler) {
		return decodePtrTextUnmarshaler, nil
	}
	return getDecodeFuncBaseType(typ)
}

func getDecodeFuncBaseType(typ reflect.Type) (DecodeFunc, error) {
	typeIsPtr := false
	if typ.Kind() == reflect.Pointer {
		typeIsPtr = true
		typ = typ.Elem()
	}
	switch typ.Kind() { // nolint: exhaustive
	case reflect.String:
		if typeIsPtr {
			return decodePtrStr, nil
		}
		return decodeStr, nil
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if typeIsPtr {
			return decodePtrIntFunc(typ.Bits()), nil
		}
		return decodeIntFunc(typ.Bits()), nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		if typeIsPtr {
			return decodePtrUintFunc(typ.Bits()), nil
		}
		return decodeUintFunc(typ.Bits()), nil
	case reflect.Bool:
		if typeIsPtr {
			return decodePtrBool, nil
		}
		return decodeBool, nil
	case reflect.Float32, reflect.Float64:
		if typeIsPtr {
			return decodePtrFloatFunc(typ.Bits()), nil
		}
		return decodeFloatFunc(typ.Bits()), nil
	case reflect.Interface:
		if typeIsPtr {
			return decodePtrInterface, nil
		}
		return decodeInterface, nil
	default:
		return nil, fmt.Errorf("%w: %v", ErrTypeUnsupported, typ.Kind())
	}
}

func decodeTextUnmarshaler(s string, v reflect.Value) error {
	initAndIndirectValue(v)
	return v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s)) // nolint: forcetypeassert
}

func decodePtrTextUnmarshaler(s string, v reflect.Value) error {
	if v.CanAddr() {
		return decodeTextUnmarshaler(s, v.Addr())
	}
	// Fallback to process the value dynamically
	decodeFn, err := getDecodeFuncBaseType(v.Type())
	if err != nil {
		return err
	}
	return decodeFn(s, v)
}

func decodeCSVUnmarshaler(s string, v reflect.Value) error {
	initAndIndirectValue(v)
	return v.Interface().(CSVUnmarshaler).UnmarshalCSV([]byte(s)) // nolint: forcetypeassert
}

func decodePtrCSVUnmarshaler(s string, v reflect.Value) error {
	if v.CanAddr() {
		return decodeCSVUnmarshaler(s, v.Addr())
	}
	// Fallback to process the value dynamically
	decodeFn, err := getDecodeFuncBaseType(v.Type())
	if err != nil {
		return err
	}
	return decodeFn(s, v)
}

func decodeStr(s string, v reflect.Value) error {
	v.SetString(s)
	return nil
}

func decodePtrStr(s string, v reflect.Value) error {
	return decodeStr(s, initAndIndirectValue(v))
}

func decodeBool(s string, v reflect.Value) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("%w: %v (%s)", ErrDecodeValueType, v.Type(), s)
	}
	v.SetBool(b)
	return nil
}

func decodePtrBool(s string, v reflect.Value) error {
	return decodeBool(s, initAndIndirectValue(v))
}

func decodeInt(s string, v reflect.Value, bits int) error {
	n, err := strconv.ParseInt(s, 10, bits)
	if err != nil {
		return fmt.Errorf("%w: %v (%s)", ErrDecodeValueType, v.Type(), s)
	}
	v.SetInt(n)
	return nil
}

func decodeIntFunc(bits int) DecodeFunc {
	return func(s string, v reflect.Value) error {
		return decodeInt(s, v, bits)
	}
}

func decodePtrIntFunc(bits int) DecodeFunc {
	return func(s string, v reflect.Value) error {
		return decodeInt(s, initAndIndirectValue(v), bits)
	}
}

func decodeUint(s string, v reflect.Value, bits int) error {
	n, err := strconv.ParseUint(s, 10, bits)
	if err != nil {
		return fmt.Errorf("%w: %v (%s)", ErrDecodeValueType, v.Type(), s)
	}
	v.SetUint(n)
	return nil
}

func decodeUintFunc(bits int) DecodeFunc {
	return func(s string, v reflect.Value) error {
		return decodeUint(s, v, bits)
	}
}

func decodePtrUintFunc(bits int) DecodeFunc {
	return func(s string, v reflect.Value) error {
		return decodeUint(s, initAndIndirectValue(v), bits)
	}
}

func decodeFloat(s string, v reflect.Value, bits int) error {
	n, err := strconv.ParseFloat(s, bits)
	if err != nil {
		return fmt.Errorf("%w: %v (%s)", ErrDecodeValueType, v.Type(), s)
	}
	v.SetFloat(n)
	return nil
}

func decodeFloatFunc(bits int) DecodeFunc {
	return func(s string, v reflect.Value) error {
		return decodeFloat(s, v, bits)
	}
}

func decodePtrFloatFunc(bits int) DecodeFunc {
	return func(s string, v reflect.Value) error {
		return decodeFloat(s, initAndIndirectValue(v), bits)
	}
}

func decodeInterface(s string, v reflect.Value) error {
	v.Set(reflect.ValueOf(s))
	return nil
}

func decodePtrInterface(s string, v reflect.Value) error {
	initAndIndirectValue(v).Set(reflect.ValueOf(s))
	return nil
}
