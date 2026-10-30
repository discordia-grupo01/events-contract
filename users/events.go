package users

import (
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const (
	RoutingKeyProfileUpdated    = "users.profile_updated"
	RoutingKeyAccountRegistered = "users.account_registered"
	RoutingKeyEmailVerified     = "users.email_verified"
)

// AccountRegistered is published by identify once, when an account is created.
// It carries the data of the account itself (as opposed to the public profile
// of ProfileUpdated) for the one consumer that needs it: admin's back-office
// listing, which searches users by email. Email and CreatedAt never change, so
// there is no version: a repeated delivery is harmless because applying it
// again leaves the same values. Do not add this event to consumers that only
// need the public profile -- the email is private data.
//
// EmailVerified is true when the account was born verified (OAuth sign-up)
// and false for an email sign-up, which stays pending until EmailVerified
// arrives. Events published before this field existed don't carry it, so
// they decode as false.
type AccountRegistered struct {
	envelope.Meta
	UserID        string    `json:"user_id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
	EmailVerified bool      `json:"email_verified"`
}

// EmailVerified is published by identify once, when the user confirms the
// email of an account created with email sign-up (OAuth sign-ups never
// publish it: they are already verified in AccountRegistered). Like
// AccountRegistered it only exists for admin's back-office listing, which
// shows unverified accounts as pending. Verification is never undone, so
// there is no version: a repeated or out-of-order delivery is harmless, and
// it may arrive before AccountRegistered.
type EmailVerified struct {
	envelope.Meta
	UserID     string    `json:"user_id"`
	VerifiedAt time.Time `json:"verified_at"`
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
