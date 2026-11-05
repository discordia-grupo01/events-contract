package platform

import "github.com/discordia-grupo01/events-contract/envelope"

const RoutingKeyStatusChanged = "platform.status_changed"

type StatusChanged struct {
	envelope.Meta
	Status string `json:"status"`
}
