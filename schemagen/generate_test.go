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
