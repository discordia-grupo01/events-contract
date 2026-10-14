package users

import (
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const (
	RoutingKeyProfileUpdated    = "users.profile_updated"
	RoutingKeyAccountRegistered = "users.account_registered"
)

// AccountRegistered is published by identify once, when an account is created.
// It carries the data of the account itself (as opposed to the public profile
// of ProfileUpdated) for the one consumer that needs it: admin's back-office
// listing, which searches users by email. Email and CreatedAt never change, so
// there is no version: a repeated delivery is harmless because applying it
// again leaves the same values. Do not add this event to consumers that only
// need the public profile -- the email is private data.
type AccountRegistered struct {
	envelope.Meta
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ProfileUpdated is published by identify when a user is created and every
// time their public profile or status changes. It is a full snapshot, not a
// diff, so consumers don't need the previous state. Version is identify's
// profile_version: it grows with every change of the same user, and a
// consumer applies a snapshot only if its version is higher than the one it
// already has (duplicates and out-of-order deliveries are then harmless).
// Only public fields travel: the email never does.
type ProfileUpdated struct {
	envelope.Meta
	UserID      string `json:"user_id"`
	Version     int64  `json:"version"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Description string `json:"description"`
	StatusText  string `json:"status_text"`
	StatusEmoji string `json:"status_emoji"`
}
