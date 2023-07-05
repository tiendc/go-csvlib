# Decoding CSV

## Index
- [First example](#first-example)
- [Preprocessor and Validator](#preprocessor-and-validator)
- [When StopOnError is false](#when-stoponerror-is-false)
- [Optional and Unrecognized columns](#optional-and-unrecognized-columns)
- [Allow unordered header columns](#allow-unordered-header-columns)
- [Fixed inline columns](#fixed-inline-columns)
- [Dynamic inline columns](#dynamic-inline-columns)
- [Custom unmarshaler](#custom-unmarshaler)
- [Custom column delimiter](#custom-column-delimiter)
- [Header localization](#header-localization)
- [Render error as human-readable format](#render-error-as-human-readable-format)

## Content

### First example

```go
    data := []byte(`
name,birthdate,address
jerry,1990-01-01T15:00:00Z,
tom,1989-11-11T00:00:00Z,new york`)
    
    type Student struct {
        Name      string    `csv:"name"`
        Birthdate time.Time `csv:"birthdate"`
        Address   string    `csv:"address,omitempty"`
    }
    
    var students []Student
    result, err := csvlib.Unmarshal(data, &students)
    if err != nil {
        fmt.Println("error:", err)
    }
    
    fmt.Printf("%+v\n", *result)
    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }
    
    // Output:
    // {totalRow:3 unrecognizedColumns:[] missingOptionalColumns:[]}
    // {Name:jerry Birthdate:1990-01-01 15:00:00 +0000 UTC Address:}
    // {Name:tom Birthdate:1989-11-11 00:00:00 +0000 UTC Address:new york}
```

### Preprocessor and Validator

- Preprocessor functions will be called before cell values are decoded and validator functions will be called after that.

```go
    data := []byte(`
name,age,address
jerry ,20,
tom ,26,new york`)

    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address,omitempty"`
    }

    var students []Student
    result, err := csvlib.Unmarshal(data, &students, func(cfg *csvlib.DecodeConfig) {
        cfg.ConfigureColumn("name", func(cfg *csvlib.DecodeColumnConfig) {
            cfg.TrimSpace = true // you can use csvlib.ProcessorTrim as an alternative
            cfg.PreprocessorFuncs = []csvlib.ProcessorFunc{csvlib.ProcessorUpper}
        })

        cfg.ConfigureColumn("age", func(cfg *csvlib.DecodeColumnConfig) {
            cfg.ValidatorFuncs = []csvlib.ValidatorFunc{csvlib.ValidatorRange(20, 30)}
        })
    })
    if err != nil {
        fmt.Println("error:", err)
    }

    fmt.Printf("%+v\n", *result)
    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }

    // Success: when use ValidatorRange(20, 30) as above
    // {Name:JERRY Age:20 Address:}
    // {Name:TOM Age:26 Address:new york}

    // Failure: when use ValidatorRange(10, 20) instead
    // error: ErrValidation: Range
```

### When StopOnError is false

- When set `StopOnError = false`, the decoding will continue to process the data even when errors occur. You can handle all errors of the process at once.

```go
    data := []byte(`
name,age,address
jerry,30,
tom,26,new york`)

    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address,omitempty"`
    }

    var students []Student
    result, err := csvlib.Unmarshal(data, &students, func(cfg *csvlib.DecodeConfig) {
        cfg.StopOnError = false
        cfg.ConfigureColumn("age", func(cfg *csvlib.DecodeColumnConfig) {
            cfg.ValidatorFuncs = []csvlib.ValidatorFunc{csvlib.ValidatorRange(10, 20)}
        })
    })
    if err != nil {
        csvErr := err.(*csvlib.Errors)
        fmt.Println("error 1:", csvErr.Unwrap()[0])
        fmt.Println("error 2:", csvErr.Unwrap()[1])
    }

    // Output:
    // error 1: ErrValidation: Range
    // error 2: ErrValidation: Range
```

### Optional and Unrecognized columns

- Optional column is a column which is defined in the struct tag but not exist in the input data
- Unrecognized column is a column which exists in the input data but is not defined in the struct tag

```go
    data := []byte(`
name,age,mark
jerry,20,10
tom,26,9`)

    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address,optional"` // without `optional` tag, it will fail to decode
    }

    var students []Student
    result, err := csvlib.Unmarshal(data, &students, func(cfg *csvlib.DecodeConfig) {
        cfg.AllowUnrecognizedColumns = true // without this config, it will fail to decode
    })
    if err != nil {
        fmt.Println("error:", err)
    }

    fmt.Printf("%+v\n", *result)
    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }

    // Output:
    // {totalRow:3 unrecognizedColumns:[mark] missingOptionalColumns:[address]}
    // {Name:jerry Age:20 Address:}
    // {Name:tom Age:26 Address:}
```

### Allow unordered header columns

- By default, the header order in the input data must match the order defined in the struct tag.

```go
    // `address` appears before `age`
    data := []byte(`
name,address,age
jerry,tokyo,10
tom,new york,9`)

    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address,optional"`
    }

    var students []Student
    result, err := csvlib.Unmarshal(data, &students, func(cfg *csvlib.DecodeConfig) {
        cfg.RequireColumnOrder = false // without this config, it will fail to decode
    })
    if err != nil {
        fmt.Println("error:", err)
    }

    fmt.Printf("%+v\n", *result)
    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }
    
    // Output:
    // {Name:jerry Age:10 Address:tokyo}
    // {Name:tom Age:9 Address:new york}
```

### Fixed inline columns

- Fixed inline columns are represented by an inner struct. `prefix` can be set for inline columns.

```go
    data := []byte(`
name,age,mark_math,mark_chemistry,mark_physics
jerry,20,9,8,7
tom,19,7,8,9`)

    type Marks struct {
        Math      int `csv:"math"`
        Chemistry int `csv:"chemistry"`
        Physics   int `csv:"physics"`
    }
    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address,optional"`
        Marks   Marks  `csv:"marks,inline,prefix=mark_"`
    }

    var students []Student
    result, err := csvlib.Unmarshal(data, &students, func(cfg *csvlib.DecodeConfig) {
        cfg.ConfigureColumn("mark_math", func(cfg *csvlib.DecodeColumnConfig) { // `math` field will be decoded with these configurations
            cfg.ValidatorFuncs = []csvlib.ValidatorFunc{csvlib.ValidatorRange(0, 10)}
        })
        cfg.ConfigureColumn("marks", func(cfg *csvlib.DecodeColumnConfig) { // `chemistry` and `physics` fields will be decoded with these configurations
            cfg.ValidatorFuncs = []csvlib.ValidatorFunc{csvlib.ValidatorRange(1, 10)}
        })
    })
    if err != nil {
        fmt.Println("error:", err)
    }

    fmt.Printf("%+v\n", *result)
    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }
    
    // Output:
    // {Name:jerry Age:20 Address: Marks:{Math:9 Chemistry:8 Physics:7}}
    // {Name:tom Age:19 Address: Marks:{Math:7 Chemistry:8 Physics:9}}
```

### Dynamic inline columns

- Dynamic inline columns are represented by an inner struct with variable number of columns. `prefix` can also be set for them.

```go
    data := []byte(`
name,age,mark_math,mark_chemistry,mark_physics
jerry,20,9,8,7
tom,19,7,8,9`)

    type Student struct {
        Name    string                   `csv:"name"`
        Age     int                      `csv:"age"`
        Address string                   `csv:"address,optional"`
        Marks   csvlib.InlineColumn[int] `csv:"marks,inline,prefix=mark_"`
    }

    var students []Student
    result, err := csvlib.Unmarshal(data, &students, func(cfg *csvlib.DecodeConfig) {
        cfg.ConfigureColumn("marks", func(cfg *csvlib.DecodeColumnConfig) { // all inline columns are validated by this
            cfg.ValidatorFuncs = []csvlib.ValidatorFunc{csvlib.ValidatorRange(1, 10)}
        })
    })
    if err != nil {
        fmt.Println("error:", err)
    }

    fmt.Printf("%+v\n", *result)
    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }
    
    // Output:
    // {Name:jerry Age:20 Address: Marks:{Header:[mark_math mark_chemistry mark_physics] Values:[9 8 7]}}
    // {Name:tom Age:19 Address: Marks:{Header:[mark_math mark_chemistry mark_physics] Values:[7 8 9]}}
```

### Custom unmarshaler

- Any user custom type can be decoded by implementing either `encoding.TextUnmarshaler` or `csvlib.CSVUnmarshaler` or both. `csvlib.CSVUnmarshaler` has higher priority.

```go
type BirthDate time.Time

func (d *BirthDate) UnmarshalCSV(data []byte) error {
    t, err := time.Parse("2006-1-2", string(data)) // RFC3339 format
    if err != nil {
        return err
    }
    *d = BirthDate(t)
    return nil
}

func (d BirthDate) String() string { // this is for displaying the result only
    return time.Time(d).Format("2006-01-02")
}
```
```go
    data := []byte(`
name,birthdate
jerry,1990-01-01
tom,1989-11-11`)

    type Student struct {
        Name      string    `csv:"name"`
        BirthDate BirthDate `csv:"birthdate"`
    }

    var students []Student
    result, err := csvlib.Unmarshal(data, &students)
    if err != nil {
        fmt.Println("error:", err)
    }

    fmt.Printf("%+v\n", *result)
    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }
    
    // Output:
    // {Name:jerry BirthDate:1990-01-01}
    // {Name:tom BirthDate:1989-11-11}
```

### Custom column delimiter

- By default, the decoder detects columns via comma delimiter, if you want to decode custom one, use `csv.Reader` from the built-in package `encoding/csv` as the input reader.

```go
    data := []byte(
        "name\tage\taddress\n" +
        "jerry\t10\ttokyo\n" +
        "tom\t9\tnew york\n",
    )

    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address,optional"`
    }

    var students []Student
    reader := csv.NewReader(bytes.NewReader(data))
    reader.Comma = '\t'
    _, err := csvlib.NewDecoder(reader).Decode(&students)
    if err != nil {
        fmt.Println("error:", err)
    }

    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }
    
    // Output:
    // {Name:jerry Age:10 Address:tokyo}
    // {Name:tom Age:9 Address:new york}
```

### Header localization

- This functionality allows to decode multiple input data with header translated into specific language

```go
// Sample localization data (you can define them somewhere else such as in a json file)
var (
    mapLanguageEn = map[string]string{
        "name":    "name",
        "age":     "age",
        "address": "address",
    }

    mapLanguageVi = map[string]string{
        "name":    "tên",
        "age":     "tuổi",
        "address": "địa chỉ",
    }
)

func localizeViVn(k string, params csvlib.ParameterMap) (string, error) {
    return mapLanguageVi[k], nil
}

func localizeEnUs(k string, params csvlib.ParameterMap) (string, error) {
    return mapLanguageEn[k], nil
}
```

```go
    dataEn := []byte(`
name,age,address
jerry,20,tokyo
tom,19,new york`)

    type Student struct {
        Name    string `csv:"name"` // this tag is now used as localization key to find the real header
        Age     int    `csv:"age"`
        Address string `csv:"address"`
    }

    var students []Student
    result, _ := csvlib.Unmarshal(dataEn, &students, func(cfg *csvlib.DecodeConfig) {
        cfg.ParseLocalizedHeader = true
        cfg.LocalizationFunc = localizeEnUs
    })

    fmt.Printf("%+v\n", *result)
    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }

    // Output:
    // {Name:jerry Age:20 Address:tokyo}
    // {Name:tom Age:19 Address:new york}
    
    dataVi := []byte(`
tên,tuổi,địa chỉ
jerry,20,tokyo
tom,19,new york`)

    result, _ = csvlib.Unmarshal(dataVi, &students, func(cfg *csvlib.DecodeConfig) {
        cfg.ParseLocalizedHeader = true
        cfg.LocalizationFunc = localizeViVn
    })

    fmt.Printf("%+v\n", *result)
    for _, u := range students {
        fmt.Printf("%+v\n", u)
    }

    // Output: the same
    // {Name:jerry Age:20 Address:tokyo}
    // {Name:tom Age:19 Address:new york}
```

### Render error as human-readable format

- Decoding errors can be rendered as a more human-readable content such as text or CSV.

```go
    data := []byte(`
name,age,address
jerry,20,tokyo
tom,26,new york
tintin from paris,32,paris
jj,40,`)

    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address"`
    }

    var students []Student
    _, err := csvlib.Unmarshal(data, &students, func(cfg *csvlib.DecodeConfig) {
        cfg.StopOnError = false
        cfg.DetectRowLine = true
        cfg.ConfigureColumn("name", func(cfg *csvlib.DecodeColumnConfig) {
            cfg.ValidatorFuncs = []csvlib.ValidatorFunc{csvlib.ValidatorRange(1, 10)}
            cfg.OnCellErrorFunc = func(e *csvlib.CellError) {
                if errors.Is(e, csvlib.ErrValidationStrLen) {
                    e.SetLocalizationKey("Column {{.Column}} - '{{.Value}}': Name length must be from {{.MinLen}} to {{.MaxLen}}")
                    e.WithParam("MinLen", 3).WithParam("MaxLen", 10)
                }
            }
        })
        cfg.ConfigureColumn("age", func(cfg *csvlib.DecodeColumnConfig) {
            cfg.ValidatorFuncs = []csvlib.ValidatorFunc{csvlib.ValidatorRange(10, 30)}
            cfg.OnCellErrorFunc = func(e *csvlib.CellError) {
                if errors.Is(e, csvlib.ErrValidationRange) {
                    e.SetLocalizationKey("Column {{.Column}} - '{{.Value}}': Age must be from {{.MinVal}} to {{.MaxVal}}")
                    e.WithParam("MinVal", 10).WithParam("MaxVal", 30)
                }
            }
        })
    })
    if err != nil {
        renderer, _ := csvlib.NewRenderer(err.(*csvlib.Errors))
        msg, transErr, _ := renderer.Render()
        fmt.Println(transErr) // Translation error can be ignored if that is not important to you
        fmt.Println(msg)
    }
    
    // Output:
    // Error content: TotalRow: 5, TotalRowError: 2, TotalCellError: 4, TotalError: 4
    // Row 4 (line 5): Column 0 - 'tintin from paris': Name length must be from 3 to 10, Column 1 - '32': Age must be from 10 to 30
    // Row 5 (line 6): Column 0 - 'jj': Name length must be from 3 to 10, Column 1 - '40': Age must be from 10 to 30
```

- Render error as CSV content.


```go
    renderer, _ := csvlib.NewCSVRenderer(err.(*csvlib.Errors), func(cfg *csvlib.CSVRenderConfig) {
        cfg.CellSeparator = "\n"
    })
    msg, _, _ := renderer.RenderAsString()
    fmt.Println(msg)

    // Output:
    // |------|--------|--------------|--------------------------------------------------------|----------------------------------|----------|                            
    // |  Row |  Line  | CommonError  |  name                                                  | age                              | address  |                            
    // |------|--------|--------------|--------------------------------------------------------|----------------------------------|----------|                            
    // |  4   |  5     |              |  'tintin from paris': Name length must be from 3 to 10 | '32': Age must be from 10 to 30  |          |                            
    // |------|--------|--------------|--------------------------------------------------------|----------------------------------|----------|                            
    // |  5   |  6     |              |  'jj': Name length must be from 3 to 10                | '40': Age must be from 10 to 30  |          |                            
    // |------|--------|--------------|--------------------------------------------------------|----------------------------------|----------|
```