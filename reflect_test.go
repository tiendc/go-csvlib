package csvlib

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isKindOrPtrOf(t *testing.T) {
	var v int
	assert.True(t, isKindOrPtrOf(reflect.TypeOf(v), reflect.Int))
	assert.True(t, isKindOrPtrOf(reflect.TypeOf(&v), reflect.Int))
	assert.False(t, isKindOrPtrOf(reflect.TypeOf(&v), reflect.String))
}

func Test_indirectType(t *testing.T) {
	var v string
	assert.Equal(t, reflect.TypeOf(""), indirectType(reflect.TypeOf(v)))
	assert.Equal(t, reflect.TypeOf(""), indirectType(reflect.TypeOf(&v)))
}

func Test_indirectValue(t *testing.T) {
	v := "abc"
	assert.Equal(t, "abc", indirectValue(reflect.ValueOf(v)).String())
	assert.Equal(t, "abc", indirectValue(reflect.ValueOf(&v)).String())
}

func Test_initAndIndirectValue(t *testing.T) {
	v := "abc"
	assert.Equal(t, "abc", initAndIndirectValue(reflect.ValueOf(v)).String())
	assert.Equal(t, "abc", initAndIndirectValue(reflect.ValueOf(&v)).String())

	var v2 int
	assert.Equal(t, int64(0), initAndIndirectValue(reflect.ValueOf(v2)).Int())
	assert.Equal(t, int64(0), initAndIndirectValue(reflect.ValueOf(&v2)).Int())

	s := []*string{nil}
	assert.Equal(t, "", initAndIndirectValue(reflect.ValueOf(s).Index(0)).String())
}
