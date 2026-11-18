package node

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

type FieldType string

const (
	TypeString  FieldType = "string"
	TypeNumber  FieldType = "number"
	TypeBoolean FieldType = "boolean"
)

type Field struct {
	Key         string    `json:"key"`
	Type        FieldType `json:"type"`
	Required    bool      `json:"required"`
	Description string    `json:"description"`
}

// Schema is ordered: clients render the form fields in this order.
type Schema []Field

// Names and keys end up in "service.name" IDs and {{ node.key }} references,
// so no dots or spaces.
var identRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func (s Schema) Lookup(key string) (Field, bool) {
	for _, f := range s {
		if f.Key == key {
			return f, true
		}
	}
	return Field{}, false
}

// Validate accepts a {{ node.key }} reference for any type: the workflow
// validator checks it and the engine resolves it before Execute.
func (s Schema) Validate(values Values) error {
	var problems []string
	for _, f := range s {
		v := values[f.Key]
		switch {
		case v == nil || v == "":
			if f.Required {
				problems = append(problems, fmt.Sprintf("%q is required", f.Key))
			}
		case isTemplate(v):
		case !f.Type.matches(v):
			problems = append(problems, fmt.Sprintf("%q must be a %s", f.Key, f.Type))
		}
	}
	for _, key := range slices.Sorted(maps.Keys(values)) {
		if _, ok := s.Lookup(key); !ok {
			problems = append(problems, fmt.Sprintf("%q is not a parameter of this node", key))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrInvalidConfig, strings.Join(problems, "; "))
}

func (s Schema) check() error {
	seen := make(map[string]bool, len(s))
	for _, f := range s {
		switch {
		case !identRe.MatchString(f.Key):
			return fmt.Errorf("invalid field key %q", f.Key)
		case seen[f.Key]:
			return fmt.Errorf("duplicate field key %q", f.Key)
		case !f.Type.valid():
			return fmt.Errorf("field %q has unknown type %q", f.Key, f.Type)
		}
		seen[f.Key] = true
	}
	return nil
}

func (t FieldType) valid() bool {
	switch t {
	case TypeString, TypeNumber, TypeBoolean:
		return true
	default:
		return false
	}
}

func (t FieldType) matches(v any) bool {
	switch t {
	case TypeString:
		_, ok := v.(string)
		return ok
	case TypeNumber:
		switch v.(type) {
		case float64, float32, int, int32, int64:
			return true
		default:
			return false
		}
	case TypeBoolean:
		_, ok := v.(bool)
		return ok
	default:
		return false
	}
}

func isTemplate(v any) bool {
	s, ok := v.(string)
	return ok && strings.Contains(s, "{{")
}
