package moderation_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
	"github.com/discordia-grupo01/events-contract/moderation"
)

func TestMemberBannedJSONShape(t *testing.T) {
	event := moderation.MemberBanned{
		Meta: envelope.Meta{
			EventID:    "evt-1",
			OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ServerID: "server-1",
		UserID:   "user-1",
		BannedBy: "user-2",
	}

	assertJSONShape(t, event, map[string]string{
		"event_id":    "evt-1",
		"server_id":   "server-1",
		"user_id":     "user-1",
		"banned_by":   "user-2",
		"occurred_at": "2026-01-01T00:00:00Z",
	})
}

func TestMemberUnbannedJSONShape(t *testing.T) {
	event := moderation.MemberUnbanned{
		Meta: envelope.Meta{
			EventID:    "evt-2",
			OccurredAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		ServerID:   "server-1",
		UserID:     "user-1",
		UnbannedBy: "user-2",
	}

	assertJSONShape(t, event, map[string]string{
		"event_id":    "evt-2",
		"server_id":   "server-1",
		"user_id":     "user-1",
		"unbanned_by": "user-2",
		"occurred_at": "2026-01-02T00:00:00Z",
	})
}

func TestRoutingKeys(t *testing.T) {
	if moderation.RoutingKeyMemberBanned != "servers.member_banned" {
		t.Errorf("RoutingKeyMemberBanned = %q", moderation.RoutingKeyMemberBanned)
	}
	if moderation.RoutingKeyMemberUnbanned != "servers.member_unbanned" {
		t.Errorf("RoutingKeyMemberUnbanned = %q", moderation.RoutingKeyMemberUnbanned)
	}
}

// assertJSONShape checks that event marshals to exactly the keys in want --
// no missing field and no unexpected one (e.g. a ban reason leaking into the
// event).
func assertJSONShape(t *testing.T, event any, want map[string]string) {
	t.Helper()

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
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
