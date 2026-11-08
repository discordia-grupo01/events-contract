package reports

import (
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const (
	RoutingKeyReportUpdated = "servers.report_updated"
)

const (
	KindUser    = "user"
	KindMessage = "message"
)

const (
	StatusPending   = "pending"
	StatusResolved  = "resolved"
	StatusDismissed = "dismissed"
)

// ReportUpdated is published by servers when a user or message report is
// created and every time its status changes. Admin consumes it to build the
// investigation dossier of a user without asking servers.
//
// It is a full snapshot of the report, not a diff. There is no version: the
// consumer keeps the snapshot with the newest occurred_at of each report, so a
// repeated or out-of-order delivery doesn't overwrite newer data (same rule as
// servers.ServerUpdated).
//
// Optional fields are always present and marshal as null when empty (no
// omitempty): ReportedUserID is null when Kind is KindMessage, ChannelID and
// MessageID are null when Kind is KindUser, and ResolvedAt is null while the
// report is pending.
//
// It carries neither the reporter nor the moderator who resolved it (data
// minimization): the dossier is about the reported user, and who reported or
// resolved is only needed inside servers.
type ReportUpdated struct {
	envelope.Meta
	ReportID       string     `json:"report_id"`
	ServerID       string     `json:"server_id"`
	Kind           string     `json:"kind"`
	ReportedUserID *string    `json:"reported_user_id"`
	ChannelID      *string    `json:"channel_id"`
	MessageID      *string    `json:"message_id"`
	Reason         string     `json:"reason"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
}
