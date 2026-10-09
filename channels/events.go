package channels

import "github.com/discordia-grupo01/events-contract/envelope"

const (
	RoutingKeyChannelCreated       = "servers.channel_created"
	RoutingKeyChannelDeleted       = "servers.channel_deleted"
	RoutingKeyChannelAccessRevoked = "servers.channel_access_revoked"
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

// ChannelAccessRevoked is published when members who could see a channel
// (VIEW_CHANNELS after its per-role overrides) no longer can: an override or
// a role's permissions changed, a role was assigned, removed or deleted, the
// ownership was transferred, or the member was kicked, banned or left. There
// is one event per affected channel, text or voice; UserIDs holds only the
// members who lost access, never empty. Consumers cut what those users have
// open in the channel (a messaging socket, a voice room); gaining access is
// never published.
type ChannelAccessRevoked struct {
	envelope.Meta
	ServerID  string   `json:"server_id"`
	ChannelID string   `json:"channel_id"`
	UserIDs   []string `json:"user_ids"`
}
