package servers

import (
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const (
	RoutingKeyServerDeleted = "servers.server_deleted"
	RoutingKeyServerUpdated = "servers.server_updated"
)

type ServerDeleted struct {
	envelope.Meta
	ServerID   string   `json:"server_id"`
	ChannelIDs []string `json:"channel_ids"`
}

// ServerUpdated is published by servers when a server is created and every
// time its name or owner changes. It is a full snapshot, not a diff, so
// consumers don't need the previous state. There is no version: the consumer
// keeps the snapshot with the newest occurred_at of each server, so a repeated
// or out-of-order delivery doesn't overwrite newer data (same rule as the
// permissions events). Built for admin's back-office listing, which shows the
// name, owner and creation date of every server without asking servers.
type ServerUpdated struct {
	envelope.Meta
	ServerID  string    `json:"server_id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}
