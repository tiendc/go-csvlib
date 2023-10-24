package csvlib

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/hashicorp/go-multierror"
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

func processTemplate(templ string, params ParameterMap) (detail string, retErr error) {
	detail = templ
	t, err := template.New("error").Parse(detail)
	if err != nil {
		retErr = multierror.Append(retErr, err)
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, 100)) // nolint: gomnd
	err = t.Execute(buf, params)
	if err != nil {
		retErr = multierror.Append(retErr, err)
	} else {
		detail = buf.String()
	}
	return
}
