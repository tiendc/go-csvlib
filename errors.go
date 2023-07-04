package csvlib

import (
	"errors"
	"fmt"
)

var (
	ErrTypeInvalid     = errors.New("ErrTypeInvalid")
	ErrTypeUnsupported = errors.New("ErrTypeUnsupported")
	ErrTypeUnmatched   = errors.New("ErrTypeUnmatched")
	ErrValueNil        = errors.New("ErrValueNil")
	ErrAlreadyFailed   = errors.New("ErrAlreadyFailed")
	ErrFinished        = errors.New("ErrFinished")
	ErrUnexpected      = errors.New("ErrUnexpected")

	ErrTagOptionInvalid    = errors.New("ErrTagOptionInvalid")
	ErrConfigOptionInvalid = errors.New("ErrConfigOptionInvalid")
	ErrLocalization        = errors.New("ErrLocalization")

	ErrHeaderColumnInvalid                      = errors.New("ErrHeaderColumnInvalid")
	ErrHeaderColumnUnrecognized                 = errors.New("ErrHeaderColumnUnrecognized")
	ErrHeaderColumnRequired                     = errors.New("ErrHeaderColumnRequired")
	ErrHeaderColumnDuplicated                   = errors.New("ErrHeaderColumnDuplicated")
	ErrHeaderColumnOrderInvalid                 = errors.New("ErrHeaderColumnOrderInvalid")
	ErrHeaderDynamicTypeInvalid                 = errors.New("ErrHeaderDynamicTypeInvalid")
	ErrHeaderDynamicNotAllowNoHeaderMode        = errors.New("ErrHeaderDynamicNotAllowNoHeaderMode")
	ErrHeaderDynamicRequireColumnOrder          = errors.New("ErrHeaderDynamicRequireColumnOrder")
	ErrHeaderDynamicNotAllowUnrecognizedColumns = errors.New("ErrHeaderDynamicNotAllowUnrecognizedColumns")
	ErrHeaderDynamicNotAllowLocalizedHeader     = errors.New("ErrHeaderDynamicNotAllowLocalizedHeader")

	ErrValidationConversion = errors.New("ErrValidationConversion")
	ErrValidation           = errors.New("ErrValidation")
	ErrValidationLT         = fmt.Errorf("%w: LT", ErrValidation)
	ErrValidationLTE        = fmt.Errorf("%w: LTE", ErrValidation)
	ErrValidationGT         = fmt.Errorf("%w: GT", ErrValidation)
	ErrValidationGTE        = fmt.Errorf("%w: GTE", ErrValidation)
	ErrValidationRange      = fmt.Errorf("%w: Range", ErrValidation)
	ErrValidationIN         = fmt.Errorf("%w: IN", ErrValidation)
	ErrValidationStrLen     = fmt.Errorf("%w: StrLen", ErrValidation)
	ErrValidationStrPrefix  = fmt.Errorf("%w: StrPrefix", ErrValidation)
	ErrValidationStrSuffix  = fmt.Errorf("%w: StrSuffix", ErrValidation)

	ErrDecodeValueType     = errors.New("ErrDecodeValueType")
	ErrDecodeRowFieldCount = errors.New("ErrDecodeRowFieldCount")
	ErrDecodeQuoteInvalid  = errors.New("ErrDecodeQuoteInvalid")

	ErrEncodeValueType = errors.New("ErrEncodeValueType")
)

type Errors struct { // nolint: errname
	errs     []error
	totalRow int
	header   []string
}

func NewErrors() *Errors {
	return &Errors{}
}

func (e *Errors) TotalRow() int {
	return e.totalRow
}

func (e *Errors) Header() []string {
	return e.header
}

func (e *Errors) Error() string {
	return getErrorMsg(e.errs)
}

func (e *Errors) HasError() bool {
	return len(e.errs) > 0
}

func (e *Errors) TotalRowError() int {
	c := 0
	for _, e := range e.errs {
		if _, ok := e.(*RowErrors); ok { // nolint: errorlint
			c++
		}
	}
	return c
}

func (e *Errors) TotalCellError() int {
	c := 0
	for _, e := range e.errs {
		if rowErr, ok := e.(*RowErrors); ok { // nolint: errorlint
			c += rowErr.TotalCellError()
		}
	}
	return c
}

func (e *Errors) TotalError() int {
	c := 0
	for _, e := range e.errs {
		if rowErr, ok := e.(*RowErrors); ok { // nolint: errorlint
			c += rowErr.TotalError()
		} else {
			c++
		}
	}
	return c
}

func (e *Errors) Add(errs ...error) {
	e.errs = append(e.errs, errs...)
}

func (e *Errors) Is(err error) bool {
	for _, er := range e.errs {
		if errors.Is(er, err) {
			return true
		}
	}
	return false
}

func (e *Errors) Unwrap() []error {
	return e.errs
}

type RowErrors struct { // nolint: errname
	errs []error
	row  int
	line int
}

func NewRowErrors(row, line int) *RowErrors {
	return &RowErrors{row: row, line: line}
}

func (e *RowErrors) Row() int {
	return e.row
}

func (e *RowErrors) Line() int {
	return e.line
}

func (e *RowErrors) Error() string {
	return getErrorMsg(e.errs)
}

func (e *RowErrors) HasError() bool {
	return len(e.errs) > 0
}

func (e *RowErrors) TotalError() int {
	return len(e.errs)
}

func (e *RowErrors) TotalCellError() int {
	c := 0
	for _, e := range e.errs {
		if _, ok := e.(*CellError); ok { // nolint: errorlint
			c++
		}
	}
	return c
}

func (e *RowErrors) Add(errs ...error) {
	e.errs = append(e.errs, errs...)
}

func (e *RowErrors) Is(err error) bool {
	for _, er := range e.errs {
		if errors.Is(er, err) {
			return true
		}
	}
	return false
}

func (e *RowErrors) Unwrap() []error {
	return e.errs
}

type CellError struct {
	err             error
	fields          map[string]interface{}
	localizationKey string

	column int
	header string
	value  string
}

func NewCellError(err error, column int, header string) *CellError {
	return &CellError{err: err, column: column, header: header, fields: map[string]interface{}{}}
}

func (e *CellError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *CellError) Column() int {
	return e.column
}

func (e *CellError) Header() string {
	return e.header
}

func (e *CellError) Value() string {
	return e.value
}

func (e *CellError) HasError() bool {
	return e.err != nil
}

func (e *CellError) Is(err error) bool {
	return errors.Is(e.err, err)
}

func (e *CellError) Unwrap() error {
	return e.err
}

func (e *CellError) WithParam(k string, v interface{}) *CellError {
	e.fields[k] = v
	return e
}

func (e *CellError) LocalizationKey() string {
	return e.localizationKey
}

func (e *CellError) SetLocalizationKey(k string) {
	e.localizationKey = k
}

func getErrorMsg(errs []error) string {
	s := ""
	for i, e := range errs {
		if i > 0 {
			s += ", "
		}
		s += e.Error()
	}
	return s
}
