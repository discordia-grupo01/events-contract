package schemagen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/discordia-grupo01/events-contract/catalog"
	"github.com/discordia-grupo01/events-contract/validate"
)

var (
	schemasHeader = regexp.MustCompile(`^  schemas:\s*$`)
	indent2       = regexp.MustCompile(`^ {0,3}\S`)
)

func docsSchemaBlock(e catalog.Entry, c Comments) ([]string, error) {
	full, err := EventSchema(e, c)
	if err != nil {
		return nil, err
	}
	body := NewObj()
	for _, k := range full.Keys() {
		if k == "$schema" || k == "$id" || k == "title" || k == "description" {
			continue
		}
		v, _ := full.Get(k)
		body.Set(k, v)
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(body.YAMLNode()); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	var out []string
	for _, l := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		out = append(out, "      "+l)
	}
	return out, nil
}

func replaceSchema(lines []string, name string, block []string) ([]string, bool) {
	start := -1
	inSchemas := false
	for i, l := range lines {
		if schemasHeader.MatchString(l) {
			inSchemas = true
			continue
		}
		if inSchemas && l == "    "+name+":" {
			start = i
			break
		}
	}
	if start < 0 {
		return lines, false
	}
	end := start + 1
	for end < len(lines) {
		l := lines[end]
		if strings.TrimSpace(l) != "" && indent2.MatchString(l) || (strings.HasPrefix(l, "    ") && !strings.HasPrefix(l, "     ") && strings.TrimSpace(l) != "") {
			break
		}
		end++
	}
	trailing := end
	for trailing > start+1 && strings.TrimSpace(lines[trailing-1]) == "" {
		trailing--
	}
	out := append([]string{}, lines[:start+1]...)
	out = append(out, block...)
	out = append(out, lines[trailing:]...)
	return out, true
}

func SyncDocsSchemas(dir string, c Comments) (map[string][]byte, error) {
	events, err := LoadDocsEvents(dir)
	if err != nil {
		return nil, err
	}
	byFile := map[string][]catalog.Entry{}
	for _, e := range catalog.All() {
		if e.Internal {
			continue
		}
		if d, ok := events[e.RoutingKey]; ok && d.SchemaName != "" {
			byFile[d.File] = append(byFile[d.File], e)
		}
	}
	out := map[string][]byte{}
	files := make([]string, 0, len(byFile))
	for f := range byFile {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(raw), "\n")
		for _, e := range byFile[f] {
			block, err := docsSchemaBlock(e, c)
			if err != nil {
				return nil, err
			}
			var ok bool
			lines, ok = replaceSchema(lines, events[e.RoutingKey].SchemaName, block)
			if !ok {
				return nil, fmt.Errorf("%s: no se encontró components.schemas.%s", f, events[e.RoutingKey].SchemaName)
			}
		}
		out[f] = []byte(strings.Join(lines, "\n"))
	}
	return out, nil
}

func StaleDocsSchemas(dir string, c Comments) ([]string, error) {
	want, err := SyncDocsSchemas(dir, c)
	if err != nil {
		return nil, err
	}
	var out []string
	for f, content := range want {
		got, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(got, content) {
			out = append(out, f+": los schemas no coinciden con el contrato; correr `eventsgen docs-sync`")
		}
	}
	sort.Strings(out)
	return out, nil
}

func DocsExamples(dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	var problems []string
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
			if strings.HasPrefix(addr, "broadcast.") {
				continue
			}
			msgs, _ := chm["messages"].(map[string]any)
			for _, m := range msgs {
				msg := resolve(m, doc)
				examples, _ := msg["examples"].([]any)
				for i, ex := range examples {
					exm, _ := ex.(map[string]any)
					body, err := json.Marshal(exm["payload"])
					if err != nil {
						return nil, err
					}
					if err := validate.Validate(addr, body); err != nil {
						problems = append(problems, fmt.Sprintf("%s: el ejemplo %d de %s no cumple el contrato: %v", addr, i+1, filepath.Base(path), err))
					}
				}
			}
		}
	}
	sort.Strings(problems)
	return problems, nil
}
