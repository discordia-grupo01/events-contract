package validate_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/discordia-grupo01/events-contract/catalog"
	"github.com/discordia-grupo01/events-contract/schemagen"
	"github.com/discordia-grupo01/events-contract/schemas"
	"github.com/discordia-grupo01/events-contract/validate"
)

func TestEveryCatalogEventHasASchemaThatCompiles(t *testing.T) {
	keys, err := validate.RoutingKeys()
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, k := range keys {
		have[k] = true
	}
	for _, e := range catalog.All() {
		if !have[e.RoutingKey] {
			t.Errorf("%s no tiene schema", e.RoutingKey)
		}
		delete(have, e.RoutingKey)
	}
	for k := range have {
		t.Errorf("schema %s sin evento en el catálogo", k)
	}
}

func TestExamplesRoundTripThroughTheGoTypes(t *testing.T) {
	comments, err := schemagen.LoadComments("..")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range catalog.All() {
		example, err := schemagen.Example(e, comments)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(example)
		if err := validate.Validate(e.RoutingKey, raw); err != nil {
			t.Errorf("el ejemplo de %s no cumple su propio schema: %v", e.RoutingKey, err)
			continue
		}
		decoded := reflect.New(e.Type())
		if err := json.Unmarshal(raw, decoded.Interface()); err != nil {
			t.Errorf("%s: el ejemplo no entra en la struct: %v", e.RoutingKey, err)
			continue
		}
		again, _ := json.Marshal(decoded.Interface())
		if err := validate.Validate(e.RoutingKey, again); err != nil {
			t.Errorf("la struct de %s emite algo que su schema rechaza: %v", e.RoutingKey, err)
		}
	}
}

func TestMissingRequiredFieldIsRejected(t *testing.T) {
	comments, _ := schemagen.LoadComments("..")
	for _, e := range catalog.All() {
		example, err := schemagen.Example(e, comments)
		if err != nil {
			t.Fatal(err)
		}
		for _, field := range requiredFields(t, e.RoutingKey) {
			broken := map[string]any{}
			for k, v := range example {
				if k != field {
					broken[k] = v
				}
			}
			raw, _ := json.Marshal(broken)
			if err := validate.Validate(e.RoutingKey, raw); err == nil {
				t.Errorf("%s sin %s debería rechazarse", e.RoutingKey, field)
			}
		}
	}
}

func TestWrongTypesAndFormatsAreRejected(t *testing.T) {
	cases := map[string]string{
		"servers.member_left": `{"event_id":"no-es-uuid","occurred_at":"2026-11-10T15:00:00Z","server_id":"6d5c4b3a-2e1f-4a0b-9c8d-7e6f5a4b3c2d","user_id":"3f2e1d0c-9b8a-4766-b554-433221100ffe"}`,
	}
	for key, payload := range cases {
		err := validate.Validate(key, []byte(payload))
		var invalid *validate.InvalidError
		if !errors.As(err, &invalid) {
			t.Errorf("%s: se esperaba InvalidError, llegó %v", key, err)
		}
	}
	raw := `{"event_id":"0b6f1c8e-7a2d-4e3f-9c1b-5d4a3e2f1a0b","occurred_at":"2026-11-10T15:00:00Z","server_id":5,"user_id":"3f2e1d0c-9b8a-4766-b554-433221100ffe"}`
	if err := validate.Validate("servers.member_left", []byte(raw)); err == nil {
		t.Error("un server_id numérico debería rechazarse")
	}
}

func TestUnknownFieldsAreAcceptedSoAddingFieldsIsSafe(t *testing.T) {
	raw := `{"event_id":"0b6f1c8e-7a2d-4e3f-9c1b-5d4a3e2f1a0b","occurred_at":"2026-11-10T15:00:00Z","server_id":"6d5c4b3a-2e1f-4a0b-9c8d-7e6f5a4b3c2d","user_id":"3f2e1d0c-9b8a-4766-b554-433221100ffe","campo_nuevo":1}`
	if err := validate.Validate("servers.member_left", []byte(raw)); err != nil {
		t.Errorf("un campo extra no debería romper a nadie: %v", err)
	}
}

func TestUnknownEvent(t *testing.T) {
	if err := validate.Validate("nadie.nada", []byte(`{}`)); !errors.Is(err, validate.ErrUnknownEvent) {
		t.Errorf("se esperaba ErrUnknownEvent, llegó %v", err)
	}
}

func requiredFields(t *testing.T, key string) []string {
	raw, err := schemas.FS.ReadFile(key + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	return parsed.Required
}
