package platform

import "github.com/discordia-grupo01/events-contract/envelope"

const RoutingKeyStatusChanged = "platform.status_changed"
const RoutingKeyHealthChanged = "platform.health_changed"

type StatusChanged struct {
	envelope.Meta
	Status string `json:"status"`
}

type HealthChanged struct {
	envelope.Meta
	Service             string   `json:"service"`
	PreviousStatus      string   `json:"previous_status"`
	Status              string   `json:"status"`
	ConsecutiveFailures int      `json:"consecutive_failures"`
	AudienceUserIDs     []string `json:"audience_user_ids"`
}
