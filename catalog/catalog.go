package catalog

import (
	"reflect"

	"github.com/discordia-grupo01/events-contract/admin"
	"github.com/discordia-grupo01/events-contract/billing"
	"github.com/discordia-grupo01/events-contract/boosts"
	"github.com/discordia-grupo01/events-contract/broadcast"
	"github.com/discordia-grupo01/events-contract/broker"
	"github.com/discordia-grupo01/events-contract/channels"
	"github.com/discordia-grupo01/events-contract/membership"
	"github.com/discordia-grupo01/events-contract/messages"
	"github.com/discordia-grupo01/events-contract/moderation"
	"github.com/discordia-grupo01/events-contract/permissions"
	"github.com/discordia-grupo01/events-contract/platform"
	"github.com/discordia-grupo01/events-contract/presence"
	"github.com/discordia-grupo01/events-contract/reports"
	"github.com/discordia-grupo01/events-contract/servers"
	"github.com/discordia-grupo01/events-contract/users"
	"github.com/discordia-grupo01/events-contract/voice"
)

type Entry struct {
	RoutingKey   string
	Payload      any
	Producer     string
	Consumers    []string
	Exchange     string
	ExchangeType string
	Signed       bool
	Internal     bool
	Omit         []string
	Require      []string
}

func (e Entry) Type() reflect.Type {
	return reflect.TypeOf(e.Payload)
}

func event(key string, payload any, producer string, consumers ...string) Entry {
	return Entry{
		RoutingKey:   key,
		Payload:      payload,
		Producer:     producer,
		Consumers:    consumers,
		Exchange:     broker.ExchangeName,
		ExchangeType: broker.ExchangeType,
	}
}

func signed(e Entry) Entry {
	e.Signed = true
	return e
}

func internal(event string, payload any) Entry {
	return Entry{
		RoutingKey:   broadcast.KeyPrefix + event,
		Payload:      payload,
		Producer:     "messaging",
		Consumers:    []string{"messaging"},
		Exchange:     broadcast.ExchangeName,
		ExchangeType: broadcast.ExchangeType,
		Internal:     true,
	}
}

func without(e Entry, fields ...string) Entry {
	e.Omit = fields
	return e
}

func requiring(e Entry, fields ...string) Entry {
	e.Require = fields
	return e
}

func All() []Entry {
	return []Entry{
		signed(event(admin.RoutingKeyServerDeletionRequested, admin.ServerDeletionRequested{}, "admin", "servers")),
		signed(event(admin.RoutingKeyServerReactivated, admin.ServerReactivated{}, "admin", "messaging", "servers")),
		signed(event(admin.RoutingKeyServerSuspended, admin.ServerSuspended{}, "admin", "messaging", "servers", "voice-sfu")),
		signed(event(admin.RoutingKeyStaffMemberAdded, admin.StaffMemberAdded{}, "admin")),
		signed(event(admin.RoutingKeyStaffMemberRemoved, admin.StaffMemberRemoved{}, "admin")),
		signed(event(admin.RoutingKeyUserReactivated, admin.UserReactivated{}, "admin", "identify-service", "messaging", "servers", "voice-sfu")),
		signed(event(admin.RoutingKeyUserSuspended, admin.UserSuspended{}, "admin", "identify-service", "messaging", "servers", "voice-sfu")),
		signed(event(billing.RoutingKeyPremiumUpdated, billing.PremiumUpdated{}, "billing", "identify-service", "messaging")),
		signed(without(event(boosts.RoutingKeyRequested, boosts.SagaEvent{}, "billing", "servers"), "reason")),
		signed(without(event(boosts.RoutingKeyCancelled, boosts.SagaEvent{}, "billing", "servers"), "reason")),
		signed(without(event(boosts.RoutingKeyRemoved, boosts.SagaEvent{}, "billing", "servers"), "reason")),
		signed(without(event(boosts.RoutingKeyApplied, boosts.SagaEvent{}, "servers", "billing"), "reason")),
		signed(requiring(event(boosts.RoutingKeyRejected, boosts.SagaEvent{}, "servers", "billing"), "reason")),
		event(boosts.RoutingKeyUpdated, boosts.ServerBoostUpdated{}, "servers", "messaging"),
		event(channels.RoutingKeyChannelAccessRevoked, channels.ChannelAccessRevoked{}, "servers", "messaging", "voice-sfu"),
		event(channels.RoutingKeyChannelCreated, channels.ChannelCreated{}, "servers", "messaging", "notifications"),
		event(channels.RoutingKeyChannelDeleted, channels.ChannelDeleted{}, "servers", "messaging", "notifications", "voice-sfu"),
		event(membership.RoutingKeyMemberJoined, membership.MemberJoined{}, "servers", "admin", "identify-service", "notifications"),
		event(membership.RoutingKeyMemberLeft, membership.MemberLeft{}, "servers", "admin", "identify-service", "notifications"),
		event(messages.RoutingKeyDirectMessageCreated, messages.DirectMessageCreated{}, "messaging", "notifications"),
		event(messages.RoutingKeyMessageCreated, messages.MessageCreated{}, "messaging", "notifications"),
		event(messages.RoutingKeyMessageDeleted, messages.MessageDeleted{}, "messaging", "notifications"),
		event(messages.RoutingKeyMessageUpdated, messages.MessageUpdated{}, "messaging", "notifications"),
		event(moderation.RoutingKeyMemberBanned, moderation.MemberBanned{}, "servers"),
		event(moderation.RoutingKeyMemberKicked, moderation.MemberKicked{}, "servers"),
		event(moderation.RoutingKeyMemberMuted, moderation.MemberMuted{}, "servers"),
		event(moderation.RoutingKeyMemberUnbanned, moderation.MemberUnbanned{}, "servers"),
		event(moderation.RoutingKeyMemberUnmuted, moderation.MemberUnmuted{}, "servers"),
		event(moderation.RoutingKeyModerationActionRecorded, moderation.ModerationActionRecorded{}, "servers"),
		event(moderation.RoutingKeyWordFilterUpdated, moderation.WordFilterUpdated{}, "servers"),
		event(permissions.RoutingKeyChannelOverridesUpdated, permissions.ChannelOverridesUpdated{}, "servers", "messaging", "notifications"),
		event(permissions.RoutingKeyMemberRolesUpdated, permissions.MemberRolesUpdated{}, "servers", "messaging", "notifications"),
		event(permissions.RoutingKeyOwnerUpdated, permissions.OwnerUpdated{}, "servers", "messaging", "notifications"),
		event(permissions.RoutingKeyRoleDeleted, permissions.RoleDeleted{}, "servers", "messaging", "notifications"),
		event(permissions.RoutingKeyRoleUpdated, permissions.RoleUpdated{}, "servers", "messaging", "notifications"),
		event(platform.RoutingKeyHealthChanged, platform.HealthChanged{}, "metrics", "notifications"),
		event(platform.RoutingKeyStatusChanged, platform.StatusChanged{}, "metrics", "notifications"),
		event(presence.RoutingKeyUserPresenceChanged, presence.UserPresenceChanged{}, "messaging", "notifications"),
		event(reports.RoutingKeyReportUpdated, reports.ReportUpdated{}, "servers", "admin"),
		event(servers.RoutingKeyEmojiCreated, servers.EmojiCreated{}, "servers", "messaging"),
		event(servers.RoutingKeyEmojiDeleted, servers.EmojiDeleted{}, "servers", "messaging"),
		event(servers.RoutingKeyServerDeleted, servers.ServerDeleted{}, "servers", "admin", "messaging", "notifications", "voice-sfu"),
		event(servers.RoutingKeyServerUpdated, servers.ServerUpdated{}, "servers", "admin"),
		event(users.RoutingKeyAccountRegistered, users.AccountRegistered{}, "identify-service", "admin"),
		event(users.RoutingKeyEmailVerified, users.EmailVerified{}, "identify-service", "admin"),
		event(users.RoutingKeyProfileUpdated, users.ProfileUpdated{}, "identify-service", "admin", "messaging", "servers"),
		event(voice.RoutingKeyParticipantJoined, voice.ParticipantJoined{}, "voice-sfu"),
		event(voice.RoutingKeyParticipantLeft, voice.ParticipantLeft{}, "voice-sfu"),
		event(voice.RoutingKeyPrivateCall, voice.PrivateCall{}, "voice-sfu"),
		internal(broadcast.EventNewMessage, broadcast.NewMessage{}),
		internal(broadcast.EventMessageUpdated, broadcast.MessageUpdated{}),
		internal(broadcast.EventMessageDeleted, broadcast.MessageDeleted{}),
		internal(broadcast.EventReactionUpdated, broadcast.ReactionUpdated{}),
		internal(broadcast.EventNewDM, broadcast.NewDM{}),
		internal(broadcast.EventMention, broadcast.Mention{}),
		internal(broadcast.EventAccessRevoked, broadcast.AccessRevoked{}),
		internal(broadcast.EventServerLeft, broadcast.ServerLeft{}),
		internal(broadcast.EventServerSuspended, broadcast.ServerSuspended{}),
		internal(broadcast.EventServerReactivated, broadcast.ServerReactivated{}),
		internal(broadcast.EventDisconnect, broadcast.Disconnect{}),
		internal(broadcast.EventPresenceUpdated, broadcast.PresenceUpdated{}),
		internal(broadcast.EventOwnStatusUpdated, broadcast.OwnStatusUpdated{}),
	}
}

func ByRoutingKey(key string) (Entry, bool) {
	for _, e := range All() {
		if e.RoutingKey == key {
			return e, true
		}
	}
	return Entry{}, false
}
