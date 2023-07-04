package csvlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidatorLT(t *testing.T) {
	assert.Nil(t, ValidatorLT(100)(99))
	assert.Nil(t, ValidatorLT(100)(0))
	assert.ErrorIs(t, ValidatorLT(100)("abc"), ErrValidationConversion)
	assert.ErrorIs(t, ValidatorLT(100)(100), ErrValidationLT)
	assert.ErrorIs(t, ValidatorLT(100)(101), ErrValidation)
}

func Test_ValidatorLTE(t *testing.T) {
	assert.Nil(t, ValidatorLTE(int64(0))(int64(0)))
	assert.Nil(t, ValidatorLTE(int64(0))(int64(-1)))
	assert.ErrorIs(t, ValidatorLTE(100)(true), ErrValidationConversion)
	assert.ErrorIs(t, ValidatorLTE(int64(0))(int64(1)), ErrValidationLTE)
	assert.ErrorIs(t, ValidatorLTE(int64(0))(int64(100)), ErrValidation)
}

func Test_ValidatorGT(t *testing.T) {
	assert.Nil(t, ValidatorGT(100)(101))
	assert.Nil(t, ValidatorGT(100)(10000))
	assert.ErrorIs(t, ValidatorGT(100)(int8(1)), ErrValidationConversion)
	assert.ErrorIs(t, ValidatorGT(100)(100), ErrValidationGT)
	assert.ErrorIs(t, ValidatorGT(100)(99), ErrValidation)
}

func Test_ValidatorGTE(t *testing.T) {
	assert.Nil(t, ValidatorGTE(int64(0))(int64(0)))
	assert.Nil(t, ValidatorGTE(int64(0))(int64(1)))
	assert.ErrorIs(t, ValidatorGTE(100)(int32(1)), ErrValidationConversion)
	assert.ErrorIs(t, ValidatorGTE(int64(0))(int64(-1)), ErrValidationGTE)
	assert.ErrorIs(t, ValidatorGTE(int64(0))(int64(-10)), ErrValidation)
}

func Test_ValidatorRange(t *testing.T) {
	assert.Nil(t, ValidatorRange(0, 10)(0))
	assert.Nil(t, ValidatorRange(0, 10)(10))
	assert.ErrorIs(t, ValidatorRange(0, 10)(int32(1)), ErrValidationConversion)
	assert.ErrorIs(t, ValidatorRange("a", "g")("h"), ErrValidationRange)
	assert.ErrorIs(t, ValidatorRange("a", "g")("0bc"), ErrValidation)
}

func Test_ValidatorIN(t *testing.T) {
	assert.Nil(t, ValidatorIN("a", "b", "c")("b"))
	assert.Nil(t, ValidatorIN("a", "b", "")(""))
	assert.ErrorIs(t, ValidatorIN("a", "b", "")(1), ErrValidationConversion)
	assert.ErrorIs(t, ValidatorIN("a", "b", "")("c"), ErrValidationIN)
	assert.ErrorIs(t, ValidatorIN("a", "b", "")("d"), ErrValidation)
}

func Test_ValidatorStrLen(t *testing.T) {
	lenFn := func(s string) int { return len(s) }
	assert.Nil(t, ValidatorStrLen[string](0, 5)("abc"))
	assert.Nil(t, ValidatorStrLen[string](0, 5)(""))
	assert.Nil(t, ValidatorStrLen[string](0, 5, lenFn)("abc"))
	assert.Nil(t, ValidatorStrLen[StrType](0, 5)(StrType("abc")))
	assert.ErrorIs(t, ValidatorStrLen[string](0, 5)(StrType("abc")), ErrValidationConversion)
	assert.ErrorIs(t, ValidatorStrLen[string](1, 5)(""), ErrValidationStrLen)
	assert.ErrorIs(t, ValidatorStrLen[string](0, 5)("abc123"), ErrValidation)
	assert.ErrorIs(t, ValidatorStrLen[string](0, 5, lenFn)("abc123"), ErrValidation)
	assert.ErrorIs(t, ValidatorStrLen[StrType](0, 5)(StrType("abc123")), ErrValidation)
}

func Test_ValidatorStrPrefix(t *testing.T) {
	assert.Nil(t, ValidatorStrPrefix[string]("a")("abc"))
	assert.Nil(t, ValidatorStrPrefix[string](" a")(" abc"))
	assert.Nil(t, ValidatorStrPrefix[StrType](" a")(StrType(" abc")))
	assert.ErrorIs(t, ValidatorStrPrefix[string]("x")(StrType("abc")), ErrValidationConversion)
	assert.ErrorIs(t, ValidatorStrPrefix[string]("x")("abc"), ErrValidationStrPrefix)
	assert.ErrorIs(t, ValidatorStrPrefix[string]("x")("abc"), ErrValidation)
	assert.ErrorIs(t, ValidatorStrPrefix[StrType]("x")(StrType("abc123")), ErrValidationStrPrefix)
}

func Test_ValidatorStrSuffix(t *testing.T) {
	assert.Nil(t, ValidatorStrSuffix[string]("c")("abc"))
	assert.Nil(t, ValidatorStrSuffix[string]("c ")(" abc "))
	assert.Nil(t, ValidatorStrSuffix[StrType]("c ")(StrType("abc ")))
	assert.ErrorIs(t, ValidatorStrSuffix[string]("x")(StrType("abc")), ErrValidationConversion)
	assert.ErrorIs(t, ValidatorStrSuffix[string]("x")("abc"), ErrValidationStrSuffix)
	assert.ErrorIs(t, ValidatorStrSuffix[string]("x")("abc"), ErrValidation)
	assert.ErrorIs(t, ValidatorStrSuffix[StrType]("x")(StrType("abc123")), ErrValidationStrSuffix)
}
