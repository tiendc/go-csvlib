[![Go Version][gover-img]][gover] [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoReport][rpt-img]][rpt]

# High level CSV lib for Go 1.18+

This is a library for decoding and encoding CSV at high level as it provides convenient methods and many configuration options to process the data the way you expect.

## Functionalities

**Decoding**
  - Decode CSV data into Go struct
  - Support Go interface `encoding.TextUnmarshaler` (with function `UnmarshalText`)
  - Support custom interface `CSVUnmarshaler` (with function `UnmarshalCSV`)
  - Ability to continue decoding when error occurs (collect all errors at once)
  - Ability to perform custom preprocessor functions on cell data before decoding
  - Ability to perform custom validator functions on cell data after decoding
  - Ability to decode dynamic columns into Go struct field (inline columns)
  - Support rendering the result errors into human-readable content (row-by-row text and CSV)
  - Support localization to render the result errors into a specific language

**Encoding**
  - Encode Go struct into CSV data
  - Support Go interface `encoding.TextMarshaler` (with function `MarshalText`)
  - Support custom interface `CSVMarshaler` (with function `MarshalCSV`)
  - Ability to perform custom postprocessor functions on cell data after encoding
  - Ability to encode dynamic columns defined via inner Go struct (inline columns)
  - Ability to localize the header into a specific language

## Installation

```shell
go get github.com/tiendc/go-csvlib
```

## Usage

- [Decoding](docs/DECODING.md)
- [Encoding](docs/ENCODING.md)

## Benchmarks

### csvlib vs csvutil vs gocsv vs easycsv

[Benchmark code](https://gist.github.com/tiendc/c394677a846233bf8de819da3bb7093c)

### Unmarshal

```
BenchmarkUnmarshal/csvlib.Unmarshal/100_records
BenchmarkUnmarshal/csvlib.Unmarshal/100_records-10         	   20876	     55942 ns/op
BenchmarkUnmarshal/csvlib.Unmarshal/1000_records
BenchmarkUnmarshal/csvlib.Unmarshal/1000_records-10        	    2592	    462079 ns/op
BenchmarkUnmarshal/csvlib.Unmarshal/10000_records
BenchmarkUnmarshal/csvlib.Unmarshal/10000_records-10       	     252	   4766143 ns/op
BenchmarkUnmarshal/csvlib.Unmarshal/100000_records
BenchmarkUnmarshal/csvlib.Unmarshal/100000_records-10      	      26	  44989255 ns/op

BenchmarkUnmarshal/csvutil.Unmarshal/100_records
BenchmarkUnmarshal/csvutil.Unmarshal/100_records-10        	   27765	     43558 ns/op
BenchmarkUnmarshal/csvutil.Unmarshal/1000_records
BenchmarkUnmarshal/csvutil.Unmarshal/1000_records-10       	    2958	    405543 ns/op
BenchmarkUnmarshal/csvutil.Unmarshal/10000_records
BenchmarkUnmarshal/csvutil.Unmarshal/10000_records-10      	     296	   4030927 ns/op
BenchmarkUnmarshal/csvutil.Unmarshal/100000_records
BenchmarkUnmarshal/csvutil.Unmarshal/100000_records-10     	      28	  40479180 ns/op

BenchmarkUnmarshal/gocsv.Unmarshal/100_records
BenchmarkUnmarshal/gocsv.Unmarshal/100_records-10          	    9786	    118998 ns/op
BenchmarkUnmarshal/gocsv.Unmarshal/1000_records
BenchmarkUnmarshal/gocsv.Unmarshal/1000_records-10         	    1020	   1165965 ns/op
BenchmarkUnmarshal/gocsv.Unmarshal/10000_records
BenchmarkUnmarshal/gocsv.Unmarshal/10000_records-10        	      85	  12592398 ns/op
BenchmarkUnmarshal/gocsv.Unmarshal/100000_records
BenchmarkUnmarshal/gocsv.Unmarshal/100000_records-10       	       9	 120881625 ns/op

BenchmarkUnmarshal/easycsv.ReadAll/100_records
BenchmarkUnmarshal/easycsv.ReadAll/100_records-10          	    3772	    312755 ns/op
BenchmarkUnmarshal/easycsv.ReadAll/1000_records
BenchmarkUnmarshal/easycsv.ReadAll/1000_records-10         	     387	   3088959 ns/op
BenchmarkUnmarshal/easycsv.ReadAll/10000_records
BenchmarkUnmarshal/easycsv.ReadAll/10000_records-10        	      36	  31415667 ns/op
BenchmarkUnmarshal/easycsv.ReadAll/100000_records
BenchmarkUnmarshal/easycsv.ReadAll/100000_records-10       	       4	 319926677 ns/op
```

### Marshal

```
BenchmarkMarshal/csvlib.Marshal/100_records
BenchmarkMarshal/csvlib.Marshal/100_records-10         	   19753	     58890 ns/op
BenchmarkMarshal/csvlib.Marshal/1000_records
BenchmarkMarshal/csvlib.Marshal/1000_records-10        	    2149	    554537 ns/op
BenchmarkMarshal/csvlib.Marshal/10000_records
BenchmarkMarshal/csvlib.Marshal/10000_records-10       	     214	   5575920 ns/op
BenchmarkMarshal/csvlib.Marshal/100000_records
BenchmarkMarshal/csvlib.Marshal/100000_records-10      	      19	  55735281 ns/op

BenchmarkMarshal/csvutil.Marshal/100_records
BenchmarkMarshal/csvutil.Marshal/100_records-10        	   24388	     48931 ns/op
BenchmarkMarshal/csvutil.Marshal/1000_records
BenchmarkMarshal/csvutil.Marshal/1000_records-10       	    2557	    467704 ns/op
BenchmarkMarshal/csvutil.Marshal/10000_records
BenchmarkMarshal/csvutil.Marshal/10000_records-10      	     256	   4720885 ns/op
BenchmarkMarshal/csvutil.Marshal/100000_records
BenchmarkMarshal/csvutil.Marshal/100000_records-10     	      22	  48627754 ns/op

BenchmarkMarshal/gocsv.Marshal/100_records
BenchmarkMarshal/gocsv.Marshal/100_records-10          	   13254	     90873 ns/op
BenchmarkMarshal/gocsv.Marshal/1000_records
BenchmarkMarshal/gocsv.Marshal/1000_records-10         	    1294	    898938 ns/op
BenchmarkMarshal/gocsv.Marshal/10000_records
BenchmarkMarshal/gocsv.Marshal/10000_records-10        	     132	   9017481 ns/op
BenchmarkMarshal/gocsv.Marshal/100000_records
BenchmarkMarshal/gocsv.Marshal/100000_records-10       	      12	  90260420 ns/op
```

## Contributing

- You are welcome to make pull requests for new functions and bug fixes.

## Authors

- Dao Cong Tien ([tiendc](https://github.com/tiendc))

## License

- [MIT License](LICENSE)

[doc-img]: https://pkg.go.dev/badge/github.com/tiendc/go-csvlib
[doc]: https://pkg.go.dev/github.com/tiendc/go-csvlib
[gover-img]: https://img.shields.io/badge/Go-%3E%3D%201.18-blue
[gover]: https://img.shields.io/badge/Go-%3E%3D%201.18-blue
[ci-img]: https://github.com/tiendc/go-csvlib/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/tiendc/go-csvlib/actions/workflows/go.yml
[cov-img]: https://codecov.io/gh/tiendc/go-csvlib/branch/main/graph/badge.svg
[cov]: https://codecov.io/gh/tiendc/go-csvlib
[rpt-img]: https://goreportcard.com/badge/github.com/tiendc/go-csvlib
[rpt]: https://goreportcard.com/report/github.com/tiendc/go-csvlib
