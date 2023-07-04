package csvlib

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiendc/gofn"
)

func Test_GetHeaderDetails(t *testing.T) {
	t.Run("#1: success", func(t *testing.T) {
		type Item struct {
			Col1 int     `csv:"col1,omitempty"`
			Col2 *string `csv:"col2,optional"`
			Col3 bool    `csv:"-"`
			Col4 float32
			Col5 InlineColumn[int] `csv:"col5,inline"`
		}
		details, err := GetHeaderDetails(Item{}, "csv")
		assert.Nil(t, err)
		assert.Equal(t, []ColumnDetail{
			{Name: "col1", DataType: reflect.TypeOf(int(0)), OmitEmpty: true},
			{Name: "col2", DataType: reflect.TypeOf(gofn.New("")), Optional: true},
			{Name: "col5", DataType: reflect.TypeOf(InlineColumn[int]{}), Inline: true},
		}, details)
	})

	t.Run("#2: invalid type", func(t *testing.T) {
		_, err := GetHeaderDetails("abc", "csv")
		assert.ErrorIs(t, err, ErrTypeInvalid)
	})
}

func Test_GetHeader(t *testing.T) {
	t.Run("#1: success", func(t *testing.T) {
		type Item struct {
			Col1 int     `csv:"col1,omitempty"`
			Col2 *string `csv:"col2,optional"`
			Col3 bool    `csv:"-"`
			Col4 float32
			Col5 InlineColumn[int] `csv:"col5,inline"`
		}
		header, err := GetHeader(Item{}, "csv")
		assert.Nil(t, err)
		assert.Equal(t, []string{"col1", "col2", "col5"}, header)
	})

	t.Run("#2: invalid type", func(t *testing.T) {
		_, err := GetHeader(0, "csv")
		assert.ErrorIs(t, err, ErrTypeInvalid)
	})
}
