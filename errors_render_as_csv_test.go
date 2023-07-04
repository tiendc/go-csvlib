package csvlib

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiendc/gofn"
)

func Test_ErrorRenderAsCSV(t *testing.T) {
	// CSV error has 2 row errors
	csvErr := NewErrors()
	csvErr.totalRow = 200
	csvErr.header = []string{"Name", "Age", "Address"}

	rowErr1 := NewRowErrors(10, 12)
	rowErr2 := NewRowErrors(20, 22)
	csvErr.Add(rowErr1, rowErr2)

	// First row error has 2 cell errors and an unexpected error
	cellErr11 := NewCellError(ErrValidationStrLen, 0, "Name")
	cellErr11.SetLocalizationKey("ERR_NAME_TOO_LONG")
	cellErr11.value = "David David David"
	_ = cellErr11.WithParam("MinLen", 1).WithParam("MaxLen", 10)

	cellErr12 := NewCellError(ErrValidationRange, 1, "Age")
	cellErr12.SetLocalizationKey("ERR_AGE_OUT_OF_RANGE")
	cellErr12.value = "101"
	_ = cellErr12.WithParam("MinValue", 1).WithParam("MaxValue", 100)

	cellErr13 := NewCellError(ErrDecodeQuoteInvalid, -1, "") // error not relate to any column
	rowErr1.Add(cellErr11, cellErr12, cellErr13)

	// Second row error has 2 other cell errors
	cellErr21 := NewCellError(ErrValidationStrLen, 0, "Name")
	cellErr22 := NewCellError(ErrValidationRange, 1, "Age")
	rowErr2.Add(cellErr21, cellErr22)

	// An unexpected error
	csvErr.Add(ErrTypeUnsupported)

	t.Run("#1: default rendering", func(t *testing.T) {
		r, err := NewCSVRenderer(csvErr)
		assert.Nil(t, err)
		msg, _, err := r.RenderAsString()
		assert.Nil(t, err)
		assert.Equal(t, gofn.MultilineString(
			`Row,Line,CommonError,Name,Age,Address
			10,12,ErrDecodeQuoteInvalid,ERR_NAME_TOO_LONG,ERR_AGE_OUT_OF_RANGE,
			20,22,,ErrValidation: StrLen,ErrValidation: Range,
			`), msg)
	})

	t.Run("#2: translate en_US", func(t *testing.T) {
		r, err := NewCSVRenderer(csvErr, func(cfg *CSVRenderConfig) {
			cfg.LocalizationFunc = localizeEnUs
		})
		assert.Nil(t, err)
		msg, _, err := r.RenderAsString()
		assert.Nil(t, err)
		// nolint: lll
		assert.Equal(t, gofn.MultilineString(
			`Row,Line,CommonError,Name,Age,Address
			10,12,ErrDecodeQuoteInvalid,'David David David' at column 0 - Name length must be from 1 to 10,'101' at column 1 - Age must be from 1 to 100,
			20,22,,ErrValidation: StrLen,ErrValidation: Range,
			`), msg)
	})

	t.Run("#3: translate vi_VN", func(t *testing.T) {
		r, err := NewCSVRenderer(csvErr, func(cfg *CSVRenderConfig) {
			cfg.LocalizationFunc = localizeViVn
			cfg.CellRenderFunc = func(rowErr *RowErrors, cellErr *CellError, params ParameterMap) (string, bool) {
				if errors.Is(cellErr, ErrDecodeQuoteInvalid) {
					return "nội dung bị bao sai (quote)", true
				}
				return "", true
			}
			cfg.RenderRowNumberColumnIndex = 1
			cfg.RenderCommonErrorColumnIndex = 0
			cfg.RenderLineNumberColumnIndex = -1
			cfg.HeaderRenderFunc = func(header []string, params ParameterMap) {
				header[cfg.RenderRowNumberColumnIndex] = "ROW"
				header[cfg.RenderCommonErrorColumnIndex] = "COMMONERR"
			}
		})
		assert.Nil(t, err)
		msg, _, err := r.RenderAsString()
		assert.Nil(t, err)
		// nolint: lll
		assert.Equal(t, gofn.MultilineString(
			`COMMONERR,ROW,Name,Age,Address
			nội dung bị bao sai (quote),10,'David David David' at column 0 - Tên phải dài từ 1 đến 10 ký tự,'101' at column 1 - Tuổi phải từ 1 đến 100,
			,20,ErrValidation: StrLen,ErrValidation: Range,
			`), msg)
	})
}
