package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/discordia-grupo01/events-contract/catalog"
	"github.com/discordia-grupo01/events-contract/schemagen"
)

const (
	schemasDir  = "schemas"
	asyncapiDir = "asyncapi"
	asyncapiOut = "events.asyncapi.yaml"
	examplesDir = "examples"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: eventsgen gen | check | docs-check <carpeta asyncapi de docs>")
		os.Exit(2)
	}
	comments, err := schemagen.LoadComments(".")
	fail(err)
	switch os.Args[1] {
	case "gen":
		schemas, doc := build(comments)
		fail(schemagen.WriteAll(schemasDir, schemas, "*.json"))
		fail(schemagen.WriteAll(asyncapiDir, doc, "*.yaml"))
		examples, err := schemagen.GenerateExamples(".")
		fail(err)
		fail(schemagen.WriteAll(examplesDir, examples, "*.json"))
		fmt.Printf("%d schemas en %s/ y %s/%s\n", len(schemas), schemasDir, asyncapiDir, asyncapiOut)
	case "check":
		schemas, doc := build(comments)
		d1, err := schemagen.Diff(schemasDir, schemas, "*.json")
		fail(err)
		d2, err := schemagen.Diff(asyncapiDir, doc, "*.yaml")
		fail(err)
		examples, err := schemagen.GenerateExamples(".")
		fail(err)
		d3, err := schemagen.Diff(examplesDir, examples, "*.json")
		fail(err)
		out := d1.String() + d2.String() + d3.String()
		if out != "" {
			fmt.Fprint(os.Stderr, out)
			fmt.Fprintln(os.Stderr, "ejecutá `go run ./cmd/eventsgen gen` y commiteá el resultado")
			os.Exit(1)
		}
	case "docs-check":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "uso: eventsgen docs-check <carpeta asyncapi de docs>")
			os.Exit(2)
		}
		problems, err := schemagen.CompareDocs(os.Args[2], comments)
		fail(err)
		stale, err := internalDocStale(os.Args[2], comments)
		fail(err)
		problems = append(problems, stale...)
		staleSchemas, err := schemagen.StaleDocsSchemas(os.Args[2], comments)
		fail(err)
		problems = append(problems, staleSchemas...)
		badExamples, err := schemagen.DocsExamples(os.Args[2])
		fail(err)
		problems = append(problems, badExamples...)
		for _, p := range problems {
			fmt.Println(p)
		}
		if len(problems) > 0 {
			fmt.Fprintf(os.Stderr, "%d diferencias entre el contrato y el AsyncAPI\n", len(problems))
			os.Exit(1)
		}
	case "compat":
		compat()
	case "docs-sync":
		docsSync(comments)
	default:
		fmt.Fprintln(os.Stderr, "comando desconocido:", os.Args[1])
		os.Exit(2)
	}
}

func build(comments schemagen.Comments) (map[string][]byte, map[string][]byte) {
	schemas, err := schemagen.GenerateSchemas(".")
	fail(err)
	doc, err := schemagen.AsyncAPI(catalog.All(), comments)
	fail(err)
	return schemas, map[string][]byte{asyncapiOut: doc}
}

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const allowFile = "compat.allow"

func compat() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "uso: eventsgen compat <ref de git, por ejemplo origin/main>")
		os.Exit(2)
	}
	ref := os.Args[2]
	old := gitSchemas(ref)
	if len(old) == 0 {
		fmt.Println("la referencia no tiene schemas todavía: nada que comparar")
		return
	}
	current, err := schemagen.GenerateSchemas(".")
	fail(err)
	problems, err := schemagen.Breaking(old, current)
	fail(err)
	allowed := allowedKeys()
	var blocking []string
	for _, p := range problems {
		key, _, _ := strings.Cut(p, ":")
		if allowed[key] {
			fmt.Println("permitido por", allowFile+":", p)
			continue
		}
		blocking = append(blocking, p)
	}
	for _, p := range blocking {
		fmt.Println(p)
	}
	if len(blocking) > 0 {
		fmt.Fprintf(os.Stderr, "%d cambios que rompen a los consumidores o publicadores existentes\n", len(blocking))
		fmt.Fprintf(os.Stderr, "si es intencional, agregá el routing key a %s (una línea por evento) y explicalo en el PR\n", allowFile)
		os.Exit(1)
	}
}

func allowedKeys() map[string]bool {
	out := map[string]bool{}
	f, err := os.Open(allowFile)
	if err != nil {
		return out
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			out[line] = true
		}
	}
	return out
}

func gitSchemas(ref string) map[string][]byte {
	list, err := exec.Command("git", "ls-tree", "--name-only", ref, schemasDir+"/").Output()
	if err != nil {
		return nil
	}
	out := map[string][]byte{}
	for _, path := range strings.Fields(string(list)) {
		if !strings.HasSuffix(path, ".json") {
			continue
		}
		raw, err := exec.Command("git", "show", ref+":"+path).Output()
		fail(err)
		out[filepath.Base(path)] = raw
	}
	return out
}

const internalDocName = "messaging-broadcast.yaml"

func internalDoc(comments schemagen.Comments) []byte {
	var internal []catalog.Entry
	for _, e := range catalog.All() {
		if e.Internal {
			internal = append(internal, e)
		}
	}
	doc, err := schemagen.AsyncAPIWith(internal, comments, schemagen.Info{
		Title:       "Discordia — Canal interno de messaging",
		Version:     "generated",
		Description: "Eventos que las instancias de `messaging` se mandan entre sí por el exchange `messaging.broadcast` (fanout, una cola durable por instancia) para que cada una empuje el aviso a los sockets que tiene abiertos. Solo los consume `messaging`: no son parte del bus entre servicios. Documento generado desde `events-contract` con `go run ./cmd/eventsgen docs-sync <carpeta asyncapi>`; no se edita a mano (ver ADR 0001, 0004 y 0006 para el diseño).",
	})
	fail(err)
	return doc
}

func docsSync(comments schemagen.Comments) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "uso: eventsgen docs-sync <carpeta asyncapi de docs>")
		os.Exit(2)
	}
	synced, err := schemagen.SyncDocsSchemas(os.Args[2], comments)
	fail(err)
	for name, content := range synced {
		fail(os.WriteFile(filepath.Join(os.Args[2], name), content, 0o644))
		fmt.Println("actualizado", name)
	}
	path := filepath.Join(os.Args[2], internalDocName)
	fail(os.WriteFile(path, internalDoc(comments), 0o644))
	fmt.Println("escrito", path)
}

func internalDocStale(dir string, comments schemagen.Comments) ([]string, error) {
	got, err := os.ReadFile(filepath.Join(dir, internalDocName))
	if os.IsNotExist(err) {
		return []string{internalDocName + ": falta; correr `eventsgen docs-sync`"}, nil
	}
	if err != nil {
		return nil, err
	}
	if string(got) != string(internalDoc(comments)) {
		return []string{internalDocName + ": desactualizado; correr `eventsgen docs-sync`"}, nil
	}
	return nil, nil
}
