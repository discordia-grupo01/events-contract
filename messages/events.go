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
	ServerID  string `json:"server_id" contract:"format=uuid"`
	ChannelID string `json:"channel_id" contract:"format=uuid"`
	// UUID v7 del mensaje, ordenable por tiempo de creación.
	MessageID string `json:"message_id" contract:"format=uuid"`
	AuthorID  string `json:"author_id" contract:"format=uuid"`
	// Usuarios mencionados y ya validados por `messaging`, sin el autor. Puede ser una lista vacía.
	MentionedUserIDs []string `json:"mentioned_user_ids" contract:"itemFormat=uuid"`
	// Nombre del canal. Opcional; solo para armar el aviso push.
	ChannelName string `json:"channel_name,omitempty"`
	// Extracto del texto (unos 140 caracteres), opcional. Solo para armar el aviso push; el consumidor no debe guardarlo.
	ContentPreview string `json:"content_preview,omitempty" contract:"maxLength=160"`
}

type DirectMessageCreated struct {
	envelope.Meta
	// Id de la conversación (determinístico a partir de los dos participantes).
	ConversationID string `json:"conversation_id"`
	// UUID v7 del mensaje, ordenable por tiempo de creación.
	MessageID string `json:"message_id" contract:"format=uuid"`
	// Quien escribió el mensaje.
	AuthorID string `json:"author_id"`
	// El otro participante de la conversación.
	RecipientID string `json:"recipient_id"`
	// Extracto del texto (unos 140 caracteres), opcional. Solo para armar el aviso push; el consumidor no debe guardarlo.
	ContentPreview string `json:"content_preview,omitempty" contract:"maxLength=160"`
}

// MessageUpdated is published when the text of a message in a server channel
// is edited. It carries the mentions the message has after the edit, so a
// consumer can add the notification of whoever the edit mentioned and remove
// the one of whoever it stopped mentioning. OccurredAt is the moment of the
// edit. Direct messages never publish it: they have no mentions.
type MessageUpdated struct {
	envelope.Meta
	ServerID  string `json:"server_id" contract:"format=uuid"`
	ChannelID string `json:"channel_id" contract:"format=uuid"`
	MessageID string `json:"message_id" contract:"format=uuid"`
	AuthorID  string `json:"author_id" contract:"format=uuid"`
	// Menciones que el mensaje tiene después de la edición, ya validadas y sin el autor. Puede ser una lista vacía.
	MentionedUserIDs []string `json:"mentioned_user_ids" contract:"itemFormat=uuid"`
	// Nombre del canal. Opcional; solo para armar el aviso push.
	ChannelName string `json:"channel_name,omitempty"`
	// Extracto del texto editado (unos 140 caracteres), opcional. Solo para armar el aviso push; el consumidor no debe guardarlo.
	ContentPreview string `json:"content_preview,omitempty" contract:"maxLength=160"`
}

// MessageDeleted is published when a message is deleted, in a server channel
// or in a direct conversation (ServerID is empty then and ChannelID is the
// conversation). OccurredAt is the moment of the deletion. A consumer drops
// whatever it kept about MessageID.
type MessageDeleted struct {
	envelope.Meta
	// Servidor del canal. No viaja si el mensaje era directo.
	ServerID string `json:"server_id,omitempty" contract:"format=uuid"`
	// Canal del mensaje, o la conversación si era un mensaje directo.
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id" contract:"format=uuid"`
	AuthorID  string `json:"author_id" contract:"format=uuid"`
}
