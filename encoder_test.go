package csvlib

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiendc/gofn"
)

func Test_Encode_configOption(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int16   `csv:"col1"`
		Col2 StrType `csv:"col2"`
	}

	t.Run("#1: column option not found", func(t *testing.T) {
		v := []Item{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.ColumnConfigMap = EncodeColumnConfigMap{
				"ColX": {},
				"col1": {},
				"colX": {},
			}
		}).Encode(v)
		assert.ErrorIs(t, err, ErrConfigOptionInvalid)
	})

	t.Run("#2: localize header without localization func", func(t *testing.T) {
		v := []Item{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.LocalizeHeader = true
		}).Encode(v)
		assert.ErrorIs(t, err, ErrConfigOptionInvalid)
	})

	t.Run("#3: invalid input var", func(t *testing.T) {
		var v []Item
		err := NewEncoder(nil).Encode(v)
		assert.ErrorIs(t, err, ErrValueNil)
	})

	t.Run("#4: invalid input var", func(t *testing.T) {
		v := []string{}
		err := NewEncoder(nil).Encode(&v)
		assert.ErrorIs(t, err, ErrTypeInvalid)
	})

	t.Run("#5: invalid input var", func(t *testing.T) {
		var v Item
		err := NewEncoder(nil).Encode(&v)
		assert.ErrorIs(t, err, ErrTypeInvalid)
	})
}

func Test_Encode_withHeader(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int     `csv:"col1"`
		Col2 float32 `csv:"col2"`
	}

	t.Run("#1: column optional", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Col2: 2.123},
			{Col1: 100, Col2: 200},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2
				false,1,2.123
				false,100,200
			`), buf.String())
	})

	t.Run("#2: no header mode", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Col2: 2.123},
			{Col1: 100, Col2: 200},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.NoHeaderMode = true
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`false,1,2.123
				false,100,200
			`), buf.String())
	})
}

func Test_Encode_withPostprocessor(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional,omitempty"`
		ColY bool
		Col1 int    `csv:"col1"`
		Col2 string `csv:"col2"`
	}

	t.Run("#1: trim/upper specific column after encoding", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Col2: "\tabcXYZ "},
			{Col1: 100, Col2: " xYz123 "},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.ColumnConfigMap = EncodeColumnConfigMap{
				"col2": {
					PostprocessorFuncs: []ProcessorFunc{ProcessorTrim, ProcessorUpper},
				},
			}
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2
				,1,ABCXYZ
				,100,XYZ123
			`), buf.String())
	})

	t.Run("#2: add comma to numbers after encoding", func(t *testing.T) {
		type Item struct {
			ColX bool `csv:",optional,omitempty"`
			ColY bool
			Col1 int     `csv:"col1"`
			Col2 float64 `csv:"col2"`
		}
		v := []Item{
			{Col1: 12345, Col2: 1234.1234567},
			{Col1: 1234567, Col2: 1.1234},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.ColumnConfigMap = EncodeColumnConfigMap{
				"col1": {
					PostprocessorFuncs: []ProcessorFunc{ProcessorNumberGroupComma},
				},
				"col2": {
					PostprocessorFuncs: []ProcessorFunc{ProcessorNumberGroupComma},
				},
			}
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2
				,"12,345","1,234.1234567"
				,"1,234,567",1.1234
			`), buf.String())
	})
}

func Test_Encode_withSpecialChars(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional,omitempty"`
		ColY bool
		Col1 int    `csv:"col1"`
		Col2 string `csv:"col2"`
	}

	t.Run("#1: values have CRLF and TAB", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Col2: " a\tbc\nXYZ "},
			{Col1: 100, Col2: " xYz\n123 "},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.ColumnConfigMap = EncodeColumnConfigMap{
				"col2": {
					PostprocessorFuncs: []ProcessorFunc{ProcessorTrim, ProcessorUpper},
				},
			}
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2
				,1,"A	BC
					XYZ"
				,100,"XYZ
					123"
			`), buf.String())
	})

	t.Run("#2: values have bare double-quote and single quote", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Col2: "abc\"XYZ"},
			{Col1: 100, Col2: "xYz'123"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2
				,1,"abc""XYZ"
				,100,xYz'123
			`), buf.String())
	})

	t.Run("#3: values have valid double-quote", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Col2: "\"abcXYZ\""},
			{Col1: 100, Col2: "xYz'123"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2
				,1,"""abcXYZ"""
				,100,xYz'123
			`), buf.String())
	})
}

func Test_Encode_withOmitEmpty(t *testing.T) {
	type Item struct {
		ColX bool   `csv:",optional,omitempty"`
		Col1 int    `csv:"col1,omitempty"`
		Col2 string `csv:"col2,omitempty"`
		Col3 *int   `csv:"col3,omitempty"`
	}

	t.Run("#1: success", func(t *testing.T) {
		v := []*Item{
			{},
			nil,
			{Col3: gofn.New(0)},
			{Col1: 123},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2,col3
				,,,
				,,,
				,123,,
			`), buf.String())
	})
}

func Test_Encode_multipleCalls(t *testing.T) {
	type Item struct {
		ColY bool
		Col1 int     `csv:"col1"`
		Col2 float32 `csv:"col2"`
	}
	type Item2 struct {
		ColY bool
		Col1 int     `csv:"col1"`
		Col2 float32 `csv:"col2"`
	}

	t.Run("#1: encode multiple times", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Col2: 1.12345},
			{Col1: 2, Col2: 6.543210},
		}
		buf := bytes.NewBuffer([]byte{})
		e := NewEncoder(csv.NewWriter(buf))
		err := e.Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`col1,col2
				1,1.12345
				2,6.54321
			`), buf.String())

		// Second call
		v = []Item{
			{Col1: 3, Col2: 1.12345},
			{Col1: 4, Col2: 6.543},
		}
		err = e.Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`col1,col2
				1,1.12345
				2,6.54321
				3,1.12345
				4,6.543
			`), buf.String())

		// Finish encoding, then try to encode more
		e.Finish()
		err = e.Encode(v)
		assert.ErrorIs(t, err, ErrEncodeAlreadyFinished)
	})

	t.Run("#2: encode different types of data", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Col2: 1.12345},
			{Col1: 2, Col2: 6.543210},
		}
		buf := bytes.NewBuffer([]byte{})
		e := NewEncoder(csv.NewWriter(buf))
		err := e.Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`col1,col2
				1,1.12345
				2,6.54321
			`), buf.String())

		// Second call use different data type
		v2 := []Item2{
			{Col1: 3, Col2: 1.12345},
			{Col1: 4, Col2: 6.543},
		}
		err = e.Encode(v2)
		assert.ErrorIs(t, err, ErrTypeUnmatched)
	})
}

func Test_Encode_withFixedInlineColumn(t *testing.T) {
	type Sub struct {
		ColZ bool `csv:",optional"`
		ColY bool
		Col1 int16  `csv:"sub1"`
		Col2 string `csv:"sub2,optional"`
	}
	type Item struct {
		ColX bool `csv:",optional,omitempty"`
		ColY bool
		Col1 int    `csv:"col1"`
		Sub1 Sub    `csv:"sub1,inline"`
		Col2 string `csv:"col2"`
	}
	type Item2 struct {
		ColX bool `csv:",optional,omitempty"`
		ColY bool
		Col1 int    `csv:"col1"`
		Sub1 *Sub   `csv:"sub1,inline"`
		Col2 string `csv:"col2"`
	}

	t.Run("#1: success", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Sub1: Sub{Col1: 111}, Col2: "abcxyz123"},
			{Col1: 1000, Sub1: Sub{Col1: 222}, Col2: "abc123"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,ColZ,sub1,sub2,col2
				,1,false,111,,abcxyz123
				,1000,false,222,,abc123
			`), buf.String())
	})

	t.Run("#2: success with pointer type", func(t *testing.T) {
		v := []Item2{
			{Col1: 1, Sub1: &Sub{Col1: 111}, Col2: "abcxyz123"},
			{Col1: 1000, Sub1: &Sub{Col1: 222}, Col2: "abc123"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,ColZ,sub1,sub2,col2
				,1,false,111,,abcxyz123
				,1000,false,222,,abc123
			`), buf.String())
	})

	t.Run("#3: with no header mode", func(t *testing.T) {
		v := []Item{
			{Col1: 1, Sub1: Sub{Col1: 111}, Col2: "abcxyz123"},
			{Col1: 1000, Sub1: Sub{Col1: 222}, Col2: "abc123"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.NoHeaderMode = true
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`,1,false,111,,abcxyz123
				,1000,false,222,,abc123
			`), buf.String())
	})

	t.Run("#4: with column prefix", func(t *testing.T) {
		type Item struct {
			ColX bool `csv:",optional,omitempty"`
			ColY bool
			Col1 int    `csv:"col1"`
			Sub1 *Sub   `csv:"sub1,inline,prefix=sub_"`
			Col2 string `csv:"col2"`
		}
		v := []Item{
			{Col1: 1, Sub1: &Sub{Col1: 111}, Col2: "abcxyz123"},
			{Col1: 1000, Sub1: &Sub{Col1: 222}, Col2: "abc123"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,sub_ColZ,sub_sub1,sub_sub2,col2
				,1,false,111,,abcxyz123
				,1000,false,222,,abc123
			`), buf.String())
	})

	t.Run("#5: invalid inline column", func(t *testing.T) {
		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int    `csv:"col1,inline"`
			Col2 string `csv:"col2"`
		}

		v := []Item{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
	})

	t.Run("#6: with prefix and custom postprocessor", func(t *testing.T) {
		type Item struct {
			Col1 int    `csv:"col1"`
			Sub1 Sub    `csv:"sub1,inline,prefix=sub_"`
			Col2 string `csv:"col2"`
		}

		v := []Item{
			{Col1: 1, Sub1: Sub{Col1: 12345, Col2: "abC"}, Col2: "abcxyz123"},
			{Col1: 1000, Sub1: Sub{Col1: 4321, Col2: "xYz123"}, Col2: "abc123"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.ColumnConfigMap = EncodeColumnConfigMap{
				"sub_sub1": {
					PostprocessorFuncs: []ProcessorFunc{ProcessorNumberGroupComma},
				},
				"sub1": {
					PostprocessorFuncs: []ProcessorFunc{ProcessorUpper},
				},
			}
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`col1,sub_ColZ,sub_sub1,sub_sub2,col2
				1,FALSE,"12,345",ABC,abcxyz123
				1000,FALSE,"4,321",XYZ123,abc123
			`), buf.String())
	})
}

func Test_Encode_withDynamicInlineColumn(t *testing.T) {
	type Item struct {
		ColX bool              `csv:",optional"`
		ColY bool              `csv:"-"`
		Col1 int               `csv:"col1"`
		Sub1 InlineColumn[int] `csv:"sub1,inline"`
		Col2 *string           `csv:"col2"`
		ColZ bool              `csv:",optional"`
	}
	type Item2 struct {
		ColX bool               `csv:",optional"`
		ColY bool               `csv:"-"`
		Col1 int                `csv:"col1"`
		Sub1 *InlineColumn[int] `csv:"sub1,inline"`
		Col2 *string            `csv:"col2"`
		ColZ bool               `csv:",optional"`
	}

	t.Run("#1: success", func(t *testing.T) {
		header := []string{"sub1", "sub2"}
		v := []Item{
			{Col1: 1, Sub1: InlineColumn[int]{Header: header, Values: []int{111, 11}}, Col2: gofn.New("abcxyz123")},
			{Col1: 1000, Sub1: InlineColumn[int]{Header: header, Values: []int{222, 22}}, Col2: gofn.New("abc123")},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,sub1,sub2,col2,ColZ
				false,1,111,11,abcxyz123,false
				false,1000,222,22,abc123,false
			`), buf.String())
	})

	t.Run("#2: success with using pointer", func(t *testing.T) {
		header := []string{"sub1", "sub2"}
		v := []Item2{
			{Col1: 1, Sub1: &InlineColumn[int]{Header: header, Values: []int{111, 11}}, Col2: gofn.New("abcxyz123")},
			{Col1: 1000, Sub1: &InlineColumn[int]{Header: header, Values: []int{222, 22}}, Col2: gofn.New("abc123")},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,sub1,sub2,col2,ColZ
				false,1,111,11,abcxyz123,false
				false,1000,222,22,abc123,false
			`), buf.String())
	})

	t.Run("#3: no header mode", func(t *testing.T) {
		header := []string{"sub1", "sub2"}
		v := []Item{
			{Col1: 1, Sub1: InlineColumn[int]{Header: header, Values: []int{111, 11}}, Col2: gofn.New("abcxyz123")},
			{Col1: 1000, Sub1: InlineColumn[int]{Header: header, Values: []int{222, 22}}, Col2: gofn.New("abc123")},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.NoHeaderMode = true
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`false,1,111,11,abcxyz123,false
				false,1000,222,22,abc123,false
			`), buf.String())
	})

	t.Run("#4: with column prefix", func(t *testing.T) {
		type Item struct {
			ColX bool              `csv:",optional"`
			ColY bool              `csv:"-"`
			Col1 *int              `csv:"col1"`
			Sub1 InlineColumn[int] `csv:"sub1,inline,prefix=sub_"`
			Col2 string            `csv:"col2"`
		}

		header := []string{"col1", "col2"}
		v := []Item{
			{Col1: gofn.New(1), Sub1: InlineColumn[int]{Header: header, Values: []int{1234, 11}}, Col2: "abcxyz123"},
			{Col1: gofn.New(1000), Sub1: InlineColumn[int]{Header: header, Values: []int{12345, 22}}, Col2: "abc123"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.LocalizeHeader = true
			cfg.LocalizationFunc = localizeViVn
			cfg.ColumnConfigMap = EncodeColumnConfigMap{
				"sub_col1": {
					PostprocessorFuncs: []ProcessorFunc{ProcessorNumberGroupComma},
				},
			}
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`Cột-X,cột-1,cột-phụ-1,cột-phụ-2,cột-2
				false,1,"1,234",11,abcxyz123
				false,1000,"12,345",22,abc123
			`), buf.String())
	})

	t.Run("#5: invalid dynamic inline type", func(t *testing.T) {
		type InlineColumn2[T any] struct {
			Header []string
		}
		type Item struct {
			Col1 int                `csv:"col1"`
			Sub1 InlineColumn2[int] `csv:"sub1,inline"`
			Col2 string             `csv:"col2"`
		}

		header := []string{"sub1", "sub2"}
		v := []*Item{
			{Col1: 1, Sub1: InlineColumn2[int]{Header: header}, Col2: "aBc123"},
			{Col1: 2, Sub1: InlineColumn2[int]{Header: header}, Col2: "xyzZ"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
	})

	t.Run("#6: invalid dynamic inline type", func(t *testing.T) {
		type InlineColumn2[T any] struct {
			Header []string
			Values *[]T
		}
		type Item struct {
			Col1 int                `csv:"col1"`
			Sub1 InlineColumn2[int] `csv:"sub1,inline"`
			Col2 string             `csv:"col2"`
		}

		header := []string{"sub1", "sub2"}
		v := []*Item{
			{Col1: 1, Sub1: InlineColumn2[int]{Header: header}, Col2: "aBc123"},
			{Col1: 2, Sub1: InlineColumn2[int]{Header: header}, Col2: "xyzZ"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
	})
}

func Test_Encode_withLocalization(t *testing.T) {
	type Item struct {
		ColX bool    `csv:",optional"`
		ColY bool    `csv:"-"`
		Col1 int16   `csv:"col1"`
		Col2 StrType `csv:"col2"`
	}

	t.Run("#1: localization fails", func(t *testing.T) {
		v := []Item{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.LocalizeHeader = true
			cfg.LocalizationFunc = localizeFail
		}).Encode(v)
		assert.ErrorIs(t, err, ErrLocalizationFailed)
		assert.ErrorIs(t, err, errKeyNotFound)
	})

	t.Run("#2: no header mode, empty data", func(t *testing.T) {
		v := []Item{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.NoHeaderMode = true
			cfg.LocalizeHeader = true
			cfg.LocalizationFunc = localizeEnUs
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, "", buf.String())
	})

	t.Run("#3: no header mode, have data", func(t *testing.T) {
		v := []Item{
			{Col1: 111},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.NoHeaderMode = true
			cfg.LocalizeHeader = true
			cfg.LocalizationFunc = localizeEnUs
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`false,111,
			`), buf.String())
	})
}

func Test_Encode_withCustomMarshaler(t *testing.T) {
	t.Run("#1: no encode func matching", func(t *testing.T) {
		type Item struct {
			Col1 int            `csv:"col1"`
			Col2 float32        `csv:"col2"`
			Col3 map[int]string `csv:"col3"`
		}

		v := []Item{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.ErrorIs(t, err, ErrTypeUnsupported)
	})

	t.Run("#2: ptr of type implements MarshalText/MarshalCSV", func(t *testing.T) {
		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int          `csv:"col1"`
			Col2 StrUpperType `csv:"col2"`
			Col3 StrLowerType `csv:"col3"`
		}

		v := []Item{
			{Col1: 1, Col2: "aBcXyZ123", Col3: "aA"},
			{Col1: 1000, Col2: "aBc123", Col3: "bB"},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2,col3
				false,1,ABCXYZ123,aa
				false,1000,ABC123,bb
			`), buf.String())
	})

	t.Run("#3: type implements MarshalText/MarshalCSV", func(t *testing.T) {
		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int           `csv:"col1"`
			Col2 *StrUpperType `csv:"col2"`
			Col3 *StrLowerType `csv:"col3"`
		}

		v := []Item{
			{Col1: 1, Col2: gofn.New[StrUpperType]("aBcXyZ123"), Col3: gofn.New[StrLowerType]("aA")},
			{Col1: 1000, Col2: gofn.New[StrUpperType]("aBc123"), Col3: gofn.New[StrLowerType]("bB")},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2,col3
				false,1,ABCXYZ123,aa
				false,1000,ABC123,bb
			`), buf.String())
	})
}

func Test_Encode_specialCases(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int     `csv:"col1"`
		Col2 float32 `csv:"col2"`
	}

	t.Run("#1: no input data", func(t *testing.T) {
		v := []Item{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(v))
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2
			`), buf.String())
	})

	t.Run("#2: no input data and no header mode", func(t *testing.T) {
		v := []Item{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf), func(cfg *EncodeConfig) {
			cfg.NoHeaderMode = true
		}).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, "", buf.String())
	})

	t.Run("#3: duplicated column from struct", func(t *testing.T) {
		type Item2 struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int     `csv:"col1"`
			Col2 float32 `csv:"col2"`
			Col3 int     `csv:"col1"`
		}
		v := []Item2{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.ErrorIs(t, err, ErrHeaderColumnDuplicated)
	})

	t.Run("#4: invalid header (contains space)", func(t *testing.T) {
		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int     `csv:"col1 "`
			Col2 float32 `csv:" col2"`
		}
		v := []Item{}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.ErrorIs(t, err, ErrHeaderColumnInvalid)
	})

	t.Run("#5: item type is pointer and nil item is ignored", func(t *testing.T) {
		v := []*Item{
			{Col1: 1, Col2: 2.123},
			nil,
			{Col1: 100, Col2: 200},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`ColX,col1,col2
				false,1,2.123
				false,100,200
			`), buf.String())
	})
}

func Test_Encode_specialTypes(t *testing.T) {
	t.Run("#1: interface{} type", func(t *testing.T) {
		type Item struct {
			Col1 int         `csv:"col1"`
			Col2 interface{} `csv:"col2"`
		}

		v := []Item{
			{Col1: 1, Col2: 2.123},
			{Col1: 2, Col2: "200"},
			{Col1: 3, Col2: true},
			{Col1: 4, Col2: nil},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`col1,col2
				1,2.123
				2,200
				3,true
				4,
			`), buf.String())
	})

	t.Run("#2: ptr interface{} type", func(t *testing.T) {
		type Item struct {
			Col1 int          `csv:"col1"`
			Col2 *interface{} `csv:"col2"`
		}

		v := []*Item{
			{Col1: 1, Col2: gofn.New[interface{}](2.123)},
			nil,
			{Col1: 100, Col2: gofn.New[interface{}]("200")},
			{Col1: 100, Col2: gofn.New[interface{}](true)},
		}
		buf := bytes.NewBuffer([]byte{})
		err := NewEncoder(csv.NewWriter(buf)).Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`col1,col2
				1,2.123
				100,200
				100,true
			`), buf.String())
	})

	t.Run("#3: all base types", func(t *testing.T) {
		type Item struct {
			Col1  int     `csv:"c1"`
			Col2  *int    `csv:"c2"`
			Col3  int8    `csv:"c3"`
			Col4  *int8   `csv:"c4"`
			Col5  int16   `csv:"c5"`
			Col6  *int16  `csv:"c6"`
			Col7  int32   `csv:"c7"`
			Col8  *int32  `csv:"c8"`
			Col9  int64   `csv:"c9"`
			Col10 *int64  `csv:"c10"`
			Col11 uint    `csv:"c11"`
			Col12 *uint   `csv:"c12"`
			Col13 uint8   `csv:"c13"`
			Col14 *uint8  `csv:"c14"`
			Col15 uint16  `csv:"c15"`
			Col16 *uint16 `csv:"c16"`
			Col17 uint32  `csv:"c17"`
			Col18 *uint32 `csv:"c18"`
			Col19 uint64  `csv:"c19"`
			Col20 *uint64 `csv:"c20"`

			Col21 float32  `csv:"c21"`
			Col22 *float32 `csv:"c22"`
			Col23 float64  `csv:"c23"`
			Col24 *float64 `csv:"c24"`

			Col25 bool    `csv:"c25"`
			Col26 *bool   `csv:"c26"`
			Col27 string  `csv:"c27"`
			Col28 *string `csv:"c28"`
		}

		v := []Item{
			{-1, gofn.New(-1), int8(-1), gofn.New(int8(-1)), int16(-1), gofn.New(int16(-1)),
				int32(-1), gofn.New(int32(-1)), int64(-1), gofn.New(int64(-1)), uint(1), gofn.New(uint(1)),
				uint8(1), gofn.New(uint8(1)), uint16(1), gofn.New(uint16(1)), uint32(1), gofn.New(uint32(1)),
				uint64(1), gofn.New(uint64(1)), float32(1), gofn.New(float32(1)), float64(1), gofn.New(float64(1)),
				false, gofn.New(false), "abc", gofn.New("123")},
		}

		buf := bytes.NewBuffer([]byte{})
		e := NewEncoder(csv.NewWriter(buf))
		err := e.Encode(v)
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`c1,c2,c3,c4,c5,c6,c7,c8,c9,c10,c11,c12,c13,c14,c15,c16,c17,c18,c19,c20,c21,c22,c23,c24,c25,c26,c27,c28
			-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,false,false,abc,123
			`), buf.String())
	})
}

func Test_EncodeOne(t *testing.T) {
	type Item struct {
		ColY bool
		Col1 int      `csv:"col1,omitempty"`
		Col2 *float32 `csv:"col2,omitempty"`
	}
	type Item2 struct {
		ColY bool
		Col1 int     `csv:"col1,omitempty"`
		Col2 float32 `csv:"col2,omitempty"`
	}

	t.Run("#1: encode one multiple times", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		e := NewEncoder(csv.NewWriter(buf))
		err := e.EncodeOne(Item{Col1: 1, Col2: gofn.New[float32](1.12345)})
		e.flushWriter()
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`col1,col2
				1,1.12345
			`), buf.String())

		// Second call
		err = e.EncodeOne(Item{Col1: 2, Col2: gofn.New[float32](0.0)})
		e.flushWriter()
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`col1,col2
				1,1.12345
				2,
			`), buf.String())

		// Third call
		err = e.EncodeOne(Item{Col1: 0, Col2: gofn.New[float32](2.22)})
		e.flushWriter()
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`col1,col2
				1,1.12345
				2,
				,2.22
			`), buf.String())

		// Encode a different data type
		err = e.EncodeOne(Item2{Col1: 11, Col2: 11})
		assert.ErrorIs(t, err, ErrTypeUnmatched)

		// Finish encoding, then try to encode more
		e.Finish()
		err = e.EncodeOne(Item{Col1: 0})
		assert.ErrorIs(t, err, ErrEncodeAlreadyFinished)
	})
}
