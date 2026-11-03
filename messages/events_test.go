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

func TestDirectMessageRoutingKey(t *testing.T) {
	if messages.RoutingKeyDirectMessageCreated != "messaging.direct_message_created" {
		t.Fatalf("routing key = %q", messages.RoutingKeyDirectMessageCreated)
	}
}

func TestDirectMessageCreatedJSONShape(t *testing.T) {
	event := messages.DirectMessageCreated{
		Meta: envelope.Meta{
			EventID:    "evt-dm-1",
			OccurredAt: time.Date(2026, 10, 31, 15, 0, 0, 0, time.UTC),
		},
		ConversationID: "conversation-1",
		MessageID:      "message-1",
		AuthorID:       "user-1",
		RecipientID:    "user-2",
	}

	got := marshalToMap(t, event)

	if len(got) != 6 {
		t.Fatalf("got %d fields, want 6 in %v", len(got), got)
	}
	for key, want := range map[string]string{
		"event_id":        "evt-dm-1",
		"occurred_at":     "2026-10-31T15:00:00Z",
		"conversation_id": "conversation-1",
		"message_id":      "message-1",
		"author_id":       "user-1",
		"recipient_id":    "user-2",
	} {
		if got[key] != want {
			t.Errorf("key %q = %v, want %v", key, got[key], want)
		}
	}
}

func TestDirectMessageCreatedRoundTrip(t *testing.T) {
	want := messages.DirectMessageCreated{
		Meta:           envelope.Meta{EventID: "evt-dm-2", OccurredAt: time.Date(2026, 10, 31, 15, 0, 0, 0, time.UTC)},
		ConversationID: "conversation-1",
		MessageID:      "message-2",
		AuthorID:       "user-1",
		RecipientID:    "user-2",
	}

	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got messages.DirectMessageCreated
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.EventID != want.EventID || !got.OccurredAt.Equal(want.OccurredAt) || got.ConversationID != want.ConversationID ||
		got.MessageID != want.MessageID || got.AuthorID != want.AuthorID || got.RecipientID != want.RecipientID {
		t.Errorf("round trip = %+v, want %+v", got, want)
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

func TestMessageCreatedWithPreviewJSONShape(t *testing.T) {
	event := messages.MessageCreated{
		Meta:             envelope.Meta{EventID: "evt-4", OccurredAt: time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC)},
		ServerID:         "server-1",
		ChannelID:        "channel-1",
		MessageID:        "message-4",
		AuthorID:         "user-1",
		MentionedUserIDs: []string{},
		ChannelName:      "general",
		ContentPreview:   "hola a todos",
	}

	got := marshalToMap(t, event)

	if len(got) != 9 {
		t.Fatalf("got %d fields, want 9 in %v", len(got), got)
	}
	if got["channel_name"] != "general" || got["content_preview"] != "hola a todos" {
		t.Errorf("channel_name = %v, content_preview = %v", got["channel_name"], got["content_preview"])
	}
}

func TestMessageCreatedOmitsThePreviewWhenItIsEmpty(t *testing.T) {
	got := marshalToMap(t, messages.MessageCreated{
		Meta:             envelope.Meta{EventID: "evt-5", OccurredAt: time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC)},
		MentionedUserIDs: []string{},
	})

	for _, key := range []string{"channel_name", "content_preview"} {
		if _, present := got[key]; present {
			t.Errorf("key %q should be omitted when empty, got %v", key, got)
		}
	}
}

func TestMessageCreatedFromAnOldPayloadHasNoPreview(t *testing.T) {
	body := `{"event_id":"evt-6","occurred_at":"2026-10-24T15:00:00Z","server_id":"server-1","channel_id":"channel-1","message_id":"message-6","author_id":"user-1","mentioned_user_ids":["user-2"]}`

	var got messages.MessageCreated
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.ChannelName != "" || got.ContentPreview != "" || got.MessageID != "message-6" {
		t.Errorf("old payload = %+v", got)
	}
}

func TestMessageCreatedPreviewRoundTrip(t *testing.T) {
	want := messages.MessageCreated{
		Meta:             envelope.Meta{EventID: "evt-7", OccurredAt: time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC)},
		ServerID:         "server-1",
		ChannelID:        "channel-1",
		MessageID:        "message-7",
		AuthorID:         "user-1",
		MentionedUserIDs: []string{"user-2"},
		ChannelName:      "anuncios",
		ContentPreview:   "mirá esto @usuario",
	}

	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got messages.MessageCreated
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.ChannelName != want.ChannelName || got.ContentPreview != want.ContentPreview {
		t.Errorf("round trip = %+v, want %+v", got, want)
	}
}

func TestDirectMessageCreatedWithPreviewJSONShape(t *testing.T) {
	event := messages.DirectMessageCreated{
		Meta:           envelope.Meta{EventID: "evt-dm-3", OccurredAt: time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC)},
		ConversationID: "conversation-1",
		MessageID:      "message-3",
		AuthorID:       "user-1",
		RecipientID:    "user-2",
		ContentPreview: "¿venís?",
	}

	got := marshalToMap(t, event)

	if len(got) != 7 {
		t.Fatalf("got %d fields, want 7 in %v", len(got), got)
	}
	if got["content_preview"] != "¿venís?" {
		t.Errorf("content_preview = %v", got["content_preview"])
	}
}

func TestDirectMessageCreatedOmitsThePreviewWhenItIsEmpty(t *testing.T) {
	got := marshalToMap(t, messages.DirectMessageCreated{
		Meta: envelope.Meta{EventID: "evt-dm-4", OccurredAt: time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC)},
	})

	if _, present := got["content_preview"]; present {
		t.Errorf("content_preview should be omitted when empty, got %v", got)
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

func TestMessageUpdatedAndDeletedRoutingKeys(t *testing.T) {
	if messages.RoutingKeyMessageUpdated != "messaging.message_updated" {
		t.Fatalf("routing key = %q", messages.RoutingKeyMessageUpdated)
	}
	if messages.RoutingKeyMessageDeleted != "messaging.message_deleted" {
		t.Fatalf("routing key = %q", messages.RoutingKeyMessageDeleted)
	}
}

func TestMessageUpdatedJSONShape(t *testing.T) {
	event := messages.MessageUpdated{
		Meta:             envelope.Meta{EventID: "evt-5", OccurredAt: time.Date(2026, 11, 3, 16, 0, 0, 0, time.UTC)},
		ServerID:         "server-1",
		ChannelID:        "channel-1",
		MessageID:        "message-5",
		AuthorID:         "user-1",
		MentionedUserIDs: []string{"user-2"},
		ChannelName:      "general",
		ContentPreview:   "hola @usuario",
	}

	got := marshalToMap(t, event)

	if len(got) != 9 {
		t.Fatalf("got %d fields, want 9 in %v", len(got), got)
	}
	if got["message_id"] != "message-5" || got["content_preview"] != "hola @usuario" {
		t.Errorf("message_id = %v, content_preview = %v", got["message_id"], got["content_preview"])
	}
}

func TestMessageUpdatedOmitsOptionalFieldsAndKeepsAnEmptyMentionList(t *testing.T) {
	event := messages.MessageUpdated{
		Meta:             envelope.Meta{EventID: "evt-6", OccurredAt: time.Date(2026, 11, 3, 16, 0, 0, 0, time.UTC)},
		ServerID:         "server-1",
		ChannelID:        "channel-1",
		MessageID:        "message-6",
		AuthorID:         "user-1",
		MentionedUserIDs: []string{},
	}

	got := marshalToMap(t, event)

	if _, ok := got["channel_name"]; ok {
		t.Errorf("channel_name should be omitted: %v", got)
	}
	if _, ok := got["content_preview"]; ok {
		t.Errorf("content_preview should be omitted: %v", got)
	}
	if list, ok := got["mentioned_user_ids"].([]any); !ok || len(list) != 0 {
		t.Errorf("mentioned_user_ids = %v, want an empty list", got["mentioned_user_ids"])
	}
}

func TestMessageUpdatedRoundTrip(t *testing.T) {
	want := messages.MessageUpdated{
		Meta:             envelope.Meta{EventID: "evt-7", OccurredAt: time.Date(2026, 11, 3, 16, 0, 0, 0, time.UTC)},
		ServerID:         "server-1",
		ChannelID:        "channel-1",
		MessageID:        "message-7",
		AuthorID:         "user-1",
		MentionedUserIDs: []string{"user-2", "user-3"},
		ChannelName:      "general",
		ContentPreview:   "hola",
	}

	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got messages.MessageUpdated
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}

	if got.MessageID != want.MessageID || len(got.MentionedUserIDs) != 2 || got.ContentPreview != "hola" {
		t.Errorf("round trip = %+v", got)
	}
}

func TestMessageDeletedJSONShape(t *testing.T) {
	event := messages.MessageDeleted{
		Meta:      envelope.Meta{EventID: "evt-8", OccurredAt: time.Date(2026, 11, 3, 17, 0, 0, 0, time.UTC)},
		ServerID:  "server-1",
		ChannelID: "channel-1",
		MessageID: "message-8",
		AuthorID:  "user-1",
	}

	got := marshalToMap(t, event)

	if len(got) != 6 {
		t.Fatalf("got %d fields, want 6 in %v", len(got), got)
	}
	if got["message_id"] != "message-8" || got["server_id"] != "server-1" {
		t.Errorf("message_id = %v, server_id = %v", got["message_id"], got["server_id"])
	}
}

func TestMessageDeletedOfADirectMessageHasNoServer(t *testing.T) {
	event := messages.MessageDeleted{
		Meta:      envelope.Meta{EventID: "evt-9", OccurredAt: time.Date(2026, 11, 3, 17, 0, 0, 0, time.UTC)},
		ChannelID: "conversation-1",
		MessageID: "message-9",
		AuthorID:  "user-1",
	}

	got := marshalToMap(t, event)

	if _, ok := got["server_id"]; ok {
		t.Errorf("server_id should be omitted: %v", got)
	}
	if len(got) != 5 {
		t.Fatalf("got %d fields, want 5 in %v", len(got), got)
	}
}
