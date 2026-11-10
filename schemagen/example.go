package schemagen

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/discordia-grupo01/events-contract/catalog"
)

var topicPattern = regexp.MustCompile(`^\^\(?([a-z_|]+)\)?:\.\+$`)

var samplePatterns = map[string]string{
	"^[0-9a-f]{64}$":       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	"^[A-Za-z0-9_]{2,32}$": "custom_emoji",
}

func Example(e catalog.Entry, c Comments) (map[string]any, error) {
	s, err := EventSchema(e, c)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	v, err := sample(parsed, "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", e.RoutingKey, err)
	}
	return v.(map[string]any), nil
}

func sample(schema map[string]any, name string) (any, error) {
	typ := ""
	switch t := schema["type"].(type) {
	case string:
		typ = t
	case []any:
		for _, x := range t {
			if s, ok := x.(string); ok && s != "null" {
				typ = s
			}
		}
	}
	if en, ok := schema["enum"].([]any); ok && len(en) > 0 {
		return en[0], nil
	}
	switch typ {
	case "string":
		return sampleString(schema, name)
	case "boolean":
		return true, nil
	case "integer", "number":
		if m, ok := schema["minimum"].(float64); ok {
			return int(m) + 1, nil
		}
		return 1, nil
	case "array":
		items, _ := schema["items"].(map[string]any)
		n := 1
		if m, ok := schema["minItems"].(float64); ok && int(m) > n {
			n = int(m)
		}
		out := make([]any, 0, n)
		for i := 0; i < n; i++ {
			v, err := sample(items, name)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case "object":
		out := map[string]any{}
		props, _ := schema["properties"].(map[string]any)
		for k, v := range props {
			child, err := sample(v.(map[string]any), k)
			if err != nil {
				return nil, err
			}
			out[k] = child
		}
		if ap, ok := schema["additionalProperties"].(map[string]any); ok && len(props) == 0 {
			child, err := sample(ap, "valor")
			if err != nil {
				return nil, err
			}
			out["clave"] = child
		}
		return out, nil
	}
	if typ == "" && len(schema) == 0 {
		return "valor", nil
	}
	return nil, fmt.Errorf("tipo %q sin ejemplo", typ)
}

func sampleString(schema map[string]any, name string) (any, error) {
	if p, ok := schema["pattern"].(string); ok {
		v, found := samplePatterns[p]
		if m := topicPattern.FindStringSubmatch(p); m != nil {
			return strings.Split(m[1], "|")[0] + ":0b6f1c8e-7a2d-4e3f-9c1b-5d4a3e2f1a0b", nil
		}
		if !found {
			return nil, fmt.Errorf("falta un ejemplo para el patrón %s", p)
		}
		return v, nil
	}
	switch schema["format"] {
	case "uuid":
		return "0b6f1c8e-7a2d-4e3f-9c1b-5d4a3e2f1a0b", nil
	case "date-time":
		return "2026-11-10T15:00:00Z", nil
	case "email":
		return "ana@example.com", nil
	case "uri":
		return "https://cdn.example.com/imagen.png", nil
	case nil:
	default:
		return nil, fmt.Errorf("falta un ejemplo para el formato %v", schema["format"])
	}
	v := strings.ReplaceAll(name, "_", "-")
	if v == "" {
		v = "texto"
	}
	if m, ok := schema["maxLength"].(float64); ok && len(v) > int(m) {
		v = v[:int(m)]
	}
	if m, ok := schema["minLength"].(float64); ok {
		for len(v) < int(m) {
			v += "x"
		}
	}
	return v, nil
}
