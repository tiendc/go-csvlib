[![Go Version][gover-img]][gover] [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoReport][rpt-img]][rpt]

# High level CSV lib for Go 1.18+

This is a library for decoding and encoding CSV at high level as it provides convenient methods and many configuration options to process the data the way you expect.

## Functionalities

**Decoding**
  - Decode CSV data into Go struct
  - Support Go interface `encoding.TextUnmarshaler` (with function `UnmarshalText`)
  - Support custom interface `CSVUnmarshaler` (with function `UnmarshalCSV`)
  - Able to continue decoding when error occurs (collect all errors at once)
  - Able to perform custom validator functions on cell data after decoding
  - Able to decode dynamic columns into Go struct field (inline columns)
  - Support rendering the result errors into human-readable content (row-by-row text and CSV)
  - Support localization to render the result errors into a specific language

**Encoding**
  - Encode Go struct into CSV data
  - Support Go interface `encoding.TextMarshaler` (with function `MarshalText`)
  - Support custom interface `CSVMarshaler` (with function `MarshalCSV`)
  - Able to encode dynamic columns defined via inner Go struct
  - Able to localize the header into a specific language

## Installation

```shell
go get github.com/tiendc/go-csvlib
```

## Usage

- [Decoding](docs/DECODING.md)
- [Encoding](docs/ENCODING.md)

## Benchmarks

TBD

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
[cov-img]: https://codecov.io/gh/tiendc/go-csvlib/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/tiendc/go-csvlib
[rpt-img]: https://goreportcard.com/badge/github.com/tiendc/go-csvlib
[rpt]: https://goreportcard.com/report/github.com/tiendc/go-csvlib
