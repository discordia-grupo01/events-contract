package schemagen_test

import (
	"testing"

	"github.com/discordia-grupo01/events-contract/catalog"
	"github.com/discordia-grupo01/events-contract/schemagen"
)

func TestGeneratedFilesAreUpToDate(t *testing.T) {
	schemas, err := schemagen.GenerateSchemas("..")
	if err != nil {
		t.Fatal(err)
	}
	d, err := schemagen.Diff("../schemas", schemas, "*.json")
	if err != nil {
		t.Fatal(err)
	}
	if !d.Empty() {
		t.Fatalf("schemas desactualizados, correr `go run ./cmd/eventsgen gen`:\n%s", d)
	}
	comments, err := schemagen.LoadComments("..")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := schemagen.AsyncAPI(catalog.All(), comments)
	if err != nil {
		t.Fatal(err)
	}
	d, err = schemagen.Diff("../asyncapi", map[string][]byte{"events.asyncapi.yaml": doc}, "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !d.Empty() {
		t.Fatalf("AsyncAPI desactualizado, correr `go run ./cmd/eventsgen gen`:\n%s", d)
	}
}

func TestEventsWithTheSameTypeNameKeepTheirOwnDescription(t *testing.T) {
	comments, err := schemagen.LoadComments("..")
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := map[string]string{
		"billing.account_changed":  "Changed is published by billing",
		"admin.backoffice_changed": "Changed is published by admin",
	}
	for _, e := range catalog.All() {
		prefix, ok := wantPrefix[e.RoutingKey]
		if !ok {
			continue
		}
		s, err := schemagen.EventSchema(e, comments)
		if err != nil {
			t.Fatal(err)
		}
		d, _ := s.Get("description")
		got, _ := d.(string)
		if len(got) < len(prefix) || got[:len(prefix)] != prefix {
			t.Fatalf("%s: la descripción es de otro tipo: %q", e.RoutingKey, got)
		}
	}
}
