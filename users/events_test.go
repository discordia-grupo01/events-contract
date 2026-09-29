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

func TestRoutingKeys(t *testing.T) {
	if users.RoutingKeyProfileUpdated != "users.profile_updated" {
		t.Errorf("RoutingKeyProfileUpdated = %q", users.RoutingKeyProfileUpdated)
	}
}
