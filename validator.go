package csvlib

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// ValidatorLT validate a value to be less than the given value
func ValidatorLT[T LTComparable](val T) ValidatorFunc {
	return func(v interface{}) error {
		v1, err := convertValue[T](v)
		if err != nil {
			return err
		}
		if v1 < val {
			return nil
		}
		return ErrValidationLT
	}
}

// ValidatorLTE validate a value to be less than or equal to the given value
func ValidatorLTE[T LTComparable](val T) ValidatorFunc {
	return func(v interface{}) error {
		v1, err := convertValue[T](v)
		if err != nil {
			return err
		}
		if v1 <= val {
			return nil
		}
		return ErrValidationLTE
	}
}

// ValidatorGT validate a value to be greater than the given value
func ValidatorGT[T LTComparable](val T) ValidatorFunc {
	return func(v interface{}) error {
		v1, err := convertValue[T](v)
		if err != nil {
			return err
		}
		if v1 > val {
			return nil
		}
		return ErrValidationGT
	}
}

// ValidatorGTE validate a value to be greater than or equal to the given value
func ValidatorGTE[T LTComparable](val T) ValidatorFunc {
	return func(v interface{}) error {
		v1, err := convertValue[T](v)
		if err != nil {
			return err
		}
		if v1 >= val {
			return nil
		}
		return ErrValidationGTE
	}
}

// ValidatorRange validate a value to be in the given range (min and max are inclusive)
func ValidatorRange[T LTComparable](min, max T) ValidatorFunc {
	return func(v interface{}) error {
		v1, err := convertValue[T](v)
		if err != nil {
			return err
		}
		if min <= v1 && v1 <= max {
			return nil
		}
		return ErrValidationRange
	}
}

// ValidatorIN validate a value to be one of the specific values
func ValidatorIN[T LTComparable](vals ...T) ValidatorFunc {
	return func(v interface{}) error {
		v1, err := convertValue[T](v)
		if err != nil {
			return err
		}
		for _, val := range vals {
			if v1 == val {
				return nil
			}
		}
		return ErrValidationIN
	}
}

// ValidatorStrLen validate a string to have length in the given range
// Pass argument -1 to skip the equivalent validation.
func ValidatorStrLen[T StringEx](minLen, maxLen int, lenFuncs ...func(s string) int) ValidatorFunc {
	return func(v interface{}) error {
		s, err := convertValue[T](v)
		if err != nil {
			return err
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

// ValidatorStrPrefix validate a string to have prefix matching the given one
func ValidatorStrPrefix[T StringEx](prefix string) ValidatorFunc {
	return func(v interface{}) error {
		s, err := convertValue[T](v)
		if err != nil {
			return err
		}
		if strings.HasPrefix(*(*string)(unsafe.Pointer(&s)), prefix) {
			return nil
		}
		return ErrValidationStrPrefix
	}
}

// ValidatorStrSuffix validate a string to have suffix matching the given one
func ValidatorStrSuffix[T StringEx](suffix string) ValidatorFunc {
	return func(v interface{}) error {
		s, err := convertValue[T](v)
		if err != nil {
			return err
		}
		if strings.HasSuffix(*(*string)(unsafe.Pointer(&s)), suffix) {
			return nil
		}
		return ErrValidationStrSuffix
	}
}

func convertValue[T any](v interface{}) (T, error) {
	val, ok := v.(T)
	if !ok {
		return val, fmt.Errorf("%w: (%v -> %v)", ErrValidationConversion, reflect.TypeOf(v), reflect.TypeOf(val))
	}
	return val, nil
}
