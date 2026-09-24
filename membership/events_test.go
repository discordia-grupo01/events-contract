package membership_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
	"github.com/discordia-grupo01/events-contract/membership"
)

func TestMemberJoinedJSONShape(t *testing.T) {
	event := membership.MemberJoined{
		Meta: envelope.Meta{
			EventID:    "evt-1",
			OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ServerID: "server-1",
		UserID:   "user-1",
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
		"user_id":     "user-1",
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
