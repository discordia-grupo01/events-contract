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
  events.go       -- ChannelCreated, ChannelDeleted, ChannelAccessRevoked (Go, publicados por servers)
  events_test.go
  events.ex       -- mismo binding en Elixir (consumido por messaging)
moderation/
  events.go       -- MemberBanned, MemberUnbanned, MemberKicked, MemberMuted,
                     MemberUnmuted, WordFilterUpdated (Go)
  events_test.go
servers/
  events.go       -- ServerDeleted, ServerUpdated (Go, publicados por servers)
  events_test.go
  events.ex       -- mismo binding en Elixir (consumido por messaging)
admin/
  events.go       -- UserSuspended, UserReactivated (Go, publicados por admin,
                     firmados con HMAC: Sign/Verify)
  events_test.go
  vectors_test.go -- vectores de firma fijos que todo port a Elixir debe reproducir
  events.ex       -- binding Elixir de referencia con sign/6 y verify/3
                     (consumido por messaging y voice)
users/
  events.go       -- ProfileUpdated, AccountRegistered (Go), publicados por identify
  events_test.go
permissions/
  events.go       -- RoleUpdated, RoleDeleted, MemberRolesUpdated,
                     ChannelOverridesUpdated, OwnerUpdated (Go, publicados por servers)
  events_test.go
  events.ex       -- binding de referencia en Elixir (consumido por messaging)
voice/
  events.go       -- voice.participants_joined, voice.participant_left (consume server)
  events.ex       -- Mensajes publicados por Elixir signaling
  events_test.go 
messages/
  events.go       -- MessageCreated (Go), publicado por messaging al crear un mensaje
                     en un canal de servidor (sin el texto del mensaje)
  events_test.go
  events.ex       -- binding Elixir (publicado por messaging, consumido por notifications)
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

`servers.channel_access_revoked` (de `channels/`) avisa qui�nes dejaron de
ver un canal: miembros que ten�an `VIEW_CHANNELS` en �l (despu�s de los
overrides por rol, #57) y ya no lo tienen. servers lo calcula comparando el
antes y el despu�s dentro de la misma transacci�n del cambio, y publica uno
por canal afectado (de texto o de voz), con `user_ids` = solo los que
perdieron acceso (nunca vac�o). Lo disparan: crear/editar/borrar un override,
editar o borrar un rol, asignar o quitar un rol, aceptar una transferencia de
propiedad, y expulsar, banear o abandonar (en esos tres casos adem�s del
`servers.member_left`). Ganar acceso no se publica. Pensado para cortar lo que
el usuario tenga abierto en ese canal: `messaging` cierra su socket y
`voice-sfu` lo saca de la sala de voz. Como solo corta y al reconectar se
vuelve a chequear el estado actual, un evento repetido o fuera de orden cuesta
a lo sumo una reconexi�n de m�s -- no hace falta descartar eventos viejos.

`users.profile_updated` es una foto completa del perfil público (nombre,
avatar, descripción y estado), no un diff: identify lo publica al crear un
usuario y en cada cambio de perfil o estado. `version` crece con cada cambio
del mismo usuario; el consumidor aplica una foto solo si su `version` es
mayor a la que ya tiene, así que eventos repetidos o fuera de orden no pisan
datos más nuevos. El email no viaja nunca (minimización de datos).

`users.account_registered` lo publica identify una sola vez, al crear la cuenta,
con los datos de la cuenta que no son perfil público: `email`, `name` y
`created_at`. Existe únicamente para el listado del backoffice (`admin`), que
tiene que buscar usuarios por email; por eso es la excepción a "el email no
viaja nunca" de `users.profile_updated`, y ningún otro consumidor debería
suscribirse. `email` y `created_at` no cambian, así que no lleva `version`:
aplicarlo dos veces deja los mismos valores. Para las cuentas que ya existían
cuando se sumó el evento, identify re-publica uno por usuario (backfill).

`servers.server_updated` (de `servers/`) es una foto completa del servidor
(nombre, dueño y fecha de creación), no un diff: servers la publica al crear el
servidor y en cada cambio de nombre o de dueño. Como los eventos de
`permissions/`, no lleva `version`: el consumidor se queda con la foto de
`occurred_at` más nueva de cada servidor, así que un evento repetido o fuera de
orden no pisa datos más recientes. Lo consume `admin` para armar su listado de
servidores sin preguntarle a servers; el número de miembros lo calcula con
`servers.member_joined` y `servers.member_left`. Al sumar el consumidor hay que
relanzar el backfill de servers, porque RabbitMQ no guarda lo publicado antes
de que existiera su cola.

`admin.user_suspended` y `admin.user_reactivated` (de `admin/`) los publica
`admin` cuando el staff suspende o reactiva una cuenta, en la misma
transacción que el cambio (outbox). Los consumen identify (revoca los refresh
tokens y rechaza login y refresh), servers, messaging y voz (rechazan toda
acción del usuario y cortan sus sockets y su sala de voz). `version` crece con
cada suspensión o reactivación del mismo usuario y el consumidor aplica el
evento solo si es mayor a la que ya tiene, así que un evento repetido o fuera
de orden es inofensivo. El motivo de la suspensión no viaja.

Son los únicos eventos **firmados**: una suspensión falsa dejaría a cualquiera
sin acceso y RabbitMQ no impide que otro cliente con credenciales del broker
publique una routing key `admin.*`. `admin` y cada consumidor comparten
`ADMIN_EVENTS_HMAC_SECRET`; el publisher llama a `Sign(secret)` justo antes de
publicar y el consumidor descarta (sin reencolar) todo evento cuyo `Verify(secret)`
da falso. La firma es HMAC-SHA256 sobre la routing key, `event_id`, `user_id`,
`version` y `occurred_at`; incluir la routing key evita reusar una suspensión
válida como si fuera una reactivación. Un secreto vacío nunca verifica.

`servers.word_filter_updated` (de `moderation/`) es la lista completa de
palabras prohibidas de un servidor, no un diff: servers la publica en cada
cambio (lista vacía si se borraron todas). `messaging` la usa para rechazar
mensajes que las contengan; se queda con la de `occurred_at` más nuevo, así
que un evento repetido o fuera de orden no pisa una lista más reciente.

Los eventos de `permissions/` son la réplica de "quién puede qué" para quien
no quiere preguntarle a servers en cada pedido (`messaging` autoriza cada
mensaje con ellos). Son fotos completas de una entidad, no diferencias, igual
que `users.profile_updated`: `servers.role_updated` (nombre de los permisos del
rol e `is_everyone`; se publica al crear y en cada cambio),
`servers.role_deleted`, `servers.member_roles_updated` (el conjunto completo de
roles de un miembro, @everyone incluido; se publica al unirse y en cada
asignación o quita), `servers.channel_overrides_updated` (todos los overrides de
un canal; lista vacía si no queda ninguno) y `servers.owner_updated` (al crear
el servidor y en cada transferencia). Los permisos viajan por nombre
(`VIEW_CHANNELS`, `SEND_MESSAGES`, ...), nunca como máscara de bits, para no
atar al consumidor al orden interno de servers. El consumidor se queda con la
foto de `occurred_at` más nueva de cada entidad, así que un evento repetido o
fuera de orden no pisa datos más recientes. Borrar un rol no publica un evento
por cada miembro u override que lo tenía: el consumidor lo saca de los
miembros y de los overrides al recibir `servers.role_deleted`, y borrar un
canal o un servidor arrastra sus overrides. Al sumar un consumidor nuevo hay
que relanzar el backfill (migración de servers que re-publica todo) porque
RabbitMQ no guarda lo que se publicó antes de que su cola existiera.

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

Este repo no se despliega -- el "CD" es taguear, y **el tag lo crea el CI solo**: cuando
un merge llega a `main` y pasan los tests, el job `tag` de `.github/workflows/ci.yml` calcula
la próxima versión, crea el tag y un GitHub Release con las notas generadas. No hay que hacerlo
a mano.

- Si algún commit desde el último tag empieza con `feat` / `[feat]` (o dice `BREAKING CHANGE`),
  sube la **minor** (`v0.9.1` -> `v0.10.0`).
- Si no (`fix`, `docs`, etc.), sube la **patch** (`v0.9.1` -> `v0.9.2`).
- Si `HEAD` ya tiene tag, no hace nada.

Para forzar un número distinto, se sigue pudiendo taguear a mano:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Los consumidores (`servers`, `identify-service`) fijan la versión en su
`go.mod` y la actualizan con:

```bash
go get github.com/discordia-grupo01/events-contract@v0.3.0
```

Un cambio que rompe compatibilidad (renombrar un campo, sacar uno) primero
se decide en el AsyncAPI de `docs`, y después se migra publisher y
consumer juntos en el mismo despliegue.
