package validate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/discordia-grupo01/events-contract/schemas"
)

var ErrUnknownEvent = errors.New("evento sin schema")

type InvalidError struct {
	RoutingKey string
	Reasons    []string
}

func (e *InvalidError) Error() string {
	return fmt.Sprintf("%s no cumple el contrato: %s", e.RoutingKey, strings.Join(e.Reasons, "; "))
}

var (
	once     sync.Once
	compiled map[string]*jsonschema.Schema
	loadErr  error
)

func load() {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft7
	compiler.AssertFormat = true
	entries, err := schemas.FS.ReadDir(".")
	if err != nil {
		loadErr = err
		return
	}
	compiled = map[string]*jsonschema.Schema{}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		raw, err := schemas.FS.ReadFile(name)
		if err != nil {
			loadErr = err
			return
		}
		if err := compiler.AddResource(name, bytes.NewReader(raw)); err != nil {
			loadErr = fmt.Errorf("%s: %w", name, err)
			return
		}
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		s, err := compiler.Compile(name)
		if err != nil {
			loadErr = fmt.Errorf("%s: %w", name, err)
			return
		}
		compiled[strings.TrimSuffix(name, ".json")] = s
	}
}

func RoutingKeys() ([]string, error) {
	once.Do(load)
	if loadErr != nil {
		return nil, loadErr
	}
	keys := make([]string, 0, len(compiled))
	for k := range compiled {
		keys = append(keys, k)
	}
	return keys, nil
}

func Validate(routingKey string, payload []byte) error {
	once.Do(load)
	if loadErr != nil {
		return loadErr
	}
	s, ok := compiled[routingKey]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownEvent, routingKey)
	}
	var instance any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&instance); err != nil {
		return &InvalidError{RoutingKey: routingKey, Reasons: []string{"el cuerpo no es JSON válido"}}
	}
	if err := s.Validate(instance); err != nil {
		return &InvalidError{RoutingKey: routingKey, Reasons: reasons(err)}
	}
	return nil
}

func reasons(err error) []string {
	var verr *jsonschema.ValidationError
	if !errors.As(err, &verr) {
		return []string{err.Error()}
	}
	var out []string
	var walk func(v *jsonschema.ValidationError)
	walk = func(v *jsonschema.ValidationError) {
		if len(v.Causes) == 0 {
			loc := v.InstanceLocation
			if loc == "" {
				loc = "/"
			}
			out = append(out, fmt.Sprintf("%s: %s", loc, v.Message))
			return
		}
		for _, c := range v.Causes {
			walk(c)
		}
	}
	walk(verr)
	return out
}
