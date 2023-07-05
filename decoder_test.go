package csvlib

import (
	"encoding/csv"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiendc/gofn"
)

func Test_Decode_configOption(t *testing.T) {
	type Item struct {
		ColX bool    `csv:",optional"`
		ColY bool    `csv:"-"`
		Col1 int16   `csv:"col1"`
		Col2 StrType `csv:"col2"`
	}

	t.Run("#1: column option not found", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			1000,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.StopOnError = false
			cfg.ConfigureColumn("col1", func(config *DecodeColumnConfig) {})
			cfg.ConfigureColumn("colX", func(config *DecodeColumnConfig) {})
		}).Decode(&v)
		assert.Nil(t, ret)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrConfigOptionInvalid)
	})

	t.Run("#2: localized header without localization func", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			1000,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.ParseLocalizedHeader = true
		}).Decode(&v)
		assert.Nil(t, ret)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrConfigOptionInvalid)
	})

	t.Run("#3: invalid output var", func(t *testing.T) {
		var v []Item
		ret, err := NewDecoder(nil).Decode(v)
		assert.Nil(t, ret)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrTypeInvalid)
	})

	t.Run("#4: invalid output var", func(t *testing.T) {
		var v []int
		ret, err := NewDecoder(nil).Decode(&v)
		assert.Nil(t, ret)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrTypeInvalid)
	})

	t.Run("#5: invalid output var", func(t *testing.T) {
		var v Item
		ret, err := NewDecoder(nil).Decode(&v)
		assert.Nil(t, ret)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrTypeInvalid)
	})

	t.Run("#6: define prefix on non-inline column", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			1000,abc123`)
		type Item struct {
			Col1 int16   `csv:"col1,prefix=abc"`
			Col2 StrType `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrTagOptionInvalid)
	})

	t.Run("#7: define optional on inline column", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			1000,abc123`)
		type Item struct {
			Col1 InlineColumn[int] `csv:"col1,inline,optional"`
			Col2 StrType           `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrTagOptionInvalid)
	})
}

func Test_Decode_withOptionalColumn(t *testing.T) {
	type Item struct {
		ColX bool    `csv:",optional"`
		ColY bool    `csv:"-"`
		Col1 int     `csv:"col1"`
		Col2 float32 `csv:"col2"`
	}

	t.Run("#1: column optional", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,2.123
			100,200`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []string{"ColX"}, ret.MissingOptionalColumns())
		assert.Equal(t, []Item{{Col1: 1, Col2: 2.123}, {Col1: 100, Col2: 200}}, v)
	})

	t.Run("#2: column required", func(t *testing.T) {
		data := gofn.MultilineString(
			`ColX,col2
			1,2.123
			100,200`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrHeaderColumnRequired)
	})
}

func Test_Decode_withUnrecognizedColumn(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int64   `csv:"col1"`
		Col2 float64 `csv:"col2"`
	}
	data := gofn.MultilineString(
		`col-x,col1,col2,col-y
			a,1,2.123,b
			c,100,200,d`)

	t.Run("#1: success", func(t *testing.T) {
		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.AllowUnrecognizedColumns = true
		}).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []string{"ColX"}, ret.MissingOptionalColumns())
		assert.Equal(t, []string{"col-x", "col-y"}, ret.UnrecognizedColumns())
		assert.Equal(t, []Item{{Col1: 1, Col2: 2.123}, {Col1: 100, Col2: 200}}, v)
	})

	t.Run("#1: failure", func(t *testing.T) {
		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrHeaderColumnUnrecognized)
	})
}

func Test_Decode_withPreprocessor(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int     `csv:"col1"`
		Col2 float32 `csv:"col2"`
	}

	t.Run("#1: trim all column before decoding", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1, 2.123
			100,200`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.TrimSpace = true
		}).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.totalRow)
		assert.Equal(t, []Item{{Col1: 1, Col2: 2.123}, {Col1: 100, Col2: 200}}, v)
	})

	t.Run("#2: trim specific column before decoding", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1, 2.123
			100,200`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.ConfigureColumn("col2", func(cfg *DecodeColumnConfig) {
				cfg.TrimSpace = true
			})
		}).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item{{Col1: 1, Col2: 2.123}, {Col1: 100, Col2: 200}}, v)
	})

	t.Run("#3: number un-comma before decoding", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1, 2.123
			100," 200,123.45"`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.ConfigureColumn("col2", func(cfg *DecodeColumnConfig) {
				cfg.PreprocessorFuncs = []ProcessorFunc{ProcessorTrim, ProcessorNumberUngroupComma}
			})
		}).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item{{Col1: 1, Col2: 2.123}, {Col1: 100, Col2: 200123.45}}, v)
	})

	t.Run("#4: error parsing", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1, 2.123
			100,200 `)

		var v []Item
		_, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.ErrorIs(t, err, ErrDecodeValueType)
	})
}

func Test_Decode_withValidator(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int16   `csv:"col1,omitempty"`
		Col2 StrType `csv:"col2"`
	}

	t.Run("#1: validate number range", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			1000,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.ConfigureColumn("col1", func(cfg *DecodeColumnConfig) {
				cfg.ValidatorFuncs = []ValidatorFunc{ValidatorRange(int16(0), int16(1000))}
			})
		}).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item{{Col1: 1, Col2: "abcxyz123"}, {Col1: 1000, Col2: "abc123"}}, v)
	})

	t.Run("#2: validate number range with omitempty", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			,abcxyz123
			1000,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.ConfigureColumn("col1", func(cfg *DecodeColumnConfig) {
				cfg.ValidatorFuncs = []ValidatorFunc{ValidatorRange(int16(1), int16(1000))}
			})
		}).Decode(&v)
		assert.Equal(t, 3, ret.TotalRow())
		assert.ErrorIs(t, err, ErrValidationRange)
	})

	t.Run("#3: validate number range and str length range", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			1000,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.StopOnError = false
			cfg.ConfigureColumn("col1", func(cfg *DecodeColumnConfig) {
				cfg.ValidatorFuncs = []ValidatorFunc{ValidatorRange(int16(0), int16(999))}
			})
			cfg.ConfigureColumn("col2", func(cfg *DecodeColumnConfig) {
				cfg.ValidatorFuncs = []ValidatorFunc{ValidatorStrLen[StrType](5, 7)}
			})
		}).Decode(&v)
		assert.Nil(t, v)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, 2, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrValidationRange)
		assert.ErrorIs(t, err, ErrValidationStrLen)
	})
}

func Test_Decode_multipleCalls(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int     `csv:"col1"`
		Col2 float32 `csv:"col2"`
	}

	t.Run("#1: 2nd call will fail", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,2.123
			100,200`)

		var v []Item
		d := NewDecoder(csv.NewReader(strings.NewReader(data)))
		ret, err := d.Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []string{"ColX"}, ret.MissingOptionalColumns())
		assert.Equal(t, []Item{{Col1: 1, Col2: 2.123}, {Col1: 100, Col2: 200}}, v)

		// Second call
		var v2 []Item
		ret, err = d.Decode(&v2)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrFinished)
	})

	t.Run("#2: 1st and 2nd call will fail", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,2.123
			100,abc`)

		var v []Item
		d := NewDecoder(csv.NewReader(strings.NewReader(data)))
		ret, err := d.Decode(&v)
		assert.NotNil(t, ret)
		assert.ErrorIs(t, err, ErrDecodeValueType)

		// Second call
		var v2 []Item
		ret, err = d.Decode(&v2)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrAlreadyFailed)
	})
}

func Test_Decode_withFixedInlineColumn(t *testing.T) {
	type Sub struct {
		ColZ bool `csv:",optional"`
		ColY bool
		Col1 int16  `csv:"sub1"`
		Col2 string `csv:"sub2,optional"`
	}
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int    `csv:"col1"`
		Sub1 Sub    `csv:"sub1,inline"`
		Col2 string `csv:"col2"`
	}
	type Item2 struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int    `csv:"col1"`
		Sub1 *Sub   `csv:"sub1,inline"`
		Col2 string `csv:"col2"`
	}

	t.Run("#1: success", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,sub1,col2
			1,111,abcxyz123
			1000,222,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item{{Col1: 1, Sub1: Sub{Col1: 111}, Col2: "abcxyz123"},
			{Col1: 1000, Sub1: Sub{Col1: 222}, Col2: "abc123"}}, v)
	})

	t.Run("#2: with column unordered", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2,sub1
			1,abcxyz123,111
			1000,abc123,222`)

		var v []Item2
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.RequireColumnOrder = false
		}).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item2{{Col1: 1, Sub1: &Sub{Col1: 111}, Col2: "abcxyz123"},
			{Col1: 1000, Sub1: &Sub{Col1: 222}, Col2: "abc123"}}, v)
	})

	t.Run("#3: with no header mode", func(t *testing.T) {
		data := gofn.MultilineString(
			`1,111,abcxyz123
			1000,222,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.NoHeaderMode = true
		}).Decode(&v)
		assert.Nil(t, ret)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrHeaderDynamicNotAllowNoHeaderMode)
	})

	t.Run("#4: with column prefix", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2,sub_sub1
			1,abcxyz123,111
			1000,abc123,222`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int    `csv:"col1"`
			Sub1 Sub    `csv:"sub1,inline,prefix=sub_"`
			Col2 string `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.RequireColumnOrder = false
		}).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item{{Col1: 1, Sub1: Sub{Col1: 111}, Col2: "abcxyz123"},
			{Col1: 1000, Sub1: Sub{Col1: 222}, Col2: "abc123"}}, v)
	})

	t.Run("#5: invalid inline column", func(t *testing.T) {
		data := gofn.MultilineString(
			`1,111,abcxyz123
			1000,222,abc123`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int    `csv:"col1,inline"`
			Col2 string `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.NoHeaderMode = true
		}).Decode(&v)
		assert.Nil(t, ret)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
	})

	t.Run("#6: with prefix and custom preprocessor/validator", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2,sub_sub1,sub_sub2
			1,abcxyz123,111,abc123
			1000,abc123,222,xyz`)

		type Item struct {
			Col1 int    `csv:"col1"`
			Sub1 Sub    `csv:"sub1,inline,prefix=sub_"`
			Col2 string `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.StopOnError = false
			cfg.RequireColumnOrder = false
			cfg.ConfigureColumn("sub_sub1", func(cfg *DecodeColumnConfig) {
				cfg.PreprocessorFuncs = []ProcessorFunc{ProcessorTrim}
				cfg.ValidatorFuncs = []ValidatorFunc{ValidatorRange(int16(0), 100)}
			})
			cfg.ConfigureColumn("sub1", func(cfg *DecodeColumnConfig) {
				cfg.PreprocessorFuncs = []ProcessorFunc{ProcessorTrim}
				cfg.ValidatorFuncs = []ValidatorFunc{ValidatorStrLen[string](0, 5)}
			})
		}).Decode(&v)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, 3, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrValidationRange)
		assert.ErrorIs(t, err, ErrValidationStrLen)
	})
}

func Test_Decode_withDynamicInlineColumn(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int               `csv:"col1"`
		Sub1 InlineColumn[int] `csv:"sub1,inline"`
		Col2 *string           `csv:"col2"`
		ColZ bool              `csv:",optional"`
	}

	t.Run("#1: success", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,sub1,sub2,col2
			1,111,11,abcxyz123
			1000,222,22,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		header := []string{"sub1", "sub2"}
		assert.Equal(t, []Item{
			{Col1: 1, Sub1: InlineColumn[int]{Header: header, Values: []int{111, 11}}, Col2: gofn.New("abcxyz123")},
			{Col1: 1000, Sub1: InlineColumn[int]{Header: header, Values: []int{222, 22}}, Col2: gofn.New("abc123")},
		}, v)
	})

	t.Run("#2: with column unordered", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2,sub1,sub2
			1,abcxyz123,111,11
			1000,abc123,222,22`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.RequireColumnOrder = false
		}).Decode(&v)
		assert.Nil(t, ret)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrHeaderDynamicRequireColumnOrder)
	})

	t.Run("#3: with no header mode", func(t *testing.T) {
		data := gofn.MultilineString(
			`1,abcxyz123,111,11
			1000,abc123,222,22`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.NoHeaderMode = true
		}).Decode(&v)
		assert.Nil(t, ret)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrHeaderDynamicNotAllowNoHeaderMode)
	})

	t.Run("#4: with column prefix", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,sub_sub1,sub_sub2,col2
			1,111,11,abcxyz123
			1000,222,22,abc123`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 *int              `csv:"col1"`
			Sub1 InlineColumn[int] `csv:"sub1,inline,prefix=sub_"`
			Col2 string            `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		header := []string{"sub_sub1", "sub_sub2"}
		assert.Equal(t, []Item{
			{Col1: gofn.New(1), Sub1: InlineColumn[int]{Header: header, Values: []int{111, 11}}, Col2: "abcxyz123"},
			{Col1: gofn.New(1000), Sub1: InlineColumn[int]{Header: header, Values: []int{222, 22}}, Col2: "abc123"},
		}, v)
	})

	t.Run("#5: with custom preprocessor/validator", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,sub1,sub2,col2
			1,111, 11,abcxyz123
			1000,222 ,22,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.StopOnError = false
			cfg.ConfigureColumn("sub1", func(cfg *DecodeColumnConfig) {
				cfg.PreprocessorFuncs = []ProcessorFunc{ProcessorTrim}
				cfg.ValidatorFuncs = []ValidatorFunc{ValidatorRange(0, 100)}
			})
		}).Decode(&v)
		assert.Equal(t, 2, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrValidationRange)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Nil(t, v)
	})

	t.Run("#4: with prefix and custom preprocessor/validator", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,sub_sub1,sub_sub2,col2
			1,111, 11,abcxyz123
			1000,222 ,22,abc123`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int               `csv:"col1"`
			Sub1 InlineColumn[int] `csv:"sub1,inline,prefix=sub_"`
			Col2 string            `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.StopOnError = false
			cfg.ConfigureColumn("sub_sub1", func(cfg *DecodeColumnConfig) {
				cfg.PreprocessorFuncs = []ProcessorFunc{ProcessorTrim}
				cfg.ValidatorFuncs = []ValidatorFunc{ValidatorRange(0, 100)}
			})
		}).Decode(&v)
		assert.Equal(t, 2, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrValidationRange)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Nil(t, v)
	})

	t.Run("#5: invalid dynamic inline type", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,sub1,sub2,col2
			1,111,11,abcxyz123
			1000,222,22,abc123`)

		type InlineColumn2[T any] struct {
			Header []string
		}
		type Item struct {
			Col1 int                `csv:"col1"`
			Sub1 InlineColumn2[int] `csv:"sub1,inline"`
			Col2 string             `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
		assert.Nil(t, ret)
		assert.Nil(t, v)
	})

	t.Run("#6: invalid dynamic inline type", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,sub1,sub2,col2
			1,111,11,abcxyz123
			1000,222,22,abc123`)

		type InlineColumn2[T any] struct {
			Header []string
			Values *[]T
		}
		type Item struct {
			Col1 int                `csv:"col1"`
			Sub1 InlineColumn2[int] `csv:"sub1,inline"`
			Col2 string             `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.ErrorIs(t, err, ErrHeaderDynamicTypeInvalid)
		assert.Nil(t, ret)
		assert.Nil(t, v)
	})
}

func Test_Decode_withLocalization(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int16   `csv:"col1"`
		Col2 StrType `csv:"col2"`
	}

	t.Run("#1: localization fails", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			1000,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.ParseLocalizedHeader = true
			cfg.LocalizationFunc = localizeFail
		}).Decode(&v)
		assert.Nil(t, ret)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrLocalization)
		assert.ErrorIs(t, err, errKeyNotFound)
	})

	t.Run("#2: map localization key for error", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			100000,abc123`)

		var v []Item
		_, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.ConfigureColumn("col1", func(cfg *DecodeColumnConfig) {
				cfg.OnCellErrorFunc = func(e *CellError) {
					if errors.Is(e, ErrDecodeValueType) {
						e.SetLocalizationKey("LOCALE_KEY_1")
					}
				}
			})
		}).Decode(&v)
		assert.Nil(t, v)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrDecodeValueType)
		assert.Equal(t, "LOCALE_KEY_1",
			err.(*Errors).Unwrap()[0].(*RowErrors).Unwrap()[0].(*CellError).LocalizationKey())
	})
}

func Test_Decode_withCustomUnmarshaler(t *testing.T) {
	t.Run("#1: no decode func matching", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2,col3
			1,abcxyz123,a
			1000,abc123,b`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int            `csv:"col1"`
			Col2 float32        `csv:"col2"`
			Col3 map[int]string `csv:"col3"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrTypeUnsupported)
	})

	t.Run("#2: ptr of type implements UnmarshalText/UnmarshalCSV", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2,col3
			1,abcxyz123,AA
			1000,abc123,Bb`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int          `csv:"col1"`
			Col2 StrUpperType `csv:"col2"`
			Col3 StrLowerType `csv:"col3"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item{
			{Col1: 1, Col2: "ABCXYZ123", Col3: "aa"},
			{Col1: 1000, Col2: "ABC123", Col3: "bb"},
		}, v)
	})

	t.Run("#3: type implements UnmarshalText/UnmarshalCSV", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2,col3
			1,abcxyz123,AA
			1000,abc123,Bb`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int           `csv:"col1"`
			Col2 *StrUpperType `csv:"col2"`
			Col3 *StrLowerType `csv:"col3"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item{
			{Col1: 1, Col2: gofn.New[StrUpperType]("ABCXYZ123"), Col3: gofn.New[StrLowerType]("aa")},
			{Col1: 1000, Col2: gofn.New[StrUpperType]("ABC123"), Col3: gofn.New[StrLowerType]("bb")},
		}, v)
	})
}

func Test_Decode_specialCases(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int     `csv:"col1"`
		Col2 float32 `csv:"col2"`
	}

	t.Run("#1: no input data", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 1, ret.TotalRow())
		assert.Equal(t, []string{"ColX"}, ret.MissingOptionalColumns())
		assert.Equal(t, 0, len(v))
	})

	t.Run("#2: no header mode", func(t *testing.T) {
		data := gofn.MultilineString(
			`1,2.2
			100,3.3`)

		type Item struct {
			ColY bool
			Col1 int     `csv:"col1"`
			Col2 float32 `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.NoHeaderMode = true
		}).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 2, ret.TotalRow())
		assert.Equal(t, []Item{{Col1: 1, Col2: 2.2}, {Col1: 100, Col2: 3.3}}, v)
	})

	t.Run("#3: duplicated column from struct", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			1000,abc123`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int     `csv:"col1"`
			Col2 float32 `csv:"col2"`
			Col3 int     `csv:"col1"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrHeaderColumnDuplicated)
	})

	t.Run("#4: duplicated column from input", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2,col1
			1,abcxyz123,a
			1000,abc123,b`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrHeaderColumnDuplicated)
	})

	t.Run("#5: invalid header (contains space)", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,abcxyz123
			1000,abc123`)

		type Item3 struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int     `csv:"col1 "`
			Col2 float32 `csv:" col2"`
		}

		var v []Item3
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrHeaderColumnInvalid)
	})

	t.Run("#6: header order unmatched", func(t *testing.T) {
		data := gofn.MultilineString(
			`col2,col1
			1,abcxyz123
			1000,abc123`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrHeaderColumnOrderInvalid)
	})

	t.Run("#7: item type is pointer", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,2.123
			100,200`)

		var v []*Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []string{"ColX"}, ret.MissingOptionalColumns())
		assert.Equal(t, []*Item{{Col1: 1, Col2: 2.123}, {Col1: 100, Col2: 200}}, v)
	})
}

func Test_Decode_specialTypes(t *testing.T) {
	t.Run("#1: interface{} type", func(t *testing.T) {
		data := gofn.MultilineString(`col1,col2
			1,2.2
			100,3.3`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int         `csv:"col1"`
			Col2 interface{} `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item{{Col1: 1, Col2: "2.2"}, {Col1: 100, Col2: "3.3"}}, v)
	})

	t.Run("#2: ptr interface{} type", func(t *testing.T) {
		data := gofn.MultilineString(`col1,col2
			1,2.2
			100,3.3`)

		type Item struct {
			ColX bool `csv:",optional"`
			ColY bool
			Col1 int          `csv:"col1"`
			Col2 *interface{} `csv:"col2"`
		}

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, []Item{{Col1: 1, Col2: gofn.New[interface{}]("2.2")},
			{Col1: 100, Col2: gofn.New[interface{}]("3.3")}}, v)
	})

	t.Run("#3: all base types", func(t *testing.T) {
		data := gofn.MultilineString(
			`c1,c2,c3,c4,c5,c6,c7,c8,c9,c10,c11,c12,c13,c14,c15,c16,c17,c18,c19,c20,c21,c22,c23,c24,c25,c26,c27,c28
			-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,false,false,abc,123
			`)

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

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, err)
		assert.Equal(t, 2, ret.TotalRow())
		assert.Equal(t, []Item{
			{-1, gofn.New(-1), int8(-1), gofn.New(int8(-1)), int16(-1), gofn.New(int16(-1)),
				int32(-1), gofn.New(int32(-1)), int64(-1), gofn.New(int64(-1)), uint(1), gofn.New(uint(1)),
				uint8(1), gofn.New(uint8(1)), uint16(1), gofn.New(uint16(1)), uint32(1), gofn.New(uint32(1)),
				uint64(1), gofn.New(uint64(1)), float32(1), gofn.New(float32(1)), float64(1), gofn.New(float64(1)),
				false, gofn.New(false), "abc", gofn.New("123")},
		}, v)
	})
}

func Test_Decode_incorrectStructure(t *testing.T) {
	type Item struct {
		ColX bool `csv:",optional"`
		ColY bool
		Col1 int     `csv:"col1"`
		Col2 float32 `csv:"col2"`
	}

	t.Run("#1: row field count not match header", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,1.1
			1000,2.2,invalid,
			2,2.2,abc,123
			3`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.ErrorIs(t, err, ErrDecodeRowFieldCount)
	})

	t.Run("#2: row field count not match header (TreatAsError = false)", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,1.1
			1000,2.2,invalid,
			2,2.2,abc,123
			3`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.TreatIncorrectStructureAsError = false
			cfg.StopOnError = false
		}).Decode(&v)
		assert.Equal(t, 5, ret.TotalRow())
		assert.Equal(t, 3, err.(*Errors).TotalError())
		assert.ErrorIs(t, err.(*Errors).Unwrap()[0], ErrDecodeRowFieldCount)
		assert.ErrorIs(t, err.(*Errors).Unwrap()[1], ErrDecodeRowFieldCount)
		assert.ErrorIs(t, err.(*Errors).Unwrap()[2], ErrDecodeRowFieldCount)
	})

	t.Run("#3: invalid field quote", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,1.1
			"1000"",2.2,
			2,2.2""`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data))).Decode(&v)
		assert.Nil(t, ret)
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrDecodeQuoteInvalid)
	})

	t.Run("#4: invalid field quote (TreatAsError = false)", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,1.1
			"1000"",2.2
			2,2.2
			3,3.3`)

		var v []Item
		ret, err := NewDecoder(csv.NewReader(strings.NewReader(data)), func(cfg *DecodeConfig) {
			cfg.TreatIncorrectStructureAsError = false
			cfg.StopOnError = false
		}).Decode(&v)
		assert.Equal(t, 3, ret.TotalRow())
		assert.Equal(t, 1, err.(*Errors).TotalError())
		assert.ErrorIs(t, err, ErrDecodeQuoteInvalid)
	})
}

func Test_DecodeOne(t *testing.T) {
	type Item struct {
		ColX bool          `csv:",optional"`
		ColY bool          `csv:"-"`
		Col1 int           `csv:"col1"`
		Col2 float32       `csv:"col2"`
		Col3 StrUpperType  `csv:"col3,optional"`
		Col4 *StrLowerType `csv:"col4,optional"`
	}

	t.Run("#1: decode one until finishes", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,2.123
			100,200`)

		var v1, v2, v3 Item
		d := NewDecoder(csv.NewReader(strings.NewReader(data)))
		err := d.DecodeOne(&v1)
		assert.Nil(t, err)
		assert.Equal(t, Item{Col1: 1, Col2: 2.123}, v1)
		err = d.DecodeOne(&v2)
		assert.Nil(t, err)
		assert.Equal(t, Item{Col1: 100, Col2: 200}, v2)
		err = d.DecodeOne(&v3)
		assert.ErrorIs(t, err, ErrFinished)
	})

	t.Run("#2: using nil ptr as input", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,2.123
			100,200`)

		var v1, v2, v3 *Item
		d := NewDecoder(csv.NewReader(strings.NewReader(data)))
		err := d.DecodeOne(v1)
		assert.ErrorIs(t, err, ErrValueNil)
		err = d.DecodeOne(v2)
		assert.ErrorIs(t, err, ErrValueNil)
		err = d.DecodeOne(v3)
		assert.ErrorIs(t, err, ErrValueNil)
	})

	t.Run("#4: invalid input type", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,2.123
			100,200`)

		var v1 []string
		d := NewDecoder(csv.NewReader(strings.NewReader(data)))
		err := d.DecodeOne(v1)
		assert.ErrorIs(t, err, ErrTypeInvalid)
		var v2 int
		err = d.DecodeOne(&v2)
		assert.ErrorIs(t, err, ErrTypeInvalid)
		var v3 Item
		err = d.DecodeOne(v3)
		assert.ErrorIs(t, err, ErrTypeInvalid)
	})

	t.Run("#5: call decode when finished", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,2.123
			100,200`)

		var v1, v2 Item
		d := NewDecoder(csv.NewReader(strings.NewReader(data)))
		err := d.DecodeOne(&v1)
		assert.Nil(t, err)
		assert.Equal(t, Item{Col1: 1, Col2: 2.123}, v1)
		_, _ = d.Finish()
		err = d.DecodeOne(&v2)
		assert.ErrorIs(t, err, ErrFinished)
	})

	t.Run("#6: pass different types between calls", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2
			1,2.123
			100,200`)
		type Item2 struct {
			Col1 int     `csv:"col1"`
			Col2 float32 `csv:"col2"`
		}

		var v1 Item
		var v2 Item2
		d := NewDecoder(csv.NewReader(strings.NewReader(data)))
		err := d.DecodeOne(&v1)
		assert.Nil(t, err)
		assert.Equal(t, Item{Col1: 1, Col2: 2.123}, v1)
		err = d.DecodeOne(&v2)
		assert.ErrorIs(t, err, ErrTypeUnmatched)
	})

	t.Run("#7: no input data", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2`)

		var v Item
		d := NewDecoder(csv.NewReader(strings.NewReader(data)))
		err := d.DecodeOne(&v)
		assert.ErrorIs(t, err, ErrFinished)
	})

	t.Run("#8: decode one until finishes with unmarshaler", func(t *testing.T) {
		data := gofn.MultilineString(
			`col1,col2,col3,col4
			1,2.123,AAa,AaA
			100,200,bbB,bBb`)

		var v1, v2, v3 Item
		d := NewDecoder(csv.NewReader(strings.NewReader(data)))
		err := d.DecodeOne(&v1)
		assert.Nil(t, err)
		assert.Equal(t, Item{Col1: 1, Col2: 2.123, Col3: "AAA", Col4: gofn.New[StrLowerType]("aaa")}, v1)
		err = d.DecodeOne(&v2)
		assert.Nil(t, err)
		assert.Equal(t, Item{Col1: 100, Col2: 200, Col3: "BBB", Col4: gofn.New[StrLowerType]("bbb")}, v2)
		err = d.DecodeOne(&v3)
		assert.ErrorIs(t, err, ErrFinished)
	})
}

func Test_parseColumnDetailsFromStructType(t *testing.T) {
	type Item struct {
		Col0 InlineColumn[int64]  `csv:"dynA,inline"`
		Col1 int                  `csv:"col1,optional"`
		Col2 *int                 `csv:"col2,omitempty"`
		Col3 string               `csv:"-"`
		Col4 string               `csv:""`
		Col5 InlineColumn[int]    `csv:"dynB,inline"`
		Col6 int                  `csv:"col6"`
		Col7 InlineColumn[string] `csv:"dynC,inline"`
	}
	structType := reflect.TypeOf(Item{})
	fileHeader := []string{"dyn1", "col1", "col2", "Col4", "dyn2", "dyn3", "col6", "dyn4", "col7"}

	t.Run("#1: success", func(t *testing.T) {
		colDetails, err := NewDecoder(nil).parseColumnsMetaFromStructType(structType, fileHeader)
		assert.Nil(t, err)
		parsedHeader := gofn.MapSlice(colDetails, func(v *decodeColumnMeta) string { return v.headerText })
		assert.Equal(t, fileHeader, parsedHeader)
	})

	t.Run("#2: config invalid NoHeaderMode", func(t *testing.T) {
		cfg := defaultDecodeConfig()
		cfg.NoHeaderMode = true
		_, err := NewDecoder(nil, func(cfg *DecodeConfig) {
			cfg.NoHeaderMode = true
		}).parseColumnsMetaFromStructType(structType, fileHeader)
		assert.ErrorIs(t, err, ErrHeaderDynamicNotAllowNoHeaderMode)
	})
	t.Run("#3: config invalid not RequireColumnOrder", func(t *testing.T) {
		_, err := NewDecoder(nil, func(cfg *DecodeConfig) {
			cfg.RequireColumnOrder = false
		}).parseColumnsMetaFromStructType(structType, fileHeader)
		assert.ErrorIs(t, err, ErrHeaderDynamicRequireColumnOrder)
	})
	t.Run("#4: config invalid AllowUnrecognizedColumns", func(t *testing.T) {
		_, err := NewDecoder(nil, func(cfg *DecodeConfig) {
			cfg.AllowUnrecognizedColumns = true
		}).parseColumnsMetaFromStructType(structType, fileHeader)
		assert.ErrorIs(t, err, ErrHeaderDynamicNotAllowUnrecognizedColumns)
	})
	t.Run("#5: config invalid ParseLocalizedHeader", func(t *testing.T) {
		_, err := NewDecoder(nil, func(cfg *DecodeConfig) {
			cfg.ParseLocalizedHeader = true
			cfg.LocalizationFunc = func(k string, params ParameterMap) (string, error) { return k, nil }
		}).parseColumnsMetaFromStructType(structType, fileHeader)
		assert.ErrorIs(t, err, ErrHeaderDynamicNotAllowLocalizedHeader)
	})
}
