package servers_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
	"github.com/discordia-grupo01/events-contract/servers"
)

func TestServerDeletedJSONShape(t *testing.T) {
	event := servers.ServerDeleted{
		Meta: envelope.Meta{
			EventID:    "evt-1",
			OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ServerID:   "server-1",
		ChannelIDs: []string{"channel-1", "channel-2"},
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
		"event_id":    "evt-1",
		"server_id":   "server-1",
		"channel_ids": []any{"channel-1", "channel-2"},
		"occurred_at": "2026-01-01T00:00:00Z",
	}
	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Errorf("missing key %q in %s", key, body)
			continue
		}
		if !reflect.DeepEqual(gotValue, wantValue) {
			t.Errorf("key %q = %v, want %v", key, gotValue, wantValue)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d keys, want %d -- unexpected field in %s", len(got), len(want), body)
	}
}

func TestServerDeletedWithoutChannelsEncodesAnEmptyList(t *testing.T) {
	body, err := json.Marshal(servers.ServerDeleted{ServerID: "server-1", ChannelIDs: []string{}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	ids, ok := got["channel_ids"].([]any)
	if !ok || len(ids) != 0 {
		t.Errorf("channel_ids = %v, want an empty JSON list in %s", got["channel_ids"], body)
	}
}

func TestRoutingKey(t *testing.T) {
	if servers.RoutingKeyServerDeleted != "servers.server_deleted" {
		t.Errorf("RoutingKeyServerDeleted = %q", servers.RoutingKeyServerDeleted)
	}
}
