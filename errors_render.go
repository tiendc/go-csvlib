package csvlib

import (
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/tiendc/gofn"
)

const (
	newLine = "\n"
)

type ErrorRenderConfig struct {
	// HeaderFormatKey header format string
	// You can use a localization key as the value to force the renderer to translate the key first.
	// If the translation fails, the original value is used for next step.
	// For example, header format key can be:
	//   - "HEADER_FORMAT_KEY" (a localization key you define in your localization data such as a json file,
	//         HEADER_FORMAT_KEY = "CSV decoding result: total errors is {{.TotalError}}")
	//   - "CSV decoding result: total errors is {{.TotalError}}" (direct string)
	//
	// Supported params:
	//   {{.TotalRow}}       - number of rows in the CSV data
	//   {{.TotalRowError}}  - number of rows have error
	//   {{.TotalCellError}} - number of cells have error
	//   {{.TotalError}}     - number of errors
	//
	// Extra params:
	//   {{.CrLf}} - line break
	//   {{.Tab}}  - tab character
	HeaderFormatKey string

	// RowFormatKey format string for each row
	// Similar to the header format key, this can be a localization key or a direct string.
	// For example, row format key can be:
	//   - "ROW_FORMAT_KEY" (a localization key you define in your localization data such as a json file)
	//   - "Row {{.Row}} (line {{.Line}}): {{.Error}}" (direct string)
	//
	// Supported params:
	//   {{.Row}}   - row index (1-based, row 1 can be the header row if present)
	//   {{.Line}}  - line of row in source file (can be -1 if undetected)
	//   {{.Error}} - error content of the row which is a list of cell errors
	RowFormatKey string

	// RowSeparator separator to join row error details, normally a row is in a separated line
	RowSeparator string

	// CellSeparator separator to join cell error details within a row, normally a comma (`,`)
	CellSeparator string

	// LineBreak custom new line character (default is `\n`)
	LineBreak string

	// LocalizeCellFields localize cell's fields before rendering the cell error (default is `true`)
	LocalizeCellFields bool

	// LocalizeCellHeader localize cell header before rendering the cell error (default is `true`)
	LocalizeCellHeader bool

	// Params custom params user wants to send to the localization (optional)
	Params ParameterMap

	// LocalizationFunc function to translate message (optional)
	LocalizationFunc LocalizationFunc

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

func defaultRenderConfig() *ErrorRenderConfig {
	return &ErrorRenderConfig{
		HeaderFormatKey: "Error content: TotalRow: {{.TotalRow}}, TotalRowError: {{.TotalRowError}}, " +
			"TotalCellError: {{.TotalCellError}}, TotalError: {{.TotalError}}",
		RowFormatKey: "Row {{.Row}} (line {{.Line}}): {{.Error}}",

		RowSeparator:  newLine,
		CellSeparator: ", ",
		LineBreak:     newLine,

		LocalizeCellFields: true,
		LocalizeCellHeader: true,
	}
}

type SimpleRenderer struct {
	cfg       *ErrorRenderConfig
	sourceErr *Errors
	transErr  error
}

func NewRenderer(err *Errors, options ...func(*ErrorRenderConfig)) (*SimpleRenderer, error) {
	cfg := defaultRenderConfig()
	for _, opt := range options {
		opt(cfg)
	}
	return &SimpleRenderer{cfg: cfg, sourceErr: err}, nil
}

// Render render Errors object as text
// Sample output:
//
//	There are 5 total errors in your CSV file
//	Row 20 (line 21): column 2: invalid type (Int), column 4: value (12345) too big
//	Row 30 (line 33): column 2: invalid type (Int), column 4: value (12345) too big, column 6: unexpected
//	Row 35 (line 38): column 2: invalid type (Int), column 4: value (12345) too big
//	Row 40 (line 44): column 2: invalid type (Int), column 4: value (12345) too big
//	Row 41 (line 50): invalid number of columns (10)
func (r *SimpleRenderer) Render() (msg string, transErr error, err error) {
	cfg := r.cfg
	errs := r.sourceErr.Unwrap()
	content := make([]string, 0, len(errs)+1)
	params := gofn.MapUpdate(ParameterMap{
		"CrLf": cfg.LineBreak,
		"Tab":  "\t",

		"TotalRow":       r.sourceErr.TotalRow(),
		"TotalError":     r.sourceErr.TotalError(),
		"TotalRowError":  r.sourceErr.TotalRowError(),
		"TotalCellError": r.sourceErr.TotalCellError(),
	}, cfg.Params)

	// Header line
	if cfg.HeaderFormatKey != "" {
		header := r.localizeKeySkipError(cfg.HeaderFormatKey, params)
		if header != "" {
			content = append(content, header)
		}
	}

	// Body part (simply each RowErrors object is rendered as a line)
	for _, err := range errs {
		var detail string
		if rowErr, ok := err.(*RowErrors); ok { // nolint: errorlint
			detail = r.renderRow(rowErr, params)
		} else {
			detail = r.renderCommonError(err, params)
		}
		if detail != "" {
			content = append(content, detail)
		}
	}

	return strings.Join(content, cfg.RowSeparator), r.transErr, nil
}

func (r *SimpleRenderer) renderRow(rowErr *RowErrors, exparams ParameterMap) string {
	cfg := r.cfg
	errs := rowErr.Unwrap()
	content := make([]string, 0, len(errs))

	params := gofn.MapUpdate(ParameterMap{}, exparams)
	params["Row"] = rowErr.Row()
	params["Line"] = rowErr.Line()

	for _, err := range errs {
		var detail string
		if cellErr, ok := err.(*CellError); ok { // nolint: errorlint
			detail = r.renderCell(rowErr, cellErr, params)
		} else {
			detail = r.renderCommonError(err, params)
		}
		if detail != "" {
			content = append(content, detail)
		}
	}

	if cfg.RowFormatKey != "" {
		params["Error"] = strings.Join(content, cfg.CellSeparator)
		return r.localizeKeySkipError(cfg.RowFormatKey, params)
	}
	return ""
}

func (r *SimpleRenderer) renderCell(rowErr *RowErrors, cellErr *CellError, exparams ParameterMap) string {
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

func (r *SimpleRenderer) renderCellFields(cellErr *CellError, params ParameterMap) ParameterMap {
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

func (r *SimpleRenderer) renderCellHeader(cellErr *CellError, params ParameterMap) string {
	if !r.cfg.LocalizeCellHeader {
		return cellErr.Header()
	}
	return r.localizeKeySkipError(cellErr.Header(), params)
}

func (r *SimpleRenderer) renderCommonError(err error, params ParameterMap) string {
	if r.cfg.CommonErrorRenderFunc == nil {
		return r.localizeKeySkipError(err.Error(), params)
	}
	msg, err := r.cfg.CommonErrorRenderFunc(err, params)
	if err != nil {
		r.transErr = multierror.Append(r.transErr, err)
	}
	return msg
}

func (r *SimpleRenderer) localizeKey(key string, params ParameterMap) (string, error) {
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

func (r *SimpleRenderer) localizeKeySkipError(key string, params ParameterMap) string {
	s, err := r.localizeKey(key, params)
	if err != nil {
		s = key
	}
	return processParams(s, params)
}
