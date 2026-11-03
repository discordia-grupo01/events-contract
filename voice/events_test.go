package voice_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
	"github.com/discordia-grupo01/events-contract/voice"
)

func TestRoutingKeys(t *testing.T) {
	if voice.RoutingKeyParticipantJoined != "voice.participant_joined" {
		t.Fatalf("joined routing key = %q", voice.RoutingKeyParticipantJoined)
	}
	if voice.RoutingKeyParticipantLeft != "voice.participant_left" {
		t.Fatalf("left routing key = %q", voice.RoutingKeyParticipantLeft)
	}
	if voice.RoutingKeyPrivateCall != "voice.private_call" {
		t.Fatalf("private call routing key = %q", voice.RoutingKeyPrivateCall)
	}
}

func TestPrivateCallJSONShape(t *testing.T) {
	event := voice.PrivateCall{
		Meta: envelope.Meta{
			EventID:    "evt-call",
			OccurredAt: time.Date(2026, 11, 3, 19, 26, 0, 0, time.UTC),
		},
		EventType:      voice.RoutingKeyPrivateCall,
		CallID:         "call-1",
		ConversationID: "conversation-1",
		CallerID:       "caller-1",
		RecipientID:    "recipient-1",
		Status:         "ringing",
	}

	assertJSONShape(t, event, map[string]any{
		"event_id":        "evt-call",
		"event_type":      "voice.private_call",
		"occurred_at":     "2026-11-03T19:26:00Z",
		"call_id":         "call-1",
		"conversation_id": "conversation-1",
		"caller_id":       "caller-1",
		"recipient_id":    "recipient-1",
		"status":          "ringing",
	})
}

func TestParticipantJoinedJSONShape(t *testing.T) {
	event := voice.ParticipantJoined{
		Meta: envelope.Meta{
			EventID:    "evt-joined",
			OccurredAt: time.Date(2026, 10, 11, 20, 30, 0, 0, time.UTC),
		},
		EventType:     voice.RoutingKeyParticipantJoined,
		ServerID:      "server-1",
		RoomID:        "channel-1",
		ParticipantID: "user-1",
		Name:          "Ada",
	}

	assertJSONShape(t, event, map[string]any{
		"event_id":       "evt-joined",
		"event_type":     "voice.participant_joined",
		"occurred_at":    "2026-10-11T20:30:00Z",
		"server_id":      "server-1",
		"room_id":        "channel-1",
		"participant_id": "user-1",
		"name":           "Ada",
	})
}

func TestParticipantLeftJSONShape(t *testing.T) {
	event := voice.ParticipantLeft{
		Meta: envelope.Meta{
			EventID:    "evt-left",
			OccurredAt: time.Date(2026, 10, 11, 20, 35, 0, 0, time.UTC),
		},
		EventType:     voice.RoutingKeyParticipantLeft,
		ServerID:      "server-1",
		RoomID:        "channel-1",
		ParticipantID: "user-1",
		Name:          "Ada",
	}

	assertJSONShape(t, event, map[string]any{
		"event_id":       "evt-left",
		"event_type":     "voice.participant_left",
		"occurred_at":    "2026-10-11T20:35:00Z",
		"server_id":      "server-1",
		"room_id":        "channel-1",
		"participant_id": "user-1",
		"name":           "Ada",
	})
}

func assertJSONShape(t *testing.T, event any, want map[string]any) {
	t.Helper()

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d fields, want %d in %s", len(got), len(want), body)
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
}
