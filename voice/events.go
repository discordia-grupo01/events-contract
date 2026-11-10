package voice

import "github.com/discordia-grupo01/events-contract/envelope"

const (
	RoutingKeyParticipantJoined = "voice.participant_joined"
	RoutingKeyParticipantLeft   = "voice.participant_left"
	RoutingKeyPrivateCall       = "voice.private_call"
)

type PrivateCall struct {
	envelope.Meta
	EventType      string `json:"event_type" contract:"enum=voice.private_call"`
	CallID         string `json:"call_id"`
	ConversationID string `json:"conversation_id" contract:"format=uuid"`
	CallerID       string `json:"caller_id"`
	RecipientID    string `json:"recipient_id"`
	RoomName       string `json:"room_name,omitempty"`
	Status         string `json:"status" contract:"enum=ringing|active|rejected|missed|ended"`
	Reason         string `json:"reason,omitempty"`
	EndedBy        string `json:"ended_by,omitempty"`
}

type ParticipantJoined struct {
	envelope.Meta
	EventType     string `json:"event_type" contract:"enum=voice.participant_joined|voice.participant_left"`
	ServerID      string `json:"server_id" contract:"format=uuid"`
	RoomID        string `json:"room_id" contract:"format=uuid"`
	ParticipantID string `json:"participant_id"`
	Name          string `json:"name"`
}

type ParticipantLeft struct {
	envelope.Meta
	EventType     string `json:"event_type" contract:"enum=voice.participant_joined|voice.participant_left"`
	ServerID      string `json:"server_id" contract:"format=uuid"`
	RoomID        string `json:"room_id" contract:"format=uuid"`
	ParticipantID string `json:"participant_id"`
	Name          string `json:"name"`
}
