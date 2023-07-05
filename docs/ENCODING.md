# Encoding CSV

## Index
- [First example](#first-example)
- [No header mode](#no-header-mode)
- [Postprocessor](#postprocessor)
- [Skip columns](#skip-columns)
- [Fixed inline columns](#fixed-inline-columns)
- [Dynamic inline columns](#dynamic-inline-columns)
- [Custom marshaler](#custom-marshaler)
- [Custom column delimiter](#custom-column-delimiter)
- [Header localization](#header-localization)

## Content

### First example

```go
    type Student struct {
        Name      string    `csv:"name"`
        Birthdate time.Time `csv:"birthdate"`
        Address   string    `csv:"address,omitempty"`
    }

    students := []Student{
        {Name: "jerry", Birthdate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC), Address: "tokyo"},
        {Name: "tom", Birthdate: time.Date(1989, time.November, 11, 0, 0, 0, 0, time.UTC), Address: "new york"},
    }
    data, err := csvlib.Marshal(students)
    if err != nil {
        fmt.Println("error:", err)
    }
    fmt.Println(string(data))
    
    // Output:
    // name,birthdate,address
    // jerry,1990-01-01T00:00:00Z,tokyo
    // tom,1989-11-11T00:00:00Z,new york
```

### No header mode

```go
    type Student struct {
        Name      string    `csv:"name"`
        Birthdate time.Time `csv:"birthdate"`
        Address   string    `csv:"address,omitempty"`
    }

    students := []Student{
        {Name: "jerry", Birthdate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC), Address: "tokyo"},
        {Name: "tom", Birthdate: time.Date(1989, time.November, 11, 0, 0, 0, 0, time.UTC), Address: "new york"},
    }
    data, err := csvlib.Marshal(students, func(cfg *csvlib.EncodeConfig) {
        cfg.NoHeaderMode = true
    })
    if err != nil {
        fmt.Println("error:", err)
    }
    fmt.Println(string(data))

    // Output:
    // jerry,1990-01-01T00:00:00Z,tokyo
    // tom,1989-11-11T00:00:00Z,new york
```

### Postprocessor

- Postprocessor functions will be called after Go values are encoded into CSV string.

```go
    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address,omitempty"`
    }

    students := []Student{
        {Name: "jerry ", Age: 20, Address: "tokyo"},
        {Name: " tom ", Age: 19, Address: "new york"},
    }
    data, err := csvlib.Marshal(students, func(cfg *csvlib.EncodeConfig) {
        cfg.ConfigureColumn("name", func(cfg *csvlib.EncodeColumnConfig) {
            cfg.PostprocessorFuncs = []csvlib.ProcessorFunc{csvlib.ProcessorTrim, csvlib.ProcessorUpper}
        })
    })
    if err != nil {
        fmt.Println("error:", err)
    }
    fmt.Println(string(data))

    // Output:
    // name,age,address
    // JERRY,20,tokyo
    // TOM,19,new york
```

### Skip columns

- To skip encoding a column, you can omit the `csv` tag from the struct field or use `csv:"-"`. You can also use an equivalent configuration option for the column.

```go
    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"-"`
        Address string `csv:"address,omitempty"`
    }

    students := []Student{
        {Name: "jerry", Age: 20, Address: "tokyo"},
        {Name: "tom", Age: 19, Address: "new york"},
    }
    data, err := csvlib.Marshal(students, func(cfg *csvlib.EncodeConfig) {
        cfg.ConfigureColumn("address", func(cfg *csvlib.EncodeColumnConfig) {
            cfg.Skip = true
        })
    })
    if err != nil {
        fmt.Println("error:", err)
    }
    fmt.Println(string(data))

    // Output:
    // name
    // jerry
    // tom
```

### Fixed inline columns

- Fixed inline columns are represented by an inner struct. `prefix` can be set for inline columns.

```go
    type Marks struct {
        Math      int `csv:"math"`
        Chemistry int `csv:"-"`
        Physics   int `csv:"physics"`
    }
    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age,omitempty"`
        Address string `csv:"address,optional"`
        Marks   Marks  `csv:"marks,inline,prefix=mark_"`
    }

    students := []Student{
        {Name: "jerry", Age: 20, Address: "tokyo", Marks: Marks{Math: 9, Chemistry: 8, Physics: 7}},
        {Name: "tom", Age: 0, Address: "new york", Marks: Marks{Math: 7, Chemistry: 8, Physics: 9}},
    }
    data, err := csvlib.Marshal(students)
    if err != nil {
        fmt.Println("error:", err)
    }
    fmt.Println(string(data))
    
    // Output:
    // name,age,address,mark_math,mark_physics
    // jerry,20,tokyo,9,7
    // tom,,new york,7,9
```

### Dynamic inline columns

- Dynamic inline columns are represented by an inner struct with variable number of columns. `prefix` can also be set for them.

```go
    type Student struct {
        Name    string                   `csv:"name"`
        Age     int                      `csv:"age,omitempty"`
        Address string                   `csv:"address,optional"`
        Marks   csvlib.InlineColumn[int] `csv:"marks,inline,prefix=mark_"`
    }

    marksHeader := []string{"math", "chemistry", "physics"}
    students := []Student{
        {Name: "jerry", Age: 20, Address: "tokyo", Marks: csvlib.InlineColumn[int]{Header: marksHeader, Values: []int{9, 8, 7}}},
        {Name: "tom", Age: 0, Address: "new york", Marks: csvlib.InlineColumn[int]{Header: marksHeader, Values: []int{7, 8, 9}}},
    }
    data, err := csvlib.Marshal(students)
    if err != nil {
        fmt.Println("error:", err)
    }
    fmt.Println(string(data))

    // Output:
    // name,age,address,mark_math,mark_chemistry,mark_physics
    // jerry,20,tokyo,9,8,7
    // tom,,new york,7,8,9
```

### Custom marshaler

- Any user custom type can be decoded by implementing either `encoding.TextMarshaler` or `csvlib.CSVMarshaler` or both. `csvlib.CSVMarshaler` has higher priority.

```go
type BirthDate time.Time

func (d BirthDate) MarshalCSV() ([]byte, error) {
    return []byte(time.Time(d).Format("2006-02-01")), nil // RFC3339
}
```
```go
    type Student struct {
        Name      string    `csv:"name"`
        Birthdate BirthDate `csv:"birthdate"`
        Address   string    `csv:"address,omitempty"`
    }

    students := []Student{
        {Name: "jerry", Birthdate: BirthDate(time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)), Address: "tokyo"},
        {Name: "tom", Birthdate: BirthDate(time.Date(1989, time.November, 11, 0, 0, 0, 0, time.UTC)), Address: "new york"},
    }
    data, err := csvlib.Marshal(students)
    if err != nil {
        fmt.Println("error:", err)
    }
    fmt.Println(string(data))

    // Output:
    // name,birthdate,address
    // jerry,1990-01-01,tokyo
    // tom,1989-11-11,new york
```

### Custom column delimiter

- By default, the encoder uses comma as the delimiter, if you want to encode custom one, use `csv.Writer` from the built-in package `encoding/csv` as the input writer.

```go
    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address"`
    }

    students := []Student{
        {Name: "jerry", Age: 20, Address: "tokyo"},
        {Name: "tom", Age: 19, Address: "new york"},
    }

    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)
    writer.Comma = '\t'

    err := csvlib.NewEncoder(writer).Encode(students)
    if err != nil {
        fmt.Println("error:", err)
    }
    writer.Flush()
    fmt.Println(buf.String())

    // Output:
    // name	age	address
    // jerry	20	tokyo
    // tom	19	new york
```

### Header localization

- This functionality allows to encode CSV data with header translated into a specific language.

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
    type Student struct {
        Name    string `csv:"name"`
        Age     int    `csv:"age"`
        Address string `csv:"address,omitempty"`
    }
    
    students := []Student{
        {Name: "jerry", Age: 20, Address: "tokyo"},
        {Name: "tom", Age: 19, Address: "new york"},
    }
    data, err := csvlib.Marshal(students, func(cfg *csvlib.EncodeConfig) {
        cfg.LocalizeHeader = true
        cfg.LocalizationFunc = localizeViVn
    })
    if err != nil {
        fmt.Println("error:", err)
    }
    fmt.Println(string(data))
	
    // Output:
    // tên,tuổi,địa chỉ
    // jerry,20,tokyo
    // tom,19,new york
```
