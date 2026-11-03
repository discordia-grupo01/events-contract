package presence_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
	"github.com/discordia-grupo01/events-contract/presence"
)

func TestRoutingKey(t *testing.T) {
	if presence.RoutingKeyUserPresenceChanged != "messaging.user_presence_changed" {
		t.Fatalf("routing key = %q", presence.RoutingKeyUserPresenceChanged)
	}
}

func TestConnectionStates(t *testing.T) {
	for got, want := range map[string]string{
		presence.ConnectionsActive: "active",
		presence.ConnectionsIdle:   "idle",
		presence.ConnectionsNone:   "none",
	} {
		if got != want {
			t.Errorf("connection state = %q, want %q", got, want)
		}
	}
}

func TestUserPresenceChangedJSONShape(t *testing.T) {
	event := presence.UserPresenceChanged{
		Meta:         envelope.Meta{EventID: "evt-1", OccurredAt: time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC)},
		UserID:       "user-1",
		Connections:  presence.ConnectionsActive,
		DoNotDisturb: true,
		Version:      7,
	}

	got := marshalToMap(t, event)

	if len(got) != 6 {
		t.Fatalf("got %d fields, want 6 in %v", len(got), got)
	}
	for key, want := range map[string]any{
		"event_id":       "evt-1",
		"occurred_at":    "2026-11-03T15:00:00Z",
		"user_id":        "user-1",
		"connections":    "active",
		"do_not_disturb": true,
		"version":        float64(7),
	} {
		if got[key] != want {
			t.Errorf("key %q = %v, want %v", key, got[key], want)
		}
	}
}

func TestUserPresenceChangedAlwaysSendsDoNotDisturb(t *testing.T) {
	got := marshalToMap(t, presence.UserPresenceChanged{
		Meta:        envelope.Meta{EventID: "evt-2", OccurredAt: time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC)},
		UserID:      "user-1",
		Connections: presence.ConnectionsNone,
		Version:     1,
	})

	value, present := got["do_not_disturb"]
	if !present || value != false {
		t.Errorf("do_not_disturb = %v (present: %v), want an explicit false", value, present)
	}
}

func TestUserPresenceChangedRoundTrip(t *testing.T) {
	want := presence.UserPresenceChanged{
		Meta:         envelope.Meta{EventID: "evt-3", OccurredAt: time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC)},
		UserID:       "user-2",
		Connections:  presence.ConnectionsIdle,
		DoNotDisturb: false,
		Version:      42,
	}

	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got presence.UserPresenceChanged
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.EventID != want.EventID || !got.OccurredAt.Equal(want.OccurredAt) || got.UserID != want.UserID ||
		got.Connections != want.Connections || got.DoNotDisturb != want.DoNotDisturb || got.Version != want.Version {
		t.Errorf("round trip = %+v, want %+v", got, want)
	}
}

func marshalToMap(t *testing.T, event any) map[string]any {
	t.Helper()

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return got
}
