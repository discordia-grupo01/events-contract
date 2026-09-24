package envelope

import "time"

type Meta struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
