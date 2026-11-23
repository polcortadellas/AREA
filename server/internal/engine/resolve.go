package engine

import (
	"fmt"

	"github.com/polcortadellas/AREA/server/internal/node"
)

type step struct {
	outputs node.Values
	schema  node.Schema
}

// resolve replaces every {{ node.key }} in config with the output of an earlier
// step. done must also hold the trigger under "trigger". A missing optional key
// becomes "", a missing required key is an error.
func resolve(config node.Values, done map[string]step) (node.Values, error) {
	out := make(node.Values, len(config))
	for key, v := range config {
		s, ok := v.(string)
		if !ok {
			out[key] = v
			continue
		}
		// A value that is exactly one reference keeps its type, so a number stays a number.
		if m := node.RefRe.FindStringSubmatch(s); m != nil && m[0] == s {
			val, err := lookup(done, m[1], m[2])
			if err != nil {
				return nil, fmt.Errorf("%q: %w", key, err)
			}
			out[key] = val
			continue
		}
		var err error
		out[key] = node.RefRe.ReplaceAllStringFunc(s, func(ref string) string {
			m := node.RefRe.FindStringSubmatch(ref)
			val, e := lookup(done, m[1], m[2])
			if e != nil && err == nil {
				err = e
			}
			return fmt.Sprint(val)
		})
		if err != nil {
			return nil, fmt.Errorf("%q: %w", key, err)
		}
	}
	return out, nil
}

func lookup(done map[string]step, ref, key string) (any, error) {
	st, ok := done[ref]
	if !ok {
		return nil, fmt.Errorf("node %q has not run", ref)
	}
	if v, ok := st.outputs[key]; ok && v != nil {
		return v, nil
	}
	if f, ok := st.schema.Lookup(key); ok && f.Required {
		return nil, fmt.Errorf("node %q did not return required output %q", ref, key)
	}
	return "", nil
}
