package messages

import "github.com/discordia-grupo01/events-contract/envelope"

const RoutingKeyMessageCreated = "messaging.message_created"

type MessageCreated struct {
	envelope.Meta
	ServerID         string   `json:"server_id"`
	ChannelID        string   `json:"channel_id"`
	MessageID        string   `json:"message_id"`
	AuthorID         string   `json:"author_id"`
	MentionedUserIDs []string `json:"mentioned_user_ids"`
}
