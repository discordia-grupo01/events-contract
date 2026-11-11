package account

import "github.com/discordia-grupo01/events-contract/envelope"

const RoutingKeyChanged = "billing.account_changed"

// Changed is published by billing every time something of a user's account that the clients show changes: the points
// balance and its history (a purchase confirmed, points spent or given back by a boost), the boosts the user holds, or
// the purchases. It carries no data, only what to ask for again, so a repeated or out-of-order delivery is harmless.
// Premium has its own signed snapshot, billing.premium_updated. It is not signed: the worst a forged one can do is make a
// screen ask again for what it already shows.
type Changed struct {
	envelope.Meta
	// Usuario al que se le avisa (ID de discordia-identify).
	UserID string `json:"user_id" contract:"format=uuid"`
	// Qué volver a pedir: el saldo de puntos, los boosts del usuario o sus compras.
	Resource string `json:"resource" contract:"enum=points|boosts|purchases"`
}
