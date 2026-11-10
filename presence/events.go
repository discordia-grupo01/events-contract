package presence

import "github.com/discordia-grupo01/events-contract/envelope"

const RoutingKeyUserPresenceChanged = "messaging.user_presence_changed"

const (
	ConnectionsActive = "active"
	ConnectionsIdle   = "idle"
	ConnectionsNone   = "none"
)

type UserPresenceChanged struct {
	envelope.Meta
	UserID string `json:"user_id" contract:"format=uuid"`
	// Estado real de las conexiones del usuario, de todas sus instancias y dispositivos.
	Connections string `json:"connections" contract:"enum=active|idle|none"`
	// `true` si el usuario eligió No molestar. Se manda siempre.
	DoNotDisturb bool `json:"do_not_disturb"`
	// Crece con cada cambio del mismo usuario. El consumidor aplica el evento solo si es mayor a la que ya tiene.
	Version int64 `json:"version" contract:"format=int64"`
}
