package idem

import (
	"bytes"
	"encoding/json"
)

func CanonicalizeJSON(reqBytes []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(reqBytes))
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, ErrInvalidJSON
	}

	if dec.More() {
		return nil, ErrInvalidJSON
	}

	n := normalizeJSON(v)
	return json.Marshal(n)
}

func normalizeJSON(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			out[k] = normalizeJSON(vv)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i := range x {
			out[i] = normalizeJSON(x[i])
		}
		return out
	default:
		return x
	}
}
