package csvlib

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseTag(t *testing.T) {
	type Item struct {
		Col0 bool
		Col1 int               `csv:"col1,optional"`
		Col2 *int              `csv:"col2,omitempty"`
		Col3 string            `csv:"-"`
		Col4 string            `csv:""`
		Col5 string            `csv:",unsupported"`
		Col6 InlineColumn[int] `csv:"col6,inline,prefix=xyz"`
	}
	structType := reflect.TypeOf(Item{})

	col0, _ := structType.FieldByName("Col0")
	tag0, err := parseTag(DefaultTagName, col0)
	assert.Nil(t, err)
	assert.Nil(t, tag0)

	col1, _ := structType.FieldByName("Col1")
	tag1, err := parseTag(DefaultTagName, col1)
	assert.Nil(t, err)
	assert.True(t, tag1.name == "col1" && tag1.optional)

	col2, _ := structType.FieldByName("Col2")
	tag2, err := parseTag(DefaultTagName, col2)
	assert.Nil(t, err)
	assert.True(t, tag2.name == "col2" && tag2.omitEmpty)

	col3, _ := structType.FieldByName("Col3")
	tag3, err := parseTag(DefaultTagName, col3)
	assert.Nil(t, err)
	assert.True(t, tag3.ignored)

	col4, _ := structType.FieldByName("Col4")
	tag4, err := parseTag(DefaultTagName, col4)
	assert.Nil(t, err)
	assert.True(t, tag4.name == "Col4" && tag4.empty)

	col5, _ := structType.FieldByName("Col5")
	tag5, err := parseTag(DefaultTagName, col5)
	assert.Nil(t, err)
	assert.True(t, tag5.name == "Col5")

	col6, _ := structType.FieldByName("Col6")
	tag6, err := parseTag(DefaultTagName, col6)
	assert.Nil(t, err)
	assert.True(t, tag6.name == "col6" && tag6.inline && tag6.prefix == "xyz")
}
