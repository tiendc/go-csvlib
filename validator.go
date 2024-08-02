package csvlib

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// ValidatorLT validates a value to be less than the given value
func ValidatorLT[T LTComparable](val T) ValidatorFunc {
	return func(v any) error {
		v1, ok := v.(T)
		if !ok {
			return errValidationConversion(v, v1)
		}
		if v1 < val {
			return nil
		}
		return ErrValidationLT
	}
}

// ValidatorLTE validates a value to be less than or equal to the given value
func ValidatorLTE[T LTComparable](val T) ValidatorFunc {
	return func(v any) error {
		v1, ok := v.(T)
		if !ok {
			return errValidationConversion(v, v1)
		}
		if v1 <= val {
			return nil
		}
		return ErrValidationLTE
	}
}

// ValidatorGT validates a value to be greater than the given value
func ValidatorGT[T LTComparable](val T) ValidatorFunc {
	return func(v any) error {
		v1, ok := v.(T)
		if !ok {
			return errValidationConversion(v, v1)
		}
		if v1 > val {
			return nil
		}
		return ErrValidationGT
	}
}

// ValidatorGTE validates a value to be greater than or equal to the given value
func ValidatorGTE[T LTComparable](val T) ValidatorFunc {
	return func(v any) error {
		v1, ok := v.(T)
		if !ok {
			return errValidationConversion(v, v1)
		}
		if v1 >= val {
			return nil
		}
		return ErrValidationGTE
	}
}

// ValidatorRange validates a value to be in the given range (min and max are inclusive)
func ValidatorRange[T LTComparable](min, max T) ValidatorFunc {
	return func(v any) error {
		v1, ok := v.(T)
		if !ok {
			return errValidationConversion(v, v1)
		}
		if min <= v1 && v1 <= max {
			return nil
		}
		return ErrValidationRange
	}
}

// ValidatorIN validates a value to be one of the specific values
func ValidatorIN[T LTComparable](vals ...T) ValidatorFunc {
	return func(v any) error {
		v1, ok := v.(T)
		if !ok {
			return errValidationConversion(v, v1)
		}
		for _, val := range vals {
			if v1 == val {
				return nil
			}
		}
		return ErrValidationIN
	}
}

// ValidatorStrLen validates a string to have length in the given range.
// Pass argument -1 to skip the equivalent validation.
func ValidatorStrLen[T StringEx](minLen, maxLen int, lenFuncs ...func(s string) int) ValidatorFunc {
	return func(v any) error {
		s, ok := v.(T)
		if !ok {
			return errValidationConversion(v, s)
		}
		lenFunc := utf8.RuneCountInString
		if len(lenFuncs) > 0 {
			lenFunc = lenFuncs[0]
		}
		length := lenFunc(*(*string)(unsafe.Pointer(&s)))
		if (minLen == -1 || minLen <= length) && (maxLen == -1 || length <= maxLen) {
			return nil
		}
		return ErrValidationStrLen
	}
}

// ValidatorStrPrefix validates a string to have prefix matching the given one
func ValidatorStrPrefix[T StringEx](prefix string) ValidatorFunc {
	return func(v any) error {
		s, ok := v.(T)
		if !ok {
			return errValidationConversion(v, s)
		}
		if strings.HasPrefix(*(*string)(unsafe.Pointer(&s)), prefix) {
			return nil
		}
		return ErrValidationStrPrefix
	}
}

// ValidatorStrSuffix validates a string to have suffix matching the given one
func ValidatorStrSuffix[T StringEx](suffix string) ValidatorFunc {
	return func(v any) error {
		s, ok := v.(T)
		if !ok {
			return errValidationConversion(v, s)
		}
		if strings.HasSuffix(*(*string)(unsafe.Pointer(&s)), suffix) {
			return nil
		}
		return ErrValidationStrSuffix
	}
}

func errValidationConversion[T any](v1 any, v2 T) error {
	return fmt.Errorf("%w: (%v -> %v)", ErrValidationConversion, reflect.TypeOf(v1), reflect.TypeOf(v2))
}
