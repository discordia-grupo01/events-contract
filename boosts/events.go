// Package boosts has the events of the saga that boosts a server. billing owns
// the points and servers owns the membership, the suspension and the level, so
// they coordinate by events:
//
//	billing.server_boost_requested  billing reserved the points
//	servers.server_boost_applied    servers counted the boost
//	servers.server_boost_rejected   servers refused it: billing refunds
//	billing.server_boost_cancelled  no answer in time: billing refunded
//	billing.server_boost_removed    the owner withdrew it, no refund
//
// Those five are signed: a forged rejection would give points away and a forged
// request would give boosts for free. billing and servers share a secret, and
// the consumer drops any event whose signature doesn't verify.
//
// servers.server_boost_updated tells everyone else the server's new level.
package boosts

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const (
	RoutingKeyRequested = "billing.server_boost_requested"
	RoutingKeyCancelled = "billing.server_boost_cancelled"
	RoutingKeyRemoved   = "billing.server_boost_removed"
	RoutingKeyApplied   = "servers.server_boost_applied"
	RoutingKeyRejected  = "servers.server_boost_rejected"
	RoutingKeyUpdated   = "servers.server_boost_updated"
)

// Why servers rejected a boost. A rejection is always a business rule: a
// technical failure requeues the event instead, so it never refunds points.
const (
	ReasonServerNotFound  = "server_not_found"
	ReasonServerSuspended = "server_suspended"
	ReasonNotAMember      = "not_a_member"
	ReasonCancelled       = "cancelled"
)

// SagaEvent is the message of every signed step. Processing it twice has no
// effect: BoostID identifies the boost on both sides.
type SagaEvent struct {
	envelope.Meta
	BoostID  string `json:"boost_id"`
	ServerID string `json:"server_id"`
	UserID   string `json:"user_id"`
	// Reason is only set on servers.server_boost_rejected.
	Reason    string `json:"reason,omitempty"`
	Signature string `json:"signature"`
}

// Sign sets Signature using secret, right before publishing with routingKey. It
// also moves OccurredAt to UTC, so the JSON carries exactly the text signed.
func (e *SagaEvent) Sign(routingKey, secret string) {
	e.OccurredAt = e.OccurredAt.UTC()
	e.Signature = hex.EncodeToString(e.signature(routingKey, secret))
}

// Verify reports whether Signature is valid for the routing key the event
// arrived with. An empty secret or an empty/malformed signature never verifies.
func (e SagaEvent) Verify(routingKey, secret string) bool {
	if secret == "" || e.Signature == "" {
		return false
	}
	got, err := hex.DecodeString(e.Signature)
	if err != nil {
		return false
	}
	return hmac.Equal(got, e.signature(routingKey, secret))
}

// signature is HMAC-SHA256(secret, routing_key \n event_id \n boost_id \n
// server_id \n user_id \n reason \n occurred_at). The routing key is part of it
// so a signed "applied" can't be replayed as a "removed". Times are RFC 3339 in
// UTC, so the text is the same after a JSON round trip.
func (e SagaEvent) signature(routingKey, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.Join([]string{
		routingKey,
		e.EventID,
		e.BoostID,
		e.ServerID,
		e.UserID,
		e.Reason,
		e.OccurredAt.UTC().Format(time.RFC3339Nano),
	}, "\n")))
	return mac.Sum(nil)
}

// ServerBoostUpdated is a full snapshot of a server's boosts, not a diff. A
// consumer applies it only if Version is higher than the one it has, so
// duplicated or out-of-order deliveries are harmless.
type ServerBoostUpdated struct {
	envelope.Meta
	ServerID   string `json:"server_id"`
	BoostCount int    `json:"boost_count"`
	Level      int    `json:"level"`
	Version    int64  `json:"version"`
}
