package csvlib

import (
	"strings"

	"github.com/tiendc/gofn"
)

var (
	// ProcessorTrim trims space of string
	ProcessorTrim = strings.TrimSpace
	// ProcessorTrimPrefix trims prefix from a string
	ProcessorTrimPrefix = strings.TrimPrefix
	// ProcessorTrimSuffix trims suffix from a string
	ProcessorTrimSuffix = strings.TrimSuffix
	// ProcessorReplace replaces a substring in a string
	ProcessorReplace = strings.Replace
	// ProcessorReplaceAll replaces all occurrences of a substring in a string
	ProcessorReplaceAll = strings.ReplaceAll
	// ProcessorLower converts a string to lowercase
	ProcessorLower = strings.ToLower
	// ProcessorUpper converts a string to uppercase
	ProcessorUpper = strings.ToUpper
	// ProcessorNumberGroup formats a number with grouping its digits
	ProcessorNumberGroup = gofn.NumberFmtGroup
	// ProcessorNumberUngroup ungroups number digits
	ProcessorNumberUngroup = gofn.NumberFmtUngroup
)

// ProcessorNumberGroupComma formats a number with grouping its digits by comma
func ProcessorNumberGroupComma(s string) string {
	return gofn.NumberFmtGroup(s, '.', ',')
}

// ProcessorNumberUngroupComma ungroups number digits by comma
func ProcessorNumberUngroupComma(s string) string {
	return gofn.NumberFmtUngroup(s, ',')
}
