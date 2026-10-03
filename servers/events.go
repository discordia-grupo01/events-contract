package servers

import "github.com/discordia-grupo01/events-contract/envelope"

const RoutingKeyServerDeleted = "servers.server_deleted"

type ServerDeleted struct {
	envelope.Meta
	ServerID   string   `json:"server_id"`
	ChannelIDs []string `json:"channel_ids"`
}
