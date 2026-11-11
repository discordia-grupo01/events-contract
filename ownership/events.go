package ownership

import (
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const RoutingKeyTransferUpdated = "servers.ownership_transfer_updated"

// TransferUpdated is published by servers every time an ownership transfer changes state: when the owner offers it
// (pending), when the recipient accepts or rejects it, when the owner cancels it and when it expires after 24 hours.
// It is a full snapshot, so a consumer keeps the one with the newest occurred_at of each transfer and a repeated or
// out-of-order delivery doesn't overwrite newer data. OccurredAt is the moment of the change: the creation time for
// pending and the resolution time for the rest. The notification goes to the recipient while the transfer is pending
// and to the previous owner when the recipient answers or the offer expires.
type TransferUpdated struct {
	envelope.Meta
	// Identificador de la transferencia.
	TransferID string `json:"transfer_id" contract:"format=uuid"`
	// Servidor cuya propiedad se ofrece.
	ServerID string `json:"server_id" contract:"format=uuid"`
	// Dueño que hace el ofrecimiento.
	FromUserID string `json:"from_user_id"`
	// Miembro que recibe el ofrecimiento.
	ToUserID string `json:"to_user_id"`
	// Estado de la transferencia después del cambio. `expired` es un ofrecimiento sin respuesta a las 24 horas.
	Status string `json:"status" contract:"enum=pending|accepted|rejected|cancelled|expired"`
	// Momento en que se hizo el ofrecimiento.
	CreatedAt time.Time `json:"created_at"`
	// Momento en que se resolvió; no viaja mientras está pendiente.
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}
