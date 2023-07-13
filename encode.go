package csvlib

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
)

var (
	textMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	csvMarshaler  = reflect.TypeOf((*CSVMarshaler)(nil)).Elem()
)

func getEncodeFunc(typ reflect.Type) (EncodeFunc, error) {
	if typ.Implements(csvMarshaler) {
		return encodeCSVMarshaler, nil
	}
	if reflect.PtrTo(typ).Implements(csvMarshaler) {
		return encodePtrCSVMarshaler, nil
	}
	if typ.Implements(textMarshaler) {
		return encodeTextMarshaler, nil
	}
	if reflect.PtrTo(typ).Implements(textMarshaler) {
		return encodePtrTextMarshaler, nil
	}
	return getEncodeFuncBaseType(typ)
}

func getEncodeFuncBaseType(typ reflect.Type) (EncodeFunc, error) {
	typeIsPtr := false
	if typ.Kind() == reflect.Pointer {
		typeIsPtr = true
		typ = typ.Elem()
	}
	switch typ.Kind() { // nolint: exhaustive
	case reflect.String:
		if typeIsPtr {
			return encodePtrStr, nil
		}
		return encodeStr, nil
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if typeIsPtr {
			return encodePtrInt, nil
		}
		return encodeInt, nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		if typeIsPtr {
			return encodePtrUint, nil
		}
		return encodeUint, nil
	case reflect.Bool:
		if typeIsPtr {
			return encodePtrBool, nil
		}
		return encodeBool, nil
	case reflect.Float32, reflect.Float64:
		if typeIsPtr {
			return encodePtrFloatFunc(typ.Bits()), nil
		}
		return encodeFloatFunc(typ.Bits()), nil
	case reflect.Interface:
		if typeIsPtr {
			return encodePtrInterface, nil
		}
		return encodeInterface, nil
	default:
		return nil, fmt.Errorf("%w: %v", ErrTypeUnsupported, typ.Kind())
	}
}

func encodeCSVMarshaler(v reflect.Value, _ bool) (string, error) {
	if !v.IsValid() {
		return "", nil
	}
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return "", nil
	}
	b, err := v.Interface().(CSVMarshaler).MarshalCSV()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncodeValueType, v.Type())
	}
	return string(b), nil
}

func encodePtrCSVMarshaler(v reflect.Value, omitempty bool) (string, error) {
	if v.CanAddr() {
		return encodeCSVMarshaler(v.Addr(), omitempty)
	}
	// Fallback to process the value dynamically
	encodeFn, err := getEncodeFuncBaseType(v.Type())
	if err != nil {
		return "", err
	}
	return encodeFn(v, omitempty)
}

func encodeTextMarshaler(v reflect.Value, _ bool) (string, error) {
	if !v.IsValid() {
		return "", nil
	}
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return "", nil
	}
	b, err := v.Interface().(encoding.TextMarshaler).MarshalText()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncodeValueType, v.Type())
	}
	return string(b), nil
}

func encodePtrTextMarshaler(v reflect.Value, omitempty bool) (string, error) {
	if v.CanAddr() {
		return encodeTextMarshaler(v.Addr(), omitempty)
	}
	// Fallback to process the value dynamically
	encodeFn, err := getEncodeFuncBaseType(v.Type())
	if err != nil {
		return "", err
	}
	return encodeFn(v, omitempty)
}

func encodeStr(v reflect.Value, _ bool) (string, error) {
	return v.String(), nil
}

func encodePtrStr(v reflect.Value, _ bool) (string, error) {
	v = v.Elem()
	if !v.IsValid() {
		return "", nil
	}
	return v.String(), nil
}

func encodeBool(v reflect.Value, omitempty bool) (string, error) {
	t := v.Bool()
	if !t && omitempty {
		return "", nil
	}
	return strconv.FormatBool(t), nil
}

func encodePtrBool(v reflect.Value, omitempty bool) (string, error) {
	v = v.Elem()
	if !v.IsValid() {
		return "", nil
	}
	return encodeBool(v, omitempty)
}

func encodeInt(v reflect.Value, omitempty bool) (string, error) {
	n := v.Int()
	if n == 0 && omitempty {
		return "", nil
	}
	return strconv.FormatInt(n, 10), nil
}

func encodePtrInt(v reflect.Value, omitempty bool) (string, error) {
	v = v.Elem()
	if !v.IsValid() {
		return "", nil
	}
	return encodeInt(v, omitempty)
}

func encodeUint(v reflect.Value, omitempty bool) (string, error) {
	n := v.Uint()
	if n == 0 && omitempty {
		return "", nil
	}
	return strconv.FormatUint(n, 10), nil
}

func encodePtrUint(v reflect.Value, omitempty bool) (string, error) {
	v = v.Elem()
	if !v.IsValid() {
		return "", nil
	}
	return encodeUint(v, omitempty)
}

func encodeFloat(v reflect.Value, omitempty bool, bits int) (string, error) {
	f := v.Float()
	if f == 0 && omitempty {
		return "", nil
	}
	return strconv.FormatFloat(f, 'f', -1, bits), nil
}

func encodePtrFloat(v reflect.Value, omitempty bool, bits int) (string, error) {
	v = v.Elem()
	if !v.IsValid() {
		return "", nil
	}
	return encodeFloat(v, omitempty, bits)
}

func encodeFloatFunc(bits int) EncodeFunc {
	return func(v reflect.Value, omitempty bool) (string, error) {
		return encodeFloat(v, omitempty, bits)
	}
}

func encodePtrFloatFunc(bits int) EncodeFunc {
	return func(v reflect.Value, omitempty bool) (string, error) {
		return encodePtrFloat(v, omitempty, bits)
	}
}

func encodeInterface(v reflect.Value, omitempty bool) (string, error) {
	val := v.Elem()
	if !val.IsValid() {
		return "", nil
	}
	encodeFn, err := getEncodeFunc(val.Type())
	if err != nil {
		return "", err
	}
	return encodeFn(val, omitempty)
}

func encodePtrInterface(v reflect.Value, omitempty bool) (string, error) {
	val := v.Elem()
	if !val.IsValid() {
		return "", nil
	}
	val = val.Elem()
	if !val.IsValid() {
		return "", nil
	}
	encodeFn, err := getEncodeFunc(val.Type())
	if err != nil {
		return "", err
	}
	return encodeFn(val, omitempty)
}
