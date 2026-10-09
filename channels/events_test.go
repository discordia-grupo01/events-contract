package channels_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/channels"
	"github.com/discordia-grupo01/events-contract/envelope"
)

func TestChannelCreatedJSONShape(t *testing.T) {
	event := channels.ChannelCreated{
		Meta: envelope.Meta{
			EventID:    "evt-1",
			OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ServerID:  "server-1",
		ChannelID: "channel-1",
		Name:      "general",
		Kind:      "text",
	}

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]string{
		"event_id":    "evt-1",
		"server_id":   "server-1",
		"channel_id":  "channel-1",
		"name":        "general",
		"kind":        "text",
		"occurred_at": "2026-01-01T00:00:00Z",
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

func TestChannelDeletedJSONShape(t *testing.T) {
	event := channels.ChannelDeleted{
		Meta: envelope.Meta{
			EventID:    "evt-2",
			OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ServerID:  "server-1",
		ChannelID: "channel-1",
	}

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]string{
		"event_id":    "evt-2",
		"server_id":   "server-1",
		"channel_id":  "channel-1",
		"occurred_at": "2026-01-01T00:00:00Z",
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

func TestChannelAccessRevokedJSONShape(t *testing.T) {
	event := channels.ChannelAccessRevoked{
		Meta: envelope.Meta{
			EventID:    "evt-3",
			OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ServerID:  "server-1",
		ChannelID: "channel-1",
		UserIDs:   []string{"user-1", "user-2"},
	}

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]string{
		"event_id":    "evt-3",
		"server_id":   "server-1",
		"channel_id":  "channel-1",
		"occurred_at": "2026-01-01T00:00:00Z",
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

	userIDs, ok := got["user_ids"].([]any)
	if !ok || len(userIDs) != 2 || userIDs[0] != "user-1" || userIDs[1] != "user-2" {
		t.Errorf("user_ids = %v, want [user-1 user-2]", got["user_ids"])
	}
	if len(got) != len(want)+1 {
		t.Errorf("got %d keys, want %d -- unexpected field in %s", len(got), len(want)+1, body)
	}
}
