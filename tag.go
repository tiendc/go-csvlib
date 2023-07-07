package csvlib

import (
	"fmt"
	"reflect"
	"strings"
)

type tagDetail struct {
	name      string
	prefix    string
	ignored   bool
	empty     bool
	omitEmpty bool
	optional  bool
	inline    bool
}

func parseTag(tagName string, field reflect.StructField) (*tagDetail, error) {
	tagValue, ok := field.Tag.Lookup(tagName)
	if !ok {
		return nil, nil
	}

	tag := &tagDetail{}
	tags := strings.Split(tagValue, ",")
	if len(tags) == 1 && tags[0] == "" {
		tag.name = field.Name
		tag.empty = true
	} else {
		switch tags[0] {
		case "-":
			tag.ignored = true
		case "":
			tag.name = field.Name
		default:
			tag.name = tags[0]
		}

		for _, tagOpt := range tags[1:] {
			switch {
			case tagOpt == "optional":
				tag.optional = true
			case tagOpt == "omitempty":
				tag.omitEmpty = true
			case tagOpt == "inline":
				tag.inline = true
			case strings.HasPrefix(tagOpt, "prefix="):
				tag.prefix = tagOpt[len("prefix="):]
			}
		}
	}

	// Validation: struct field unexported
	if !tag.ignored && !field.IsExported() {
		return nil, fmt.Errorf("%w: struct field %s unexported", ErrTagOptionInvalid, field.Name)
	}
	// Validation: only inline column can have prefix
	if tag.prefix != "" && !tag.inline {
		return nil, fmt.Errorf("%w: prefix tag is only accepted for inline column", ErrTagOptionInvalid)
	}
	// Validation: inline column must not be optional
	if tag.inline && tag.optional {
		return nil, fmt.Errorf("%w: inline column must not be optional", ErrTagOptionInvalid)
	}

	return tag, nil
}
