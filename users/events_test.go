package users_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
	"github.com/discordia-grupo01/events-contract/users"
)

// ProfileUpdated marshals to exactly these keys: no missing field and no
// unexpected one (e.g. the email leaking into the event).
func TestProfileUpdatedJSONShape(t *testing.T) {
	event := users.ProfileUpdated{
		Meta: envelope.Meta{
			EventID:    "evt-1",
			OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		UserID:      "user-1",
		Version:     3,
		Name:        "Manu",
		AvatarURL:   "https://cdn/a.png",
		Description: "hola",
		StatusText:  "jugando",
		StatusEmoji: "🎮",
	}

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"event_id":     "evt-1",
		"occurred_at":  "2026-01-01T00:00:00Z",
		"user_id":      "user-1",
		"version":      float64(3),
		"name":         "Manu",
		"avatar_url":   "https://cdn/a.png",
		"description":  "hola",
		"status_text":  "jugando",
		"status_emoji": "🎮",
	}
	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Errorf("missing key %q in %s", key, body)
			continue
		}
		if gotValue != wantValue {
			t.Errorf("key %q = %v, want %v", key, gotValue, wantValue)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d keys, want %d -- unexpected field in %s", len(got), len(want), body)
	}
}

// AccountRegistered marshals to exactly these keys: no missing field and no
// unexpected one (e.g. the password hash leaking into the event).
func TestAccountRegisteredJSONShape(t *testing.T) {
	event := users.AccountRegistered{
		Meta: envelope.Meta{
			EventID:    "evt-2",
			OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		UserID:        "user-1",
		Email:         "manu@example.com",
		Name:          "Manu",
		CreatedAt:     time.Date(2025, 12, 31, 23, 0, 0, 0, time.UTC),
		EmailVerified: true,
	}

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"event_id":       "evt-2",
		"occurred_at":    "2026-01-01T00:00:00Z",
		"user_id":        "user-1",
		"email":          "manu@example.com",
		"name":           "Manu",
		"created_at":     "2025-12-31T23:00:00Z",
		"email_verified": true,
	}
	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Errorf("missing key %q in %s", key, body)
			continue
		}
		if gotValue != wantValue {
			t.Errorf("key %q = %v, want %v", key, gotValue, wantValue)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d keys, want %d -- unexpected field in %s", len(got), len(want), body)
	}
}

func TestAccountRegisteredRoundTrip(t *testing.T) {
	in := users.AccountRegistered{
		Meta:      envelope.Meta{EventID: "evt-3", OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		UserID:    "user-1",
		Email:     "manu@example.com",
		Name:      "Manu",
		CreatedAt: time.Date(2025, 12, 31, 23, 0, 0, 0, time.UTC),
	}
	body, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out users.AccountRegistered
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("round trip = %+v, want %+v", out, in)
	}
}

// An account_registered published before email_verified existed decodes as
// not verified.
func TestAccountRegisteredWithoutEmailVerified(t *testing.T) {
	body := []byte(`{"event_id":"evt-4","occurred_at":"2026-01-01T00:00:00Z","user_id":"user-1","email":"manu@example.com","name":"Manu","created_at":"2025-12-31T23:00:00Z"}`)
	var out users.AccountRegistered
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	if out.EmailVerified {
		t.Errorf("EmailVerified = true, want false")
	}
}

// EmailVerified marshals to exactly these keys.
func TestEmailVerifiedJSONShape(t *testing.T) {
	event := users.EmailVerified{
		Meta: envelope.Meta{
			EventID:    "evt-5",
			OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		UserID:     "user-1",
		VerifiedAt: time.Date(2025, 12, 31, 23, 30, 0, 0, time.UTC),
	}

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"event_id":    "evt-5",
		"occurred_at": "2026-01-01T00:00:00Z",
		"user_id":     "user-1",
		"verified_at": "2025-12-31T23:30:00Z",
	}
	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Errorf("missing key %q in %s", key, body)
			continue
		}
		if gotValue != wantValue {
			t.Errorf("key %q = %v, want %v", key, gotValue, wantValue)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d keys, want %d -- unexpected field in %s", len(got), len(want), body)
	}
}

func TestRoutingKeys(t *testing.T) {
	if users.RoutingKeyProfileUpdated != "users.profile_updated" {
		t.Errorf("RoutingKeyProfileUpdated = %q", users.RoutingKeyProfileUpdated)
	}
	if users.RoutingKeyAccountRegistered != "users.account_registered" {
		t.Errorf("RoutingKeyAccountRegistered = %q", users.RoutingKeyAccountRegistered)
	}
	if users.RoutingKeyEmailVerified != "users.email_verified" {
		t.Errorf("RoutingKeyEmailVerified = %q", users.RoutingKeyEmailVerified)
	}
}
