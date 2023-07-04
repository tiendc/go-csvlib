package csvlib

// func Test_parseInlineColumnDynamicType(t *testing.T) {
//	parent := &decodeColumnMeta{}
//	_, err := parseInlineColumnDynamicType(reflect.TypeOf([]string{}), parent)
//	assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
//
//	type DynInline1 struct {
//	}
//	_, err = parseInlineColumnDynamicType(reflect.TypeOf(DynInline1{}), parent)
//	assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
//
//	type DynInline2 struct {
//		Header []int
//	}
//	_, err = parseInlineColumnDynamicType(reflect.TypeOf(DynInline2{}), parent)
//	assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
//
//	type DynInline3 struct {
//		Header []string
//	}
//	_, err = parseInlineColumnDynamicType(reflect.TypeOf(DynInline3{}), parent)
//	assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
//
//	type DynInline4 struct {
//		Header []string
//		Values [10]int32
//	}
//	_, err = parseInlineColumnDynamicType(reflect.TypeOf(DynInline4{}), parent)
//	assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
//
//	type DynInlineOK struct {
//		Header []string
//		Values []int32
//	}
//	typ, err := parseInlineColumnDynamicType(reflect.TypeOf(DynInlineOK{}), parent)
//	assert.Nil(t, err)
//	assert.Equal(t, reflect.TypeOf(int32(0)), typ)
// }
