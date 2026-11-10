package schemagen

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func Breaking(oldSchemas, newSchemas map[string][]byte) ([]string, error) {
	var problems []string
	names := make([]string, 0, len(oldSchemas))
	for name := range oldSchemas {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		key := strings.TrimSuffix(name, ".json")
		raw, ok := newSchemas[name]
		if !ok {
			problems = append(problems, fmt.Sprintf("%s: se quitó el evento", key))
			continue
		}
		var before, after map[string]any
		if err := json.Unmarshal(oldSchemas[name], &before); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if err := json.Unmarshal(raw, &after); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		for _, p := range compareObject(before, after, "") {
			problems = append(problems, key+": "+p)
		}
	}
	return problems, nil
}

func compareObject(before, after map[string]any, path string) []string {
	var out []string
	bprops, _ := before["properties"].(map[string]any)
	aprops, _ := after["properties"].(map[string]any)
	breq := stringSet(before["required"])
	areq := stringSet(after["required"])
	names := make([]string, 0, len(bprops))
	for n := range bprops {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		full := path + n
		ap, ok := aprops[n].(map[string]any)
		if !ok {
			out = append(out, fmt.Sprintf("se quitó el campo %s", full))
			continue
		}
		bp := bprops[n].(map[string]any)
		out = append(out, compareField(bp, ap, full)...)
		if areq[n] && !breq[n] {
			out = append(out, fmt.Sprintf("el campo %s pasó a ser obligatorio: los publicadores viejos no lo mandan", full))
		}
	}
	addedRequired := []string{}
	for n := range areq {
		if _, existed := bprops[n]; !existed {
			addedRequired = append(addedRequired, n)
		}
	}
	sort.Strings(addedRequired)
	for _, n := range addedRequired {
		out = append(out, fmt.Sprintf("el campo nuevo %s es obligatorio: los publicadores viejos no lo mandan", path+n))
	}
	return out
}

func compareField(before, after map[string]any, path string) []string {
	var out []string
	if bt, at := normalizeType(before["type"]), normalizeType(after["type"]); bt != at {
		out = append(out, fmt.Sprintf("el campo %s cambió de tipo: %s -> %s", path, bt, at))
		return out
	}
	if bf, af := before["format"], after["format"]; af != nil && !reflect.DeepEqual(bf, af) {
		out = append(out, fmt.Sprintf("el campo %s cambió de formato: %v -> %v", path, bf, af))
	}
	if be, ae := stringSet(before["enum"]), stringSet(after["enum"]); len(ae) > 0 {
		if len(be) == 0 {
			out = append(out, fmt.Sprintf("el campo %s ahora solo admite un conjunto de valores", path))
		} else {
			for v := range be {
				if !ae[v] {
					out = append(out, fmt.Sprintf("el campo %s ya no admite el valor %q", path, v))
				}
			}
		}
	}
	for _, k := range []string{"maxLength", "maximum", "maxItems"} {
		if bv, ok := before[k].(float64); ok {
			if av, ok2 := after[k].(float64); ok2 && av < bv {
				out = append(out, fmt.Sprintf("el campo %s bajó %s: %v -> %v", path, k, bv, av))
			}
		} else if av, ok := after[k]; ok {
			out = append(out, fmt.Sprintf("el campo %s ahora tiene %s=%v", path, k, av))
		}
	}
	for _, k := range []string{"minLength", "minimum", "minItems"} {
		if bv, ok := before[k].(float64); ok {
			if av, ok2 := after[k].(float64); ok2 && av > bv {
				out = append(out, fmt.Sprintf("el campo %s subió %s: %v -> %v", path, k, bv, av))
			}
		} else if av, ok := after[k]; ok {
			out = append(out, fmt.Sprintf("el campo %s ahora tiene %s=%v", path, k, av))
		}
	}
	if bp, ok := before["pattern"]; ok {
		if ap := after["pattern"]; !reflect.DeepEqual(bp, ap) {
			out = append(out, fmt.Sprintf("el campo %s cambió de patrón", path))
		}
	} else if ap, ok := after["pattern"]; ok {
		out = append(out, fmt.Sprintf("el campo %s ahora tiene el patrón %v", path, ap))
	}
	if _, isObject := before["properties"]; isObject {
		out = append(out, compareObject(before, after, path+".")...)
	}
	if bi, ok := before["items"].(map[string]any); ok {
		if ai, ok := after["items"].(map[string]any); ok {
			out = append(out, compareField(bi, ai, path+"[]")...)
			if _, isObject := bi["properties"]; isObject {
				out = append(out, compareObject(bi, ai, path+"[].")...)
			}
		}
	}
	return out
}

func stringSet(v any) map[string]bool {
	out := map[string]bool{}
	list, _ := v.([]any)
	for _, x := range list {
		out[fmt.Sprint(x)] = true
	}
	return out
}
