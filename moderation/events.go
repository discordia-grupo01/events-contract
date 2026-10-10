package moderation

import (
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const (
	RoutingKeyMemberBanned   = "servers.member_banned"
	RoutingKeyMemberUnbanned = "servers.member_unbanned"
	RoutingKeyMemberKicked   = "servers.member_kicked"
	RoutingKeyMemberMuted    = "servers.member_muted"
	RoutingKeyMemberUnmuted  = "servers.member_unmuted"

	RoutingKeyWordFilterUpdated = "servers.word_filter_updated"
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

// MemberKicked is published when a member is kicked from a server. Unlike a
// ban, nothing stops them from rejoining with a valid invitation. The same
// transaction also publishes membership.MemberLeft, so consumers that only
// track membership don't need to know about kicks.
type MemberKicked struct {
	envelope.Meta
	ServerID string `json:"server_id"`
	UserID   string `json:"user_id"`
	KickedBy string `json:"kicked_by"`
}

// MemberMuted is published when a member loses the ability to send messages
// and publish voice for a limited period. The session remains connected.
type MemberMuted struct {
	envelope.Meta
	ServerID string    `json:"server_id"`
	UserID   string    `json:"user_id"`
	MutedBy  string    `json:"muted_by"`
	UntilAt  time.Time `json:"until_at"`
	Reason   string    `json:"reason"`
}

// MemberUnmuted is published when a mute is manually revoked or expires.
type MemberUnmuted struct {
	envelope.Meta
	ServerID  string `json:"server_id"`
	UserID    string `json:"user_id"`
	UnmutedBy string `json:"unmuted_by"`
	Reason    string `json:"reason"`
}

// WordFilterUpdated is published every time a server's banned words list
// changes. It carries the whole list, not a diff, so consumers keep the one
// with the newest occurred_at and a repeated or late event never undoes a
// newer change. Words must never be nil: an empty list (no filter) has to
// marshal as [] and not as null.
type WordFilterUpdated struct {
	envelope.Meta
	ServerID string   `json:"server_id"`
	Words    []string `json:"words"`
}
