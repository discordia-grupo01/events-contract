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

func TestMemberKickedJSONShape(t *testing.T) {
	event := moderation.MemberKicked{
		Meta: envelope.Meta{
			EventID:    "evt-3",
			OccurredAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
		},
		ServerID: "server-1",
		UserID:   "user-1",
		KickedBy: "user-2",
	}

	assertJSONShape(t, event, map[string]string{
		"event_id":    "evt-3",
		"server_id":   "server-1",
		"user_id":     "user-1",
		"kicked_by":   "user-2",
		"occurred_at": "2026-01-03T00:00:00Z",
	})
}

func TestMemberMutedJSONShape(t *testing.T) {
	event := moderation.MemberMuted{
		Meta: envelope.Meta{
			EventID:    "evt-4",
			OccurredAt: time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC),
		},
		ServerID: "server-1",
		UserID:   "user-1",
		MutedBy:  "user-2",
		UntilAt:  time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
		Reason:   "spam",
	}

	assertJSONShape(t, event, map[string]string{
		"event_id":    "evt-4",
		"server_id":   "server-1",
		"user_id":     "user-1",
		"muted_by":    "user-2",
		"until_at":    "2026-01-05T00:00:00Z",
		"reason":      "spam",
		"occurred_at": "2026-01-04T00:00:00Z",
	})
}

func TestMemberUnmutedJSONShape(t *testing.T) {
	event := moderation.MemberUnmuted{
		Meta: envelope.Meta{
			EventID:    "evt-5",
			OccurredAt: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
		},
		ServerID:  "server-1",
		UserID:    "user-1",
		UnmutedBy: "user-2",
		Reason:    "expired",
	}

	assertJSONShape(t, event, map[string]string{
		"event_id":    "evt-5",
		"server_id":   "server-1",
		"user_id":     "user-1",
		"unmuted_by":  "user-2",
		"reason":      "expired",
		"occurred_at": "2026-01-05T00:00:00Z",
	})
}

func TestRoutingKeys(t *testing.T) {
	if moderation.RoutingKeyMemberBanned != "servers.member_banned" {
		t.Errorf("RoutingKeyMemberBanned = %q", moderation.RoutingKeyMemberBanned)
	}
	if moderation.RoutingKeyMemberUnbanned != "servers.member_unbanned" {
		t.Errorf("RoutingKeyMemberUnbanned = %q", moderation.RoutingKeyMemberUnbanned)
	}
	if moderation.RoutingKeyMemberKicked != "servers.member_kicked" {
		t.Errorf("RoutingKeyMemberKicked = %q", moderation.RoutingKeyMemberKicked)
	}
	if moderation.RoutingKeyMemberMuted != "servers.member_muted" {
		t.Errorf("RoutingKeyMemberMuted = %q", moderation.RoutingKeyMemberMuted)
	}
	if moderation.RoutingKeyMemberUnmuted != "servers.member_unmuted" {
		t.Errorf("RoutingKeyMemberUnmuted = %q", moderation.RoutingKeyMemberUnmuted)
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
