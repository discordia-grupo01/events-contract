# events-contract

Contrato de los eventos que viajan por RabbitMQ entre los servicios de Discordia.

## Tabla de contenidos

- [Introducción](#introducción)
- [Pre-requisitos](#pre-requisitos)
- [Cómo correr](#cómo-correr)
- [Variables de entorno](#variables-de-entorno)
- [Comandos Docker](#comandos-docker)
- [Tests](#tests)
- [Más documentación](#más-documentación)

## Introducción

Los servicios de Discordia se avisan los cambios por eventos (por ejemplo, `servers.channel_created`). Para que ninguno mantenga su propia copia de cómo es cada evento, este repositorio define el contrato en un solo lugar: el nombre de cada evento, sus campos, quién lo publica y quién lo consume.

La fuente de verdad son las structs de Go de este repositorio. De ellas se generan solos:

```
structs de Go  ──go run ./cmd/eventsgen gen──▶  schemas/<evento>.json   (JSON Schema)
                                                 examples/<evento>.json  (un ejemplo válido de cada uno)
                                                 asyncapi/events.asyncapi.yaml (AsyncAPI completo)
```

Los servicios lo usan como dependencia: los de Go importan las structs, y los de Elixir usan un validador que lee los mismos JSON Schema. Hoy los servicios validan cada evento que publican y consumen y, si no cumple el contrato, lo anotan en el log sin rechazarlo.

Hay tres cosas que conviene saber:

- **Los eventos firmados.** Los que cambian el estado de un usuario o de un servidor (`admin.user_suspended`, `billing.premium_updated`) van firmados con HMAC, para que otro servicio no pueda falsificarlos.
- **Los avisos de cambio en tiempo real.** `servers.resource_changed`, `billing.account_changed` y `admin.backoffice_changed` no llevan datos, solo qué cambió. `notifications` se los empuja a los clientes por WebSocket y cada cliente vuelve a pedir lo que necesita por REST.
- **La compatibilidad.** En cada pull request, `eventsgen compat` compara los schemas con los de la rama base y falla si el cambio rompería a algún consumidor: quitar un evento o un campo, cambiar un tipo o agregar un campo obligatorio nuevo.

## Pre-requisitos

- **Go 1.22**, para generar y probar el contrato.
- **Elixir 1.16.2** y **Erlang/OTP 26.2**, solo para probar el validador en Elixir.
- **Docker**, solo si preferís correr los tests sin instalar Go.

## Cómo correr

No es un servicio: se genera y se prueba.

### Agregar un evento

1. Escribir la struct del evento en la carpeta de su dominio, con la constante `RoutingKey...`.
2. Sumar una línea en `catalog/catalog.go` con quién lo publica y quién lo consume.
3. Regenerar los schemas, los ejemplos y el AsyncAPI:

```bash
go run ./cmd/eventsgen gen
```

4. Comitear lo que se generó. Si no coincide con el código, el CI falla.

Los detalles de cada campo (formato, largo, valores permitidos, si es opcional) se escriben en la etiqueta `contract` de la struct, separados por `;`: `contract:"format=uuid"`, `contract:"maxLength=160"`, `contract:"enum=a|b"`, `contract:"optional=true"`. Los comentarios de Go pasan a ser las descripciones.

### Usarlo desde un servicio en Go

```go
import (
    "github.com/discordia-grupo01/events-contract/envelope"
    "github.com/discordia-grupo01/events-contract/membership"
)

publisher.Publish(ctx, membership.RoutingKeyMemberJoined, membership.MemberJoined{
    Meta:     envelope.Meta{EventID: id, OccurredAt: time.Now()},
    ServerID: serverID,
    UserID:   userID,
})
```

Para validar un evento: `validate.Validate(routingKey, cuerpoJSON)`. En Elixir: `Discordia.EventsContract.Validator.validate(routing_key, mapa_o_json)`.

Cada servicio fija la versión en su `go.mod` y la actualiza con:

```bash
go get github.com/discordia-grupo01/events-contract@v0.23.0
```

### Verificar contra los documentos

El repositorio [`docs`](https://github.com/discordia-grupo01/docs) conserva la explicación de cada evento en prosa. Para comprobar que coincide con el contrato:

```bash
go run ./cmd/eventsgen docs-check ../docs/asyncapi
```

## Variables de entorno

No usa ninguna.

## Comandos Docker

No tiene `docker-compose.yml`: no es un servicio. Para correr los tests sin instalar Go:

```bash
docker run --rm -v "$PWD":/app -w /app golang:1.22-alpine go test ./... -cover
```

## Tests

```bash
go test ./... -cover
mix test
```

El primero prueba las structs de Go, el generador y los vectores de firma. El segundo prueba el validador de Elixir. Los vectores de firma de `admin/vectors_test.go` y `admin/events_test.exs` son los mismos: si el texto firmado cambia, los dos se rompen y las copias que usan `messaging` y `voice-sfu` tienen que cambiar con ellos.

El CI corre Go y Elixir, y no publica una versión nueva si falla alguno.

### Versiones

Este repositorio no se despliega: el "despliegue" es un tag, y el CI lo crea solo cuando un merge llega a `main` y pasan los tests. Si algún commit desde el último tag empieza con `feat` o dice `BREAKING CHANGE`, sube la versión minor (`v0.9.1` a `v0.10.0`); si no, sube la patch (`v0.9.1` a `v0.9.2`).

## Más documentación

- [`asyncapi/events.asyncapi.yaml`](asyncapi/events.asyncapi.yaml): todos los eventos, generados desde el código.
- [`catalog/catalog.go`](catalog/catalog.go): quién publica y quién consume cada evento.
- [`docs`](https://github.com/discordia-grupo01/docs/tree/main/asyncapi): la explicación de cada evento y de cada dominio.
