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

	RoutingKeyModerationActionRecorded = "servers.moderation_action_recorded"
)

// Values of ModerationActionRecorded.Action. They are the same texts servers
// stores in its moderation audit log.
const (
	ActionBan    = "ban"
	ActionUnban  = "unban"
	ActionKick   = "kick"
	ActionMute   = "mute"
	ActionUnmute = "unmute"
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

// ModerationActionRecorded is published by servers every time a moderation
// action (ban, unban, kick, mute or unmute) is written to its audit log. Admin
// consumes it to build the investigation dossier of a user without asking
// servers. Meta.OccurredAt is the moment of the action, the same occurred_at as
// the audit log entry.
//
// An audit entry never changes, so the event is immutable: ActionID is the id of
// the audit entry and the idempotency key, and a repeated delivery with the same
// ActionID must be ignored.
//
// It carries neither the moderator who acted nor the reason (data
// minimization): the dossier is about the target user, and who acted is only
// needed inside servers.
type ModerationActionRecorded struct {
	envelope.Meta
	ActionID     string `json:"action_id"`
	ServerID     string `json:"server_id"`
	Action       string `json:"action"`
	TargetUserID string `json:"target_user_id"`
}
