package csvlib

import (
	"reflect"
)

type inlineColumnStructType int8

const (
	inlineColumnStructDynamic = inlineColumnStructType(1)
	inlineColumnStructFixed   = inlineColumnStructType(2)

	dynamicInlineColumnHeader = "Header"
	dynamicInlineColumnValues = "Values"
)

// InlineColumn represents inline columns of type `T`
type InlineColumn[T any] struct {
	Header []string
	Values []T
}

// inlineColumnMeta metadata of inline columns
type inlineColumnMeta struct {
	headerText  []string
	inlineType  inlineColumnStructType
	targetField reflect.StructField
	dataType    reflect.Type

	// columnCurrIndex current processing column (used for dynamic inline columns)
	columnCurrIndex int
}

func (m *inlineColumnMeta) decodePrepareForNextRow() {
	m.columnCurrIndex = -1
}

func (m *inlineColumnMeta) decodeInitInlineStruct(inlineStruct reflect.Value) {
	if inlineStruct.Kind() == reflect.Pointer {
		inlineStruct.Set(reflect.New(inlineStruct.Type().Elem())) // No need to check Nil as this is decoding
		inlineStruct = inlineStruct.Elem()
	}

	switch m.inlineType {
	case inlineColumnStructFixed:
		// Do nothing
	case inlineColumnStructDynamic:
		if m.columnCurrIndex != -1 {
			return
		}
		numCols := len(m.headerText)
		inlineStruct.FieldByName(dynamicInlineColumnHeader).Set(reflect.ValueOf(m.headerText))

		columnValues := inlineStruct.Field(m.targetField.Index[0])
		columnValues.Set(reflect.MakeSlice(reflect.SliceOf(m.dataType), numCols, numCols))
		m.columnCurrIndex = 0
	}
}

func (m *inlineColumnMeta) decodeGetColumnValue(inlineStruct reflect.Value) reflect.Value {
	if inlineStruct.Kind() == reflect.Pointer {
		inlineStruct.Set(reflect.New(inlineStruct.Type().Elem())) // No need to check Nil as this is decoding
		inlineStruct = inlineStruct.Elem()
	}

	switch m.inlineType {
	case inlineColumnStructFixed:
		return inlineStruct.Field(m.targetField.Index[0])
	case inlineColumnStructDynamic:
		if m.columnCurrIndex == -1 {
			m.decodeInitInlineStruct(inlineStruct)
		}
		colVal := inlineStruct.Field(m.targetField.Index[0]).Index(m.columnCurrIndex)
		m.columnCurrIndex++
		return colVal
	}
	return reflect.Value{}
}

func (m *inlineColumnMeta) encodePrepareForNextRow() {
	m.columnCurrIndex = 0
}

func (m *inlineColumnMeta) encodeGetColumnValue(inlineStruct reflect.Value) reflect.Value {
	if inlineStruct.Kind() == reflect.Pointer {
		inlineStruct = inlineStruct.Elem()
		if !inlineStruct.IsValid() {
			return reflect.Value{}
		}
	}

	switch m.inlineType {
	case inlineColumnStructFixed:
		return inlineStruct.Field(m.targetField.Index[0])
	case inlineColumnStructDynamic:
		colVal := inlineStruct.Field(m.targetField.Index[0]).Index(m.columnCurrIndex)
		m.columnCurrIndex++
		return colVal
	}
	return reflect.Value{}
}
