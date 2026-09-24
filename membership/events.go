package membership

import "github.com/discordia-grupo01/events-contract/envelope"

const (
	RoutingKeyMemberJoined = "servers.member_joined"
	RoutingKeyMemberLeft   = "servers.member_left"
)

// MemberJoined is published when a user joins a server.
type MemberJoined struct {
	envelope.Meta
	ServerID string `json:"server_id"`
	UserID   string `json:"user_id"`
}

// MemberLeft is published when a user leaves (or is removed from) a server.
type MemberLeft struct {
	envelope.Meta
	ServerID string `json:"server_id"`
	UserID   string `json:"user_id"`
}
