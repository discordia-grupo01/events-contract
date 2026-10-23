package messages_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
	"github.com/discordia-grupo01/events-contract/messages"
)

func TestRoutingKey(t *testing.T) {
	if messages.RoutingKeyMessageCreated != "messaging.message_created" {
		t.Fatalf("routing key = %q", messages.RoutingKeyMessageCreated)
	}
}

func TestMessageCreatedJSONShape(t *testing.T) {
	event := messages.MessageCreated{
		Meta: envelope.Meta{
			EventID:    "evt-1",
			OccurredAt: time.Date(2026, 10, 24, 15, 0, 0, 0, time.UTC),
		},
		ServerID:         "server-1",
		ChannelID:        "channel-1",
		MessageID:        "message-1",
		AuthorID:         "user-1",
		MentionedUserIDs: []string{"user-2", "user-3"},
	}

	got := marshalToMap(t, event)

	if len(got) != 7 {
		t.Fatalf("got %d fields, want 7 in %v", len(got), got)
	}
	for key, want := range map[string]string{
		"event_id":    "evt-1",
		"occurred_at": "2026-10-24T15:00:00Z",
		"server_id":   "server-1",
		"channel_id":  "channel-1",
		"message_id":  "message-1",
		"author_id":   "user-1",
	} {
		if got[key] != want {
			t.Errorf("key %q = %v, want %v", key, got[key], want)
		}
	}

	mentioned, ok := got["mentioned_user_ids"].([]any)
	if !ok || len(mentioned) != 2 || mentioned[0] != "user-2" || mentioned[1] != "user-3" {
		t.Errorf("mentioned_user_ids = %v", got["mentioned_user_ids"])
	}
}

func TestMessageCreatedWithoutMentionsIsAnEmptyList(t *testing.T) {
	event := messages.MessageCreated{
		Meta:             envelope.Meta{EventID: "evt-2", OccurredAt: time.Date(2026, 10, 24, 15, 0, 0, 0, time.UTC)},
		ServerID:         "server-1",
		ChannelID:        "channel-1",
		MessageID:        "message-2",
		AuthorID:         "user-1",
		MentionedUserIDs: []string{},
	}

	mentioned, ok := marshalToMap(t, event)["mentioned_user_ids"].([]any)
	if !ok || len(mentioned) != 0 {
		t.Errorf("mentioned_user_ids = %v, want an empty list", mentioned)
	}
}

func TestMessageCreatedRoundTrip(t *testing.T) {
	want := messages.MessageCreated{
		Meta:             envelope.Meta{EventID: "evt-3", OccurredAt: time.Date(2026, 10, 24, 15, 0, 0, 0, time.UTC)},
		ServerID:         "server-1",
		ChannelID:        "channel-1",
		MessageID:        "message-3",
		AuthorID:         "user-1",
		MentionedUserIDs: []string{"user-2"},
	}

	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got messages.MessageCreated
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.EventID != want.EventID || !got.OccurredAt.Equal(want.OccurredAt) || got.MessageID != want.MessageID ||
		got.AuthorID != want.AuthorID || len(got.MentionedUserIDs) != 1 || got.MentionedUserIDs[0] != "user-2" {
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
