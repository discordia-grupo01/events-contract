package permissions

import "github.com/discordia-grupo01/events-contract/envelope"

const (
	RoutingKeyRoleUpdated             = "servers.role_updated"
	RoutingKeyRoleDeleted             = "servers.role_deleted"
	RoutingKeyMemberRolesUpdated      = "servers.member_roles_updated"
	RoutingKeyChannelOverridesUpdated = "servers.channel_overrides_updated"
	RoutingKeyOwnerUpdated            = "servers.owner_updated"
)

type RoleUpdated struct {
	envelope.Meta
	ServerID string `json:"server_id" contract:"format=uuid"`
	RoleID   string `json:"role_id" contract:"format=uuid"`
	// Permisos del rol a nivel servidor, ordenados, sin repetidos. Puede ser vacío, nunca null.
	Permissions []string `json:"permissions"`
	// true para el rol `@everyone`, que todo miembro tiene y no se puede borrar.
	IsEveryone bool `json:"is_everyone"`
}

type RoleDeleted struct {
	envelope.Meta
	ServerID string `json:"server_id" contract:"format=uuid"`
	RoleID   string `json:"role_id" contract:"format=uuid"`
}

type MemberRolesUpdated struct {
	envelope.Meta
	ServerID string `json:"server_id" contract:"format=uuid"`
	// ID del usuario en discordia-identify.
	UserID string `json:"user_id"`
	// Todos los roles del miembro, `@everyone` incluido, ordenados, sin repetidos. Nunca null.
	RoleIDs []string `json:"role_ids" contract:"itemFormat=uuid"`
}

type RoleOverride struct {
	RoleID string   `json:"role_id"`
	Allow  []string `json:"allow"`
	Deny   []string `json:"deny"`
}

type ChannelOverridesUpdated struct {
	envelope.Meta
	ServerID  string `json:"server_id" contract:"format=uuid"`
	ChannelID string `json:"channel_id" contract:"format=uuid"`
	// Un elemento por rol que tiene override en el canal. Vacío si no queda ninguno, nunca null.
	Overrides []RoleOverride `json:"overrides"`
}

type OwnerUpdated struct {
	envelope.Meta
	ServerID string `json:"server_id" contract:"format=uuid"`
	// ID del dueño en discordia-identify.
	OwnerID string `json:"owner_id"`
}
