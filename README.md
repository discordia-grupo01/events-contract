# events-contract

Structs/tipos compartidos para los eventos que viajan por `discordia.events`
(RabbitMQ) -- para que ningún servicio mantenga su propia copia del mismo
evento.

La fuente de verdad del *schema* sigue siendo el AsyncAPI en
[`docs`](https://github.com/discordia-grupo01/docs); acá viven las
implementaciones de ese schema. Un cambio de schema se hace primero en el
AsyncAPI y después se replica acá.

## Estructura

Una carpeta por dominio/evento, no por lenguaje -- por ahora todo es Go
(único lenguaje que hay hoy), y el día que un evento necesite binding
Elixir, el archivo va al lado del `.go` en la carpeta de ese mismo evento,
no en un árbol paralelo:

```
envelope/   -- campos comunes a todo evento (event_id, occurred_at)
broker/     -- topología compartida del exchange (nombre, tipo)
membership/
  events.go       -- MemberJoined, MemberLeft (Go)
  events_test.go
  events.ex       -- binding Elixir de MemberLeft (consumido por messaging)
channels/
  events.go       -- ChannelCreated, ChannelDeleted (Go, publicados por servers)
  events_test.go
  events.ex       -- mismo binding en Elixir (consumido por messaging)
moderation/
  events.go       -- MemberBanned, MemberUnbanned, MemberKicked (Go)
  events_test.go
servers/
  events.go       -- ServerDeleted (Go, publicado por servers)
  events_test.go
  events.ex       -- mismo binding en Elixir (consumido por messaging)
users/
  events.go       -- ProfileUpdated (Go), publicado por identify
  events_test.go
```

Un ban publica, en la misma transacción, `servers.member_left` (de
`membership/`) y `servers.member_banned` (de `moderation/`): quien solo
lleva la membresía escucha el primero; quien tiene que reaccionar al ban
(cortar sesiones de mensajería/voz, auditoría) escucha el segundo. El
motivo del ban no viaja en el evento (minimización de datos).

Una expulsión hace lo mismo con `servers.member_kicked`: publica también
`servers.member_left` en la misma transacción. A diferencia del ban, el
expulsado puede volver a unirse con una invitación válida. Pensado para que
`notifications` le avise al expulsado.

Eliminar un servidor publica, en la misma transacción que el borrado, un
`servers.member_left` por cada miembro (quien lleva la membresía, como
`identify-service`, no necesita saber nada nuevo) y un
`servers.server_deleted` (de `servers/`) con los ids de todos sus canales, para
quien guarda datos por canal (`messaging` borra ahí los mensajes).

`users.profile_updated` es una foto completa del perfil público (nombre,
avatar, descripción y estado), no un diff: identify lo publica al crear un
usuario y en cada cambio de perfil o estado. `version` crece con cada cambio
del mismo usuario; el consumidor aplica una foto solo si su `version` es
mayor a la que ya tiene, así que eventos repetidos o fuera de orden no pisan
datos más nuevos. El email no viaja nunca (minimización de datos).

Cada evento es un tipo con nombre propio (embebe los campos comunes, no los
repite) -- así uno puede evolucionar sin arrastrar al otro (ej. agregarle
`Reason` a `MemberLeft` no toca `MemberJoined`).

Al sumar el próximo evento (de `servers` o de otro servicio), se agrega una
carpeta nueva con el mismo patrón -- no se toca `membership/`.

## Uso (Go)

```go
import (
    "github.com/discordia-grupo01/events-contract/broker"
    "github.com/discordia-grupo01/events-contract/membership"
)

publisher.Publish(ctx, membership.RoutingKeyMemberJoined, membership.MemberJoined{
    Meta:     envelope.Meta{EventID: id, OccurredAt: time.Now()},
    ServerID: serverID,
    UserID:   userID,
})
```

Como el repo es público, `go get` no necesita token ni `GOPRIVATE` -- ni en
CI ni en el build de Docker de los consumidores.

## Versionado

Este repo no se despliega -- el "CD" es taguear:

```bash
git tag v0.3.0
git push origin v0.3.0
```

Los consumidores (`servers`, `identify-service`) fijan la versión en su
`go.mod` y la actualizan con:

```bash
go get github.com/discordia-grupo01/events-contract@v0.3.0
```

Un cambio que rompe compatibilidad (renombrar un campo, sacar uno) primero
se decide en el AsyncAPI de `docs`, y después se migra publisher y
consumer juntos en el mismo despliegue.
