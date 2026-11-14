package servers

import (
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const (
	RoutingKeyServerDeleted = "servers.server_deleted"
	RoutingKeyServerUpdated = "servers.server_updated"
	RoutingKeyEmojiCreated  = "servers.emoji_created"
	RoutingKeyEmojiDeleted  = "servers.emoji_deleted"

	RoutingKeyResourceChanged = "servers.resource_changed"
)

type ServerDeleted struct {
	envelope.Meta
	ServerID string `json:"server_id" contract:"format=uuid"`
	// Ids de todos los canales (texto y voz) que tenía el servidor. Puede ser vacío, nunca null.
	ChannelIDs []string `json:"channel_ids" contract:"itemFormat=uuid"`
}

// ServerUpdated is published by servers when a server is created and every
// time its name, icon or owner changes. It is a full snapshot, not a diff, so
// consumers don't need the previous state. There is no version: the consumer
// keeps the snapshot with the newest occurred_at of each server, so a repeated
// or out-of-order delivery doesn't overwrite newer data (same rule as the
// permissions events). Built for admin's back-office listing, which shows the
// name, icon, owner and creation date of every server without asking servers.
type ServerUpdated struct {
	envelope.Meta
	ServerID string `json:"server_id" contract:"format=uuid"`
	Name     string `json:"name"`
	// ID del dueño del servidor (emitido por discordia-identify).
	OwnerID string `json:"owner_id"`
	// Fecha de creación del servidor, ISO 8601 en UTC. No cambia entre fotos.
	CreatedAt time.Time `json:"created_at"`
	// Dirección del ícono tal como la guarda servers: URL pública de Supabase Storage o ruta del gateway (`/v1/servers/{id}/icon?v=...`). Vacío si no tiene ícono. Los eventos anteriores a la versión 0.22.0 no lo traen: el consumidor lo toma como vacío.
	IconURL string `json:"icon_url" contract:"optional=true"`
}

// EmojiCreated is published by servers when a custom emoji is added to a server.
// Like EmojiDeleted, it carries the whole emoji: an emoji never changes after it
// is created, so its only transition is being deleted, and that transition wins
// whatever order the two events arrive in.
type EmojiCreated struct {
	envelope.Meta
	EmojiID  string `json:"emoji_id" contract:"format=uuid"`
	ServerID string `json:"server_id" contract:"format=uuid"`
	Name     string `json:"name" contract:"pattern=^[A-Za-z0-9_]{2,32}$"`
	// true si la imagen es un GIF; usarlo requiere Premium.
	Animated bool   `json:"animated"`
	ImageURL string `json:"image_url" contract:"format=uri"`
}

// EmojiDeleted is published by servers when a custom emoji is deleted. It can no
// longer be used in new messages or reactions, but the image stays available:
// consumers keep it so the messages and reactions that already use it still show
// it.
type EmojiDeleted struct {
	envelope.Meta
	EmojiID  string `json:"emoji_id" contract:"format=uuid"`
	ServerID string `json:"server_id" contract:"format=uuid"`
	Name     string `json:"name" contract:"pattern=^[A-Za-z0-9_]{2,32}$"`
	// true si la imagen es un GIF; usarlo requiere Premium.
	Animated bool   `json:"animated"`
	ImageURL string `json:"image_url" contract:"format=uri"`
}

// ResourceChanged is published by servers when something a client keeps on screen changed and no other event says so:
// the server's name, icon, banner or vanity link, its channels and categories (rename, move, reorder, create), or its
// roles (reorder, default role). It carries no data, only what to ask for again, so a consumer never needs the new
// state and a repeated or out-of-order delivery is harmless.
type ResourceChanged struct {
	envelope.Meta
	// Servidor al que pertenece lo que cambió.
	ServerID string `json:"server_id" contract:"format=uuid"`
	// Qué volver a pedir: los datos del servidor, sus canales y categorías, o sus roles.
	Resource string `json:"resource" contract:"enum=server|channels|roles"`
}
