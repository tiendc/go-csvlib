package csvlib

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/tiendc/gofn"
)

type EncodeConfig struct {
	// TagName tag name to parse the struct (default is `csv`)
	TagName string

	// NoHeaderMode indicates whether to write header or not (default is `false`)
	NoHeaderMode bool

	// LocalizeHeader indicates whether to localize the header or not (default is `false`)
	LocalizeHeader bool

	// LocalizationFunc localization function, required when LocalizeHeader is true
	LocalizationFunc LocalizationFunc

	// columnConfigMap a map consists of configuration for specific columns (optional)
	columnConfigMap map[string]*EncodeColumnConfig
}

func defaultEncodeConfig() *EncodeConfig {
	return &EncodeConfig{
		TagName: DefaultTagName,
	}
}

func (c *EncodeConfig) ConfigureColumn(name string, fn func(*EncodeColumnConfig)) {
	if c.columnConfigMap == nil {
		c.columnConfigMap = map[string]*EncodeColumnConfig{}
	}
	columnCfg, ok := c.columnConfigMap[name]
	if !ok {
		columnCfg = defaultEncodeColumnConfig()
		c.columnConfigMap[name] = columnCfg
	}
	fn(columnCfg)
}

type EncodeColumnConfig struct {
	// Skip whether skip encoding the column or not (this is equivalent to use `csv:"-"` in struct tag)
	// (default is "false")
	Skip bool

	// EncodeFunc custom encode function (optional)
	EncodeFunc EncodeFunc

	// PostprocessorFuncs a list of functions will be called after encoding a cell value (optional)
	PostprocessorFuncs []ProcessorFunc
}

func defaultEncodeColumnConfig() *EncodeColumnConfig {
	return &EncodeColumnConfig{}
}

type EncodeOption func(cfg *EncodeConfig)

type Encoder struct {
	w                       Writer
	cfg                     *EncodeConfig
	err                     error
	finished                bool
	itemType                reflect.Type
	headerWritten           bool
	hasDynamicInlineColumns bool
	hasFixedInlineColumns   bool
	colsMeta                []*encodeColumnMeta
}

func NewEncoder(w Writer, options ...EncodeOption) *Encoder {
	cfg := defaultEncodeConfig()
	for _, opt := range options {
		opt(cfg)
	}
	return &Encoder{
		w:   w,
		cfg: cfg,
	}
}

// Encode encode input data stored in the given variable
// The input var must be a slice, e.g. `[]Student` or `[]*Student`
func (e *Encoder) Encode(v any) error {
	if e.finished {
		return ErrFinished
	}
	if e.err != nil {
		return ErrAlreadyFailed
	}

	val := reflect.ValueOf(v)
	if e.itemType == nil {
		if err := e.prepareEncode(val); err != nil {
			e.err = err
			return err
		}
	} else {
		itemType, err := e.parseInputVar(val)
		if err != nil {
			return err
		}
		if itemType != e.itemType {
			return fmt.Errorf("%w: %v (expect %v)", ErrTypeUnmatched, itemType, e.itemType)
		}
	}

	totalRow := val.Len()
	itemKindIsPtr := e.itemType.Kind() == reflect.Pointer
	for row := 0; row < totalRow; row++ {
		rowVal := val.Index(row)
		if itemKindIsPtr {
			if rowVal.IsNil() {
				continue
			}
			rowVal = rowVal.Elem()
		}
		if err := e.encodeRow(rowVal); err != nil {
			e.err = err
			break
		}
	}
	return e.err
}

// EncodeOne encode single object into a single CSV row
func (e *Encoder) EncodeOne(v any) error {
	if e.finished {
		return ErrFinished
	}
	if e.err != nil {
		return ErrAlreadyFailed
	}

	rowVal := reflect.ValueOf(v)
	itemType := rowVal.Type()
	if !isKindOrPtrOf(itemType, reflect.Struct) {
		return fmt.Errorf("%w: must be a struct", ErrTypeInvalid)
	}
	if e.itemType == nil {
		slice := reflect.MakeSlice(reflect.SliceOf(itemType), 1, 1)
		slice.Index(0).Set(rowVal)
		err := e.prepareEncode(slice)
		if err != nil {
			e.err = err
			return err
		}
	} else if itemType != e.itemType {
		return fmt.Errorf("%w: %v (expect %v)", ErrTypeUnmatched, itemType, e.itemType)
	}

	if err := e.encodeRow(rowVal); err != nil {
		e.err = err
		return err
	}
	return nil
}

func (e *Encoder) Finish() error {
	e.finished = true
	return e.err
}

func (e *Encoder) prepareEncode(v reflect.Value) error {
	if e.itemType != nil {
		return fmt.Errorf("%w: item type already parsed", ErrUnexpected)
	}
	itemType, err := e.parseInputVar(v)
	if err != nil {
		return err
	}
	e.itemType = itemType

	if err = e.validateConfig(); err != nil {
		return err
	}

	if err = e.parseColumnsMeta(itemType, v); err != nil {
		return err
	}

	if err = e.buildColumnEncoders(); err != nil {
		return err
	}

	if err = e.encodeHeader(); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encodeHeader() error {
	if e.headerWritten {
		return fmt.Errorf("%w: header already encoded", ErrUnexpected)
	}
	record := make([]string, 0, len(e.colsMeta))
	for _, colMeta := range e.colsMeta {
		if colMeta.skipColumn {
			continue
		}
		record = append(record, colMeta.headerText)
	}
	if err := validateHeader(record); err != nil {
		return err
	}
	if e.cfg.NoHeaderMode {
		e.headerWritten = true
		return nil
	} else {
		err := e.w.Write(record)
		e.headerWritten = err == nil
		return err
	}
}

func (e *Encoder) encodeRow(rowVal reflect.Value) error {
	colsMeta := e.colsMeta
	if e.hasDynamicInlineColumns || e.hasFixedInlineColumns {
		for _, colMeta := range colsMeta {
			if colMeta.inlineColumnMeta != nil {
				colMeta.inlineColumnMeta.encodePrepareForNextRow()
			}
		}
	}

	record := make([]string, 0, len(colsMeta))
	for _, colMeta := range colsMeta {
		if colMeta.skipColumn {
			continue
		}
		colVal := colMeta.getColumnValue(rowVal)
		if !colVal.IsValid() {
			record = append(record, "")
			continue
		}
		text, err := colMeta.encodeFunc(colVal, colMeta.omitEmpty)
		if err != nil {
			return err
		}
		for _, fn := range colMeta.postprocessorFuncs {
			text = fn(text)
		}
		record = append(record, text)
	}
	return e.w.Write(record)
}

func (e *Encoder) parseInputVar(v reflect.Value) (itemType reflect.Type, err error) {
	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		err = fmt.Errorf("%w: %v", ErrTypeInvalid, kind)
		return
	}
	if v.IsNil() {
		err = fmt.Errorf("%w: must be non-nil pointer", ErrValueNil)
		return
	}

	typ := v.Type()
	itemType = typ.Elem() // E.g. val: []Item, typ: []Item, itemType: Item

	if indirectType(itemType).Kind() != reflect.Struct {
		err = fmt.Errorf("%w: %v", ErrTypeInvalid, itemType.Kind())
		return
	}
	return
}

func (e *Encoder) validateConfig() error {
	if e.cfg.LocalizeHeader && e.cfg.LocalizationFunc == nil {
		return fmt.Errorf("%w: localization function required", ErrConfigOptionInvalid)
	}
	return nil
}

func (e *Encoder) parseColumnsMeta(itemType reflect.Type, val reflect.Value) error {
	colsMeta, err := e.parseColumnsMetaFromStructType(itemType, val)
	if err != nil {
		return err
	}

	if err = e.validateColumnsMeta(colsMeta); err != nil {
		return err
	}

	e.colsMeta = colsMeta
	return nil
}

func (e *Encoder) validateColumnsMeta(colsMeta []*encodeColumnMeta) error {
	cfg := e.cfg
	// Make sure all column options valid
	for colKey := range cfg.columnConfigMap {
		if !gofn.ContainPred(colsMeta, func(colMeta *encodeColumnMeta) bool {
			return colMeta.headerKey == colKey || colMeta.parentKey == colKey
		}) {
			return fmt.Errorf("%w: column \"%s\" not found", ErrConfigOptionInvalid, colKey)
		}
	}

	// Make sure all columns are unique
	if err := e.validateHeaderUniqueness(colsMeta); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) parseColumnsMetaFromStructType(itemType reflect.Type, val reflect.Value) (
	colsMeta []*encodeColumnMeta, err error) {
	cfg := e.cfg

	// Get first row data from the input
	firstRowVal := reflect.Value{}
	if val.Len() > 0 {
		firstRowVal = val.Index(0)
	}
	if firstRowVal.IsValid() && firstRowVal.Kind() == reflect.Pointer {
		firstRowVal = firstRowVal.Elem()
	}

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

		colMeta := &encodeColumnMeta{
			column:      len(colsMeta),
			headerKey:   tag.name,
			headerText:  tag.name,
			prefix:      tag.prefix,
			omitEmpty:   tag.omitEmpty,
			targetField: field,
		}
		if tag.inline {
			inlineColsMeta, err := e.parseInlineColumn(field, colMeta, firstRowVal)
			if err != nil {
				return nil, err
			}
			colsMeta = append(colsMeta, inlineColsMeta...)
			continue
		}

		colMeta.copyConfig(cfg.columnConfigMap[colMeta.headerKey])
		if err = colMeta.localizeHeader(cfg); err != nil {
			return nil, err
		}

		colsMeta = append(colsMeta, colMeta)
	}

	for i, colMeta := range colsMeta {
		colMeta.column = i
	}
	return colsMeta, err
}

func (e *Encoder) parseInlineColumn(field reflect.StructField, parentCol *encodeColumnMeta, firstRowVal reflect.Value) (
	colsMeta []*encodeColumnMeta, err error) {
	if firstRowVal.IsValid() {
		inlineStruct := firstRowVal.Field(field.Index[0])
		inlineColumnsMeta, err := e.parseInlineColumnDynamicType(inlineStruct, parentCol)
		if err == nil {
			e.hasDynamicInlineColumns = true
			return inlineColumnsMeta, nil
		}
	}
	inlineColumnsMeta, err := e.parseInlineColumnFixedType(field.Type, parentCol)
	if err == nil && len(inlineColumnsMeta) > 0 {
		e.hasFixedInlineColumns = true
		return inlineColumnsMeta, nil
	}
	return nil, fmt.Errorf("%w: %v", ErrHeaderDynamicTypeInvalid, field.Type)
}

func (e *Encoder) parseInlineColumnFixedType(typ reflect.Type, parent *encodeColumnMeta) ([]*encodeColumnMeta, error) {
	cfg := e.cfg
	typ = indirectType(typ)
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: not struct type", ErrHeaderDynamicTypeInvalid)
	}
	numFields := typ.NumField()
	colsMeta := make([]*encodeColumnMeta, 0, numFields)
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
		colMeta := &encodeColumnMeta{
			column:      len(colsMeta),
			headerKey:   headerKey,
			headerText:  headerKey,
			parentKey:   parent.headerKey,
			targetField: parent.targetField,
			inlineColumnMeta: &inlineColumnMeta{
				inlineType:  inlineColumnStructFixed,
				targetField: field,
				dataType:    field.Type,
			},
		}

		columnCfg := cfg.columnConfigMap[headerKey]
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

func (e *Encoder) parseInlineColumnDynamicType(inlineStruct reflect.Value, parent *encodeColumnMeta) (
	[]*encodeColumnMeta, error) {
	cfg := e.cfg
	inlineStruct = indirectValue(inlineStruct)
	if !inlineStruct.IsValid() {
		return []*encodeColumnMeta{}, nil
	}
	if inlineStruct.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: not struct type", ErrHeaderDynamicTypeInvalid)
	}
	headerField := inlineStruct.FieldByName(dynamicInlineColumnHeader)
	if !headerField.IsValid() {
		return nil, fmt.Errorf("%w: field Header not found", ErrHeaderDynamicTypeInvalid)
	}
	if headerField.Type() != reflect.TypeOf([]string{}) {
		return nil, fmt.Errorf("%w: field Header not []string", ErrHeaderDynamicTypeInvalid)
	}

	valuesField, ok := inlineStruct.Type().FieldByName(dynamicInlineColumnValues)
	if !ok {
		return nil, fmt.Errorf("%w: field Values not found", ErrHeaderDynamicTypeInvalid)
	}
	if valuesField.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("%w: field Values not slice", ErrHeaderDynamicTypeInvalid)
	}

	dataType := valuesField.Type.Elem()
	header, _ := headerField.Interface().([]string)

	colsMeta := make([]*encodeColumnMeta, 0, len(header))
	inlineColumnMeta := &inlineColumnMeta{
		headerText:  header,
		inlineType:  inlineColumnStructDynamic,
		targetField: valuesField,
		dataType:    dataType,
	}
	for _, h := range header {
		headerKey := parent.prefix + h
		colMeta := *parent
		colMeta.headerKey = headerKey
		colMeta.headerText = headerKey
		colMeta.parentKey = parent.headerKey
		colMeta.inlineColumnMeta = inlineColumnMeta

		columnCfg := cfg.columnConfigMap[colMeta.headerKey]
		if columnCfg == nil {
			columnCfg = cfg.columnConfigMap[colMeta.parentKey]
		}
		colMeta.copyConfig(columnCfg)

		// Try to localize header (ignore the error when fail)
		_ = colMeta.localizeHeader(cfg)

		colsMeta = append(colsMeta, &colMeta)
	}

	return colsMeta, nil
}

func (e *Encoder) buildColumnEncoders() error {
	for _, colMeta := range e.colsMeta {
		if colMeta.encodeFunc != nil {
			continue
		}
		dataType := colMeta.targetField.Type
		if colMeta.inlineColumnMeta != nil {
			dataType = colMeta.inlineColumnMeta.dataType
		}
		encodeFunc, err := getEncodeFunc(dataType)
		if err != nil {
			return err
		}
		colMeta.encodeFunc = encodeFunc
	}
	return nil
}

// validateHeaderUniqueness validate to make sure header columns are unique
func (e *Encoder) validateHeaderUniqueness(colsMeta []*encodeColumnMeta) error {
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

type encodeColumnMeta struct {
	column     int
	headerKey  string
	headerText string
	parentKey  string
	prefix     string
	omitEmpty  bool
	skipColumn bool

	targetField      reflect.StructField
	inlineColumnMeta *inlineColumnMeta

	encodeFunc         EncodeFunc
	postprocessorFuncs []ProcessorFunc
}

func (m *encodeColumnMeta) localizeHeader(cfg *EncodeConfig) error {
	if cfg.LocalizeHeader {
		headerText, err := cfg.LocalizationFunc(m.headerKey, nil)
		if err != nil {
			return multierror.Append(ErrLocalization, err)
		}
		m.headerText = headerText
	}
	return nil
}

func (m *encodeColumnMeta) copyConfig(columnCfg *EncodeColumnConfig) {
	if columnCfg == nil {
		return
	}
	m.skipColumn = columnCfg.Skip
	m.encodeFunc = columnCfg.EncodeFunc
	m.postprocessorFuncs = columnCfg.PostprocessorFuncs
}

func (m *encodeColumnMeta) getColumnValue(rowVal reflect.Value) reflect.Value {
	colVal := rowVal.Field(m.targetField.Index[0])
	if m.inlineColumnMeta != nil {
		colVal = m.inlineColumnMeta.encodeGetColumnValue(colVal)
	}
	return colVal
}
