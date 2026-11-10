package platform

import "github.com/discordia-grupo01/events-contract/envelope"

const RoutingKeyStatusChanged = "platform.status_changed"
const RoutingKeyHealthChanged = "platform.health_changed"

type StatusChanged struct {
	envelope.Meta
	Status string `json:"status" contract:"enum=operational|degraded|maintenance"`
}

type HealthChanged struct {
	envelope.Meta
	Service             string   `json:"service"`
	PreviousStatus      string   `json:"previous_status" contract:"enum=unknown|operational|degraded"`
	Status              string   `json:"status" contract:"enum=operational|degraded"`
	ConsecutiveFailures int      `json:"consecutive_failures" contract:"minimum=0"`
	AudienceUserIDs     []string `json:"audience_user_ids"`
}
