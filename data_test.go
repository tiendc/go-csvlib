package csvlib

import (
	"errors"
	"fmt"
	"strings"
)

type StrType string

type StrUpperType string

func (s *StrUpperType) UnmarshalText(b []byte) error {
	*s = StrUpperType(strings.ToUpper(string(b)))
	return nil
}

func (s StrUpperType) MarshalText() ([]byte, error) {
	return []byte(strings.ToUpper(string(s))), nil
}

type StrLowerType string

func (s *StrLowerType) UnmarshalCSV(b []byte) error {
	*s = StrLowerType(strings.ToLower(string(b)))
	return nil
}

func (s StrLowerType) MarshalCSV() ([]byte, error) {
	return []byte(strings.ToLower(string(s))), nil
}

type SaleStats struct {
	ID            int               `csv:"id"`
	Name          string            `csv:"name"`
	Year          int               `csv:"year"`
	QuarterCount  QuarterSaleCount  `csv:"quarter_count,inline"`
	QuarterAmount QuarterSaleAmount `csv:"quarter_amount,inline"`
	TotalCount    int64             `csv:"total_count"`
	TotalAmount   int64             `csv:"total_amount"`
	Success       bool              `csv:"success"`
}

type QuarterSaleCount struct {
	Q1 int `csv:"q1-count"`
	Q2 int `csv:"q2-count"`
	Q3 int `csv:"q3-count"`
	Q4 int `csv:"q4-count"`
}

type QuarterSaleAmount struct {
	Q1 int64 `csv:"q1-amount"`
	Q2 int64 `csv:"q2-amount"`
	Q3 int64 `csv:"q3-amount"`
	Q4 int64 `csv:"q4-amount"`
}

type SaleStats2 struct {
	ID            int                 `csv:"id"`
	Name          string              `csv:"name"`
	Year          int                 `csv:"year"`
	QuarterCount  InlineColumn[int]   `csv:"quarter_count,inline"`
	QuarterAmount InlineColumn[int64] `csv:"quarter_amount,inline"`
	TotalCount    int64               `csv:"total_count"`
	TotalAmount   int64               `csv:"total_amount"`
	Success       bool                `csv:"success"`
}

var (
	mapLanguageEn = map[string]string{
		"col1": "col-1",
		"col2": "col-2",
		"col3": "col-3",
		"col4": "col-4",
		"col5": "col-5",
		"colX": "col-X",
		"colY": "col-Y",
		"colZ": "col-Z",
		"Col1": "Col-1",
		"Col2": "Col-2",
		"Col3": "Col-3",
		"Col4": "Col-4",
		"Col5": "Col-5",
		"ColX": "Col-X",
		"ColY": "Col-Y",
		"ColZ": "Col-Z",

		"sub":      "sub",
		"sub_col1": "sub-col-1",
		"sub_col2": "sub-col-2",
		"sub_col3": "sub-col-3",
		"sub_col4": "sub-col-4",
		"sub_col5": "sub-col-5",

		"ERR_HEADER_FORMAT": "Error content: TotalRow: {{.TotalRow}}, TotalRowError: {{.TotalRowError}}, " +
			"TotalCellError: {{.TotalCellError}}, TotalError: {{.TotalError}}",
		"ERR_ROW_FORMAT": "Row {{.Row}} (at line {{.Line}}): {{.ErrorContent}}",

		"ERR_NAME_TOO_LONG":    "'{{.Value}}' at column {{.Column}} - Name length must be from {{.MinLen}} to {{.MaxLen}}",
		"ERR_AGE_OUT_OF_RANGE": "'{{.Value}}' at column {{.Column}} - Age must be from {{.MinValue}} to {{.MaxValue}}",
		"ERR_AGE_INVALID":      "'{{.Value}}' at column {{.Column}} - Age must be a number",
	}

	mapLanguageVi = map[string]string{
		"col1": "cột-1",
		"col2": "cột-2",
		"col3": "cột-3",
		"col4": "cột-4",
		"col5": "cột-5",
		"colX": "cột-X",
		"colY": "cột-Y",
		"colZ": "cột-Z",
		"Col1": "Cột-1",
		"Col2": "Cột-2",
		"Col3": "Cột-3",
		"Col4": "Cột-4",
		"Col5": "Cột-5",
		"ColX": "Cột-X",
		"ColY": "Cột-Y",
		"ColZ": "Cột-Z",

		"sub":      "phụ",
		"sub_col1": "cột-phụ-1",
		"sub_col2": "cột-phụ-2",
		"sub_col3": "cột-phụ-3",
		"sub_col4": "cột-phụ-4",
		"sub_col5": "cột-phụ-5",

		"ERR_HEADER_FORMAT": "Error details: Số hàng: {{.TotalRow}}, Số hàng bị lỗi: {{.TotalRowError}}, " +
			"Số ô bị lỗi: {{.TotalCellError}}, Tổng lỗi: {{.TotalError}}",
		"ERR_ROW_FORMAT": "Hàng {{.Row}} (dòng {{.Line}}): {{.ErrorDetails}}",

		"ERR_NAME_TOO_LONG":    "'{{.Value}}' at column {{.Column}} - Tên phải dài từ {{.MinLen}} đến {{.MaxLen}} ký tự",
		"ERR_AGE_OUT_OF_RANGE": "'{{.Value}}' at column {{.Column}} - Tuổi phải từ {{.MinValue}} đến {{.MaxValue}}",
		"ERR_AGE_INVALID":      "'{{.Value}}' at column {{.Column}} - Tuổi phải là dạng số",
	}

	errKeyNotFound = errors.New("key not found")
)

func localizeViVn(k string, params ParameterMap) (string, error) {
	s, ok := mapLanguageVi[k]
	if !ok {
		return "", fmt.Errorf("%w: '%s'", errKeyNotFound, k)
	}
	return processTemplate(s, params)
}

func localizeEnUs(k string, params ParameterMap) (string, error) {
	s, ok := mapLanguageEn[k]
	if !ok {
		return "", fmt.Errorf("%w: '%s'", errKeyNotFound, k)
	}
	return processTemplate(s, params)
}

func localizeFail(k string, params ParameterMap) (string, error) {
	return "", fmt.Errorf("%w: '%s'", errKeyNotFound, k)
}
