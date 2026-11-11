package schemagen

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const Draft = "http://json-schema.org/draft-07/schema#"

var timeType = reflect.TypeOf(time.Time{})

type Comments struct {
	Types  map[string]string
	Fields map[string]map[string]string
}

func SchemaFor(t reflect.Type, comments Comments) (*Obj, error) {
	return objectSchema(t, comments)
}

func objectSchema(t reflect.Type, comments Comments) (*Obj, error) {
	props := NewObj()
	var required []string
	if err := collect(t, comments, props, &required); err != nil {
		return nil, err
	}
	out := NewObj().Set("type", "object").Set("properties", props)
	if len(required) > 0 {
		out.Set("required", required)
	}
	return out, nil
}

func collect(t reflect.Type, comments Comments, props *Obj, required *[]string) error {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous && f.Tag.Get("json") == "" {
			ft := f.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() != reflect.Struct {
				return fmt.Errorf("campo embebido %s.%s no es struct", t.Name(), f.Name)
			}
			if err := collect(ft, comments, props, required); err != nil {
				return err
			}
			continue
		}
		if !f.IsExported() {
			continue
		}
		name, omit, skip := jsonName(f)
		if skip {
			continue
		}
		contract := f.Tag.Get("contract")
		optional := false
		if contract == "optional=true" || strings.HasSuffix(contract, ";optional=true") || strings.Contains(contract, "optional=true;") {
			optional = true
			contract = strings.NewReplacer(";optional=true", "", "optional=true;", "", "optional=true", "").Replace(contract)
		}
		s, err := fieldSchema(f.Type, omit || optional, contract, comments)
		if err != nil {
			return fmt.Errorf("%s.%s: %w", t.Name(), f.Name, err)
		}
		if desc := comments.Fields[t.String()][f.Name]; desc != "" {
			s.Set("description", desc)
		}
		props.Set(name, s)
		if !omit && !optional {
			*required = append(*required, name)
		}
	}
	return nil
}

func jsonName(f reflect.StructField) (name string, omit bool, skip bool) {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return "", false, true
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "" {
		name = f.Name
	}
	for _, p := range parts[1:] {
		if p == "omitempty" || p == "omitzero" {
			omit = true
		}
	}
	return name, omit, false
}

func fieldSchema(t reflect.Type, omit bool, contract string, comments Comments) (*Obj, error) {
	nullable := false
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		nullable = !omit
	}
	s, err := typeSchema(t, comments)
	if err != nil {
		return nil, err
	}
	if err := applyContract(s, t, contract); err != nil {
		return nil, err
	}
	if nullable {
		typ, _ := s.Get("type")
		s.Set("type", []string{typ.(string), "null"})
	}
	return s, nil
}

func typeSchema(t reflect.Type, comments Comments) (*Obj, error) {
	if t == timeType {
		return NewObj().Set("type", "string").Set("format", "date-time"), nil
	}
	switch t.Kind() {
	case reflect.String:
		return NewObj().Set("type", "string"), nil
	case reflect.Bool:
		return NewObj().Set("type", "boolean"), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewObj().Set("type", "integer"), nil
	case reflect.Float32, reflect.Float64:
		return NewObj().Set("type", "number"), nil
	case reflect.Slice, reflect.Array:
		items, err := fieldSchema(t.Elem(), false, "", comments)
		if err != nil {
			return nil, err
		}
		return NewObj().Set("type", "array").Set("items", items), nil
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("mapa con clave %s no soportado", t.Key())
		}
		values, err := fieldSchema(t.Elem(), false, "", comments)
		if err != nil {
			return nil, err
		}
		return NewObj().Set("type", "object").Set("additionalProperties", values), nil
	case reflect.Struct:
		return objectSchema(t, comments)
	case reflect.Interface:
		return NewObj(), nil
	}
	return nil, fmt.Errorf("tipo %s no soportado", t)
}

func applyContract(s *Obj, t reflect.Type, contract string) error {
	if contract == "" {
		return nil
	}
	for _, part := range strings.Split(contract, ";") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return fmt.Errorf("etiqueta contract %q sin valor", part)
		}
		switch key {
		case "format", "pattern":
			s.Set(key, value)
		case "minLength", "maxLength", "minimum", "maximum", "minItems", "maxItems":
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("etiqueta contract %s: %w", key, err)
			}
			s.Set(key, n)
		case "enum":
			vals := strings.Split(value, "|")
			items := make([]any, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			s.Set("enum", items)
		case "itemFormat":
			items, ok := s.Get("items")
			if !ok {
				return fmt.Errorf("itemFormat sobre un campo que no es lista")
			}
			items.(*Obj).Set("format", value)
		default:
			return fmt.Errorf("etiqueta contract desconocida %q", key)
		}
	}
	return nil
}
