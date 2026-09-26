package moderation

import "github.com/discordia-grupo01/events-contract/envelope"

const (
	RoutingKeyMemberBanned   = "servers.member_banned"
	RoutingKeyMemberUnbanned = "servers.member_unbanned"
)

// MemberBanned is published when a member is banned from a server. The same
// transaction also publishes membership.MemberLeft, so consumers that only
// track membership don't need to know about bans.
type MemberBanned struct {
	envelope.Meta
	ServerID string `json:"server_id"`
	UserID   string `json:"user_id"`
	BannedBy string `json:"banned_by"`
}

// MemberUnbanned is published when a ban is revoked. It does not make the
// user a member again: they still need a valid invitation to rejoin.
type MemberUnbanned struct {
	envelope.Meta
	ServerID   string `json:"server_id"`
	UserID     string `json:"user_id"`
	UnbannedBy string `json:"unbanned_by"`
}
