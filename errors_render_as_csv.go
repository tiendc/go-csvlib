package csvlib

import (
	"bytes"
	"encoding/csv"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/tiendc/gofn"
)

type CSVRenderConfig struct {
	// CellSeparator separator to join cell error details within a row, normally a comma (`,`)
	CellSeparator string

	// LineBreak custom new line character (default is `\n`)
	LineBreak string

	// RenderHeader whether render header row or not
	RenderHeader bool

	// RenderRowNumberColumnIndex index of `row` column to render, set `-1` to not render it (default is `0`)
	RenderRowNumberColumnIndex int

	// RenderLineNumberColumnIndex index of `line` column to render, set `-1` to not render it (default is `-1`)
	RenderLineNumberColumnIndex int

	// RenderCommonErrorColumnIndex index of `common error` column to render, set `-1` to not render it
	// (default is `1`)
	RenderCommonErrorColumnIndex int

	// LocalizeCellFields localize cell's fields before rendering the cell error (default is `true`)
	LocalizeCellFields bool

	// LocalizeCellHeader localize cell header before rendering the cell error (default is `true`)
	LocalizeCellHeader bool

	// Params custom params user wants to send to the localization (optional)
	Params ParameterMap

	// LocalizationFunc function to translate message (optional)
	LocalizationFunc LocalizationFunc

	// HeaderRenderFunc custom render function for rendering header row (optional)
	HeaderRenderFunc func([]string, ParameterMap)

	// CellRenderFunc custom render function for rendering a cell error (optional)
	// The func can return ("", false) to skip rendering the cell error, return ("", true) to let the
	// renderer continue using its solution, and return ("<str>", true) to override the value.
	//
	// Supported params:
	//   {{.Column}}       - column index (0-based)
	//   {{.ColumnHeader}} - column name
	//   {{.Value}}        - cell value
	//   {{.Error}}        - error detail which is result of calling err.Error()
	//
	// Use cellErr.WithParam() to add more extra params
	CellRenderFunc func(*RowErrors, *CellError, ParameterMap) (string, bool)

	// CommonErrorRenderFunc renders common error (not RowErrors, CellError) (optional)
	CommonErrorRenderFunc func(error, ParameterMap) (string, error)
}

func defaultCSVRenderConfig() *CSVRenderConfig {
	return &CSVRenderConfig{
		CellSeparator: ", ",
		LineBreak:     newLine,

		RenderHeader:                 true,
		RenderRowNumberColumnIndex:   0,
		RenderLineNumberColumnIndex:  1,
		RenderCommonErrorColumnIndex: 2,

		LocalizeCellFields: true,
		LocalizeCellHeader: true,
	}
}

type CSVRenderer struct {
	cfg               *CSVRenderConfig
	sourceErr         *Errors
	transErr          error
	numColumns        int
	startCellErrIndex int
	data              [][]string
}

func NewCSVRenderer(err *Errors, options ...func(*CSVRenderConfig)) (*CSVRenderer, error) {
	cfg := defaultCSVRenderConfig()
	for _, opt := range options {
		opt(cfg)
	}
	// Validate/Correct the base columns to render
	baseColumns := make([]*int, 0, 3) // nolint: gomnd
	if cfg.RenderRowNumberColumnIndex >= 0 {
		baseColumns = append(baseColumns, &cfg.RenderRowNumberColumnIndex)
	}
	if cfg.RenderLineNumberColumnIndex >= 0 {
		baseColumns = append(baseColumns, &cfg.RenderLineNumberColumnIndex)
	}
	if cfg.RenderCommonErrorColumnIndex >= 0 {
		baseColumns = append(baseColumns, &cfg.RenderCommonErrorColumnIndex)
	}
	sort.Slice(baseColumns, func(i, j int) bool {
		return *baseColumns[i] < *baseColumns[j]
	})
	for i := range baseColumns {
		*baseColumns[i] = i
	}

	return &CSVRenderer{cfg: cfg, sourceErr: err}, nil
}

// Render render Errors object as CSV rows data
// Sample output:
//
//	There are 5 total errors in your CSV file
//	Row 20 (line 21): column 2: invalid type (Int), column 4: value (12345) too big
//	Row 30 (line 33): column 2: invalid type (Int), column 4: value (12345) too big, column 6: unexpected
//	Row 35 (line 38): column 2: invalid type (Int), column 4: value (12345) too big
//	Row 40 (line 44): column 2: invalid type (Int), column 4: value (12345) too big
//	Row 41 (line 50): invalid number of columns (10)
func (r *CSVRenderer) Render() (data [][]string, transErr error, err error) {
	cfg := r.cfg
	r.startCellErrIndex = 0
	if cfg.RenderRowNumberColumnIndex >= 0 {
		r.startCellErrIndex++
	}
	if cfg.RenderLineNumberColumnIndex >= 0 {
		r.startCellErrIndex++
	}
	if cfg.RenderCommonErrorColumnIndex >= 0 {
		r.startCellErrIndex++
	}

	r.numColumns = len(r.sourceErr.Header()) + r.startCellErrIndex
	errs := r.sourceErr.Unwrap()
	r.data = make([][]string, 0, len(errs)+1)

	params := gofn.MapUpdate(ParameterMap{
		"CrLf": cfg.LineBreak,
		"Tab":  "\t",

		"TotalRow":       r.sourceErr.TotalRow(),
		"TotalError":     r.sourceErr.TotalError(),
		"TotalRowError":  r.sourceErr.TotalRowError(),
		"TotalCellError": r.sourceErr.TotalCellError(),
	}, cfg.Params)

	// Render header row
	r.renderHeader(params)

	// Render rows content
	for _, err := range errs {
		if rowErr, ok := err.(*RowErrors); ok { // nolint: errorlint
			rowContent := r.renderRow(rowErr, params)
			r.data = append(r.data, rowContent)
		} else {
			_ = r.renderCommonError(err, params)
		}
	}
	return r.data, r.transErr, nil
}

func (r *CSVRenderer) RenderAsString() (msg string, transErr error, err error) {
	csvData, transErr, err := r.Render()
	if err != nil {
		return "", transErr, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, r.estimateCSVBuffer(csvData)))
	w := csv.NewWriter(buf)
	if err = w.WriteAll(csvData); err != nil {
		return "", transErr, err
	}
	w.Flush()
	return buf.String(), transErr, nil
}

func (r *CSVRenderer) RenderTo(w Writer) (transErr error, err error) {
	csvData, transErr, err := r.Render()
	if err != nil {
		return transErr, err
	}
	writeAll, canWriteAll := w.(interface{ WriteAll([][]string) error })
	if canWriteAll {
		return transErr, writeAll.WriteAll(csvData)
	}
	for _, row := range csvData {
		err = w.Write(row)
		if err != nil {
			return transErr, err
		}
	}
	return transErr, nil
}

func (r *CSVRenderer) renderHeader(exparams ParameterMap) {
	if !r.cfg.RenderHeader {
		return
	}
	cfg := r.cfg
	header := make([]string, r.numColumns)
	if cfg.RenderRowNumberColumnIndex >= 0 {
		header[cfg.RenderRowNumberColumnIndex] = "Row"
	}
	if cfg.RenderLineNumberColumnIndex >= 0 {
		header[cfg.RenderLineNumberColumnIndex] = "Line"
	}
	if cfg.RenderCommonErrorColumnIndex >= 0 {
		header[cfg.RenderCommonErrorColumnIndex] = "CommonError"
	}
	for i := r.startCellErrIndex; i < r.numColumns; i++ {
		header[i] = r.sourceErr.header[i-r.startCellErrIndex]
	}

	if cfg.HeaderRenderFunc != nil {
		cfg.HeaderRenderFunc(header, exparams)
	}
	r.data = append(r.data, header)
}

func (r *CSVRenderer) renderRow(rowErr *RowErrors, exparams ParameterMap) []string {
	cfg := r.cfg
	content := make([]string, r.numColumns)

	if cfg.RenderRowNumberColumnIndex >= 0 {
		content[cfg.RenderRowNumberColumnIndex] = strconv.FormatInt(int64(rowErr.row), 10)
	}
	if cfg.RenderLineNumberColumnIndex >= 0 {
		content[cfg.RenderLineNumberColumnIndex] = strconv.FormatInt(int64(rowErr.line), 10)
	}

	errs := rowErr.Unwrap()
	mapErrByIndex := make(map[int][]string, r.numColumns)
	params := gofn.MapUpdate(ParameterMap{}, exparams)
	params["Row"] = rowErr.Row()
	params["Line"] = rowErr.Line()

	for _, err := range errs {
		if cellErr, ok := err.(*CellError); ok { // nolint: errorlint
			detail := r.renderCell(rowErr, cellErr, params)
			colIndex := cellErr.column + r.startCellErrIndex
			if cellErr.column == -1 {
				colIndex = cfg.RenderCommonErrorColumnIndex
			}
			if listItems, ok := mapErrByIndex[colIndex]; ok {
				mapErrByIndex[colIndex] = append(listItems, detail)
			} else {
				mapErrByIndex[colIndex] = []string{detail}
			}
			continue
		}
		// Common error
		detail := r.renderCommonError(err, params)
		if listItems, ok := mapErrByIndex[cfg.RenderCommonErrorColumnIndex]; ok {
			mapErrByIndex[cfg.RenderCommonErrorColumnIndex] = append(listItems, detail)
		} else {
			mapErrByIndex[cfg.RenderCommonErrorColumnIndex] = []string{detail}
		}
	}

	for index, items := range mapErrByIndex {
		content[index] = strings.Join(items, cfg.CellSeparator)
	}
	return content
}

func (r *CSVRenderer) renderCell(rowErr *RowErrors, cellErr *CellError, exparams ParameterMap) string {
	params := gofn.MapUpdate(ParameterMap{}, exparams)
	params = gofn.MapUpdate(params, r.renderCellFields(cellErr, params))
	params["Column"] = cellErr.Column()
	params["ColumnHeader"] = r.renderCellHeader(cellErr, params)
	params["Value"] = cellErr.Value()
	params["Error"] = cellErr.Error()

	if r.cfg.CellRenderFunc != nil {
		msg, flag := r.cfg.CellRenderFunc(rowErr, cellErr, exparams)
		if !flag {
			return ""
		}
		if msg != "" {
			return msg
		}
	}

	locKey := cellErr.LocalizationKey()
	if locKey == "" {
		locKey = cellErr.Error()
	}
	return r.localizeKeySkipError(locKey, params)
}

func (r *CSVRenderer) renderCellFields(cellErr *CellError, params ParameterMap) ParameterMap {
	if !r.cfg.LocalizeCellFields {
		return cellErr.fields
	}
	result := make(ParameterMap, len(cellErr.fields))
	for k, v := range cellErr.fields {
		vAsStr, ok := v.(string)
		if !ok {
			result[k] = v
			continue
		}
		if translated, err := r.localizeKey(vAsStr, params); err != nil {
			result[k] = v
		} else {
			result[k] = translated
		}
	}
	return result
}

func (r *CSVRenderer) renderCellHeader(cellErr *CellError, params ParameterMap) string {
	if !r.cfg.LocalizeCellHeader {
		return cellErr.Header()
	}
	return r.localizeKeySkipError(cellErr.Header(), params)
}

func (r *CSVRenderer) renderCommonError(err error, params ParameterMap) string {
	if r.cfg.CommonErrorRenderFunc == nil {
		return r.localizeKeySkipError(err.Error(), params)
	}
	msg, err := r.cfg.CommonErrorRenderFunc(err, params)
	if err != nil {
		r.transErr = multierror.Append(r.transErr, err)
	}
	return msg
}

func (r *CSVRenderer) localizeKey(key string, params ParameterMap) (string, error) {
	if r.cfg.LocalizationFunc == nil {
		return processParams(key, params), nil
	}
	msg, err := r.cfg.LocalizationFunc(key, params)
	if err != nil {
		err = multierror.Append(ErrLocalization, err)
		r.transErr = multierror.Append(r.transErr, err)
		return "", err
	}
	return msg, nil
}

func (r *CSVRenderer) localizeKeySkipError(key string, params ParameterMap) string {
	s, err := r.localizeKey(key, params)
	if err != nil {
		s = key
	}
	return processParams(s, params)
}

func (r *CSVRenderer) estimateCSVBuffer(data [][]string) int {
	if len(data) <= 1 {
		return 512 // nolint: gomnd
	}
	row := data[1]
	rowSz := 0
	for _, v := range row {
		rowSz += len(v)
	}
	return (rowSz + len(row)) * len(data)
}
