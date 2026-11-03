package messages

import "github.com/discordia-grupo01/events-contract/envelope"

const (
	RoutingKeyMessageCreated       = "messaging.message_created"
	RoutingKeyDirectMessageCreated = "messaging.direct_message_created"
	RoutingKeyMessageUpdated       = "messaging.message_updated"
	RoutingKeyMessageDeleted       = "messaging.message_deleted"
)

type MessageCreated struct {
	envelope.Meta
	ServerID         string   `json:"server_id"`
	ChannelID        string   `json:"channel_id"`
	MessageID        string   `json:"message_id"`
	AuthorID         string   `json:"author_id"`
	MentionedUserIDs []string `json:"mentioned_user_ids"`
	ChannelName      string   `json:"channel_name,omitempty"`
	ContentPreview   string   `json:"content_preview,omitempty"`
}

type DirectMessageCreated struct {
	envelope.Meta
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	AuthorID       string `json:"author_id"`
	RecipientID    string `json:"recipient_id"`
	ContentPreview string `json:"content_preview,omitempty"`
}

// MessageUpdated is published when the text of a message in a server channel
// is edited. It carries the mentions the message has after the edit, so a
// consumer can add the notification of whoever the edit mentioned and remove
// the one of whoever it stopped mentioning. OccurredAt is the moment of the
// edit. Direct messages never publish it: they have no mentions.
type MessageUpdated struct {
	envelope.Meta
	ServerID         string   `json:"server_id"`
	ChannelID        string   `json:"channel_id"`
	MessageID        string   `json:"message_id"`
	AuthorID         string   `json:"author_id"`
	MentionedUserIDs []string `json:"mentioned_user_ids"`
	ChannelName      string   `json:"channel_name,omitempty"`
	ContentPreview   string   `json:"content_preview,omitempty"`
}

// MessageDeleted is published when a message is deleted, in a server channel
// or in a direct conversation (ServerID is empty then and ChannelID is the
// conversation). OccurredAt is the moment of the deletion. A consumer drops
// whatever it kept about MessageID.
type MessageDeleted struct {
	envelope.Meta
	ServerID  string `json:"server_id,omitempty"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	AuthorID  string `json:"author_id"`
}
