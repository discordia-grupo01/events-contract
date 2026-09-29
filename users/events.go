package users

import "github.com/discordia-grupo01/events-contract/envelope"

const RoutingKeyProfileUpdated = "users.profile_updated"

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
