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
	UserID       string `json:"user_id"`
	Connections  string `json:"connections"`
	DoNotDisturb bool   `json:"do_not_disturb"`
	Version      int64  `json:"version"`
}
