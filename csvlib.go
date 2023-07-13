package csvlib

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
)

const (
	DefaultTagName = "csv"
)

// Reader reader object interface required by the lib to read CSV data.
// Should use csv.Reader from the built-in package "encoding/csv".
type Reader interface {
	Read() ([]string, error)
}

// Writer writer object interface required by the lib to write CSV data to.
// Should use csv.Writer from the built-in package "encoding/csv".
type Writer interface {
	Write(record []string) error
}

// CSVUnmarshaler unmarshaler interface for decoding custom type
type CSVUnmarshaler interface {
	UnmarshalCSV([]byte) error
}

// CSVMarshaler marshaler interface for encoding custom type
type CSVMarshaler interface {
	MarshalCSV() ([]byte, error)
}

// DecodeFunc decode function for a given cell text
type DecodeFunc func(text string, v reflect.Value) error

// EncodeFunc encode function for a given Go value
type EncodeFunc func(v reflect.Value, omitempty bool) (string, error)

// ProcessorFunc function to transform cell value before decoding or after encoding
type ProcessorFunc func(s string) string

// ValidatorFunc function to validate the values of decoded cells
type ValidatorFunc func(v any) error

type ParameterMap map[string]any

// LocalizationFunc function to translate message into a specific language
type LocalizationFunc func(key string, params ParameterMap) (string, error)

// OnCellErrorFunc function to be called when error happens on decoding cell value
type OnCellErrorFunc func(e *CellError)

type ColumnDetail struct {
	Name      string
	Optional  bool
	OmitEmpty bool
	Inline    bool
	DataType  reflect.Type
}

// Unmarshal convenient method to decode CVS data into a slice of structs
func Unmarshal(data []byte, v any, options ...DecodeOption) (*DecodeResult, error) {
	decoder := NewDecoder(csv.NewReader(bytes.NewReader(data)), options...)
	return decoder.Decode(v)
}

// Marshal convenient method to encode a slice of structs into CSV format
func Marshal(v any, options ...EncodeOption) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	encoder := NewEncoder(w, options...)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetHeaderDetails get CSV header details from the given struct type
func GetHeaderDetails(v any, tagName string) (columnDetails []ColumnDetail, err error) {
	t := reflect.TypeOf(v)
	t = indirectType(t)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: must be struct", ErrTypeInvalid)
	}
	numFields := t.NumField()
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		tag, _ := parseTag(tagName, field)
		if tag == nil || tag.ignored {
			continue
		}
		columnDetails = append(columnDetails, ColumnDetail{
			Name:      tag.name,
			Optional:  tag.optional,
			OmitEmpty: tag.omitEmpty,
			Inline:    tag.inline,
			DataType:  field.Type,
		})
	}
	return
}

// GetHeader get CSV header from the given struct
func GetHeader(v any, tagName string) ([]string, error) {
	details, err := GetHeaderDetails(v, tagName)
	if err != nil {
		return nil, err
	}
	header := make([]string, 0, len(details))
	for i := range details {
		header = append(header, details[i].Name)
	}
	return header, nil
}
