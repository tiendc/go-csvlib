package csvlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ProcessorTrim(t *testing.T) {
	assert.Equal(t, "", ProcessorTrim(""))
	assert.Equal(t, "abc", ProcessorTrim("\n abc\t"))
}

func Test_ProcessorTrimPrefix(t *testing.T) {
	assert.Equal(t, "", ProcessorTrimPrefix("", "abc"))
	assert.Equal(t, "bc\t", ProcessorTrimPrefix("\n abc\t", "\n a"))
}

func Test_ProcessorTrimSuffix(t *testing.T) {
	assert.Equal(t, "", ProcessorTrimSuffix("", "abc"))
	assert.Equal(t, "\n a", ProcessorTrimSuffix("\n abc\t", "bc\t"))
}

func Test_ProcessorReplace(t *testing.T) {
	assert.Equal(t, "", ProcessorReplace("", "abc", "xyz", 1))
	assert.Equal(t, "xyzxyz", ProcessorReplace("abcxyz", "abc", "xyz", 1))
}

func Test_ProcessorReplaceAll(t *testing.T) {
	assert.Equal(t, "", ProcessorReplaceAll("", "abc", "xyz"))
	assert.Equal(t, "xyzxyzxyz", ProcessorReplaceAll("abcxyzabc", "abc", "xyz"))
}

func Test_ProcessorLower(t *testing.T) {
	assert.Equal(t, "", ProcessorLower(""))
	assert.Equal(t, "abc123", ProcessorLower("aBc123"))
}

func Test_ProcessorUpper(t *testing.T) {
	assert.Equal(t, "", ProcessorUpper(""))
	assert.Equal(t, "ABC123", ProcessorUpper("aBc123"))
}

func Test_ProcessorNumberGroup(t *testing.T) {
	assert.Equal(t, "", ProcessorNumberGroup("", '.', ','))
	assert.Equal(t, "aBc123", ProcessorNumberGroup("aBc123", '.', ','))
	assert.Equal(t, "123", ProcessorNumberGroup("123", '.', ','))
	assert.Equal(t, "12,345,678", ProcessorNumberGroup("12345678", '.', ','))
}

func Test_ProcessorNumberUngroup(t *testing.T) {
	assert.Equal(t, "", ProcessorNumberUngroup("", ','))
	assert.Equal(t, "aBc123", ProcessorNumberUngroup("aBc,123", ','))
	assert.Equal(t, "123", ProcessorNumberUngroup("123", ','))
	assert.Equal(t, "1234567.8", ProcessorNumberUngroup("12,3456,7.8", ','))
}

func Test_ProcessorNumberGroupComma(t *testing.T) {
	assert.Equal(t, "", ProcessorNumberGroupComma(""))
	assert.Equal(t, "aBc123", ProcessorNumberGroupComma("aBc123"))
	assert.Equal(t, "123", ProcessorNumberGroupComma("123"))
	assert.Equal(t, "12,345,678", ProcessorNumberGroupComma("12345678"))
}

func Test_ProcessorNumberUngroupComma(t *testing.T) {
	assert.Equal(t, "", ProcessorNumberUngroupComma(""))
	assert.Equal(t, "aBc123", ProcessorNumberUngroupComma("aBc,123"))
	assert.Equal(t, "123", ProcessorNumberUngroupComma("123"))
	assert.Equal(t, "1234567.8", ProcessorNumberUngroupComma("12,3456,7.8"))
}
