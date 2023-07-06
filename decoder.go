package csvlib

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/tiendc/gofn"
)

type DecodeConfig struct {
	// TagName tag name to parse the struct (default is "csv")
	TagName string

	// NoHeaderMode indicates the input data have no header (default is "false")
	NoHeaderMode bool

	// StopOnError when error occurs, stop the processing (default is "true")
	StopOnError bool

	// TrimSpace trim all cell values before processing (default is "false")
	TrimSpace bool

	// RequireColumnOrder order of columns defined in struct must match the order of columns
	// in the input data (default is "true")
	RequireColumnOrder bool

	// ParseLocalizedHeader header in the input data is localized (default is "false")
	// For example:
	//  type Student struct {
	//     Name string `csv:"name"`  -> `name` is header key now, the actual header is localized based on the key
	//     Age  int    `csv:"age"`
	//  }
	ParseLocalizedHeader bool

	// AllowUnrecognizedColumns allow a column in the input data but not in the struct tag definition
	// (default is "false")
	AllowUnrecognizedColumns bool

	// TreatIncorrectStructureAsError treat incorrect data structure as error (default is "true")
	// For example: header has 5 columns, if there is a row having 6 columns, it will be treated as error
	// and the decoding process will stop even StopOnError flag is false.
	TreatIncorrectStructureAsError bool

	// DetectRowLine detect exact lines of rows (default is "false")
	// If turn this flag on, the input reader should be an instance of "encoding/csv" Reader
	// as this lib uses Reader.FieldPos() function to get the line of a row.
	DetectRowLine bool

	// LocalizationFunc localization function, required when ParseLocalizedHeader is true
	LocalizationFunc LocalizationFunc

	// columnConfigMap a map consists of configuration for specific columns
	columnConfigMap map[string]*DecodeColumnConfig
}

func defaultDecodeConfig() *DecodeConfig {
	return &DecodeConfig{
		TagName:                        DefaultTagName,
		StopOnError:                    true,
		RequireColumnOrder:             true,
		TreatIncorrectStructureAsError: true,
	}
}

func (c *DecodeConfig) ConfigureColumn(name string, fn func(*DecodeColumnConfig)) {
	if c.columnConfigMap == nil {
		c.columnConfigMap = map[string]*DecodeColumnConfig{}
	}
	columnCfg, ok := c.columnConfigMap[name]
	if !ok {
		columnCfg = defaultDecodeColumnConfig()
		c.columnConfigMap[name] = columnCfg
	}
	fn(columnCfg)
}

// DecodeColumnConfig configuration for a specific column
type DecodeColumnConfig struct {
	// TrimSpace if `true` and DecodeConfig.TrimSpace is `false`, only trim space this column
	// (default is "false")
	TrimSpace bool

	// StopOnError if `true` and DecodeConfig.StopOnError is `false`, only stop when error occurs
	// within this column processing (default is "false")
	StopOnError bool

	// DecodeFunc custom decode function (optional)
	DecodeFunc DecodeFunc

	// PreprocessorFuncs a list of functions will be called before decoding a cell value (optional)
	PreprocessorFuncs []ProcessorFunc

	// ValidatorFuncs a list of functions will be called after decoding (optional)
	ValidatorFuncs []ValidatorFunc

	// OnCellErrorFunc function will be called every time an error happens when decode a cell
	// This func can be helpful to set localization key and additional params for the error
	// to localize the error message later on. (optional)
	OnCellErrorFunc OnCellErrorFunc
}

func defaultDecodeColumnConfig() *DecodeColumnConfig {
	return &DecodeColumnConfig{}
}

type DecodeOption func(cfg *DecodeConfig)

type DecodeResult struct {
	totalRow               int
	unrecognizedColumns    []string
	missingOptionalColumns []string
}

func (r *DecodeResult) TotalRow() int {
	return r.totalRow
}

func (r *DecodeResult) UnrecognizedColumns() []string {
	return r.unrecognizedColumns
}

func (r *DecodeResult) MissingOptionalColumns() []string {
	return r.missingOptionalColumns
}

type Decoder struct {
	r                       Reader
	cfg                     *DecodeConfig
	err                     *Errors
	result                  *DecodeResult
	finished                bool
	rowsData                []*rowData
	itemType                reflect.Type
	shouldStop              bool
	hasDynamicInlineColumns bool
	hasFixedInlineColumns   bool
	colsMeta                []*decodeColumnMeta
}

func NewDecoder(r Reader, options ...DecodeOption) *Decoder {
	cfg := defaultDecodeConfig()
	for _, opt := range options {
		opt(cfg)
	}
	return &Decoder{
		r:   r,
		cfg: cfg,
		err: NewErrors(),
	}
}

// Decode decode input data and store the result in the given variable
// The input var must be a pointer to a slice, e.g. `*[]Student` (recommended) or `*[]*Student`
func (d *Decoder) Decode(v interface{}) (*DecodeResult, error) {
	if d.finished {
		return nil, ErrFinished
	}
	if d.shouldStop {
		return nil, ErrAlreadyFailed
	}

	val := reflect.ValueOf(v)
	if d.itemType == nil {
		if err := d.prepareDecode(val); err != nil {
			d.err.Add(err)
			d.shouldStop = true
			return nil, d.err
		}
	} else {
		itemType, err := d.parseOutputVar(val)
		if err != nil {
			return nil, err
		}
		if itemType != d.itemType {
			return nil, fmt.Errorf("%w: %v (expect %v)", ErrTypeUnmatched, itemType, d.itemType)
		}
	}

	outSlice := reflect.MakeSlice(val.Type().Elem(), len(d.rowsData), len(d.rowsData))
	itemKindIsPtr := d.itemType.Kind() == reflect.Pointer
	row := 0
	for !d.shouldStop && len(d.rowsData) > 0 {
		// Reduce memory consumption by splitting the source data into chunks (10000 items each)
		// After each chunk is processed, resize the slice to allow Go to free the memory when necessary
		chunkSz := gofn.Min(10000, len(d.rowsData)) // nolint: gomnd
		chunk := d.rowsData[0:chunkSz]
		d.rowsData = d.rowsData[chunkSz:]

		for _, rowData := range chunk {
			rowVal := outSlice.Index(row)
			row++
			if itemKindIsPtr {
				rowVal.Set(reflect.New(d.itemType.Elem()))
				rowVal = rowVal.Elem()
			}
			if err := d.decodeRow(rowData, rowVal); err != nil {
				d.err.Add(err)
				if d.cfg.StopOnError || d.shouldStop {
					d.shouldStop = true
					break
				}
			}
		}
	}

	if d.err.HasError() {
		return d.result, d.err
	}
	val.Elem().Set(outSlice)
	d.finished = len(d.rowsData) == 0
	return d.result, nil
}

// DecodeOne decode the next one row data
// The input var must be a pointer to a struct (e.g. *Student)
// This func returns error of the current row processing only, after finishing the last row decoding,
// call Finish() to get the overall result and error.
func (d *Decoder) DecodeOne(v interface{}) error {
	if d.finished {
		return ErrFinished
	}
	if d.shouldStop {
		return ErrAlreadyFailed
	}

	rowVal := reflect.ValueOf(v)
	rowVal, itemType, err := d.parseOutputVarOne(rowVal)
	if err != nil {
		return err
	}
	if d.itemType == nil {
		if err := d.prepareDecode(reflect.New(reflect.SliceOf(itemType))); err != nil {
			d.err.Add(err)
			d.shouldStop = true
			return err
		}
	} else {
		if itemType != d.itemType {
			return fmt.Errorf("%w: %v (expect %v)", ErrTypeUnmatched, itemType, d.itemType)
		}
	}

	if len(d.rowsData) == 0 {
		d.finished = true
		return ErrFinished
	}
	rowData := d.rowsData[0]
	d.rowsData = d.rowsData[1:]
	err = d.decodeRow(rowData, rowVal)
	if err != nil {
		d.err.Add(err)
		if d.cfg.StopOnError {
			d.shouldStop = true
		}
	}
	d.finished = len(d.rowsData) == 0
	return err
}

// Finish decoding, after calling this func, you can't decode more even there is data
func (d *Decoder) Finish() (*DecodeResult, error) {
	d.finished = true
	if d.err.HasError() {
		return d.result, d.err
	}
	return d.result, nil
}

// prepareDecode prepare for decoding by parsing the struct tags and build column decoders
// This step is performed one time only before the first row decoding
func (d *Decoder) prepareDecode(v reflect.Value) error {
	d.result = &DecodeResult{}
	itemType, err := d.parseOutputVar(v)
	if err != nil {
		return err
	}
	d.itemType = itemType

	if err = d.validateConfig(); err != nil {
		return err
	}

	if err = d.parseColumnsMeta(itemType); err != nil { // typ: []Item, itemTyp: Item
		return err
	}

	if err = d.buildColumnDecoders(); err != nil {
		return err
	}

	if err = d.readRowData(); err != nil {
		return err
	}

	totalRow := len(d.rowsData)
	if !d.cfg.NoHeaderMode {
		totalRow++
	}
	d.result.totalRow = totalRow
	d.err.totalRow = totalRow
	for _, colMeta := range d.colsMeta {
		d.err.header = append(d.err.header, colMeta.headerText)
	}
	return nil
}

// decodeRow decode row data and write the result to the row target value
// `rowVal` is normally a slice item at a specific index
// nolint: gocyclo,gocognit
func (d *Decoder) decodeRow(rowData *rowData, rowVal reflect.Value) error {
	cfg, colsMeta := d.cfg, d.colsMeta
	if rowData.err != nil {
		rowErr := NewRowErrors(rowData.row, rowData.line)
		rowErr.Add(d.handleCellError(rowData.err, "", nil))
		return rowErr
	}

	if d.hasDynamicInlineColumns || d.hasFixedInlineColumns {
		for _, colMeta := range colsMeta {
			if colMeta.inlineColumnMeta != nil {
				colMeta.inlineColumnMeta.decodePrepareForNextRow()
			}
		}
	}

	var cellErrs []error
	for col, cellText := range rowData.records {
		colMeta := colsMeta[col]
		if colMeta.unrecognized {
			continue
		}
		if cfg.TrimSpace || colMeta.trimSpace {
			cellText = strings.TrimSpace(cellText)
		}
		for _, fn := range colMeta.preprocessorFuncs {
			cellText = fn(cellText)
		}

		outVal := rowVal.Field(colMeta.targetField.Index[0])
		if colMeta.inlineColumnMeta != nil {
			outVal = colMeta.inlineColumnMeta.decodeGetColumnValue(outVal)
		}

		var errs []error
		hasDecodeErr := false
		if !colMeta.omitempty || cellText != "" {
			if err := colMeta.decodeFunc(cellText, outVal); err != nil {
				errs = []error{err}
				hasDecodeErr = true
			}
		}
		if !hasDecodeErr && len(colMeta.validatorFuncs) > 0 {
			errs = d.validateParsedCell(outVal, colMeta)
		}
		for _, err := range errs {
			cellErrs = append(cellErrs, d.handleCellError(err, rowData.records[col], colMeta))
			if cfg.StopOnError || colMeta.stopOnError {
				d.shouldStop = true
				break
			}
		}
	}
	if len(cellErrs) > 0 {
		rowErr := NewRowErrors(rowData.row, rowData.line)
		rowErr.Add(cellErrs...)
		return rowErr
	}
	return nil
}

// validateParsedCell validate a cell value after decoding
func (d *Decoder) validateParsedCell(v reflect.Value, colMeta *decodeColumnMeta) []error {
	var errs []error
	vAsIface := v.Interface()
	for _, validatorFunc := range colMeta.validatorFuncs {
		err := validatorFunc(vAsIface)
		if err != nil {
			if _, ok := err.(*CellError); !ok { // nolint: errorlint
				err = NewCellError(err, colMeta.column, colMeta.headerText)
			}
			errs = append(errs, err)
			if d.cfg.StopOnError || colMeta.stopOnError {
				return errs
			}
		}
	}
	return errs
}

// handleCellError build cell error for the given error and call the onCellErrorFunc
func (d *Decoder) handleCellError(err error, value string, colMeta *decodeColumnMeta) error {
	cellErr, ok := err.(*CellError) // nolint: errorlint
	if !ok {
		if colMeta != nil {
			cellErr = NewCellError(err, colMeta.column, colMeta.headerText)
		} else {
			// This is error that not relate to any column (e.g. RwoFieldCount error)
			cellErr = NewCellError(err, -1, "")
		}
	}
	cellErr.value = value
	if colMeta != nil && colMeta.onCellErrorFunc != nil {
		colMeta.onCellErrorFunc(cellErr)
	}
	return cellErr
}

// parseOutputVar parse and validate the input var
func (d *Decoder) parseOutputVar(v reflect.Value) (itemType reflect.Type, err error) {
	if v.Kind() != reflect.Pointer || v.IsNil() {
		err = fmt.Errorf("%w: %v", ErrTypeInvalid, v.Kind())
		return
	}

	typ := v.Type().Elem() // E.g. []Item
	switch typ.Kind() {    // nolint: exhaustive
	case reflect.Slice, reflect.Array:
	default:
		err = fmt.Errorf("%w: %v", ErrTypeInvalid, typ.Kind())
		return
	}

	itemType = typ.Elem()
	if indirectType(itemType).Kind() != reflect.Struct {
		err = fmt.Errorf("%w: %v", ErrTypeInvalid, itemType.Kind())
		return
	}
	return
}

func (d *Decoder) parseOutputVarOne(v reflect.Value) (val reflect.Value, itemType reflect.Type, err error) {
	itemType = v.Type()
	if itemType.Kind() != reflect.Pointer || itemType.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("%w: must be a pointer to struct", ErrTypeInvalid)
	}
	itemType = itemType.Elem()
	if v.IsNil() {
		return reflect.Value{}, nil, fmt.Errorf("%w: must be a non-nil pointer", ErrValueNil)
	}
	val = v.Elem()
	return
}

// readRowData read data of all rows from the input to struct type
// If you use `encoding/csv` Reader, we can determine the lines of rows (via Reader.FieldPos func).
// Otherwise, `line` will be set to `-1` which mean undetected.
func (d *Decoder) readRowData() error {
	cfg, r := d.cfg, d.r
	row := 2
	if d.cfg.NoHeaderMode {
		row = 1
	}
	getLine, ableToGetLine := r.(interface {
		FieldPos(field int) (line, column int) // Reader from "encoding/csv" provides this func
	})
	if !cfg.DetectRowLine {
		ableToGetLine = false
		getLine = nil
	}
	rowDataItems := make([]*rowData, 0, 10000) // nolint: gomnd

	for ; ; row++ {
		records, err := r.Read()
		line := -1
		if err == nil {
			if ableToGetLine {
				line, _ = getLine.FieldPos(0)
			}
			rowDataItems = append(rowDataItems, &rowData{
				records: records,
				line:    line,
				row:     row,
			})
			continue
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if errors.Is(err, csv.ErrFieldCount) {
			err = fmt.Errorf("%w: row %d", ErrDecodeRowFieldCount, row)
			if cfg.TreatIncorrectStructureAsError || cfg.StopOnError {
				return err
			}
			if ableToGetLine {
				line, _ = getLine.FieldPos(0)
			}
			rowDataItems = append(rowDataItems, &rowData{row: row, line: line, err: err})
			continue
		}
		if errors.Is(err, csv.ErrQuote) || errors.Is(err, csv.ErrBareQuote) {
			err = fmt.Errorf("%w: row %d", ErrDecodeQuoteInvalid, row)
			if cfg.TreatIncorrectStructureAsError || cfg.StopOnError {
				return err
			}
			// NOTE: it seems when invalid quote, calling getLine will panic
			rowDataItems = append(rowDataItems, &rowData{row: row, line: line, err: err})
			continue
		}
		return err
	}

	d.rowsData = rowDataItems
	return nil
}

// parseColumnsMeta parse struct metadata
func (d *Decoder) parseColumnsMeta(itemType reflect.Type) error {
	cfg, result := d.cfg, d.result
	fileHeader, err := d.readFileHeader()
	if err != nil {
		return err
	}

	colsMetaFromStruct, err := d.parseColumnsMetaFromStructType(itemType, fileHeader)
	if err != nil {
		return err
	}
	if len(fileHeader) == 0 {
		d.colsMeta = colsMetaFromStruct
		return nil
	}

	mapColMetaFromStruct := make(map[string]*decodeColumnMeta, len(colsMetaFromStruct))
	for _, colMeta := range colsMetaFromStruct {
		mapColMetaFromStruct[colMeta.headerText] = colMeta
	}

	colsMeta := make([]*decodeColumnMeta, 0, len(fileHeader))
	for _, headerText := range fileHeader {
		colMeta := mapColMetaFromStruct[headerText]
		if colMeta == nil {
			if !cfg.AllowUnrecognizedColumns {
				return fmt.Errorf("%w: \"%s\"", ErrHeaderColumnUnrecognized, headerText)
			}
			colMeta = &decodeColumnMeta{
				headerKey:    headerText,
				headerText:   headerText,
				unrecognized: true,
			}
		}
		colMeta.column = len(colsMeta)
		colsMeta = append(colsMeta, colMeta)
	}

	mapColMeta := make(map[string]*decodeColumnMeta, len(colsMeta))
	for _, colMeta := range colsMeta {
		mapColMeta[colMeta.headerText] = colMeta
		if colMeta.unrecognized {
			result.unrecognizedColumns = append(result.unrecognizedColumns, colMeta.headerText)
		}
	}
	for _, colMeta := range colsMetaFromStruct {
		if _, ok := mapColMeta[colMeta.headerText]; !ok {
			if !colMeta.optional {
				return fmt.Errorf("%w: \"%s\"", ErrHeaderColumnRequired, colMeta.headerText)
			}
			result.missingOptionalColumns = append(result.missingOptionalColumns, colMeta.headerText)
		}
	}

	if cfg.RequireColumnOrder {
		if err = d.validateHeaderOrder(colsMeta, colsMetaFromStruct); err != nil {
			return err
		}
	}

	if err = d.validateColumnsMeta(colsMeta, colsMetaFromStruct); err != nil {
		return err
	}

	d.colsMeta = colsMeta
	return nil
}

// validateColumnsMeta validate struct metadata
func (d *Decoder) validateColumnsMeta(colsMeta, colsMetaFromStruct []*decodeColumnMeta) error {
	cfg := d.cfg
	// Make sure all column options valid
	for colKey := range cfg.columnConfigMap {
		if !gofn.ContainPred(colsMetaFromStruct, func(colMeta *decodeColumnMeta) bool {
			return colMeta.headerKey == colKey || colMeta.parentKey == colKey
		}) {
			return fmt.Errorf("%w: column \"%s\" not found", ErrConfigOptionInvalid, colKey)
		}
	}

	// Make sure all columns are unique
	if err := d.validateHeaderUniqueness(colsMeta); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) readFileHeader() (fileHeader []string, err error) {
	if !d.cfg.NoHeaderMode {
		fileHeader, err = d.r.Read()
		if err != nil {
			return nil, err
		}
	}
	if err = validateHeader(fileHeader); err != nil {
		return nil, err
	}
	return
}

func (d *Decoder) parseColumnsMetaFromStructType(itemType reflect.Type, fileHeader []string) (
	colsMeta []*decodeColumnMeta, err error) {
	cfg := d.cfg
	itemType = indirectType(itemType)
	numFields := itemType.NumField()
	for i := 0; i < numFields; i++ {
		field := itemType.Field(i)
		tag, err := parseTag(cfg.TagName, field)
		if err != nil {
			return nil, err
		}
		if tag == nil || tag.ignored {
			continue
		}

		colMeta := &decodeColumnMeta{
			column:      len(colsMeta),
			headerKey:   tag.name,
			headerText:  tag.name,
			prefix:      tag.prefix,
			optional:    tag.optional,
			omitempty:   tag.omitEmpty,
			targetField: field,
		}

		if tag.inline {
			inlineColumnsMeta, err := d.parseInlineColumn(field, colMeta)
			if err != nil {
				return nil, err
			}
			colsMeta = append(colsMeta, inlineColumnsMeta...)
			continue
		}

		colMeta.copyConfig(cfg.columnConfigMap[colMeta.headerKey])
		if err = colMeta.localizeHeader(cfg); err != nil {
			return nil, err
		}

		colsMeta = append(colsMeta, colMeta)
	}
	if err = d.validateHeaderUniqueness(colsMeta); err != nil {
		return nil, err
	}

	if d.hasFixedInlineColumns || d.hasDynamicInlineColumns {
		if err = d.validateConfigOnInlineColumns(fileHeader); err != nil {
			return nil, err
		}
		// Parse dynamic inline columns based on file header
		if d.hasDynamicInlineColumns {
			colsMeta, err = d.parseDynamicInlineColumns(colsMeta, fileHeader)
			if err != nil {
				return nil, err
			}
		}
	}

	// Correct column index (0-index)
	for i, colMeta := range colsMeta {
		colMeta.column = i
	}
	return colsMeta, err
}

func (d *Decoder) parseInlineColumn(field reflect.StructField, parentCol *decodeColumnMeta) (
	colsMeta []*decodeColumnMeta, err error) {
	inlineColumnsMeta, err := d.parseInlineColumnDynamicType(field.Type, parentCol)
	if err == nil {
		d.hasDynamicInlineColumns = true
		return inlineColumnsMeta, nil
	}
	inlineColumnsMeta, err = d.parseInlineColumnFixedType(field.Type, parentCol)
	if err == nil && len(inlineColumnsMeta) > 0 {
		d.hasFixedInlineColumns = true
		return inlineColumnsMeta, nil
	}
	return nil, fmt.Errorf("%w: %v", ErrHeaderDynamicTypeInvalid, field.Type)
}

func (d *Decoder) parseInlineColumnFixedType(typ reflect.Type, parent *decodeColumnMeta) ([]*decodeColumnMeta, error) {
	cfg := d.cfg
	typ = indirectType(typ)
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: not struct type", ErrHeaderDynamicTypeInvalid)
	}
	numFields := typ.NumField()
	colsMeta := make([]*decodeColumnMeta, 0, numFields)
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		tag, err := parseTag(cfg.TagName, field)
		if err != nil {
			return nil, err
		}
		if tag == nil || tag.ignored {
			continue
		}

		headerKey := parent.prefix + tag.name
		colMeta := &decodeColumnMeta{
			column:      len(colsMeta),
			headerKey:   headerKey,
			headerText:  headerKey,
			parentKey:   parent.headerKey,
			optional:    tag.optional,
			omitempty:   tag.omitEmpty,
			targetField: parent.targetField,
			inlineColumnMeta: &inlineColumnMeta{
				inlineType:  inlineColumnStructFixed,
				targetField: field,
				dataType:    field.Type,
			},
		}

		columnCfg := cfg.columnConfigMap[colMeta.headerKey]
		if columnCfg == nil {
			columnCfg = cfg.columnConfigMap[colMeta.parentKey]
		}
		colMeta.copyConfig(columnCfg)
		if err = colMeta.localizeHeader(cfg); err != nil {
			return nil, err
		}

		colsMeta = append(colsMeta, colMeta)
	}
	return colsMeta, nil
}

func (d *Decoder) parseInlineColumnDynamicType(typ reflect.Type, parent *decodeColumnMeta) (
	[]*decodeColumnMeta, error) {
	cfg := d.cfg
	typ = indirectType(typ)
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: not struct type", ErrHeaderDynamicTypeInvalid)
	}
	headerField, ok := typ.FieldByName(dynamicInlineColumnHeader)
	if !ok {
		return nil, fmt.Errorf("%w: field Header not found", ErrHeaderDynamicTypeInvalid)
	}
	if headerField.Type != reflect.TypeOf([]string{}) {
		return nil, fmt.Errorf("%w: field Header not []string", ErrHeaderDynamicTypeInvalid)
	}

	valuesField, ok := typ.FieldByName(dynamicInlineColumnValues)
	if !ok {
		return nil, fmt.Errorf("%w: field Values not found", ErrHeaderDynamicTypeInvalid)
	}
	if valuesField.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("%w: field Values not slice", ErrHeaderDynamicTypeInvalid)
	}

	dataType := valuesField.Type.Elem()
	colMeta := *parent
	colMeta.headerKey = parent.prefix + parent.headerKey
	colMeta.parentKey = parent.headerKey
	colMeta.inlineColumnMeta = &inlineColumnMeta{
		inlineType:  inlineColumnStructDynamic,
		targetField: valuesField,
		dataType:    dataType,
	}

	columnCfg := cfg.columnConfigMap[colMeta.headerKey]
	if columnCfg == nil {
		columnCfg = cfg.columnConfigMap[colMeta.parentKey]
	}
	colMeta.copyConfig(columnCfg)

	return []*decodeColumnMeta{&colMeta}, nil
}

func (d *Decoder) parseDynamicInlineColumns(colsMetaFromStruct []*decodeColumnMeta, fileHeader []string) (
	[]*decodeColumnMeta, error) {
	newColsMetaFromStruct := make([]*decodeColumnMeta, 0, len(colsMetaFromStruct)*2) // nolint: gomnd
	fileHeaderIndex := 0
	for i, colMetaFromStruct := range colsMetaFromStruct {
		if colMetaFromStruct.inlineColumnMeta == nil {
			colOptional := colMetaFromStruct.optional
			expectHeader := colMetaFromStruct.headerText
			if fileHeaderIndex < len(fileHeader) && fileHeader[fileHeaderIndex] == expectHeader {
				newColsMetaFromStruct = append(newColsMetaFromStruct, colMetaFromStruct)
				fileHeaderIndex++
				continue
			}
			if colOptional {
				newColsMetaFromStruct = append(newColsMetaFromStruct, colMetaFromStruct)
				continue
			}
			return nil, fmt.Errorf("%w: \"%s\"", ErrHeaderDynamicNotAllowUnrecognizedColumns, expectHeader)
		}

		inlineColumnMeta := colMetaFromStruct.inlineColumnMeta
		for j := fileHeaderIndex; j < len(fileHeader); j++ {
			if (i+1) < len(colsMetaFromStruct) && colsMetaFromStruct[i+1].headerText == fileHeader[j] {
				break
			}
			newColMeta := *colMetaFromStruct
			newColMeta.headerText = fileHeader[j]
			inlineColumnMeta.headerText = append(inlineColumnMeta.headerText, newColMeta.headerText)
			newColsMetaFromStruct = append(newColsMetaFromStruct, &newColMeta)
			fileHeaderIndex++
		}
	}
	return newColsMetaFromStruct, nil
}

// buildColumnDecoders build decoders for each column type
// If the type of column is determined, e.g. `int`, the decode function for that will be determined at
// the prepare step, and it will be fast at decoding. If it is `interface`, the decode function will parse
// the actual type at decoding, and it will be slower.
func (d *Decoder) buildColumnDecoders() error {
	for _, colMeta := range d.colsMeta {
		if colMeta.decodeFunc != nil || colMeta.unrecognized {
			continue
		}
		dataType := colMeta.targetField.Type
		if colMeta.inlineColumnMeta != nil {
			dataType = colMeta.inlineColumnMeta.dataType
		}
		decodeFunc, err := getDecodeFunc(dataType)
		if err != nil {
			return err
		}
		colMeta.decodeFunc = decodeFunc
	}
	return nil
}

// validateHeaderUniqueness validate to make sure header columns are unique
func (d *Decoder) validateHeaderUniqueness(colsMeta []*decodeColumnMeta) error {
	mapCheckUniq := make(map[string]struct{}, len(colsMeta))
	for _, colMeta := range colsMeta {
		h := colMeta.headerKey
		hh := strings.TrimSpace(h)
		if h != hh || len(hh) == 0 {
			return fmt.Errorf("%w: \"%s\" invalid", ErrHeaderColumnInvalid, h)
		}
		isDynamicInline := colMeta.inlineColumnMeta != nil &&
			colMeta.inlineColumnMeta.inlineType == inlineColumnStructDynamic
		if _, ok := mapCheckUniq[hh]; ok && !isDynamicInline {
			return fmt.Errorf("%w: \"%s\" duplicated", ErrHeaderColumnDuplicated, h)
		}
		mapCheckUniq[hh] = struct{}{}
	}
	return nil
}

// validateHeaderOrder validate to make sure the order of columns in the struct must match
// the order of columns in the input data
func (d *Decoder) validateHeaderOrder(colsMeta, colsMetaFromStruct []*decodeColumnMeta) error {
	mapColMeta := make(map[string]*decodeColumnMeta, len(colsMeta))
	for _, colMeta := range colsMeta {
		mapColMeta[colMeta.headerText] = colMeta
	}

	header := make([]string, 0, len(colsMeta))
	for _, colMeta := range colsMeta {
		if colMeta.unrecognized {
			continue
		}
		header = append(header, colMeta.headerText)
	}

	headerFromStruct := make([]string, 0, len(colsMetaFromStruct))
	for _, colMeta := range colsMetaFromStruct {
		if colMeta.optional && mapColMeta[colMeta.headerText] == nil {
			continue
		}
		headerFromStruct = append(headerFromStruct, colMeta.headerText)
	}

	if !reflect.DeepEqual(header, headerFromStruct) {
		return fmt.Errorf("%w: %v (expect %v)", ErrHeaderColumnOrderInvalid, header, headerFromStruct)
	}
	return nil
}

// validateConfig validate the configuration sent from user
func (d *Decoder) validateConfig() error {
	if d.cfg.ParseLocalizedHeader && d.cfg.LocalizationFunc == nil {
		return fmt.Errorf("%w: localization function required", ErrConfigOptionInvalid)
	}

	return nil
}

// validateConfigOnInlineColumns validate the configuration on inline columns
func (d *Decoder) validateConfigOnInlineColumns(fileHeader []string) error {
	cfg := d.cfg
	// If file has inline columns, there are some restrictions
	if cfg.NoHeaderMode || len(fileHeader) == 0 {
		return ErrHeaderDynamicNotAllowNoHeaderMode
	}
	if d.hasDynamicInlineColumns && !cfg.RequireColumnOrder {
		return ErrHeaderDynamicRequireColumnOrder
	}
	if d.hasDynamicInlineColumns && cfg.AllowUnrecognizedColumns {
		return ErrHeaderDynamicNotAllowUnrecognizedColumns
	}
	if d.hasDynamicInlineColumns && cfg.ParseLocalizedHeader {
		return ErrHeaderDynamicNotAllowLocalizedHeader
	}
	return nil
}

// rowData input data of each row
// `line` can be different from `row`, as a row can be in multiple rows and empty lines are skipped
type rowData struct {
	records []string
	line    int
	row     int
	err     error
}

// decodeColumnMeta metadata for decoding a specific column
type decodeColumnMeta struct {
	column       int
	headerKey    string
	headerText   string
	parentKey    string
	prefix       string
	optional     bool
	unrecognized bool
	omitempty    bool
	trimSpace    bool
	stopOnError  bool

	targetField      reflect.StructField
	inlineColumnMeta *inlineColumnMeta

	decodeFunc        DecodeFunc
	preprocessorFuncs []ProcessorFunc
	validatorFuncs    []ValidatorFunc
	onCellErrorFunc   OnCellErrorFunc
}

func (m *decodeColumnMeta) localizeHeader(cfg *DecodeConfig) error {
	if cfg.ParseLocalizedHeader {
		headerText, err := cfg.LocalizationFunc(m.headerKey, nil)
		if err != nil {
			return multierror.Append(ErrLocalization, err)
		}
		m.headerText = headerText
	}
	return nil
}

func (m *decodeColumnMeta) copyConfig(columnCfg *DecodeColumnConfig) {
	if columnCfg == nil {
		return
	}
	m.trimSpace = columnCfg.TrimSpace
	m.stopOnError = columnCfg.StopOnError
	m.decodeFunc = columnCfg.DecodeFunc
	m.validatorFuncs = columnCfg.ValidatorFuncs
	m.preprocessorFuncs = columnCfg.PreprocessorFuncs
	m.onCellErrorFunc = columnCfg.OnCellErrorFunc
}
