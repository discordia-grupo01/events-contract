// Package admin has the events published by the back-office service when staff
// suspends or reactivates an account.
//
// These events have a big effect (a forged one could lock anyone out), so
// unlike the rest they are signed: admin and every consumer share a secret, and
// the consumer drops any event whose signature doesn't verify. RabbitMQ alone
// doesn't stop another client with broker credentials from publishing an
// admin.* routing key.
package admin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const (
	RoutingKeyUserSuspended   = "admin.user_suspended"
	RoutingKeyUserReactivated = "admin.user_reactivated"
)

// UserSuspended is published by admin when staff suspends an account. Every
// service that serves users (identify, servers, messaging, voice) reacts by
// rejecting anything that user does and cutting their open sessions.
//
// Version is admin's suspension_version for that user: it grows with every
// suspension or reactivation, and a consumer applies an event only if its
// version is higher than the one it already has, so duplicated or out-of-order
// deliveries are harmless. The reason for the suspension never travels
// (data minimisation).
type UserSuspended struct {
	envelope.Meta
	UserID    string `json:"user_id"`
	Version   int64  `json:"version"`
	Signature string `json:"signature"`
}

// UserReactivated is published by admin when staff lifts a suspension. Same
// fields and rules as UserSuspended.
type UserReactivated struct {
	envelope.Meta
	UserID    string `json:"user_id"`
	Version   int64  `json:"version"`
	Signature string `json:"signature"`
}

// Sign sets Signature using secret. Admin calls it right before publishing.
func (e *UserSuspended) Sign(secret string) {
	e.Signature = sign(secret, RoutingKeyUserSuspended, e.Meta, e.UserID, e.Version)
}

// Verify reports whether Signature is valid for secret. An empty secret or an
// empty/malformed signature never verifies.
func (e UserSuspended) Verify(secret string) bool {
	return verify(secret, RoutingKeyUserSuspended, e.Meta, e.UserID, e.Version, e.Signature)
}

// Sign sets Signature using secret. Admin calls it right before publishing.
func (e *UserReactivated) Sign(secret string) {
	e.Signature = sign(secret, RoutingKeyUserReactivated, e.Meta, e.UserID, e.Version)
}

// Verify reports whether Signature is valid for secret. An empty secret or an
// empty/malformed signature never verifies.
func (e UserReactivated) Verify(secret string) bool {
	return verify(secret, RoutingKeyUserReactivated, e.Meta, e.UserID, e.Version, e.Signature)
}

// signature is HMAC-SHA256(secret, routing_key \n event_id \n user_id \n
// version \n occurred_at), hex-encoded. The routing key is part of the signed
// text so a valid UserSuspended can't be replayed as a UserReactivated.
// occurred_at is formatted in UTC so it is the same after a JSON round trip.
func signature(secret, routingKey string, meta envelope.Meta, userID string, version int64) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.Join([]string{
		routingKey,
		meta.EventID,
		userID,
		strconv.FormatInt(version, 10),
		meta.OccurredAt.UTC().Format(time.RFC3339Nano),
	}, "\n")))
	return mac.Sum(nil)
}

func sign(secret, routingKey string, meta envelope.Meta, userID string, version int64) string {
	return hex.EncodeToString(signature(secret, routingKey, meta, userID, version))
}

func verify(secret, routingKey string, meta envelope.Meta, userID string, version int64, got string) bool {
	if secret == "" || got == "" {
		return false
	}
	gotBytes, err := hex.DecodeString(got)
	if err != nil {
		return false
	}
	return hmac.Equal(gotBytes, signature(secret, routingKey, meta, userID, version))
}
