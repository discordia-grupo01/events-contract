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
	ServerID    string   `json:"server_id"`
	RoleID      string   `json:"role_id"`
	Permissions []string `json:"permissions"`
	IsEveryone  bool     `json:"is_everyone"`
}

type RoleDeleted struct {
	envelope.Meta
	ServerID string `json:"server_id"`
	RoleID   string `json:"role_id"`
}

type MemberRolesUpdated struct {
	envelope.Meta
	ServerID string   `json:"server_id"`
	UserID   string   `json:"user_id"`
	RoleIDs  []string `json:"role_ids"`
}

type RoleOverride struct {
	RoleID string   `json:"role_id"`
	Allow  []string `json:"allow"`
	Deny   []string `json:"deny"`
}

type ChannelOverridesUpdated struct {
	envelope.Meta
	ServerID  string         `json:"server_id"`
	ChannelID string         `json:"channel_id"`
	Overrides []RoleOverride `json:"overrides"`
}

type OwnerUpdated struct {
	envelope.Meta
	ServerID string `json:"server_id"`
	OwnerID  string `json:"owner_id"`
}
