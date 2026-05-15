package output

import "strings"

// selectFields applies dot-path projection to a JSON-decoded value. Each path
// like "id" or "to.name" extracts that field; unknown paths are skipped.
//
// On a slice, projection is applied to each element. On a map, projection
// keeps only the named keys (preserving the requested order in input order).
// Non-container values are returned unchanged.
func selectFields(v any, paths []string) any {
	if len(paths) == 0 {
		return v
	}
	switch x := v.(type) {
	case []any:
		out := make([]any, 0, len(x))
		for _, item := range x {
			out = append(out, selectFields(item, paths))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(paths))
		for _, p := range paths {
			head, rest, more := splitPath(p)
			val, ok := x[head]
			if !ok {
				continue
			}
			if !more {
				out[head] = val
				continue
			}
			sub := selectFields(val, []string{rest})
			out[head] = sub
		}
		return out
	default:
		return v
	}
}

func splitPath(p string) (head, rest string, more bool) {
	i := strings.IndexByte(p, '.')
	if i < 0 {
		return p, "", false
	}
	return p[:i], p[i+1:], true
}
