package channels

import "github.com/discordia-grupo01/events-contract/envelope"

const (
	RoutingKeyChannelCreated = "servers.channel_created"
	RoutingKeyChannelDeleted = "servers.channel_deleted"
)

// ChannelCreated is published when a text or voice channel is created
// inside a server.
type ChannelCreated struct {
	envelope.Meta
	ServerID  string `json:"server_id"`
	ChannelID string `json:"channel_id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
}

// ChannelDeleted is published when a channel is removed from a server.
type ChannelDeleted struct {
	envelope.Meta
	ServerID  string `json:"server_id"`
	ChannelID string `json:"channel_id"`
}
