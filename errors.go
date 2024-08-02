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

// Errors represents errors returned by the encoder or decoder
type Errors struct { // nolint: errname
	errs     []error
	totalRow int
	header   []string
}

// NewErrors creates a new Errors object
func NewErrors() *Errors {
	return &Errors{}
}

// TotalRow gets total rows of CSV data
func (e *Errors) TotalRow() int {
	return e.totalRow
}

// Header gets list of column headers
func (e *Errors) Header() []string {
	return e.header
}

// Error implements Go error interface
func (e *Errors) Error() string {
	return getErrorMsg(e.errs)
}

// HasError checks if there is at least one error in the list
func (e *Errors) HasError() bool {
	return len(e.errs) > 0
}

// TotalRowError gets the total number of error of rows
func (e *Errors) TotalRowError() int {
	c := 0
	for _, e := range e.errs {
		if _, ok := e.(*RowErrors); ok { // nolint: errorlint
			c++
		}
	}
	return c
}

// TotalCellError gets the total number of error of cells
func (e *Errors) TotalCellError() int {
	c := 0
	for _, e := range e.errs {
		if rowErr, ok := e.(*RowErrors); ok { // nolint: errorlint
			c += rowErr.TotalCellError()
		}
	}
	return c
}

// TotalError gets the total number of errors including row errors and cell errors
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

// Add appends errors to the list
func (e *Errors) Add(errs ...error) {
	e.errs = append(e.errs, errs...)
}

// Is checks if there is at least an error in the list kind of the specified error
func (e *Errors) Is(err error) bool {
	for _, er := range e.errs {
		if errors.Is(er, err) {
			return true
		}
	}
	return false
}

// Unwrap implements Go error unwrap function
func (e *Errors) Unwrap() []error {
	return e.errs
}

// RowErrors data structure of error of a row
type RowErrors struct { // nolint: errname
	errs []error
	row  int
	line int
}

// NewRowErrors creates a new RowErrors
func NewRowErrors(row, line int) *RowErrors {
	return &RowErrors{row: row, line: line}
}

// Row gets the row contains the error
func (e *RowErrors) Row() int {
	return e.row
}

// Line gets the line contains the error (line equals to row in most cases)
func (e *RowErrors) Line() int {
	return e.line
}

// Error implements Go error interface
func (e *RowErrors) Error() string {
	return getErrorMsg(e.errs)
}

// HasError checks if there is at least one error in the list
func (e *RowErrors) HasError() bool {
	return len(e.errs) > 0
}

// TotalError gets the total number of errors
func (e *RowErrors) TotalError() int {
	return len(e.errs)
}

// TotalCellError gets the total number of error of cells
func (e *RowErrors) TotalCellError() int {
	c := 0
	for _, e := range e.errs {
		if _, ok := e.(*CellError); ok { // nolint: errorlint
			c++
		}
	}
	return c
}

// Add appends errors to the list
func (e *RowErrors) Add(errs ...error) {
	e.errs = append(e.errs, errs...)
}

// Is checks if there is at least an error in the list kind of the specified error
func (e *RowErrors) Is(err error) bool {
	for _, er := range e.errs {
		if errors.Is(er, err) {
			return true
		}
	}
	return false
}

// Unwrap implements Go error unwrap function
func (e *RowErrors) Unwrap() []error {
	return e.errs
}

// CellError data structure of error of a cell
type CellError struct {
	err             error
	fields          map[string]any
	localizationKey string

	column int
	header string
	value  string
}

// NewCellError creates a new CellError
func NewCellError(err error, column int, header string) *CellError {
	return &CellError{err: err, column: column, header: header, fields: map[string]any{}}
}

// Error implements Go error interface
func (e *CellError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

// Column gets the column of the cell
func (e *CellError) Column() int {
	return e.column
}

// Header gets the header of the column
func (e *CellError) Header() string {
	return e.header
}

// Value gets the value of the cell
func (e *CellError) Value() string {
	return e.value
}

// HasError checks if the error contains an error
func (e *CellError) HasError() bool {
	return e.err != nil
}

// Is checks if the inner error is kind of the specified error
func (e *CellError) Is(err error) bool {
	return errors.Is(e.err, err)
}

// Unwrap implements Go error unwrap function
func (e *CellError) Unwrap() error {
	return e.err
}

// WithParam sets a param of error
func (e *CellError) WithParam(k string, v any) *CellError {
	e.fields[k] = v
	return e
}

// LocalizationKey gets localization key of error
func (e *CellError) LocalizationKey() string {
	return e.localizationKey
}

// SetLocalizationKey sets localization key of error
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
