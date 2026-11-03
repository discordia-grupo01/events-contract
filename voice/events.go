package voice

import "github.com/discordia-grupo01/events-contract/envelope"

const (
	RoutingKeyParticipantJoined = "voice.participant_joined"
	RoutingKeyParticipantLeft   = "voice.participant_left"
	RoutingKeyPrivateCall       = "voice.private_call"
)

type PrivateCall struct {
	envelope.Meta
	EventType      string `json:"event_type"`
	CallID         string `json:"call_id"`
	ConversationID string `json:"conversation_id"`
	CallerID       string `json:"caller_id"`
	RecipientID    string `json:"recipient_id"`
	Status         string `json:"status"`
}

type ParticipantJoined struct {
	envelope.Meta
	EventType     string `json:"event_type"`
	ServerID      string `json:"server_id"`
	RoomID        string `json:"room_id"`
	ParticipantID string `json:"participant_id"`
	Name          string `json:"name"`
}

type ParticipantLeft struct {
	envelope.Meta
	EventType     string `json:"event_type"`
	ServerID      string `json:"server_id"`
	RoomID        string `json:"room_id"`
	ParticipantID string `json:"participant_id"`
	Name          string `json:"name"`
}
