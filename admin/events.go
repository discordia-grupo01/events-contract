// Package admin has the events published by the back-office service when staff
// suspends or reactivates an account, and when it suspends, reactivates or
// asks to delete a server.
//
// These events have a big effect (a forged one could lock anyone out or delete
// a server for good), so unlike the rest they are signed: admin and every
// consumer share a secret, and
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

	RoutingKeyServerSuspended         = "admin.server_suspended"
	RoutingKeyServerReactivated       = "admin.server_reactivated"
	RoutingKeyServerDeletionRequested = "admin.server_deletion_requested"
	RoutingKeyStaffMemberAdded        = "admin.staff_member_added"
	RoutingKeyStaffMemberRemoved      = "admin.staff_member_removed"
)

type StaffMemberAdded struct {
	envelope.Meta
	UserID    string `json:"user_id" contract:"format=uuid"`
	Signature string `json:"signature"`
}

type StaffMemberRemoved struct {
	envelope.Meta
	UserID    string `json:"user_id" contract:"format=uuid"`
	Signature string `json:"signature"`
}

func (e *StaffMemberAdded) Sign(secret string) {
	e.Signature = signStaff(secret, RoutingKeyStaffMemberAdded, e.Meta, e.UserID)
}

func (e StaffMemberAdded) Verify(secret string) bool {
	return verifyStaff(secret, RoutingKeyStaffMemberAdded, e.Meta, e.UserID, e.Signature)
}

func (e *StaffMemberRemoved) Sign(secret string) {
	e.Signature = signStaff(secret, RoutingKeyStaffMemberRemoved, e.Meta, e.UserID)
}

func (e StaffMemberRemoved) Verify(secret string) bool {
	return verifyStaff(secret, RoutingKeyStaffMemberRemoved, e.Meta, e.UserID, e.Signature)
}

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
	// Usuario suspendido o reactivado (ID de discordia-identify).
	UserID string `json:"user_id" contract:"format=uuid"`
	// suspension_version del usuario en admin. Crece con cada suspensión o reactivación; ordena los eventos del mismo usuario.
	Version int64 `json:"version" contract:"format=int64;minimum=1"`
	// HMAC-SHA256 en hexadecimal. Ver la sección "Firma" de la descripción del documento.
	Signature string `json:"signature"`
}

// UserReactivated is published by admin when staff lifts a suspension. Same
// fields and rules as UserSuspended.
type UserReactivated struct {
	envelope.Meta
	// Usuario suspendido o reactivado (ID de discordia-identify).
	UserID string `json:"user_id" contract:"format=uuid"`
	// suspension_version del usuario en admin. Crece con cada suspensión o reactivación; ordena los eventos del mismo usuario.
	Version int64 `json:"version" contract:"format=int64;minimum=1"`
	// HMAC-SHA256 en hexadecimal. Ver la sección "Firma" de la descripción del documento.
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

// ServerSuspended is published by admin when staff suspends a server. The
// server keeps its content, but nobody can use it until it is reactivated:
// servers rejects every action on it except leaving, messaging rejects reads
// and writes in its channels and closes their open sockets, and voice takes
// whoever is connected out of the room.
//
// Version is admin's moderation_version for that server: it grows with every
// suspension, reactivation or deletion request, and a consumer applies an
// event only if its version is higher than the one it already has. The reason
// never travels (data minimisation).
type ServerSuspended struct {
	envelope.Meta
	// Servidor suspendido, reactivado o a eliminar (ID de discordia-servers).
	ServerID string `json:"server_id" contract:"format=uuid"`
	// moderation_version del servidor en admin. Crece con cada suspensión, reactivación o eliminación; ordena los eventos del mismo servidor.
	Version int64 `json:"version" contract:"format=int64;minimum=1"`
	// HMAC-SHA256 en hexadecimal. Ver la sección "Firma" de la descripción del documento.
	Signature string `json:"signature"`
}

// ServerReactivated is published by admin when staff lifts a server
// suspension. Same fields and rules as ServerSuspended.
type ServerReactivated struct {
	envelope.Meta
	// Servidor suspendido, reactivado o a eliminar (ID de discordia-servers).
	ServerID string `json:"server_id" contract:"format=uuid"`
	// moderation_version del servidor en admin. Crece con cada suspensión, reactivación o eliminación; ordena los eventos del mismo servidor.
	Version int64 `json:"version" contract:"format=int64;minimum=1"`
	// HMAC-SHA256 en hexadecimal. Ver la sección "Firma" de la descripción del documento.
	Signature string `json:"signature"`
}

// ServerDeletionRequested is published by admin when staff asks to delete a
// server. It is a request, not the fact: servers runs the same cascade it runs
// when the owner deletes the server and then publishes servers.server_deleted,
// which the other services already consume. It is irreversible, so servers must
// check the signature before touching anything. Same fields as ServerSuspended.
type ServerDeletionRequested struct {
	envelope.Meta
	// Servidor suspendido, reactivado o a eliminar (ID de discordia-servers).
	ServerID string `json:"server_id" contract:"format=uuid"`
	// moderation_version del servidor en admin. Crece con cada suspensión, reactivación o eliminación; ordena los eventos del mismo servidor.
	Version int64 `json:"version" contract:"format=int64;minimum=1"`
	// HMAC-SHA256 en hexadecimal. Ver la sección "Firma" de la descripción del documento.
	Signature string `json:"signature"`
}

// Sign sets Signature using secret. Admin calls it right before publishing.
func (e *ServerSuspended) Sign(secret string) {
	e.Signature = sign(secret, RoutingKeyServerSuspended, e.Meta, e.ServerID, e.Version)
}

// Verify reports whether Signature is valid for secret. An empty secret or an
// empty/malformed signature never verifies.
func (e ServerSuspended) Verify(secret string) bool {
	return verify(secret, RoutingKeyServerSuspended, e.Meta, e.ServerID, e.Version, e.Signature)
}

// Sign sets Signature using secret. Admin calls it right before publishing.
func (e *ServerReactivated) Sign(secret string) {
	e.Signature = sign(secret, RoutingKeyServerReactivated, e.Meta, e.ServerID, e.Version)
}

// Verify reports whether Signature is valid for secret. An empty secret or an
// empty/malformed signature never verifies.
func (e ServerReactivated) Verify(secret string) bool {
	return verify(secret, RoutingKeyServerReactivated, e.Meta, e.ServerID, e.Version, e.Signature)
}

// Sign sets Signature using secret. Admin calls it right before publishing.
func (e *ServerDeletionRequested) Sign(secret string) {
	e.Signature = sign(secret, RoutingKeyServerDeletionRequested, e.Meta, e.ServerID, e.Version)
}

// Verify reports whether Signature is valid for secret. An empty secret or an
// empty/malformed signature never verifies.
func (e ServerDeletionRequested) Verify(secret string) bool {
	return verify(secret, RoutingKeyServerDeletionRequested, e.Meta, e.ServerID, e.Version, e.Signature)
}

// signature is HMAC-SHA256(secret, routing_key \n event_id \n subject_id \n
// version \n occurred_at), hex-encoded. The subject is the user_id of the
// account events and the server_id of the server events. The routing key is
// part of the signed text so a valid UserSuspended can't be replayed as a
// UserReactivated, nor an account event as a server event.
// occurred_at is formatted in UTC so it is the same after a JSON round trip.
func signature(secret, routingKey string, meta envelope.Meta, subjectID string, version int64) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.Join([]string{
		routingKey,
		meta.EventID,
		subjectID,
		strconv.FormatInt(version, 10),
		meta.OccurredAt.UTC().Format(time.RFC3339Nano),
	}, "\n")))
	return mac.Sum(nil)
}

func sign(secret, routingKey string, meta envelope.Meta, subjectID string, version int64) string {
	return hex.EncodeToString(signature(secret, routingKey, meta, subjectID, version))
}

func verify(secret, routingKey string, meta envelope.Meta, subjectID string, version int64, got string) bool {
	if secret == "" || got == "" {
		return false
	}

	gotBytes, err := hex.DecodeString(got)
	if err != nil {
		return false
	}
	return hmac.Equal(gotBytes, signature(secret, routingKey, meta, subjectID, version))
}

func signStaff(secret, routingKey string, meta envelope.Meta, userID string) string {
	return hex.EncodeToString(staffSignature(secret, routingKey, meta, userID))
}

func verifyStaff(secret, routingKey string, meta envelope.Meta, userID, got string) bool {
	if secret == "" || got == "" {
		return false
	}
	gotBytes, err := hex.DecodeString(got)
	if err != nil {
		return false
	}
	return hmac.Equal(gotBytes, staffSignature(secret, routingKey, meta, userID))
}

func staffSignature(secret, routingKey string, meta envelope.Meta, userID string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.Join([]string{
		routingKey, meta.EventID, userID, meta.OccurredAt.UTC().Format(time.RFC3339Nano),
	}, "\n")))
	return mac.Sum(nil)
}
