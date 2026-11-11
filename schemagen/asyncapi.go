package schemagen

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/discordia-grupo01/events-contract/catalog"
)

func pascal(s string) string {
	var b strings.Builder
	for _, p := range strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '.' || r == '-' }) {
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}

func camel(s string) string {
	p := pascal(s)
	return strings.ToLower(p[:1]) + p[1:]
}

func MessageName(key string) string {
	prefix, suffix, _ := strings.Cut(key, ".")
	if prefix == "broadcast" {
		return pascal(key)
	}
	return pascal(suffix)
}

func serviceTitle(s string) string {
	return "discordia-" + s
}

type Info struct {
	Title       string
	Version     string
	Description string
}

var defaultInfo = Info{
	Title:       "Discordia — Contrato de eventos (generado)",
	Version:     "generated",
	Description: "Documento generado por `go run ./cmd/eventsgen gen` a partir de las structs de Go de `events-contract`. No se edita a mano: el CI falla si no coincide con el código. Cada payload se valida con los JSON Schema de la carpeta `schemas/`.",
}

func AsyncAPI(entries []catalog.Entry, c Comments) ([]byte, error) {
	return AsyncAPIWith(entries, c, defaultInfo)
}

func AsyncAPIWith(entries []catalog.Entry, c Comments, info Info) ([]byte, error) {
	channels := NewObj()
	operations := NewObj()
	messages := NewObj()
	schemas := NewObj()
	names := map[string]string{}

	for _, e := range entries {
		name := MessageName(e.RoutingKey)
		if other, dup := names[name]; dup {
			return nil, fmt.Errorf("el nombre %s sale de %s y de %s", name, other, e.RoutingKey)
		}
		names[name] = e.RoutingKey
		full, err := EventSchema(e, c)
		if err != nil {
			return nil, err
		}
		body := NewObj()
		for _, k := range full.Keys() {
			if k == "$schema" || k == "$id" || k == "title" {
				continue
			}
			v, _ := full.Get(k)
			body.Set(k, v)
		}
		schemas.Set(name, body)

		msg := NewObj().Set("name", name).Set("title", e.RoutingKey).
			Set("contentType", "application/json").
			Set("payload", NewObj().Set("$ref", "#/components/schemas/"+name))
		example, err := Example(e, c)
		if err != nil {
			return nil, err
		}
		msg.Set("examples", []any{map[string]any{"payload": example}})
		if e.Signed {
			msg.Set("description", "Evento firmado con HMAC: el campo `signature` se calcula con el secreto compartido y el consumidor debe verificarlo antes de aplicar el cambio.")
		}
		messages.Set(name, msg)

		chID := camel(e.RoutingKey)
		ch := NewObj().Set("address", e.RoutingKey).Set("title", e.RoutingKey)
		if d := c.Types[e.Type().String()]; d != "" {
			ch.Set("description", d)
		}
		ch.Set("messages", NewObj().Set(camel(name), NewObj().Set("$ref", "#/components/messages/"+name)))
		ch.Set("bindings", NewObj().Set("$ref", "#/components/channelBindings/"+bindingName(e)))
		channels.Set(chID, ch)

		chRef := NewObj().Set("$ref", "#/channels/"+chID)
		operations.Set(camel(e.Producer)+"Publishes"+name, NewObj().
			Set("action", "send").
			Set("title", fmt.Sprintf("%s publica %s", serviceTitle(e.Producer), e.RoutingKey)).
			Set("channel", chRef).
			Set("bindings", NewObj().Set("amqp", NewObj().Set("deliveryMode", 2))))
		for _, cons := range e.Consumers {
			operations.Set(camel(cons)+"Receives"+name, NewObj().
				Set("action", "receive").
				Set("title", fmt.Sprintf("%s consume %s", serviceTitle(cons), e.RoutingKey)).
				Set("channel", NewObj().Set("$ref", "#/channels/"+chID)).
				Set("bindings", NewObj().Set("amqp", NewObj().Set("ack", true))))
		}
	}

	root := NewObj().
		Set("asyncapi", "3.0.0").
		Set("info", NewObj().
			Set("title", info.Title).
			Set("version", info.Version).
			Set("description", info.Description)).
		Set("servers", NewObj().Set("rabbitmq", NewObj().
			Set("host", "{host}").
			Set("protocol", "amqp").
			Set("description", "Broker RabbitMQ.").
			Set("variables", NewObj().Set("host", NewObj().Set("default", "localhost:5672"))))).
		Set("defaultContentType", "application/json").
		Set("channels", channels).
		Set("operations", operations).
		Set("components", NewObj().
			Set("channelBindings", channelBindings(entries)).
			Set("messages", messages).
			Set("schemas", schemas))

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root.YAMLNode()); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func bindingName(e catalog.Entry) string {
	if e.ExchangeType == "fanout" {
		return "messagingBroadcast"
	}
	return "discordiaEvents"
}

func channelBindings(entries []catalog.Entry) *Obj {
	out := NewObj()
	seen := map[string]bool{}
	for _, e := range entries {
		name := bindingName(e)
		if seen[name] {
			continue
		}
		seen[name] = true
		ex := NewObj().Set("name", e.Exchange).Set("type", e.ExchangeType).Set("durable", true).Set("autoDelete", false).Set("vhost", "/")
		is := "routingKey"
		out.Set(name, NewObj().Set("amqp", NewObj().Set("is", is).Set("exchange", ex)))
	}
	return out
}
