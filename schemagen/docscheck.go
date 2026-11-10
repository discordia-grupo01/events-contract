package schemagen

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/discordia-grupo01/events-contract/catalog"
)

var constraintKeys = []string{"format", "pattern", "minLength", "maxLength", "minimum", "maximum", "minItems", "maxItems", "enum"}

type DocsEvent struct {
	File       string
	SchemaName string
	Required   []string
	Props      map[string]map[string]any
}

func LoadDocsEvents(dir string) (map[string]DocsEvent, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	out := map[string]DocsEvent{}
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var doc map[string]any
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		channels, _ := doc["channels"].(map[string]any)
		for _, ch := range channels {
			chm, _ := ch.(map[string]any)
			addr, _ := chm["address"].(string)
			if addr == "" {
				continue
			}
			payload := channelPayload(chm, doc)
			if payload == nil {
				continue
			}
			schemaName := channelSchemaName(chm, doc)
			props := map[string]map[string]any{}
			var required []string
			flatten(payload, doc, props, &required)
			sort.Strings(required)
			out[addr] = DocsEvent{File: filepath.Base(path), SchemaName: schemaName, Required: required, Props: props}
		}
	}
	return out, nil
}

func resolve(node any, doc map[string]any) map[string]any {
	m, _ := node.(map[string]any)
	for i := 0; i < 8 && m != nil; i++ {
		ref, ok := m["$ref"].(string)
		if !ok {
			return m
		}
		parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
		var cur any = doc
		for _, p := range parts {
			cm, _ := cur.(map[string]any)
			cur = cm[p]
		}
		m, _ = cur.(map[string]any)
	}
	return m
}

func channelPayload(ch map[string]any, doc map[string]any) map[string]any {
	msgs, _ := ch["messages"].(map[string]any)
	for _, m := range msgs {
		msg := resolve(m, doc)
		if msg == nil {
			continue
		}
		if p, ok := msg["payload"]; ok {
			return resolve(p, doc)
		}
	}
	return nil
}

func flatten(schema map[string]any, doc map[string]any, props map[string]map[string]any, required *[]string) {
	if schema == nil {
		return
	}
	if all, ok := schema["allOf"].([]any); ok {
		for _, part := range all {
			flatten(resolve(part, doc), doc, props, required)
		}
	}
	if p, ok := schema["properties"].(map[string]any); ok {
		for name, v := range p {
			pm := resolve(v, doc)
			props[name] = pm
		}
	}
	if r, ok := schema["required"].([]any); ok {
		for _, v := range r {
			if s, ok := v.(string); ok {
				*required = append(*required, s)
			}
		}
	}
}

func generatedBody(e catalog.Entry, c Comments) (map[string]map[string]any, []string, error) {
	s, err := EventSchema(e, c)
	if err != nil {
		return nil, nil, err
	}
	raw, err := Marshal(s)
	if err != nil {
		return nil, nil, err
	}
	var parsed map[string]any
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return nil, nil, err
	}
	props := map[string]map[string]any{}
	for k, v := range parsed["properties"].(map[string]any) {
		props[k] = v.(map[string]any)
	}
	var required []string
	if r, ok := parsed["required"].([]any); ok {
		for _, v := range r {
			required = append(required, v.(string))
		}
	}
	sort.Strings(required)
	return props, required, nil
}

func normalizeType(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		var parts []string
		for _, x := range t {
			if s, ok := x.(string); ok && s != "null" {
				parts = append(parts, s)
			}
		}
		sort.Strings(parts)
		return strings.Join(parts, "|")
	}
	return ""
}

func CompareDocs(dir string, c Comments) ([]string, error) {
	docs, err := LoadDocsEvents(dir)
	if err != nil {
		return nil, err
	}
	var problems []string
	seen := map[string]bool{}
	for _, e := range catalog.All() {
		if e.Internal {
			continue
		}
		seen[e.RoutingKey] = true
		d, ok := docs[e.RoutingKey]
		if !ok {
			problems = append(problems, fmt.Sprintf("%s: está en el contrato pero no tiene canal en el AsyncAPI de docs", e.RoutingKey))
			continue
		}
		props, required, err := generatedBody(e, c)
		if err != nil {
			return nil, err
		}
		for name, gp := range props {
			dp, ok := d.Props[name]
			if !ok {
				problems = append(problems, fmt.Sprintf("%s: el campo %q está en el contrato pero no en %s", e.RoutingKey, name, d.File))
				continue
			}
			if dt, gt := normalizeType(dp["type"]), normalizeType(gp["type"]); dt != gt {
				problems = append(problems, fmt.Sprintf("%s.%s: tipo %q en %s y %q en el contrato", e.RoutingKey, name, dt, d.File, gt))
			}
			for _, k := range constraintKeys {
				if dv, has := dp[k]; has {
					if gv, ok := gp[k]; !ok || !reflect.DeepEqual(normalizeNumber(dv), normalizeNumber(gv)) {
						problems = append(problems, fmt.Sprintf("%s.%s: %s=%v en %s y %v en el contrato", e.RoutingKey, name, k, dv, d.File, gp[k]))
					}
				} else if gv, ok := gp[k]; ok {
					problems = append(problems, fmt.Sprintf("%s.%s: %s=%v en el contrato y nada en %s", e.RoutingKey, name, k, gv, d.File))
				}
			}
		}
		for name := range d.Props {
			if _, ok := props[name]; !ok {
				problems = append(problems, fmt.Sprintf("%s: el campo %q está en %s pero no en el contrato", e.RoutingKey, name, d.File))
			}
		}
		if !reflect.DeepEqual(required, d.Required) && !(len(required) == 0 && len(d.Required) == 0) {
			problems = append(problems, fmt.Sprintf("%s: obligatorios %v en %s y %v en el contrato", e.RoutingKey, d.Required, d.File, required))
		}
	}
	var extra []string
	for addr := range docs {
		if !seen[addr] && !strings.HasPrefix(addr, "broadcast.") {
			extra = append(extra, addr)
		}
	}
	sort.Strings(extra)
	for _, addr := range extra {
		problems = append(problems, fmt.Sprintf("%s: tiene canal en %s pero no está en el contrato", addr, docs[addr].File))
	}
	sort.Strings(problems)
	return problems, nil
}

func normalizeNumber(v any) any {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return v
}

func channelSchemaName(ch map[string]any, doc map[string]any) string {
	msgs, _ := ch["messages"].(map[string]any)
	for _, m := range msgs {
		msg := resolve(m, doc)
		payload, _ := msg["payload"].(map[string]any)
		ref, _ := payload["$ref"].(string)
		const prefix = "#/components/schemas/"
		if strings.HasPrefix(ref, prefix) {
			return strings.TrimPrefix(ref, prefix)
		}
	}
	return ""
}
