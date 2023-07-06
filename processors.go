package csvlib

import (
	"strings"

	"github.com/tiendc/gofn"
)

var (
	ProcessorTrim          = strings.TrimSpace
	ProcessorTrimPrefix    = strings.TrimPrefix
	ProcessorTrimSuffix    = strings.TrimSuffix
	ProcessorReplace       = strings.Replace
	ProcessorReplaceAll    = strings.ReplaceAll
	ProcessorLower         = strings.ToLower
	ProcessorUpper         = strings.ToUpper
	ProcessorNumberGroup   = gofn.NumberFmtGroup
	ProcessorNumberUngroup = gofn.NumberFmtUngroup
)

func ProcessorNumberGroupComma(s string) string {
	return gofn.NumberFmtGroup(s, '.', ',')
}

func ProcessorNumberUngroupComma(s string) string {
	return gofn.NumberFmtUngroup(s, ',')
}
