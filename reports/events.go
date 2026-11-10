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
	ReportID string `json:"report_id" contract:"format=uuid"`
	ServerID string `json:"server_id" contract:"format=uuid"`
	// Qué se reportó.
	Kind string `json:"kind" contract:"enum=user|message"`
	// Usuario reportado (ID de discordia-identify). `null` si `kind` es `message`.
	ReportedUserID *string `json:"reported_user_id" contract:"format=uuid"`
	// Canal del mensaje reportado. `null` si `kind` es `user`.
	ChannelID *string `json:"channel_id" contract:"format=uuid"`
	// Mensaje reportado. `null` si `kind` es `user`.
	MessageID *string `json:"message_id" contract:"format=uuid"`
	// Motivo que escribió quien reportó.
	Reason string `json:"reason" contract:"minLength=1;maxLength=512"`
	Status string `json:"status" contract:"enum=pending|resolved|dismissed"`
	// Fecha de creación del reporte, ISO 8601 en UTC.
	CreatedAt time.Time `json:"created_at"`
	// Fecha de cierre del reporte, ISO 8601 en UTC. `null` mientras está `pending`.
	ResolvedAt *time.Time `json:"resolved_at"`
}
