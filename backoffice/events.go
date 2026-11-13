package backoffice

import "github.com/discordia-grupo01/events-contract/envelope"

const RoutingKeyChanged = "admin.backoffice_changed"

// Changed is published by admin when something the back-office shows to staff changed: a user, a server, a report, the
// staff itself, the investigation file of a user or a server, or an account recovery request. It carries no data, only what to ask for again, and the
// staff members who should be told, so notifications can push it to them without knowing who is staff. It is not signed:
// the worst a forged one can do is make a staff screen ask again for what it already shows.
type Changed struct {
	envelope.Meta
	// Qué volver a pedir: usuarios, servidores, reportes, el equipo de staff, un expediente, las métricas o las solicitudes de recuperación de cuenta.
	Resource string `json:"resource" contract:"enum=users|servers|reports|staff|dossier|metrics|recovery_requests"`
	// Miembros del staff a los que hay que avisar.
	AudienceUserIDs []string `json:"audience_user_ids"`
}
