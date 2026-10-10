defmodule Discordia.EventsContract.Permissions do
  @moduledoc """
  Elixir binding for the `permissions` event domain: the snapshots that let a
  consumer keep its own copy of who may do what in a server, instead of asking
  `servers` on every request. Reference copy, hand-ported by `messaging` into
  its own `Messaging.Events.ChannelEvent` module; the JSON wire shape is the
  contract.
  """

  @routing_key_role_updated "servers.role_updated"
  @routing_key_role_deleted "servers.role_deleted"
  @routing_key_member_roles_updated "servers.member_roles_updated"
  @routing_key_channel_overrides_updated "servers.channel_overrides_updated"
  @routing_key_owner_updated "servers.owner_updated"

  def routing_key_role_updated, do: @routing_key_role_updated
  def routing_key_role_deleted, do: @routing_key_role_deleted
  def routing_key_member_roles_updated, do: @routing_key_member_roles_updated
  def routing_key_channel_overrides_updated, do: @routing_key_channel_overrides_updated
  def routing_key_owner_updated, do: @routing_key_owner_updated

  defmodule RoleUpdated do
    @moduledoc "Full snapshot of a role, published on creation and on every change."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            role_id: String.t(),
            permissions: [String.t()],
            is_everyone: boolean()
          }

    defstruct [:event_id, :occurred_at, :server_id, :role_id, :permissions, :is_everyone]
  end

  defmodule RoleDeleted do
    @moduledoc "Published when a role is deleted. Its members and overrides lose it."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            role_id: String.t()
          }

    defstruct [:event_id, :occurred_at, :server_id, :role_id]
  end

  defmodule MemberRolesUpdated do
    @moduledoc "Full set of roles of a member, @everyone included."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            user_id: String.t(),
            role_ids: [String.t()]
          }

    defstruct [:event_id, :occurred_at, :server_id, :user_id, :role_ids]
  end

  defmodule ChannelOverridesUpdated do
    @moduledoc "Every per-role override of a channel; an empty list means none."

    @type override :: %{
            String.t() => String.t() | [String.t()]
          }

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            channel_id: String.t(),
            overrides: [override()]
          }

    defstruct [:event_id, :occurred_at, :server_id, :channel_id, :overrides]
  end

  defmodule OwnerUpdated do
    @moduledoc "Owner of a server, published at its creation and on every ownership transfer."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            owner_id: String.t()
          }

    defstruct [:event_id, :occurred_at, :server_id, :owner_id]
  end
end
