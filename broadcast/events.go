package broadcast

import "time"

const (
	ExchangeName = "messaging.broadcast"
	ExchangeType = "fanout"
	KeyPrefix    = "broadcast."
)

const (
	EventNewMessage        = "new_message"
	EventMessageUpdated    = "message_updated"
	EventMessageDeleted    = "message_deleted"
	EventReactionUpdated   = "reaction_updated"
	EventNewDM             = "new_dm"
	EventMention           = "mention"
	EventAccessRevoked     = "access_revoked"
	EventServerLeft        = "server_left"
	EventServerSuspended   = "server_suspended"
	EventServerReactivated = "server_reactivated"
	EventDisconnect        = "disconnect"
	EventPresenceUpdated   = "presence_updated"
	EventOwnStatusUpdated  = "own_status_updated"
)

type Attachment map[string]any

// Message is the JSON of a message as clients receive it (Message.to_api in messaging).
type Message struct {
	ID              string       `json:"id"`
	ChannelID       string       `json:"channel_id"`
	ServerID        *string      `json:"server_id"`
	UserID          string       `json:"user_id"`
	Content         string       `json:"content"`
	InsertedAt      time.Time    `json:"inserted_at"`
	Edited          bool         `json:"edited"`
	EditedAt        *time.Time   `json:"edited_at"`
	Mentions        []string     `json:"mentions"`
	MentionRoles    []string     `json:"mention_roles"`
	MentionEveryone bool         `json:"mention_everyone"`
	Attachments     []Attachment `json:"attachments"`
	ClientID        string       `json:"client_id,omitempty"`
}

type DeletedMessage struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	ServerID  *string   `json:"server_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

type ReactionChange struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	Emoji     string `json:"emoji"`
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
	Count     int    `json:"count" contract:"minimum=0"`
}

type DirectMessage struct {
	ConversationID string  `json:"conversation_id"`
	PartnerID      string  `json:"partner_id"`
	Message        Message `json:"message"`
}

type MessageMention struct {
	ServerID  *string `json:"server_id"`
	ChannelID string  `json:"channel_id"`
	Message   Message `json:"message"`
}

type Revocation struct {
	UserID    string `json:"user_id"`
	Reason    string `json:"reason"`
	ServerID  string `json:"server_id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
}

type ServerRef struct {
	ServerID string `json:"server_id"`
}

type Presence struct {
	UserID  string `json:"user_id"`
	Status  string `json:"status"`
	Version int64  `json:"version" contract:"minimum=0"`
}

type OwnStatus struct {
	Status  string `json:"status"`
	Version int64  `json:"version" contract:"minimum=0"`
}

type Empty struct{}

// NewMessage is sent to every instance of messaging when a message is created, so each one pushes it to the sockets it holds.
type NewMessage struct {
	EventID string  `json:"event_id"`
	Topic   string  `json:"topic" contract:"pattern=^channel:.+"`
	Event   string  `json:"event" contract:"enum=new_message"`
	Payload Message `json:"payload"`
}

// MessageUpdated is sent when the text of a message is edited.
type MessageUpdated struct {
	EventID string  `json:"event_id"`
	Topic   string  `json:"topic" contract:"pattern=^channel:.+"`
	Event   string  `json:"event" contract:"enum=message_updated"`
	Payload Message `json:"payload"`
}

// MessageDeleted is sent when a message is deleted. It carries no content.
type MessageDeleted struct {
	EventID string         `json:"event_id"`
	Topic   string         `json:"topic" contract:"pattern=^channel:.+"`
	Event   string         `json:"event" contract:"enum=message_deleted"`
	Payload DeletedMessage `json:"payload"`
}

// ReactionUpdated is sent when somebody adds or removes a reaction.
type ReactionUpdated struct {
	EventID string         `json:"event_id"`
	Topic   string         `json:"topic" contract:"pattern=^channel:.+"`
	Event   string         `json:"event" contract:"enum=reaction_updated"`
	Payload ReactionChange `json:"payload"`
}

// NewDM is sent to each participant of a direct conversation when a message arrives.
type NewDM struct {
	EventID string        `json:"event_id"`
	Topic   string        `json:"topic" contract:"pattern=^user:.+"`
	Event   string        `json:"event" contract:"enum=new_dm"`
	Payload DirectMessage `json:"payload"`
}

// Mention is sent to every user a message mentions, wherever they are.
type Mention struct {
	EventID string         `json:"event_id"`
	Topic   string         `json:"topic" contract:"pattern=^user:.+"`
	Event   string         `json:"event" contract:"enum=mention"`
	Payload MessageMention `json:"payload"`
}

// AccessRevoked is sent when a user loses access to a server or a channel, so their open sockets leave it.
type AccessRevoked struct {
	EventID string     `json:"event_id"`
	Topic   string     `json:"topic" contract:"pattern=^(channel|presence):.+"`
	Event   string     `json:"event" contract:"enum=access_revoked"`
	Payload Revocation `json:"payload"`
}

// ServerLeft is sent to a user who stopped being a member, so the client drops the server from its list.
type ServerLeft struct {
	EventID string    `json:"event_id"`
	Topic   string    `json:"topic" contract:"pattern=^user:.+"`
	Event   string    `json:"event" contract:"enum=server_left"`
	Payload ServerRef `json:"payload"`
}

// ServerSuspended is sent to every member when staff suspends the server.
type ServerSuspended struct {
	EventID string    `json:"event_id"`
	Topic   string    `json:"topic" contract:"pattern=^user:.+"`
	Event   string    `json:"event" contract:"enum=server_suspended"`
	Payload ServerRef `json:"payload"`
}

// ServerReactivated is sent to every member when staff lifts the suspension.
type ServerReactivated struct {
	EventID string    `json:"event_id"`
	Topic   string    `json:"topic" contract:"pattern=^user:.+"`
	Event   string    `json:"event" contract:"enum=server_reactivated"`
	Payload ServerRef `json:"payload"`
}

// Disconnect makes every instance close the sockets of a user.
type Disconnect struct {
	EventID string `json:"event_id"`
	Topic   string `json:"topic" contract:"pattern=^user_socket:.+"`
	Event   string `json:"event" contract:"enum=disconnect"`
	Payload Empty  `json:"payload"`
}

// PresenceUpdated is sent to whoever should see the connection status of a user.
type PresenceUpdated struct {
	EventID string   `json:"event_id"`
	Topic   string   `json:"topic" contract:"pattern=^(presence|user):.+"`
	Event   string   `json:"event" contract:"enum=presence_updated"`
	Payload Presence `json:"payload"`
}

// OwnStatusUpdated is sent to a user's own sockets when they change the status they chose.
type OwnStatusUpdated struct {
	EventID string    `json:"event_id"`
	Topic   string    `json:"topic" contract:"pattern=^user:.+"`
	Event   string    `json:"event" contract:"enum=own_status_updated"`
	Payload OwnStatus `json:"payload"`
}
