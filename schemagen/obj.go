package schemagen

import (
	"bytes"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type Obj struct {
	keys []string
	vals map[string]any
}

func NewObj() *Obj {
	return &Obj{vals: map[string]any{}}
}

func (o *Obj) Set(key string, value any) *Obj {
	if _, ok := o.vals[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.vals[key] = value
	return o
}

func (o *Obj) Get(key string) (any, bool) {
	v, ok := o.vals[key]
	return v, ok
}

func (o *Obj) Has(key string) bool {
	_, ok := o.vals[key]
	return ok
}

func (o *Obj) Keys() []string {
	return append([]string(nil), o.keys...)
}

func (o *Obj) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range o.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		vb, err := json.Marshal(o.vals[k])
		if err != nil {
			return nil, err
		}
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o *Obj) YAMLNode() *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, k := range o.keys {
		n.Content = append(n.Content, scalar(k), toNode(o.vals[k]))
	}
	return n
}

func scalar(s string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
	return n
}

func toNode(v any) *yaml.Node {
	switch t := v.(type) {
	case *Obj:
		return t.YAMLNode()
	case string:
		n := scalar(t)
		if containsNewline(t) {
			n.Style = yaml.LiteralStyle
		}
		return n
	case []string:
		n := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		if len(t) <= 8 {
			n.Style = yaml.FlowStyle
		}
		for _, s := range t {
			n.Content = append(n.Content, scalar(s))
		}
		return n
	case []any:
		n := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, e := range t {
			n.Content = append(n.Content, toNode(e))
		}
		return n
	case bool:
		n := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool"}
		if t {
			n.Value = "true"
		} else {
			n.Value = "false"
		}
		return n
	case int:
		n := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int"}
		b, _ := json.Marshal(t)
		n.Value = string(b)
		return n
	default:
		n := &yaml.Node{}
		_ = n.Encode(v)
		return n
	}
}

func containsNewline(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return true
		}
	}
	return false
}
