package csvlib

import (
	"fmt"
	"strings"
)

func validateHeader(header []string) error {
	mapCheckUniq := make(map[string]struct{}, len(header))
	for _, h := range header {
		hh := strings.TrimSpace(h)
		if h != hh || len(hh) == 0 {
			return fmt.Errorf("%w: \"%s\" invalid", ErrHeaderColumnInvalid, h)
		}
		if _, ok := mapCheckUniq[hh]; ok {
			return fmt.Errorf("%w: \"%s\" duplicated", ErrHeaderColumnDuplicated, h)
		}
		mapCheckUniq[hh] = struct{}{}
	}
	return nil
}

func processParams(s string, params ParameterMap) string {
	if len(params) == 0 {
		return s
	}
	for k, v := range params {
		key := "{{." + k + "}}"
		val := fmt.Sprintf("%v", v)
		s = strings.ReplaceAll(s, key, val)
	}
	return s
}
