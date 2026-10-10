package permissions_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
	"github.com/discordia-grupo01/events-contract/permissions"
)

var meta = envelope.Meta{
	EventID:    "evt-1",
	OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
}

func decode(t *testing.T, event any) map[string]any {
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

func TestRoleUpdatedJSONShape(t *testing.T) {
	got := decode(t, permissions.RoleUpdated{
		Meta:        meta,
		ServerID:    "server-1",
		RoleID:      "role-1",
		Permissions: []string{"SEND_MESSAGES", "VIEW_CHANNELS"},
		IsEveryone:  true,
	})

	want := map[string]any{
		"event_id":    "evt-1",
		"occurred_at": "2026-01-01T00:00:00Z",
		"server_id":   "server-1",
		"role_id":     "role-1",
		"permissions": []any{"SEND_MESSAGES", "VIEW_CHANNELS"},
		"is_everyone": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRoleDeletedJSONShape(t *testing.T) {
	got := decode(t, permissions.RoleDeleted{Meta: meta, ServerID: "server-1", RoleID: "role-1"})

	want := map[string]any{
		"event_id":    "evt-1",
		"occurred_at": "2026-01-01T00:00:00Z",
		"server_id":   "server-1",
		"role_id":     "role-1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMemberRolesUpdatedJSONShape(t *testing.T) {
	got := decode(t, permissions.MemberRolesUpdated{
		Meta:     meta,
		ServerID: "server-1",
		UserID:   "user-1",
		RoleIDs:  []string{"role-1", "role-2"},
	})

	want := map[string]any{
		"event_id":    "evt-1",
		"occurred_at": "2026-01-01T00:00:00Z",
		"server_id":   "server-1",
		"user_id":     "user-1",
		"role_ids":    []any{"role-1", "role-2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestChannelOverridesUpdatedJSONShape(t *testing.T) {
	got := decode(t, permissions.ChannelOverridesUpdated{
		Meta:      meta,
		ServerID:  "server-1",
		ChannelID: "channel-1",
		Overrides: []permissions.RoleOverride{
			{RoleID: "role-1", Allow: []string{"SEND_MESSAGES"}, Deny: []string{}},
			{RoleID: "role-2", Allow: []string{}, Deny: []string{"VIEW_CHANNELS"}},
		},
	})

	want := map[string]any{
		"event_id":    "evt-1",
		"occurred_at": "2026-01-01T00:00:00Z",
		"server_id":   "server-1",
		"channel_id":  "channel-1",
		"overrides": []any{
			map[string]any{"role_id": "role-1", "allow": []any{"SEND_MESSAGES"}, "deny": []any{}},
			map[string]any{"role_id": "role-2", "allow": []any{}, "deny": []any{"VIEW_CHANNELS"}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestOwnerUpdatedJSONShape(t *testing.T) {
	got := decode(t, permissions.OwnerUpdated{Meta: meta, ServerID: "server-1", OwnerID: "user-1"})

	want := map[string]any{
		"event_id":    "evt-1",
		"occurred_at": "2026-01-01T00:00:00Z",
		"server_id":   "server-1",
		"owner_id":    "user-1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
