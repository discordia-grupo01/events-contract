package envelope

import "time"

type Meta struct {
	// Identificador único del evento. Usar para deduplicar (at-least-once).
	EventID string `json:"event_id" contract:"format=uuid"`
	// Momento del hecho de negocio, ISO 8601 en UTC.
	OccurredAt time.Time `json:"occurred_at"`
}
