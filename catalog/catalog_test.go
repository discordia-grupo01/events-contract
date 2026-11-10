package catalog_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/discordia-grupo01/events-contract/catalog"
)

func TestEveryRoutingKeyInTheRepoIsInTheCatalog(t *testing.T) {
	inCatalog := map[string]bool{}
	for _, e := range catalog.All() {
		if inCatalog[e.RoutingKey] {
			t.Errorf("%s repetido en el catálogo", e.RoutingKey)
		}
		inCatalog[e.RoutingKey] = true
	}
	files, err := filepath.Glob("../*/events.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no se encontraron events.go")
	}
	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			vs, ok := n.(*ast.ValueSpec)
			if !ok {
				return true
			}
			for i, name := range vs.Names {
				if !strings.HasPrefix(name.Name, "RoutingKey") {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok {
					continue
				}
				key, _ := strconv.Unquote(lit.Value)
				if !inCatalog[key] {
					t.Errorf("%s (%s) no está en catalog.All()", key, path)
				}
			}
			return true
		})
	}
}

func TestProducersAndConsumersAreSet(t *testing.T) {
	for _, e := range catalog.All() {
		if e.Producer == "" {
			t.Errorf("%s sin productor", e.RoutingKey)
		}
		for _, c := range e.Consumers {
			if c == e.Producer && !e.Internal {
				t.Errorf("%s: %s figura como productor y consumidor", e.RoutingKey, c)
			}
		}
	}
}
