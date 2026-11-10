package schemagen_test

import (
	"strings"
	"testing"

	"github.com/discordia-grupo01/events-contract/schemagen"
)

const base = `{"type":"object","properties":{"a":{"type":"string","maxLength":10},"b":{"type":"string","enum":["x","y"]}},"required":["a"]}`

func compare(t *testing.T, next string) []string {
	t.Helper()
	got, err := schemagen.Breaking(
		map[string][]byte{"k.e.json": []byte(base)},
		map[string][]byte{"k.e.json": []byte(next)},
	)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestAddingAnOptionalFieldIsSafe(t *testing.T) {
	next := `{"type":"object","properties":{"a":{"type":"string","maxLength":10},"b":{"type":"string","enum":["x","y","z"]},"c":{"type":"integer"}},"required":["a"]}`
	if got := compare(t, next); len(got) != 0 {
		t.Fatalf("agregar un campo opcional y un valor de enum no debería romper: %v", got)
	}
}

func TestBreakingChangesAreReported(t *testing.T) {
	cases := map[string]string{
		"se quitó el campo b":             `{"type":"object","properties":{"a":{"type":"string","maxLength":10}},"required":["a"]}`,
		"cambió de tipo":                  `{"type":"object","properties":{"a":{"type":"integer"},"b":{"type":"string","enum":["x","y"]}},"required":["a"]}`,
		"pasó a ser obligatorio":          `{"type":"object","properties":{"a":{"type":"string","maxLength":10},"b":{"type":"string","enum":["x","y"]}},"required":["a","b"]}`,
		"el campo nuevo c es obligatorio": `{"type":"object","properties":{"a":{"type":"string","maxLength":10},"b":{"type":"string","enum":["x","y"]},"c":{"type":"string"}},"required":["a","c"]}`,
		"ya no admite el valor":           `{"type":"object","properties":{"a":{"type":"string","maxLength":10},"b":{"type":"string","enum":["x"]}},"required":["a"]}`,
		"bajó maxLength":                  `{"type":"object","properties":{"a":{"type":"string","maxLength":5},"b":{"type":"string","enum":["x","y"]}},"required":["a"]}`,
	}
	for want, next := range cases {
		got := compare(t, next)
		if len(got) == 0 || !strings.Contains(strings.Join(got, "\n"), want) {
			t.Errorf("se esperaba %q, llegó %v", want, got)
		}
	}
}

func TestRemovingAnEventIsBreaking(t *testing.T) {
	got, err := schemagen.Breaking(map[string][]byte{"k.e.json": []byte(base)}, map[string][]byte{})
	if err != nil || len(got) != 1 || !strings.Contains(got[0], "se quitó el evento") {
		t.Fatalf("llegó %v, %v", got, err)
	}
}
