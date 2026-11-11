package schemagen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	eventscontract "github.com/discordia-grupo01/events-contract"
	"github.com/discordia-grupo01/events-contract/catalog"
)

const SchemaIDBase = "https://github.com/discordia-grupo01/events-contract/schemas/"

var spaces = regexp.MustCompile(`\s+`)

func LoadComments(_ string) (Comments, error) {
	c := Comments{Types: map[string]string{}, Fields: map[string]map[string]string{}}
	files, err := fs.Glob(eventscontract.Source, "*/*.go")
	if err != nil {
		return c, err
	}
	for _, path := range files {
		raw, err := eventscontract.Source.ReadFile(path)
		if err != nil {
			return c, err
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, raw, parser.ParseComments)
		if err != nil {
			return c, err
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				doc := ts.Doc
				if doc == nil {
					doc = gen.Doc
				}
				if doc != nil {
					c.Types[file.Name.Name+"."+ts.Name.Name] = clean(doc.Text())
				}
				fields := map[string]string{}
				for _, f := range st.Fields.List {
					text := ""
					if f.Doc != nil {
						text = clean(f.Doc.Text())
					} else if f.Comment != nil {
						text = clean(f.Comment.Text())
					}
					if text == "" {
						continue
					}
					for _, n := range f.Names {
						fields[n.Name] = text
					}
				}
				if len(fields) > 0 {
					c.Fields[file.Name.Name+"."+ts.Name.Name] = fields
				}
			}
		}
	}
	return c, nil
}

func clean(s string) string {
	return strings.TrimSpace(spaces.ReplaceAllString(s, " "))
}

func FileName(key string) string {
	return key + ".json"
}

func EventSchema(e catalog.Entry, c Comments) (*Obj, error) {
	body, err := SchemaFor(e.Type(), c)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", e.RoutingKey, err)
	}
	out := NewObj().
		Set("$schema", Draft).
		Set("$id", SchemaIDBase+FileName(e.RoutingKey)).
		Set("title", e.RoutingKey)
	if d := c.Types[e.Type().String()]; d != "" {
		out.Set("description", d)
	}
	applyOverrides(body, e)
	for _, k := range body.Keys() {
		v, _ := body.Get(k)
		out.Set(k, v)
	}
	return out, nil
}

func Marshal(o *Obj) ([]byte, error) {
	raw, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func GenerateSchemas(root string) (map[string][]byte, error) {
	c, err := LoadComments(root)
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, e := range catalog.All() {
		s, err := EventSchema(e, c)
		if err != nil {
			return nil, err
		}
		b, err := Marshal(s)
		if err != nil {
			return nil, err
		}
		out[FileName(e.RoutingKey)] = b
	}
	return out, nil
}

type Drift struct {
	Missing []string
	Stale   []string
	Extra   []string
}

func (d Drift) Empty() bool {
	return len(d.Missing)+len(d.Stale)+len(d.Extra) == 0
}

func (d Drift) String() string {
	var b strings.Builder
	for _, m := range d.Missing {
		fmt.Fprintf(&b, "falta generar: %s\n", m)
	}
	for _, m := range d.Stale {
		fmt.Fprintf(&b, "desactualizado: %s\n", m)
	}
	for _, m := range d.Extra {
		fmt.Fprintf(&b, "sin evento en el catálogo: %s\n", m)
	}
	return b.String()
}

func Diff(dir string, want map[string][]byte, glob string) (Drift, error) {
	var d Drift
	for name, content := range want {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if os.IsNotExist(err) {
			d.Missing = append(d.Missing, name)
			continue
		}
		if err != nil {
			return d, err
		}
		if !bytes.Equal(got, content) {
			d.Stale = append(d.Stale, name)
		}
	}
	existing, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		return d, err
	}
	for _, p := range existing {
		name := filepath.Base(p)
		if _, ok := want[name]; !ok {
			d.Extra = append(d.Extra, name)
		}
	}
	sort.Strings(d.Missing)
	sort.Strings(d.Stale)
	sort.Strings(d.Extra)
	return d, nil
}

func WriteAll(dir string, files map[string][]byte, glob string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	existing, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		return err
	}
	for _, p := range existing {
		if _, ok := files[filepath.Base(p)]; !ok {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func applyOverrides(body *Obj, e catalog.Entry) {
	if len(e.Omit) == 0 && len(e.Require) == 0 {
		return
	}
	propsAny, _ := body.Get("properties")
	props := propsAny.(*Obj)
	kept := NewObj()
	for _, k := range props.Keys() {
		if contains(e.Omit, k) {
			continue
		}
		v, _ := props.Get(k)
		kept.Set(k, v)
	}
	body.Set("properties", kept)
	reqAny, _ := body.Get("required")
	req, _ := reqAny.([]string)
	var next []string
	for _, r := range req {
		if !contains(e.Omit, r) {
			next = append(next, r)
		}
	}
	for _, r := range e.Require {
		if !contains(next, r) {
			next = append(next, r)
		}
	}
	body.Set("required", next)
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func GenerateExamples(root string) (map[string][]byte, error) {
	c, err := LoadComments(root)
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, e := range catalog.All() {
		ex, err := Example(e, c)
		if err != nil {
			return nil, err
		}
		raw, err := json.MarshalIndent(ex, "", "  ")
		if err != nil {
			return nil, err
		}
		out[FileName(e.RoutingKey)] = append(raw, '\n')
	}
	return out, nil
}
