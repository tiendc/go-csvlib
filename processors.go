package csvlib

import (
	"strings"

	"github.com/tiendc/gofn"
)

func ProcessorTrim(s string) string {
	return strings.TrimSpace(s)
}

func ProcessorTrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

func ProcessorTrimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

func ProcessorReplace(s string, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

func ProcessorReplaceAll(s string, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func ProcessorLower(s string) string {
	return strings.ToLower(s)
}

func ProcessorUpper(s string) string {
	return strings.ToUpper(s)
}

func ProcessorNumberGroup(s string, fractionSep, groupSep byte) string {
	return gofn.NumberFmtGroup(s, fractionSep, groupSep)
}

func ProcessorNumberUngroup(s string, groupSep byte) string {
	return gofn.NumberFmtUngroup(s, groupSep)
}

func ProcessorNumberGroupComma(s string) string {
	return gofn.NumberFmtGroup(s, '.', ',')
}

func ProcessorNumberUngroupComma(s string) string {
	return gofn.NumberFmtUngroup(s, ',')
}
